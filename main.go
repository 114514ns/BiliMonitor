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
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	url2 "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icoMn src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.
var client = resty.New()
var Cookie = ""
var Special = make([]User, 0)
var RecordedDynamic = make([]string, 0)
var RecordedMedias = make([]string, 0)
var GiftPrice = map[string]float32{}
var mailClient = resend.NewClient("")

const USER_AGENT = "User-Agent\", \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

type Config struct {
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
type UserResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Card struct {
			Name string `json:"name"`
			Face string `json:"face"`
		}
		Followers int `json:"follower"`
	} `json:"data"`
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
type VideoResponse struct {
	Data struct {
		Cover     string `json:"pic"`
		Title     string `json:"title"`
		Duration  int    `json:"duration"`
		PublishAt int64  `json:"pubdate"`
		Desc      string `json:"desc"`
		Owner     struct {
			Mid  int64  `json:"mid"`
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"owner"`
		Pages []struct {
			Cid      int    `json:"cid"`
			Title    string `json:"part"`
			Duration int    `json:"duration"`
		}
	} `json:"data"`
}
type PlayListResponse struct {
	Data struct {
		Archives []struct {
			BV       string `json:"bvid"`
			CreateAt int    `json:"pubdate"`
			Cover    string `json:"pic"`
			Duration int    `json:"duration"`
			Title    string `json:"title"`
		} `json:"archives"`
		Meta struct {
			Name string `json:"name"`
		} `json:"meta"`
	} `json:"data"`
}
type Status struct {
	Live         bool
	LastActive   int64
	UName        string
	UID          string
	Area         string
	Title        string
	StartAt      string
	RemainTrying int8
	Face         string
	Cover        string
	LiveRoom     string
	Danmuku      []FrontLiveAction
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
type CollectionList struct {
	Data struct {
		List []struct {
			Title string `json:"title"`
			ID    int    `json:"id"`
		}
	}
}
type CollectionMedias struct {
	Data struct {
		Medias []struct {
			Title string `json:"title"`
			BV    string `json:"bvid"`
		}
	}
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

func FetchUser(mid string) User {
	var url = "https://api.bilibili.com/x/web-interface/card?mid=" + mid
	res, _ := client.R().SetHeader("User-Agent", USER_AGENT).Get(url)
	var userResponse = UserResponse{}
	sonic.Unmarshal(res.Body(), &userResponse)

	var user = User{}
	user.Name = userResponse.Data.Card.Name
	user.Face = userResponse.Data.Card.Face
	user.Fans = userResponse.Data.Followers

	user.UserID = mid
	return user

}
func UpdateCommon() {
	for i := range Followings {
		if i > len(Followings)-1 {
			continue
		}
		var id = Followings[i].UserID
		var user = FetchUser(id)
		user.Face = ""
		db.Save(&user)
		time.Sleep(3 * time.Second)
	}
}

func UpdateGuard() {

}
func RefreshCookie() {
	//var url = "https://www.bilibili.com/correspond/1/" + getCorrespondPath(time.Now().UnixMilli())
	//_, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", USER_AGENT).Get(url)

}
func GetDefaultCookie() {
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
	alist, err := client.R().SetBody(req).Post(config.AlistServer + "api/auth/login/hash")

	if err != nil {
		log.Println(err)
	}
	var res = LoginResponse{}
	sonic.Unmarshal(alist.Body(), &res)
	return res.Data.Token
}
func UploadFile(path string, alistPath string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("打开文件失败: %w", err)
	}
	defer file.Close()

	bodyReader, bodyWriter := io.Pipe() // 创建 Pipe
	writer := multipart.NewWriter(bodyWriter)

	go func() {
		defer bodyWriter.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			bodyWriter.CloseWithError(fmt.Errorf("创建表单文件失败: %w", err))
			return
		}

		_, err = io.Copy(part, file)
		if err != nil {
			bodyWriter.CloseWithError(fmt.Errorf("复制文件数据失败: %w", err))
			return
		}

		writer.Close()
	}()

	req, err := http.NewRequest("PUT", config.AlistServer+"api/fs/form", bodyReader)
	if err != nil {
		log.Println("创建请求失败:", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", GetAlistToken())
	req.Header.Set("File-Path", alistPath)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("上传请求失败: %w", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应失败: %w", err)
	}

	fmt.Printf("[%s] %d %s\n", alistPath, resp.StatusCode, string(body))
	return nil
}

func UploadArchive(video Video) string {
	log.Printf("[%s] 开始下载", video.Title)
	os.Mkdir("cache", 066)
	var videolink = "https://bilibili.com/video/" + video.BV + "?p=" + strconv.Itoa(video.Part)
	vRes, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com").SetHeader("User-Agent", USER_AGENT).Get(videolink)
	htmlContent := vRes.Body()
	reader := bytes.NewReader(htmlContent)
	var bv = video.BV
	root, _ := html.Parse(reader)
	find := goquery.NewDocumentFromNode(root).Find("script")
	var final = ""
	find.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "m4s") && strings.Contains(s.Text(), "backup_url") {
			var json = strings.Replace(s.Text(), "window.__playinfo__=", "", 1)
			var v = Dash{}
			sonic.Unmarshal([]byte(json), &v)
			audio, _ := client.R().SetDoNotParseResponse(true).SetHeader("Referer", "https://www.bilibili.com").SetHeader("Cookie", config.Cookie).Get(v.Data.Dash0.Audio[0].Link)
			//defer audio.RawBody().Close()
			os.WriteFile("cache/"+video.BV+".mp3", audio.Body(), 066)
			audioFile, _ := os.Create("cache/" + video.BV + ".mp3")
			//defer audioFile.Close()
			io.Copy(audioFile, audio.RawBody())

			videoLink, _ := client.R().SetDoNotParseResponse(true).SetHeader("Referer", "https://www.bilibili.com").Get(v.Data.Dash0.Video[0].Link)
			//defer video.RawBody().Close()
			videoFile, _ := os.Create("cache/" + bv + ".m4s")
			//defer videoFile.Close()
			io.Copy(videoFile, videoLink.RawBody())
			cmd := exec.Command("ffmpeg", "-i", videoFile.Name(), "-i", audioFile.Name(), "-vcodec", "copy", "-acodec", "copy", "cache/"+bv+".mp4")
			//out, _ := cmd.CombinedOutput()
			cmd.Run() // 执行命令

			if video.ParentTitle != "" {
				video.ParentTitle = video.ParentTitle + "/"
			}
			//log.Println(string(out))
			final = config.AlistPath + "/Archive/" + video.Author + "/" + video.ParentTitle + "[" + strings.ReplaceAll(video.PublishAt, ":", "-") + "] " + video.Title + ".mp4"
			UploadFile("cache/"+bv+".mp4", final)

			os.Remove("cache/" + bv + ".mp4")
			os.Remove("cache/" + bv + ".mp3")
			os.Remove("cache/" + bv + ".m4s")
		}
	})
	return final

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
			PushDynamic("你关注的up主："+userName+"转发了动态 ", item.Modules.ModuleDynamic.Desc.Nodes[0].Text)
		}
	} else if Type == "DYNAMIC_TYPE_AV" { //发布视频
		archive.Type = "v"
		archive.BiliID = item.IDStr
		archive.Title = item.Modules.ModuleDynamic.Major.Archive.Title
		if push {
			PushDynamic("你关注的up主："+userName+"投稿了视频 ", archive.Title)
			go UploadArchive(ParseSingleVideo(item.Modules.ModuleDynamic.Major.Archive.Bvid)[0])
		}

	} else if Type == "DYNAMIC_TYPE_DRAW" { //图文
		archive.Type = "i"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Desc.Text
		if push {
			PushDynamic("你关注的up主："+userName+"发布了动态 ", item.Modules.ModuleDynamic.Major.Archive.Title)
		}

	} else if Type == "DYNAMIC_TYPE_WORD" { //文字
		archive.Type = "t"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
		if push {
			PushDynamic("你关注的up主："+userName+"转发了动态 ", item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
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
					if d.BiliID != "" {
						db.Save(&d)
					}

					json := make([]byte, 0)
					sonic.Unmarshal(json, item)
					//PushDynamic("动态json", string(json))

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

func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	secs := (duration % time.Minute) / time.Second

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
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
func GetCollectionId() int {
	var url = "https://api.bilibili.com/x/v3/fav/folder/created/list?ps=50&pn=1&up_mid=" + config.User
	res, _ := client.R().Get(url)
	var list = CollectionList{}
	sonic.Unmarshal(res.Body(), &list)
	for _, s := range list.Data.List {
		if s.Title == "Monitor" {
			return s.ID
		}
	}
	return 0
}
func RefreshCollection(id string) {

	var url = "https://api.bilibili.com/x/v3/fav/resource/list?ps=1&media_id=" + id
	res, _ := client.R().Get(url)
	var medias = CollectionMedias{}
	sonic.Unmarshal(res.Body(), &medias)
	if len(RecordedMedias) == 0 {
		for _, media := range medias.Data.Medias {
			RecordedMedias = append(RecordedMedias, media.BV)
		}
	} else {
		for i := range RecordedMedias {
			var item = RecordedMedias[i]
			var title = ""
			var found = false
			for _, media := range medias.Data.Medias {
				if item == media.BV {
					found = true
					title = media.Title
				}
			}
			if !found {
				RecordedMedias = append(RecordedMedias, item)
				go func() {
					var link = UploadArchive(ParseSingleVideo(item)[0])
					PushDynamic("你收藏的"+title+"已下载完成", link)
				}()
			}
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

var lives = map[string]*Status{} //[]string{}
var file = time.Now().Format(time.DateTime) + ".log"
var logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)

func main() {
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	content, err := os.ReadFile("config.json")
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	if err != nil {
		content = []byte("")

		config.SpecialDelay = "2m"
		config.CommonDelay = "30m"
		config.User = "451537183"
		config.RefreshFollowingsDelay = "30m"
		config.SpecialList = []int{}
		config.EnableQQBot = false
		config.EnableEmail = true
		config.FromMail = "bili@ikun.dev"
		config.ToMail = []string{"3212329718@qq.com"}
		config.ReportTo = []string{"3212329718"}
		config.BackServer = "http://127.0.0.1:3090"
		config.Tracing = []string{"544853"}
		config.EnableAlist = false
		config.EnableSQLite = true
		config.SQLitePath = "database.db"
		config.EnableMySQL = false
		config.EnableCollectionMonitor = false
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
		go TraceLive(config.Tracing[i])
		time.Sleep(30 * time.Second)

	}

	//var collectId = GetCollectionId()
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
