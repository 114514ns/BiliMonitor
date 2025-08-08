package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/html"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	url2 "net/url"
	"os"
	"os/exec"
	"runtime/debug"
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
var clickDb *gorm.DB

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"

type Config struct {
	UID                     int64
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
	HTTPProxy               string
	RefreshToken            string
	QueryProxy              string
	QueryAlive              int
	RequestDelay            int
}

type User struct {
	gorm.Model
	Name   string
	UserID int64
	Fans   int
	Face   string
	Bio    string
	Verify string
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
	Download  bool
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
	if config.HTTPProxy != "" {
		t := resty.New()
		t.SetProxy(config.HTTPProxy)
		r, _ := t.R().Get("https://bilibili.com")
		if r.StatusCode() != 200 {
			log.Fatal("HTTP代理配置错误")
		}
	}
}

func UpdateGuard() {

}
func ie() {
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
		var archive = Archive{}
		db.Find(&archive, "bili_id = ?", bv)
		if archive.BiliID == "" {
			archive.BiliID = bv
			archive.UID = video.UID
			archive.UName = video.Author
			archive.CreatedAt = time.Unix(resObj.Data.PublishAt, 0)
			archive.Title = video.Title
			archive.Text = video.Desc
			archive.Type = "v"
			archive.Download = false
			db.Save(&archive)
		}
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
			var a = Archive{}
			db.Find(&a, "bili_id = ?", video.BV)
			if a.BiliID == "" {
				a.BiliID = video.BV
				a.UID = video.UID
				a.UName = video.Author
				a.CreatedAt = time.Unix(int64(archive.CreateAt), 0)
				a.Title = video.Title
				a.Text = video.Desc
				a.Type = "v"
				a.Download = false
				db.Save(&archive)
			}
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
	var cpy Config
	copier.Copy(&cpy, &config)
	cpy.Slaves = []string{}
	for _, slave := range config.Slaves {
		if slave != "http://127.0.0.1:"+strconv.Itoa(int(cpy.Port)) {
			cpy.Slaves = append(cpy.Slaves, slave)
		}
	}
	content, _ := sonic.Marshal(&cpy)
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
var websocketBytes int64 = 0

const ENV = "DEV"

var totalRequests = 0
var launchTime = time.Now()
var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.3",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.6",
	"Mozilla/5.0 (Linux; Android 13; SM-S908U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone14,3; U; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) Version/10.0 Mobile/19A346 Safari/602.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:138.0) Gecko/20100101 Firefox/138.0",
	"Mozilla/5.0 (Windows 7 Enterprise; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6099.71 Safari/537.36",
	"Mozilla/5.0 (Windows Server 2012 R2 Standard; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5975.80 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.5672 Safari/537.36",
}

func CSRF() string {
	split := strings.Split(config.Cookie, ";")
	jct := ""
	for _, s := range split {
		if strings.Contains(s, "bili_jct=") {
			jct = strings.Replace(s, "bili_jct=", "", 1)
		}
	}
	return strings.TrimSpace(jct)
}

var man *SlaverManager
var consoleLogger = log.New(os.Stdout, "", log.LstdFlags) //用于在控制台输出弹幕信息，不会输出到日志文件里

var tempCount = 0 //临时的弹幕计数。每五分钟从数据库读取弹幕数量，在加上内存里临时的弹幕数量，显示给前端。大大减少数据库压力。
var tempMutex sync.Mutex
var msg1 int64 = 0
var msg5 int64 = 0
var msg60 int64 = 0
var queryClient = resty.New()

func loadConfig() {
	content, err := os.ReadFile("config.json")
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
		config.HTTPProxy = ""
		Cookie = config.Cookie
		content, _ = sonic.Marshal(&config)
		os.Create("config.json")
		os.WriteFile("config.json", content, 666)
	}
	err = sonic.Unmarshal(content, &config)
}
func setupHTTPClient() {
	client.OnAfterResponse(func(c *resty.Client, response *resty.Response) error {
		totalRequests++
		httpBytes += response.RawResponse.ContentLength
		return nil
	})

	client.OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
		request.Header.Set("User-Agent", UserAgents[rand.Uint32()%uint32(len(UserAgents))])
		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, response *resty.Response) error {
		if strings.Contains(response.Request.URL, "bilibili.com") && strings.Contains(response.Request.URL, "api") {
			var obj map[string]interface{}
			sonic.Unmarshal(response.Body(), &obj)
			_, ok := obj["code"]
			if !ok && !strings.Contains(response.String(), "ts_rpc") {
				log.Println(response.String())
			}
			if ok && obj["code"].(float64) != 0 {
				if strings.Contains(response.Request.URL, "getRoomPlayInfo") {
					if obj["message"].(string) == "参数错误" {
						return nil
					}
				}
				if strings.Contains(response.String(), "ts_rpc_return") {
					if strings.Contains(response.String(), "hdslb") {
						return nil //偷点懒
					}
				}
				log.Println(response.Request.URL)
				log.Println(response.String())
				debug.PrintStack()
			}
			if response.IsError() {
				log.Println(response.Request.URL)
				log.Println(response.Error())
				debug.PrintStack()
			}
		}

		return nil
	})

	queryClient.SetTransport(&http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second, // 保持连接的生存期
		}).DialContext,
	})
	queryClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if rand.Int()%100 == 1 {
			r.Header.Set("Connection", "close")
		}
		r.Header.Set("User-Agent", UserAgents[rand.Uint32()%uint32(len(UserAgents))])
		return nil
	})
	if config.QueryProxy != "" {
		queryClient.SetProxy(config.QueryProxy)
	}
	queryClient.OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
		request.Header.Set("User-Agent", randomUserAgent())
		return nil
	})
}

var localClient = resty.New()

const MAX_TASK = 40

func main() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Panic  restarting", r)
					time.Sleep(time.Second * 10)
				}
			}()

			main0()
		}()
	}
}

func main0() {
	loadConfig()
	setupHTTPClient()

	if config.UID == 0 {
		config.UID = SelfUID(config.Cookie)
		SaveConfig()
		os.Exit(0)
	}
	var loggerFlag = log.Ldate | log.Ltime | log.Llongfile
	consoleLogger.SetFlags(loggerFlag)

	rand.Seed(time.Now().UnixNano())

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.SetFlags(loggerFlag)

	mailClient = resend.NewClient(config.ResendToken)
	if config.EnableSQLite {
		db, _ = gorm.Open(sqlite.Open("database.db"), &gorm.Config{
			Logger: logger.New(
				log.New(multiWriter, "", loggerFlag),
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

		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN: dsl, // DSN data source name
		}), &gorm.Config{Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			logger.Config{},
		)})
	}
	if db == nil || err != nil {
		log.Println("Fail to connect to database")
		log.Fatal(err.Error())
	} else {
		log.Println("Success to connect to database")
	}
	wbi.WithRawCookies(config.Cookie)
	wbi.initWbi()
	db.Table("enter_action").AutoMigrate(&LiveAction{})
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
	if config.HTTPProxy != "" {
		client.SetProxy(config.HTTPProxy)
		log.Println(config.HTTPProxy)
	} else {
		client.SetTransport(&http.Transport{
			Proxy: nil,
		})
	}
	time.Sleep(1 * time.Second)
	res, _ := client.R().Get("https://test.ipw.cn/")
	log.Println("当前ip：" + res.String())
	res, _ = client.R().Get("https://api.bilibili.com/x/web-interface/zone")
	log.Println("当前ip：" + res.String())

	RefreshCookie()
	time.Sleep(5 * time.Second)
	c.AddFunc("@every 1m", func() {
		tempMutex.Lock()
		msg1 = 0
		tempMutex.Unlock()
	})
	c.AddFunc("@every 5m", func() {
		tempMutex.Lock()
		msg5 = 0
		tempMutex.Unlock()
	})
	c.AddFunc("@every 60m", func() {
		tempMutex.Lock()
		msg60 = 0
		tempMutex.Unlock()
	})
	if config.Mode == "Master" {
		config.Slaves = append(config.Slaves, "http://127.0.0.1:"+strconv.Itoa(int(config.Port)))
		man = NewSlaverManager(config.Slaves)
		man.OnErr = func(tasks []string) {
			log.Println("onError")
		}
		for _, slave := range man.Nodes {
			res, err := client.R().Get(slave.Address + "/monitor")
			if err == nil {
				var obj map[string]interface{}
				sonic.Unmarshal(res.Body(), &obj)
				/*
					for _, i2 := range obj["lives"].([]interface{}) {
						var room = i2.(map[string]interface{})["LiveRoom"].(string)
						slave.Tasks = append(slave.Tasks, room)
					}
					man.Nodes[i].Tasks = slave.Tasks

				*/
			}
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
		c.AddFunc("@every 720m", UpdateCommon)
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

		SortTracing()

		for i := range config.Tracing {
			man.AddTask(config.Tracing[i])
		}

	}
	c.AddFunc("@every 240m", RefreshCookie)
	c.Start()
	if config.Mode == "Slaver" {
		log.Printf("Slave Mode")
	}
	select {}
}
func randomBrowserVersion(browser string) string {
	majorVersion := 117 + rand.Intn(3)
	minorVersion := rand.Intn(1000)
	return fmt.Sprintf("%s/%d.%d", browser, majorVersion, minorVersion)
}
func randomUserAgent() string {

	var operatingSystems = []string{
		"Windows NT 10.0; Win64; x64",
		"Macintosh; Intel Mac OS X 12_6",
		"Linux; Android 13; Pixel 6",
		"Linux; U; Android 13; SM-G991U",
		"X11; Linux x86_64",
	}

	var devices = []string{
		"Mobile Safari/537.36",
		"Safari/537.36",
		"Mobile/15E148 Safari/604.1",
		"Safari/604.1",
	}
	os := operatingSystems[rand.Intn(len(operatingSystems))]
	device := devices[rand.Intn(len(devices))]

	// 定义浏览器类型
	browserTypes := []string{"Chrome", "Edge", "Firefox", "Safari"}
	browser := browserTypes[rand.Intn(len(browserTypes))]

	browserVersion := randomBrowserVersion(browser)

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) %s %s", os, browserVersion, device)
}

func RefreshCookie() {

	res, _ := client.R().SetHeader("Cookie", config.Cookie).Get("https://passport.bilibili.com/x/passport-login/web/cookie/info")
	type RefreshResponse struct {
		Data struct {
			Refresh   bool  `json:"refresh"`
			Timestamp int64 `json:"timestamp"`
		}
	}
	var obj RefreshResponse
	sonic.Unmarshal(res.Body(), &obj)
	if obj.Data.Refresh == true {
		log.Println("[CookieRefresh] Begin")
		path := getCorrespondPath(obj.Data.Timestamp)
		res, _ = client.R().SetHeader("Cookie", config.Cookie).Get("https://www.bilibili.com/correspond/1/" + path)
		htmlContent := res.Body()
		reader := bytes.NewReader(htmlContent)
		root, _ := html.Parse(reader)
		find := goquery.NewDocumentFromNode(root).Find("#1-name")
		csrf := find.Text()
		var body = fmt.Sprintf("csrf=%s&refresh_csrf=%s&source=main_web&refresh_token=%s", CSRF(), csrf, config.RefreshToken)
		res, _ := client.R().SetBody(body).SetHeader("Cookie", config.Cookie).Post("https://passport.bilibili.com/x/passport-login/web/cookie/refresh?" + body)
		log.Println("[CookieRefresh] " + res.Request.URL)
		type RefreshResult struct {
			Data struct {
				Refresh string `json:"refresh_token"`
			}
		}
		var obj RefreshResult
		sonic.Unmarshal(res.Body(), &obj)
		var newCookie = ""
		for _, s := range res.RawResponse.Header.Values("Set-Cookie") {
			newCookie = newCookie + strings.Split(s, ";")[0] + ";"
		}
		if newCookie == "" || obj.Data.Refresh == "" {
			log.Println("[CookieRefresh] Refresh Failed")
			log.Println(res.String())
			return
		}
		config.RefreshToken = obj.Data.Refresh
		config.Cookie = newCookie + "buvid3=" + uuid.New().String() + "infoc"
		log.Printf("[CookieRefresh] RefreshToken=%s", obj.Data.Refresh)
		log.Printf("[CookieRefresh] Cookie=%s", newCookie)
		SaveConfig()
		time.Now()
	} else {
		log.Println("[CookieRefresh] Skip")
	}

}
