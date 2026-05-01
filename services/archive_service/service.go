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
	"github.com/samber/lo"
	pool2 "github.com/sourcegraph/conc/pool"
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
			over = true
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

	if len(list) > 0 {
		clickDb.Create(list)
	}

}

func UpdateReaction(oid int64) {
	var items = RandomM(clients).GetReactionList(oid, 103)
	var list = struct {
		OID   int64
		Items []interface{}
	}{
		oid, nil,
	}
	for i := range items {
		list.Items = append(list.Items, struct {
			UID   int64
			UName string
		}{
			items[i].UID,
			items[i].UName,
		})
	}

	message := dynamic.NewMessage(protoMap["REACTION_LIST"])
	marshal, _ := json.Marshal(list)
	message.UnmarshalJSON(marshal)
	bytes, _ := message.Marshal()
	clickDb.Raw("delete from reactions where oid = ?", oid)
	clickDb.Table("reactions").Create(&struct {
		OID int64
		Pb  string
	}{
		oid, string(bytes),
	})

}

func UpdateFeedDetails(dyn bili.Dynamic) {

	//处理评论区，弹幕，视频流
	if dyn.BV != "" {
		if clickDb.Raw("select bv from videos where bv = ?", dyn.BV).Scan(new("")).RowsAffected == 0 {
			UpdateVideo(dyn.BV) //弹幕 视频流
		}

	}
	UpdateReplies(dyn.CommentID, dyn.CommentType, dyn.Comments) //评论区
	UpdateReaction(dyn.ID)

}

func UpdateClients() {
	var count = 0
	for {
		var c = bili.NewClient(config.Cookie, bili.ClientOptions{
			HttpProxy:       config.HttpProxy,
			ProxyUser:       config.ProxyUser,
			ProxyPass:       config.ProxyPass,
			PrintErrorStack: true,
			PrintErrorURL:   true,
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
	videos := RandomM(clients).GetVideo(bv)
	var dst Video
	clickDb.Raw("select cid from videos where aid = ? order by create_at desc limit 1", videos[0].Aid).Scan(&dst)
	if dst.Cid != videos[0].Cid {
		res, e := RandomM(clients).Resty.R().Get(config.StreamAgentEndPoint + fmt.Sprintf("?bv=%s&cid=%d", videos[0].BV, videos[0].Cid))
		if e != nil {
			log.Println(e)
		}
		fmt.Println(res.String())
		videos[0].RawResponse = ""
		clickDb.Create(&Video{
			videos[0], time.Now(), "",
		})
	}
	UpdateDanmaku(videos[0].Aid, videos[0].Cid, videos[0].Danmaku)
}

//seens字段存储爬取时间的数组
//每次爬的时候对比之前的，如果没有少，直接把新出现的加进去，不动seens字段

func UpdateDanmaku(aid int64, cid int, serverCount int) {

	var items = FetchDanmakus(aid, int64(cid))

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

	var last = ReadDanmakus(cid)

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
	clickDb.Create(&Danmaku{
		RawDanmaku:  string(bytes),
		Cid:         int64(cid),
		Count:       len(items),
		ServerCount: serverCount,
		Versions:    versions,
	})
}

func UpdateReplies(cid int64, typo int, serverCount int) {

	var last = ReadReplies(cid)

	var items = FetchReplies(cid, typo)

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID > items[j].ID
	})

	var s = ""
	clickDb.Raw("select (pb) from reply_lists where o_id = ?", cid).Scan(&s)

	var versions = ""
	clickDb.Raw("select (versions) from reply_lists where o_id = ?", cid).Scan(&versions)

	var now = time.Now()

	var versionsArray []string
	if versions != "" {
		json.Unmarshal([]byte(versions), &versionsArray)
	}
	versionsArray = append(versionsArray, now.Format(time.DateTime))

	versions0, _ := json.Marshal(versionsArray)

	versions = string(versions0)

	left, right := DifferenceBy(last, items, func(t Reply) int64 {
		return t.ID
	}) //left是被删掉的记录，right是新增的记录

	var leftM = make(map[int64]Reply)
	var rightM = make(map[int64]Reply)

	var itemsM = make(map[int64]Reply)

	for _, i := range left {
		var id = (i.ID)
		leftM[id] = i
	}

	for _, i := range right {
		var id = i.ID
		rightM[id] = i
	}

	for _, i := range items {
		var id = i.ID
		itemsM[id] = i
	}

	for i, _ := range last {
		_, ok := leftM[last[i].ID] //被删掉的弹幕，置alive为false
		if ok {
			last[i].Alive = false
		} else {
			if len(left) != 0 {
				//有被删掉的弹幕，要给每一条弹幕标上现在的时间
				//如果之前保存的所有弹幕都存在，就不需要标上时间了。

				m := last[i]
				m.Seens = append(m.Seens, now)
				last[i] = m
			}
		}
	}

	for _, v := range rightM {
		//然后加上新增的弹幕
		v.Alive = true
		v.Seens = append(v.Seens, now)
		last = append(last, v)

	}

	marshal, _ := json.Marshal(struct {
		Replies []Reply `json:"replies"`
	}{last})

	msg := dynamic.NewMessage(protoMap["REPLY_LIST"])
	err := msg.UnmarshalJSON(marshal)
	if err != nil {
		log.Println(err)
	}

	bytes, _ := msg.Marshal()
	clickDb.Raw("delete from reply_lists where o_id = ?", cid)
	var sum = 0
	for i := range items {
		sum += len(items[i].Reply.Reply)
	}
	clickDb.Create(ReplyList{
		Pb:          (bytes),
		Typo:        typo,
		OID:         cid,
		Count:       sum + len(items),
		ServerCount: serverCount,
		Versions:    versions,
		CreatedAt:   time.Now(),
	})
}

var cachedCollection = make(map[int64][]interface{})

func UpdateUserVideo(uid int64) {
	var dst []bili.Video
	var page = 1
	for {
		var l = len(dst)
		dst = append(dst, RandomM(clients).GetVideoByUser(uid, page, false)...)
		page++
		if l == len(dst) {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	var db []Video
	clickDb.Raw("select * from videos where aid in (?)", lo.Map(dst, func(item bili.Video, index int) int64 {
		return item.Aid
	}))
	_, right := DifferenceBy(dst, lo.Map(db, func(item Video, index int) bili.Video {
		return item.Video
	}), func(t bili.Video) int64 {
		return t.Aid
	})

	var pool = pool2.New().WithMaxGoroutines(8)
	for i := range right {
		pool.Go(func() {
			UpdateVideo(right[i].BV)
			UpdateDanmaku(right[i].Aid, right[i].Cid, right[i].Danmaku)
		})
	}
	pool.Wait()

}

func UpdateCollections(id, up int64, typo string) {
	var page = 1
	var tv []bili.Video

	var desc = ""
	var cName = ""
	for {
		var t, n, d = RandomM(clients).GetCollectionItems(id, typo, page, up)
		page++
		cName = n
		desc = d
		tv = append(tv, t...)
		if len(t) == 0 {
			break
		}
	}
	var db []Video

	clickDb.Raw("select * from videos where aid in (?)", lo.Map(tv, func(item bili.Video, index int) int64 {
		return item.Aid
	})).Scan(&db)

	left, _ := DifferenceBy(tv, lo.Map(db, func(item Video, index int) bili.Video {
		return db[index].Video
	}), func(t bili.Video) int64 {
		return t.Aid
	})

	var pool = pool2.New().WithMaxGoroutines(8)
	var dst = make([]bili.Video, len(left))
	for i := range left {
		pool.Go(func() {
			//UpdateVideo()
			dst[i] = left[i]
			cachedCollection[left[i].Aid] = append(cachedCollection[left[i].Aid], cName)
			cachedCollection[left[i].Aid] = append(cachedCollection[left[i].Aid], id)
			dst[i].CollectionID = id
			dst[i].CollectionName = cName
			UpdateVideo(dst[i].BV)
			time.Sleep(time.Millisecond * 100)
		})
	}
	pool.Wait()
	var dbs = ""
	clickDb.Raw("select items from collections where uid = ? and id = ?", up, id).Scan(&dbs)

	sort.Slice(dst, func(i, j int) bool {
		return dst[i].Aid > dst[j].Aid
	})

	marshal, _ := json.Marshal(lo.Map(dst, func(item bili.Video, index int) int64 {
		return item.Aid
	}))

	if dbs != string(marshal) {
		clickDb.Raw("delete from collections where uid = ? and id = ?", up, id)
		clickDb.Create(Collection{
			UID:       up,
			ID:        int(id),
			Name:      cName,
			Items:     string(marshal),
			Desc:      desc,
			CreatedAt: time.Now(),
		})
	}
	fmt.Println()

}
