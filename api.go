package main

import (
	"embed"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"log"
	"net/http"
	url2 "net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var videoCache = []Video{}
var worker = NewWorker()

func removeDuplicates(input []string) []string {
	// 创建一个空的 map，用于记录已经存在的字符串
	seen := make(map[string]bool)

	// 存储去重后的结果
	var result []string

	// 遍历输入数组
	for _, str := range input {
		// 如果字符串没有出现过，则添加到结果并在 map 中标记为已存在
		if !seen[str] {
			result = append(result, str)
			seen[str] = true
		}
	}

	return result
}

type FrontAreaLiver struct {
	AreaLiver
	LastActive time.Time
	DailyDiff  int
	Verify     string
	Bio        string
}

//go:embed Page/dist
var distFS embed.FS

func InitHTTP() {
	r := gin.Default()
	r.UseH2C = true

	r.Use(CORSMiddleware())
	//r.Static("/page", "./Page/dist/")
	//r.Static("/assets", "./Page/dist/assets")

	if ENV == "BUILD" {
		r.Use(static.Serve("/", static.EmbedFolder(distFS, "Page/dist")))
	} else {
		r.Use(static.Serve("/", static.LocalFile("./Page/dist", false)))
	}
	r.Use(gzip.Gzip(gzip.DefaultCompression))
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
		}
		for _, status := range lives {
			var cpy = Status{}
			copier.Copy(&cpy, &status)
			cpy.GuardList = []Watcher{}
			cpy.OnlineWatcher = []Watcher{}
			array = append(array, cpy)
		}
		c.JSON(http.StatusOK, gin.H{
			"lives": array,
		})
	})
	r.GET("/monitor/:id", func(c *gin.Context) {
		id := c.Param("id")
		if config.Mode == "Master" {
			add, ok := man.GetNodeByTask(id)
			if !ok {
				c.JSON(http.StatusOK, gin.H{
					"message": "<UNK>",
				})
			}
			if !man.isSelf(add) {
				r, _ := client.R().Get(add + "/monitor/" + id)
				c.String(r.StatusCode(), r.String())
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"live": lives[id],
		})
	})
	r.GET("/liver", func(c *gin.Context) {
		key := c.DefaultQuery("key", "1")
		var result0 = make([]Live, 0)
		var result = make([]string, 0)
		db.Model(&Live{}).Where("user_name like '%" + key + "%'").Find(&result0)
		for i := range result0 {
			result = append(result, result0[i].UserName)
		}
		result = removeDuplicates(result)
		c.JSON(http.StatusOK, gin.H{
			"result": result,
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
	r.GET("/sendMsg", func(c *gin.Context) {
		room := c.DefaultQuery("room", "-1")
		msg := c.DefaultQuery("msg", "-@")
		if room == "-1" || msg == "-@" {
			c.JSON(http.StatusOK, gin.H{
				"msg": "missing params",
			})
		} else {
			i, _ := strconv.Atoi(room)
			SendMessage(msg, i, func(s string) {
				c.JSON(http.StatusOK, gin.H{
					"msg": "success",
				})
			})
		}
	})
	r.GET("/liver/:id", func(c *gin.Context) {
		type S struct {
			AreaLiver
			LiveCount int
			LiveMoney float64
			FansMoney float64
		}
		uid := c.DefaultQuery("id", "-1")
		if uid == "-1" {
			c.JSON(http.StatusOK, gin.H{
				"msg": "params missing",
			})
		}
		var result0 = AreaLiver{}
		var s = S{}
		db.Model(&AreaLiver{}).Where("uid = ?", uid).Find(&result0)
		s.AreaLiver = result0
		//db.Model()
		c.JSON(http.StatusOK, gin.H{
			"liver": s,
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

		if name == "1" {
			query.Find(&f)
		} else {
			query = query.Where("user_name = ?", name)
			query.Find(&f)
			db.Model(&Live{}).Where("user_name = ?", name).Count(&totalRecords)
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
	r.GET("/refreshMoney", func(context *gin.Context) {
		go FixMoney()
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	r.GET("/parse", func(context *gin.Context) {
		id := context.DefaultQuery("bv", "10")

		var videos = ParseSingleVideo(id)
		for _, video := range videos {
			videoCache = append(videoCache, video)
		}
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    videos,
		})
	})

	r.GET("/parseList", func(context *gin.Context) {
		id := context.DefaultQuery("mid", "10")
		listId := context.DefaultQuery("season", "10")

		var found = ParsePlayList(id, listId)
		for _, video := range found {
			videoCache = append(videoCache, video)
		}
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    found,
		})
	})
	r.GET("/download", func(context *gin.Context) {
		bv := context.DefaultQuery("bv", "BV16TP5euE5s")
		partStr := context.DefaultQuery("part", "1")
		if partStr == "0" {
			partStr = "1"
		}
		for _, video := range videoCache {
			if strconv.Itoa(video.Part) == partStr && bv == video.BV {
				worker.AddTask(func() {
					UploadArchive(video)
				})
				context.JSON(http.StatusOK, gin.H{
					"message": "success",
				})
				return
			}

		}
		worker.AddTask(func() {
			UploadArchive(ParseSingleVideo(bv)[0])
		})
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
		})

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
		res, _ := client.R().SetHeader("Referer", "https://www.bilibili.com/").Get(url)
		c.Writer.Header().Set("Content-Type", res.Header().Get("Content-Type"))
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		c.Writer.WriteHeader(res.StatusCode())
		c.Writer.Write(res.Body())
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

	//直播间最近的弹幕，用于在前端的实时直播页面显示
	r.GET("/history", func(context *gin.Context) {
		last := context.DefaultQuery("last", "")
		roomStr := context.DefaultQuery("room", "1")
		if config.Mode == "Master" {
			add, _ := man.GetNodeByTask(roomStr)
			if !man.isSelf(add) {
				r, _ := client.R().Get(add + "/history?room=" + roomStr + "&last=" + last)
				context.String(r.StatusCode(), r.String())
				return
			}
		}

		var result = []FrontLiveAction{}
		if lives[roomStr] == nil {
			context.JSON(http.StatusOK, gin.H{
				"message": "error",
				"data":    []string{},
			})
			return
		}

		if last != "" {
			var match = false
			for _, action := range lives[roomStr].Danmuku {
				if action.UUID == last {
					match = true
					continue
				}
				if match {
					result = append(result, action)
				}
			}
		} else {
			result = []FrontLiveAction{}
			if lives[roomStr] != nil && lives[roomStr].Danmuku != nil {
				result = lives[roomStr].Danmuku
			}
		}
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    result,
		})

	})

	//搜索主播，暂时废弃
	r.GET("/searchLive", func(context *gin.Context) {
		url := "https://api.bilibili.com/x/web-interface/wbi/search/type?page=1&page_size=42&order=online&keyword=" + context.Query("keyword") + "&search_type=live_user"
		obj, _ := url2.Parse((url))
		now := time.Now()
		signed, _ := wbi.SignQuery(obj.Query(), now)
		final := "https://api.bilibili.com/x/web-interface/wbi/search/type?" + signed.Encode()
		res, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).SetHeader("Referer", "https://www.bilibili.com").Get(final)
		var list = LiveListResponse{}
		sonic.Unmarshal(res.Body(), &list)
		for i, s := range list.Data.Result {
			t := strings.Split(s.UName, "</em>")
			s.UName = extractTextFromHTML(s.UName)
			if len(t) == 2 {
				list.Data.Result[i].UName = s.UName + t[1]
			}

		}
		context.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    list,
		})
	})
	r.GET("/status", func(context *gin.Context) {
		var wg sync.WaitGroup
		var t1 = msg1
		var t2 = msg5
		var t3 = msg60
		var l sync.Mutex
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

				res, err := client.R().
					SetResult(&result).
					Get(n.Address + "/count")

				if err != nil {
					log.Printf("请求子节点 %s 失败: %v", n.Address, err)
					return
				}
				for i, s := range strings.Split(res.String(), ",") {
					l.Lock()
					if i == 0 {
						t1 = t1 + toInt64(s)
					}
					if i == 1 {
						t2 = t2 + toInt64(s)
					}
					if i == 2 {
						t3 = t3 + toInt64(s)
					}
					l.Unlock()
				}

			}(node)
		}

		wg.Wait()

		context.JSON(http.StatusOK, gin.H{
			"Requests":      totalRequests,
			"LaunchedAt":    launchTime.Format(time.DateTime),
			"Livers":        TotalLiver(),
			"TotalMessages": TotalMessage(),
			"Message1":      t1,
			"Message5":      t2,
			"MessageHour":   t3,
			"MessageDaily":  MinuteMessageCount(1440),
			"HTTPBytes":     httpBytes,
			"WSBytes":       websocketBytes,
			"Nodes":         man.Nodes,
		})
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

	r.GET("/setCookie", func(c *gin.Context) {
		var cookie = c.Query("cookie")
		if SelfUID(cookie) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "invalid cookie",
			})
		} else {
			config.Cookie = cookie
			SaveConfig()
			c.JSON(http.StatusOK, gin.H{
				"msg": "success",
			})
		}
	})

	//所有在虚拟区开播过的主播

	r.GET("/areaLivers", func(c *gin.Context) {
		var dist = make([]FrontAreaLiver, 0)
		copier.Copy(&dist, &cachedLivers)
		var sortType = c.Query("sort")
		if sortType == "guard" {
			sort.Slice(dist, func(i, j int) bool {
				var g1 = 0
				for _, s := range strings.Split(dist[i].Guard, ",") {
					var n, _ = strconv.ParseInt(s, 10, 64)
					g1 = g1 + int(n)
				}
				var g2 = 0
				for _, s := range strings.Split(dist[j].Guard, ",") {
					var n, _ = strconv.ParseInt(s, 10, 64)
					g1 = g1 + int(n)
				}
				return g1 > g2
			})
		} else if sortType == "l1-guard" {
			sort.Slice(dist, func(i, j int) bool {
				var g1, _ = strconv.ParseInt(strings.Split(dist[i].Guard, ",")[0], 10, 64)
				var g2, _ = strconv.ParseInt(strings.Split(dist[j].Guard, ",")[0], 10, 64)
				return g1 > g2
			})
		} else if sortType == "diff" {
			sort.Slice(dist, func(i, j int) bool {
				return dist[i].DailyDiff > dist[j].DailyDiff
			})
		} else if sortType == "diff-desc" {
			sort.Slice(dist, func(i, j int) bool {
				return dist[i].DailyDiff < dist[j].DailyDiff
			})
		} else if sortType == "guard-equal" {
			var price = []int{19998, 1998, 168}
			sort.Slice(dist, func(i, j int) bool {
				var g1 = 0
				for i, s := range strings.Split(dist[i].Guard, ",") {
					var n, _ = strconv.ParseInt(s, 10, 64)
					g1 = g1 + int(n)*price[i]
				}
				var g2 = 0
				for i, s := range strings.Split(dist[j].Guard, ",") {
					var n, _ = strconv.ParseInt(s, 10, 64)
					g2 = g2 + int(n)*price[i] //https://www.bilibili.com/video/BV1QTZdYZEn1/
				}
				return g1 > g2
			})
		} else {
			sort.Slice(dist, func(i, j int) bool {
				return dist[i].Fans > dist[j].Fans
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"list": dist,
		})

	})
	r.GET("/money", func(c *gin.Context) {
		array := make([]LiveAction, 0)
		uid := c.Query("uid")
		db.Model(&LiveAction{}).Where("from_id = ? and gift_price != 0", uid).Find(&array)
	})
	r.GET("/debug", func(c *gin.Context) {
		var room = c.Query("room")
		url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + room
		res, _ := client.R().Get(url)
		status := LiveStatusResponse{}
		sonic.Unmarshal(res.Body(), &status)
		c.JSON(http.StatusOK, gin.H{
			"ServerState": status.Data.LiveStatus,
			"IsLive":      isLive(room),
		})
	})

	//获取直播流
	r.GET("/stream", func(c *gin.Context) {
		var room = c.Query("room")
		c.JSON(http.StatusOK, gin.H{
			"Stream:": GetLiveStream(room),
		})
	})

	//Trace直播间

	r.GET("/trace", func(c *gin.Context) {
		var room = c.Query("room")
		livesMutex.Lock()
		lives[room] = &Status{RemainTrying: 40}
		lives[room].LiveRoom = room
		lives[room].Danmuku = make([]FrontLiveAction, 0)
		lives[room].OnlineWatcher = make([]Watcher, 0)
		lives[room].GuardList = make([]Watcher, 0)
		livesMutex.Unlock()
		worker.AddTask(func() {
			go TraceLive(room)
			time.Sleep(45 * time.Second)
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

	r.GET("/count", func(context *gin.Context) {
		context.String(http.StatusOK, fmt.Sprintf("%d,%d,%d", msg1, msg5, msg60))
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
		db.Raw("SELECT from_id,from_name,medal_level,medal_name FROM live_actions where live = ?  GROUP BY from_id order by medal_level desc", liveInt).Scan(&dst)
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
