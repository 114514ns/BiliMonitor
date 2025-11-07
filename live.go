package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	_ "runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	pool2 "github.com/sourcegraph/conc/pool"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SelfInfo struct {
	Data struct {
		Mid int64 `json:"mid"`
	} `json:"data"`
}

func SelfUID(cookie string) int64 {
	res, err := client.R().SetHeader("Cookie", cookie).Get("https://api.bilibili.com/x/web-interface/nav")

	if err != nil {
		fmt.Println(err)
	}
	var self = SelfInfo{}
	sonic.Unmarshal(res.Body(), &self)
	return self.Data.Mid
}
func GetServerState(room string) bool {
	url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + room
	res, err := queryClient.R().Get(url)
	if err != nil {
		log.Println(err)
	}
	status := LiveStatusResponse{}
	sonic.Unmarshal(res.Body(), &status)
	return status.Data.LiveStatus == 1
}
func RemoveEmpty() {
	//db.Where("money = 0 and message = 0").Delete(&Live{})
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

					man.AddTask(strconv.Itoa(live.RoomId))
					log.Printf("[%s] 恢复直播", live.UserName)

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
func GetOnline(room string, liver string) ([]Watcher, int) {
	var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/queryContributionRank?ruid=%s&room_id=%s", liver, room)
	res, _ := client.R().Get(url)
	var o = OnlineWatcherResponse{}
	sonic.Unmarshal(res.Body(), &o)
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
	return arr, o.Data.Count

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
func GetFansClub(liver string, callback func(g DBGuard)) []DBGuard {
	log.Printf("[%s] begin fetch fansClub", liver)
	var list = make([]DBGuard, 0)
	t := "0"
	u := fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/getFansMembersRank?ruid=%s&page=%s&page_size=30&rank_type=%s&ts=%s", liver, strconv.Itoa(1), t, strconv.FormatInt(time.Now().Unix(), 10))
	res, _ := queryClient.R().Get(u)
	obj := FansClubResponse{}
	var pool = pool2.New().WithMaxGoroutines(config.ConnectionPoolSize)
	sonic.Unmarshal(res.Body(), &obj)
	var totalPages = (obj.Data.Num + 29) / 30
	var mutex sync.Mutex
	//type=0是活跃的粉丝团用户
	//type=2是所有没上过舰长的粉丝团用户
	for i := 1; i <= totalPages; i++ {
		page := i
		var tmp []DBGuard
		pool.Go(func() {
			var retry = 3
			for {
				if retry < 0 {
					break
				}
				u := fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/getFansMembersRank?ruid=%s&page=%s&page_size=30&rank_type=%s&ts=%s", liver, strconv.Itoa(page), t, strconv.FormatInt(time.Now().Unix(), 10))
				res, _ := RandomPick(cPools).R().Get(u)
				obj := FansClubResponse{}
				sonic.Unmarshal(res.Body(), &obj)
				time.Sleep(time.Duration(config.RequestDelay) * time.Millisecond)
				if len(obj.Data.Item) == 0 {
					retry--
					continue
				}
				for _, s := range obj.Data.Item {
					var d = DBGuard{}
					d.Score = s.Score
					d.Level = s.Level
					d.Type = s.Medal.Type
					d.UID = s.UID
					d.UName = s.UName
					d.MedalName = s.Medal.Name
					tmp = append(tmp, d)
					if callback != nil {
						callback(d)
					}
				}
				mutex.Lock()
				list = append(list, tmp...)
				mutex.Unlock()
				time.Sleep(time.Duration(config.RequestDelay) * time.Millisecond)
				break
			}

		})
	}

	pool.Wait()
	log.Printf("[%s] end fetch fansClub，size=%d", liver, len(list))
	return list
}
func GetGuardList(room string, liver string) []Watcher {
	var arr = make([]Watcher, 0)
	var pool = pool2.New().WithMaxGoroutines(8)
	var mutex sync.Mutex
	var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/app-room/v2/guardTab/topListNew?roomid=%s&page=%d&ruid=%s&page_size=30", room, 1, liver)
	res, _ := queryClient.R().Get(url)
	var r = GuardListResponse{}
	sonic.Unmarshal(res.Body(), &r)
	var totalPage = r.Data.Info.Page
	for i := 1; i <= totalPage; i++ {
		page := i
		pool.Go(func() {
			var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/app-room/v2/guardTab/topListNew?roomid=%s&page=%d&ruid=%s&page_size=30", room, page, liver)
			res, _ := RandomPick(cPools).R().Get(url)
			var r = GuardListResponse{}
			sonic.Unmarshal(res.Body(), &r)
			var tmp []Watcher
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
					tmp = append(tmp, watcher)
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
				tmp = append(tmp, watcher)
			}
			mutex.Lock()
			arr = append(arr, tmp...)
			mutex.Unlock()
			time.Sleep(time.Duration(config.RequestDelay) * time.Millisecond)
		})
	}
	pool.Wait()
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
	ID       uint `gorm:"primarykey"`
	Time     time.Time
	UName    string
	UID      int64
	Room     int
	Title    string
	Area     string
	Watch    int
	LastSeen time.Time
	Duration int
	Cover    string
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

var guardWorker = NewWorker(1)

var guardWorkerMap = make(map[int64]time.Time)

func TraceArea(parent int, full bool) {
	log.Println("begin TraceArea")
	var lock sync.Mutex
	if working && full { //上一个没有爬完，而且不是全的模式，就跳过这次
		log.Println("TraceArea is still executing,break")
		return //确保不会重叠执行
	}
	var page = 1
	defer func() {
		if full {
			log.Println("set work false")
			working = false
		}
	}()
	var arr = make([]AreaLiver, 0)
	for {
		type SortInfo struct {
			Room string
			Time int64
		}
		if full {
			working = true
		}
		u, _ := url.Parse(fmt.Sprintf("https://api.live.bilibili.com/xlive/app-interface/v2/second/getList?area_id=0&build=1001016004&device=win&page=%d&parent_area_id=%d&platform=web&web_location=bilibili-electron", page, parent))
		var now = time.Now()
		s, _ := wbi.SignQuery(u.Query(), now)
		res, _ := client.R().SetHeader("User-Agent", USER_AGENT).SetHeader("Cookie", config.Cookie).Get("https://api.live.bilibili.com/xlive/app-interface/v2/second/getList?" + s.Encode())
		obj := AreaLiverListResponse{}
		var m = make([]SortInfo, 0)
		sonic.Unmarshal(res.Body(), &obj)

		log.Printf("page=%d,len=%d", page, len(obj.Data.List))
		if page == 1 && len(obj.Data.List) == 0 {
			log.Println(res.String())
		}
		go func() {
			var sum = 0
			for _, node := range man.Nodes {
				sum += len(node.Tasks)
			}
			for _, s2 := range obj.Data.List {
				if sum <= MAX_TASK*len(man.Nodes) {
					o := AreaLiver{}
					db.Model(&AreaLiver{}).Where("uid = ?", s2.UID).Last(&o)
					var fans = o.Fans
					if fans == 0 {
						user := FetchUser(strconv.FormatInt(s2.UID, 10), nil)
						fans = user.Fans
					}
					var hour = time.Now().Hour()
					//白天2500粉丝以上被爬取，晚上8000粉以上，10舰长以上
					if fans > 4000 {
						if hour > 19 {
							var guards = 0
							for _, i := range strings.Split(o.Guard, ",") {
								guards += int(toInt64(i))
							}
							if guards > 10 {
								m = append(m, SortInfo{Room: strconv.Itoa(s2.Room)})
							}
						} else {
							m = append(m, SortInfo{Room: strconv.Itoa(s2.Room)})
						}

					} else {
						if hour > 1 && hour < 17 {
							if fans > 2500 {
								m = append(m, SortInfo{Room: strconv.Itoa(s2.Room)})
							}
						}
					}
				}
			}
			for _, info := range m {
				man.AddTask(info.Room)
			}
		}()

		if full {
			for _, s2 := range obj.Data.List {
				t, ok := guardWorkerMap[s2.UID]
				if !ok || time.Since(t).Hours() > 12 {
					guardWorkerMap[s2.UID] = time.Now()
					go func() {
						guardWorker.AddTask(func() {
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
							db.Model(&AreaLive{}).Where("uid = ?", s2.UID).Last(&live)
							if live.Time.Unix() != time.Unix(info.Data.Time, 0).Unix() {

								var l = AreaLive{}
								l.UName = s2.UName
								l.UID = s2.UID
								l.Room = s2.Room
								l.Cover = s2.Cover
								l.Title = s2.Title
								l.Area = s2.Area
								l.Watch = s2.Watch.Num
								l.Time = time.Unix(info.Data.Time, 0)
								l.LastSeen = time.Now()
								live.Duration = int(live.LastSeen.Sub(live.Time).Minutes())
								db.Save(&l)
							} else {
								live.Watch = s2.Watch.Num
								live.LastSeen = time.Now()
								live.Duration = int(live.LastSeen.Sub(live.Time).Minutes())
								db.Save(&live)
							}
							//log.Printf("current Liver %s", s2.UName)
							db.Model(&AreaLiver{}).
								Where("uid = ?", s2.UID).
								Order("id DESC").
								First(&found)
							//如果这个主播在数据库里没有，或者上次更新超过两天，就更新一下
							if found.UID == 0 || time.Now().Unix()-found.UpdatedAt.Unix() > 3600*36 {
								if found.UID == 0 {
									//如果这个主播在数据库里没有，就获取活跃的粉丝团用户列表，和免费的粉丝团用户是数量
									GetFansClub(strconv.FormatInt(s2.UID, 10), func(g DBGuard) {
										club := FansClub{}
										lock.Lock()
										fansMap[g.UID] = g
										lock.Unlock()
										club.DBGuard = g
										club.Liver = s2.UName
										club.LiverID = s2.UID

										db.Save(&club)
									})

									i.FreeClubs = GetFreeClubCounts(strconv.FormatInt(s2.UID, 10), 500)

								} else {
									//数据库里已经有这个主播的情况下只要更新活跃的用户就可以了
									GetFansClub(strconv.FormatInt(s2.UID, 10), func(g DBGuard) {
										club := FansClub{}
										lock.Lock()
										fansMap[g.UID] = g
										lock.Unlock()
										db.Model(&FansClub{}).Where("uid = ? and liver_id = ?", g.UID, s2.UID).Last(&club)
										club.DBGuard = g
										club.Liver = s2.UName
										club.LiverID = s2.UID
										if club.Score != 0 {
											club.Score = g.Score //如何数据库里这名用户没有这位主播的粉丝牌，就插入一条记录
											db.Save(&club)
										} else {
											lock.Lock()
											fansMap[g.UID] = g
											lock.Unlock()
											db.Save(&club)
										}
									})

								}
								log.Printf("[%s] begin fetch guardList", s2.UName)
								var user = FetchUser(strconv.FormatInt(s2.UID, 10), nil)
								user.Face = ""
								var guards = make([]DBGuard, 0)
								var l1 = 0
								var l2 = 0
								var l3 = 0
								var size = 0
								for size0, watcher := range GetGuardList(strconv.Itoa(i.Room), strconv.FormatInt(s2.UID, 10)) {
									size = size0
									ins := DBGuard{}
									ins.Type = watcher.Guard
									ins.UID = watcher.UID
									ins.UName = watcher.Name
									ins.Level = watcher.Medal.Level
									ins.Liver = s2.UName
									ins.LiverID = s2.UID
									ins.MedalName = watcher.Medal.Name
									lock.Lock()
									ins.Score = fansMap[watcher.UID].Score
									lock.Unlock()
									if ins.Score == 0 {
										var found FansClub
										db.Model(&FansClub{}).Where("uid = ? and liver_id = ? ", ins.UID, ins.LiverID).Order("id desc").Limit(1).First(&found)
										ins.Score = found.Score
									}

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
								log.Printf("[%s] finish fetch guardList,size=%d", s2.UName, size)
								i.Guard = fmt.Sprintf("%d,%d,%d", l1, l2, l3)
								b, _ := sonic.Marshal(guards)
								i.GuardList = string(b)
								db.Save(&user)
								i.Fans = user.Fans
								time.Sleep(time.Second * 2)
								db.Save(&i)
								log.Printf("end %s", s2.UName)
							}
						})
					}()
				}

				//log.Printf("finish update %s", s2.UName)
			}
		}

		log.Printf("page=%d,More=%d", page, obj.Data.More)
		if len(obj.Data.List) == 0 {
			break
		}
		if !full && page >= 25 {
			return
		}
		page++
		time.Sleep(time.Second * 1)
	}
	working = false
}
func BuildAuthMessage(room string) string {
	url0 := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + room + "&web_location=444.8&isGaiaAvoided=true"
	query, _ := url.Parse(url0)
	signed, _ := wbi.SignQuery(query.Query(), time.Now())
	res, e := client.R().SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?" + signed.Encode())
	if e != nil {
		fmt.Println(e)
	}
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

func isLive(roomId string) bool {
	livesMutex.Lock()
	s, ok := lives[roomId]
	livesMutex.Unlock()

	if !ok || s == nil {
		return false
	}

	s.RLock()
	defer s.RUnlock()
	return s.Live
}

func setLive(roomId string, live bool) {
	lives[roomId].Lock()
	defer lives[roomId].Unlock()
	if s, ok := lives[roomId]; ok {
		s.Live = live
	}
}
func RandomHost() string {
	var HOST = []string{"zj-cn-live-comet.chat.bilibili.com", "bd-bj-live-comet-06.chat.bilibili.com", "bd-sz-live-comet-10.chat.bilibili.com", "broadcastlv.chat.bilibili.com"}
	return HOST[rand.Intn(len(HOST))]

}

func TraceLive(roomId string) {

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			fmt.Println("Recovered from panic:", r)
		}
	}()

	var WS_HEADER = http.Header{}
	WS_HEADER.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	var roomUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + roomId
	var rRes, _ = client.R().Get(roomUrl)
	var liver string
	var roomInfo = RoomInfo{}
	err := sonic.Unmarshal(rRes.Body(), &roomInfo)
	if err != nil {
		return
	}
	FillGiftPrice(roomId, roomInfo.Data.AreaId, roomInfo.Data.ParentAreaId)
	var dbLiveId = 0
	var liverId = strconv.FormatInt(roomInfo.Data.UID, 10)
	if roomInfo.Data.UID == 0 {
		log.Println(rRes.String())
	}
	var startAt = roomInfo.Data.LiveTime

	livesMutex.Lock()
	lives[roomId].Live = GetServerState(roomId)
	var living = lives[roomId].Live
	livesMutex.Unlock()
	var liverInfoUrl = "https://api.live.bilibili.com/live_user/v1/Master/info?uid=" + liverId
	liverRes, _ := client.R().Get(liverInfoUrl)
	var liverObj = LiverInfo{}
	sonic.Unmarshal(liverRes.Body(), &liverObj)
	liver = liverObj.Data.Info.Uname
	lives[roomId].UName = liver

	var faceUrl = "https://api.bilibili.com/x/web-interface/card?mid=" + liverId

	var faceRes, _ = client.R().SetHeader("User-Agent", USER_AGENT).Get(faceUrl)

	time.Sleep(1 * time.Second)

	var areaLiver AreaLiver
	db.Raw("select fans from area_livers where uid = ?", liverId).Scan(&areaLiver)
	lives[roomId].Fans = areaLiver.Fans

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

		//当前是开播状态
		var serverStartAt, _ = time.Parse(time.DateTime, startAt)
		lives[roomId].StartAt = startAt

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
			new.ParentAreaID = int16(roomInfo.Data.ParentAreaId)
			new.AreaID = int16(roomInfo.Data.AreaId)
			//new.UserName = roomInfo.Data

			var i, _ = strconv.Atoi(roomId)
			new.RoomId = i
			new.UserName = liver
			liver = strings.TrimSpace(liver) // 去除前后的空白字符

			db.Create(&new)
			dbLiveId = int(new.ID)
		}
	}

	url0 := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + roomId + "&web_location=444.8&isGaiaAvoided=true"
	query, _ := url.Parse(url0)
	signed, _ := wbi.SignQuery(query.Query(), time.Now())
	res, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("User-Agent", USER_AGENT).Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?" + signed.Encode())
	var liveInfo = LiveInfo{}
	sonic.Unmarshal(res.Body(), &liveInfo)
	if len(liveInfo.Data.HostList) == 0 {
		log.Println("error,break" + res.String())
		return
	}
	u := url.URL{Scheme: "wss", Host: liveInfo.Data.HostList[0].Host + ":2245", Path: "/sub"}
	var dialer = &websocket.Dialer{
		Proxy:            nil,
		HandshakeTimeout: 45 * time.Second,
	}
	if config.HTTPProxy != "" {
		u, _ := url.Parse(config.HTTPProxy)
		dialer.Proxy = http.ProxyURL(u)
	}
	var c *websocket.Conn
	if lives[roomId].Live {

		c, _, err = dialer.Dial(u.String(), WS_HEADER)
	}
	if err != nil {
		log.Println("["+liver+"]  "+"Dial:", err)
	}
	ticker := time.NewTicker(45 * time.Second)

	SafeGoRetry(func() {
		if lives[roomId].Live {

			err := c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
			if err != nil {
				return
			}
			log.Printf("[%s] 成功连接到弹幕服务器", liver)
		}
		for {
			var msg = ""
			if isLive(roomId) {
				_, message, err := c.ReadMessage()
				if err != nil && isLive(roomId) {
					log.Printf("[%s] 断开连接，尝试重连次数："+strconv.FormatInt(int64(lives[roomId].RemainTrying), 10), liver)
					for {
						if lives[roomId].RemainTrying > 0 {
							time.Sleep(time.Duration(rand.Int()%10000) * time.Millisecond)
							c, _, err := dialer.Dial(u.String(), WS_HEADER)
							lives[roomId].RemainTrying--
							if err != nil {
								log.Println(err)
							}
							err = c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
							if err == nil {
								log.Printf("[%s] 重连成功", liver)
								break
							} else {
								log.Printf("[%s] %v", liver, err)
							}
						}
					}
					return
				}
				websocketBytes += int64(len(message))
				reader := io.NewSectionReader(bytes.NewReader(message), 16, int64(len(message)-16))
				brotliReader := brotli.NewReader(reader)
				var decompressedData bytes.Buffer
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
				action.LiveRoom, _ = strconv.Atoi(roomId)
				action.GiftPrice = sql.NullFloat64{Float64: 0, Valid: false}
				action.GiftAmount = sql.NullInt16{Int16: 0, Valid: false}
				var text = LiveText{}
				sonic.Unmarshal(msgData, &text)
				var front = FrontLiveAction{}
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") { // 弹幕
					action.ActionName = "msg"
					action.ActionType = Message
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = int64(text.Info[2].([]interface{})[0].(float64))
					action.Extra = text.Info[1].(string)
					action.HonorLevel = int8(text.Info[16].([]interface{})[0].(float64))
					front.Emoji = make(map[string]string)
					value, ok := text.Info[0].([]interface{})[15].(map[string]interface{})
					e1, ok := text.Info[0].([]interface{})[13].(map[string]interface{})
					if ok {
						e2, ok := e1["emoticon_unique"].(string)
						if ok {
							front.Emoji[strings.Replace(e2, "upower_", "", 1)] = e1["url"].(string)
						}
					}
					var o interface{}
					sonic.Unmarshal([]byte(text.Info[0].([]interface{})[15].(map[string]interface{})["extra"].(string)), &o)
					e, ok := o.(map[string]interface{})["emots"]
					if e != nil {
						emots := e.(map[string]interface{})
						if len(emots) != 0 {
							for s, i := range emots {
								front.Emoji[s] = i.(map[string]interface{})["url"].(string)
							}
						}
					}
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
								ruid, exists := medal["ruid"]
								if exists {
									action.MedalLiver = int64(ruid.(float64))
								}

							}
						}
					}
					go func() {
						db.Create(&action)
					}()
					go func() {
						var dst FaceCache
						db.Raw("SELECT * FROM face_caches where uid = ?", action.FromId)
						if dst.UID == 0 {
							db.Create(FaceCache{Face: front.Face, UID: action.FromId})
						} else {
							db.Model(&FaceCache{}).Where("uid = ?", action.FromId).UpdateColumns(FaceCache{Face: front.Face})
						}
					}()
					consoleLogger.Println("[" + liver + "]  " + text.Info[2].([]interface{})[1].(string) + "  " + text.Info[1].(string))

				} else if strings.Contains(obj, "SEND_GIFT") { //送礼物
					var info = GiftInfo{}
					sonic.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.ActionType = Gift
					action.MedalLevel = int8(info.Data.Medal.Level)
					action.HonorLevel = info.Data.HonorLevel
					action.MedalName = info.Data.Medal.Name
					action.FromId = info.Data.SenderUinfo.UID
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
					go func() {
						db.Create(&action)
					}()
					consoleLogger.Printf("[%s] %s 投喂了 %d 个 %s，%.2f元", liver, info.Data.Uname, info.Data.Num, info.Data.GiftName, price)
				} else if strings.Contains(obj, "INTERACT_WORD") { //进入直播间
					var enter = EnterLive{}
					sonic.Unmarshal(msgData, &enter)
					action.FromId = enter.Data.UID
					action.FromName = enter.Data.Uname
					action.ActionName = "enter"
					go func() {
						db.Table("enter_action").Create(&action)
					}()
				} else if strings.Contains(obj, "PREPARING") {
				} else if text.Cmd == "LIVE" {
				} else if text.Cmd == "SUPER_CHAT_MESSAGE" { //SC
					var sc = SuperChatInfo{}
					sonic.Unmarshal(msgData, &sc)
					action.ActionType = SuperChat
					action.ActionName = "sc"
					action.FromName = sc.Data.UserInfo.Uname
					action.FromId = sc.Data.Uid
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sc.Data.Price), 64)
					action.GiftPrice = sql.NullFloat64{Float64: result, Valid: true}
					action.MedalName = sc.Data.MedalInfo.MedalName
					action.MedalLiver = sc.Data.MedalInfo.MedalLiver
					action.MedalLevel = int8(sc.Data.MedalInfo.MedalLevel)
					action.GuardLevel = int8(sc.Data.MedalInfo.GuardLevel)
					action.GiftAmount = sql.NullInt16{Valid: true, Int16: 1}
					action.Extra = sc.Data.Message
					if action.FromId != 0 {
						db.Create(&action)
					}
				} else if text.Cmd == ("GUARD_BUY") { //上舰
					//GUARD_BUY不返回粉丝牌信息
					var guard = GuardInfo{}
					action.ActionType = Guard
					sonic.Unmarshal(msgData, &guard)
					action.FromId = guard.Data.Uid
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
				} else if text.Cmd == "CUT_OFF" {
					action.ActionName = "cut"
					action.ActionType = Cut
					var o = make(map[string]interface{})
					sonic.Unmarshal(msgData, &o)
					action.Extra = o["msg"].(string)
					db.Save(action)
				} else if text.Cmd == "ROOM_BLOCK_MSG" {
					action.ActionName = "mute"
					var o = make(map[string]interface{})
					sonic.Unmarshal(msgData, &o)
					action.FromId = toInt64(o["uid"].(string))
					action.FromName = o["uname"].(string)
					action.ActionType = Block
					db.Save(action)
				} else if text.Cmd == "ANCHOR_LOT_START" {
					PushServerHime(liver+"直播间发起了天选抽奖", "")
				}
				front.LiveAction = action
				if action.ActionName != "" {
					/*
						front.UUID = uuid.New().String()
						_, ok := lives[roomId]
						if ok {
							lives[roomId].Danmuku = AppendElement(lives[roomId].Danmuku, 500, front)
						}


					*/
				}
				if buffer.Len() < 16 {
					break
				}

			}
			if !strings.Contains(msg, "[object") {

				//log.Printf("Received: %s", substr(msg, 16, len(msg)))
			}

		}
	}, 5, time.Second*10)

	for {
		select {
		case <-ticker.C:
			if isLive(roomId) {
				err = c.WriteMessage(websocket.TextMessage, BuildMessage("[object Object]", 2))
				//lives[roomId].LastActive = time.Now().Unix() + 3600*8
				if err != nil {
					log.Printf("[%s] write:  %v", liver, err)
					c, _, err = dialer.Dial(u.String(), WS_HEADER)
					err = c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
					if err == nil {
						log.Printf("[%s] 重新连接成功", liver)
						setLive(roomId, true)
					}
				}
			}
			url := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=" + roomId
			res, _ = client.R().Get(url)
			status := LiveStatusResponse{}
			sonic.Unmarshal(res.Body(), &status)

			if status.Data.LiveStatus == 1 && !isLive(roomId) {
				rRes, _ = client.R().Get(roomUrl)
				sonic.Unmarshal(rRes.Body(), &roomInfo)
				log.Printf("[%s] 直播开始，连接ws服务器,id=%d", liver, dbLiveId)
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
				livesMutex.Lock()
				_, ok := lives[roomId]
				if !ok {
					log.Println(roomId)
					log.Println(lives)
					livesMutex.Unlock()
					return
				}
				lives[roomId].StartAt = time.Now().Format(time.DateTime)
				lives[roomId].Title = roomInfo.Data.Title
				livesMutex.Unlock()
				liver = strings.TrimSpace(liver)
				db.Create(&new)
				dbLiveId = int(new.ID) //似乎直播间的ws服务器有概率不发送开播消息，导致漏数据，这里做个兜底。
				var msg = "你关注的主播： " + liver + " 开始直播"

				if Has(config.Tracing, roomId) {
					PushDynamic(msg, roomInfo.Data.Title)
				}

				c, _, err = dialer.Dial(u.String(), WS_HEADER)
				err = c.WriteMessage(websocket.TextMessage, BuildMessage(BuildAuthMessage(roomId), 7))
				if err == nil {
					log.Printf("[%s] 连接成功", liver)
					setLive(roomId, true)
				}

			}
			e, ok := lives[roomId]
			if status.Data.LiveStatus != 1 && (isLive(roomId) || (ok && e.Live)) {
				if status.Message != "" {
					setLive(roomId, false)
					var sum float64
					db.Table("live_actions").Select("SUM(gift_price)").Where("live = ?", dbLiveId).Scan(&sum)
					tx := db.Model(&Live{}).Where("id= ?", dbLiveId).Updates(map[string]interface{}{
						"end_at": time.Now().Unix(),
					})
					tx = db.Model(&Live{}).Where("id= ?", dbLiveId).Updates(map[string]interface{}{
						"money": sum,
					})
					var msg = 0
					db.Raw("select count(*) from live_actions where live = ?", dbLiveId).Scan(&msg)
					db.Raw("update lives set message = ? where id = ?", msg, dbLiveId).Scan(&msg)
					if tx.Error != nil {
						log.Println(tx.Error)
					}
					living = false
					log.Printf("[%s] 直播结束，断开连接,live=%d", liver, dbLiveId)
					c.Close()
					if !Has(config.Tracing, roomId) {
						log.Println("不在关注列表，结束")
						livesMutex.Lock()
						delete(lives, roomId)
						livesMutex.Unlock()
						return
					}
				}
			} else {
				if rand.Int()%5 == 0 {
					go func() {
						var msg = 0
						var money = 0.0
						db.Raw("select count(*) from live_actions where live = ?", dbLiveId).Scan(&msg)
						db.Raw("select sum(gift_price) from live_actions where live = ?", dbLiveId).Scan(&money)
						db.Raw("update lives set money = ? where id = ?", money, dbLiveId).Scan(&msg)
						db.Raw("update lives set message = ? where id = ?", msg, dbLiveId).Scan(&msg)
					}()
				}
				if isLive(roomId) || (ok && e.Live) {
					_, count := GetOnline(roomId, liverId)
					var obj = OnlineStatus{
						Live:  dbLiveId,
						Count: count,
						Time:  time.Now(),
					}
					db.Save(&obj)
				}
			}
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
	Uid      int64  `json:"uid"`
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
			UID int64 `json:"uid"`
		} `json:"sender_uinfo"`
		UID   int `json:"uid"`
		Medal struct {
			Name  string `json:"medal_name"`
			Level int    `json:"medal_level"`
			Color int    `json:"medal_color"`
		} `json:"medal_info"`
		Uname      string `json:"uname"`
		Face       string `json:"face"`
		HonorLevel int8   `json:"wealth_level"`
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

const (
	Message   int = 1
	Gift      int = 2
	Guard     int = 3
	SuperChat int = 4
	Cut       int = 5
	Block     int = 6
)

type LiveAction struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	Live       uint
	FromName   string
	FromId     int64
	LiveRoom   int
	ActionName string
	ActionType int
	GiftName   string
	GiftPrice  sql.NullFloat64 `gorm:"scale:2;precision:7"`
	GiftAmount sql.NullInt16
	Extra      string
	MedalName  string
	MedalLevel int8
	GuardLevel int8
	HonorLevel int8
	MedalLiver int64
}
type FrontLiveAction struct {
	LiveAction
	Face        string
	UUID        string
	MedalColor  string
	GiftPicture string
	Emoji       map[string]string
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
type OnlinePoint struct {
	Time   time.Time
	Online int
}
type Live struct {
	gorm.Model
	Title          string
	StartAt        int64
	EndAt          int64
	UserName       string
	UserID         string
	Area           string
	RoomId         int
	Money          float64 //`gorm:"type:decimal(7,2)"`
	Message        int
	Watch          int
	Cover          string
	AreaID         int16
	ParentAreaID   int16
	OnlinePoints   datatypes.JSON
	SuperChatMoney float64
	GuardMoney     float64
	BoxDiff        float64
}
type OnlineStatus struct {
	ID    int `gorm:"primarykey"`
	Live  int
	Count int
	Time  time.Time
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		UID       int64  `json:"uid"`
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
		Uid      int64   `json:"uid"`
		UserInfo struct {
			Uname string `json:"uname"`
		} `json:"user_info"`
		MedalInfo struct {
			MedalLiver int64  `json:"target_id"`
			MedalName  string `json:"medal_name"`
			MedalLevel int    `json:"medal_level"`
			GuardLevel int    `json:"guard_level"`
		} `json:"medal_info"`
	} `json:"data"`
}

type GuardInfo struct {
	Data struct {
		Uid        int64  `json:"uid"`
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

func SafeGoRetry(fn func(), maxRetries int, retryDelay time.Duration) {
	go func() {
		for i := 0; i <= maxRetries; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("goroutine panic (attempt %d): %v", i+1, r)
					}
				}()

				fn()
				return
			}()

			if i < maxRetries {
				time.Sleep(retryDelay)
			}
		}
	}()
}
