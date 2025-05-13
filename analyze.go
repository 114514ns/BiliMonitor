package main

import (
	"github.com/jinzhu/copier"
	"log"
	"sort"
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
	var start = time.Now().Unix()
	var stage1 int64 = 0
	var stage2 int64 = 0
	db.Raw(`
    WITH LastVerify AS (
        SELECT user_id, verify, bio,
               ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY id DESC) AS rn
        FROM users
    )
    SELECT MAX(A.u_name) AS u_name, 
           A.uid,
           MAX(A.room) AS room,
           MAX(A.area) AS area,
           MAX(A.fans) AS fans,
           MAX(A.guard) AS guard,
           MAX(B.verify) AS verify,
           MAX(B.bio) AS bio
    FROM area_livers A
    INNER JOIN LastVerify B ON A.uid = B.user_id
    WHERE B.rn = 1
    GROUP BY A.uid
    ORDER BY A.fans DESC
`).Scan(&result)
	var m = make(map[int64]FrontAreaLiver)
	for _, v := range result {
		m[v.UID] = v
	}
	result = []FrontAreaLiver{}
	for _, liver := range m {
		result = append(result, liver)
	}
	stage1 = time.Now().Unix() - start
	temp := make([]FrontAreaLiver, 0)
	var wg sync.WaitGroup
	ch := make(chan struct{}, 4)

	for i, liver := range result {
		ch <- struct{}{}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var o0 = User{}
			var o1 = User{}
			var arr []User

			db.Raw(`
SELECT * FROM (
    SELECT *
    FROM users
    WHERE user_id = ?
    ORDER BY created_at DESC
    LIMIT 1
) AS LatestRecord

UNION ALL

SELECT * FROM (
    SELECT *
    FROM users
    WHERE user_id = ?
      AND created_at <= (
          SELECT MAX(created_at) FROM users WHERE user_id = ?
      ) - INTERVAL 1 DAY
    ORDER BY created_at DESC
    LIMIT 1
) AS OldRecord;
`, liver.UID, liver.UID, liver.UID).Find(&arr)

			o0 = arr[0]
			if len(arr) == 2 {
				o0 = arr[1]
			}
			o1 = arr[0]
			var f = liver
			secondsDiff := float64(o1.CreatedAt.Unix() - o0.CreatedAt.Unix())
			days := secondsDiff / 86400.0
			fansDiff := float64(o1.Fans - o0.Fans)
			f.DailyDiff = int(fansDiff / days)
			var lastLive = AreaLive{}
			db.Model(&AreaLive{}).Where("uid = ?", liver.UID).Order("id desc").Find(&lastLive)
			f.LastActive = lastLive.Time
			temp = append(temp, f)
			<-ch
		}(i)

	}
	stage2 = time.Now().Unix() - stage1
	wg.Wait()
	sort.Slice(temp, func(i, j int) bool {
		return temp[i].Fans > temp[j].Fans
	})
	log.Printf(strconv.FormatInt(stage2, 10))
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
