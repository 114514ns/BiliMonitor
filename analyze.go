package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jinzhu/copier"
	pool2 "github.com/sourcegraph/conc/pool"

	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MessagePoint struct {
	Time  string
	Count int
}

var cachedMessagesPoint = make([][]MessagePoint, 3)
var cachedFans = make(map[int][][]MessagePoint)
var cachedWatcher []FansClub //搜索那边用的
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
	for i := range result {
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
func RefreshMessagePoints() {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		db.Raw(`
	SELECT
    	DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') AS time,
    	COUNT(*) AS count
	FROM live_actions
	WHERE created_at >= NOW() - INTERVAL 24-8 HOUR
	GROUP BY time
	ORDER BY time;
`).Scan(&cachedMessagesPoint[0])
		wg.Done()
	}()

	go func() {
		db.Raw(`
			SELECT
		    	DATE_FORMAT(created_at, '%Y-%m-%d 00:00:00') AS time,
		    	COUNT(*) AS count
			FROM live_actions
			WHERE created_at >= NOW() - INTERVAL 24*30-8 HOUR
			GROUP BY time
			ORDER BY time;
		`).Scan(&cachedMessagesPoint[0x01])
		wg.Done()
	}()

	go func() {
		db.Raw(`
			SELECT
		    	DATE_FORMAT(created_at, '%Y-%m-31 00:00:00') AS time,
		    	COUNT(*) AS count
			FROM live_actions
			WHERE created_at >= NOW() - INTERVAL 24*30*6-8 HOUR
			GROUP BY time
			ORDER BY time;
		`).Scan(&cachedMessagesPoint[2])
		wg.Done()
	}()

	wg.Wait()
}

func TotalLiver() int {

	var count = 0
	db.Raw("SELECT COUNT(DISTINCT uid) FROM area_livers").Scan(&count)
	return count
}

type FlowLiver struct {
	FrontAreaLiver
	Type string
}

var cachedFlowLiver []FlowLiver

func RefreshFlow() {
	var mutex sync.Mutex
	var pool = pool2.New().WithMaxGoroutines(32)
	for i := range cachedLivers {
		var item = cachedLivers[i]
		pool.Go(func() {
			if item.Fans > 300 {
				var currentMonth = time.Now().Month()
				var currentYear = time.Now().Year()

				if time.Now().Day() < 7 {
					currentMonth--
				}
				var dst []int //当月
				db.Raw("select id from area_lives where uid = ? and year(time) = ? and month(time) = ? limit 5", item.UID, currentYear, currentMonth).Scan(&dst)

				var dst2 []int //上月
				db.Raw("select id from area_lives where uid = ? and year(time) = ? and month(time) = ? limit 5", item.UID, currentYear, currentMonth-1).Scan(&dst2)

				if dst == nil {
					dst = append(dst, currentYear)
				}
				if (dst2) == nil {
					dst2 = append(dst2, currentYear)
				}

				if item.UID == 3690970457573801 {
					time.Now()
				}

				if len(dst) < 5 && len(dst2) == 5 {
					//Leave

					mutex.Lock()
					cachedFlowLiver = append(cachedFlowLiver, FlowLiver{
						FrontAreaLiver: item,
						Type:           "Leave",
					})
					mutex.Unlock()
				}
				if len(dst2) < 5 && len(dst) == 5 {
					mutex.Lock()
					cachedFlowLiver = append(cachedFlowLiver, FlowLiver{
						FrontAreaLiver: item,
						Type:           "Enter",
					})
					mutex.Unlock()
				}

			}
		})
	}
	pool.Wait()
}

var listRef = ""

func RefreshLivers() {

	var start = time.Now()
	log.Println("[RefreshLivers] Start")
	var result []FrontAreaLiver
	var ids []int64
	var idMap = make(map[int64][]FrontAreaLiver)
	db.Raw(`SELECT uid FROM area_livers GROUP BY uid`).Find(&ids)
	var wg = pool2.New().WithMaxGoroutines(6)
	var mutex sync.Mutex
	for _, id := range ids {
		id := id
		wg.Go(func() {
			var dst []FrontAreaLiver

			db.Raw(`select uid,fans,updated_at,guard,u_name,room from area_livers where uid=?`, id).Find(&dst)
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
		if len(livers) > 0 {
			if livers[len(livers)-1].Fans > 1000 {
				if !Has(config.BlackAreaLiver, livers[len(livers)-1].UID) {
					result = append(result, livers[len(livers)-1])
				}
			}
		}

	}

	type DiffStruct struct {
		UserID int64
		Diff   int
	}
	var order = []string{"DESC", "ASC"}
	var m = make(map[int64]DiffStruct)

	for _, s := range order {
		var dst []DiffStruct
		db.Raw(`
		WITH ranked AS (
		  SELECT user_id, name, fans, created_at,
		         ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at ASC)  AS rn_asc,
		         ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) AS rn_desc
		  FROM users
		  WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)   -- 最近1个月内的数据
		)
		SELECT e.user_id,
		       e.name,
		       e.fans AS start_fans,
		       e.created_at AS start_time,
		       l.fans AS end_fans,
		       l.created_at AS end_time,
		       l.fans - e.fans AS diff,
		       TIMESTAMPDIFF(DAY, e.created_at, l.created_at) AS days_span,
		       (l.fans - e.fans) / NULLIF(TIMESTAMPDIFF(DAY, e.created_at, l.created_at) / 30.4375, 0) AS monthly_growth
		FROM ranked e
		JOIN ranked l ON e.user_id = l.user_id
		WHERE e.rn_asc = 1 AND l.rn_desc = 1
		ORDER BY monthly_growth ` + s + " limit 1000;").Scan(&dst)
		for _, diffStruct := range dst {
			m[diffStruct.UserID] = diffStruct
		}
	}

	for i, liver := range result {

		liverMap[liver.Room] = liver
		var dst User
		db.Raw("select * from users where user_id=? order by id desc limit 1", liver.UID).Scan(&dst)
		if dst.ID != 0 && dst.Fans > 1000 {
			result[i].Bio = dst.Bio
			result[i].Verify = dst.Verify
			result[i].Fans = dst.Fans
			_, ok := m[liver.UID]
			if ok {
				if m[liver.UID].Diff != 0 {
					result[i].MonthlyDiff = m[liver.UID].Diff
				}
			}
			var live AreaLive
			db.Raw("select time from area_lives where uid=? order by time desc limit 1", liver.UID).Scan(&live)
			result[i].LastActive = live.Time
			var dst1 User
			db.Raw("select * from users where user_id=? and created_at < ? order by id desc limit 1", liver.UID, dst.CreatedAt.Add(time.Hour*-24)).Scan(&dst1)
			if dst1.ID != 0 {
				result[i].DailyDiff = int((float64(dst.Fans - dst1.Fans)) / (float64(dst.CreatedAt.Unix()-dst1.CreatedAt.Unix()) / 86400))
				if abs(result[i].DailyDiff) > 10000 {
					time.Now()
				}
			}

		}
	}
	log.Println("[RefreshLivers] Done " + time.Since(start).String())
	sort.Slice(result, func(i, j int) bool {
		return result[i].Fans > result[j].Fans
	})
	copier.Copy(&cachedLivers, &result)

	type Temp struct {
		List []FrontAreaLiver `json:"list"`
	}

	bytes, _ := json.Marshal(Temp{
		List: cachedLivers,
	})

	UploadBytes(bytes, "/Microsoft365/static/areaLivers.json")

	ListFile("/Microsoft365/static/")

	time.Sleep(5 * time.Second)

	//listRef = GetFile("/139/Msic/areaLivers.json")

}
func RefreshWatchers() {
	var month = 7
	var dst []AreaLiver
	db.Raw(`
SELECT t.uid, t.u_name, t.updated_at,t.guard,t.guard_list
FROM area_livers t
WHERE MONTH(t.updated_at) = ?
  AND t.updated_at = (
    SELECT MAX(updated_at)
    FROM area_livers
    WHERE uid = t.uid
      AND MONTH(updated_at) = ?
  )
  AND (
    SELECT COUNT(*)
    FROM area_livers
    WHERE uid = t.uid AND MONTH(updated_at) = 8
  ) >= 3;`, month, month).Scan(&dst)
	var l1 int64 = 0
	var l2 int64 = 0
	var l3 int64 = 0
	var m = make(map[int64][]DBGuard)
	sort.Slice(dst, func(i, j int) bool {
		var count1 int64 = 0
		for _, s := range strings.Split(dst[i].Guard, ",") {
			count1 += toInt64(s)
		}
		var count2 int64 = 0
		for _, s := range strings.Split(dst[j].Guard, ",") {
			count2 += toInt64(s)
		}
		return count2 < count1
	})
	for _, liver := range dst {
		var list []DBGuard
		json.Unmarshal([]byte(liver.GuardList), &list)
		for _, guard := range list {
			m[guard.UID] = append(m[guard.UID], guard)
		}
		if liver.UpdatedAt.Day() > 10 {
			for i, s := range strings.Split(liver.Guard, ",") {
				var num = toInt64(s)
				if i == 0 {
					l1 += num
				}
				if i == 1 {
					l2 += num
				}
				if i == 2 {
					l3 += num
				}
			}
		}

	}

	fmt.Println(l1, l2, month)
}

func RefreshWatcher() {
	db.Raw("select u_name,uid,medal_name,level from fans_clubs group by uid order by level desc ").Scan(&cachedWatcher)
	time.Now()
}
func MinuteMessageCount(minute int64) int64 {
	var count int64
	db.
		Raw("SELECT count(*) FROM live_actions WHERE created_at >= (NOW() + INTERVAL 8 HOUR) - INTERVAL  ? MINUTE ", minute).
		Scan(&count)
	return count
}
func TotalMessage() int64 {
	var count int64
	db.
		Raw("SELECT id FROM live_actions  ORDER BY id desc limit 1").
		Scan(&count)
	return count
}

var cachedLivers = make([]FrontAreaLiver, 0)
var liverMap = make(map[int]FrontAreaLiver)
