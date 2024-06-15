package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// 全局数据库配置变量
var DBConfig = DatabaseConfig{
	Host:     "localhost",
	Port:     3306,
	Username: "root",
	Password: "123mysql",
	Database: "stay33",
}

type Wechat struct {
	// 微信公众号
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
	Token     string `json:"token"`
	// 支付商
	MchId     string `yaml:"mch_id"`
	MchKey    string `yaml:"mch_key"`
	MchNumber string `yaml:"mch_number"`
	// 私钥路径
	PrivateKey string `yaml:"private_key"`
	Domain     string `yaml:"domain"`
	Port       string `yaml:"port"`
}

type Config struct {
	Wechat Wechat `yaml:"wechat"`
}

var (
	C = new(Config)
)

func createDBConnection() (*sql.DB, error) {
	const (
		username = "root"
		password = "123mysql"
		host     = "localhost"
		port     = "3306"
		dbname   = "stay33"
	)
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", username, password, host, port, dbname)
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Printf("无法初始化数据库连接：%v", err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		log.Printf("无法连接到数据库：%v", err)
		return nil, err
	}

	return db, nil
}

func Load(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Printf("配置文件打开失败：%v", err)
		return err
	}
	if err = yaml.Unmarshal(bytes, &C); err != nil {
		log.Printf("配置文件解析失败：%v", err)
		return err
	}

	db, err := createDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT AppID, AppSecret, MchId, MchKey, MchNumber FROM wpay LIMIT 1")
	if err != nil {
		log.Printf("数据库查询失败：%v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&C.Wechat.AppID, &C.Wechat.AppSecret, &C.Wechat.MchId, &C.Wechat.MchKey, &C.Wechat.MchNumber); err != nil {
			log.Printf("扫描数据库结果失败：%v", err)
			return err
		}
	}
	// fmt.Println(C.Wechat)
	return nil
}

// type Wechat struct {
// 	// 微信公众号
// 	AppID string `yaml:"app_id"`

// 	AppSecret string `yaml:"app_secret"`

// 	Token string `json:"token"`
// 	// 支付商
// 	MchId string `yaml:"mch_id"`

// 	MchKey string `yaml:"mch_key"`

// 	MchNumber string `yaml:"mch_number"`
// 	// 私钥路径
// 	PrivateKey string `yaml:"private_key"`
// 	Domain     string `yaml:"domain"`
// }

// type Config struct {
// 	Wechat Wechat `yaml:"wechat"`
// }

// var (
// 	C = new(Config)
// )

// func Load(path string) (err error) {
// 	bytes, err := os.ReadFile(path)
// 	if err != nil {
// 		fmt.Printf("配置文件打开失败：%v\n", err)
// 		return
// 	}
// 	err = yaml.Unmarshal(bytes, C)
// 	if err != nil {
// 		fmt.Printf("配置文件解析失败：%v\n", err)
// 		return
// 	}
// 	return nil
// }
