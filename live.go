package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SelfInfo struct {
	Data struct {
		Mid int `json:"mid"`
	} `json:"data"`
}

func SelfUID(cookie string) string {
	res, _ := client.R().Get("https://api.bilibili.com/x/web-interface/nav")

	var self = SelfInfo{}
	sonic.Unmarshal(res.Body(), &self)
	return strconv.Itoa(self.Data.Mid)
}
func FixMoney() {
	var lives0 []Live
	db.Find(&lives0)

	for _, v := range lives0 {
		var sum float64
		db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", v.ID).Scan(&sum)
		result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sum), 64)
		var msgCount int64
		db.Model(&LiveAction{}).Where("live = ? and action_name = 'msg'", v.ID).Count(&msgCount)
		v.Money = result
		v.Message = int(msgCount)
		db.Save(&v) // 分别更新每条记录
	}
}
func RemoveEmpty() {
	db.Where("money = 0 and message = 0").Delete(&Live{})
}
func TraceLive(roomId string) {
	var roomUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + roomId
	var rRes, _ = client.R().Get(roomUrl)
	var liver string
	var roomInfo = RoomInfo{}
	sonic.Unmarshal(rRes.Body(), &roomInfo)
	FillGiftPrice(roomId, roomInfo.Data.AreaId, roomInfo.Data.ParentAreaId)
	var dbLiveId = 0
	var liverId = strconv.Itoa(roomInfo.Data.UID)
	var startAt = roomInfo.Data.LiveTime

	var living = false
	var liverInfoUrl = "https://api.live.bilibili.com/live_user/v1/Master/info?uid=" + liverId
	liverRes, _ := client.R().Get(liverInfoUrl)
	var liverObj = LiverInfo{}
	sonic.Unmarshal(liverRes.Body(), &liverObj)
	liver = liverObj.Data.Info.Uname
	lives[roomId].UName = liver

	lives[roomId].UID = liverId
	lives[roomId].Area = roomInfo.Data.Area
	lives[roomId].Title = roomInfo.Data.Title
	if !strings.Contains(startAt, "0000-00-00 00:00:00") {
		lives[roomId].Live = true

		//当前是开播状态
		var serverStartAt, _ = time.Parse(time.DateTime, startAt)
		living = true

		var foundLive = Live{}

		db.Where("user_id=?", roomInfo.Data.UID).Last(&foundLive)

		var diff = abs(int(foundLive.StartAt - serverStartAt.Unix()))

		log.Println("diff  " + strconv.Itoa(diff))

		if diff < 90 {
			log.Println("续")

			dbLiveId = int(foundLive.ID)
		} else {

			var new = Live{}
			new.UserID = liverId
			new.StartAt = serverStartAt.Unix()
			//new.ID,_ : = strconv.Atoi(roomId)
			new.Title = roomInfo.Data.Title
			new.Area = roomInfo.Data.Area
			//new.UserName = roomInfo.Data

			var i, _ = strconv.Atoi(roomId)
			new.RoomId = i
			new.UserName = liver
			liver = strings.TrimSpace(liver) // 去除前后的空白字符

			db.Create(&new)
			dbLiveId = int(new.ID)
		}
	}

	var url0 = "https://api.bilibili.com/x/web-interface/nav"
	url0 = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + roomId
	res, _ := client.R().SetHeader("Cookie", config.Cookie).Get(url0)
	var liveInfo = LiveInfo{}
	sonic.Unmarshal(res.Body(), &liveInfo)
	if len(liveInfo.Data.HostList) == 0 {
		log.Println(res.String())
	}
	u := url.URL{Scheme: "wss", Host: liveInfo.Data.HostList[0].Host + ":2245", Path: "/sub"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	ticker := time.NewTicker(45 * time.Second)
	go func() {
		log.Printf("[%s] 成功连接到弹幕服务器", liver)
		var cer = Certificate{}
		cer.Uid = 3546580934199673
		id, _ := strconv.Atoi(roomId)
		cer.RoomId = id
		cer.Type = 2
		cer.Key = liveInfo.Data.Token
		cer.Cookie = strings.Replace(config.Cookie, "buvid3=", "", 1)
		cer.Protover = 3
		json, _ := sonic.Marshal(&cer)

		err := c.WriteMessage(websocket.TextMessage, BuildMessage(string(json), 7))
		if err != nil {
			return
		}
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("[System] 登录失败，请更换Cookie")
				lives[roomId].LastActive = 114514
				return
			}
			// 打印接收到的消息
			reader := io.NewSectionReader(bytes.NewReader(message), 16, int64(len(message)-16))
			brotliReader := brotli.NewReader(reader)
			var decompressedData bytes.Buffer

			// 通过 io.Copy 将解压后的数据写入到缓冲区
			var msg = ""
			_, err0 := io.Copy(&decompressedData, brotliReader)
			if err0 != nil {
				msg = string(message)
			} else {
				msg = string(decompressedData.Bytes())
			}
			buffer := bytes.NewReader([]byte(msg))

			for {
				// 检查是否有足够的数据来解码头部
				if buffer.Len() < 16 {
					log.Println("Insufficient data to read header")
					break
				}

				// 解码头部
				var totalSize uint32
				var headerLength uint16
				var protocolVersion uint16
				var operation uint32
				var sequence uint32

				binary.Read(buffer, binary.BigEndian, &totalSize)
				binary.Read(buffer, binary.BigEndian, &headerLength)
				binary.Read(buffer, binary.BigEndian, &protocolVersion)
				binary.Read(buffer, binary.BigEndian, &operation)
				binary.Read(buffer, binary.BigEndian, &sequence)

				// 验证数据完整性
				if buffer.Len() < int(totalSize-16) {
					log.Println("Insufficient data for complete packet")
					break
				}

				// 读取消息体部分
				msgData := make([]byte, totalSize-16)
				buffer.Read(msgData)

				var obj = string(msgData)
				var action = LiveAction{}
				action.Live = uint(dbLiveId)
				action.LiveRoom = roomId
				action.GiftPrice = sql.NullFloat64{Float64: 0, Valid: false}
				action.GiftAmount = sql.NullInt16{Int16: 0, Valid: false}
				var text = LiveText{}
				sonic.Unmarshal(msgData, &text)
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") { // 弹幕
					action.ActionName = "msg"
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = strconv.Itoa(int(text.Info[2].([]interface{})[0].(float64)))
					action.Extra = text.Info[1].(string)
					db.Create(&action)
					log.Println("[" + liver + "]  " + text.Info[2].([]interface{})[1].(string) + "  " + text.Info[1].(string))

				} else if strings.Contains(obj, "SEND_GIFT") { //送礼物
					var info = GiftInfo{}
					sonic.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.FromId = strconv.Itoa(info.Data.SenderUinfo.UID)
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)
					action.GiftPrice = sql.NullFloat64{Float64: result, Valid: true}
					action.GiftAmount = sql.NullInt16{Int16: int16(info.Data.Num), Valid: true}
					db.Create(&action)
					log.Printf("[%s] %s 赠送了 %d 个 %s，%.2f元", liver, info.Data.Uname, info.Data.Num, info.Data.GiftName, price)
				} else if strings.Contains(obj, "INTERACT_WORD") { //进入直播间
					/*
						var entet = EnterLive{}
						sonic.Unmarshal(msgData, &entet)
						action.FromId = strconv.Itoa(entet.Data.UID)
						action.FromName = entet.Data.Uname
						action.ActionName = "enter"
						db.Create(&action)

					*/
				} else if strings.Contains(obj, "PREPARING") { //下播
					lives[roomId].Live = false
					var sum float64
					db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", dbLiveId).Scan(&sum)

					db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{EndAt: time.Now().Unix(), Money: sum})
					living = false

				} else if text.Cmd == "LIVE" {

					//else if strings.Contains(obj, "LIVE") && !strings.Contains(obj, "STOP_LIVE_ROOM_LIST") && !strings.Contains(obj, "LIVE_MULTI_VIEW") && !strings.Contains(obj, "live_time") && !strings.Contains(obj, "LIVE_INTERACT_GAME") && !strings.Contains(obj, "LIVE_OPEN_PLATFORM") && !strings.Contains(obj, "LIVE_ROOM_TOAST") { //开播
					var new = Live{}
					living = true
					new.UserID = liverId
					time.Sleep(time.Second * 5) //如果马上去请求直播间信息会有问题
					var r, _ = client.R().Get(roomUrl)
					sonic.Unmarshal(r.Body(), &roomInfo)

					var serverStartAt = time.Now() //time.Parse(time.DateTime, roomInfo.Data.LiveTime)

					var foundLive = Live{}

					db.Where("user_id=?", roomInfo.Data.UID).Last(&foundLive)

					var diff = abs(int(foundLive.StartAt - serverStartAt.Unix()))

					if diff-8*3600 > 60*15 && !strings.Contains(msg, "WATCHED_CHANGE") /*&& roomInfo.Data.LiveTime != "0000-00-00 00:00:00"*/ {

						v := serverStartAt
						v = v.Add(time.Hour * 8)
						new.StartAt = v.Unix()
						new.Title = roomInfo.Data.Title
						new.Area = roomInfo.Data.Area
						var i, _ = strconv.Atoi(roomId)
						new.RoomId = i
						new.UserName = liver
						lives[roomId].Live = true
						liver = strings.TrimSpace(liver) // 去除前后的空白字符

						db.Create(&new)
						dbLiveId = int(new.ID)
						var msg = "你关注的主播： " + liver + " 开始直播"
						PushDynamic(msg, roomInfo.Data.Title)
					} else {
						strings.Contains("", "")
					}

				} else if strings.Contains(obj, "SUPER_CHAT_MESSAGE") { //SC
					var sc = SuperChatInfo{}
					sonic.Unmarshal(msgData, &sc)

					action.ActionName = "sc"
					action.FromName = sc.Data.UserInfo.Uname
					action.FromId = strconv.Itoa(sc.Data.Uid)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sc.Data.Price), 64)
					action.GiftPrice = sql.NullFloat64{Float64: result, Valid: true}

					action.GiftAmount = sql.NullInt16{Valid: true, Int16: 1}
					action.Extra = sc.Data.Message
					if action.FromId != "0" {
						db.Create(&action)
					}
				} else if strings.Contains(obj, "GUARD_BUY") { //上舰
					var guard = GuardInfo{}
					sonic.Unmarshal(msgData, &guard)
					action.FromId = strconv.Itoa(guard.Data.Uid)
					action.FromName = guard.Data.Username
					action.GiftName = guard.Data.GiftName
					switch action.GiftName {
					case "舰长":
						action.GiftPrice = sql.NullFloat64{Float64: float64(138 * guard.Data.Num), Valid: true}
					case "提督":
						action.GiftPrice = sql.NullFloat64{Float64: float64(1998 * guard.Data.Num), Valid: true}
					case "总督":
						action.GiftPrice = sql.NullFloat64{Float64: float64(19998 * guard.Data.Num), Valid: true}
					}

					db.Create(&action)
				} else if text.Cmd == "WATCHED_CHANGE" {
					if living {
						var obj = Watched{}
						sonic.Unmarshal(msgData, &obj)
						db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{Watch: obj.Data.Num})
					}
				}

				// 假设读取完一个封包，如果已没有足够数据来读取下一个头部，退出循环
				if buffer.Len() < 16 {
					break
				}
			}
			if !strings.Contains(msg, "[object") {

				//log.Printf("Received: %s", substr(msg, 16, len(msg)))
			}
		}
	}()
	for {
		select {
		case <-ticker.C:
			// 每30秒向服务端发送一次消息
			err = c.WriteMessage(websocket.TextMessage, BuildMessage("[object Object]", 2))
			if err != nil {
				lives[roomId].LastActive = 114514
				log.Println("write:", err)
				return
			}
		}

	}
}
func BuildMessage(str string, opCode int) []byte {
	buffer := new(bytes.Buffer)
	totalSize := uint32(16 + len(str)) // 封包总大小
	headerLength := uint16(16)         // 头部长度
	protocolVersion := uint16(1)       // 协议版本
	operation := uint32(opCode)        // 操作码
	sequence := uint32(1)              // sequence

	binary.Write(buffer, binary.BigEndian, totalSize)
	binary.Write(buffer, binary.BigEndian, headerLength)
	binary.Write(buffer, binary.BigEndian, protocolVersion)
	binary.Write(buffer, binary.BigEndian, operation)
	binary.Write(buffer, binary.BigEndian, sequence)
	buffer.Write([]byte(str))

	return buffer.Bytes()
}
func Index(s string, index int) string {
	runes := bytes.Runes([]byte(s))
	for i, rune := range runes {
		if i == int(index) {
			return string(rune)
		}
	}
	return ""
}
func FillGiftPrice(room string, area int, parent int) {
	htmlRes, _ := client.R().Get("https://live.bilibili.com/13878454")
	htmlStr := htmlRes.String()
	strings.Index(htmlStr, `"area_id"`)
	Index(htmlStr, strings.Index(htmlStr, `"area_id"`))
	res, _ := client.R().Get("https://api.live.bilibili.com/xlive/web-room/v1/giftPanel/roomGiftList?platform=pc&room_id=" + room + "&area_id=" + strconv.Itoa(area) + "&area_parent_id" + strconv.Itoa(parent))
	var gift = GiftList{}
	sonic.Unmarshal(res.Body(), &gift)
	for i := range gift.Data.GiftConfig.BaseConfig.List {
		var item = gift.Data.GiftConfig.BaseConfig.List[i]

		if strings.Contains(item.Name, "盲盒") {
			res, _ := client.R().SetHeader("Cookie", config.Cookie).Get("https://api.live.bilibili.com/xlive/general-interface/v1/blindFirstWin/getInfo?gift_id=" + strconv.Itoa(item.ID))

			var box = GiftBox{}
			sonic.Unmarshal(res.Body(), &box)
			for i2 := range box.Data.Gifts {
				var item0 = box.Data.Gifts[i2]
				GiftPrice[item0.GiftName] = float32(item0.Price) / 1000.0
			}
		} else {
			GiftPrice[item.Name] = float32(item.Price) / 1000.0
		}

	}
	for i := range gift.Data.GiftConfig.RoomConfig {
		var item = gift.Data.GiftConfig.RoomConfig[i]
		GiftPrice[item.Name] = float32(item.Price) / 1000.0
	}
	fmt.Println()

}

type Certificate struct {
	Uid      int    `json:"uid"`
	RoomId   int    `json:"roomid"`
	Key      string `json:"key"`
	Protover int    `json:"protover"`
	Cookie   string `json:"buvid"`
	Type     int    `json:"type"`
}
type LiveInfo struct {
	Data struct {
		Group            string  `json:"group"`
		BusinessID       int     `json:"business_id"`
		RefreshRowFactor float64 `json:"refresh_row_factor"`
		RefreshRate      int     `json:"refresh_rate"`
		MaxDelay         int     `json:"max_delay"`
		Token            string  `json:"token"`
		HostList         []struct {
			Host    string `json:"host"`
			Port    int    `json:"port"`
			WssPort int    `json:"wss_port"`
			WsPort  int    `json:"ws_port"`
		} `json:"host_list"`
	} `json:"data"`
}
type GiftList struct {
	Data struct {
		GiftConfig struct {
			BaseConfig struct {
				List []struct {
					ID    int    `json:"id"`
					Name  string `json:"name"`
					Price int    `json:"price"`
				} `json:"list"`
			} `json:"base_config"`
			RoomConfig []struct {
				Name  string `json:"name"`
				Price int    `json:"price"`
			} `json:"room_config"`
		} `json:"gift_config"`
	} `json:"data"`
}
type GiftInfo struct {
	Cmd  string `json:"cmd"`
	Data struct {
		GiftName        string `json:"giftName"`
		Num             int    `json:"num"`
		ReceiveUserInfo struct {
			UID   int    `json:"uid"`
			Uname string `json:"uname"`
		} `json:"receive_user_info"`
		SenderUinfo struct {
			Base struct {
				Name string `json:"name"`
			} `json:"base"`
			UID int `json:"uid"`
		} `json:"sender_uinfo"`
		UID   int    `json:"uid"`
		Uname string `json:"uname"`
	} `json:"data"`
}
type LiveText struct {
	Cmd  string        `json:"cmd"`
	DmV2 string        `json:"dm_v2"`
	Info []interface{} `json:"info"`
}
type GiftBox struct {
	Data struct {
		Gifts []struct {
			Price    int    `json:"price"`
			GiftName string `json:"gift_name"`
		} `json:"gifts"`
	} `json:"data"`
}
type LiveAction struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	Live       uint
	FromName   string
	FromId     string
	LiveRoom   string
	ActionName string
	GiftName   string
	GiftPrice  sql.NullFloat64 `gorm:"scale:2;precision:7"`
	GiftAmount sql.NullInt16
	Extra      string
}
type RoomInfo struct {
	Data struct {
		LiveTime     string `json:"live_time"`
		UID          int    `json:"uid"`
		Title        string `json:"title"`
		Area         string `json:"area_name"`
		AreaId       int    `json:"area_id"`
		ParentAreaId int    `json:"parent_area_id"`
	} `json:"data"`
}
type Live struct {
	gorm.Model
	Title    string
	StartAt  int64
	EndAt    int64
	UserName string
	UserID   string
	Area     string
	RoomId   int
	Money    float64 `gorm:"type:decimal(7,2)"`
	Message  int
	Watch    int
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		TriggerTime int64  `json:"trigger_time"`
		UID         int    `json:"uid"`
		Uname       string `json:"uname"`
	} `json:"data"`
}

type LiverInfo struct {
	Data struct {
		Info struct {
			Uname string `json:"uname"`
		} `json:"info"`
	} `json:"data"`
}

type SuperChatInfo struct {
	Cmd  string `json:"cmd"`
	Data struct {
		Message  string  `json:"message"`
		Price    float64 `json:"price"`
		Uid      int     `json:"uid"`
		UserInfo struct {
			Uname string `json:"uname"`
		} `json:"user_info"`
	} `json:"data"`
}

type GuardInfo struct {
	Cmd  string `json:"cmd"`
	Data struct {
		Uid        int    `json:"uid"`
		Username   string `json:"username"`
		GuardLevel int    `json:"guard_level"`
		Num        int    `json:"num"`
		Price      int    `json:"price"`
		GiftId     int    `json:"gift_id"`
		GiftName   string `json:"gift_name"`
		StartTime  int    `json:"start_time"`
		EndTime    int    `json:"end_time"`
	} `json:"data"`
}
type Watched struct {
	Cmd  string `json:"cmd"`
	Data struct {
		Num       int    `json:"num"`
		TextSmall string `json:"text_small"`
		TextLarge string `json:"text_large"`
	} `json:"data"`
}
