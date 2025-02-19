package main

import (
	"database/sql"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log"
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
	Tracing                []string
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
			IDStr   string      `json:"id_str"`
			Orig    UserDynamic `json:"orig"`
			Modules struct {
				ModuleDynamic struct {
					Major struct {
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
							Pics []struct {
								URL string `json:"url"`
							} `json:"pics"`
							Summary struct {
								Text string `json:"text"`
							} `json:"summary"`
						} `json:"opus"`
						Desc struct {
							Text string `json:"text"`
						} `json:"desc"`
						Type string `json:"type"`
					} `json:"major"`
					Topic interface{} `json:"topic"`
					Desc  struct {
						Nodes []struct {
							Text string `json:"text"`
						} `json:"rich_text_nodes"`
					} `json:"desc"`
				} `json:"module_dynamic"`
				ModuleAuthor struct {
					Name string `json:"name"`
					Mid  int64  `json:"mid"`
				} `json:"module_author"`
			} `json:"modules"`
			Type string `json:"type"`
		} `json:"items"`
	} `json:"data"`
}
type User struct {
	gorm.Model
	Name   string
	UserID string
	Fans   int
}

type Status struct {
	Live       bool
	LastActive int64
	UName      string
	UID        string
	Area       string
	Title      string
}

type GuardResponse struct {
	Data struct {
		List []struct {
			UID      int64  `json:"uid"`
			UserName string `json:"username"`
		} `json:"guard_top_list"`
	} `json:"data"`
}

type Guard struct {
}

type Archive struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UName     string
	UID       int64
	Images    string
	Type      string
	Title     string
	Text      string
}

func UpdateCommon() {
	for i := range Followings {
		if i > len(Followings)-1 {
			continue
		}
		var id = Followings[i].UserID
		res, _ := client.R().Get("https://api.bilibili.com/x/relation/stat?vmid=" + id)
		var state = UserState{}
		sonic.Unmarshal(res.Body(), &state)
		var user = User{}
		user.Fans = state.Data.Follower
		user.UserID = Followings[i].UserID
		user.Name = Followings[i].Name
		db.Save(&user)
		time.Sleep(3 * time.Second)
	}

	/*

		for i := range config.Tracing {
			var id = config.Tracing[i]
			var live0 = Live{}
			db.Model(&Live{}).Where("user_id = ?", id).Find(&live0)
			if live0.RoomId != 0 {

			}
		}

	*/
}

func UpdateGuard() {

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

func FixPrice() {
	var actions []LiveAction
	db.Where("action_name = ? AND gift_price = ?", "gift", 0).Find(&actions)

	for _, action := range actions {
		action.GiftPrice = sql.NullFloat64{Float64: float64(GiftPrice[action.GiftName] * float32(action.GiftAmount.Int16)), Valid: true}
		db.Save(&action) // 分别更新每条记录
	}
}
func UpdateSpecial() {
	var flag = false
	if len(RecordedDynamic) == 0 {
		flag = true
	}
	for i := range config.SpecialList {
		var id = config.SpecialList[i]
		resp, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + (strconv.Itoa(id)) + "&timezone_offset=-480&features=itemOpusStyle")
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
					var Type = item.Type
					var userName = item.Modules.ModuleAuthor.Name
					var archive = Archive{}
					archive.UName = userName
					archive.UID = item.Modules.ModuleAuthor.Mid

					if Type == "DYNAMIC_TYPE_FORWARD" { //转发
						archive.Type = "f"
						PushDynamic("你关注的up主：转发了动态 "+userName, item.Modules.ModuleDynamic.Desc.Nodes[0].Text)
						db.Save(&archive)
					} else if Type == "DYNAMIC_TYPE_AV" { //发布视频
						archive.Type = "v"
						archive.Title = item.Modules.ModuleDynamic.Major.Archive.Title
						PushDynamic("你关注的up主：发布了视频 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
					} else if Type == "DYNAMIC_TYPE_DRAW" { //图文
						archive.Type = "i"
						archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
						db.Save(&archive)
						PushDynamic("你关注的up主：发布了动态 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
					} else if Type == "DYNAMIC_TYPE_WORD" { //文字
						archive.Type = "t"
						archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
						db.Save(&archive)
						PushDynamic("你关注的up主：发布了动态 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
					} else if Type == "DYNAMIC_TYPE_LIVE_RCMD" {

					} else if Type == "DYNAMIC_TYPE_COMMON_SQUARE" {

					} else {
						archive.Type = Type
						archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
						db.Save(&archive)
					}
					json := make([]byte, 0)
					sonic.Unmarshal(json, item)
					PushDynamic("动态json", string(json))

				}
			}
		}

		time.Sleep(time.Second * 10)

	}

	//log.Printf(RecordedDynamic)
}
func RefreshFollowings() {

	var Followings0 = make([]User, 0)
	var page = 1
	var Special0 = make([]User, 0)
	for true {
		resp, err := client.R().Get("https://line3-h5-mobile-api.biligame.com/game/center/h5/user/relationship/following_list?vmid=" + string(config.User) + "&ps=50&pn=" + strconv.Itoa(page))
		if err != nil {
			fmt.Println(err)
		}
		var list = FansList{}
		sonic.Unmarshal(resp.Body(), &list)
		var users = list.Data.List
		for i := 0; i < len(users); i++ {
			var user = User{}
			user.Name = users[i].Uname
			user.UserID = users[i].Mid
			Followings0 = append(Followings0, user)
			for j := 0; j < len(config.SpecialList); j++ {
				if strconv.Itoa(config.SpecialList[j]) == user.UserID {
					Special0 = append(Special0, user)
				}
			}
		}
		if len(users) == 0 {
			break
		}
		page++
	}
	Followings = Followings0
	Special = Special0
}

func SaveConfig() {
	content, _ := sonic.Marshal(&config)
	os.WriteFile("config.json", content, 666)
}

var config = Config{}
var Followings = make([]User, 0)
var db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

var lives = map[string]*Status{} //[]string{}

func main() {

	db.AutoMigrate(&Live{})
	db.AutoMigrate(&LiveAction{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Archive{})
	FixMoney()
	RemoveEmpty()
	db.Exec("PRAGMA journal_mode=WAL;")
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
	go InitHTTP()
	for i := range config.Tracing {
		var roomId = config.Tracing[i]

		lives[roomId] = &Status{}
		go TraceLive(config.Tracing[i])
		time.Sleep(30 * time.Second)

	}
	c := cron.New()
	RefreshFollowings()
	UpdateCommon()
	c.AddFunc("@every 2m", func() { UpdateSpecial() })
	c.AddFunc("@every 120m", RefreshFollowings)
	c.AddFunc("@every 10m", UpdateCommon)
	c.AddFunc("@every 5m", FixMoney)

	c.Start()

	select {}
}
