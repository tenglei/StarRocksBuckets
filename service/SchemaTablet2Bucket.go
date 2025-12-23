package service

import (
	"fmt"
	"github.com/fatih/color"
	"math"
	"regexp"
	"setbuckets/util"
	"strconv"
	"strings"
)

// TabletRepSize 分片容量计算公式
func TabletRepSize(DataSize string) {
	datasize := size(DataSize)
	c := color.New()
	s := datasize / (3 * size(fmt.Sprintf("%d mb", util.P.TabletSize)))
	
	// 将浮点数分桶数转换为整数（向上取整，确保足够容纳数据）
	bucketCount := int64(math.Ceil(s))
	
	fmt.Println(c.Add(color.FgHiYellow).Sprint("分桶数 = [分区总容量 / (3副本 * 单个tablet容量)]"))
	fmt.Println(fmt.Sprintf("%d BUCKETS = [%.1f / (3 * %.1f)]", bucketCount, datasize, size(fmt.Sprintf("%d mb", util.P.TabletSize))))
	fmt.Println(fmt.Sprintf("%d BUCKETS = [%s / (3 * %s)]", bucketCount, DataSize, fmt.Sprintf("%dMB", util.P.TabletSize)))
	fmt.Println(fmt.Sprintf("%d BUCKETS = 综上所述，如按【Tablet/(%dMB)】计算，建议分桶数为 %s", bucketCount, util.P.TabletSize, c.Add(color.FgHiGreen).Sprint(bucketCount)))
	fmt.Println()
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
