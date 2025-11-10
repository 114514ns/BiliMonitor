package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	bili "github.com/114514ns/BiliClient"
	pool2 "github.com/sourcegraph/conc/pool"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
)

var worker = NewWorker(1)

type FrontAreaLiver struct {
	AreaLiver
	LastActive  time.Time
	DailyDiff   int
	MonthlyDiff int
	Verify      string
	Bio         string
}

//go:embed Page/dist
var distFS embed.FS

func InitHTTP() {
	r := gin.Default()
	r.UseH2C = true
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(CORSMiddleware())
	r.Use(TTLMiddleware())
	if ENV == "BUILD" {
		r.Use(static.Serve("/", static.EmbedFolder(distFS, "Page/dist")))
	} else {
		r.Use(static.Serve("/", static.LocalFile("./Page/dist", false)))
	}

	r.GET("/monitor", func(c *gin.Context) {

		var array = make([]Status, 0)
		var wg sync.WaitGroup
		var lock sync.Mutex

		if config.Mode == "Master" {
			if man == nil {
				return
			}
			for _, node := range man.Nodes {
				if !node.Alive {
					continue
				}
				if node.Address == "http://127.0.0.1:"+strconv.Itoa(int(config.Port)) {
					continue
				}

				wg.Add(1)
				go func(n SlaverNode) {
					defer wg.Done()
					var result map[string][]Status

					_, err := client.R().
						SetResult(&result).
						Get(n.Address + "/monitor")

					if err != nil {
						log.Printf("请求子节点 %s 失败: %v", n.Address, err)
						return
					}
					if remoteStatuses, ok := result["lives"]; ok {
						lock.Lock()
						array = append(array, remoteStatuses...)
						lock.Unlock()
					}
				}(node)
			}

			wg.Wait()
		} else {
			for _, status := range array {
				if status.Live {
					array = append(array, status)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"lives": array,
		})
	})
	r.GET("/searchLiver", func(c *gin.Context) {
		key := c.DefaultQuery("key", "")
		type LiverInfo struct {
			UName string
			Room  int
			UID   int64
			Money float64 `json:"-"`
		}
		var results []LiverInfo
		query := `
        SELECT 
            l1.user_name as u_name,
            l1.room_id as room,
            CAST(l1.user_id AS SIGNED) as uid,
            l1.money
        FROM lives l1
        INNER JOIN (
            SELECT room_id, MAX(money) as max_money
            FROM lives
            WHERE user_name LIKE ?
            GROUP BY room_id
        ) l2 ON l1.room_id = l2.room_id AND l1.money = l2.max_money
        WHERE l1.user_name LIKE ?
        ORDER BY l1.money DESC
        LIMIT 30
    `
		searchPattern := "%" + key + "%"
		err := db.Raw(query, searchPattern, searchPattern).Scan(&results).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "查询失败",
				"message": err.Error(),
				"result":  []interface{}{},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"result": results,
			"count":  len(results),
		})
	})
	r.GET("/searchAreaLiver", func(c *gin.Context) {
		key := c.DefaultQuery("key", "1")
		var result = make([]AreaLiver, 0)

		if key != "1" {
			for _, liver := range cachedLivers {
				if strings.Contains(liver.UName, key) {
					result = append(result, liver.AreaLiver)
				}
			}
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].Fans > result[j].Fans
		})

		if len(result) > 10 {
			result = result[:10]
		}

		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})

	})
	r.GET("/live", func(c *gin.Context) {
		var f []Live
		name := c.DefaultQuery("name", "1")
		pageStr := c.DefaultQuery("page", "1")
		page, _ := strconv.Atoi(pageStr)
		limitStr := c.DefaultQuery("limit", "10")
		limit, _ := strconv.Atoi(limitStr)
		order := c.DefaultQuery("order", "id")
		//var liverStr = c.DefaultQuery("liver", "")
		var uidStr = c.DefaultQuery("uid", "-1")
		var noDM = c.DefaultQuery("no_dm", "false")

		if noDM == "true" {
			var dst []AreaLive
			db.Raw("select * from area_lives where uid =? order by id desc", uidStr).Scan(&dst)
			copier.Copy(&f, dst)
			for i, live := range dst {
				f[i].CreatedAt = live.Time
			}
			c.JSON(http.StatusOK, gin.H{
				"lives": f,
			})
			return

		}

		offset := (page - 1) * limit
		var totalRecords int64

		if err := db.Model(&Live{}).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}
		validOrders := map[string]bool{
			"id":      true,
			"money":   true,
			"message": true,
			"watch":   true,
		}
		if !validOrders[order] {
			order = "id"
		}

		query := db.Model(&Live{}).Order(order + " desc").Offset(offset).Limit(limit)
		if toInt64(uidStr) > 0 {
			query = query.Where("user_id = ?", uidStr)
			query.Count(&totalRecords)
		}

		if name == "1" {
			query.Find(&f)
		} else {
			query = query.Where("user_name = ?", name)
			query.Find(&f)
			db.Model(&Live{}).Where("user_name = ?", name).Count(&totalRecords)
		}

		if toInt64(uidStr) != 0 {
			query = query.Where("user_id = ?", uidStr)
			query.Find(&f)
		}

		var off int64 = 1
		if totalRecords%int64(limit) == 0 {
			off = 0
		}

		c.JSON(http.StatusOK, gin.H{
			"totalPage": totalRecords/int64(limit) + off,
			"lives":     f,
		})
	})
	r.GET("/liveDetail/:id/", func(c *gin.Context) {
		id := c.Param("id")
		var obj = Live{}
		db.Model(&Live{}).Where("id = ?", id).Find(&obj)
		c.JSON(http.StatusOK, gin.H{
			"live": obj,
		})
	})
	r.GET("/live/:id/", func(c *gin.Context) {
		id := c.Param("id")
		Type := c.DefaultQuery("type", "")
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		mid := c.DefaultQuery("mid", "0")
		midInt := toInt64(mid)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
			return
		}

		limitStr := c.DefaultQuery("limit", "10")
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit

		orderStr := c.DefaultQuery("order", "ascend")

		query := db.Model(&LiveAction{}).Where("live = ? and action_name != 'enter'", id)

		if Type != "" {
			query = query.Where("action_name = ?", Type)
		}

		orderQuery := ""
		if orderStr == "ascend" {
			orderQuery = "gift_price asc"
		} else if orderStr == "descend" {
			orderQuery = "gift_price desc"
		} else {
			orderQuery = "id asc"
		}

		var records []LiveAction
		query = db.Model(&LiveAction{}).Where("live = ? and action_name != 'enter'", id)

		if Type != "" {
			query = query.Where("action_name = ?", Type)
		}
		if midInt != 0 {
			query = query.Where("from_id = ?", midInt)
		}
		var totalRecords int64
		if err := query.Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if err := query.Order(orderQuery).Offset(offset).Limit(limit).Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database query error"})
			return
		}

		var liveObj = &Live{}

		db.Model(&Live{}).Where("id = ?", id).Find(&liveObj)

		c.JSON(http.StatusOK, gin.H{
			"totalPages":   totalPages,
			"totalRecords": totalRecords,
			"page":         page,
			"records":      records,
			"liver":        liveObj.UserName,
		})
	})
	r.GET("/add/:id", func(context *gin.Context) {
		id := context.Param("id")
		if lives[id] == nil {
			lives[id] = &Status{}
			man.AddTask(id)
			config.Tracing = append(config.Tracing, id)
			SaveConfig()
			context.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		} else {
			context.JSON(http.StatusOK, gin.H{
				"message": "live has already exist",
			})
		}
	})
	r.NoRoute(func(c *gin.Context) {
		c.File("./Page/dist/index.html")
	})

	r.GET("/proxy", func(c *gin.Context) {
		var url = c.Query("url")
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is empty"})
			return
		}
		res, _ := queryClient.R().SetHeader("Referer", "https://www.bilibili.com/").Get(url)
		c.Writer.Header().Set("Content-Type", res.Header().Get("Content-Type"))
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		c.Writer.WriteHeader(res.StatusCode())
		c.Writer.Write(res.Body())
	})

	r.GET("/status", func(c *gin.Context) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("upgrade err: %v", err)
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		conn.SetPongHandler(func(appData string) error {
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			return nil
		})

		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Printf("read error (client inactive?): %v", err)
					conn.Close()
					return
				}
				_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			}
		}()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Printf("[StatusResponse]\n")
				var wg sync.WaitGroup
				var t1, t2, t3 int
				var l sync.Mutex
				var bytes int64
				var m []interface{}

				wg.Add(3)
				go func() {
					t1 = int(MinuteMessageCount(1))
					wg.Done()
				}()
				go func() {
					t2 = int(MinuteMessageCount(60))
					wg.Done()
				}()
				go func() {
					t3 = int(MinuteMessageCount(1440))
					wg.Done()
				}()
				wg.Wait()

				if man == nil || man.Nodes == nil {
					data := gin.H{
						"LaunchedAt":    launchTime.Format(time.DateTime),
						"WSBytes":       websocketBytes,
						"TotalMessages": TotalMessage(),
						"Message1":      t1,
						"MessageHour":   t2,
						"MessageDaily":  t3,
					}
					_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
					if err := conn.WriteJSON(data); err != nil {
						log.Printf("write err: %v", err)
						return
					}
					continue
				}

				for _, node := range man.Nodes {
					if !node.Alive {
						continue
					}
					if node.Address == "http://127.0.0.1:"+strconv.Itoa(int(config.Port)) {
						continue
					}
					wg.Add(1)
					go func(n SlaverNode) {
						defer wg.Done()
						res, _ := client.R().Get(n.Address + "/status")
						type S struct {
							WSBytes int64
							Rooms   map[string]interface{}
						}
						var s S
						sonic.Unmarshal(res.Body(), &s)
						l.Lock()
						bytes += s.WSBytes
						for _, status := range lives {
							m = append(m, status)
						}
						l.Unlock()
					}(node)
				}
				wg.Wait()

				data := gin.H{
					"Requests":      totalRequests,
					"Tasks":         guardWorker.QueueLen(),
					"LaunchedAt":    launchTime.Format(time.DateTime),
					"TotalMessages": TotalMessage(),
					"Message1":      t1,
					"MessageHour":   t2,
					"MessageDaily":  t3,
					"HTTPBytes":     httpBytes,
					"WSBytes":       websocketBytes + bytes,
					"Nodes":         man.Nodes,
					"Rooms":         m,
				}

				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteJSON(data); err != nil {
					log.Printf("write err: %v", err)
					return
				}
			}
		}
	})

	//获取头像，会302到b站的bfs

	r.GET("/face", func(c *gin.Context) {
		if c.Query("mid") == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing mid",
			})
		}
		c.Redirect(http.StatusMovedPermanently, GetFace(c.Query("mid")))
	})

	//所有在虚拟区开播过的主播

	r.GET("/areaLivers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"list": cachedLivers,
		})

	})
	//获取直播流
	r.GET("/stream", func(c *gin.Context) {
		var room = c.Query("room")
		c.JSON(http.StatusOK, gin.H{
			"Stream": GetLiveStream(room),
		})
	})

	//Trace直播间

	r.GET("/trace", func(c *gin.Context) {
		var room = c.Query("room")
		livesMutex.Lock()
		lives[room] = &Status{RemainTrying: 40}
		lives[room].LiveRoom = room
		livesMutex.Unlock()
		worker.AddTask(func() {
			go TraceLive(room)
			time.Sleep(10 * time.Second)
		})
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
		log.Printf("从[%s]接收到任务：%s", c.RemoteIP(), room)
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	//数据库内所有粉丝牌，每个用户在不同主播的粉丝牌分开计算
	r.GET("/fansRank", func(context *gin.Context) {
		var page = context.DefaultQuery("page", "1")
		var size = context.DefaultQuery("size", "20")
		var liver = context.DefaultQuery("liver", "")
		var active = context.DefaultQuery("active", "false")
		pageInt, _ := strconv.Atoi(page)
		sizeInt, _ := strconv.Atoi(size)
		var result []FansClub
		query := db.Model(&FansClub{})
		var countQuery = db.Model(&FansClub{})
		var s = ""
		if active == "true" {
			s = "and TO_DAYS(NOW()) -TO_DAYS(updated_at) <7"
		}
		if liver != "" {
			query.Where("liver_id = ? "+s, liver)
			countQuery.Where("liver_id = ? "+s, liver)
		}

		var count int64
		query.Count(&count)
		query.Offset(sizeInt * (pageInt - 1)).Limit(sizeInt).Order("level desc").Find(&result)
		context.JSON(http.StatusOK, gin.H{
			"list":  result,
			"total": count,
		})
	})

	//获取某场直播内发送弹幕或礼物的所有用户

	r.GET("/liveUser", func(context *gin.Context) {
		var live = context.DefaultQuery("live", "-1")
		var keyword = context.DefaultQuery("keyword", "")
		var liveInt = toInt64(live)
		if live == "-1" || liveInt == 0 {
			context.JSON(http.StatusOK, gin.H{
				"message": "missing or wrong params:live",
			})
			return
		}

		var dst []FrontLiveAction
		db.Raw("SELECT from_id,from_name,medal_level,medal_name,guard_level FROM live_actions where live = ?  GROUP BY from_id order by medal_level desc", liveInt).Scan(&dst)
		if keyword != "" {
			var tmp []FrontLiveAction
			for _, watcher := range dst {
				if strings.Contains(watcher.FromName, keyword) {
					tmp = append(tmp, watcher)
				}
			}
			dst = tmp
		}
		if len(dst) > 50 {
			dst = dst[:50]
		}
		if dst == nil {
			dst = []FrontLiveAction{}
		}
		context.JSON(http.StatusOK, gin.H{
			"list": dst,
		})

	})

	r.GET("/medals", func(context *gin.Context) {
		var midStr = context.DefaultQuery("mid", "0")
		mid := toInt64(midStr)
		if mid == 0 {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid mid",
			})
		}

		var dst []FansClub
		db.Raw("select * FROM fans_clubs WHERE uid = ?", mid).Scan(&dst)
		if dst == nil {
			dst = []FansClub{}
		}
		sort.Slice(dst, func(i, j int) bool {
			return dst[i].Score > dst[j].Score
		})
		context.JSON(http.StatusOK, gin.H{
			"list": dst,
		})

	})

	r.GET("/user/space", func(context *gin.Context) {
		var uidStr = context.DefaultQuery("uid", "")
		if uidStr == "" {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid uid",
			})
			return
		}
		var uid = toInt64(uidStr)
		var wg sync.WaitGroup
		var totalMessage = 0
		var totalMoney = 0.0
		var lastSeen time.Time
		var firstSeen time.Time
		var name = ""
		type Room struct {
			Liver    string
			LiverID  int64
			Count    int
			Rate     float64
			LiveRoom int
		}
		var medals int
		var topMedal int
		var rooms []Room
		wg.Add(7)
		go func() {
			db.Raw("select count(*) from live_actions where from_id = ? ", uid).Scan(&totalMessage)
			wg.Done()
		}()
		go func() {
			db.Raw("SELECT COALESCE(SUM(gift_price), 0)  FROM live_actions WHERE from_id = ?", uid).Scan(&totalMoney)
			wg.Done()
		}()
		go func() {
			var obj LiveAction
			db.Raw("select from_name,created_at from live_actions where from_id = ? order by id desc limit 1", uid).Scan(&obj)
			lastSeen = obj.CreatedAt
			name = obj.FromName
			wg.Done()
		}()
		go func() {
			db.Raw("select created_at from live_actions where from_id = ? order by id limit 1", uid).Scan(&firstSeen)
			wg.Done()
		}()
		go func() {
			db.Raw("SELECT live_room,COUNT(live_room) as count,user_id as liver_id,user_name as liver  FROM live_actions,lives where from_id = ? and live_actions.live = lives.id  GROUP BY live_room limit 100", uid).Scan(&rooms)
			var totalCount = 0
			for _, room := range rooms {
				totalCount += room.Count
			}
			for j, room := range rooms {
				rooms[j].Rate = float64(room.Count) / float64(totalCount)
			}
			wg.Done()
		}()
		go func() {
			db.Raw("select count(*) from fans_clubs where uid = ?", uid).Scan(&medals)
			wg.Done()
		}()
		go func() {
			db.Raw("select level from fans_clubs where uid = ? order by level desc limit 1", uid).Scan(&topMedal)
			wg.Done()
		}()
		wg.Wait()

		context.JSON(http.StatusOK, gin.H{
			"Money":        totalMoney,
			"Message":      totalMessage,
			"UName":        name,
			"LastSeen":     lastSeen,
			"FirstSeen":    firstSeen,
			"Rooms":        rooms,
			"Medals":       medals,
			"HighestLevel": topMedal,
		})
	})

	r.GET("/user/action", func(context *gin.Context) {
		type CustomLiveAction struct {
			LiveAction
			UserName string
			UserID   string
		}

		var uidStr = context.DefaultQuery("uid", "")
		var pageSizeStr = context.DefaultQuery("pageSize", "10")
		var pageStr = context.DefaultQuery("page", "1")
		var order = context.DefaultQuery("order", "time")
		var typo = context.DefaultQuery("type", "")
		var room = context.DefaultQuery("room", "")
		var showEnter = context.DefaultQuery("enter", "")

		var total = 0
		var queryOrder = ""
		var tableName = "live_actions"
		if showEnter != "" {
			tableName = "enter_action"
		} else {
			if order == "timeDesc" {
				queryOrder = "order by live_actions.id desc"
			}
			if order == "money" {
				queryOrder = "order by gift_price desc"
			}
		}

		if queryOrder == "" {
			queryOrder = "order by " + tableName + ".id"
		}

		if uidStr == "" {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid uid",
			})
			return
		}
		var typeQuery = ""
		if typo == "sc" {
			typeQuery = "and action_name = 'sc' "
		}
		if typo == "gift" {
			typeQuery = "and action_name = 'gift'"
		}
		if typo == "msg" {
			typeQuery = "and action_name = 'msg'"
		}
		if typo == "guard" {
			typeQuery = "and action_name = 'guard'"
		}
		if room != "" {
			typeQuery = typeQuery + "and live_room =  '" + room + "'"
		}
		uid := toInt64(uidStr)
		pageSize := toInt64(pageSizeStr)
		page := toInt64(pageStr)
		if page < 1 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		var dst []CustomLiveAction
		err := db.Raw(
			fmt.Sprintf("select *,%s.id,%s.created_at from %s,lives where from_id = ? and lives.id = %s.live %s %s limit ? offset ?", tableName, tableName, tableName, tableName, typeQuery, queryOrder),
			uid, pageSize, offset,
		).Scan(&dst).Error

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"message": "query failed",
				"error":   err.Error(),
			})
			return
		}
		if showEnter != "" {
			db.Raw("select count(*) from enter_action where from_id = ?", uid).Scan(&total)
		} else {
			if room != "" {
				db.Raw("select count(*) from live_actions where from_id = ? and live_room = ? and  (select lives.id from lives where live_actions.live = lives.id ) != 0",
					uid, room).Scan(&total)
			} else {
				db.Raw("select count(*) from live_actions where from_id = ? and  (select lives.id from lives where live_actions.live = lives.id ) != 0", uid).Scan(&total)
			}
		}

		for _, action := range dst {
			action.ID = action.LiveAction.ID
		}
		context.JSON(http.StatusOK, gin.H{
			"data":  dst,
			"total": total,
		})
	})

	//一场直播内每分钟弹幕数
	r.GET("/chart/live", func(c *gin.Context) {
		var id = c.Query("id")
		type LiveActionCount struct {
			MinuteTime  string `gorm:"column:minute_time"`
			RecordCount int    `gorm:"column:record_count"`
		}

		var results []LiveActionCount

		db.Raw(`
    SELECT 
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') AS minute_time, 
        COUNT(*) AS record_count
    FROM live_actions
    WHERE live = ?
    GROUP BY minute_time
    ORDER BY minute_time;
`, id).Scan(&results)
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    results,
		})
	})

	r.GET("/chart/fans", func(context *gin.Context) {

		var month, _ = strconv.ParseInt(context.DefaultQuery("month", "-3"), 10, 64)
		const POINT = 30
		var end = time.Now()
		var start = end.AddDate(0, int(month)*-1, 0)

		var uid = context.DefaultQuery("uid", "")
		if uid == "" {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid uid",
			})

			return
		}
		var result []User
		//gemini写的
		db.Raw(`
WITH TimeRange AS (
    -- 第一步：获取指定时间范围内的最小和最大时间戳
    SELECT
        MIN(created_at) AS min_time,
        MAX(created_at) AS max_time
    FROM
        users
    WHERE
        user_id = ? -- Placeholder for user_id
        AND created_at BETWEEN ? AND ? -- Placeholders for start_date and end_date
),
TimeBuckets AS (
    -- 第二步：为每一行数据计算它所属的时间桶 (time bucket)
    SELECT
        u.fans,
        u.created_at,
        -- 计算总时间跨度（秒），然后将每条记录的相对时间位置映射到对应的桶中
        -- 使用 GREATEST 防止当 min_time 和 max_time 相同时出现除以零的错误
        FLOOR(
            TIMESTAMPDIFF(SECOND, tr.min_time, u.created_at) * ? / -- Placeholder for num_points
            GREATEST(TIMESTAMPDIFF(SECOND, tr.min_time, tr.max_time), 1)
        ) AS time_bucket
    FROM
        users u,
        TimeRange tr
    WHERE
        u.user_id = ? -- Placeholder for user_id
        AND u.created_at BETWEEN ? AND ? -- Placeholders for start_date and end_date
        AND tr.min_time IS NOT NULL -- 确保时间范围内有数据
),
RankedByBucket AS (
    -- 第三步：在每个时间桶内，根据时间顺序为记录编号
    SELECT
        fans,
        created_at,
        time_bucket,
        ROW_NUMBER() OVER (PARTITION BY time_bucket ORDER BY created_at ASC) AS rn
    FROM
        TimeBuckets
)
-- 第四步：从每个时间桶中只选取第一条记录
SELECT
    fans ,
    created_at
FROM
    RankedByBucket
WHERE
    rn = 1
ORDER BY
    created_at;

	`, uid, start, end, 30, uid, start, end).Scan(&result)

		context.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	})

	r.GET("/chart/guard", func(context *gin.Context) {
		var month, _ = strconv.ParseInt(context.DefaultQuery("month", "-3"), 10, 64)
		var uidStr = context.DefaultQuery("uid", "")
		if uidStr == "" || toInt64(uidStr) == 0 {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid uid",
			})
			return
		}
		var uid = toInt64(uidStr)
		var dst []AreaLiver
		db.Raw(`
WITH TimeRange AS (
    -- 第一步：获取指定时间范围内的最小和最大时间戳
    SELECT
        MIN(updated_at) AS min_time,
        MAX(updated_at) AS max_time
    FROM
        area_livers
    WHERE
        uid = @user_id
        AND updated_at BETWEEN @start_date AND @end_date
),
TimeBuckets AS (
    -- 第二步：为每一行数据计算它所属的时间桶 (time bucket)
    SELECT
        u.guard,
        u.updated_at,
		u.id,
        -- 计算总时间跨度（秒），然后将每条记录的相对时间位置映射到对应的桶中
        -- 使用 GREATEST 防止当 min_time 和 max_time 相同时出现除以零的错误
        FLOOR(
            TIMESTAMPDIFF(SECOND, tr.min_time, u.updated_at) * @num_points /
            GREATEST(TIMESTAMPDIFF(SECOND, tr.min_time, tr.max_time), 1)
        ) AS time_bucket
    FROM
        area_livers u,
        TimeRange tr
    WHERE
        u.uid = @user_id
        AND u.updated_at BETWEEN @start_date AND @end_date
        AND tr.min_time IS NOT NULL -- 确保时间范围内有数据
),
RankedByBucket AS (
    -- 第三步：在每个时间桶内，根据时间顺序为记录编号
    SELECT
        guard,
        updated_at,
		id,
        time_bucket,
        ROW_NUMBER() OVER (PARTITION BY time_bucket ORDER BY updated_at ASC) AS rn
    FROM
        TimeBuckets
)
-- 第四步：从每个时间桶中只选取第一条记录
SELECT
    guard,
    updated_at,
	id
FROM
    RankedByBucket
WHERE
    rn = 1
ORDER BY
    updated_at;

`, map[string]interface{}{
			"user_id":    uid,
			"start_date": time.Now().AddDate(0, int(month)*-1, 0),
			"end_date":   time.Now(),
			"num_points": 30,
		}).Scan(&dst)

		context.JSON(http.StatusOK, gin.H{
			"data": dst,
		})

	})

	r.GET("/chart/msg", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    cachedMessagesPoint,
		})
	})

	r.GET("/liver/space", func(context *gin.Context) {
		var midStr = context.Query("uid")
		if midStr == "" || toInt64(midStr) == 0 {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid uid",
			})
			return
		}

		var mid = toInt64(midStr)
		var l AreaLiver
		var wg sync.WaitGroup
		var medal = ""
		wg.Add(3)
		go func() {

			db.Raw("select u_name,fans,guard,updated_at,area from area_livers where uid = ?  order by id desc", mid).Scan(&l)

			wg.Done()
		}()
		var user = User{}
		go func() {
			db.Raw("select * from users where user_id = ? order by id desc limit 1", mid).Scan(&user)
			wg.Done()
		}()
		go func() {
			db.Raw("select medal_name from fans_clubs where liver_id = ? order by id desc limit 1", mid).Scan(&medal)
			wg.Done()
		}()

		wg.Wait()

		context.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"UName":   l.UName,
			"Fans":    user.Fans,
			"Guard":   l.Guard,
			"Time":    l.UpdatedAt,
			"Area":    l.Area,
			"Bio":     user.Bio,
			"Verify":  user.Verify,
			"Medal":   medal,
		})
	})
	r.GET("/queryPage", func(context *gin.Context) {
		var live = context.DefaultQuery("live", "")
		var id = context.DefaultQuery("id", "")
		var size = context.DefaultQuery("size", "10")
		if live == "" || id == "" || toInt64(live) == 0 || toInt64(id) == 0 || toInt64(size) == 0 {
			context.JSON(http.StatusOK, gin.H{
				"message": "invalid params",
			})
			return
		}
		var dst []int64
		db.Raw("select id from live_actions where live = ? order by id", live).Scan(&dst)

		var index = -1
		var idn = toInt64(id)
		for i, val := range dst {
			if val == idn {
				index = i
				break
			}
		}
		if index == -1 {
			context.JSON(http.StatusOK, gin.H{
				"message": "id not found",
			})
			return
		}
		s := int(toInt64(size))
		page := index/s + 1

		context.JSON(http.StatusOK, gin.H{
			"page": page,
		})
	})

	r.GET("/reload", func(context *gin.Context) {
		livesMutex.Lock()
		loadConfig()
		livesMutex.Unlock()
		context.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	r.GET("/refreshCookie", func(context *gin.Context) {
		RefreshCookie()
		context.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	r.GET("/guard", func(c *gin.Context) {
		var idStr = c.Query("id")
		var inspect = c.Query("inspect")
		if toInt64(idStr) <= 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "invalid id",
			})
			return
		}
		type ExtendGuard struct {
			DBGuard
			MessageCount int
			GiftCount    int
			GuardCount   int
			TimeOut      bool
			OverEnter    bool //满足最低进房次数要求
		}
		var dst []ExtendGuard
		var str = ""
		db.Raw("select guard_list from area_livers where  id = ?", toInt64(idStr)).Scan(&str)
		sonic.Unmarshal([]byte(str), &dst)
		if inspect == "true" {
			var pool = pool2.New().WithMaxGoroutines(32)
			for i, guard := range dst {

				var id = guard.UID
				pool.Go(func() {
					dst[i].TimeOut = false
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					var guardCount = 0
					e := db.WithContext(ctx).Raw("select count(*) from live_actions where from_id = ? and action_type = 3", id).Scan(&guardCount)

					var msgCount = 0
					e1 := db.WithContext(ctx).Raw("select count(*) from live_actions where from_id = ? and action_type = 1", id).Scan(&msgCount)

					if e.Error == context.DeadlineExceeded || e1.Error == context.DeadlineExceeded {
						dst[i].TimeOut = true
					}

					var d []LiveAction
					c0 := db.Raw("select id,created_at from enter_action where from_id = ? limit 25", id).Scan(&d)
					if c0.Error != nil {
						log.Println(c0.Error)
					}

					if len(d) >= 20 {
						dst[i].OverEnter = true
					} else {
						dst[i].OverEnter = false
					}

					dst[i].MessageCount = msgCount
					dst[i].GuardCount = guardCount

					if id == 447689051 {
						time.Now()
					}
				})
			}
			pool.Wait()
			sort.Slice(dst, func(i, j int) bool {
				return dst[i].Level > dst[j].Level
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})
	})
	r.GET("/search", func(context *gin.Context) {
		var typo = context.DefaultQuery("type", "name")
		var key = context.DefaultQuery("key", "")
		var api = context.DefaultQuery("api", "")
		if api == "" {
			context.File("./Page/dist/index.html")
			return
		}
		type Response struct {
			UName      string
			UID        int64
			ExtraInt   int64
			MedalLevel int
			MedalName  string
		}

		var dst []Response
		if typo == "name" {
			var count = 0
			for _, liver := range cachedLivers {
				if count > 30 {
					break
				}
				if strings.Contains(liver.UName, key) {
					count++
					dst = append(dst, Response{
						UName:    liver.UName,
						UID:      liver.UID,
						ExtraInt: int64(liver.Fans),
					})
				}

			}
			sort.Slice(dst, func(i, j int) bool {
				return dst[i].ExtraInt > dst[j].ExtraInt
			})
		}
		if typo == "uid" {
			for _, liver := range cachedLivers {
				if toInt64(key) == liver.UID {
					dst = append(dst, Response{
						UName: liver.UName,
						UID:   liver.UID,
					})
					break
				}

			}
		}
		if typo == "room" {
			var dst0 AreaLive
			db.Raw("select * from area_lives where room = ? limit 1", key).Scan(&dst0)
			dst = append(dst, Response{
				UName: dst0.UName,
				UID:   dst0.UID,
			})
		}
		if typo == "watcher-name" {
			var count = 0
			for _, club := range cachedWatcher {
				if count > 30 {
					break
				}
				if strings.Contains(club.UName, key) {
					count++
					dst = append(dst, Response{
						UName:      club.UName,
						UID:        club.UID,
						MedalLevel: int(club.Level),
						MedalName:  club.MedalName,
					})
				}
			}
		}

		context.JSON(http.StatusOK, gin.H{
			"data": dst,
		})

	})

	// 可配置的超时时间
	const (
		QueryTimeout     = 5 * time.Second // 数据查询超时
		CountTimeout     = 3 * time.Second // 计数查询超时
		FastCountTimeout = 1 * time.Second // 快速计数超时
	)
	r.GET("/raw", func(c *gin.Context) {
		var room = c.DefaultQuery("room", "")
		var typo = c.DefaultQuery("type", "")
		var order = c.DefaultQuery("order", "")
		var size = c.DefaultQuery("size", "10")
		var page = c.DefaultQuery("page", "1")

		if toInt64(room) == 0 {
			room = ""
		}

		pageSize, err := strconv.Atoi(size)
		if err != nil || pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 100 { // 限制最大页面大小
			pageSize = 100
		}

		pageNum, err := strconv.Atoi(page)
		if err != nil || pageNum <= 0 {
			pageNum = 1
		}
		offset := (pageNum - 1) * pageSize

		// SQL 查询语句
		query := `
        SELECT la.*, l.user_id, l.user_name 
        FROM live_actions la
        LEFT JOIN lives l ON l.id = la.live
        WHERE 1=1
    `
		var args []interface{}

		if room != "" {
			query += " AND la.live_room = ?"
			args = append(args, room)
		}
		if typo != "" {
			query += " AND la.action_type = ?"
			args = append(args, typo)
		}

		// 添加排序
		switch order {
		case "created_at_asc":
			query += " ORDER BY la.created_at ASC"
		case "created_at_desc":
			query += " ORDER BY la.created_at DESC"
		case "money_desc":
			query += " ORDER BY la.gift_price DESC"
		default:
			// 默认排序可以根据需要添加
			// query += " ORDER BY la.id DESC"
		}

		query += " LIMIT ? OFFSET ?"
		args = append(args, pageSize, offset)

		type ExtendAction struct {
			LiveAction
			UserID   int64
			UserName string
		}
		var results []ExtendAction
		var total int64
		var wg sync.WaitGroup
		var queryErr, countErr error

		// 创建一个共享的取消上下文（可选）
		rootCtx, rootCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer rootCancel()

		wg.Add(2)

		// 查询数据
		go func() {
			defer wg.Done()

			// 创建查询超时上下文
			ctx, cancel := context.WithTimeout(rootCtx, QueryTimeout)
			defer cancel()

			// 创建一个通道来接收查询结果
			done := make(chan error, 1)

			go func() {
				done <- db.WithContext(ctx).Raw(query, args...).Scan(&results).Error
			}()

			select {
			case queryErr = <-done:
				// 查询完成
			case <-ctx.Done():
				// 超时
				if ctx.Err() == context.DeadlineExceeded {
					queryErr = fmt.Errorf("查询超时（%v），请尝试缩小查询范围或使用更精确的筛选条件", QueryTimeout)
				} else {
					queryErr = ctx.Err()
				}
			}
		}()

		// 查询总数
		go func() {
			defer wg.Done()
			// 对于全量查询使用近似值
			if (typo == "1" || typo == "") && room == "" {
				// 快速获取最大ID
				ctx, cancel := context.WithTimeout(rootCtx, FastCountTimeout)
				defer cancel()

				var maxID int64
				err := db.WithContext(ctx).Raw("SELECT id FROM live_actions ORDER BY id DESC LIMIT 1").Scan(&maxID).Error

				if err == nil {
					total = maxID
				} else if err == context.DeadlineExceeded {
					// 快速查询超时，尝试缓存或返回估算值
					total = -1 // 或者使用缓存的值
					log.Printf("Fast count query timeout: %v", err)
				} else {
					// 其他错误，尝试精确计数
					fallbackCtx, fallbackCancel := context.WithTimeout(rootCtx, CountTimeout)
					defer fallbackCancel()

					countQuery := "SELECT COUNT(*) FROM live_actions WHERE 1=1"
					if err := db.WithContext(fallbackCtx).Raw(countQuery).Scan(&total).Error; err != nil {
						if err == context.DeadlineExceeded {
							total = -1
						} else {
							countErr = err
						}
					}
				}
			} else {
				// 带条件的精确计数
				countQuery := "SELECT COUNT(*) FROM live_actions WHERE 1=1"
				countArgs := []interface{}{}

				if room != "" {
					countQuery += " AND live_room = ?"
					countArgs = append(countArgs, room)
				}
				if typo != "" {
					countQuery += " AND action_type = ?"
					countArgs = append(countArgs, typo)
				}

				ctx, cancel := context.WithTimeout(rootCtx, CountTimeout)
				defer cancel()
				err := db.WithContext(ctx).Raw(countQuery, countArgs...).Scan(&total).Error

				if err == context.DeadlineExceeded {
					// 计数超时，使用估算值
					total = -1
					log.Printf("Count query timeout for room=%s, type=%s", room, typo)
				} else if err != nil {
					countErr = err
				}
			}
		}()

		// 等待所有goroutine完成
		wg.Wait()

		// 错误处理
		if queryErr != nil {
			status := 500
			if queryErr == context.DeadlineExceeded {
				status = 504 // Gateway Timeout
			}
			c.JSON(status, gin.H{
				"error":   "查询失败",
				"message": queryErr.Error(),
				"data":    []ExtendAction{}, // 返回空数组而不是null
			})
			return
		}

		if countErr != nil {
			// 计数失败但数据查询成功，可以返回数据但标记总数未知
			log.Printf("Count error: %v", countErr)
			total = -1 // 标记为未知
		}

		// 计算页数
		pages := int64(-1)
		if total > 0 {
			pages = (total + int64(pageSize) - 1) / int64(pageSize)
		}

		// 返回结果
		c.JSON(200, gin.H{
			"data":  results,
			"page":  pageNum,
			"size":  pageSize,
			"total": total,
			"pages": pages,
		})
	})
	r.GET("/online", func(c *gin.Context) {
		var idStr = c.Query("id")
		var id = int(toInt64(idStr))
		if id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "wrong params",
			})
			return
		}
		var dst []OnlineStatus
		db.Raw("select * from online_statuses where live=?", id).Scan(&dst)
		c.JSON(200, gin.H{
			"data": dst,
		})
	})
	r.GET("/reaction", func(c *gin.Context) {
		var midStr = c.Query("mid")
		var id = int(toInt64(midStr))

		var sizeStr = c.Query("size")
		var pageStr = c.Query("page")
		var page = int(toInt64(pageStr))
		var size = int(toInt64(sizeStr))
		if page == 0 {
			page = 1
		}
		if size == 0 {
			size = 3000
		}

		type ActionItemResponse struct {
			UName      string
			UID        int64
			TargetID   int64
			OID        string
			Text       string
			CreatedAt  time.Time
			Title      string
			Type       string
			TargetName string
			Images     string
			Like       int
			Comments   int
			BV         string
		}

		type Dynamic struct {
			Text      string
			CreatedAt time.Time
			Title     string
			Type      string
			ID        int64
			CreateAt  time.Time
			UName     string
			Images    string
			Like      int
			Comments  int
			BV        string
		}

		if id >= 1 {
			var dst []bili.ActionItem
			clickDb.Raw("select * from action_items where uid = ?  limit ? offset ?", id, size, (size * (page - 1))).Scan(&dst)
			var m0 = make(map[int64][]int64)
			seen := make(map[int64]bool)
			var result []bili.ActionItem
			for _, u := range dst {
				_, ok := seen[u.OID]
				if !ok {
					seen[u.OID] = true
					result = append(result, u)
				}
			}
			dst = []bili.ActionItem{}
			for i := range result {
				dst = append(dst, result[i])
			}
			for i := range dst {
				m0[dst[i].TargetID] = append(m0[dst[i].TargetID], dst[i].OID)
			}
			var m = make(map[int64]Dynamic)

			var pool = pool2.New().WithMaxGoroutines(128)
			var mutex = sync.Mutex{}
			for i := range m0 {
				pool.Go(func() {
					var query = "select u_name,id,title,text,type,create_at,images,like,comments,bv from dynamics where uid = " + strconv.FormatInt(i, 10)
					query = query + " and (id = "
					var array = m0[i]
					for i2 := range array {
						query = query + strconv.FormatInt(array[i2], 10)
						if i2 != len(array)-1 {
							query = query + " or id = "
						}
					}
					query = query + ")"
					var dst0 []Dynamic
					clickDb.Raw(query).Scan(&dst0)
					mutex.Lock()
					for i2 := range dst0 {
						m[dst0[i2].ID] = dst0[i2]
					}
					mutex.Unlock()
				})
			}
			pool.Wait()

			var response []ActionItemResponse
			for _, item := range dst {
				var id = strconv.FormatInt(item.OID, 10)
				response = append(response, ActionItemResponse{
					UName:      item.UName,
					UID:        item.UID,
					TargetID:   item.TargetID,
					OID:        id,
					Text:       m[item.OID].Text,
					CreatedAt:  m[item.OID].CreateAt,
					Title:      m[item.OID].Title,
					Type:       m[item.OID].Type,
					TargetName: m[item.OID].UName,
					Like:       m[item.OID].Like,
					Images:     m[item.OID].Images,
					Comments:   m[item.OID].Comments,
					BV:         m[item.OID].BV,
				})
			}

			sort.Slice(response, func(i, j int) bool {
				return time.Since(response[i].CreatedAt).Seconds() < time.Since(response[j].CreatedAt).Seconds()
			})
			c.JSON(200, gin.H{
				"data": response,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "wrong params",
			})
		}
	})
	r.GET("/pk", func(c *gin.Context) {
		var guild = toInt64(c.Query("guild"))
		var month = toInt(c.DefaultQuery("month", time.Now().Month().String()))
		type Scanned struct {
			UserName       string
			Money          float64
			Hours          float64
			BoxDiff        float64
			SuperChatMoney float64
			Guard          string
			Fans           int
			UserID         int64
			GuardLiveMoney float64 //直播捕获到的上舰事件的金额，由于存在自动续舰的情况，在前端显示的总流水应去除捕获到的上舰金额，再加上当月或当前舰长列表里计算所得金额
			Gift           float64
			Guild          string
		}
		if month == 0 {
			month = int((time.Now().Month()))
		}
		var dst []Scanned
		if guild == 0 {
			db.Raw(fmt.Sprintf(`
SELECT
    lives.user_name,
        lives.user_id,
    SUM(money) AS money,
    SUM(28800 +lives.end_at - lives.start_at)/3600 AS hours,
    sum(lives.box_diff) as box_diff,
    sum(lives.super_chat_money) as super_chat_money,
    sum(lives.guard_money) as guard_live_money,
    sum(lives.money-lives.super_chat_money-guard_money) as gift
FROM
    lives
WHERE
    MONTH(lives.created_at) = %d
    AND end_at != 0
GROUP BY
    lives.user_name
ORDER BY money DESC
LIMIT 400;

`, month)).Scan(&dst)
		} else {
			db.Raw(fmt.Sprintf(`
SELECT 
    lives.user_name, 
    lives.user_id,
    SUM(money) AS money,
    SUM(28800 +lives.end_at - lives.start_at)/3600 AS hours,
    sum(lives.box_diff) as box_diff,
    sum(lives.super_chat_money) as super_chat_money,
    sum(lives.guard_money) as guard_live_money,
    sum(lives.money-lives.super_chat_money-guard_money) as gift

FROM 
    lives
JOIN 
    guild_infos ON guild_infos.uid = lives.user_id
WHERE 
    guild_infos.guild_id = %d
    AND MONTH(lives.created_at) = %d
    AND end_at != 0
GROUP BY 
    lives.user_name
ORDER BY money DESC;


`, guild, month)).Scan(&dst)
		}
		for i, _ := range dst {
			dst[i].Money = math.Round(dst[i].Money*100) / 100
			dst[i].Hours = math.Round(dst[i].Hours*100) / 100
			dst[i].BoxDiff = math.Round(dst[i].BoxDiff*100) / 100
			dst[i].SuperChatMoney = math.Round(dst[i].SuperChatMoney*100) / 100
			dst[i].Gift = math.Round(dst[i].Gift*100) / 100
		}

		var currentMonth = int(time.Now().Month())
		var pool = pool2.New().WithMaxGoroutines(6)
		for j, scanned := range dst {
			var id = scanned.UserID
			pool.Go(func() {
				var obj AreaLiver
				if month == currentMonth {
					db.Raw("select guard,fans from area_livers where uid = ? order by id desc limit 1", id).Scan(&obj)
				} else {
					db.Raw("select guard,fans from area_livers where uid = ? and MONTH(updated_at) = ? order by id desc limit 1", id, month).Scan(&obj)
				}
				db.Raw("select guild_name from guild_infos where uid = ?", id).Scan(&dst[j].Guild)
				dst[j].Fans = obj.Fans
				dst[j].Guard = obj.Guard
			})
		}
		pool.Wait()

		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})

	})
	r.GET("/api/geo", func(c *gin.Context) {
		type Location struct {
			UID      string
			Location string
		}
		var dst []Location
		db.Raw("select * from locations where location is not null ").Scan(&dst)
		var m = make(map[string]int)
		for i := range dst {
			dst[i].Location = strings.Replace(dst[i].Location, "中国", "", 99)
			_, ok := m[dst[i].Location]
			if dst[i].Location == "" {
				continue
			}
			if !ok {
				m[dst[i].Location] = 1
			} else {
				m[dst[i].Location] = m[dst[i].Location] + 1
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"data": m,
		})
	})
	type ProvinceResponse struct {
		UID  int64
		Name string
		Fans int
	}
	var cacheProvince = make(map[string][]ProvinceResponse)

	r.GET("/api/geo/province", func(c *gin.Context) {
		province := c.DefaultQuery("name", "")
		if province == "香港" || province == "澳门" || province == "台湾" {
			province = "中国" + province
		}

		if data, ok := cacheProvince[province]; ok {
			c.JSON(http.StatusOK, gin.H{
				"data": data,
			})
			return
		}

		var dst []ProvinceResponse
		db.Raw(`
        SELECT u.user_id AS uid, MAX(u.fans) AS fans, u.name
        FROM users u
        JOIN locations l ON u.user_id = l.uid
        WHERE l.location = ?
        GROUP BY u.user_id
        ORDER BY fans DESC
        LIMIT 1000;
    `, province).Scan(&dst)

		cacheProvince[province] = dst

		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})
	})
	r.GET("/minute", func(c *gin.Context) {
		var id = c.Query("id")
		type LiveActionCount struct {
			MinuteTime  string `gorm:"column:minute_time"`
			RecordCount int    `gorm:"column:record_count"`
		}

		var results []LiveActionCount

		db.Raw(`
    SELECT 
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') AS minute_time, 
        COUNT(*) AS record_count
    FROM live_actions
    WHERE live = ?
    GROUP BY minute_time
    ORDER BY minute_time;
`, id).Scan(&results)
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    results,
		})
	})

	r.GET("/living", func(c *gin.Context) {
		tasks := man.GetAllTasks(false)
		var dst []AreaLiver
		var pool = pool2.New().WithMaxGoroutines(32)
		for _, task := range tasks {
			pool.Go(func() {
				var d AreaLiver
				db.Raw(`select fans,u_name from area_livers where room = ? limit 1`, task).Scan(&d)
				dst = append(dst, d)
			})
		}
		pool.Wait()
		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})
	})
	r.Run(":" + strconv.Itoa(int(config.Port)))
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func TTLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		c.Next()
		duration := time.Since(startTime)
		originalBody := writer.body.Bytes()

		var finalBody []byte

		contentType := writer.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var data map[string]interface{}
			err := sonic.Unmarshal(originalBody, &data)

			if err == nil {
				// 添加 ttl 字段，单位为毫秒
				data["ttl"] = duration.Milliseconds()

				// 重新序列化为 JSON
				newBody, marshalErr := sonic.Marshal(data)
				if marshalErr == nil {
					finalBody = newBody
				} else {
					// 如果重新序列化失败，则保留原始响应
					fmt.Println("Error re-marshalling JSON:", marshalErr)
					finalBody = originalBody
				}
			} else {
				// 如果解析失败（比如响应体是JSON数组或无效JSON），则保留原始响应
				finalBody = originalBody
			}
		} else {
			// 如果不是 JSON，直接使用原始响应
			finalBody = originalBody
		}

		// 6. 更新 Content-Length 并将最终响应写回
		c.Header("Content-Length", strconv.Itoa(len(finalBody)))
		// 使用原始的 ResponseWriter 将内容写入网络连接
		writer.ResponseWriter.Write(finalBody)
	}
}
