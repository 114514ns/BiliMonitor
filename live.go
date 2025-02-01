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

func TraceLive(roomId string) {
	FillGiftPrice(roomId)
	var roomUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + roomId
	var rRes, _ = client.R().Get(roomUrl)
	var liver string
	var roomInfo = RoomInfo{}
	sonic.Unmarshal(rRes.Body(), &roomInfo)
	var dbLiveId = 0
	var liverId = strconv.Itoa(roomInfo.Data.UID)
	var startAt = roomInfo.Data.LiveTime
	if !strings.Contains(startAt, "0000-00-00 00:00:00") {
		//当前是开播状态
		var serverStartAt, _ = time.Parse(time.DateTime, startAt)

		var foundLive = Live{}
		/*
					_ = db.Raw(`
				SELECT *
				FROM lives
				WHERE user_id = '?'
				ORDER BY id DESC
				LIMIT 1;
			    `, strconv.Itoa(roomInfo.Data.UID)).Scan(&foundLive).Error

		*/

		db.Where("user_id=?", roomInfo.Data.UID).Last(&foundLive)

		var diff = abs(int(foundLive.StartAt - serverStartAt.Unix()))

		log.Println("diff  " + strconv.Itoa(diff))
		htmlRes, _ := client.R().Get("https://live.bilibili.com/" + roomId)
		startStr := `meta name="keywords" content="`
		startIndex := strings.Index(htmlRes.String(), startStr)
		if startIndex == -1 {
			fmt.Println("Meta tag not found")
			return
		}

		// 计算content内容起始的实际索引
		contentStartIndex := startIndex + len(startStr)

		// 找到content内容的结束引号位置
		contentEndIndex := strings.Index(htmlRes.String()[contentStartIndex:], `"`)
		if contentEndIndex == -1 {
			fmt.Println("Content end quote not found")
			return
		}

		// 提取完整的content字符串
		fullContent := htmlRes.String()[contentStartIndex : contentStartIndex+contentEndIndex]

		// 找到第一个逗号的位置
		commaIndex := strings.Index(fullContent, ",")

		if commaIndex == -1 {
			liver = fullContent // 如果没有逗号，则使用完整的content内容
		} else {
			liver = fullContent[:commaIndex] // 提取逗号之前的内容
		}
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
	log.Printf("Connecting to %s", u.String())

	// 建立连接
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	ticker := time.NewTicker(45 * time.Second)
	// 启动一个goroutine来接收消息
	go func() {
		log.Println("成功连接ws服务器")
		var cer = Certificate{}
		cer.Uid = 3546580934199673
		id, _ := strconv.Atoi(roomId)
		cer.RoomId = id
		cer.Type = 2
		cer.Key = liveInfo.Data.Token
		cer.Cookie = strings.Replace(config.Cookie, "buvid3=", "", 1)
		cer.Protover = 3
		json, _ := sonic.Marshal(&cer)

		c.WriteMessage(websocket.TextMessage, BuildMessage(string(json), 7))

		for {
			// 读取从服务端传来的消息
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
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
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") {
					sonic.Unmarshal(msgData, &text)
					action.ActionName = "msg"
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = strconv.Itoa(int(text.Info[2].([]interface{})[0].(float64)))
					action.Extra = text.Info[1].(string)
					db.Create(&action)
					log.Println("[" + liver + "]  " + text.Info[2].([]interface{})[1].(string) + "  " + text.Info[1].(string))

				} else if strings.Contains(obj, "SEND_GIFT") {
					var info = GiftInfo{}
					sonic.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.FromId = strconv.Itoa(info.Data.SenderUinfo.UID)
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					action.GiftPrice = sql.NullFloat64{Float64: price, Valid: true}
					action.GiftAmount = sql.NullInt16{Int16: int16(info.Data.Num), Valid: true}
					db.Create(&action)
					log.Printf("[%s] %s 赠送了 %d 个 %s，%.2f元", liver, info.Data.Uname, info.Data.Num, info.Data.GiftName, price)
				} else if strings.Contains(obj, "INTERACT_WORD") {

					var entet = EnterLive{}
					sonic.Unmarshal(msgData, &entet)
					action.FromId = strconv.Itoa(entet.Data.UID)
					action.FromName = entet.Data.Uname
					action.ActionName = "enter"
					db.Create(&action)
				} else if strings.Contains(obj, "PREPARING") {
					//猜测是下播
					db.Model(&Live{}).Where("id= ?", dbLiveId).Update("end_at", time.Now().Unix())
					log.Println("下拨")

				} else if text.Cmd == "LIVE" {
					var new = Live{}
					new.UserID = liverId
					new.StartAt = time.Now().Unix()
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
					var msg = "你关注的主播： " + liver + " 开始直播"
					PushDynamic(msg, roomInfo.Data.Title)

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
func FillGiftPrice(room string) {
	res, _ := client.R().Get("https://api.live.bilibili.com/xlive/web-room/v1/giftPanel/roomGiftList?platform=pc&room_id=" + room)
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
		LiveTime string `json:"live_time"`
		UID      int    `json:"uid"`
		Title    string `json:"title"`
		Area     string `json:"area_name"`
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
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		TriggerTime int64  `json:"trigger_time"`
		UID         int    `json:"uid"`
		Uname       string `json:"uname"`
	} `json:"data"`
}
