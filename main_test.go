package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	bili "github.com/114514ns/BiliClient"
	"github.com/bytedance/sonic"
	"github.com/jinzhu/copier"
	pool2 "github.com/sourcegraph/conc/pool"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func loadDB() {

	content, _ := os.ReadFile("config.json")
	sonic.Unmarshal(content, &config)
	config.Port = 8081
	var dsl = "#user:#pass@tcp(#server)/#name?charset=utf8mb4&parseTime=True&loc=Local"
	dsl = strings.Replace(dsl, "#user", config.SQLUser, 1)
	dsl = strings.Replace(dsl, "#pass", config.SQLPass, 1)
	dsl = strings.Replace(dsl, "#server", config.SQLServer, 1)
	dsl = strings.Replace(dsl, "#name", config.SQLName, 1)

	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN: dsl, // DSN data source name
	}))
	if err != nil {

	}
	db.AutoMigrate(&Comment{})
	s, _ := db.DB()
	s.SetMaxOpenConns(200)
	s.SetMaxIdleConns(100)
	clickDb, _ = gorm.Open(clickhouse.Open(config.ClickServer))

}

func MockRefreshLivers() {
	type Response struct {
		List []FrontAreaLiver
	}
	get, _ := client.R().Get("https://live.ikun.dev/areaLivers")
	var dst Response
	sonic.Unmarshal(get.Body(), &dst)
	var from = dst.List
	cachedLivers = append(cachedLivers, from...)

}
func TestHttp(test *testing.T) {
	loadDB()
	go func() {
		for {
			//cachedToken = GetAlistToken()
			time.Sleep(120 * time.Minute)
		}
	}()

	//MockRefreshLivers()

	//RefreshFlow()
	setupHTTPClient()
	go func() {
		//RefreshLivers()
	}()
	go func() {
		//RefreshMessagePoints()
	}()
	//clickDb = clickDb.Debug()
	InitHTTP()

}

func TestGenMoneyRankList(test *testing.T) {
	loadDB()
	var dst []LiveAction
	var m = make(map[int64]float64)
	type Struct struct {
		UID       int64
		UName     string
		Amount    float64
		Liver     string
		MedalName string
		Level     int
		LiverID   int64
	}
	var array []Struct
	db.Raw("SELECT from_id,gift_price from live_actions WHERE action_type > 1").Scan(&dst)
	for i := range dst {
		_, ok := m[dst[i].FromId]
		if ok {
			m[dst[i].FromId] = m[dst[i].FromId] + dst[i].GiftPrice.Float64
		} else {
			m[dst[i].FromId] = dst[i].GiftPrice.Float64
		}
	}
	for i := range m {
		array = append(array, Struct{
			UID:    i,
			Amount: m[i],
		})
	}
	sort.Slice(array, func(i, j int) bool {
		return array[i].Amount > array[j].Amount
	})

	array = array[:10000]

	for i := range array {
		var obj DBGuard
		db.Raw("select * from fans_clubs where uid = ? order by level desc limit 1", array[i].UID).Scan(&obj)
		var amount = array[i].Amount
		copier.Copy(&array[i], &obj)
		array[i].Amount = amount
	}

	bytes, _ := sonic.Marshal(&array)
	os.WriteFile("rank.json", bytes, 666)

	time.Now()

}

func TestGetGuardList(t *testing.T) {
	loadConfig()
	setupHTTPClient()
	GetGuardList("22625027", "672342685")
}

func TestGetFansClub(t *testing.T) {
	loadConfig()
	setupHTTPClient()
	GetFansClub("672342685", nil)
}

func TestRefreshLiver(t *testing.T) {
	loadDB()
	setupHTTPClient()
	RefreshLiver(22886883)

}
func TestTraceArea(t *testing.T) {
	//loadDB()
	loadConfig()
	setupHTTPClient()
	db, _ = gorm.Open(sqlite.Open("database.db"))
	man = NewSlaverManager([]string{})
	TraceArea(9, true)
}

func TestAnalyzeWatcher(test *testing.T) {
	loadConfig()
	loadDB()
	RefreshWatchers()
}

func TestMerge(test *testing.T) {
	loadConfig()
	var dsl = "#user:#pass@tcp(#server)/#name?charset=utf8mb4&parseTime=True&loc=Local"
	dsl = strings.Replace(dsl, "#user", config.SQLUser, 1)
	dsl = strings.Replace(dsl, "#pass", config.SQLPass, 1)
	dsl = strings.Replace(dsl, "#server", config.SQLServer, 1)
	dsl = strings.Replace(dsl, "#name", config.SQLName, 1)

	mariadb, _ := gorm.Open(mysql.New(mysql.Config{
		DSN: dsl, // DSN data source name
	}), &gorm.Config{Logger: logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)})

	click, _ := gorm.Open(clickhouse.Open("tcp://127.0.0.1:19000/bili?&username=default&password=@&read_timeout=10s"))
	const pageSize = 1000
	var lastID = 0
	total := 0

	click.AutoMigrate(FansClub{})
	for {
		var rows []FansClub
		// 使用主键递增分页，避免 offset 性能问题
		if err := mariadb.
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(pageSize).
			Find(&rows).Error; err != nil {
			log.Fatalf("查询失败: %v", err)
		}

		if len(rows) == 0 {
			fmt.Println("迁移完成，总计迁移行数：", total)
			break
		}

		if err := click.Create(&rows).Error; err != nil {
			log.Fatalf("写入失败: %v", err)
		}

		lastID = int(rows[len(rows)-1].ID)
		total += len(rows)
		fmt.Printf("已迁移 %d 行，最后ID=%d\n", total, lastID)
	}
}

func TestRefreshLiveCount(test *testing.T) {
	loadConfig()
	loadDB()
	var dst []Live
	db.AutoMigrate(&Live{})
	db.Raw("select * from lives where lives.end_at != 0 and month(created_at) = 11 order by id desc").Scan(&dst)
	//dst = append(dst, Live{})
	//dst[0].ID = 209926
	for _, live := range dst {

		var sc = 0.0
		var guard = 0.0
		var total = 0.0
		//var diff = 0.0
		db.Raw("select COALESCE(sum(gift_price)) from live_actions where live = ? and action_type = 4", live.ID).Scan(&sc)
		db.Raw("select COALESCE(sum(gift_price)) from live_actions where live = ? and action_type = 3", live.ID).Scan(&guard)
		var boxes []LiveAction
		db.Raw("select extra,gift_price from live_actions where live = ? and action_type = 2 and extra like '%盲盒%'", live.ID).Scan(&boxes)
		db.Raw("select sum(gift_price) from live_actions where live = ?", live.ID).Scan(&total)
		var diff = 0.0
		for _, box := range boxes {
			var spent = float64(toInt(strings.Split(box.Extra, ",")[1]))
			var d = box.GiftPrice.Float64 - spent
			diff = diff + d
		}
		db.Exec("update lives set box_diff = ? where id = ?", diff, live.ID)
		db.Exec("update lives set super_chat_money = ? where id = ?", sc, live.ID)
		db.Exec("update lives set guard_money = ? where id = ?", guard, live.ID)
		db.Exec("update lives set money = ? where id = ?", total, live.ID)

	}
}

func TestFixLiver(test *testing.T) {
	loadDB()
	/*
		var ids []string
		db.Raw("SELECT medal_name FROM fans_clubs GROUP BY medal_name").Scan(&ids)
		var m = make(map[string]int64)

		for _, id := range ids {
			var id0 int64
			db.Raw("select liver_id from fans_clubs where medal_name = ? and liver_id != 0 limit 1", id).Scan(&id0)
			m[id] = id0
		}

		for s, i := range m {
			if s != "粉丝牌" {
				if i != 0 {
					db.Model(&FansClub{}).Where("medal_name = ?", s).Update("liver_id", i)
				}
			}
		}

	*/
	var ids []int64
	db.Raw("SELECT liver_id FROM fans_clubs GROUP BY liver_id").Scan(&ids)

	for _, id := range ids {

		var name = ""
		db.Raw("select u_name from area_livers where uid = ?", id).Scan(&name)

		if id == 0 || name == "" {
			continue
		}

		db.Model(&FansClub{}).Where("liver_id = ?", id).Update("liver", name)
	}

	print()

}

func TestUpdateCommon(test *testing.T) {
	loadConfig()
	loadDB()
	setupHTTPClient()
	RefreshFollowings()
	var m = make(map[*bili.BiliClient]string)
	for j := 0; j < 64; j++ {

		c := bili.NewAnonymousClient(bili.ClientOptions{
			HttpProxy:       "207.2.122.68:8443",
			ProxyUser:       "fg27msTTyo",
			ProxyPass:       "PZ8u9Pr2oz",
			RandomUserAgent: true,
			NoCookie:        true,
		})

		ip := check(c)
		var found = false
		for _, s := range m {
			if ip == s {
				found = true
				break
			}
		}
		if !found {
			fmt.Println(len(m))
			m[c] = ip
		} else {
			transport, _ := c.Resty.Transport()
			transport.CloseIdleConnections()
			j--
		}

	}

	var pool = pool2.New().WithMaxGoroutines(64)
	for i := range Followings {

		pool.Go(func() {
			var id = strconv.FormatInt((Followings[i].UserID), 10)
			var url = "https://api.bilibili.com/x/web-interface/card?mid=" + id
			res, err := RandomKey(m).Resty.R().Get(url)
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

			user.UserID, _ = strconv.ParseInt(id, 10, 64)

			user.Face = ""
			if user.Fans != 0 {
				db.Save(&user)
			} else {
				fmt.Println(res.String())
			}

			time.Sleep(1500 * time.Millisecond)
		})
	}
}
func check(client *bili.BiliClient) string {
	res, _ := client.Resty.R().SetHeader("Connection", "close").Get("https://api.bilibili.com/x/web-interface/zone")
	return res.String()
}
func RandomKey[K comparable, V any](m map[K]V) K {

	if len(m) == 0 {
		var zero K
		return zero
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(keys))

	return keys[randomIndex]
}

func TestDeleteDirty(test *testing.T) {

	var ids []int64
	loadConfig()
	loadDB()
	var remov []int64

	db.Raw("select uid from area_livers group by uid").Scan(&ids)

	for _, id := range ids {
		var dst User
		db.Raw("select fans from users where user_id=? order by id desc limit 1", id).Scan(&dst)

		if dst.Fans < 100 && dst.Fans != 0 {
			remov = append(remov, id)
		}
	}

	for _, i := range remov {
		tx := db.Raw("delete from area_livers where uid=?", i)
		if tx.Error != nil {
			fmt.Println(tx.Error)
		}
	}

	fmt.Printf("ids: %v\n", remov)

}

func TestExportBox(test *testing.T) {
	loadDB()
	var dst []interface{}
	db.Raw("SELECT extra,gift_amount,gift_price FROM live_actions where action_type = 2 and extra like '%盲盒%' ").Scan(&dst)
	marshal, _ := sonic.Marshal(dst)
	os.WriteFile("box.json", marshal, os.ModePerm)
}
