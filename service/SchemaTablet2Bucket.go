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
	fmt.Println(c.Add(color.FgHiYellow).Sprint("分桶数 = [分区总容量 / (3副本 * 单个tablet容量)]"))
	fmt.Println(fmt.Sprintf("%.1f BUCKETS = [%.1f / (3 * %.1f)]", s, datasize, size(fmt.Sprintf("%d mb", util.P.TabletSize))))
	fmt.Println(fmt.Sprintf("%.1f BUCKETS = [%s / (3 * %s)]", s, DataSize, fmt.Sprintf("%dMB", util.P.TabletSize)))
	fmt.Println(fmt.Sprintf("%.1f BUCKETS = 综上所述，如按【TabletId/(%dMB)】", math.Round(s), util.P.TabletSize))
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
