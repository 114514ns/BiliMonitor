package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func InitHTTP() {
	r := gin.Default()
	r.GET("/monitor", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"lives": lives,
		})
	})
	r.GET("/lives", func(c *gin.Context) {
		var f []Live
		db.Find(&f)
		c.JSON(http.StatusOK, gin.H{

			"lives": f,
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
		limit := 10
		offset := (page - 1) * limit

		// 从数据库查询

		var totalRecords int64
		if err := db.Model(&LiveAction{}).Where("live = ? and action_name != 'enter'", id).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database count error"})
			return
		}

		// 计算总页数
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit)) // 向上取整

		var records []LiveAction
		if err := db.Where("live = ? and action_name != 'enter'", id).Offset(offset).Limit(limit).Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database query error"})
			return
		}

		// 返回查询结果
		c.JSON(http.StatusOK, gin.H{
			"totalPages":   totalPages,
			"totalRecords": totalRecords,
			"page":         page,
			"records":      records,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
