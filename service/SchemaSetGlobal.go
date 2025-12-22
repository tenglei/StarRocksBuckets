/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaAutoSetGlobal
 *@date    2025/3/18 16:14
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/conn"
	"setbuckets/tools"
	"setbuckets/util"
	"time"
)

func SetParGlobal(table, taskname string, bucket int64) {
	c := color.New()
	stime := time.Now()

	util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", "自动调整内表所有分桶数")
	tgr, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", err.Error())
		return
	}
	before := tools.GetTabletNum(tgr, table)
	bc := tools.CheckRow(tgr, table)

	defer func() {
		last := tools.GetTabletNum(tgr, table)
		lc := tools.CheckRow(tgr, table)

		var title string
		if before > last {
			title = fmt.Sprintf("%s(%s)倍", c.Add(color.FgHiWhite).Sprint("降低"), c.Add(color.FgHiGreen).Sprint(before/last))
		} else if last > before {
			title = fmt.Sprintf("%s(%s)倍", c.Add(color.FgHiYellow).Sprint("升高"), c.Add(color.FgHiRed).Sprint(last/before))
		}
		msg := fmt.Sprintf("count:(%d/%d) %s %s %s:(%s)/last:(%s) %s", bc, lc, c.Add(color.FgHiGreen).Sprint(time.Now().Sub(stime).String()), table, c.Add(color.FgHiWhite).Sprint("tablet before"), c.Add(color.FgHiYellow).Sprint(before), c.Add(color.FgHiGreen).Sprint(last), title)
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", msg)

		fix = util.Fix{
			App:     util.SrConfig.Host,
			Edtime:  int64(time.Now().Sub(stime).Seconds()),
			Count:   lc,
			Before:  int64(before),
			Last:    int64(last),
			Comment: msg,
		}
	}()

	splitkey := tools.GetSplitKey(tgr, table)
	if bc == 0 {
		bucket = 1
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", "分桶数重置:(1)")
	}
	util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, " ", fmt.Sprintf("表数据量:(%s)，分桶数:(%s)", c.Add(color.FgHiGreen).Sprint(bc), c.Add(color.FgHiWhite).Sprint(bucket)))
	t.Signal <- []string{table, fmt.Sprintf("%d", bucket), splitkey, taskname}

	//整体阻塞
	var i int
	for {
		c := color.New()
		if inProgress == 0 && i >= 5 {
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, "EXIT!")
			break
		}
		i++
		time.Sleep(time.Second)
		if i%20 == 0 && inProgress == 0 {
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), taskname, inProgress, c.Add(color.FgHiYellow).Sprint(" GLOBAL BLOCKING..."))
		}
	}
}
