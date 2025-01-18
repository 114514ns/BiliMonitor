package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.
var client = resty.New()
var Cookie = ""
var Special = make([]User, 0)
var RecordedDynamic = make([]string, 0)
var GiftPrice = map[string]float32{}
var mailClient = resend.NewClient("re_TLeNcEDu_Ht8QFPBRPH6JyKZjfnxmztwB")

type Config struct {
	SpecialDelay           string
	CommonDelay            string
	RefreshFollowingsDelay string
	User                   string
	SpecialList            []int
	Cookie                 string
	LoginMode              bool
	EnableEmail            bool
	FromMail               string
	ToMail                 []string
	EnableQQBot            bool
	ReportTo               []string
	BackServer             string
}

type FansList struct {
	Data struct {
		List []struct {
			Mid                string `json:"mid"`
			Attribute          int    `json:"attribute"`
			Uname              string `json:"uname"`
			Face               string `json:"face"`
			AttestationDisplay struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"attestation_display"`
		} `json:"list"`
	} `json:"data"`
	Ts        int64  `json:"ts"`
	RequestID string `json:"request_id"`
}
type UserState struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid       int `json:"mid"`
		Following int `json:"following"`
		Whisper   int `json:"whisper"`
		Black     int `json:"black"`
		Follower  int `json:"follower"`
	} `json:"data"`
}
type Basic struct {
	CommentIDStr string `json:"comment_id_str"`
	CommentType  int    `json:"comment_type"`
	LikeIcon     struct {
		ActionURL string `json:"action_url"`
		EndURL    string `json:"end_url"`
		ID        int    `json:"id"`
		StartURL  string `json:"start_url"`
	} `json:"like_icon"`
	RidStr string `json:"rid_str"`
}
type UserDynamic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Items []struct {
			Basic   Basic
			IDStr   string `json:"id_str"`
			Modules struct {
				ModuleDynamic struct {
					Additional interface{} `json:"additional"`
					Desc       interface{} `json:"desc"`
					Major      struct {
						Archive struct {
							Aid   string `json:"aid"`
							Badge struct {
								BgColor string      `json:"bg_color"`
								Color   string      `json:"color"`
								IconURL interface{} `json:"icon_url"`
								Text    string      `json:"text"`
							} `json:"badge"`
							Bvid           string `json:"bvid"`
							Cover          string `json:"cover"`
							Desc           string `json:"desc"`
							DisablePreview int    `json:"disable_preview"`
							DurationText   string `json:"duration_text"`
							JumpURL        string `json:"jump_url"`
							Stat           struct {
								Danmaku string `json:"danmaku"`
								Play    string `json:"play"`
							} `json:"stat"`
							Title string `json:"title"`
							Type  int    `json:"type"`
						} `json:"archive"`
						Opus struct {
							FoldAction []string `json:"fold_action"`
							JumpURL    string   `json:"jump_url"`
							Pics       []struct {
								Height  int         `json:"height"`
								LiveURL interface{} `json:"live_url"`
								Size    float64     `json:"size"`
								URL     string      `json:"url"`
								Width   int         `json:"width"`
							} `json:"pics"`
							Summary struct {
								RichTextNodes []struct {
									OrigText string `json:"orig_text"`
									Text     string `json:"text"`
									Type     string `json:"type"`
								} `json:"rich_text_nodes"`
								Text string `json:"text"`
							} `json:"summary"`
							Title interface{} `json:"title"`
						} `json:"opus"`
						Type string `json:"type"`
					} `json:"major"`
					Topic interface{} `json:"topic"`
				} `json:"module_dynamic"`
				ModuleMore struct {
					ThreePointItems []struct {
						Label string `json:"label"`
						Type  string `json:"type"`
					} `json:"three_point_items"`
				} `json:"module_more"`
			} `json:"modules"`
			Type    string `json:"type"`
			Visible bool   `json:"visible"`
			Basic0  Basic
		} `json:"items"`
	} `json:"data"`
}
type User struct {
	gorm.Model
	Name       string
	UserID     string
	Fans       int
	LastActive int
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
	ToId       string
	ToName     string
	LiveRoom   string
	ActionName string
	GiftName   string
	GiftPrice  float64 `gorm:"scale:2;precision:7"`
	GiftAmount int
	Extra      string
}
type RoomInfo struct {
	Data struct {
		LiveTime string `json:"live_time"`
		UID      int    `json:"uid"`
		Title    string `json:"title"`
		Area     string `json:"area"`
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
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		TriggerTime int64  `json:"trigger_time"`
		UID         int    `json:"uid"`
		Uname       string `json:"uname"`
	} `json:"data"`
}

func UpdateCommon() {
	for i := range Followings {
		var id = Followings[i].UserID
		res, _ := client.R().Get("https://api.bilibili.com/x/relation/stat?vmid=" + id)
		var state = UserState{}
		sonic.Unmarshal(res.Body(), &state)
		var user = User{}
		user.Fans = state.Data.Following
	}
}

func RefreshCookie() {
	resp, err := client.R().Get("https://space.bilibili.com/504140200/dynamic")
	if err != nil {
		panic(err)
	}
	var cookie = resp.Header().Get("Set-Cookie")
	Cookie = strings.Split(cookie, ";")[0]
	config.Cookie = Cookie
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
func PushDynamic(title, msg string) {

	for i := range config.ReportTo {
		if config.EnableQQBot {
			var qq = config.ReportTo[i]
			var body = `{
  "message_type": "private",
  "user_id": #to,
  "message": {
    "type": "text",
    "data": {"text":"#msg"}
  }
}`
			body = strings.Replace(body, "#to", qq, 1)
			body = strings.Replace(body, "#msg", msg, 1)
			res, err := client.R().SetBody([]byte(body)).Post(config.BackServer + "/send_msg")
			if err != nil {
				log.Println(err)
			}
			if strings.Contains(string(res.Body()), "ok") {
				log.Println("发送QQ信息：" + msg)
			}
		}

	}
	if config.EnableEmail {
		params := &resend.SendEmailRequest{
			From:    config.FromMail,
			To:      config.ToMail,
			Subject: title,
			Html:    msg,
		}
		_, err := mailClient.Emails.Send(params)
		if err != nil {
			log.Println(err)
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
func FixPrice() {
	var actions []LiveAction
	db.Where("action_name = ? AND gift_price = ?", "gift", 0).Find(&actions)

	for _, action := range actions {
		action.GiftPrice = float64(GiftPrice[action.GiftName] * float32(action.GiftAmount))
		db.Save(&action) // 分别更新每条记录
	}
}
func TraceLive(roomId string) {

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
				action.ToName = liver
				action.LiveRoom = roomId
				action.ToId = liverId
				if strings.Contains(obj, "DANMU_MSG") {
					var text = LiveText{}
					sonic.Unmarshal(msgData, &text)
					action.ActionName = "msg"
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = strconv.Itoa(int(text.Info[2].([]interface{})[0].(float64)))
					action.Extra = text.Info[1].(string)
					db.Create(&action)
					log.Println(text.Info[2].([]interface{})[1].(string) + "  " + text.Info[1].(string))

				} else if strings.Contains(obj, "SEND_GIFT") {
					var info = GiftInfo{}
					sonic.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.FromId = strconv.Itoa(info.Data.SenderUinfo.UID)
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					action.GiftPrice = price
					action.GiftAmount = info.Data.Num
					db.Create(&action)
					log.Printf("%s 赠送了 %d 个 %s，%.2f元", info.Data.Uname, info.Data.Num, info.Data.GiftName, price)
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

				} else if strings.Contains(obj, "PREPARATION") {

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
func UpdateSpecial() {
	var flag = false
	if len(RecordedDynamic) == 0 {
		flag = true
	}
	for i := range config.SpecialList {
		var id = config.SpecialList[i]
		resp, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + strconv.Itoa(id) + "&timezone_offset=-480&features=itemOpusStyle")
		var result UserDynamic
		sonic.Unmarshal(resp.Body(), &result)
		for i2 := range result.Data.Items {
			var item = result.Data.Items[i2]
			if flag {
				RecordedDynamic = append(RecordedDynamic, item.IDStr)
			} else {
				var ShouldPush = true
				for i3 := range RecordedDynamic {
					var item1 = RecordedDynamic[i3]
					if item1 == item.IDStr {
						ShouldPush = false
					}
				}
				if ShouldPush {
					RecordedDynamic = append(RecordedDynamic, item.IDStr)
					log.Println(item)
				}
			}
		}

	}
}
func RefreshFollowings() {

	Followings = make([]User, 0)
	var page = 1
	Special = make([]User, 0)
	for true {
		resp, err := client.R().Get("https://line3-h5-mobile-api.biligame.com/game/center/h5/user/relationship/following_list?vmid=" + string(config.User) + "&ps=50&pn=" + strconv.Itoa(page))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(resp)
		var list = FansList{}
		sonic.Unmarshal(resp.Body(), &list)
		var users = list.Data.List
		for i := 0; i < len(users); i++ {
			var user = User{}
			user.Name = users[i].Uname
			user.UserID = users[i].Mid
			Followings = append(Followings, user)
			for j := 0; j < len(config.SpecialList); j++ {
				if strconv.Itoa(config.SpecialList[j]) == user.UserID {
					Special = append(Special, user)
				}
			}
		}
		if len(users) == 0 {
			break
		}
		page++
	}
}

var config = Config{}
var Followings = make([]User, 0)
var db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

func main() {
	db.AutoMigrate(&Live{})
	db.AutoMigrate(&LiveAction{})
	content, err := os.ReadFile("config.json")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	if err != nil {
		content = []byte("")

		config.SpecialDelay = "2m"
		config.CommonDelay = "30m"
		config.User = "451537183"
		config.RefreshFollowingsDelay = "30m"
		config.SpecialList = []int{504140200}
		fmt.Println("请输入Cookie，如不需要登陆按回车即可")

		var cookie = ""
		fmt.Scanln(&cookie)

		if len(cookie) == 0 {
			RefreshCookie()
			config.LoginMode = false
		} else {
			config.Cookie = cookie

		}
		config.LoginMode = true
		config.EnableQQBot = true
		config.EnableEmail = false
		config.FromMail = "bili@ikun.dev"
		config.ToMail = []string{"3212329718@qq.com"}
		config.ReportTo = []string{"3212329718"}
		config.BackServer = "http://127.0.0.1:3090"
		Cookie = config.Cookie
		content, _ = sonic.Marshal(&config)
		os.Create("config.json")
		os.WriteFile("config.json", content, 666)
	}
	err = sonic.Unmarshal(content, &config)
	FillGiftPrice("761662")
	FixPrice()
	TraceLive("761662")
	c := cron.New()
	c.AddFunc("@every 2m", func() { UpdateSpecial() })

	UpdateSpecial()
	c.Start()
	select {}
}
