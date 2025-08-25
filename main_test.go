package main

import (
	"fmt"
	"github.com/114514ns/BiliClient"
	"github.com/bytedance/sonic"
	pool2 "github.com/sourcegraph/conc/pool"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
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

	db, _ = gorm.Open(mysql.New(mysql.Config{
		DSN: dsl, // DSN data source name
	}))

}
func TestHttp(test *testing.T) {

	loadDB()
	setupHTTPClient()
	go func() {
		RefreshLivers()
	}()
	go func() {
		//RefreshMessagePoints()
	}()
	InitHTTP()

}

func TestTraceArea(t *testing.T) {
	//loadDB()
	loadConfig()
	TraceArea(9)
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

		// 写入 db2
		if err := click.Create(&rows).Error; err != nil {
			log.Fatalf("写入失败: %v", err)
		}

		lastID = int(rows[len(rows)-1].ID)
		total += len(rows)
		fmt.Printf("已迁移 %d 行，最后ID=%d\n", total, lastID)
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
