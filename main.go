package main

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"os"
	"strconv"
	"strings"
)

// TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.
var client = resty.New()
var Cookie = ""
var Special = make([]User, 0)
var RecordedDynamic = make([]string, 0)

type Config struct {
	SpecialDelay           string
	CommonDelay            string
	RefreshFollowingsDelay string
	User                   string
	SpecialList            []int
	Cookie                 string
}

type FansList struct {
	Code int `json:"code"`
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
type UserState struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid       int `json:"mid"`
		Following int `json:"following"`
		Whisper   int `json:"whisper"`
		Black     int `json:"black"`
		Follower  int `json:"follower"`
	} `json:"data"`
}
type Basic struct {
	CommentIDStr string `json:"comment_id_str"`
	CommentType  int    `json:"comment_type"`
	LikeIcon     struct {
		ActionURL string `json:"action_url"`
		EndURL    string `json:"end_url"`
		ID        int    `json:"id"`
		StartURL  string `json:"start_url"`
	} `json:"like_icon"`
	RidStr string `json:"rid_str"`
}
type UserDynamic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			Basic   Basic
			IDStr   string `json:"id_str"`
			Modules struct {
				ModuleDynamic struct {
					Additional interface{} `json:"additional"`
					Desc       interface{} `json:"desc"`
					Major      struct {
						Archive struct {
							Aid   string `json:"aid"`
							Badge struct {
								BgColor string      `json:"bg_color"`
								Color   string      `json:"color"`
								IconURL interface{} `json:"icon_url"`
								Text    string      `json:"text"`
							} `json:"badge"`
							Bvid           string `json:"bvid"`
							Cover          string `json:"cover"`
							Desc           string `json:"desc"`
							DisablePreview int    `json:"disable_preview"`
							DurationText   string `json:"duration_text"`
							JumpURL        string `json:"jump_url"`
							Stat           struct {
								Danmaku string `json:"danmaku"`
								Play    string `json:"play"`
							} `json:"stat"`
							Title string `json:"title"`
							Type  int    `json:"type"`
						} `json:"archive"`
						Opus struct {
							FoldAction []string `json:"fold_action"`
							JumpURL    string   `json:"jump_url"`
							Pics       []struct {
								Height  int         `json:"height"`
								LiveURL interface{} `json:"live_url"`
								Size    float64     `json:"size"`
								URL     string      `json:"url"`
								Width   int         `json:"width"`
							} `json:"pics"`
							Summary struct {
								RichTextNodes []struct {
									OrigText string `json:"orig_text"`
									Text     string `json:"text"`
									Type     string `json:"type"`
								} `json:"rich_text_nodes"`
								Text string `json:"text"`
							} `json:"summary"`
							Title interface{} `json:"title"`
						} `json:"opus"`
						Type string `json:"type"`
					} `json:"major"`
					Topic interface{} `json:"topic"`
				} `json:"module_dynamic"`
				ModuleMore struct {
					ThreePointItems []struct {
						Label string `json:"label"`
						Type  string `json:"type"`
					} `json:"three_point_items"`
				} `json:"module_more"`
				ModuleStat struct {
					Comment struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
					} `json:"comment"`
					Forward struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
					} `json:"forward"`
					Like struct {
						Count     int  `json:"count"`
						Forbidden bool `json:"forbidden"`
						Status    bool `json:"status"`
					} `json:"like"`
				} `json:"module_stat"`
				ModuleTag struct {
					Text string `json:"text"`
				} `json:"module_tag"`
			} `json:"modules"`
			Type    string `json:"type"`
			Visible bool   `json:"visible"`
			Basic0  Basic
		} `json:"items"`
		Offset         string `json:"offset"`
		UpdateBaseline string `json:"update_baseline"`
		UpdateNum      int    `json:"update_num"`
	} `json:"data"`
}
type User struct {
	gorm.Model
	Name       string
	UserID     string
	Fans       int
	LastActive int
}

func UpdateCommon() {
	for i := range Followings {
		var id = Followings[i].UserID
		res, _ := client.R().Get("https://api.bilibili.com/x/relation/stat?vmid=" + id)
		var state = UserState{}
		sonic.Unmarshal(res.Body(), &state)
		var user = User{}
		user.Fans = state.Data.Following
	}
}

func RefreshCookie() {
	resp, err := client.R().Get("https://space.bilibili.com/1265680561/dynamic")
	if err != nil {
		panic(err)
	}
	var cookie = resp.Header().Get("Set-Cookie")
	Cookie = strings.Split(cookie, ";")[0]
	config.Cookie = Cookie
}

func PushDynamic() {

}

func UpdateSpecial() {
	var flag = false
	if len(RecordedDynamic) == 0 {
		flag = true
	}
	for i := range config.SpecialList {
		var id = config.SpecialList[i]
		resp, _ := client.R().SetHeader("Cookie", config.Cookie).SetHeader("Referer", "https://www.bilibili.com/").SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + strconv.Itoa(id) + "&timezone_offset=-480&features=itemOpusStyle")
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
					PushDynamic()
				}
			}
		}

		fmt.Println(result)
	}
}
func RefreshFollowings() {

	Followings = make([]User, 0)
	var page = 1
	Special = make([]User, 0)
	for true {
		resp, err := client.R().Get("https://line3-h5-mobile-api.biligame.com/game/center/h5/user/relationship/following_list?vmid=" + string(config.User) + "&ps=50&pn=" + strconv.Itoa(page))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(resp)
		var list = FansList{}
		sonic.Unmarshal(resp.Body(), &list)
		var users = list.Data.List
		for i := 0; i < len(users); i++ {
			var user = User{}
			user.Name = users[i].Uname
			user.UserID = users[i].Mid
			Followings = append(Followings, user)
			for j := 0; j < len(config.SpecialList); j++ {
				if strconv.Itoa(config.SpecialList[j]) == user.UserID {
					Special = append(Special, user)
				}
			}
		}
		if len(users) == 0 {
			break
		}
		page++
	}
}

var config = Config{}
var Followings = make([]User, 0)

func main() {
	content, err := os.ReadFile("config.json")

	if err != nil {
		os.Create("config.json")
		config.SpecialDelay = "2m"
		config.CommonDelay = "30m"
		config.User = "2"
		config.RefreshFollowingsDelay = "30m"
		config.SpecialList = []int{1265680561}
		RefreshCookie()
		content, _ = sonic.Marshal(&config)
		os.WriteFile("config.json", content, 666)

	}
	Cookie = config.Cookie
	err = sonic.Unmarshal(content, &config)
	c := cron.New()
	c.AddFunc("@every 2s", func() { fmt.Println("Every hour thirty") })
	UpdateSpecial()
	c.Start()
	select {}
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
