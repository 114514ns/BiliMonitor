package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/jordan-wright/email"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"io"
	"log"
	"net/smtp"
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
var GiftPrice = map[string]int{}

type Config struct {
	SpecialDelay           string
	CommonDelay            string
	RefreshFollowingsDelay string
	User                   string
	SpecialList            []int
	Cookie                 string
	LoginMode              bool
	EnableEmail            bool
	SMTPServer             string
	FromMail               string
	Code                   string
	ToMail                 string
	EnableQQBot            bool
	ReportTo               []string
	BackServer             string
}

type FansList struct {
	Code int `json:"code"`
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
		HasMore bool `json:"has_more"`
		Items   []struct {
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
				ModuleStat struct {
					Comment struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
					} `json:"comment"`
					Forward struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
					} `json:"forward"`
					Like struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
						Status    bool `json:"status"`
					} `json:"like"`
				} `json:"module_stat"`
				ModuleTag struct {
					Text string `json:"text"`
				} `json:"module_tag"`
			} `json:"modules"`
			Type    string `json:"type"`
			Visible bool   `json:"visible"`
			Basic0  Basic
		} `json:"items"`
		Offset         string `json:"offset"`
		UpdateBaseline string `json:"update_baseline"`
		UpdateNum      int    `json:"update_num"`
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
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
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
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		GiftConfig struct {
			BaseConfig struct {
				Hited bool `json:"hited"`
				List  []struct {
					ID                int    `json:"id"`
					Name              string `json:"name"`
					Price             int    `json:"price"`
					Type              int    `json:"type"`
					CoinType          string `json:"coin_type"`
					BagGift           int    `json:"bag_gift"`
					Effect            int    `json:"effect"`
					CornerMark        string `json:"corner_mark"`
					CornerBackground  string `json:"corner_background"`
					Broadcast         int    `json:"broadcast"`
					Draw              int    `json:"draw"`
					StayTime          int    `json:"stay_time"`
					AnimationFrameNum int    `json:"animation_frame_num"`
					Desc              string `json:"desc"`
					Rule              string `json:"rule"`
					Rights            string `json:"rights"`
					PrivilegeRequired int    `json:"privilege_required"`
					CountMap          []struct {
						Num            int    `json:"num"`
						Text           string `json:"text"`
						Desc           string `json:"desc"`
						WebSvga        string `json:"web_svga"`
						VerticalSvga   string `json:"vertical_svga"`
						HorizontalSvga string `json:"horizontal_svga"`
						SpecialColor   string `json:"special_color"`
						EffectID       int    `json:"effect_id"`
					} `json:"count_map"`
					ImgBasic             string      `json:"img_basic"`
					ImgDynamic           string      `json:"img_dynamic"`
					FrameAnimation       string      `json:"frame_animation"`
					Gif                  string      `json:"gif"`
					Webp                 string      `json:"webp"`
					FullScWeb            string      `json:"full_sc_web"`
					FullScHorizontal     string      `json:"full_sc_horizontal"`
					FullScVertical       string      `json:"full_sc_vertical"`
					FullScHorizontalSvga string      `json:"full_sc_horizontal_svga"`
					FullScVerticalSvga   string      `json:"full_sc_vertical_svga"`
					BulletHead           string      `json:"bullet_head"`
					BulletTail           string      `json:"bullet_tail"`
					LimitInterval        int         `json:"limit_interval"`
					BindRuid             int         `json:"bind_ruid"`
					BindRoomid           int         `json:"bind_roomid"`
					GiftType             int         `json:"gift_type"`
					ComboResourcesID     int         `json:"combo_resources_id"`
					MaxSendLimit         int         `json:"max_send_limit"`
					Weight               int         `json:"weight"`
					GoodsID              int         `json:"goods_id"`
					HasImagedGift        int         `json:"has_imaged_gift"`
					LeftCornerText       string      `json:"left_corner_text"`
					LeftCornerBackground string      `json:"left_corner_background"`
					GiftBanner           interface{} `json:"gift_banner"`
					DiyCountMap          int         `json:"diy_count_map"`
					EffectID             int         `json:"effect_id"`
					FirstTips            string      `json:"first_tips"`
					GiftAttrs            []int       `json:"gift_attrs"`
					CornerMarkColor      string      `json:"corner_mark_color"`
					CornerColorBg        string      `json:"corner_color_bg"`
					WebLight             struct {
						CornerMark       string `json:"corner_mark"`
						CornerBackground string `json:"corner_background"`
						CornerMarkColor  string `json:"corner_mark_color"`
						CornerColorBg    string `json:"corner_color_bg"`
					} `json:"web_light"`
					WebDark struct {
						CornerMark       string `json:"corner_mark"`
						CornerBackground string `json:"corner_background"`
						CornerMarkColor  string `json:"corner_mark_color"`
						CornerColorBg    string `json:"corner_color_bg"`
					} `json:"web_dark"`
				} `json:"list"`
				Version int64 `json:"version"`
				TTL     int64 `json:"ttl"`
			} `json:"base_config"`
		} `json:"gift_config"`
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
	resp, err := client.R().Get("https://space.bilibili.com/1265680561/dynamic")
	if err != nil {
		panic(err)
	}
	var cookie = resp.Header().Get("Set-Cookie")
	Cookie = strings.Split(cookie, ";")[0]
	config.Cookie = Cookie
}
func FillGiftPrice() {
	res, _ := client.R().Get("https://api.live.bilibili.com/xlive/web-room/v1/giftPanel/roomGiftList?platform=pc&room_id=30849380")
	var gift = GiftList{}
	sonic.Unmarshal(res.Body(), &gift)
	for i := range gift.Data.GiftConfig.BaseConfig.List {
		var item = gift.Data.GiftConfig.BaseConfig.List[i]
		GiftPrice[item.Name] = item.Price / 1000
	}
}
func PushDynamic(msg string) {

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
		e := email.NewEmail()
		e.From = config.FromMail
		e.To = []string{config.FromMail}
		e.Subject = msg
		e.Text = []byte("Text Body is, of course, supported!")
		err := e.SendWithStartTLS(config.SMTPServer, smtp.PlainAuth("", config.FromMail, config.Code, strings.Split(config.SMTPServer, ":")[0]), &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Fatal(err)
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
func TraceLive(roomId string) {

	var url0 = "https://api.bilibili.com/x/web-interface/nav"
	url0 = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + roomId
	res, _ := client.R().SetHeader("Cookie", config.Cookie).Get(url0)
	var liveInfo = LiveInfo{}
	sonic.Unmarshal(res.Body(), &liveInfo)

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

		var print = 0
		var send = 0
		for {
			// 读取从服务端传来的消息
			_, message, err := c.ReadMessage()
			send++
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

			if !strings.Contains(msg, "[object") {
				print++
				log.Printf("Received: %s", msg)

			}
			log.Println("" + strconv.Itoa(print) + "" + strconv.Itoa(send))
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
			log.Println("Send heart")
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
					//PushDynamic()
				}
			}
		}

		fmt.Println(result)
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

func main() {

	content, err := os.ReadFile("config.json")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	FillGiftPrice()
	if err != nil {
		content = []byte("")

		config.SpecialDelay = "2m"
		config.CommonDelay = "30m"
		config.User = "2"
		config.RefreshFollowingsDelay = "30m"
		config.SpecialList = []int{1265680561}
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
		config.FromMail = "from@example.com"
		config.SMTPServer = "smtp.office365.com:587"
		config.ToMail = "to@example.com"
		config.Code = "rhybghptfswxlwbk"
		config.ReportTo = []string{"10001"}
		config.BackServer = "http://127.0.0.1:3090"
		Cookie = config.Cookie
		content, _ = sonic.Marshal(&config)
		os.Create("config.json")
		os.WriteFile("config.json", content, 666)
	}
	err = sonic.Unmarshal(content, &config)
	TraceLive("30849380")
	//PushDynamic("Hello World")
	c := cron.New()
	c.AddFunc("@every 2s", func() { fmt.Println("Every hour thirty") })

	//UpdateSpecial()
	c.Start()
	select {}
}
