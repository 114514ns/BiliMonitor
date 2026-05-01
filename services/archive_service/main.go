package main

import (
	"encoding/json"
	"log"
	"os"

	bili "github.com/114514ns/BiliClient"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

var clickDb *gorm.DB

var config Config

func loadConfig() {
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bytes, &config)
}

var clients = make(map[string]*bili.BiliClient)

var protoMap = make(map[string]*desc.MessageDescriptor)

//评论区是每次完整记录，转成pb直接存clickhouse ，每次更新的时候完整爬取评论区，留一个特殊列，里面是合并了所有评论。
//用户主页。用card接口，直接存原始json
//弹幕 每次更新的时候完整爬取

// 视频流

func loadPb() {
	parser := protoparse.Parser{}
	var fds, _ = parser.ParseFiles("pb/pb.proto")
	protoMap["REPLY_LIST"] = fds[0].FindMessage("models.ReplyList")
	protoMap["REACTION"] = fds[0].FindMessage("models.Reaction")
	protoMap["REACTION_LIST"] = fds[0].FindMessage("models.ReactionList")

	fds, _ = parser.ParseFiles("pb/danmaku.proto")
	protoMap["DANMAKU_LIST"] = fds[0].FindMessage("models.DmSegMobileReply")
}

//UpdateCollections 判断之前是否记录过，如果纪录过只会更新新增加的

func main() {
	loadConfig()
	loadPb()
	clickDb0, e := gorm.Open(clickhouse.Open(config.ClickDSL))

	if e != nil {
		log.Fatal(e)
	}
	clickDb = clickDb0

	clickDb.AutoMigrate(&Danmaku{})
	clickDb.AutoMigrate(&ReplyList{})
	clickDb.AutoMigrate(&Collection{})

	clickDb.Table("reactions").AutoMigrate(&struct {
		OID int64
		Pb  string
	}{})

	clickDb0.Exec("SET max_query_size = 67108864")
	UpdateClients()
	UpdateUserVideo(3546757543758795)
	//UpdateVideo("BV1M2YqzpENF")
	UpdateCollections(1160346, 504140200, "season")
	//UpdateFeedDetails(RandomM(clients).GetDynamicDetail(1177090080201768985)[0])

	select {}
}
