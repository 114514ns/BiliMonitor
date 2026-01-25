package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	pool2 "github.com/sourcegraph/conc/pool"
)

// 获取up主的粉丝数
func FetchUser(mid string, onError func()) User {
	var url = "https://api.bilibili.com/x/web-interface/card?mid=" + mid
	res, err := RandomPick(cPools).R().SetHeader("User-Agent", USER_AGENT).Get(url)
	if err != nil {
		log.Println(err)
	}
	var userResponse = UserResponse{}
	sonic.Unmarshal(res.Body(), &userResponse)

	var user = User{}
	user.Name = userResponse.Data.Card.Name
	user.Face = userResponse.Data.Card.Face
	user.Fans = userResponse.Data.Followers
	user.Bio = userResponse.Data.Card.Bio
	user.Verify = userResponse.Data.Card.Verify.Content

	user.UserID, _ = strconv.ParseInt(mid, 10, 64)
	if user.Fans == 0 {
		fmt.Println(string(res.Body()))
		if onError != nil {
			onError()
		}
	}
	return user

}

var commonDone = true

// 刷新粉丝数
func UpdateCommon() {
	var start = time.Now()
	var pool = pool2.New().WithMaxGoroutines(config.ConnectionPoolSize)
	var dst []AreaLiver
	db.Raw("SELECT fans,uid FROM area_livers where fans > 1500 GROUP BY uid").Scan(&dst)
	for i := range dst {
		if i > len(Followings)-1 {
			continue
		}
		var id = dst[i].UID
		pool.Go(func(uid int64) func() {
			return func() {
				var user = FetchUser(strconv.FormatInt(uid, 10), nil)
				user.Face = ""
				if user.Fans != 0 {
					db.Save(&user)
				}
				time.Sleep(2 * time.Second)
			}
		}(id))

	}

	pool.Wait()
	log.Println("[UpdateCommon] finished ", time.Since(start))
	commonDone = true
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
			var t = ""
			if len(item.Modules.ModuleDynamic.Desc.Nodes) > 0 {
				t = item.Modules.ModuleDynamic.Desc.Nodes[0].Text
			}
			PushDynamic("你关注的up主："+userName+"转发了动态 ", t)
		}
	} else if Type == "DYNAMIC_TYPE_AV" { //发布视频
		archive.Type = "v"
		archive.BiliID = item.IDStr
		archive.Title = item.Modules.ModuleDynamic.Major.Archive.Title
		if push {
			PushDynamic("你关注的up主："+userName+"投稿了视频 ", archive.Title)
		}

	} else if Type == "DYNAMIC_TYPE_DRAW" { //图文
		archive.Type = "i"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Desc.Text
		if push {
			PushDynamic("你关注的up主："+userName+"发布了动态 ", item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
		}

	} else if Type == "DYNAMIC_TYPE_WORD" { //文字
		archive.Type = "t"
		archive.BiliID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
		if push {
			PushDynamic("你关注的up主："+userName+"发布了动态 ", item.Modules.ModuleDynamic.Major.Opus.Summary.Text)
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

// 刷新特别关注的up主的动态
func UpdateSpecial() {
	var flag = false
	if len(RecordedDynamic) == 0 {
		flag = true
	}
	for i := range config.SpecialList {
		var id = config.SpecialList[i]
		resp, _ := client.R().SetHeader("Cookie", PickCookie()).SetHeader("Referer", "https://www.bilibili.com/").Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + (strconv.FormatInt(id, 10)) + "&timezone_offset=-480&features=itemOpusStyle")
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
						if d.Type == "v" {
							d.Download = true
						}
						db.Save(&d)
					}

					json := make([]byte, 0)
					sonic.Unmarshal(json, item)
				}
			}
		}
		time.Sleep(time.Second * 10)
	}
	//log.Printf(RecordedDynamic)
}

// RefreshFollowings 刷新用户关注列表
func RefreshFollowings() {

	var Followings0 = make([]User, 0)
	var page = 1
	var Special0 = make([]User, 0)
	for true {
		resp, err := client.R().Get("https://line3-h5-mobile-api.biligame.com/game/center/h5/user/relationship/following_list?vmid=" + string(config.User) + "&ps=50&pn=" + strconv.Itoa(page))
		if err != nil {
			log.Println(err)
		}
		var list = FansList{}
		sonic.Unmarshal(resp.Body(), &list)
		var users = list.Data.List
		for i := 0; i < len(users); i++ {
			var user = User{}
			user.Name = users[i].Uname
			user.UserID, _ = strconv.ParseInt(users[i].Mid, 10, 64)
			Followings0 = append(Followings0, user)
			for j := 0; j < len(config.SpecialList); j++ {
				if config.SpecialList[j] == user.UserID {
					Special0 = append(Special0, user)
				}
			}
		}
		if len(users) == 0 {
			break
		}
		page++
	}
	var livers = make([]AreaLiver, 0)
	db.Model(&AreaLiver{}).Where("fans").Group("uid").Find(&livers)
	for _, liver := range livers {
		var user = User{}
		user.Name = liver.UName
		user.UserID = liver.UID
		Followings0 = append(Followings0, user)
	}
	Followings = Followings0
	Special = Special0
}

func GetFace(uid string) string {
	var obj = FaceCache{}
	db.Model(&obj).Where("uid = ?", uid).First(&obj)
	if obj.UID == 0 || time.Now().Unix()-obj.UpdateAt.Unix() > 3600*24*30 {
		var s = "https://api.live.bilibili.com/xlive/fuxi-interface/UserService/getUserInfo?_ts_rpc_args_=[[" + uid
		s = s + `],true,""]`
		res, _ := client.R().Get(s)
		type Response struct {
			TsRpcReturn struct {
				Data map[string]struct {
					UID   string `json:"uid"`
					UName string `json:"uname"`
					Face  string `json:"face"`
				} `json:"data"`
			} `json:"_ts_rpc_return_"`
		}

		var r = Response{}
		sonic.Unmarshal(res.Body(), &r)
		if obj.Face != r.TsRpcReturn.Data[uid].Face {
			if obj.UID == 0 {
				obj.Face = r.TsRpcReturn.Data[uid].Face
				obj.UID, _ = strconv.ParseInt(uid, 10, 64)
				obj.UpdateAt = time.Now()
				db.Create(&obj)
			} else {
				if obj.Face != "" {
					db.Model(&FaceCache{}).Where("uid = ?", uid).UpdateColumns(FaceCache{Face: r.TsRpcReturn.Data[uid].Face})
				}
			}
		}
		return r.TsRpcReturn.Data[uid].Face
	} else {
		return obj.Face
	}
}

func GetFansLocal(mid int64) int {
	var found User
	db.Model(&User{}).Where("user_id = ?", mid).Last(&found)
	return found.Fans
}

func GetCharge(uid int64) []ChargeInfo {

	var array []ChargeInfo

	var url = fmt.Sprintf("https://api.bilibili.com/x/upower/up/member/rank/v2?pn=1&privilege_type=10&ps=100&up_mid=%d", uid)

	res, _ := queryClient.R().Get(url)
	var obj map[string]interface{}
	sonic.Unmarshal(res.Body(), &obj)
	for _, o := range getArray(obj, "data.level_info") {
		array = append(array, ChargeInfo{
			Name:  getString(o, "name"),
			Price: float64(getInt(o, "price") / 100.0),
			Count: getInt(o, "member_total"),
		})
	}

	return array
}
