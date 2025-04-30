/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaDynamicPartition
 *@date    2025/4/21 13:14
 */

package service

import (
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"setbuckets/util"
	"time"
)

func dynamicPartition(db *gorm.DB, table string) []string {
	/*获取分区信息*/
	var m []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", table)).Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return nil
	}

	var after []string
	for _, m2 := range m {
		regex := regexp.MustCompile(`keys:\s*\[[\w|\d{4}(-?\d{2}){2}]+\]`)
		matches := regex.FindAllStringSubmatch(m2["Range"].(string), -1)
		regex2 := regexp.MustCompile(`keys:\s*\[(\d+|\d{4}(-?\d{2}){2})\]`)
		if matches == nil {
			continue
		}
		// FindStringSubmatch 用于找到第一个匹配的完整模式和捕获组
		sindex := regex2.FindStringSubmatch(matches[0][0])
		//eindex := regex2.FindStringSubmatch(matches[1][0])

		if isDateTodayOrAfter(sindex[1]) {
			after = append(after, m2["PartitionName"].(string))
		}
	}
	return after
}

// isDateTodayOrAfter checks if the given date is today or after today.
func isDateTodayOrAfter(date string) bool {
	// Parse the given date string into a time.Time object
	givenDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return false
	}
	// Get today's date with zeroed time
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Compare the given date with today's date
	return givenDate.After(todayDate) || givenDate.Equal(todayDate)
}
