package conn

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"setbuckets/util"
	"time"
)

func StarRocks(cfg util.SrAvgs) (*gorm.DB, error) {
	// 验证连接参数
	if err := util.ValidateConnectionParams(cfg); err != nil {
		return nil, fmt.Errorf("连接参数验证失败: %v", err)
	}

	newLogger := logger.New(nil,
		logger.Config{
			SlowThreshold: time.Second * 1000, // 控制慢SQL阈值
		},
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println(fmt.Sprintf("%s", err))
	}
	return db, err
}

func StarRocksApp(cfg util.SrAvgs, host string) (*gorm.DB, error) {
	if len(host) == 0 {
		return nil, errors.New("登录信息有误，或为空。")
	}
	
	// 验证连接参数
	if err := util.ValidateConnectionParams(cfg); err != nil {
		return nil, fmt.Errorf("连接参数验证失败: %v", err)
	}

	newLogger := logger.New(nil,
		logger.Config{
			SlowThreshold: time.Second * 1000, // 控制慢SQL阈值
		},
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Pass, host, cfg.Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名

		},
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println(err.Error())
		return db, err
	}
	return db, err
}
