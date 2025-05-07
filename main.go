package main

import (
	"database/sql"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	url2 "net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var client = resty.New()
var Cookie = ""
var Special = make([]User, 0)
var RecordedDynamic = make([]string, 0)
var RecordedMedias = make([]string, 0)
var GiftPrice = map[string]float32{}
var mailClient = resend.NewClient("")

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"

type Config struct {
	Port                    int32
	SpecialDelay            string
	CommonDelay             string
	RefreshFollowingsDelay  string
	User                    string
	SpecialList             []int64
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
	Mode                    string
	Slaves                  []string
	TraceArea               bool
	BlackTracing            []string
}

type User struct {
	gorm.Model
	Name   string
	UserID int64
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
	sync.RWMutex
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
type FaceCache struct {
	UID      int64
	Face     string
	UpdateAt time.Time
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
	resp, err := client.R().Get("https://bilibili.com/")
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
	var user = FetchUser(mid, nil)
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

func SortTracing() {
	var m = make(map[string]bool)
	for _, s := range config.Tracing {
		var roomUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + s
		var rRes, _ = client.R().Get(roomUrl)
		var roomInfo = RoomInfo{}
		sonic.Unmarshal(rRes.Body(), &roomInfo)
		if roomInfo.Data.LiveTime == "0000-00-00 00:00:00" {
			m[s] = false
		} else {
			m[s] = true
		}
		time.Sleep(400 * time.Millisecond)
	}
	config.Tracing = []string{}
	for s, b := range m {
		if b {
			config.Tracing = append(config.Tracing, s)
		}
	}
	for s, b := range m {
		if !b {
			config.Tracing = append(config.Tracing, s)
		}
	}
}

func SaveConfig() {
	content, _ := sonic.Marshal(&config)
	os.WriteFile("config.json", content, 666)
}

var config = Config{}
var Followings = make([]User, 0)
var db *gorm.DB

// var db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

var livesMutex sync.Mutex
var lives = map[string]*Status{} //[]string{}
var file = time.Now().Format("2006-01-02_15-04-05") + ".log"
var logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
var wbi = NewDefaultWbi()
var httpBytes int64 = 0
var websocketBytes = 0

const ENV = "DEV"

var totalRequests = 0
var launchTime = time.Now()
var USER_AGENTS = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.3",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.6",
	"Mozilla/5.0 (Linux; Android 13; SM-S908U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone14,3; U; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) Version/10.0 Mobile/19A346 Safari/602.1",
}

func CSRF() string {
	split := strings.Split(config.Cookie, ";")
	jct := ""
	for _, s := range split {
		if strings.Contains(s, "bili_jct=") {
			jct = strings.Replace(s, "bili_jct=", "", 1)
		}
	}
	jct = jct[1:]
	return jct
}

var man *SlaverManager

func main() {
	client.OnAfterResponse(func(c *resty.Client, response *resty.Response) error {
		totalRequests++
		httpBytes += response.RawResponse.ContentLength
		return nil
	})
	rand.Seed(time.Now().UnixNano())
	client.OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
		request.Header.Set("User-Agent", USER_AGENTS[rand.Uint32()%uint32(len(USER_AGENTS))])
		return nil
	})
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
		config.SpecialList = []int64{}
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
		config.Mode = "Master"
		config.TraceArea = false
		Cookie = config.Cookie
		content, _ = sonic.Marshal(&config)
		os.Create("config.json")
		os.WriteFile("config.json", content, 666)
	}
	err = sonic.Unmarshal(content, &config)
	mailClient = resend.NewClient(config.ResendToken)
	if config.EnableSQLite {
		db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{
			Logger: logger.New(
				log.New(os.Stdout, "", log.LstdFlags),
				logger.Config{
					IgnoreRecordNotFoundError: true,
				},
			),
		})
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
		}), &gorm.Config{Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: true,
			},
		)})
	}
	TotalGuards()
	wbi.WithRawCookies(config.Cookie)
	wbi.initWbi()
	db.AutoMigrate(&Live{})
	db.AutoMigrate(&LiveAction{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Archive{})
	db.AutoMigrate(&AreaLiver{})
	db.AutoMigrate(&AreaLive{})
	db.AutoMigrate(&FansClub{})
	db.AutoMigrate(&FaceCache{})
	RemoveEmpty()
	go InitHTTP()

	c := cron.New()

	if config.Mode == "Master" {
		config.Slaves = append(config.Slaves, "http://127.0.0.1:"+strconv.Itoa(int(config.Port)))
		man = NewSlaverManager(config.Slaves)
		man.OnErr = func(tasks []string) {
			log.Println("onError")
		}
		RecoverLive()
		go func() {
			RefreshFollowings()
			UpdateCommon()
		}()
		go func() {
			if config.TraceArea {
				TraceArea(9)
			}
		}()
		go func() {
			RefreshLivers()
		}()
		go func() {
			for _, slave := range config.Slaves {
				res, _ := client.R().Get(slave + "/ping")
				if res.String() != "pong" {

				}
			}
		}()
		c.AddFunc("@every 2m", func() { UpdateSpecial() })
		c.AddFunc("@every 120m", RefreshFollowings)
		c.AddFunc("@every 240m", UpdateCommon)
		c.AddFunc("@every 15m", func() {
			if config.TraceArea {
				TraceArea(9)
			}
		})
		c.AddFunc("@every 1m", FixMoney)
		c.AddFunc("@every 1m", func() { RefreshCollection(strconv.Itoa(GetCollectionId())) })
		c.AddFunc("@every 60m", RefreshLivers)
		if err != nil {
			return
		}

		c.Start()

		SortTracing()
		for i := range config.Tracing {
			man.AddTask(config.Tracing[i])
		}
	}
	if config.Mode == "Slaver" {
		log.Printf("Slave Mode")
	}

	select {}
}
