package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	bili "github.com/114514ns/BiliClient"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Port   int
	Cookie string
	Folder string
}

var client = bili.NewAnonymousClient(bili.ClientOptions{})
var drive *CMDrive

var worker = NewWorker(8)

var config Config

var cacheDir = make(map[string]string)

func main() {
	bytes, _ := os.ReadFile("config.json")
	json.Unmarshal(bytes, &config)

	drive = NewDrive(config.Cookie)

	for _, v := range drive.ListFiles(config.Folder) {
		cacheDir[v.FileName] = v.FileID
	}
	var r = gin.Default()
	r.GET("/download", func(context *gin.Context) {
		var bv = context.Query("bv")
		var cid = context.Query("cid")
		if bv == "" {
			context.JSON(http.StatusBadRequest, "BV or cid is empty")
			return
		}
		if toInt(cid) <= 0 {
			r0, _ := client.Resty.R().Post("https://api.live.bilibili.com//xlive/open-platform/v1/inner/getArchiveInfo?bv_id=" + bv)
			var o0 map[string]interface{}
			json.Unmarshal(r0.Body(), &o0)
			cid = strconv.Itoa(getInt(o0, "data.cid"))
		}
		worker.AddTask(func() {
			HandleDownload(bv, cid)
		})
		context.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"Pending": worker.QueueLen(),
		})
	})

	r.Run(":" + strconv.Itoa(config.Port))

	select {}

}

func HandleDownload(bv, cid string) {
	var array = client.GetVideoStream(bv, toInt(cid))
	var start = time.Now()
	if len(array) > 0 {
		var choose = array[0]
		if len(array) >= 3 {
			choose = array[2]
		}
		log.Printf("choose %s\n", time.Now().Sub(start))
		log.Printf("%v\n", choose)
		fmt.Println(choose)
		client.DownloadVideo(choose, "dst", false)
		log.Printf("download %s\n", time.Now().Sub(start))
		var fName = fmt.Sprintf("dst/%s-%s.mp4", bv, cid)
		var now = time.Now()
		drive.UploadFile(fName, cacheDir[fmt.Sprintf("%d%02d", now.Year(), now.Month())], fmt.Sprintf("%s-%s.mp4", bv, cid))
		log.Printf("upload %s\n", time.Now().Sub(start))
		time.Sleep(time.Second * 3)

		os.Remove("dst/" + bv + "-" + (cid) + ".mp4")
	} else {
		//os.WriteFile("/mnt/share/Stream/"+video.BV+"-"+strconv.Itoa(video.Cid)+".mp4", []byte(""), 0666)
	}
}
