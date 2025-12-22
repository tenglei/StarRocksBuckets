/*
 *@author  chengkenli
 *@project exlSchema
 *@package exl
 *@file    exl_partition_sync
 *@date    2025/2/20 19:06
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/conn"
	"setbuckets/metaload"
	"setbuckets/tools"
	"setbuckets/util"
	"strings"
	"sync"
	"time"
)

var t = Job{
	Tasks:    make(chan []string),
	Ticker:   make(chan []string),
	Signal:   make(chan []string),
	Metas:    make(chan metaload.MetaInfo),
	MetaData: make(chan []metaload.MetaInfo),
	MetaDone: make(chan struct{}),
}

var inProgress int

type Job struct {
	Signal   chan []string
	Tasks    chan []string
	Ticker   chan []string
	Metas    chan metaload.MetaInfo
	MetaData chan []metaload.MetaInfo
	MetaDone chan struct{}
}

func (j *Job) SyncPartition() {
	db, _ := conn.StarRocks(util.SrCfg)
	// 创建一个WaitGroup来等待所有goroutine完成
	var wg sync.WaitGroup
	// 定义最大并发处理的channel数量
	maxInProgress := util.P.Thread
	// 主循环处理channel数据
	for {
		select {
		case stmt, ok := <-t.Signal:
			stime := time.Now()
			table := stmt[0]
			bucket := stmt[1]
			splitkey := stmt[2]
			taskname := stmt[3]

			// 如果当前正在处理的channel数量达到上限，则等待
			var i int
			for {
				c := color.New()
				i++
				//并发堵塞
				if inProgress >= maxInProgress {
					if i%20 == 0 {
						util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", fmt.Sprintf(" <%s> %s...", c.Add(color.FgHiYellow).Sprint(inProgress), c.Add(color.FgHiBlue).Sprint("CHAN BLOCKING")))
					}
					time.Sleep(time.Second)
					continue
				}
				break
			}
			// 增加正在处理的channel数量
			inProgress++
			// 启动一个goroutine来处理数据
			c := color.New()
			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != any(nil) {
						util.Loggrs.Panic(c.Add(color.FgHiRed).Sprint("TOP:RUN > "), taskname, " ", r)
					}
					// 减少正在处理的channel数量
					inProgress--
					wg.Done()
				}()

				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", inProgress, " SUBMIT CHAN")
				// 如果channel被关闭，退出循环
				if !ok {
					util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", "关闭channel")
					return
				}

				// 触发执行exec
				msg := fmt.Sprintf("ALTER TABLE %s DISTRIBUTED BY HASH%s BUCKETS %s", table, splitkey, bucket)
				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", fmt.Sprintf("%s %s %s", c.Add(color.FgHiYellow).Sprint("SUBMIT"), time.Now().Sub(stime).String(), c.Add(color.FgHiWhite).Sprint(msg)))

				for i := 0; i < 3; i++ {
					r := db.Exec(msg)
					if r.Error != nil {
						util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", r.Error.Error())
						if strings.Contains(r.Error.Error(), "connection refused") {
							db, _ = conn.StarRocks(util.SrCfg)
						}
						continue
					}
					break
				}

				schema := strings.Split(table, ".")
				var x int
				for {
					c := color.New()
					x++
					var m []map[string]interface{}
					r := db.Raw(fmt.Sprintf("show alter table optimize from %s where TableName='%s' order by CreateTime desc limit 3", schema[0], schema[1])).Scan(&m)
					if r.Error != nil {
						util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", r.Error.Error())
						if strings.Contains(r.Error.Error(), "connection refused") {
							db, _ = conn.StarRocks(util.SrCfg)
						}
						break
					}
					time.Sleep(time.Second * 3)

					var states []string
					task := m[0]
					for _, m3 := range m {
						state := m3["State"].(string)
						states = append(states, state)
					}
					if tools.StrInSlice("RUNNING", states) || tools.StrInSlice("PENDING", states) || tools.StrInSlice("WAITING_TXN", states) {
						msg := fmt.Sprintf("%-4s %s %s", fmt.Sprintf("%s%%", task["Progress"].(string)), c.Add(color.FgHiYellow).Sprint(task["State"].(string)), time.Now().Sub(stime).String())
						if x%10 == 0 {
							util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", msg)
						}
						continue
					}
					msg = fmt.Sprintf("%-4s %s %s", fmt.Sprintf("%s%%", task["Progress"].(string)), c.Add(color.FgHiGreen).Sprint(task["State"].(string)), time.Now().Sub(stime).String())
					util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", msg)
					time.Sleep(time.Second * 3)
					break
				}
				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", c.Add(color.FgHiGreen).Sprint(fmt.Sprintf("task is done! 耗时:[%s]", time.Now().Sub(stime).String())))
			}()
		case stmt, ok := <-t.Tasks:
			stime := time.Now()

			mtkeys := stmt[0]
			bucket := stmt[1]
			taskname := stmt[2]
			splitkey := stmt[3]

			inProgress++

			c := color.New()

			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", inProgress, " SUBMIT CHAN")
			// 如果channel被关闭，退出循环
			if !ok {
				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", "关闭channel")
				goto done
			}
			// 触发执行exec
			msg := fmt.Sprintf("ALTER TABLE %s PARTITIONS (%s) DISTRIBUTED BY HASH%s BUCKETS %s", util.P.Table, mtkeys, splitkey, bucket)
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", fmt.Sprintf("%s %s %s", c.Add(color.FgHiYellow).Sprint("SUBMIT"), time.Now().Sub(stime).String(), c.Add(color.FgHiWhite).Sprint(msg)))

			for i := 0; i < 3; i++ {
				r := db.Exec(msg)
				if r.Error != nil {
					util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", r.Error.Error())
					if strings.Contains(r.Error.Error(), "connection refused") {
						db, _ = conn.StarRocks(util.SrCfg)
					}
					continue
				}
				break
			}

			schema := strings.Split(util.P.Table, ".")
			var x int
			for {
				c := color.New()
				x++
				var m []map[string]interface{}
				r := db.Raw(fmt.Sprintf("show alter table optimize from %s where TableName='%s' order by CreateTime desc limit 3", schema[0], schema[1])).Scan(&m)
				if r.Error != nil {
					util.Loggrs.Error(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", r.Error.Error())
					if strings.Contains(r.Error.Error(), "connection refused") {
						db, _ = conn.StarRocks(util.SrCfg)
					}
					break
				}
				time.Sleep(time.Second * 3)

				var states []string
				task := m[0]
				for _, m3 := range m {
					state := m3["State"].(string)
					states = append(states, state)
				}
				if tools.StrInSlice("RUNNING", states) || tools.StrInSlice("PENDING", states) || tools.StrInSlice("WAITING_TXN", states) {
					msg := fmt.Sprintf("%-4s %s %s", fmt.Sprintf("%s%%", task["Progress"].(string)), c.Add(color.FgHiYellow).Sprint(task["State"].(string)), time.Now().Sub(stime).String())
					if x%10 == 0 {
						util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", msg)
					}
					continue
				}
				msg = fmt.Sprintf("%-4s %s %s", fmt.Sprintf("%s%%", task["Progress"].(string)), c.Add(color.FgHiGreen).Sprint(task["State"].(string)), time.Now().Sub(stime).String())
				util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", msg)
				time.Sleep(time.Second * 3)
				break
			}
			inProgress--
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), taskname, " ", c.Add(color.FgHiGreen).Sprint(fmt.Sprintf("task is done! 耗时:[%s]", time.Now().Sub(stime).String())))
		}
	}
done:

	c := color.New()
	util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:RUN > "), "Job done.")
}
