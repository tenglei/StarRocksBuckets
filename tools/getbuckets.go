package tools

import (
	"fmt"
	"regexp"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
)

type BucketInfo struct {
	App string
	Best int
	Buckets string
	Conservative int
	DataSize string
	Msg string
	Normal bool
	Table string
}

func GetBuckets(stable string) *BucketInfo{
	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return &BucketInfo{}
	}
	//新的锯割--遍历所有分区,拿到最大分区
	var n []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", stable)).Scan(&n)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return &BucketInfo{}
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
	row := db.Raw(fmt.Sprintf("show partitions from %s order by LastConsistencyCheckTime,DataSize desc limit 1", stable)).Scan(&m)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		return &BucketInfo{}
	}
	var buckets float64
	var conservative, best int
	var normal bool
	Buckets := m["Buckets"].(string)
	parseInt, _ := strconv.ParseInt(Buckets, 10, 64)

	if maxDataSize <= 1073741824 && parseInt <= 3 {
		buckets = 3
		best = 3
	} else if maxDataSize < 1073741824 {
		conservative = 1
		best = 3
	} else {
		conservative = int(maxDataSize / 1073741824)
		best = conservative + 3
	}

	if buckets == 3 {
		normal = true
	} else {
		if int(parseInt) <= best && int(parseInt) >= conservative {
			normal = true
		} else {
			normal = false
		}
	}
	return &BucketInfo{
		Best:         best,
		Buckets:      Buckets,
		Conservative: conservative,
		DataSize:     fmt.Sprintf("%0.4fGB", maxDataSize/1024/1024/1024),
		Msg:          "Success",
		Normal:       normal,
		Table:        stable,
	}
}

func size(s string) float64 {
	s = strings.ToLower(s)
	//正则
	re := regexp.MustCompile(`\d+\.?\d*`)

	float, _ := strconv.ParseFloat(re.FindString(s), 64)

	if strings.Contains(s, "kb") {
		return float * 1024
	}
	if strings.Contains(s, "mb") {
		return float * 1024 * 1024
	}
	if strings.Contains(s, "gb") {
		return float * 1024 * 1024 * 1024
	}
	if strings.Contains(s, "tb") {
		return float * 1024 * 1024 * 1024 * 1024
	}
	return 0
}
