package main

import (
	"bytes"
	"fmt"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/html"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	LoginMode               string
	PoolToken               string
	PoolEndPoint            string
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
	ConnectionPoolSize      int
	ClickServer             string
	BlackAreaLiver          []int64
	PlaybackRepositories    map[string]PlaybackRepository
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
	Live         bool
	LastActive   int64
	UName        string
	UID          string
	Area         string
	Title        string
	StartAt      string
	RemainTrying int
	Face         string
	Cover        string
	LiveRoom     string
	Fans         int
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
		PushServerHime(title, msg)
	}

}
func PushServerHime(title, msg string) {
	var url = fmt.Sprintf(config.ServerPushKey+"?title=%s&desp=%s", url2.QueryEscape(title), url2.QueryEscape(msg))
	client.R().Get(url)
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
	content, _ := sonic.MarshalIndent(config, "", "   ")
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

var cPools = make([]*resty.Client, 0)
var requestCountMutex sync.Mutex

func setupHTTPClient() {
	client.OnAfterResponse(func(c *resty.Client, response *resty.Response) error {
		requestCountMutex.Lock()
		totalRequests++
		requestCountMutex.Unlock()
		httpBytes += response.RawResponse.ContentLength
		return nil
	})

	client.OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
		request.Header.Set("User-Agent", UserAgents[rand.Uint32()%uint32(len(UserAgents))])
		if config.LoginMode == "pool" || config.LoginMode == "mix" {
			if strings.Contains(request.URL, config.PoolEndPoint) {
				request.Header.Set("Authorization", config.PoolToken)
			}
		}
		return nil
	})

	var checkFailHandler = func(c *resty.Client, response *resty.Response) error {
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
	}

	client.OnAfterResponse(checkFailHandler)

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

	var ips []string
	for j := 0; j < config.ConnectionPoolSize; j++ {

		var c = resty.New().SetProxy(config.QueryProxy)
		ip := checkIP(c)
		var found = false
		for s := range ips {
			if ip == ips[s] {
				found = true
				break
			}
		}
		if !found {
			ips = append(ips, ip)
			fmt.Println(len(cPools))
			c.OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
				request.Header.Set("User-Agent", randomUserAgent())
				requestCountMutex.Lock()
				totalRequests++
				requestCountMutex.Unlock()
				return nil
			})
			c.OnAfterResponse(checkFailHandler)
			cPools = append(cPools, c)
		} else {
			transport, _ := c.Transport()
			transport.CloseIdleConnections()
			j--
		}
	}
	go func() {
		for {
			for _, pool := range cPools {
				pool.R().Get("https://api.bilibili.com/x/web-interface/zone")
				time.Sleep(2 * time.Second)
			}
			time.Sleep(time.Second * 30)
		}
	}()

}

var localClient = resty.New()

const MAX_TASK = 1250

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
	clickDb, _ = gorm.Open(clickhouse.Open(config.ClickServer))
	if config.UID == 0 {
		config.UID = SelfUID(config.Cookie)
		SaveConfig()
		os.Exit(0)
	}
	SelfUID(config.Cookie)
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
		s, _ := db.DB()
		s.SetMaxOpenConns(200)
		s.SetMaxIdleConns(100)
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
	//db.AutoMigrate(&LiveAction{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Archive{})
	db.AutoMigrate(&AreaLiver{})
	db.AutoMigrate(&AreaLive{})
	db.AutoMigrate(&FansClub{})
	db.AutoMigrate(&FaceCache{})
	db.AutoMigrate(&OnlineStatus{})
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

	if config.LoginMode == "cookie" {

		RefreshCookie()
		time.Sleep(5 * time.Second)
	}
	if config.LoginMode == "mix" {
		config.Cookie = PickCookie()
		go func() {
			for {
				config.Cookie = PickCookie()
				time.Sleep(time.Second * 300)
			}
		}()
	}
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
		//RecoverLive()
		go func() {
			RefreshFollowings()
			UpdateCommon()
		}()
		go func() {
			if config.TraceArea {
				TraceArea(9, true)
			}
		}()
		go func() {
			time.Sleep(900 * time.Second)
			RefreshMessagePoints()
			RefreshLivers()
			RefreshWatcher()
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
		c.AddFunc("@every 5m", func() {
			if config.TraceArea {
				TraceArea(9, true)
			}
		})
		c.AddFunc("@every 1m", func() {
			if config.TraceArea {
				TraceArea(9, false)
			}
		})

		c.AddFunc("@every 60m", RefreshLivers)
		c.AddFunc("@every 60m", RefreshMessagePoints)
		c.AddFunc("@every 120m", RefreshWatcher)

		if err != nil {
			return
		}
		for i := range config.Tracing {
			man.AddTask(config.Tracing[i])
		}
	}
	if config.LoginMode == "cookie" {
		c.AddFunc("@every 240m", RefreshCookie)
	}
	c.Start()
	if config.Mode == "Slaver" {
		log.Printf("Slave Mode")
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)

			var batch1 []LiveAction

			actionMutex.Lock()
			if len(cacheAction) > 0 {
				batch1 = cacheAction
				cacheAction = make([]LiveAction, 0)
			}
			actionMutex.Unlock()
			if len(batch1) > 0 {
				clickDb.Table("enter_actions").Create(&batch1)
			}

			var batch2 []LiveAction

			extraAction.Lock()
			if len(extraList) > 0 {
				batch2 = extraList
				extraList = make([]LiveAction, 0)
			}
			extraAction.Unlock()

			if len(batch2) > 0 {
				db.Save(&batch2)
			}
		}
	}()
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

var PICK_CACHETIME int64 = 60
var LAST_PICK int64 = 0
var PICK_CACHE = ""
var pickMu sync.Mutex

func PickCookie() string {
	now := time.Now().Unix()
	pickMu.Lock()
	defer pickMu.Unlock()
	if PICK_CACHE != "" && LAST_PICK != 0 && now-LAST_PICK < PICK_CACHETIME {
		return PICK_CACHE
	}
	res, err := client.R().Get(config.PoolEndPoint + "pick")
	if err != nil {
		return PICK_CACHE
	}
	var obj map[string]interface{}
	if err := sonic.Unmarshal(res.Body(), &obj); err != nil {
		return PICK_CACHE
	}
	c := getString(obj, "Cookie")
	if c == "" {
		return PICK_CACHE
	}
	PICK_CACHE = c
	LAST_PICK = now
	return PICK_CACHE
}
