package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/html"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	url2 "net/url"
	"os"
	"os/exec"
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
	EnableAlist            string
	AlistServer            string
	AlistUser              string
	AlistPass              string
	AlistPath              string
	EnableServerPush       bool
	ServerPushKey          string
	EnableLiveBackup       bool
	MikuPath               string
	EnableSQLite           bool
	SQLitePath             string
	EnableMySQL            bool
	SQLName                string
	SQLUser                string
	SQLPass                string
	SQLServer              string
	CodeToMP4              bool
	SplitAudio             bool
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
type DynamicItem struct {
	IDStr   string       `json:"id_str"`
	Orig    *DynamicItem `json:"orig"`
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
					Bvid  string `json:"bvid"`
					Cover string `json:"cover"`
					Desc  string `json:"desc"`
					Stat  struct {
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
			Name      string `json:"name"`
			Mid       int64  `json:"mid"`
			TimeStamp int64  `json:"pub_ts"`
		} `json:"module_author"`
	} `json:"modules"`
	Type string `json:"type"`
}
type UserDynamic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Items []DynamicItem `json:"items"`
	} `json:"data"`
}
type User struct {
	gorm.Model
	Name   string
	UserID string
	Fans   int
}
type Video struct {
	Data struct {
	} `json:"data"`
}
type Status struct {
	Live       bool
	LastActive int64
	UName      string
	UID        string
	Area       string
	Title      string
	StartAt    string
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
	BiliID    string
}
type Dash struct {
	Data struct {
		Dash0 struct {
			Video []struct {
				Link string `json:"base_url"`
			} `json:"video"`
			Audio []struct {
				Link string `json:"base_url"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
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

	if config.EnableServerPush {
		var url = fmt.Sprintf(config.ServerPushKey+"?title=%s&desp=%s", url2.QueryEscape(title), url2.QueryEscape(msg))
		client.R().Get(url)
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
func GetAlistToken() string {
	type LoginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var sum = sha256.Sum256([]byte(config.AlistPass + "-https://github.com/alist-org/alist"))
	var req = LoginRequest{Username: config.AlistUser, Password: hex.EncodeToString(sum[:])}
	alist, _ := client.R().SetBody(req).Post(config.AlistServer + "api/auth/login/hash")

	var res = LoginResponse{}
	sonic.Unmarshal(alist.Body(), &res)
	return res.Data.Token
}
func UploadFile(path string, alistPath string) {
	file, _ := os.Open(path)
	fi, _ := file.Stat()
	res, _ := client.R().
		SetHeader("Authorization", GetAlistToken()).
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("Content-Length", strconv.FormatInt(fi.Size(), 10)).
		SetHeader("File-Path", alistPath).
		SetFile("file", path).
		Put(config.AlistServer + "api/fs/form")

	log.Println("[" + alistPath + "   ]" + res.String())
}
func UploadArchive(bv string, cid string) {
	os.Mkdir("cache", 066)
	var videolink = "https://bilibili.com/video/" + bv
	vRes, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com").SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36").Get(videolink)
	htmlContent := vRes.Body()
	reader := bytes.NewReader(htmlContent)
	root, _ := html.Parse(reader)
	find := goquery.NewDocumentFromNode(root).Find("script")
	find.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "m4s") {
			var json = strings.Replace(s.Text(), "window.__playinfo__=", "", 1)
			var v = Dash{}
			sonic.Unmarshal([]byte(json), &v)
			audio, _ := client.R().SetDoNotParseResponse(true).SetHeader("Referer", "https://www.bilibili.com").Get(v.Data.Dash0.Audio[0].Link)
			//defer audio.RawBody().Close()
			os.WriteFile("cache/"+bv+".mp3", audio.Body(), 066)
			audioFile, _ := os.Create("cache/" + bv + ".mp3")
			//defer audioFile.Close()
			io.Copy(audioFile, audio.RawBody())

			video, _ := client.R().SetDoNotParseResponse(true).SetHeader("Referer", "https://www.bilibili.com").Get(v.Data.Dash0.Video[0].Link)
			//defer video.RawBody().Close()
			videoFile, _ := os.Create("cache/" + bv + ".m4s")
			//defer videoFile.Close()
			io.Copy(videoFile, video.RawBody())
			cmd := exec.Command("ffmpeg", "-i", videoFile.Name(), "-i", audioFile.Name(), "-vcodec", "copy", "-acodec", "copy", "cache/"+bv+".mp4")
			out, _ := cmd.CombinedOutput()
			cmd.Run() // 执行命令

			log.Println(string(out))
			UploadFile("cache/"+bv+".mp4", config.AlistPath+bv+".mp4")

			os.Remove("cache/" + bv + ".mp4")
			os.Remove("cache/" + bv + ".mp3")
			os.Remove("cache/" + bv + ".m4s")
		}
	})

}

func UploadLive(live Live) {
	time.Sleep(180 * time.Second)
	var dir = config.MikuPath + "/" + strconv.Itoa(live.RoomId) + "-" + live.UserName
	var flv, t, _ = Last(dir)
	os.MkdirAll("cache", 0777)
	if time.Now().Unix()-t.Unix() < 600 {
		var file = dir + "/" + flv
		log.Println(config.AlistPath + "Live/" + live.UserName + "/" + time.Now().Format(time.DateTime) + "/")
		split := strings.Split(file, "-")
		var ext = "flv"
		var title = strings.Replace(split[len(split)-1], ".flv", "", 10)
		var uuid = uuid.New().String() + ".mp4"

		if config.CodeToMP4 {
			file = dir + "/" + flv
			exec.Command("ffmpeg", "-i", file, "-vcodec", "copy", "-acodec", "copy", "cache/"+uuid)
			ext = "mp4"
			file = "cache/" + uuid
		}
		var alistName = config.AlistPath + "Live/" + live.UserName + "/" + strings.Replace(time.Now().Format(time.DateTime), ":", "-", 3) + "/" + title + "." + ext
		if config.SplitAudio {
			file = dir + "/" + flv
			var auido = strings.Replace("cache/"+uuid, "."+ext, ".mp3", 1)
			exec.Command("ffmpeg", "-i", file, "-vn", auido)
			UploadFile(auido, strings.Replace(alistName, ext, "mp3", 1))
		}

		UploadFile(file, alistName)
	}
}
func ParseDynamic(item DynamicItem, push bool) (Archive, Archive) {
	var Type = item.Type
	var orig = Archive{}
	var userName = item.Modules.ModuleAuthor.Name
	var archive = Archive{}
	archive.UName = userName
	archive.UID = item.Modules.ModuleAuthor.Mid
	archive.CreatedAt = time.Unix(item.Modules.ModuleAuthor.TimeStamp+8*3600, 0)
	if Type == "DYNAMIC_TYPE_FORWARD" { //转发
		archive.Type = "f"
		archive.BiliID = item.IDStr
		var txt = ""
		for _, node := range item.Modules.ModuleDynamic.Desc.Nodes {
			txt = txt + node.Text
			txt = txt + "\n"
		}
		orig, _ = ParseDynamic(*item.Orig, false)
		archive.Text = txt
		if push {
			PushDynamic("你关注的up主：转发了动态 "+userName, item.Modules.ModuleDynamic.Desc.Nodes[0].Text)
		}
	} else if Type == "DYNAMIC_TYPE_AV" { //发布视频
		archive.Type = "v"
		archive.BiliID = item.IDStr
		archive.Title = item.Modules.ModuleDynamic.Major.Archive.Title
		if push {
			PushDynamic("你关注的up主：发布了视频 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
			go UploadArchive(item.Modules.ModuleDynamic.Major.Archive.Bvid, "")
		}

	} else if Type == "DYNAMIC_TYPE_DRAW" { //图文
		archive.Type = "i"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Desc.Text
		if push {
			PushDynamic("你关注的up主：发布了动态 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
		}

	} else if Type == "DYNAMIC_TYPE_WORD" { //文字
		archive.Type = "t"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
		if push {
			PushDynamic("你关注的up主：发布了动态 "+userName, item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
		}
	} else if Type == "DYNAMIC_TYPE_LIVE_RCMD" {

	} else if Type == "DYNAMIC_TYPE_COMMON_SQUARE" {

	} else {
		archive.Type = Type
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
	}
	return archive, orig
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
					d, _ := ParseDynamic(item, true)
					db.Save(&d)
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
func FetchArchive(mid string, page int, size int) {
	var url = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?host_mid=#id&ps=#ps&pn=#pn"
	url = strings.Replace(url, "#id", mid, 1)
	url = strings.Replace(url, "#ps", strconv.Itoa(size), 1)
	url = strings.Replace(url, "#pn", strconv.Itoa(page), 1)

	res, _ := client.R().SetHeader("Cookie", config.Cookie).
		SetHeader("Referer", "https://www.bilibili.com/").Get(url)

	dynamic := UserDynamic{}
	sonic.Unmarshal(res.Body(), &dynamic)
	for _, item := range dynamic.Data.Items {
		var d, d1 = ParseDynamic(item, false)
		db.Save(&d)
		if d1.BiliID != "" {
			db.Save(&d1)
		}

	}
	log.Println(dynamic)

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
var db *gorm.DB

// var db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

var lives = map[string]*Status{} //[]string{}

func main() {

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
	if config.EnableSQLite {
		db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	}
	if config.EnableMySQL {
		var dsl = "#user:#pass@tcp(#server)/#name?charset=utf8&parseTime=True&loc=Local"
		dsl = strings.Replace(dsl, "#user", config.SQLUser, 1)
		dsl = strings.Replace(dsl, "#pass", config.SQLPass, 1)
		dsl = strings.Replace(dsl, "#server", config.SQLServer, 1)
		dsl = strings.Replace(dsl, "#name", config.SQLName, 1)
		db, _ = gorm.Open(mysql.New(mysql.Config{
			DSN: dsl, // DSN data source name
		}), &gorm.Config{})
	}
	db.AutoMigrate(&Live{})
	db.AutoMigrate(&LiveAction{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Archive{})
	db.Exec("PRAGMA journal_mode=WAL;")
	FixMoney()
	RemoveEmpty()
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
