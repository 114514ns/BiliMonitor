package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/resend/resend-go/v3"
	"github.com/samber/lo"
)

type Room struct {
	UID   int64
	UName string
	Allow bool
	Live  bool
	Room  int
	Mails []string
}

type Config struct {
	Rooms     []Room
	EndPoint  string
	Port      int
	ResendKey string
	MailFrom  string
}

var m = make(map[int]Room)
var mMu sync.RWMutex

var restyClient = resty.New()

var config Config

func RefreshStatus(id []int64) {
	var s = "https://api.live.bilibili.com/xlive/fuxi-interface/UserService/getUserInfo?_ts_rpc_args_=[["
	for i, i2 := range id {
		s = s + strconv.FormatInt(i2, 10)
		if i != len(id)-1 {
			s = s + ","
		}
	}
	s = s + `],true,""]`

	res, err := restyClient.R().Get(s)
	if err != nil || res == nil {
		time.Sleep(3 * time.Second)
		return
	}

	bodyStr := res.String()
	head := bodyStr
	if len(head) > 70 {
		head = head[:70]
	}

	var m0 map[string]interface{}
	_ = json.Unmarshal(res.Body(), &m0)

	if !strings.Contains(head, "_ts_rpc_return_") || strings.Contains(head, "服务调用超时") {
		fmt.Println(bodyStr)
		time.Sleep(3 * time.Second)
		return
	}

	for _, i := range m0["_ts_rpc_return_"].(map[string]interface{})["data"].(map[string]interface{}) {
		var room = toInt(getString(i, "roomId"))
		if getString(i, "liveStatus") == "1" {
			mMu.Lock()
			v, ok := m[room]
			if ok && !v.Live {
				log.Printf("[%s] 开始直播\n", v.UName)
				v.Live = true
				m[room] = v
				go func() {
					restyClient.R().Get(config.EndPoint + "/schedule?room=" + toString(int64(v.Room)))
					for _, mail := range v.Mails {
						resendClient.Emails.Send(&resend.SendEmailRequest{
							From:    config.MailFrom,
							To:      []string{mail},
							Subject: fmt.Sprintf("你关注的主播 %s 开始直播", v.UName),
							Text:    "",
						})
					}
				}()
			}
			mMu.Unlock()
		}
		if getString(i, "liveStatus") == "0" || getString(i, "liveStatus") == "2" {
			mMu.Lock()
			v, ok := m[room]
			if ok && v.Live {
				v.Live = false
				m[room] = v
			}
			mMu.Unlock()
		}
	}
}

func SaveConfig() {
	var array []Room
	for i := range m {
		array = append(array, m[i])
	}
	config.Rooms = array
	bytes, _ := json.MarshalIndent(config, "", "  ")
	_ = os.WriteFile("config.json", bytes, 0644)
}

func LoadConfig() {
	bytes, _ := os.ReadFile("config.json")
	_ = json.Unmarshal(bytes, &config)
	for _, i := range config.Rooms {
		m[i.Room] = i
	}
}

func InitHttp() {
	var r = gin.Default()

	r.Use(CORSMiddleware())

	r.GET("/trace_srv/list", func(context *gin.Context) {
		mMu.Lock()
		var array []Room
		for i := range m {
			array = append(array, m[i])
		}
		mMu.Unlock()

		context.JSON(http.StatusOK, gin.H{
			"list": array,
		})
	})

	r.GET("/trace_srv/info", func(context *gin.Context) {
		var mid = toInt64(context.Query("mid"))
		if mid <= 0 {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}
		res, _ := restyClient.R().Get(fmt.Sprintf("https://api.live.bilibili.com/live_user/v1/Master/info?uid=%d", mid))
		var obj map[string]interface{}
		if res != nil {
			_ = json.Unmarshal(res.Body(), &obj)
		}
		context.JSON(http.StatusOK, gin.H{
			"UID":   mid,
			"UName": getString(obj, "data.info.uname"),
			"Fans":  getInt(obj, "data.follower_num"),
		})

	})

	r.POST("/trace_srv/submit", func(context *gin.Context) {
		var uid = toInt64(context.PostForm("uid"))
		if uid <= 0 {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "bad params",
			})
			return
		}

		mMu.Lock()
		defer mMu.Unlock()

		var found = false
		for i := range m {
			v := m[i]
			if v.UID == uid {
				found = true
				break
			}
		}
		if found {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "liver already exists",
			})
			return
		}

		res, _ := restyClient.R().Get(fmt.Sprintf("https://api.live.bilibili.com/live_user/v1/Master/info?uid=%d", uid))
		var obj map[string]interface{}
		if res != nil {
			_ = json.Unmarshal(res.Body(), &obj)
		}

		if getInt(obj, "data.room_id") == 0 {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "liver not found",
			})
			return
		}

		m[getInt(obj, "data.room_id")] = Room{
			UID:   uid,
			UName: getString(obj, "data.info.uname"),
			Allow: false,
			Live:  false,
			Room:  getInt(obj, "data.room_id"),
		}

		SaveConfig()

		context.JSON(http.StatusOK, gin.H{
			"msg": "success",
		})
	})

	r.POST("/trace_srv/del", func(context *gin.Context) {
		var room = toInt(context.PostForm("room"))
		mMu.Lock()
		defer mMu.Unlock()
		delete(m, room)
		SaveConfig()
		context.JSON(http.StatusOK, gin.H{
			"msg": "success",
		})
	})

	r.POST("/trace_srv/clear", func(context *gin.Context) {
		mMu.Lock()
		defer mMu.Unlock()
		var newM = make(map[int]Room)
		for i := range m {
			if m[i].Allow {
				newM[i] = m[i]
			}
		}
		m = newM
		SaveConfig()

		context.JSON(http.StatusOK, gin.H{
			"msg": "success",
		})
	})

	r.POST("/trace_srv/allow", func(context *gin.Context) {
		var room = toInt(context.PostForm("room"))

		var uid = toInt64(context.PostForm("uid"))

		mMu.Lock()
		defer mMu.Unlock()

		v, ok := m[room]
		if !ok {
			for k := range m {
				if m[k].UID == uid {
					v = m[k]
					room = v.Room
					ok = true
				}
			}
		}
		if ok {

			v.Allow = true
			m[room] = v
			SaveConfig()
			context.JSON(http.StatusOK, gin.H{
				"msg": "success",
			})
		} else {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "room not found",
			})
		}
	})
	r.POST("/trace_srv/mail/add", func(context *gin.Context) {
		var room = toInt(context.PostForm("room"))
		var mail = context.PostForm("mail")
		mMu.Lock()
		defer mMu.Unlock()
		v, ok := m[room]
		if ok && lo.IndexOf(v.Mails, mail) == -1 {
			v.Mails = append(v.Mails, mail)
			m[room] = v
			SaveConfig()
		}
		context.JSON(http.StatusOK, gin.H{
			"msg": "success",
		})
	})
	r.POST("/trace_srv/mail/del", func(context *gin.Context) {
		var room = toInt(context.PostForm("room"))
		var mail = context.PostForm("mail")
		mMu.Lock()
		defer mMu.Unlock()
		v, ok := m[room]
		var found = false
		if ok {
			for i, i2 := range v.Mails {
				if i2 == mail {
					v.Mails = append(v.Mails[:i], v.Mails[i+1:]...)
					m[room] = v
					found = true
				}
			}
		}
		if found {
			SaveConfig()
			context.JSON(http.StatusOK, gin.H{
				"msg": "success",
			})
		} else {
			context.JSON(http.StatusBadRequest, gin.H{
				"msg": "mail not found",
			})
		}
	})

	_ = r.Run(fmt.Sprintf(":%d", config.Port))
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		c.Writer.Header().Set("Cache-Control", " public, max-age=0, stale-while-revalidate=30")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

var resendClient *resend.Client

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	LoadConfig()
	resendClient = resend.NewClient(config.ResendKey)
	restyClient.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		req.Header.Set("Origin", "https://www.bilibili.com")
		return nil
	})
	go func() {
		for {
			mMu.RLock()
			snapshot := lo.MapToSlice(m, func(key int, value Room) Room { return value })
			mMu.RUnlock()

			for _, chunk := range lo.Chunk(snapshot, 30) {
				ids := lo.Map(
					lo.Filter(chunk, func(item Room, index int) bool { return item.Allow }),
					func(item Room, index int) int64 { return item.UID },
				)
				if len(ids) > 0 {
					RefreshStatus(ids)
				}
			}

			time.Sleep(30 * time.Second)
		}
	}()

	InitHttp()
}
