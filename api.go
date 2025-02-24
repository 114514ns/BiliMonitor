package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

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
	r.GET("/monitor", func(c *gin.Context) {

		var array = make([]Status, 0)
		for s := range lives {
			array = append(array, *lives[s])
		}
		c.JSON(http.StatusOK, gin.H{
			"lives": array,
		})
	})
	r.GET("/liver", func(c *gin.Context) {
		key := c.DefaultQuery("key", "1") // 默认为第一页
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
		name := c.DefaultQuery("name", "1")    // 默认为第一页
		pageStr := c.DefaultQuery("page", "1") // 默认为第一页
		page, _ := strconv.Atoi(pageStr)
		limitStr := c.DefaultQuery("limit", "10") // 默认为第一页
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit
		var totalRecords int64
		if err := db.Model(&Live{}).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}

		// 计算总页数
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit)) // 向上取整
		if name == "1" {
			db.Offset(offset).Limit(limit).Find(&f)
		} else {
			db.Where("user_name = ?", name).Offset(offset).Limit(limit).Find(&f)
		}

		c.JSON(http.StatusOK, gin.H{

			"totalPage": totalPages,
			"lives":     f,
		})
	})
	r.GET("/live/:id/", func(c *gin.Context) {
		// 获取 ID 和 page 参数
		id := c.Param("id")
		pageStr := c.DefaultQuery("page", "1") // 默认为第一页
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
			return
		}

		// 假设每页有10条记录
		limitStr := c.DefaultQuery("limit", "10") // 默认为第一页
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit

		// 从数据库查询

		orderStr := c.DefaultQuery("order", "ascend") // 默认为第一页

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

		// 计算总页数
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit)) // 向上取整

		var records []LiveAction
		if err := db.Where("live = ? and action_name != 'enter'", id).Order(orderQuery).Offset(offset).Limit(limit).Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database query error"})
			return
		}

		var liveObj = &Live{}

		db.Model(&Live{}).Where("id = ?", id).Find(&liveObj)

		// 返回查询结果
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
