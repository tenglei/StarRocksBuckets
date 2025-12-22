/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    ScanSchemaSortKey
 *@date    2024/7/30 11:10
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"regexp"
	"setbuckets/conn"
	"setbuckets/tools"
	"setbuckets/util"
	"sort"
	"strings"
	"sync"
)

// ScanSchemaSortKey 排序键展示
func ScanSchemaSortKey() {
	fmt.Println()
	fmt.Println("排序键：")

	db, err := conn.StarRocks(util.SrCfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//提取切割键
	var cqm []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show create table %s", util.P.Table)).Scan(&cqm)
	if r.Error != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	var createSQL string
	for _, m := range cqm {
		createSQL = fmt.Sprintf("%v", m["Create Table"])
	}
	var splikKey, sortKey []string
	for _, s2 := range strings.Split(createSQL, "\n") {
		if strings.Contains(s2, "DISTRIBUTED BY") || strings.Contains(s2, "KEY(`") {
			splikKey = append(splikKey, s2)
			re1 := regexp.MustCompile(`\((.*?)\)`)
			matches2 := re1.FindAllStringSubmatch(s2, -1)
			for _, i2 := range matches2 {
				s := strings.NewReplacer(" ", "", "`", "").Replace(i2[1])
				sortKey = append(sortKey, strings.Split(s, ",")...)
			}
		}
	}
	sortKey = tools.RmStrSlice(sortKey)
	//分析切割键
	var cm []map[string]interface{}
	r = db.Raw("desc " + util.P.Table).Scan(&cm)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return
	}
	var fied []map[string]int64
	var FieldInt, FieldDecimal, FieldDate, FieldString []string

	var wg sync.WaitGroup
	for _, m2 := range cm {
		wg.Add(1)
		m2 := m2
		go func() {
			defer wg.Done()

			fied = append(fied, map[string]int64{m2["Field"].(string): tools.ScanCountDistinct(m2["Field"].(string), util.P.Table)})
			if strings.Contains(strings.ToLower(m2["Type"].(string)), "int") {
				FieldInt = append(FieldInt, m2["Field"].(string))
			}
			if strings.Contains(strings.ToLower(m2["Type"].(string)), "decimal") {
				FieldDecimal = append(FieldDecimal, m2["Field"].(string))
			}
			if strings.Contains(strings.ToLower(m2["Type"].(string)), "date") {
				FieldDate = append(FieldDate, m2["Field"].(string))
			}
			if strings.Contains(strings.ToLower(m2["Type"].(string)), "varchar") {
				FieldString = append(FieldString, m2["Field"].(string))
			}
		}()
	}
	wg.Wait()

	//test
	var intSort, decimalSort, dateSort, stringSort []map[string]int64
	for _, s2 := range FieldInt {
		adss, _ := tools.FindKeyRank(fied, s2)
		intSort = append(intSort, map[string]int64{s2: fied[adss][s2]})
	}
	for _, s2 := range FieldDecimal {
		adss, _ := tools.FindKeyRank(fied, s2)
		decimalSort = append(decimalSort, map[string]int64{s2: fied[adss][s2]})
	}
	for _, s2 := range FieldDate {
		adss, _ := tools.FindKeyRank(fied, s2)
		dateSort = append(dateSort, map[string]int64{s2: fied[adss][s2]})
	}
	for _, s2 := range FieldString {
		adss, _ := tools.FindKeyRank(fied, s2)
		stringSort = append(stringSort, map[string]int64{s2: fied[adss][s2]})
	}
	// [int]定义一个比较函数，用于sort包的比较接口
	sort.Slice(intSort, func(i, j int) bool {
		// 计算每个map的int64值总和
		sumI := tools.SumMapValues(intSort[i])
		sumJ := tools.SumMapValues(intSort[j])
		// 按照总和进行降序排序
		return sumI > sumJ
	})
	// [decimal]定义一个比较函数，用于sort包的比较接口
	sort.Slice(decimalSort, func(i, j int) bool {
		// 计算每个map的int64值总和
		sumI := tools.SumMapValues(decimalSort[i])
		sumJ := tools.SumMapValues(decimalSort[j])
		// 按照总和进行降序排序
		return sumI > sumJ
	})
	// [date]定义一个比较函数，用于sort包的比较接口
	sort.Slice(dateSort, func(i, j int) bool {
		// 计算每个map的int64值总和
		sumI := tools.SumMapValues(dateSort[i])
		sumJ := tools.SumMapValues(dateSort[j])
		// 按照总和进行降序排序
		return sumI > sumJ
	})
	// [string]定义一个比较函数，用于sort包的比较接口
	sort.Slice(stringSort, func(i, j int) bool {
		// 计算每个map的int64值总和
		sumI := tools.SumMapValues(stringSort[i])
		sumJ := tools.SumMapValues(stringSort[j])
		// 按照总和进行降序排序
		return sumI > sumJ
	})

	var split, intarr, decarr, datearr, strarr []string
	c := color.New()
	//数值
	if len(intSort) >= 1 {
		for key := range intSort[0] {
			split = append(split, fmt.Sprintf("%s(int)", key))
		}
		for _, v := range intSort {
		labeli:
			for s1, i := range v {
				for _, s2 := range sortKey {
					if s1 == s2 {
						intarr = append(intarr, c.Add(color.FgHiYellow).Sprint(fmt.Sprintf("[%s]:(%d)", s1, i)))
						break labeli
					}
				}
				intarr = append(intarr, fmt.Sprintf("[%s]:(%d)", s1, i))
			}
		}
	}
	//小数点
	if len(decimalSort) >= 1 {
		for key := range decimalSort[0] {
			split = append(split, fmt.Sprintf("%s(decimal)", key))
		}
		for _, v := range decimalSort {
		labelds:
			for s1, i := range v {
				for _, s2 := range sortKey {
					if s1 == s2 {
						decarr = append(decarr, c.Add(color.FgHiYellow).Sprint(fmt.Sprintf("[%s]:(%d)", s1, i)))
						break labelds
					}
				}
				decarr = append(decarr, fmt.Sprintf("[%s]:(%d)", s1, i))
			}
		}
	}
	//日期
	if len(dateSort) >= 1 {
		for key := range dateSort[0] {
			split = append(split, fmt.Sprintf("%s(date)", key))
		}
		for _, v := range dateSort {
		labeld:
			for s1, i := range v {
				for _, s2 := range sortKey {
					if s1 == s2 {
						datearr = append(datearr, c.Add(color.FgHiYellow).Sprint(fmt.Sprintf("[%s]:(%d)", s1, i)))
						break labeld
					}
				}
				datearr = append(datearr, fmt.Sprintf("[%s]:(%d)", s1, i))
			}
		}
	}
	//字符串
	if len(stringSort) >= 1 {
		for key := range stringSort[0] {
			split = append(split, fmt.Sprintf("%s(string)", key))
		}
		for _, v := range stringSort {
		labels:
			for s1, i := range v {
				for _, s2 := range sortKey {
					if s1 == s2 {
						strarr = append(strarr, c.Add(color.FgHiYellow).Sprint(fmt.Sprintf("[%s]:(%d)", s1, i)))
						break labels
					}
				}
				strarr = append(strarr, fmt.Sprintf("[%s]:(%d)", s1, i))
			}
		}

		// 处理最佳排序键的max and min
		var splitKeys []string
		var wgs sync.WaitGroup
		for _, key := range split {
			wgs.Add(1)
			go func(key string) {
				defer wgs.Done()

				keyv := strings.Split(key,"(")[0]
				sql := fmt.Sprintf("select max(%s) as max,min(%s) as min from (select count(*),%s from (select DISTINCT %s from (select %s from %s limit 1000) a ) b group by %s) c",
					keyv, keyv, keyv, keyv, keyv,
					util.P.Table,
					keyv,
				)
				var m map[string]interface{}
				r := db.Raw(sql).Scan(&m)
				if r.Error != nil {
					fmt.Println(r.Error.Error())
				}
				splitKeys = append(splitKeys, fmt.Sprintf("%s(max:%v,min:%v)",c.Add(color.FgHiGreen).Sprint(key),c.Add(color.FgHiBlue).Sprint(m["max"]),c.Add(color.FgHiBlue).Sprint(m["min"])))
			}(key)
		}
		wgs.Wait()

		fmt.Println(fmt.Sprintf("%-5s: %s", "当前序列", c.Add(color.FgHiYellow).Sprint(strings.Join(splikKey, "、"))))
		fmt.Println(fmt.Sprintf("%-5s: %s", "当前排序", c.Add(color.FgHiYellow).Sprint(strings.Join(sortKey, "、"))))
		fmt.Println(fmt.Sprintf("%-5s: %s", "最佳排序", strings.Join(splitKeys, "、")))
		fmt.Println(fmt.Sprintf("%-7s: %s", "数值", strings.Join(intarr, "、")))
		fmt.Println(fmt.Sprintf("%-6s: %s", "小数位", strings.Join(decarr, "、")))
		fmt.Println(fmt.Sprintf("%-7s: %s", "日期", strings.Join(datearr, "、")))
		fmt.Println(fmt.Sprintf("%-6s: %s", "字符串", strings.Join(strarr, "、")))

	}
}
