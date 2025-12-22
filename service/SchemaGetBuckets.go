package service

import (
	"fmt"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"time"
)

func getbuckets(table string) bucketStr {
	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return bucketStr{}
	}
	/*每次使用完，主动关闭连接数*/
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			util.Loggrs.Error(err.Error())
			return
		}
		sqlDB.SetMaxOpenConns(30)                  //最大连接数
		sqlDB.SetMaxIdleConns(30)                  //最大空闲连接数
		sqlDB.SetConnMaxLifetime(30 * time.Second) //空闲连接最多存活时间
		sqlDB.Close()
	}()

	//新的锯割--遍历所有分区,拿到最大分区
	var n []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", table)).Scan(&n)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return bucketStr{}
	}
	if n == nil {
		return bucketStr{}
	}
	var Max []float64
	for _, m := range n {
		Max = append(Max, size(m["DataSize"].(string)))
	}
	maxDataSize := Max[0]
	for i := 0; i < len(Max); i++ {
		if Max[i] >= maxDataSize {
			maxDataSize = Max[i]
		}
	}
	//end

	var m map[string]interface{}
	row := db.Raw(fmt.Sprintf("show partitions from %s order by LastConsistencyCheckTime,DataSize desc limit 1", table)).Scan(&m)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		return bucketStr{}
	}
	var buckets float64
	var conservative, best int
	var normal bool
	Buckets := m["Buckets"].(string)
	parseInt, _ := strconv.ParseInt(Buckets, 10, 64)

	if maxDataSize <= 1073741824 && parseInt <= 3 {
		buckets = 3
	} else if maxDataSize < 1073741824 {
		conservative = 1
		best = 10
	} else {
		conservative = int(maxDataSize / 1073741824)
		best = conservative + 10
	}
	if conservative == 0 {
		conservative = 1
	}

	if buckets == 3 {
		normal = true
	} else {
		if int(parseInt) <= conservative+20 && int(parseInt) >= conservative {
			normal = true
		} else {
			normal = false
		}
	}
	var mms string
	if !normal {
		mms = "倾斜"
	} else {
		mms = "正常"
	}


	msgs := fmt.Sprintf("说明：[%s]，表中最大的分区容量为:%s，目前分桶是:%s，按照1GB=1BUCKETS原则，允许分桶范围:（%d~%d）BUCKETS，最佳分桶是:%d", mms, fmt.Sprintf("%0.4fGB", maxDataSize/1024/1024/1024), Buckets, conservative, conservative+20, best)

	b := bucketStr{
		App:          util.SrConfig.Host,
		Best:         best,
		Buckets:      Buckets,
		Client:       "",
		Conservative: conservative,
		Datasize:     fmt.Sprintf("%0.4fGB", maxDataSize/1024/1024/1024),
		Msg:          msgs,
		Normal:       normal,
		Table:        table,
	}
	fmt.Println(msgs)
	return b
}
