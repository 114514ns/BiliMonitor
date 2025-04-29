package main

import (
	"github.com/jinzhu/copier"
	"sort"
	"strconv"
	"strings"
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
	cachedLivers = []FrontAreaLiver{}
	var result []AreaLiver
	db.Model(&AreaLiver{}).Omit("guard_list").Find(&result)
	var m = make(map[int64]AreaLiver)
	for _, v := range result {
		m[v.UID] = v
	}
	result = []AreaLiver{}
	for _, liver := range m {
		result = append(result, liver)
	}
	d1 := time.Now().AddDate(0, 0, -1) // 昨天
	d0 := time.Now().AddDate(0, 0, -2) // 前天
	temp := make([]FrontAreaLiver, 0)
	for _, liver := range result {
		var o0 = User{}
		var o1 = User{}
		db.Model(&User{}).
			Select("fans,created_at").
			Where("user_id = ? AND created_at >= ? AND created_at < ?", liver.UID, d0, d1).
			Order("created_at DESC").
			First(&o0)

		d2 := time.Now()
		db.Model(&User{}).
			Select("fans,created_at").
			Where("user_id = ? AND created_at >= ? AND created_at < ?", liver.UID, d1, d2).
			Order("created_at DESC").
			First(&o1)
		var f = FrontAreaLiver{AreaLiver: liver}
		secondsDiff := float64(o1.CreatedAt.Unix() - o0.CreatedAt.Unix())
		days := secondsDiff / 86400.0
		fansDiff := float64(o1.Fans - o0.Fans)
		f.DailyDiff = int(fansDiff / days)
		var lastLive = AreaLive{}
		db.Model(&AreaLive{}).Where("uid = ?", liver.UID).Find(&lastLive)
		f.LastActive = lastLive.Time
		temp = append(temp, f)
	}
	sort.Slice(temp, func(i, j int) bool {
		return temp[i].Fans > temp[j].Fans
	})
	copier.Copy(&cachedLivers, &temp)
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
