package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

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

func InitHttp() {
	var r = gin.Default()

	r.Use(TTLMiddleware())

	r.GET("/snapshot/space", func(context *gin.Context) {
		clickDb.Raw("select ")
	})

	r.GET("/snapshot/replies", func(context *gin.Context) {
		var oid = context.Query("oid")
		if toInt64(oid) <= 0 {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"data": ReadReplies(toInt64(oid)),
		})
	})

	r.POST("/schedule/collection", func(context *gin.Context) {

	})

	r.Run(fmt.Sprintf(":%d", config.Port))
}
