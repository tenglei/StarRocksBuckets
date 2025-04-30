/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaAutoSetPBuckets
 *@date    2025/3/12 14:50
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/xid"
	"setbuckets/conn"
	"setbuckets/tools"
	"setbuckets/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SetParBuckets() {
	c := color.New()

	stime := time.Now()
	defer func() {
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), time.Now().Sub(stime).String())
	}()
	if strings.Contains(util.P.Table, ",") {
		util.Loggrs.Info("该模式属于修改分区层级的分桶数，暂不支持多表提交")
		return
	}

	util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > ", "自动修改分区层级分桶数"))
	tgr, err := conn.StarRocks(util.P.App)
	if err != nil {
		util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), err.Error())
		return
	}
	before := tools.GetTabletNum(tgr, util.P.Table)
	bc := tools.CheckRow(tgr, util.P.Table)

	defer func() {
		last := tools.GetTabletNum(tgr, util.P.Table)
		lc := tools.CheckRow(tgr, util.P.Table)

		var title string
		if before > last {
			title = fmt.Sprintf("%s(%s)倍", c.Add(color.FgHiWhite).Sprint("降低"), c.Add(color.FgHiGreen).Sprint(before/last))
		} else if last > before {
			title = fmt.Sprintf("%s(%s)倍", c.Add(color.FgHiYellow).Sprint("升高"), c.Add(color.FgHiRed).Sprint(last/before))
		}
		msg := fmt.Sprintf("count:(%d/%d) %s %s %s:(%s)/last:(%s) %s", bc, lc, c.Add(color.FgHiGreen).Sprint(time.Now().Sub(stime).String()), util.P.Table, c.Add(color.FgHiWhite).Sprint("tablet before"), c.Add(color.FgHiYellow).Sprint(before), c.Add(color.FgHiGreen).Sprint(last), title)
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), msg)

		fix = util.Fix{
			App:     util.P.App,
			Edtime:  int64(time.Now().Sub(stime).Seconds()),
			Count:   lc,
			Before:  int64(before),
			Last:    int64(last),
			Comment: msg,
		}
	}()

	// 当没有指定分区时，自动获取
	if len(util.P.PartitionName) == 0 {

		var m2 []map[string]interface{}
		r := tgr.Raw("show partitions from " + util.P.Table).Scan(&m2)
		if r.Error != nil {
			util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), r.Error.Error())
			return
		}
		splitkey := tools.GetSplitKey(tgr, util.P.Table)

		x := 1
		var pk []string
		var done = make(chan struct{}, 100)
		var wg sync.WaitGroup
		for _, m := range m2 {
			wg.Add(1)
			m := m
			go func() {
				defer wg.Done()
				defer func() { <-done; x++ }()
				done <- struct{}{}

				bucket, _ := strconv.Atoi(m["Buckets"].(string))
				datasize := tools.Size(m["DataSize"].(string))
				dn := int64(datasize / 1024 / 1024 / 1024)
				// 数据驳回
				switch bucket {
				case 1, 2, 3:
					if strings.Contains(m["DataSize"].(string), "MB") || strings.Contains(m["DataSize"].(string), "KB") || strings.Contains(m["RowCount"].(string), "0") {
						return
					}
				}
				// 数据稽核
				var num int64
				if !strings.Contains(m["DataSize"].(string), "GB") {
					if strings.Contains(m["DataSize"].(string), "MB") {
						num = 3
					} else if strings.Contains(m["DataSize"].(string), "KB") {
						num = 2
					} else if m["RowCount"].(string) == "0" {
						num = 1
					}
				} else {
					num = dn + 5
				}
				pk = append(pk, fmt.Sprintf("%s^%d", m["PartitionName"].(string), num))
			}()
		}
		wg.Wait()

		if len(pk) == 0 {
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), c.Add(color.FgHiGreen).Sprint("NO MODIFICATIONS"))
			return
		}
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), fmt.Sprintf("正常:(%d)/异常:(%d)/总:(%d)", len(m2)-len(pk), len(pk), len(m2)))

		afterPartitions := dynamicPartition(tgr, util.P.Table)
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), "不处理未来分区:", strings.Join(afterPartitions, ","))

		mm := tools.SliceGroup(pk)
		for i, i2 := range mm {
			taskname := xid.New().String()
			c := color.New()
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", fmt.Sprintf("Jobid:%s/%s Count:%s", c.Add(color.FgHiYellow).Sprint(i), c.Add(color.FgHiYellow).Sprint(len(i2)), c.Add(color.FgHiBlack).Sprint(strings.Join(i2, ","))))
			var Stmt []string
			var bucket string
			for _, s := range i2 {
				if tools.StrInSlice(s, afterPartitions) {
					continue
				}
				bucket = strings.Split(s, "^")[1]
				Stmt = append(Stmt, strings.Split(s, "^")[0])
			}
			t.Ticker <- []string{strings.Join(Stmt, ","), bucket, taskname, splitkey}
		}
	} else {
		//如果指定了分区名称，直接传入对应分区进行重置分桶
		taskname := xid.New().String()
		splitkey := tools.GetSplitKey(tgr, util.P.Table)
		if util.P.Buckets <= 0 {
			util.Loggrs.Warn("分桶数不能为0!")
			return
		}
		t.Ticker <- []string{util.P.PartitionName, strconv.Itoa(int(util.P.Buckets)), taskname, splitkey}
	}

	//整体阻塞
	var i int
	for {
		c := color.New()
		if inProgress == 0 && i >= 5 {
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), "EXIT!")
			break
		}
		i++
		time.Sleep(time.Second)
		if i%20 == 0 && inProgress == 0 {
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), inProgress, c.Add(color.FgHiYellow).Sprint(" GLOBAL BLOCKING..."))
		}
	}
}
