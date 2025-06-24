package main

import (
	"github.com/jinzhu/copier"
	"github.com/sourcegraph/conc/pool"
	"strconv"
	"strings"
	"sync"
	"time"
)

func reverse[T any](arr []T) []T {
	n := len(arr)
	result := make([]T, n)
	for i, v := range arr {
		result[n-1-i] = v
	}
	return result
}

func TotalGuards() [3]int {
	var result []map[string]interface{}
	db.Raw("select uid,guard,u_name FROM area_livers").Scan(&result)
	var id = make(map[int64]bool)
	array := [3]int{}

	var l = len(result)
	reverse(result)
	for i, _ := range result {
		var item = result[l-1-i]
		if id[item["uid"].(int64)] != true {
			split := strings.Split(item["guard"].(string), ",")
			for i2, s := range split {
				n, _ := strconv.ParseInt(s, 10, 64)
				array[i2] += int(n)
			}
			id[item["uid"].(int64)] = true

		}

	}
	return array
}
func TotalWatcher() int {
	var count = 0
	db.Raw("SELECT COUNT(DISTINCT uid) FROM fans_clubs").Scan(&count)
	return count
}

func TotalLiver() int {

	var count = 0
	db.Raw("SELECT COUNT(DISTINCT uid) FROM area_livers").Scan(&count)
	return count
}
func RefreshLivers() {
	var result []FrontAreaLiver
	var ids []int64
	var idMap = make(map[int64][]FrontAreaLiver)
	db.Raw(`SELECT uid FROM area_livers GROUP BY uid`).Find(&ids)
	var wg = pool.New().WithMaxGoroutines(6)
	var mutex sync.Mutex
	for _, id := range ids {
		id := id // 👈 创建局部副本，避免闭包捕获同一个变量
		wg.Go(func() {
			var dst []FrontAreaLiver

			db.Raw(`select uid,fans,updated_at,guard,u_name from area_livers where uid=?`, id).Find(&dst)
			mutex.Lock()
			if len(dst) == 0 {
				time.Sleep(time.Millisecond * 100)
			}
			idMap[id] = dst

			mutex.Unlock()
		})
	}
	wg.Wait()
	for _, livers := range idMap {
		result = append(result, livers[len(livers)-1])
	}

	for i, liver := range result {
		var dst User
		db.Raw("select * from users where user_id=? order by id desc limit 1", liver.UID).Scan(&dst)
		if dst.ID != 0 {
			result[i].Bio = dst.Bio
			result[i].Verify = dst.Verify
			var live AreaLive
			db.Raw("select time from area_lives where uid=? order by time desc limit 1", liver.UID).Scan(&live)
			result[i].LastActive = live.Time
			var dst1 User
			db.Raw("select * from users where user_id=? and created_at < ? order by id desc limit 1", liver.UID, dst.CreatedAt.Add(time.Hour*-24)).Scan(&dst1)
			if dst1.ID != 0 {
				result[i].DailyDiff = int((float64(dst.Fans - dst1.Fans)) / (float64(dst.CreatedAt.Unix()-dst1.CreatedAt.Unix()) / 86400))
			}

		}
	}
	copier.Copy(&cachedLivers, &result)
}
func MinuteMessageCount(minute int64) int64 {
	var count int64
	db.Model(&LiveAction{}).
		Where("created_at >= (NOW() + INTERVAL 8 HOUR) - INTERVAL ? MINUTE", minute).
		Count(&count)
	return count
}
func TotalMessage() int64 {
	var count int64
	db.Model(&LiveAction{}).Count(&count)
	return count
}

var cachedLivers = make([]FrontAreaLiver, 0)
