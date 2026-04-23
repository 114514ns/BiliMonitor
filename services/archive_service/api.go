package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jhump/protoreflect/dynamic"
)

func InitHttp() {
	var r = gin.Default()

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
		var list ReplyList
		clickDb.Raw("select * from reply_lists where o_id=? order by created_at desc limit 1", oid).Scan(&list)

		var msg = dynamic.NewMessage(protoMap["REPLY_LIST"])
		msg.Unmarshal(list.Pb)
		jsonBytes, _ := msg.MarshalJSON()
		var stored struct {
			Replies []interface{}
		}
		json.Unmarshal(jsonBytes, &stored)
		context.JSON(http.StatusOK, gin.H{
			"data": stored,
		})
	})

	r.Run(fmt.Sprintf(":%d", config.Port))
}
