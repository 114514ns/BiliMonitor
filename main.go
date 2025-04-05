package main

import (
	"database/sql"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	url2 "net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var client = resty.New()
var Cookie = ""
var Special = make([]User, 0)
var RecordedDynamic = make([]string, 0)
var RecordedMedias = make([]string, 0)
var GiftPrice = map[string]float32{}
var mailClient = resend.NewClient("")

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

type Config struct {
	Port                    int32
	SpecialDelay            string
	CommonDelay             string
	RefreshFollowingsDelay  string
	User                    string
	SpecialList             []int
	Cookie                  string
	LoginMode               bool
	EnableEmail             bool
	ResendToken             string
	FromMail                string
	ToMail                  []string
	EnableQQBot             bool
	ReportTo                []string
	BackServer              string
	Tracing                 []string
	EnableAlist             bool
	AlistServer             string
	AlistUser               string
	AlistPass               string
	AlistPath               string
	EnableServerPush        bool
	ServerPushKey           string
	EnableLiveBackup        bool
	MikuPath                string
	EnableSQLite            bool
	SQLitePath              string
	EnableMySQL             bool
	SQLName                 string
	SQLUser                 string
	SQLPass                 string
	SQLServer               string
	CodeToMP4               bool
	SplitAudio              bool
	EnableCollectionMonitor bool
}

type User struct {
	gorm.Model
	Name   string
	UserID string
	Fans   int
	Face   string
}
type Video struct {
	Title       string
	Desc        string
	Author      string
	UID         int64
	Img         string
	BV          string
	PublishAt   string
	AuthorFace  string
	Cid         int
	Duration    string
	Part        int
	ParentTitle string
}

type Status struct {
	Live           bool
	LastActive     int64
	UName          string
	UID            string
	Area           string
	Title          string
	StartAt        string
	RemainTrying   int8
	Face           string
	Cover          string
	LiveRoom       string
	Stream         string
	StreamCacheKey int64
	OnlineWatcher  []Watcher
	OnlineCount    int
	GuardList      []Watcher
	GuardCacheKey  int64
	Danmuku        []FrontLiveAction `json:"-"`
	GuardCount     int
	StreamSplits   []string
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

func CheckConfig() {
	if config.EnableAlist {
		if GetAlistToken() == "" {
			log.Fatal("Alist密码错误")
		}
	}
	if config.EnableLiveBackup {
		dir, err := ioutil.ReadDir(config.MikuPath)
		if err != nil {
			log.Fatal("Miku录播姬路径不存在")
		}
		var found = false
		for _, info := range dir {
			if info.Name() == "config.json" {
				found = true
			}
		}
		if !found {
			log.Fatal("Miku录播姬路径错误")
		}

	}
	if config.CodeToMP4 || config.SplitAudio {
		var cmd = exec.Command("ffmpeg")
		if !strings.Contains(cmd.String(), "FFmpeg developers") {
			log.Fatal("未找到ffmpeg")
		}
	}
	if config.EnableLiveBackup && !config.EnableAlist {
		log.Fatal("直播备份需要配合Alist使用")
	}
}

func UpdateGuard() {

}
func RefreshCookie() {
	//var url = "https://www.bilibili.com/correspond/1/" + getCorrespondPath(time.Now().UnixMilli())
	//_, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", USER_AGENT).Get(url)

}
func GetDefaultCookie() {
	resp, err := client.R().Get("https://space.bilibili.com/1265680561/dynamic")
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

func UploadLive(live Live) {

	var debug = true
	time.Sleep(60 * time.Second)
	var dir = config.MikuPath + "/" + strconv.Itoa(live.RoomId) + "-" + live.UserName
	var flv, t, _ = Last(dir)
	os.MkdirAll("cache", 0777)
	if time.Now().Unix()-t.Unix() < 60000000 {
		var file = dir + "/" + flv
		log.Println(config.AlistPath + "Live/" + live.UserName + "/" + time.Now().Format(time.DateTime) + "/")
		split := strings.Split(file, "-")
		var ext = "flv"
		var title = strings.Replace(split[len(split)-1], ".flv", "", 10)
		var uuid = uuid.New().String() + ".mp4"

		if config.CodeToMP4 {
			file = dir + "/" + flv
			cmd := exec.Command("ffmpeg", "-i", file, "-vcodec", "copy", "-acodec", "copy", "cache/"+uuid)
			cmd.Run()
			out, _ := cmd.CombinedOutput()
			if debug {
				fmt.Println(string(out))
			}
			ext = "mp4"
			file = "cache/" + uuid
		}
		var alistName = config.AlistPath + "Live/" + live.UserName + "/" + strings.Replace(time.Now().Format(time.DateTime), ":", "-", 3) + "/" + title + "." + ext
		if config.SplitAudio {
			file = dir + "/" + flv
			var auido = strings.Replace("cache/"+uuid, "."+ext, ".mp3", 1)
			cmd := exec.Command("ffmpeg", "-i", file, "-vn", auido)
			cmd.Run()
			output, _ := cmd.CombinedOutput()
			if debug {
				fmt.Println(string(output))
			}
			UploadFile(auido, strings.Replace(alistName, "."+ext, ".mp3", 1))

			os.Remove(auido)
		}

		UploadFile(file, alistName)
		os.Remove(file)
	}
}

func ParseSingleVideo(bv string) (result []Video) {
	res, _ := client.R().
		SetHeader("Referer", "https://www.bilibili.com/").
		SetHeader("Cookie", config.Cookie).
		Get("https://api.bilibili.com/x/web-interface/view?bvid=" + bv)

	var resObj = VideoResponse{}
	sonic.Unmarshal(res.Body(), &resObj)
	fmt.Println(string(res.Body()))

	var array = []Video{}

	for i, item := range resObj.Data.Pages {
		var video = Video{}
		video.Author = resObj.Data.Owner.Name
		video.ParentTitle = resObj.Data.Title
		video.BV = bv
		video.Desc = resObj.Data.Desc
		video.Title = item.Title
		video.Part = i + 1
		video.Cid = item.Cid
		video.Duration = FormatDuration(item.Duration)
		video.PublishAt = time.Unix(resObj.Data.PublishAt, 0).Format(time.DateTime)
		video.Img = resObj.Data.Cover
		video.UID = resObj.Data.Owner.Mid
		video.AuthorFace = resObj.Data.Owner.Face
		array = append(array, video)
	}

	return array
}

func ParsePlayList(mid string, session string) []Video {
	var array []Video
	var page = 1
	var user = FetchUser(mid)
	for true {
		var url = "https://api.bilibili.com/x/polymer/web-space/seasons_archives_list?mid=" + mid + "&season_id=" + session + "&page_num=" + strconv.Itoa(page) + "&page_size=30"
		res, _ := client.R().SetHeader("Referer", "https://www.bilibili.com/").SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).Get(url)
		var playList = PlayListResponse{}
		sonic.Unmarshal(res.Body(), &playList)
		if len(playList.Data.Archives) == 0 {
			break
		}
		for _, archive := range playList.Data.Archives {
			var video = Video{}
			video.Cid = 0
			video.Duration = FormatDuration(archive.Duration)
			video.Img = archive.Cover
			video.BV = archive.BV
			video.Title = archive.Title
			video.ParentTitle = playList.Data.Meta.Name
			video.UID = toInt64(mid)
			video.Part = 1
			video.Author = user.Name
			video.AuthorFace = user.Face
			video.PublishAt = time.Unix(int64(archive.CreateAt), 0).Format(time.DateTime)
			video.Desc = ""
			array = append(array, video)
		}
		page++
	}
	return array

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
var file = time.Now().Format(time.DateTime) + ".log"
var logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
var wbi = NewDefaultWbi()

const ENV = "DEV"

func main() {
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	content, err := os.ReadFile("config.json")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	if err != nil {
		content = []byte("")

		config.SpecialDelay = "2m"
		config.CommonDelay = "30m"
		config.User = "2"
		config.RefreshFollowingsDelay = "30m"
		config.SpecialList = []int{}
		config.EnableQQBot = false
		config.EnableEmail = true
		config.FromMail = "bili@ikun.dev"
		config.ToMail = []string{"to@example.com"}
		config.ReportTo = []string{"10001"}
		config.BackServer = "http://127.0.0.1:3090"
		config.Tracing = []string{"544853"}
		config.EnableAlist = false
		config.EnableSQLite = true
		config.SQLitePath = "database.db"
		config.EnableMySQL = false
		config.EnableCollectionMonitor = false
		config.Port = 8081
		Cookie = config.Cookie
		content, _ = sonic.Marshal(&config)
		os.Create("config.json")
		os.WriteFile("config.json", content, 666)
	}
	err = sonic.Unmarshal(content, &config)
	mailClient = resend.NewClient(config.ResendToken)
	if config.EnableSQLite {
		db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
		db.Exec("PRAGMA journal_mode=WAL;")
	}
	if config.EnableMySQL {
		var dsl = "#user:#pass@tcp(#server)/#name?charset=utf8mb4&parseTime=True&loc=Local"
		dsl = strings.Replace(dsl, "#user", config.SQLUser, 1)
		dsl = strings.Replace(dsl, "#pass", config.SQLPass, 1)
		dsl = strings.Replace(dsl, "#server", config.SQLServer, 1)
		dsl = strings.Replace(dsl, "#name", config.SQLName, 1)

		db, _ = gorm.Open(mysql.New(mysql.Config{
			DSN: dsl, // DSN data source name
		}), &gorm.Config{})
	}
	wbi.WithRawCookies(config.Cookie)
	wbi.initWbi()
	db.AutoMigrate(&Live{})
	db.AutoMigrate(&LiveAction{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Archive{})
	RemoveEmpty()
	go InitHTTP()
	for i := range config.Tracing {
		var roomId = config.Tracing[i]
		lives[roomId] = &Status{RemainTrying: 4}
		lives[roomId].Danmuku = make([]FrontLiveAction, 0)
		lives[roomId].OnlineWatcher = make([]Watcher, 0)
		lives[roomId].GuardList = make([]Watcher, 0)
		lives[roomId].Stream = GetLiveStream(roomId)
		go RecordStream(roomId)
		go TraceLive(config.Tracing[i])
		time.Sleep(30 * time.Second)

	}

	c := cron.New()
	RefreshFollowings()
	UpdateCommon()
	c.AddFunc("@every 2m", func() { UpdateSpecial() })
	c.AddFunc("@every 120m", RefreshFollowings)
	c.AddFunc("@every 10m", UpdateCommon)
	c.AddFunc("@every 1m", FixMoney)
	//c.AddFunc("@every 1m", func() { RefreshCollection(strconv.Itoa(collectId)) })

	c.Start()

	select {}
}
