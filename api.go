package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"

	bili "github.com/114514ns/BiliClient"
	"github.com/bytedance/sonic"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	pool2 "github.com/sourcegraph/conc/pool"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var worker = NewWorker(1)

var tasks = make(map[string]Task)

var taskMutex sync.Mutex

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
	//r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gzip.Gzip(gzip.BestCompression))
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

		if strings.Contains(c.Request.Header.Get("X-Provider"), "google.com") {
			limit = 5000
		}
		offset := (page - 1) * limit

		orderStr := c.DefaultQuery("order", "ascend")

		query := db.Model(&LiveAction{}).Where("live = ? ", id)

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
		query = db.Model(&LiveAction{}).Where("live = ? ", id)

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
			c.JSON(http.StatusOK, gin.H{
				"Room": lives,
			})
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
					"Requests":        totalRequests,
					"Tasks":           guardWorker.QueueLen(),
					"LaunchedAt":      launchTime.Format(time.DateTime),
					"TotalMessages":   TotalMessage(),
					"Message1":        t1,
					"MessageHour":     t2,
					"MessageDaily":    t3,
					"HTTPBytes":       httpBytes,
					"WSBytes":         websocketBytes + bytes,
					"Nodes":           man.Nodes,
					"Rooms":           m,
					"LastInsert":      lastInsert,
					"LastInsertCount": lastInsertCount,
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

		/*
			var filter = make([]FrontAreaLiver, 0)
			for i := range cachedLivers {
				var item = cachedLivers[i]

				if item.Fans > 1000 {
					if !Has(config.BlackAreaLiver, item.UID) {
						filter = append(filter, item)
					}
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"list": filter,
			})

		*/

		c.Redirect(http.StatusTemporaryRedirect, GetFile("/Microsoft365/static/areaLivers.json"))

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
			time.Sleep(7 * time.Second)
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
			//s = "and TO_DAYS(NOW()) -TO_DAYS(updated_at) <7"
		}
		if liver != "" {
			query.Where("liver_id = ? "+s, liver)
			countQuery.Where("liver_id = ? "+s, liver)
		}

		var count int64
		query.Count(&count)
		query.Offset(sizeInt * (pageInt - 1)).Limit(sizeInt).Order("level desc").Find(&result)

		sort.Slice(result, func(i, j int) bool {
			return result[i].Level > result[j].Level
		})

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
		for i := range dst {
			var item = dst[i]
			if time.Now().After(item.UpdatedAt.Add(time.Hour * 720)) {
				dst[i].Type = 0
			}
		}
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
		wg.Add(8)
		go func() {
			db.Raw("select count(*) from live_actions where from_id = ? and action_type = 1", uid).Scan(&totalMessage)
			wg.Done()
		}()
		go func() {
			db.Raw("SELECT COALESCE(SUM(gift_price), 0)  FROM live_actions WHERE from_id = ? and (action_type = 2 or action_type = 4)", uid).Scan(&totalMoney)
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
			db.Raw("SELECT live_room,COUNT(live_room) as count,user_id as liver_id,user_name as liver  FROM live_actions,lives where from_id = ? and live_actions.live = lives.id and action_type = 1  GROUP BY live_room limit 100", uid).Scan(&rooms)
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
		var guardSum = 0.0
		go func() {
			var livers []int64
			db.Raw("select liver_id from fans_clubs where uid = ? and level >= 21", uid).Scan(&livers)

			var areas []AreaLiver
			db.Raw("SELECT id,updated_at FROM area_livers WHERE  uid IN ?", livers).Scan(&areas)

			if len(livers) == 0 {
				areas = []AreaLiver{}
				return
			}
			db.Raw(`
SELECT *
FROM (
  SELECT al.*,
         ROW_NUMBER() OVER (
           PARTITION BY al.uid, DATE_FORMAT(al.updated_at, '%Y-%m')
           ORDER BY al.updated_at DESC, al.id DESC
         ) rn
  FROM area_livers al
  WHERE al.uid IN (?)
) t
WHERE t.rn = 1
ORDER BY t.updated_at DESC, t.id DESC
`, livers).Scan(&areas)

			var m = [...]float64{0, 19998, 1998, 138}
			for _, area := range areas {
				var dg []DBGuard
				json.Unmarshal([]byte(area.GuardList), &dg)
				for _, guard := range dg {
					if guard.UID == uid {
						guardSum += m[guard.Type]
					}
				}
			}

			wg.Done()
		}()

		wg.Wait()

		context.JSON(http.StatusOK, gin.H{
			"GiftMoney":    totalMoney,
			"GuardMoney":   guardSum,
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

		var db0 *gorm.DB
		if showEnter != "" {
			tableName = "enter_actions"
			db0 = clickDb
		} else {
			if order == "timeDesc" {
				queryOrder = "order by live_actions.id desc"
			}
			if order == "money" {
				queryOrder = "order by gift_price desc"
			}
			db0 = db
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

		if context.Request.Header.Get("X-Provider") == "google.com" {
			pageSize = 5000
		}
		offset := (page - 1) * pageSize
		var dst []CustomLiveAction
		if showEnter == "" {
			err = db0.Raw(
				fmt.Sprintf("select *,%s.id,%s.created_at from %s,lives where from_id = ? and lives.id = %s.live %s %s limit ? offset ?", tableName, tableName, tableName, tableName, typeQuery, queryOrder),
				uid, pageSize, offset,
			).Scan(&dst).Error
		} else {
			err = db0.Raw("select * from enter_actions where from_id = ? limit ? offset ?", uid, pageSize, (page-1)*pageSize).Scan(&dst).Error

			lo.UniqBy(dst, func(item CustomLiveAction) int {
				return item.LiveRoom
			})
		}

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"message": "query failed",
				"error":   err.Error(),
			})
			return
		}
		if showEnter != "" {
			db0.Raw("select count(*) from enter_action where from_id = ?", uid).Scan(&total)
		} else {
			if room != "" {
				db0.Raw("select count(*) from live_actions where from_id = ? and live_room = ? and  (select lives.id from lives where live_actions.live = lives.id ) != 0 "+typeQuery,
					uid, room).Scan(&total)
			} else {
				db0.Raw("select count(*) from live_actions where from_id = ? and  (select lives.id from lives where live_actions.live = lives.id ) != 0 "+typeQuery, uid).Scan(&total)
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
		wg.Add(5)
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

		var sum = 0.0
		go func() {

			var dst []Live
			db.Raw("SELECT * FROM lives WHERE user_id = ? and created_at >= (NOW() + INTERVAL 8 HOUR) - INTERVAL 720 HOUR", mid).Scan(&dst)
			var wgInner sync.WaitGroup
			wgInner.Add(len(dst))
			var m1 sync.Mutex
			for _, live := range dst {
				go func() {
					var t = live.Money
					var sub = 0.0
					db.Raw("select sum(gift_price) from live_actions where live = ? and action_type = 3", live.ID).Scan(&sub)
					m1.Lock()
					sum = t - sub + sum
					m1.Unlock()
					wgInner.Done()
				}()
			}
			wgInner.Wait()
			wg.Done()
		}()
		var charge = ""
		go func() {
			db.Raw("select charge from area_livers where uid = ? order by id desc limit 1", mid).Scan(&charge)
			wg.Done()
		}()
		wg.Wait()
		var area = ""
		db.Raw("select guard from area_livers where uid = ? order by id desc limit 1", mid).Scan(&area)
		var index = 0
		for _, i := range strings.Split(area, ",") {
			if index == 0 {
				sum = sum + 19998.0*float64(toInt(i))
			}
			if index == 1 {
				sum = sum + 1998.0*float64(toInt(i))
			}
			if index == 2 {
				sum = sum + 168.0*float64(toInt(i))
			}
			index++
		}

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
			"Amount":  sum,
			"Charge":  charge,
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
			Amount       float64
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
					var t0 []LiveAction
					e := db.WithContext(ctx).Raw("select id from live_actions where from_id = ? and action_type = 3 limit 50", id).Scan(&t0)

					guardCount = len(t0)

					var msgCount = 0
					e1 := db.WithContext(ctx).Raw("select id from live_actions where from_id = ? and action_type = 1 limit 50", id).Scan(&t0)

					msgCount = len(t0)

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

				})
			}
			pool.Wait()
			sort.Slice(dst, func(i, j int) bool {
				return dst[i].Level > dst[j].Level
			})
		} else {
			var d1 = dst
			var room = 0
			db.Raw("select room from area_livers where uid = ?", d1[0].LiverID).Scan(&room)

			var createdAt time.Time
			db.Raw("select updated_at from area_livers where id = ?", idStr).Scan(&createdAt)
			sort.Slice(dst, func(i, j int) bool {
				return dst[i].Level > dst[j].Level
			})
			if len(dst) > 200 {
				d1 = dst[0:200]
			}
			if len(d1) == 0 {
				return
			}
			uids := make([]int64, 0, len(d1))
			for _, item := range d1 {
				uids = append(uids, item.UID)
			}
			startOfMonth := time.Date(createdAt.Year(), createdAt.Month(), 1, 0, 0, 0, 0, createdAt.Location())
			endOfMonth := startOfMonth.AddDate(0, 1, 0)
			type AggregateResult struct {
				FromID int64   `gorm:"column:from_id"`
				Total  float64 `gorm:"column:total"`
			}
			var results []AggregateResult
			err := db.Table("live_actions").
				Select("from_id, sum(gift_price) as total").
				Where("from_id IN ?", uids).
				Where("live_room = ?", room).
				Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth).
				Where("action_type = 2 or action_type= 4").
				Group("from_id").
				Scan(&results).Error
			if err != nil {
				return
			}
			amountMap := make(map[int64]float64, len(results))
			for _, r := range results {
				amountMap[r.FromID] = r.Total
			}
			for i := range d1 {
				d1[i].Amount = amountMap[d1[i].UID]
				if d1[i].Type == 1 {
					d1[i].Amount += 19998
				}
				if d1[i].Type == 2 {
					d1[i].Amount += 1998
				}
				if d1[i].Type == 3 {
					d1[i].Amount += 168
				}
			}
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

	r.GET("/dynamics", func(c *gin.Context) {
		var midStr = c.Query("mid")
		var mid = toInt64(midStr)

		sql := `select * from dynamics where uid = ?;
`
		var dst []Dynamic
		db.Raw(sql, mid).Scan(&dst)
		type ExtendDynamic struct {
			Dynamic
			IDStr string
		}
		var d []ExtendDynamic
		copier.Copy(&d, &dst)
		var m = make(map[int64]bool)
		for i := range d {
			d[i].IDStr = toString(dst[i].ID)
			if d[i].ForwardFrom != 0 {
				m[d[i].ForwardFrom] = true
			}
		}
		var pool = pool2.New().WithMaxGoroutines(12)
		var mutex sync.Mutex
		for i := range m {
			var uid = i
			pool.Go(func() {
				var d2 ExtendDynamic
				db.Raw(`select * from dynamics where id = ? limit 1`, uid).Scan(&d2)

				mutex.Lock()

				if mid != d2.UID {
					d = append(d, d2)
				}

				mutex.Unlock()

			})
		}
		c.JSON(http.StatusOK, gin.H{
			"data": d,
		})
	})
	r.GET("/dynamics/count", func(c *gin.Context) {
		var midStr = c.Query("mid")
		var mid = toInt64(midStr)

		if mid <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid mid",
			})
			return
		}

		var count = 0
		db.Raw("select count(*) from dynamics where uid = ? ", mid).Scan(&count)
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})

	r.POST("/comments/delete", func(c *gin.Context) {
		_, e := uuid.Parse(c.DefaultPostForm("session", ""))
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid session",
			})
			return
		}

		var id = toInt64(c.DefaultPostForm("id", ""))
		if id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid id",
			})
			return
		}
		var s1 = ""
		db.Raw("select session from comments where id = ?", id).Scan(&s1)
		if s1 != c.DefaultPostForm("session", "") {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid session",
			})
			return
		}
		db.Delete(&Comment{}, "id = ?", id)
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	r.POST("/comments/send", func(c *gin.Context) {
		_, e := uuid.Parse(c.DefaultPostForm("session", ""))
		text := c.DefaultPostForm("text", "")

		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid session",
			})
			return
		}

		if len(text) <= 0 || len(text) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid text",
			})
			return
		}

		db.Save(&Comment{
			Text:      text,
			CreatedAt: time.Now(),
			Session:   c.PostForm("session"),
		})

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})

	})

	r.GET("/comments/list", func(c *gin.Context) {
		var page = c.DefaultQuery("page", "1")
		var size = c.DefaultQuery("size", "10")

		var session = c.DefaultQuery("session", "")
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			pageInt = 1
		}
		sizeInt, err := strconv.Atoi(size)
		if err != nil || sizeInt < 1 {
			sizeInt = 10
		}
		if sizeInt > 100 {
			sizeInt = 100
		}
		offset := (pageInt - 1) * sizeInt
		var comments []Comment
		result := db.Raw("select * from comments order by created_at desc limit ? offset ?",
			sizeInt, offset).Scan(&comments)
		if result.Error != nil {
			c.JSON(500, gin.H{
				"msg":     "error",
				"message": result.Error.Error(),
			})
			return
		}
		var total int64
		db.Model(&Comment{}).Count(&total)

		for i := range comments {
			if comments[i].Session != session {

				h := md5.Sum([]byte(comments[i].Session))
				comments[i].DisplayName = strings.ToUpper(hex.EncodeToString(h[:])[0:6])
				comments[i].Session = ""
			} else {
				comments[i].DisplayName = "You"
			}
		}

		c.JSON(200, gin.H{
			"data":        comments,
			"page":        pageInt,
			"size":        sizeInt,
			"total":       total,
			"total_pages": (total + int64(sizeInt) - 1) / int64(sizeInt),
		})
	})

	r.GET("/hot", func(c *gin.Context) {
		type Select struct {
			ID       int
			UserName string
			UserID   int64
		}
		var dst []Select
		db.Raw("select lives.user_id,user_name,lives.id from lives,area_livers where lives.user_id = area_livers.uid  and end_at = 0 and lives.created_at >= (NOW() + INTERVAL 8 HOUR) - INTERVAL 480 MINUTE group by lives.user_id order by fans desc limit 50").Scan(&dst)
		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})
	})

	r.GET("/stress", func(c *gin.Context) {

		c.Redirect(http.StatusMovedPermanently, "http://127.0.0.1:8081"+fmt.Sprintf("/live/%d/?page=1&limit=10&order=undefined&mid=0&type=", rand.Int()%300000+1))
	})

	r.GET("/export", func(c *gin.Context) {
		var typo = c.DefaultQuery("type", "user")
		var idStr = c.DefaultQuery("id", "")

		var id = toInt64(idStr)

		if id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid id",
			})
			return
		}
		f := excelize.NewFile()
		defer f.Close()

		var sheetName = "Data"

		index, _ := f.NewSheet(sheetName)

		if typo == "user" {

			type ExtendLiveAction struct {
				LiveAction
				UserName string
			}
			var dst []ExtendLiveAction
			db.Raw("select live_actions.medal_level,live_actions.medal_name,live_actions.from_id,live_actions.action_type,live_actions.extra,live_actions.created_at,live_actions.gift_price,lives.user_name,live_actions.from_name,live_actions.gift_name from live_actions,lives where from_id = ? and live_actions.live = lives.id ", id).Scan(&dst)

			var row = 1
			f.SetCellValue(sheetName, "A1", "粉丝牌")
			f.SetCellValue(sheetName, "b1", "用户名")
			f.SetCellValue(sheetName, "C1", "时间")
			f.SetCellValue(sheetName, "D1", "主播")
			f.SetCellValue(sheetName, "E1", "弹幕/礼物")
			f.SetCellValue(sheetName, "F1", "礼物金额")
			f.SetCellValue(sheetName, "G1", "类型")

			row++
			f.SetColWidth(sheetName, "A", "C", 20)
			f.SetColWidth(sheetName, "d", "d", 30)
			f.SetColWidth(sheetName, "e", "e", 40)
			for i := range dst {
				var item = dst[i]
				var price0 = item.GiftPrice

				var price = ""

				if price0.Valid {
					price = strconv.FormatFloat(price0.Float64, 'f', -1, 64)
				}

				var extra = item.Extra

				if extra == "" {
					extra = item.GiftName
				}

				var t = ""

				if item.ActionType == Message {
					t = "弹幕"
				}
				if item.ActionType == Gift {
					t = "礼物"
				}
				if item.ActionType == SuperChat {
					t = "醒目留言"
				}
				if item.ActionType == Guard {
					t = "上舰"
				}
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), (item.MedalName)+"  LV"+toString(int64(item.MedalLevel)))
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.FromName)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.CreatedAt.Format(time.DateTime))
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.UserName)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), extra)
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), price)
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), t)

				style, _ := f.NewStyle(&excelize.Style{
					Font: &excelize.Font{
						Color: getColor(int(item.MedalLevel)),
						Bold:  true,
					},
					Alignment: &excelize.Alignment{
						Horizontal: "center",
					},
				})

				f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), style)

				if price0.Valid {
					style, _ := f.NewStyle(&excelize.Style{
						Font: &excelize.Font{
							Color: "#C55A11",
							Bold:  true,
						},
						Alignment: &excelize.Alignment{
							Horizontal: "center",
						},
					})

					f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), style)
				}

				row++
			}
			f.SetActiveSheet(index)
			rows, _ := f.GetRows(sheetName)

			style, _ := f.NewStyle(&excelize.Style{
				Alignment: &excelize.Alignment{
					Horizontal: "center",
				},
			})

			lastRow := len(rows)
			e := f.SetPanes(sheetName, &excelize.Panes{
				Freeze: true,
				Split:  false,
				XSplit: 1, // 冻结列数
				YSplit: 0,
			})
			if e != nil {
				log.Println(e)
				return
			}
			f.SetCellStyle(sheetName, "B1", fmt.Sprintf("Z%d", lastRow), style)

			buf, _ := f.WriteToBuffer()
			c.Data(
				200,
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				buf.Bytes(),
			)
		}
	})
	r.POST("/black/add", func(c *gin.Context) {
		var midStr = c.DefaultQuery("mid", "")
		var mid = toInt64(midStr)

		if mid <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}

		config.BlackAreaLiver = append(config.BlackAreaLiver, mid)
		SaveConfig()

		c.JSON(http.StatusOK, gin.H{
			"msg": "success",
		})
	})

	r.GET("/crc", func(c *gin.Context) {
		var hash = c.Query("hash")
		type Response struct {
			UName string
			UID   int64
		}
		var dst []Response
		clickDb.Raw("select uid,u_name  from user_mappings where hash = ?", hash).Scan(&dst)

		c.JSON(http.StatusOK, gin.H{
			"data": dst,
		})
	})

	r.GET("/crc/bv", func(c *gin.Context) {

		type I struct {
			Messages []struct {
				Text string `xml:",chardata"`
				P    string `xml:"p,attr"`
			} `xml:"d"`
		}

		var bv = c.Query("bv")
		res, _ := queryClient.R().Get("https://api.bilibili.com/x/web-interface/view?bvid=" + bv)

		var obj map[string]interface{}

		sonic.Unmarshal(res.Body(), &obj)

		if getInt(obj, "code") != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "video doesn't exists",
			})
			return
		}
		var cid = getInt64(obj, "data.cid")

		if cid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "unknown error",
			})
			return
		}

		xml0, _ := queryClient.R().Get("https://comment.bilibili.com/" + toString(cid) + ".xml")

		decompress, _ := DeflateDecompress(xml0.Body())

		type UserMapping struct {
			UID   int64
			UName string
		}
		type ResponseItem struct {
			Text      string
			Sender    []UserMapping
			Offset    int
			CreatedAt time.Time
		}

		var dm I
		xml.Unmarshal(decompress, &dm)

		var pool = pool2.New().WithMaxGoroutines(128)

		var results []ResponseItem

		var lck sync.Mutex

		for i := range dm.Messages {
			var item = dm.Messages[i]
			pool.Go(func() {
				var hash = strings.Split(item.P, ",")[6]

				re, _ := localClient.R().Get("http://192.168.31.178:1145/crc?hash=" + hash)

				type RPCObject struct {
					Data []UserMapping
				}
				var o RPCObject
				sonic.Unmarshal(re.Body(), &o)
				lck.Lock()
				results = append(results, ResponseItem{
					Text:      item.Text,
					Sender:    o.Data,
					CreatedAt: time.Unix(toInt64(strings.Split(item.P, ",")[4]), 0),
					Offset:    int((toFloat64(strings.Split(item.P, ",")[0]))),
				})
				lck.Unlock()

			})
		}

		pool.Wait()

		c.JSON(http.StatusOK, gin.H{
			"item": results,
		})

	})
	r.GET("/playback", func(c *gin.Context) {
		var id = toInt(c.Query("id"))

		if id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}

		var live Live
		db.Raw("select * from lives where id = ?", id).Scan(&live)

		if live.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var f1 []File
		var f2 []File

		var v []Video
		var vd0 []Video
		var vd []Video
		go func() {

			db.Raw("select * from playbacks where uid = ?", live.UserID).Scan(&v)
			for i := range v {
				var v0 = v[i]
				var start = time.Unix(live.StartAt, 0).Add(time.Hour * -8)
				v0.Title = strings.TrimSpace(v0.Title)
				var sp = strings.Split(v0.Title, "年")
				if len(sp) >= 2 {
					var o = sp[len(sp)-2]

					var year = o[len(o)-4:]

					var o0 = sp[len(sp)-1]

					o0 = strings.Replace(o0, "月", ",", 99)
					o0 = strings.Replace(o0, "日", ",", 99)
					o0 = strings.Replace(o0, "点场", ",", 99)

					var month = ""
					var day = ""
					var hour = ""
					for j, s := range strings.Split(o0, ",") {
						if j == 0 {
							month = s
						}
						if j == 1 {
							day = s
						}
						if j == 2 {
							hour = s
						}
					}

					if year == toString(int64(start.Year())) {
						if month == toString(int64(start.Month())) {
							if day == toString(int64(start.Day())) {
								if hour == toString(int64(start.Hour())) {
									vd = append(vd, v0)
								}
							}
						}
					}

				}
			}

			for i := range vd {
				var u = "https://api.bilibili.com/x/web-interface/view?bvid=" + vd[i].BV + "&isGaiaAvoided=true"
				var u0, _ = url.Parse(u)
				query, _ := wbi.SignQuery(u0.Query(), time.Now())
				res, _ := queryClient.R().SetDebug(true).EnableGenerateCurlOnDebug().SetHeader("Cookie", fmt.Sprintf("buvid3=%sinfoc", strings.ToUpper(uuid.New().String()))).SetHeader("User-Agent", USER_AGENT).Get("https://api.bilibili.com/x/web-interface/view?" + query.Encode())
				curlCmdStr := res.Request.GenerateCurlCommand()
				fmt.Println(curlCmdStr)
				var obj map[string]interface{}
				sonic.Unmarshal(res.Body(), &obj)
				for _, item := range getArray(obj, "data.pages") {
					var v0 = vd[i]
					v0.BV = v0.BV + "-" + toString(getInt64(item, "cid"))

					vd0 = append(vd0, v0)
				}

			}

			wg.Done()

		}()

		wg.Wait()

		var wg1 sync.WaitGroup
		wg1.Add(2)
		if len(vd) == 0 {
			go func() {
				f1 = ListFile(fmt.Sprintf("/Microsoft365/%s/%s", live.UserName, strings.Replace(time.Unix(live.StartAt-3600*8, 0).Format(time.DateTime), ":", "-", 1145)))
				wg1.Done()
			}()

			go func() {
				f2 = ListFile(fmt.Sprintf("/139/%s/%s", live.UserName, strings.Replace(time.Unix(live.StartAt-3600*8, 0).Format(time.DateTime), ":", "-", 1145)))
				wg1.Done()
			}()
		} else {
			wg1.Done()
			wg1.Done()
		}

		wg1.Wait()

		var f3 = f2

		if f3 == nil {
			f3 = f1
		}

		c.JSON(http.StatusOK, gin.H{
			"files":    f3,
			"archives": vd0,
		})

	})

	r.POST("/redirect", func(c *gin.Context) {
		type Req struct {
			Data string `json:"data"`
		}
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"err": err.Error()})
			return
		}

		decoded, e := base64.StdEncoding.DecodeString(req.Data)
		if e != nil {
			c.JSON(400, gin.H{"err": "invalid base64"})
			return
		}
		get, _ := client.R().SetDoNotParseResponse(true).Get(string(decoded))
		c.JSON(http.StatusOK, gin.H{
			"url": get.RawResponse.Request.URL.String(),
		})
	})

	r.GET("/bv/view", func(c *gin.Context) {
		//var start = time.Now()
		var bv = c.Query("bv")
		var cidStr = c.Query("cid")
		if bv == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invaild params",
			})
			return
		}
		cid := 0
		if toInt64(cidStr) > 0 {
			cid = int(toInt64(cidStr))
		} else {
			r0, _ := queryClient.R().Post("https://api.live.bilibili.com//xlive/open-platform/v1/inner/getArchiveInfo?bv_id=" + bv)
			var o0 map[string]interface{}
			sonic.Unmarshal(r0.Body(), &o0)
			cid = getInt(o0, "data.cid")
		}

		//log.Println(time.Since(start))
		var url0 = fmt.Sprintf("https://api.bilibili.com/x/player/wbi/playurl?bvid=%s&cid=%d&gaia_source=view-card&isGaiaAvoided=true&qn=32&fnval=4048&try_look=1", bv, cid)
		parse, _ := url.Parse(url0)
		query, _ := wbi.SignQuery(parse.Query(), time.Now())
		res, _ := queryClient.R().Get("https://api.bilibili.com/x/player/wbi/playurl?" + query.Encode())
		//log.Println(time.Since(start))
		var o map[string]interface{}
		sonic.Unmarshal(res.Body(), &o)
		//log.Println(time.Since(start))
		c.JSON(http.StatusOK, o)

		return
	})

	r.POST("/transcribe/create", func(c *gin.Context) {
		/*
			var idStr = c.Query("id")
			var id = toInt(idStr)
			if id <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "bad params",
				})
				return
			}
			var str = "uname,text,time,money"
			var

		*/
	})

	//创建模糊查询粉丝牌任务

	r.GET("/online/chart", func(c *gin.Context) {

	})

	type User struct {
		UID   int64
		UName string
	}

	r.GET("/liver/follow", func(c *gin.Context) {
		var midStr = c.Query("mid")
		var dst []User
		if toInt64(midStr) <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
		}

		var ids []int64
		db.Raw("select target_id from relations where uid = ?", toInt64(midStr)).Scan(&ids)
		for _, int64s := range chunkSlice(ids, 40) {
			maps := biliClient.BatchGetFace(int64s)
			for _, faceMap := range maps {
				dst = append(dst, User{
					UID:   faceMap.UID,
					UName: faceMap.UName,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"list": dst,
		})

	})
	r.GET("/liver/followed", func(c *gin.Context) {
		var midStr = c.Query("mid")
		var dst []User
		if toInt64(midStr) <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
		}

		var ids []int64
		db.Raw("select uid from relations where target_id = ?", toInt64(midStr)).Scan(&ids)
		for _, int64s := range chunkSlice(ids, 40) {
			maps := biliClient.BatchGetFace(int64s)
			for _, faceMap := range maps {
				dst = append(dst, User{
					UID:   faceMap.UID,
					UName: faceMap.UName,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"list": dst,
		})
	})

	r.GET("/history", func(c *gin.Context) {
		var roomStr = c.Query("room")
		var room = toInt(roomStr)
		if room <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}
		var dst []LiveAction = make([]LiveAction, 0)
		var id int
		db.Raw("select id from lives where room_id = ?", room).Scan(&id)
		if id > 0 {
			db.Raw("select * from live_actions where live = ? order by id desc ", id).Scan(&dst)
		}
		c.JSON(http.StatusOK, gin.H{
			"list": dst,
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

		c.Writer.Header().Set("Cache-Control", " public, max-age=0, stale-while-revalidate=30")
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
