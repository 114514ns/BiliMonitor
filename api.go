package main

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var cache = []Video{}
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

func InitHTTP() {
	r := gin.Default()
	r.Use(CORSMiddleware())
	//r.Static("/page", "./Page/dist/")
	//r.Static("/assets", "./Page/dist/assets")
	r.Use(static.Serve("/", static.LocalFile("./Page/dist", false)))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/monitor", func(c *gin.Context) {

		var array = make([]Status, 0)
		for s := range lives {
			copy, _ := DeepCopy(*lives[s])
			copy.Danmuku = []FrontLiveAction{}
			array = append(array, copy)
		}
		c.JSON(http.StatusOK, gin.H{
			"lives": array,
		})
	})
	r.GET("/monitor/:id", func(c *gin.Context) {
		id := c.Param("id")
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
	r.GET("/live", func(c *gin.Context) {
		var f []Live
		name := c.DefaultQuery("name", "1")
		pageStr := c.DefaultQuery("page", "1")
		page, _ := strconv.Atoi(pageStr)
		limitStr := c.DefaultQuery("limit", "10")
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit
		var totalRecords int64
		if err := db.Model(&Live{}).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}
		if name == "1" {
			db.Offset(offset).Limit(limit).Find(&f)
		} else {
			db.Where("user_name = ?", name).Offset(offset).Limit(limit).Find(&f)
			db.Model(&Live{}).Where("user_name = ", name).Count(&totalRecords)
		}
		var off int64 = 1
		if totalRecords%int64(limit) == 0 {
			off = 0
		}
		c.JSON(http.StatusOK, gin.H{

			"totalPage": totalRecords/(int64(limit)) + off,
			"lives":     f,
		})
	})
	r.GET("/live/:id/", func(c *gin.Context) {
		id := c.Param("id")
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
			return
		}

		limitStr := c.DefaultQuery("limit", "10")
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit

		orderStr := c.DefaultQuery("order", "ascend")

		var totalRecords int64
		if err := db.Model(&LiveAction{}).Where("live = ? and action_name != 'enter'", id).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}

		orderQuery := ""
		if orderStr == "ascend" {
			orderQuery = "gift_price asc"
		} else if orderStr == "descend" {
			orderQuery = "gift_price desc"
		} else {
			orderQuery = "id asc"
		}

		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

		var records []LiveAction
		if err := db.Where("live = ? and action_name != 'enter'", id).Order(orderQuery).Offset(offset).Limit(limit).Find(&records).Error; err != nil {
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
			go TraceLive(id)
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
			cache = append(cache, video)
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
			cache = append(cache, video)
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
		for _, video := range cache {
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

		context.JSON(http.StatusOK, gin.H{
			"message": "可能出bug了",
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
	r.GET("/history", func(context *gin.Context) {
		roomStr := context.DefaultQuery("room", "1")
		last := context.DefaultQuery("last", "")
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

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
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

		// 继续处理请求
		c.Next()
	}
}
