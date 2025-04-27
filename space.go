package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
	"golang.org/x/net/html"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// 获取up主的粉丝数
func FetchUser(mid string, onError func()) User {
	var url = "https://api.bilibili.com/x/web-interface/card?mid=" + mid
	res, _ := client.R().SetHeader("User-Agent", USER_AGENT).Get(url)
	var userResponse = UserResponse{}
	sonic.Unmarshal(res.Body(), &userResponse)

	var user = User{}
	user.Name = userResponse.Data.Card.Name
	user.Face = userResponse.Data.Card.Face
	user.Fans = userResponse.Data.Followers

	user.UserID, _ = strconv.ParseInt(mid, 10, 64)
	if user.Fans == 0 {
		fmt.Println(string(res.Body()))
		if onError != nil {
			onError()
		}
	}
	return user

}

// 刷新粉丝数
func UpdateCommon() {
	for i := range Followings {
		if i > len(Followings)-1 {
			continue
		}
		var id = Followings[i].UserID
		var user = FetchUser(strconv.FormatInt(id, 10), nil)
		user.Face = ""
		db.Save(&user)
		time.Sleep(3 * time.Second)
	}
}

// 解析服务端返回的动态的json结构
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

// 刷新特别关注的up主的动态
func UpdateSpecial() {
	var flag = false
	if len(RecordedDynamic) == 0 {
		flag = true
	}
	for i := range config.SpecialList {
		var id = config.SpecialList[i]
		resp, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + (strconv.FormatInt(id, 10)) + "&timezone_offset=-480&features=itemOpusStyle")
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

func FetchComments() {

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
	db.Find(&livers)
	for _, liver := range livers {
		var user = User{}
		user.Name = liver.UName
		user.UserID = liver.UID
		Followings0 = append(Followings0, user)
	}
	Followings = Followings0
	Special = Special0
}

// GetCollectionId 获取当前用户的Monitor收藏夹的id
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
		for _, media := range medias.Data.Medias {

			var found = false
			for i := range RecordedMedias {
				if RecordedMedias[i] == media.BV {
					found = true
				}
			}
			if !found {
				RecordedMedias = append(RecordedMedias, media.BV)
				go func() {
					var link = UploadArchive(ParseSingleVideo(media.BV)[0])
					PushDynamic("你收藏的"+media.Title+"已下载完成", link)
				}()
			}
		}
	}

}

// 上传稿件到Alist
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
			out, _ := cmd.CombinedOutput()
			log.Println(string(out))
			cmd.Run() // 执行命令

			if video.ParentTitle != "" {
				video.ParentTitle = video.ParentTitle + "/"
			}
			//log.Println(string(out))
			final = config.AlistPath + "/Archive/" + video.Author + "/" + video.ParentTitle + "[" + strings.ReplaceAll(video.PublishAt, ":", "-") + "] " + video.Title + ".mp4"
			if audio.StatusCode() == 200 {
				UploadFile("cache/"+bv+".mp4", final)
			} else {

			}

			os.Remove("cache/" + bv + ".mp4")
			os.Remove("cache/" + bv + ".mp3")
			os.Remove("cache/" + bv + ".m4s")
		}
	})
	return config.AlistServer + final

}

func GetFace(uid string) string {
	var obj = FaceCache{}
	db.Model(&obj).Where("uid = ?", uid).First(&obj)
	if obj.UID == 0 || time.Now().Unix()-obj.UpdateAt.Unix() > 3600*24*7 {
		var url = "https://api.bilibili.com/x/web-interface/card?mid=" + uid
		res, _ := client.R().Get(url)
		var userResponse = UserResponse{}
		sonic.Unmarshal(res.Body(), &userResponse)
		if obj.Face != userResponse.Data.Card.Face {
			if obj.UID == 0 {
				obj.Face = userResponse.Data.Card.Face
				obj.UID, _ = strconv.ParseInt(uid, 10, 64)
				obj.UpdateAt = time.Now()
				db.Create(&obj)
			} else {
				if obj.Face != "" {
					db.Model(&FaceCache{}).Where("uid = ?", uid).UpdateColumns(FaceCache{Face: userResponse.Data.Card.Face})
				}
			}
		}
		return userResponse.Data.Card.Face
	} else {
		return obj.Face
	}
}
