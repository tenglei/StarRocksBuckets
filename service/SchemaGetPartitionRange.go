/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    ScanSchemaGetPartition
 *@date    2024/8/2 14:16
 */

package service

import (
	"fmt"
	"regexp"
	"setbuckets/conn"
	"setbuckets/util"
	"strings"
)

// ScanSchemaGetPartitionRange 通过正则把每个分区的范围精准定位出来
func ScanSchemaGetPartitionRange(stable string) ([]map[string]string, error) {
	defer func() {
		if r := recover(); r != any(nil) {
			util.Loggrs.Panic("出现panic:", stable)
		}
	}()

	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		return nil, err
	}
	/*获取分区信息*/
	var m []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", stable)).Scan(&m)
	if r.Error != nil {
		return nil, err
	}

	var partion []map[string]string
	var parlist []string
	for _, m2 := range m {
		regex := regexp.MustCompile(`keys:\s*\[[\w|\d{4}(-?\d{2}){2}]+\]`)
		matches := regex.FindAllStringSubmatch(m2["Range"].(string), -1)
		regex2 := regexp.MustCompile(`keys:\s*\[(\d+|\d{4}(-?\d{2}){2})\]`)
		if matches == nil {
			continue
		}
		// FindStringSubmatch 用于找到第一个匹配的完整模式和捕获组
		sindex := regex2.FindStringSubmatch(matches[0][0])
		eindex := regex2.FindStringSubmatch(matches[1][0])
		parlist = append(parlist, fmt.Sprintf(`PARTITION p%s VALUES [("%s"), ("%s"))`, sindex[1], sindex[1], eindex[1]))
		partion = append(partion, map[string]string{fmt.Sprintf("p%s", strings.NewReplacer("-", "").Replace(sindex[1])): fmt.Sprintf(`PARTITION p%s VALUES [("%s"), ("%s"))`, strings.NewReplacer("-", "").Replace(sindex[1]), sindex[1], eindex[1])})
	}
	return partion, nil
}
