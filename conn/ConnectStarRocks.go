package conn

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"setbuckets/util"
	"time"
)

// StarRocks 使用提供的连接配置建立StarRocks数据库连接
func StarRocks(cfg util.SrAvgs) (*gorm.DB, error) {
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
