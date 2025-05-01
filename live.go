package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	_ "runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SelfInfo struct {
	Data struct {
		Mid int `json:"mid"`
	} `json:"data"`
}

func SelfUID(cookie string) int {
	res, _ := client.R().SetHeader("Cookie", cookie).Get("https://api.bilibili.com/x/web-interface/nav")

	var self = SelfInfo{}
	sonic.Unmarshal(res.Body(), &self)
	return self.Data.Mid
}
func FixMoney() {
	var lives0 []Live
	db.Find(&lives0)

	for _, v := range lives0 {
		if v.EndAt != 0 {
			continue //已经结束的直播不需要刷新
		}
		if time.Now().Unix()-v.StartAt > 3600*24*5 {
			continue //如果连续播了5天以上，大概率是直播结束的时候没有检测到，实际已经结束
		}
		var sum float64
		db.Table("live_actions").Select("SUM(gift_price)").Where("live = ? ", v.ID).Scan(&sum)
		result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sum), 64)
		var msgCount int64
		db.Model(&LiveAction{}).Where("live = ? and action_name = 'msg'", v.ID).Count(&msgCount)
		v.Money = result
		v.Message = int(msgCount)
		var last = LiveAction{}
		db.Where("live = ?", v.ID).Last(&last)
		if (time.Now().Unix() + 8*3600 - last.CreatedAt.Unix()) < 0 {

		}
		db.Save(&v)
	}
}
func ServerLiveStatus(roomId string) bool {
	url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + roomId
	res, _ := client.R().Get(url)
	status := LiveStatusResponse{}
	sonic.Unmarshal(res.Body(), &status)
	return status.Data.LiveStatus == 1
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
func GetServerState(room string) bool {
	url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + room
	res, _ := client.R().Get(url)
	status := LiveStatusResponse{}
	sonic.Unmarshal(res.Body(), &status)
	return status.Data.LiveStatus == 1
}
func RemoveEmpty() {
	db.Where("money = 0 and message = 0").Delete(&Live{})
}
func RecoverLive() {
	var array []Live
	db.Model(&array).Limit(20).Order("id desc").Find(&array)
	for _, live := range array {
		var roomId = strconv.Itoa(live.RoomId)
		if live.EndAt == 0 {
			if GetServerState(strconv.Itoa(live.RoomId)) {
				var _, ok = lives[roomId]
				if len(lives) < 25 && !ok && !Has(config.Tracing, roomId) {

					lives[roomId] = &Status{RemainTrying: 40}
					lives[roomId].Danmuku = make([]FrontLiveAction, 0)
					lives[roomId].OnlineWatcher = make([]Watcher, 0)
					lives[roomId].GuardList = make([]Watcher, 0)
					go func() {
						TraceLive(strconv.Itoa(live.RoomId))
					}()
					log.Printf("[%s] 恢复直播", live.UserName)
					time.Sleep(30 * time.Second)

				}
			}
		}
	}
}
func GetLiveStream(room string) string {

	now := time.Now()
	uri, _ := url.Parse("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?qn=10000&protocol=0,1&format=0,1,2&codec=0,1,2&web_location=444.8&room_id=" + room)
	signed, _ := wbi.SignQuery(uri.Query(), now)
	res, _ := client.R().SetHeader("User-Agent", USER_AGENT).SetHeader("Cookie", config.Cookie).Get("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?" + signed.Encode())
	var s = LiveStreamResponse{}
	sonic.Unmarshal(res.Body(), &s)
	stream := s.Data.PlayurlInfo.Playurl.Stream
	if stream != nil {
		//Format[0]是ts格式，可以直接拿来拼接，Format[1]是fmp4，需要先把ext-x-map拼到每一个分片前面，好像还有点问题
		obj := stream[len(stream)-1].Format[0].Codec[ /*len(stream[len(stream)-1].Format[0].Codec)-1*/ 0]
		if obj.UrlInfo[0].Host+obj.BaseUrl+obj.UrlInfo[0].Extra == "" {
			time.Now()
		}
		return obj.UrlInfo[0].Host + obj.BaseUrl + obj.UrlInfo[0].Extra
	} else {
		time.Now().Unix()
	}

	return ""

}
func GetOnline(room string, liver string) []Watcher {
	var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/queryContributionRank?ruid=%s&room_id=%s", liver, room)
	res, _ := client.R().Get(url)
	var o = OnlineWatcherResponse{}
	sonic.Unmarshal(res.Body(), &o)
	lives[room].OnlineCount = o.Data.Count
	var arr = make([]Watcher, 0)
	for _, s := range o.Data.Item {
		var watcher = Watcher{}
		watcher.Name = s.Name
		watcher.Face = s.Face
		watcher.Days = s.Days
		watcher.UID = s.UID
		watcher.Guard = s.Guard
		watcher.Medal.Color = s.UInfo.Medal.Color
		watcher.Medal.Name = s.UInfo.Medal.Name
		watcher.Medal.Level = s.UInfo.Medal.Level
		arr = append(arr, watcher)
	}
	return arr

}
func GetGuardCount(room string, liver string) string {
	var page = 1
	var l1 = 0
	var l2 = 0
	var l3 = 0
	var total = 0
	var entry = false
	for true {
		var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/app-room/v2/guardTab/topListNew?roomid=%s&page=%s&ruid=%s&page_size=30", room, strconv.Itoa(page), liver)
		res, _ := client.R().Get(url)
		var r = GuardListResponse{}
		sonic.Unmarshal(res.Body(), &r)
		total = r.Data.Info.Total
		if page == 1 {
			for _, item := range r.Data.Top {
				level := item.Info.Medal.GuardLevel
				if level == 1 {
					l1++
				}
				if level == 2 {
					l2++
				}
				if level == 3 {
					l3++
					entry = true
				}
			}
		}
		for _, item := range r.Data.List {
			level := item.Info.Medal.GuardLevel
			if level == 1 {
				l1++
			}
			if level == 2 {
				l2++
			}
			if level == 3 {
				l3++
				entry = true
			}
		}
		if entry {
			l3 = total - l1 - l2
			return fmt.Sprintf("%d,%d,%d", l1, l2, l3)
		}
		page++
	}
	return ""
}
func GetFreeClubCounts(liver string, delay int) int {
	left := 1
	right := 2000
	found := 0
	items := 0
	for left <= right {
		mid := (left + right) / 2
		u := fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/getFansMembersRank?ruid=%s&page=%s&page_size=30&rank_type=2&ts=%s", liver, strconv.Itoa(mid), strconv.FormatInt(time.Now().Unix(), 10))
		res, _ := client.R().Get(u)
		obj := FansClubResponse{}
		sonic.Unmarshal(res.Body(), &obj)
		l := len(obj.Data.Item)
		if obj.Message == "服务调用超时" {
			continue
		}
		if l == 0 {
			right = mid - 1
		} else {
			found = mid
			items = l
			left = mid + 1
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return 30*(found-1) + items
}
func GetFansClub(liver string, delay int, callback func(g DBGuard), logChain []Log) []DBGuard {
	var page = 1
	var list = make([]DBGuard, 0)
	t := "0"
	//type=0是活跃的粉丝团用户
	//type=2是所有没上过舰长的粉丝团用户
	for {
		u := fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/getFansMembersRank?ruid=%s&page=%s&page_size=30&rank_type=%s&ts=%s", liver, strconv.Itoa(page), t, strconv.FormatInt(time.Now().Unix(), 10))
		res, _ := client.R().Get(u)
		obj := FansClubResponse{}
		sonic.Unmarshal(res.Body(), &obj)
		if len(obj.Data.Item) == 0 && !strings.Contains(obj.Message, "服务调用超时") {
			break
		}
		if !strings.Contains(obj.Message, "服务调用超时") {
			page++
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
		//addLog(logChain, fmt.Sprintf("getting   fans msg=%s,page=%d,len=%d  url = %s", obj.Message, page, len(obj.Data.Item), u))
		for _, s := range obj.Data.Item {
			var d = DBGuard{}
			d.Score = s.Score
			d.Level = s.Level
			d.Type = s.Medal.Type
			d.UID = s.UID
			d.UName = s.UName
			d.MedalName = s.Medal.Name
			list = append(list, d)
			callback(d)
		}

	}
	addLog(logChain, fmt.Sprintf("got fansClub,len=%d", len(list)))
	return list
}
func GetGuard(room string, liver string, ignoreCache bool) []Watcher {
	if !ignoreCache && time.Now().Unix()-lives[room].GuardCacheKey < 60*10 {
		return lives[room].GuardList
	}
	if !ignoreCache {
		lives[room].GuardCacheKey = time.Now().Unix()
	}
	var arr = make([]Watcher, 0)
	var page = 1
	for true {
		var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/app-room/v2/guardTab/topListNew?roomid=%s&page=%s&ruid=%s&page_size=30", room, strconv.Itoa(page), liver)
		res, _ := client.R().Get(url)
		var r = GuardListResponse{}
		sonic.Unmarshal(res.Body(), &r)
		if page == 1 {
			for _, s := range r.Data.Top {
				var watcher = Watcher{}
				watcher.Name = s.Info.User.Name
				watcher.Face = s.Info.User.Face
				watcher.Days = s.Days
				watcher.UID = s.Info.UID
				watcher.Medal.Name = s.Info.Medal.Name
				watcher.Medal.Level = s.Info.Medal.Level
				watcher.Medal.Color = s.Info.Medal.Color
				watcher.Medal.GuardLevel = s.Info.Medal.GuardLevel
				watcher.Guard = s.Info.Medal.GuardLevel
				arr = append(arr, watcher)
			}
		}
		for _, s := range r.Data.List {
			var watcher = Watcher{}
			watcher.Name = s.Info.User.Name
			watcher.Face = s.Info.User.Face
			watcher.Days = s.Days
			watcher.UID = s.Info.UID
			watcher.Medal.Name = s.Info.Medal.Name
			watcher.Medal.Level = s.Info.Medal.Level
			watcher.Medal.Color = s.Info.Medal.Color
			watcher.Medal.GuardLevel = s.Info.Medal.GuardLevel
			watcher.Guard = s.Info.Medal.GuardLevel
			arr = append(arr, watcher)
		}
		page++
		if len(r.Data.List) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	if !ignoreCache {
		lives[room].GuardCount = len(arr)
	}
	return arr
}

// 某一时刻的主播信息
type AreaLiver struct {
	ID        uint `gorm:"primarykey"`
	UpdatedAt time.Time
	UName     string
	UID       int64
	Room      int
	Area      string
	Fans      int
	GuardList string
	Guard     string
	FreeClubs int
}

// 一场直播
type AreaLive struct {
	ID    uint `gorm:"primarykey"`
	Time  time.Time
	UName string
	UID   int64
	Room  int
	Title string
	Area  string
	Watch int
}

// 舰长，以json数组序列化后存在AreaLive的GuardList字段里
type DBGuard struct {
	UName     string
	UID       int64
	Type      int8
	Level     int8
	Score     int
	Liver     string
	LiverID   int64
	MedalName string
}

// 粉丝团，结构同上方的DBGuard，存放在一张单独的表
type FansClub struct {
	ID        uint `gorm:"primarykey"`
	UpdatedAt time.Time
	DBGuard
}
type Log struct {
	Time    time.Time
	Content string
}

var working = false

func addLog(chain []Log, content string) {
	/*
		chain = append(chain, Log{Time: time.Now(), Content: content})
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}

		// 生成完整日志内容
		timestamp := time.Now().Format("2006/01/02 15:04:05") // 和标准 log 一致
		fullMsg := fmt.Sprintf("%s %s:%d: %s", timestamp, file, line, content)
		fmt.Println(fullMsg)

	*/
}
func SendMessage(msg string, room int, onResponse func(string)) {

	//u := "https://api.live.bilibili.com/msg/send?"
	//url3, _ := url.Parse(u)
	//signed, _ := client.WBI.SignQuery(url3.Query(), time.Now())
	st := `{"appId":100,"platform":5}`
	body := fmt.Sprintf("bubble=0&msg=%s&color=16777215&mode=1&room_type=0&jumpfrom=71001&reply_mid=0&reply_attr=0&replay_dmid=&statistics=%s&fontsize=25&rnd=%d&roomid=%d&csrf=%s&csrf_token=%s", msg, st, time.Now().Unix()/1000, room, CSRF(), CSRF())
	res, _ := client.R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody(body).Post("https://api.live.bilibili.com/msg/send?" /*+ signed.Encode()*/)
	fmt.Println(res.String())
	if onResponse != nil {
		onResponse(res.String())
	}
}

func TraceArea(parent int) {
	if working {
		log.Println("TraceArea is still executing,break")
		return //确保不会重叠执行
	}
	var page = 1
	var arr = make([]AreaLiver, 0)
	for {
		working = true
		u, _ := url.Parse(fmt.Sprintf("https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=%d&area_id=0&sort_type=&page=%d&vajra_business_key=&web_location=444.43", parent, page))
		var now = time.Now()
		s, _ := wbi.SignQuery(u.Query(), now)
		res, _ := client.R().SetHeader("User-Agent", USER_AGENT).SetHeader("Cookie", config.Cookie).Get("https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?" + s.Encode())
		obj := AreaLiverListResponse{}
		sonic.Unmarshal(res.Body(), &obj)
		for _, s2 := range obj.Data.List {
			logChain := make([]Log, 0)
			var start = time.Now().Unix()
			var fansMap = make(map[int64]DBGuard)
			i := AreaLiver{}
			i.UName = s2.UName
			i.UID = s2.UID
			i.Room = s2.Room
			i.Area = s2.Area
			GetFace(strconv.FormatInt(s2.UID, 10))
			arr = append(arr, i)
			var found = AreaLiver{}
			var live = AreaLive{}
			var u0 = "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + strconv.Itoa(s2.Room)
			r, _ := client.R().Get(u0)
			var info = LiveStreamResponse{}
			sonic.Unmarshal(r.Body(), &info)
			addLog(logChain, fmt.Sprintf("Chain start"))
			addLog(logChain, fmt.Sprintf("Liver:%s  Title=%s", s2.UName, s2.Title))
			db.Model(&AreaLive{}).Where("uid = ?", s2.UID).Last(&live)
			if live.Time.Unix() != time.Unix(info.Data.Time, 0).Unix() {
				addLog(logChain, fmt.Sprintf("New Live"))
				var l = AreaLive{}
				l.UName = s2.UName
				l.UID = s2.UID
				l.Room = s2.Room
				l.Title = s2.Title
				l.Area = s2.Area
				l.Watch = s2.Watch.Num
				l.Time = time.Unix(info.Data.Time, 0)
				db.Save(&l)
			} else {
				addLog(logChain, fmt.Sprintf("Exist live,update watchNum %d->%d", live.Watch, s2.Watch.Num))
				live.Watch = s2.Watch.Num
				db.Save(&live)
			}

			db.Model(&AreaLiver{}).
				Where("uid = ?", s2.UID).
				Order("id DESC").
				First(&found)
			//如果这个主播在数据库里没有，或者上次更新超过两天，就更新一下
			if found.UID == 0 || time.Now().Unix()-found.UpdatedAt.Unix() > 3600*48 {
				if found.UID == 0 {
					addLog(logChain, fmt.Sprintf("New Liver"))
					addLog(logChain, fmt.Sprintf("getting fansClub"))
					//如果这个主播在数据库里没有，就获取活跃的粉丝团用户列表，和免费的粉丝团用户是数量
					GetFansClub(strconv.FormatInt(s2.UID, 10), 500, func(g DBGuard) {
						club := FansClub{}
						fansMap[g.UID] = g
						club.DBGuard = g
						club.Liver = s2.UName
						club.LiverID = s2.UID
						db.Save(&club)
					}, logChain)

					i.FreeClubs = GetFreeClubCounts(strconv.FormatInt(s2.UID, 10), 500)
				} else {
					//数据库里已经有这个主播的情况下只要更新活跃的用户就可以了
					GetFansClub(strconv.FormatInt(s2.UID, 10), 5, func(g DBGuard) {
						club := FansClub{}
						fansMap[g.UID] = g
						db.Model(&FansClub{}).Where("uid = ? and liver_id = ?", g.UID, s2.UID).Last(&club)
						if club.Score != g.Score {
							addLog(logChain, fmt.Sprintf("Update Medal %d->%d ", club.Score, g.Score))
						}
						if club.Score != 0 {
							club.Score = g.Score
							club.Liver = g.Liver
							db.Save(&club)
						} else {
							fansMap[g.UID] = g
							club.DBGuard = g
							club.Liver = s2.UName
							club.LiverID = s2.UID
							db.Save(&club)
						}

					}, logChain)
				}

				var user = FetchUser(strconv.FormatInt(s2.UID, 10), func() {
					time.Sleep(300 * time.Second)
					return
				})
				user.Face = ""
				var guards = make([]DBGuard, 0)
				var l1 = 0
				var l2 = 0
				var l3 = 0
				addLog(logChain, "getting guardList")
				for _, watcher := range GetGuard(strconv.Itoa(i.Room), strconv.FormatInt(i.UID, 10), true) {
					ins := DBGuard{}
					ins.Type = watcher.Guard
					ins.UID = watcher.UID
					ins.UName = watcher.Name
					ins.Level = watcher.Medal.Level
					ins.Liver = s2.UName
					ins.LiverID = s2.UID
					ins.MedalName = watcher.Medal.Name
					ins.Score = fansMap[watcher.UID].Score

					if ins.Type == 1 {
						l1++
					}
					if ins.Type == 2 {
						l2++
					}
					if ins.Type == 3 {
						l3++
					}

					guards = append(guards, ins)
				}
				i.Guard = fmt.Sprintf("%d,%d,%d", l1, l2, l3)
				addLog(logChain, fmt.Sprintf("got Guard info %s", i.Guard))
				b, _ := sonic.Marshal(guards)
				i.GuardList = string(b)
				db.Save(&user)
				i.Fans = user.Fans
				time.Sleep(time.Second * 2)
				db.Save(&i)
				addLog(logChain, fmt.Sprintf("Chain END, took %d s", time.Now().Unix()-start))
			} else {
				addLog(logChain, "Update recently,skip")
			}
			//var logStr = "" + i.UName
			o := AreaLiver{}
			db.Model(&AreaLiver{}).Where("uid = ?", s2.UID).Last(&o)
			if o.Fans > 10000 {
				//logStr = logStr + "fans > 10000"
				_, ok := lives[strconv.Itoa(i.Room)]
				if !ok {
					//logStr = logStr + "haven't recorded"
					var has = false
					recording := 0
					for _, s := range lives {
						if s.Live {
							recording++
						}
						if s.LiveRoom == strconv.Itoa(i.Room) {
							has = true
						}
					}
					if recording < 20 && !has {
						//logStr = logStr + "current recording:" + strconv.Itoa(recording)

						var roomId = strconv.Itoa(i.Room)
						lives[roomId] = &Status{RemainTrying: 40}
						lives[roomId].Danmuku = make([]FrontLiveAction, 0)
						lives[roomId].OnlineWatcher = make([]Watcher, 0)
						lives[roomId].GuardList = make([]Watcher, 0)
						lives[roomId].Live = true
						worker.AddTask(func() {
							time.Sleep(30 * time.Second)
							go TraceLive(roomId)
						})
					}

				}
			} else {
				//logStr = logStr + "fans < 10000"
			}
			//log.Printf(logStr)

		}
		if obj.Data.More == 0 {
			break
		}
		page++
		time.Sleep(time.Second * 2)
	}
	working = false
}
func BuildAuthMessage(room string) string {
	res, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + room)
	var liveInfo = LiveInfo{}
	sonic.Unmarshal(res.Body(), &liveInfo)
	if len(liveInfo.Data.HostList) == 0 {
		log.Println(res.String())
	}
	var cer = Certificate{}
	cer.Uid = SelfUID(config.Cookie)
	id, _ := strconv.Atoi(room)
	cer.RoomId = id
	cer.Type = 2
	cer.Key = liveInfo.Data.Token
	cer.Cookie = strings.Replace(config.Cookie, "buvid3=", "", 1)
	cer.Protover = 3
	json, _ := sonic.Marshal(&cer)
	return string(json)

}

var mu0 sync.RWMutex

func isLive(roomId string) bool {

	_, ok := lives[roomId]
	if !ok {
		return false
	}

	lives[roomId].RLock()
	defer lives[roomId].RUnlock()
	if s, ok := lives[roomId]; ok {
		return s.Live
	}
	return false
}

func setLive(roomId string, live bool) {
	log.Printf("[%s] setLive %b", lives[roomId].UName, live)
	lives[roomId].Lock()
	defer lives[roomId].Unlock()
	if s, ok := lives[roomId]; ok {
		s.Live = live
	}
}
func TraceLive(roomId string) {
	var roomUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + roomId
	var rRes, _ = client.R().Get(roomUrl)
	var liver string
	var roomInfo = RoomInfo{}
	sonic.Unmarshal(rRes.Body(), &roomInfo)
	FillGiftPrice(roomId, roomInfo.Data.AreaId, roomInfo.Data.ParentAreaId)
	var dbLiveId = 0
	var liverId = strconv.FormatInt(roomInfo.Data.UID, 10)
	var startAt = roomInfo.Data.LiveTime

	var living = false
	var liverInfoUrl = "https://api.live.bilibili.com/live_user/v1/Master/info?uid=" + liverId
	liverRes, _ := client.R().Get(liverInfoUrl)
	var liverObj = LiverInfo{}
	sonic.Unmarshal(liverRes.Body(), &liverObj)
	liver = liverObj.Data.Info.Uname
	lives[roomId].UName = liver

	var faceUrl = "https://api.bilibili.com/x/web-interface/card?mid=" + liverId

	var faceRes, _ = client.R().SetHeader("User-Agent", USER_AGENT).Get(faceUrl)

	time.Sleep(1 * time.Second)

	type FaceInfo struct {
		Data struct {
			Card struct {
				Face string `json:"face"`
			} `json:"card"`
		} `json:"data"`
	}
	var faceInfo = FaceInfo{}
	sonic.Unmarshal(faceRes.Body(), &faceInfo)

	lives[roomId].Face = faceInfo.Data.Card.Face
	lives[roomId].UID = liverId
	lives[roomId].Area = roomInfo.Data.Area
	lives[roomId].Cover = roomInfo.Data.Face
	lives[roomId].Title = roomInfo.Data.Title
	lives[roomId].LiveRoom = roomId
	if !strings.Contains(startAt, "0000-00-00 00:00:00") {
		lives[roomId].Live = true

		//当前是开播状态
		var serverStartAt, _ = time.Parse(time.DateTime, startAt)
		lives[roomId].StartAt = startAt
		living = true

		var foundLive = Live{}

		db.Where("user_id=?", roomInfo.Data.UID).Last(&foundLive)

		var diff = abs(int(foundLive.StartAt - serverStartAt.Unix()))

		log.Println("diff  " + strconv.Itoa(diff))

		if diff < 90 {
			log.Println("续")

			lives[roomId].Live = true
			dbLiveId = int(foundLive.ID)
		} else {

			var new = Live{}
			new.UserID = liverId
			new.StartAt = serverStartAt.Unix()
			//new.ID,_ : = strconv.Atoi(roomId)
			new.Title = roomInfo.Data.Title
			new.Area = roomInfo.Data.Area
			new.Cover = roomInfo.Data.Face
			//new.UserName = roomInfo.Data

			var i, _ = strconv.Atoi(roomId)
			new.RoomId = i
			new.UserName = liver
			liver = strings.TrimSpace(liver) // 去除前后的空白字符

			db.Create(&new)
			dbLiveId = int(new.ID)
		}
	}

	var url0 = "https://api.bilibili.com/x/web-interface/nav"
	url0 = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + roomId
	res, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).Get(url0)
	var liveInfo = LiveInfo{}
	sonic.Unmarshal(res.Body(), &liveInfo)
	if len(liveInfo.Data.HostList) == 0 {
		log.Println(res.String())
	}
	u := url.URL{Scheme: "wss", Host: liveInfo.Data.HostList[0].Host + ":2245", Path: "/sub"}
	var dialer = &websocket.Dialer{
		Proxy:            nil,
		HandshakeTimeout: 45 * time.Second,
	}
	var c *websocket.Conn
	if lives[roomId].Live {
		c, _, err = dialer.Dial(u.String(), nil)
	}
	if err != nil {
		log.Fatal("["+liver+"]  "+"Dial:", err)
	}
	ticker := time.NewTicker(45 * time.Second)

	go func() {

		if lives[roomId].Live {
			log.Printf("[%s] 成功连接到弹幕服务器", liver)
			err := c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
			if err != nil {
				return
			}
		}
		for {

			// 打印接收到的消息
			var msg = ""
			if isLive(roomId) {
				_, message, err := c.ReadMessage()
				if err != nil && isLive(roomId) {
					log.Printf("[%s] 断开连接，尝试重连次数："+strconv.FormatInt(int64(lives[roomId].RemainTrying), 10), liver)
					for {
						if lives[roomId].RemainTrying > 0 {
							time.Sleep(time.Duration(rand.Int()%10000) * time.Millisecond)
							c, _, err := dialer.Dial(u.String(), nil)
							lives[roomId].RemainTrying--
							err = c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
							if err == nil {
								log.Printf("[%s] 重连成功", liver)
								break
							}
						}
					}
					return
				}
				websocketBytes += len(message)
				reader := io.NewSectionReader(bytes.NewReader(message), 16, int64(len(message)-16))
				brotliReader := brotli.NewReader(reader)
				var decompressedData bytes.Buffer

				// 通过 io.Copy 将解压后的数据写入到缓冲区

				_, err0 := io.Copy(&decompressedData, brotliReader)
				if err0 != nil {
					msg = string(message)
				} else {
					msg = string(decompressedData.Bytes())
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			buffer := bytes.NewReader([]byte(msg))

			for {
				if buffer.Len() < 16 {
					break
				}

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
				if buffer.Len() < int(totalSize-16) {
					break
				}
				msgData := make([]byte, totalSize-16)
				buffer.Read(msgData)

				var obj = string(msgData)
				var action = LiveAction{}
				action.Live = uint(dbLiveId)
				action.LiveRoom = roomId
				action.GiftPrice = sql.NullFloat64{Float64: 0, Valid: false}
				action.GiftAmount = sql.NullInt16{Int16: 0, Valid: false}
				var text = LiveText{}
				sonic.Unmarshal(msgData, &text)
				var front = FrontLiveAction{}
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") { // 弹幕
					action.ActionName = "msg"
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = strconv.Itoa(int(text.Info[2].([]interface{})[0].(float64)))
					action.Extra = text.Info[1].(string)
					action.HonorLevel = int8(text.Info[16].([]interface{})[0].(float64))
					value, ok := text.Info[0].([]interface{})[15].(map[string]interface{})
					if ok {
						user, exists := value["user"].(map[string]interface{})
						if exists {
							base, exists := user["base"].(map[string]interface{})
							if exists {
								face, exists := base["face"]
								if exists {
									front.Face = face.(string)
								}
							}
							medal, exists := user["medal"].(map[string]interface{})
							if exists {
								name, exists := medal["name"]
								if exists {
									action.MedalName = name.(string)
								}
								level, exists := medal["level"]
								if exists {
									action.MedalLevel = int8(level.(float64))
								}
								guardLevel, exists := medal["guard_level"]
								if exists {
									action.GuardLevel = int8(guardLevel.(float64))
								}
								color, exists := medal["v2_medal_color_start"]
								if exists {
									front.MedalColor = color.(string)
								}

							}
						}
					}
					db.Create(&action)
					log.Println("[" + liver + "]  " + text.Info[2].([]interface{})[1].(string) + "  " + text.Info[1].(string))

				} else if strings.Contains(obj, "SEND_GIFT") { //送礼物
					var info = GiftInfo{}
					sonic.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.MedalLevel = int8(info.Data.Medal.Level)
					action.MedalName = info.Data.Medal.Name
					action.FromId = strconv.Itoa(info.Data.SenderUinfo.UID)
					front.MedalColor = fmt.Sprintf("#%06X", info.Data.Medal.Color)
					mu.RLock()
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					mu.RUnlock()
					if price == 0 {
						price = float64(info.Data.Price/1000) * float64(info.Data.Num)
					}
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)
					action.GiftPrice = sql.NullFloat64{Float64: result, Valid: true}
					action.GiftAmount = sql.NullInt16{Int16: int16(info.Data.Num), Valid: true}
					if info.Data.Parent.GiftName != "" {
						action.Extra = info.Data.Parent.GiftName + "," + strconv.Itoa(info.Data.Parent.Price/1000)
					}
					front.Face = info.Data.Face
					front.GiftPicture = GiftPic[info.Data.GiftName]
					db.Create(&action)
					log.Printf("[%s] %s 赠送了 %d 个 %s，%.2f元", liver, info.Data.Uname, info.Data.Num, info.Data.GiftName, price)
				} else if strings.Contains(obj, "INTERACT_WORD") { //进入直播间

					var enter = EnterLive{}
					sonic.Unmarshal(msgData, &enter)
					action.FromId = strconv.Itoa(enter.Data.UID)
					action.FromName = enter.Data.Uname
					action.ActionName = "enter"
					//db.Create(&action)

				} else if strings.Contains(obj, "PREPARING") {
					/*
						lives[roomId].Live = false
						var sum float64
						db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", dbLiveId).Scan(&sum)

						db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{EndAt: time.Now().Unix(), Money: sum})
						living = false
						i, _ := strconv.Atoi(roomId)
						if config.EnableLiveBackup {
							go UploadLive(Live{RoomId: i, UserName: liver})
						}

					*/

				} else if text.Cmd == "LIVE" {
					/*
						var new = Live{}
						living = true
						new.UserID = liverId
						time.Sleep(time.Second * 5) //如果马上去请求直播间信息会有问题
						var r, _ = client.R().Get(roomUrl)
						sonic.Unmarshal(r.Body(), &roomInfo)
						var serverStartAt = time.Now() //time.Parse(time.DateTime, roomInfo.Data.LiveTime)
						var foundLive = Live{}
						lives[roomId].Title = roomInfo.Data.Title
						db.Where("user_id=?", roomInfo.Data.UID).Last(&foundLive)
						var diff = abs(int(foundLive.StartAt - serverStartAt.Unix()))
						if diff-8*3600 > 60*15 && !lives[roomId].Live  {
							log.Println("[" + roomId + "]  " + msg)
							log.Println("[" + roomId + "]  " + string(msgData))
							v := serverStartAt
							v = v.Add(time.Hour * 8)
							new.StartAt = v.Unix()
							new.Title = roomInfo.Data.Title
							new.Area = roomInfo.Data.Area
							var i, _ = strconv.Atoi(roomId)
							new.RoomId = i
							new.UserName = liver
							lives[roomId].Live = true
							lives[roomId].StartAt = time.Now().Add(3600 * 8 * time.Second).Format(time.DateTime)
							liver = strings.TrimSpace(liver)
							db.Create(&new)
							dbLiveId = int(new.ID)
							var msg = "你关注的主播： " + liver + " 开始直播"
							PushDynamic(msg, roomInfo.Data.Title)
						} else {
							log.Println("LIVE FALSE")
						}
					*/

				} else if strings.Contains(obj, "SUPER_CHAT_MESSAGE") { //SC
					var sc = SuperChatInfo{}
					sonic.Unmarshal(msgData, &sc)

					action.ActionName = "sc"
					action.FromName = sc.Data.UserInfo.Uname
					action.FromId = strconv.Itoa(sc.Data.Uid)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sc.Data.Price), 64)
					action.GiftPrice = sql.NullFloat64{Float64: result, Valid: true}

					action.GiftAmount = sql.NullInt16{Valid: true, Int16: 1}
					action.Extra = sc.Data.Message
					if action.FromId != "0" {
						db.Create(&action)
					}
				} else if strings.Contains(obj, "GUARD_BUY") { //上舰
					var guard = GuardInfo{}
					sonic.Unmarshal(msgData, &guard)
					action.FromId = strconv.Itoa(guard.Data.Uid)
					action.ActionName = "guard"
					action.FromName = guard.Data.Username
					action.GiftName = guard.Data.GiftName
					switch action.GiftName {
					case "舰长":
						action.GiftPrice = sql.NullFloat64{Float64: float64(138 * guard.Data.Num), Valid: true}
					case "提督":
						action.GiftPrice = sql.NullFloat64{Float64: float64(1998 * guard.Data.Num), Valid: true}
					case "总督":
						action.GiftPrice = sql.NullFloat64{Float64: float64(19998 * guard.Data.Num), Valid: true}
					}

					db.Create(&action)
				} else if text.Cmd == "WATCHED_CHANGE" {
					if living {
						var obj = Watched{}
						sonic.Unmarshal(msgData, &obj)
						db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{Watch: obj.Data.Num})
					}
				}
				front.LiveAction = action
				if action.ActionName != "" {
					front.UUID = uuid.New().String()
					lives[roomId].Danmuku = AppendElement(lives[roomId].Danmuku, 500, front)
				}
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
			if isLive(roomId) {
				err = c.WriteMessage(websocket.TextMessage, BuildMessage("[object Object]", 2))
				//lives[roomId].LastActive = time.Now().Unix() + 3600*8
				if err != nil {
					log.Printf("[%s] write:  %v", liver, err)
					return
				}
			}
			url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + roomId
			res, _ = client.R().Get(url)
			status := LiveStatusResponse{}
			sonic.Unmarshal(res.Body(), &status)

			if status.Data.LiveStatus == 1 && !isLive(roomId) {

				//var sum float64
				//db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", dbLiveId).Scan(&sum)

				//db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{EndAt: time.Now().Unix(), Money: sum})
				living = true
				//i, _ := strconv.Atoi(roomId)
				var new = Live{}
				new.StartAt = time.Now().Unix() + 8*3600
				new.Title = roomInfo.Data.Title
				new.Area = roomInfo.Data.Area
				var i, _ = strconv.Atoi(roomId)
				new.RoomId = i
				new.UserName = liver
				new.UserID = liverId
				new.Cover = roomInfo.Data.Face
				lives[roomId].StartAt = time.Now().Format(time.DateTime)
				liver = strings.TrimSpace(liver)
				db.Create(&new)
				dbLiveId = int(new.ID) //似乎直播间的ws服务器有概率不发送开播消息，导致漏数据，这里做个兜底。
				var msg = "你关注的主播： " + liver + " 开始直播"
				lives[roomId].Title = roomInfo.Data.Title

				if Has(config.Tracing, roomId) {
					PushDynamic(msg, roomInfo.Data.Title)
				}
				log.Printf("[%s] 直播开始，连接ws服务器", liver)
				c, _, err = dialer.Dial(u.String(), nil)
				err = c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
				if err == nil {
					log.Printf("[%s] 连接成功", liver)
					setLive(roomId, true)
				}

			}
			e, ok := lives[roomId]
			if status.Data.LiveStatus != 1 && (isLive(roomId) || (ok && e.Live)) {
				setLive(roomId, false)
				var sum float64
				db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", dbLiveId).Scan(&sum)

				db.Model(&Live{}).Where("id= ?", dbLiveId).UpdateColumns(Live{EndAt: time.Now().Unix(), Money: sum})
				living = false
				i, _ := strconv.Atoi(roomId)
				if config.EnableLiveBackup {
					go UploadLive(Live{RoomId: i, UserName: liver})
				}
				log.Printf("[%s] 直播结束，断开连接", liver)
				c.Close()
				if !Has(config.Tracing, roomId) {
					log.Println("不在关注列表，结束")
					delete(lives, roomId)
					return
				}

			}
			lives[roomId].OnlineWatcher = GetOnline(roomId, liverId)
			go func() {
				lives[roomId].GuardList = GetGuard(roomId, liverId, false)
			}()
		}
	}
	ticker.Stop()
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

var mu sync.RWMutex
var GiftPic = make(map[string]string)

func FillGiftPrice(room string, area int, parent int) {
	res, _ := client.R().Get("https://api.live.bilibili.com/xlive/web-room/v1/giftPanel/roomGiftList?platform=pc&room_id=" + room + "&area_id=" + strconv.Itoa(area) + "&area_parent_id" + strconv.Itoa(parent))
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
				mu.Lock()
				GiftPrice[item0.GiftName] = float32(item0.Price) / 1000.0
				GiftPic[item0.GiftName] = item0.Picture
				mu.Unlock()
			}
		} else {
			mu.Lock()
			GiftPrice[item.Name] = float32(item.Price) / 1000.0
			GiftPic[item.Name] = item.Picture
			mu.Unlock()
		}

	}
	for i := range gift.Data.GiftConfig.RoomConfig {
		var item = gift.Data.GiftConfig.RoomConfig[i]
		mu.Lock()
		GiftPrice[item.Name] = float32(item.Price) / 1000.0
		mu.Unlock()
	}

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
					ID      int    `json:"id"`
					Name    string `json:"name"`
					Price   int    `json:"price"`
					Picture string `json:"webp"`
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
		GiftName string `json:"giftName"`
		Num      int    `json:"num"`
		Price    int    `json:"price"`
		Parent   struct {
			Price    int    `json:"original_gift_price"`
			GiftName string `json:"original_gift_name"`
		} `json:"blind_gift"`
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
		UID   int `json:"uid"`
		Medal struct {
			Name  string `json:"name"`
			Level int    `json:"level"`
			Color int    `json:"medal_color"`
		}
		Uname string `json:"uname"`
		Face  string `json:"face"`
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
			Picture  string `json:"webp"`
		} `json:"gifts"`
	} `json:"data"`
}
type LiveAction struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	Live       uint
	FromName   string
	FromId     string
	LiveRoom   string
	ActionName string
	GiftName   string
	GiftPrice  sql.NullFloat64 `gorm:"scale:2;precision:7"`
	GiftAmount sql.NullInt16
	Extra      string
	MedalName  string
	MedalLevel int8
	GuardLevel int8
	HonorLevel int8
}
type FrontLiveAction struct {
	LiveAction
	Face        string
	UUID        string
	MedalColor  string
	GiftPicture string
}
type RoomInfo struct {
	Data struct {
		LiveTime     string `json:"live_time"`
		UID          int64  `json:"uid"`
		Title        string `json:"title"`
		Area         string `json:"area_name"`
		AreaId       int    `json:"area_id"`
		ParentAreaId int    `json:"parent_area_id"`
		Face         string `json:"user_cover"`
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
	RoomId   int
	Money    float64 `gorm:"type:decimal(7,2)"`
	Message  int
	Watch    int
	Cover    string
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		UID       int    `json:"uid"`
		Uname     string `json:"uname"`
		FansMedal struct {
			MedalName string `json:"medal_name"`
			Level     int    `json:"medal_level"`
		} `json:"fans_medal"`
	} `json:"data"`
}

type LiverInfo struct {
	Data struct {
		Info struct {
			Uname string `json:"uname"`
		} `json:"info"`
	} `json:"data"`
}

type SuperChatInfo struct {
	Data struct {
		Message  string  `json:"message"`
		Price    float64 `json:"price"`
		Uid      int     `json:"uid"`
		UserInfo struct {
			Uname string `json:"uname"`
		} `json:"user_info"`
	} `json:"data"`
}

type GuardInfo struct {
	Data struct {
		Uid        int    `json:"uid"`
		Username   string `json:"username"`
		GuardLevel int    `json:"guard_level"`
		Num        int    `json:"num"`
		GiftName   string `json:"gift_name"`
	} `json:"data"`
}
type Watched struct {
	Data struct {
		Num       int    `json:"num"`
		TextSmall string `json:"text_small"`
		TextLarge string `json:"text_large"`
	} `json:"data"`
}
