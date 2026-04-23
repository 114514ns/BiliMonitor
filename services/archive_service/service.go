package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	bili "github.com/114514ns/BiliClient"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jinzhu/copier"
)

func UpdateSpace(uid int64) {
	profile, _ := RandomM(clients).GetUser(uid)
	clickDb.Save(profile)
}

func convertDynamic(src []bili.Dynamic) []Dynamic {
	var dst []Dynamic
	for _, dynamic := range src {
		var d = Dynamic{}
		copier.Copy(&d, &dynamic)
		d.Images = ""
		for _, image := range dynamic.Images {
			d.Images += image + ","
		}
		d.RawResponse = ""

		dst = append(dst, d)
	}

	return dst
}
func UpdateFeeds(uid int64) {
	var last Dynamic
	clickDb.Raw("select * from dynamics where uid = ? order by create_at desc limit 1", uid).Scan(&last)

	var off = ""
	var list []Dynamic
	var over = false

	for {
		if over {
			break
		}
		t, off0 := RandomM(clients).GetDynamicsByUser(uid, true, off)
		if off0 != "" && off != off0 {
			off = off0
		} else {
			break
		}

		for _, i := range convertDynamic(t) {
			if i.CreateAt.Sub(last.CreateAt).Minutes() > 0 {
				list = append(list, i)
			} else {
				over = true
				break
			}
		}

		time.Sleep(100 * time.Millisecond)

	}

	clickDb.Create(list)

}

func UpdateFeedDetails() {

}

func UpdateClients() {
	var count = 0
	for {
		var c = bili.NewClient(config.Cookie, bili.ClientOptions{
			HttpProxy:       config.HttpProxy,
			ProxyUser:       config.ProxyUser,
			ProxyPass:       config.ProxyPass,
			PrintErrorStack: true,
		})
		var addr = c.GetLocation()
		var found = false
		for s := range clients {
			if s == addr.Address {
				found = true
				break
			}
		}
		if !found {
			clients[addr.Address] = c
			count++
		}
		if count >= config.Thread {
			break
		}
	}
}

func UpdateCookie() {
	for _, s := range clients {
		s.Cookie = config.Cookie
	}
}

func UpdateVideo(bv string) {

}

func UploadStream() {
	//RandomM(clients).DownloadVideo()
}

//seens字段存储爬取时间的数组
//每次爬的时候对比之前的，如果没有少，直接把新出现的加进去，不动seens字段

func FetchDanmakus(cid int64) []interface{} {
	var items []interface{} //当前服务端返回的所有弹幕

	var off = 1

	for {
		res, _ := RandomM(clients).Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/v2/dm/web/seg.so?oid=%d&type=1&segment_index=%d", cid, off))
		msg := dynamic.NewMessage(protoMap["DANMAKU_LIST"])
		err := msg.Unmarshal(res.Body())
		if err != nil {
			log.Println(err)
		}
		marshalJSON, _ := msg.MarshalJSON()
		var obj map[string]interface{}
		json.Unmarshal(marshalJSON, &obj)
		var arr = getArray(obj, "elems")
		if len(arr) == 0 {
			break
		} else {
			items = append(items, arr...)
		}
		off++
		time.Sleep(200 * time.Millisecond)
	}

	return items
}

func FetchReplies(cid int64, typo int) []interface{} {
	var dst0 = []bili.Reply{}
	var off = ""
	var count = 0
	for {
		var client = RandomM(clients)
		var array, o = client.GetCommentRPC(cid, off, typo, bili.REPLY_SORT_TIME)
		off = o
		time.Sleep(500 * time.Millisecond)
		for _, comment := range array {
			var tmp = bili.Reply{}
			copier.Copy(&tmp, comment)
			dst0 = append(dst0, tmp)
		}

		count = count + len(array)
		fmt.Println(count)
		if len(array) == 0 || off == "" {
			break
		}
	}

	var dst []interface{}

	for i := range dst0 {
		dst = append(dst, dst[i].(interface{}))
	}

	return dst
}

func UpdateDanmaku(cid int64, serverCount int) {

	var items = FetchDanmakus(cid)

	sort.Slice(items, func(i, j int) bool {
		return getInt(items[i], "id") > getInt(items[j], "id")
	})

	var s = ""
	clickDb.Raw("select (raw_danmaku) from danmakus where cid = ?", cid).Scan(&s)

	var versions = ""
	clickDb.Raw("select (versions) from danmakus where cid = ?", cid).Scan(&versions)

	var now = time.Now()

	var versionsArray []string
	if versions != "" {
		json.Unmarshal([]byte(versions), &versionsArray)
	}
	versionsArray = append(versionsArray, now.Format(time.DateTime))

	versions0, _ := json.Marshal(versionsArray)

	versions = string(versions0)

	var last []interface{} //数据库中的弹幕

	if s != "" {
		json.Unmarshal([]byte(s), &last)
	}

	left, right := DifferenceBy(last, items, func(t interface{}) int64 {
		return getInt64(t, "id")
	}) //left是被删掉的记录，right是新增的记录

	var leftM = make(map[int64]interface{})
	var rightM = make(map[int64]interface{})

	var itemsM = make(map[int64]interface{})

	for _, i := range left {
		var id = toInt64(getString(i, "id"))
		leftM[id] = i
	}

	for _, i := range right {
		var id = toInt64(getString(i, "id"))
		rightM[id] = i
	}

	for _, i := range items {
		var id = toInt64(getString(i, "id"))
		itemsM[id] = i
	}

	for i, _ := range last {
		_, ok := leftM[getInt64(last[i], "id")] //被删掉的弹幕，置alive为false
		if ok {
			last[i].(map[string]interface{})["alive"] = false
		} else {
			if len(left) != 0 {
				//有被删掉的弹幕，要给每一条弹幕标上现在的时间
				//如果之前保存的所有弹幕都存在，就不需要标上时间了。

				m := last[i].(map[string]interface{})
				m["seens"] = append(m["seens"].([]string), now.Format(time.DateTime))

			}
		}
	}

	for _, v := range rightM {
		//然后加上新增的弹幕
		m := v.(map[string]interface{})
		m["alive"] = true
		m["seens"] = append([]string{}, now.Format(time.DateTime))
		last = append(last, m)

	}

	marshal, _ := json.Marshal(struct {
		Elems []interface{} `json:"elems"`
	}{last})

	msg := dynamic.NewMessage(protoMap["DANMAKU_LIST"])
	err := msg.UnmarshalJSON(marshal)
	if err != nil {
		log.Println(err)
	}

	bytes, _ := msg.Marshal()
	clickDb.Create(Danmaku{
		RawDanmaku:  string(bytes),
		Cid:         cid,
		Count:       len(items),
		ServerCount: serverCount,
		Versions:    versions,
	})
}

func UpdateReplies(cid int64, typo int, serverCount int) {

	var items = FetchReplies(cid, typo)

	sort.Slice(items, func(i, j int) bool {
		return getInt(items[i], "id") > getInt(items[j], "id")
	})

	var s = ""
	clickDb.Raw("select (pb) from reply_lists where oid = ?", cid).Scan(&s)

	var versions = ""
	clickDb.Raw("select (versions) from reply_lists where cid = ?", cid).Scan(&versions)

	var now = time.Now()

	var versionsArray []string
	if versions != "" {
		json.Unmarshal([]byte(versions), &versionsArray)
	}
	versionsArray = append(versionsArray, now.Format(time.DateTime))

	versions0, _ := json.Marshal(versionsArray)

	versions = string(versions0)

	var last []interface{} //数据库中的弹幕

	if s != "" {
		json.Unmarshal([]byte(s), &last)
	}

	left, right := DifferenceBy(last, items, func(t interface{}) int64 {
		return getInt64(t, "id")
	}) //left是被删掉的记录，right是新增的记录

	var leftM = make(map[int64]interface{})
	var rightM = make(map[int64]interface{})

	var itemsM = make(map[int64]interface{})

	for _, i := range left {
		var id = toInt64(getString(i, "id"))
		leftM[id] = i
	}

	for _, i := range right {
		var id = toInt64(getString(i, "id"))
		rightM[id] = i
	}

	for _, i := range items {
		var id = toInt64(getString(i, "id"))
		itemsM[id] = i
	}

	for i, _ := range last {
		_, ok := leftM[getInt64(last[i], "id")] //被删掉的弹幕，置alive为false
		if ok {
			last[i].(map[string]interface{})["alive"] = false
		} else {
			if len(left) != 0 {
				//有被删掉的弹幕，要给每一条弹幕标上现在的时间
				//如果之前保存的所有弹幕都存在，就不需要标上时间了。

				m := last[i].(map[string]interface{})
				m["seens"] = append(m["seens"].([]string), now.Format(time.DateTime))

			}
		}
	}

	for _, v := range rightM {
		//然后加上新增的弹幕
		m := v.(map[string]interface{})
		m["alive"] = true
		m["seens"] = append([]string{}, now.Format(time.DateTime))
		last = append(last, m)

	}

	marshal, _ := json.Marshal(struct {
		Elems []interface{} `json:"elems"`
	}{last})

	msg := dynamic.NewMessage(protoMap["DANMAKU_LIST"])
	err := msg.UnmarshalJSON(marshal)
	if err != nil {
		log.Println(err)
	}

	bytes, _ := msg.Marshal()
	clickDb.Create(Danmaku{
		RawDanmaku:  string(bytes),
		Cid:         cid,
		Count:       len(items),
		ServerCount: serverCount,
		Versions:    versions,
	})
}
