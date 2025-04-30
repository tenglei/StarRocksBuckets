/*
 *@author  chengkenli
 *@project exlSchema
 *@package exl
 *@file    ExlSchemaTicker
 *@date    2025/3/1 21:11
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/metaload"
	"setbuckets/tools"
	"setbuckets/util"
	"strings"
)

func (j *Job) ChanTicker() {
	c := color.New()
	var metastore []metaload.MetaInfo

	for {
		select {
		case meta, _ := <-t.Metas:
			metastore = append(metastore, meta)
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("MET:JOI > "), len(metastore))
		case <-t.MetaDone:
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("MET:SYN > "), len(metastore))
			t.MetaData <- metastore
		case stmt, _ := <-t.Ticker:
			mtkeys := stmt[0]
			bucket := stmt[1]
			taskname := stmt[2]
			splitkey := stmt[3]

			tickerSlices := tools.RmStrSlice(strings.Split(mtkeys, ","))

			c := color.New()
			var thread int
			if util.P.Thread == 3 {
				thread = 100
			} else {
				thread = util.P.Thread
			}

			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:TIK > "), taskname, " 分区合并数:", thread)
			cslice := tools.SplitSlice(tickerSlices, thread)
			for i, item := range cslice {
				c := color.New()
				msg := fmt.Sprintf("%s/%s/%s (%s|%s)",
					c.Add(color.FgHiWhite).Sprint(i),
					c.Add(color.FgHiCyan).Sprint(len(cslice)-i),
					c.Add(color.FgHiYellow).Sprint(len(cslice)),
					c.Add(color.FgHiWhite).Sprint(len(tickerSlices)-(thread*i)),
					c.Add(color.FgHiGreen).Sprint(len(tickerSlices)),
				)
				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:TIK > "), taskname, " ", msg, "分发执行~")
				t.Tasks <- []string{strings.Join(item, ","), bucket, taskname, splitkey}
			}
		}
	}
}
