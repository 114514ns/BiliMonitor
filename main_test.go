package main

import (
	"fmt"
	"github.com/bytedance/sonic"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"testing"
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
	go func() {
		RefreshLivers()
	}()
	InitHTTP()

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

	click, _ := gorm.Open(clickhouse.Open("tcp://127.0.0.1:19000/bili?&username=default&password=Zyh060813@&read_timeout=10s"))
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
	var ids []int64
	db.Raw("select liver_id from fans_clubs group by liver_id").Scan(&ids)
	for _, id := range ids {
		var liverName = ""
		db.Raw("select u_name from area_lives where uid = ? ORDER by id limit 1", id).Scan(&liverName)
		db.Model(&FansClub{}).Where("liver_id = ?", id).UpdateColumn("liver", liverName)
	}

}
