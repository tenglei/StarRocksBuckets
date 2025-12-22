package main

import (
	"fmt"
	"github.com/fatih/color"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path/filepath"
	"setbuckets/service"
	"setbuckets/util"
	"time"
)

func main() {
	printStarRocks()
	util.Parm()

	// 交互式采集StarRocks连接配置
	cfg, err := util.CollectConnectionParams()
	if err != nil {
		c := color.New()
		fmt.Println(c.Add(color.FgHiRed).Sprint("错误: "), "采集连接参数失败: ", err)
		os.Exit(1)
	}

	// 将配置存储到全局变量
	util.SrCfg = cfg

	// 连接测试
	c := color.New()
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("正在连接StarRocks集群..."))
	if err := testConnection(cfg); err != nil {
		fmt.Println(c.Add(color.FgHiRed).Sprint("错误: "), "连接StarRocks失败: ", err)
		fmt.Println(c.Add(color.FgHiYellow).Sprint("请检查："))
		fmt.Println(c.Add(color.FgHiWhite).Sprint("  1. FE地址和端口是否正确"))
		fmt.Println(c.Add(color.FgHiWhite).Sprint("  2. 网络是否可达"))
		fmt.Println(c.Add(color.FgHiWhite).Sprint("  3. 用户名和密码是否正确"))
		fmt.Println(c.Add(color.FgHiWhite).Sprint("  4. 用户是否具有足够的权限"))
		os.Exit(1)
	}
	fmt.Println(c.Add(color.FgHiGreen).Sprint("✓ 连接成功！"))
	fmt.Println()

	service.Run()
}

// testConnection 测试连接配置是否可用
func testConnection(cfg util.SrAvgs) error {
	newLogger := logger.New(nil,
		logger.Config{
			SlowThreshold: time.Second * 1000,
		},
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	// 执行简单查询验证
	var version string
	err = db.Raw("SELECT VERSION()").Scan(&version).Error
	if err != nil {
		return err
	}

	// 关闭测试连接
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	return nil
}

func printStarRocks() {
	c := color.New()
	s := c.Add(color.FgHiGreen).Sprint(`
            ____  _             ____            _                      
           / ___|| |_ __ _ _ __|  _ \ ___   ___| | _____               
           \___ \| __/ __ | ___| |_) / _ \ / __| |/ / __|              
            ___) | || (_| | |  |  _ < (_) | (__|   <\__ \              
           |____/ \__\____|_|  |_| \_\___/ \___|_|\_\___/              `)
	v := fmt.Sprintf("%s\n",
		c.Add(color.FgHiMagenta).Sprint(filepath.Base(os.Args[0])),
	)
	fmt.Println(fmt.Sprintf("%s\n\n%s\n", s, v))
}
