/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaSetPartitions
 *@date    2025/4/16 16:47
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/xid"
	"setbuckets/conn"
	"setbuckets/metaload"
	"setbuckets/tools"
	"setbuckets/util"
	"strconv"
	"strings"
	"time"
)

func SetPartitions(table string) {
	if !util.P.ClearPartition {
		return
	}

	stime := time.Now()
	Distribution(table)

	dbname := strings.Split(table, ".")[0]
	tbname := strings.Split(table, ".")[1]

	type Version struct {
		Version string `bson:"version"`
	}
	if len(util.P.Action) == 0 {
		util.P.Action = "create,insert,alter,drop"
	}

	var CreateSQL string
	db, err := conn.StarRocks(util.SrCfg)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}

	var cmm []map[string]interface{}
	err = db.Raw(fmt.Sprintf("show create table %s", table)).Scan(&cmm).Error
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	for _, m := range cmm {
		CreateSQL = fmt.Sprintf("%v", m["Create Table"])
	}
	stmt_before := CreateSQL

	cs := color.New()
	util.Loggrs.Info("原始SQL:\n" + cs.Add(color.FgHiBlue).Sprint(CreateSQL))

	NewSQL, err := ScanSchemaPartition(table)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	util.Loggrs.Info("去除空分区后SQL:\n" + cs.Add(color.FgHiYellow).Sprint(NewSQL))

	//替换表名
	oldTable := strings.Split(table, ".")[1]
	newTable := fmt.Sprintf("buckets_%d_%s", time.Now().UnixMilli(), strings.Split(table, ".")[1])
	if len(newTable) >= 40 {
		newTable = newTable[:30]
		util.Loggrs.Info(fmt.Sprintf("重构后表名过长，进行裁剪：原表名：%s，新表名：%s", strings.Split(table, ".")[1], newTable))
	}

	NewSQL = strings.ReplaceAll(strings.ReplaceAll(NewSQL, oldTable, newTable), "CREATE TABLE", fmt.Sprintf("use %s;create table if not exists", dbname))
	stmt_after := NewSQL
	/*new*/
	cs = color.New()
	util.Loggrs.Info("新生SQL:\n" + cs.Add(color.FgHiGreen).Sprint(NewSQL))
	util.Loggrs.Info("新生SQL执行中, 此过程会根据创建分区的多少, 耗费一定的时间...")
	err = db.Exec(NewSQL).Error
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}

	//-------------------------------
	//-------------------------------
	//stream load
	//这里代表着将行为进行入表
	var state string
	var operational []string
	defer func() {
		edtime := time.Now().Sub(stime).String()
		util.Loggrs.Info(fmt.Sprintf("[%s]耗时: %s", table, edtime))
		Distribution(table)

		if state != "SUCCESSED" {
			state = "CANCELLED"
		}
		operational = append(operational, util.P.Action+">")
		if util.P.Auto || util.P.Buckets > 0 {
			operational = append(operational, "bucket")
		}
		if util.P.ClearPartition {
			operational = append(operational, "partition")
		}
		if len(util.P.SortKey) != 0 {
			operational = append(operational, "sortkey")
		}
		if len(util.P.SplitKey) != 0 {
			operational = append(operational, "splitkey")
		}
		t.Metas <- metaload.MetaInfo{
			Ts:          time.Now().Format("2006-01-02"),
			App:         util.P.App,
			Database:    dbname,
			Table:       tbname,
			StmtBefore:  stmt_before,
			StmtAfter:   stmt_after,
			Operational: strings.Join(operational, ","),
			Edtime:      time.Now().Sub(stime).String(),
			State:       state,
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
			Comment:     "",
		}
	}()
	//end
	//-------------------------------
	//-------------------------------

	/*死循环*/
	for {
		r := db.Exec(fmt.Sprintf("desc %s", newTable))
		if r.Error != nil {
			util.Loggrs.Warn("waiting..." + r.Error.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}

	util.Loggrs.Info(fmt.Sprintf("建表完成: %s.%s", strings.Split(table, ".")[0], newTable))
	if !strings.Contains(util.P.Action, "insert") {
		return
	}

	//获取数据总量
	var c Ct
	sql := fmt.Sprintf("select count(*) as count from %s", table)
	util.Loggrs.Info(sql)
	row := db.Raw(sql).Scan(&c)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		db.Exec(fmt.Sprintf("drop table %s.%s", strings.Split(table, ".")[0], newTable))
		util.Loggrs.Info("发生错误，已删除临时表！")
		return
	}

	//set query_timeout=3600
	var doneC = make(chan int)
	totalSizeM := fmt.Sprintf("%.2f", float64(c.Count))

	cs = color.New()
	olapname := fmt.Sprintf("%s.%s", strings.Split(table, ".")[0], newTable)
	exlname := fmt.Sprintf("%s.%s", strings.Split(table, ".")[0], oldTable)

	if tools.Versions(db) >= 2.5 {
		taskname := fmt.Sprintf("setbuckets_%s_%d", xid.New().String(), time.Now().UnixMicro())
		execSQL := fmt.Sprintf("submit /*+set_var(query_timeout=14400,pipeline_dop=0,exec_mem_limit=214748364800)*/ task %s as insert into %s select * from %s", taskname, olapname, exlname)
		util.Loggrs.Info(cs.Add(color.FgHiGreen).Sprint(execSQL))
		if c.Count > 0 {
			err := tools.Submit(execSQL, taskname, olapname, db)
			if err != nil {
				util.Loggrs.Error(err.Error() + " -> " + execSQL)
				return
			}
		}
	} else {
		tic := time.Tick(60 * time.Second)
		go func(cs chan int) {
			for {
				select {
				case <-doneC:
				case <-tic:
					var ct Ct2
					sql = fmt.Sprintf("select count(*) as count from %s.%s", strings.Split(table, ".")[0], newTable)
					row := db.Raw(sql).Scan(&ct)
					if row.Error != nil {
						util.Loggrs.Error(row.Error.Error())
						time.Sleep(10 * time.Second)
						continue
					}
					currSizeM := fmt.Sprintf("%.2f", float64(ct.Count))
					processRate := float64(ct.Count) / float64(c.Count) * 100
					rate := strconv.FormatFloat(processRate, 'f', 2, 64)
					util.Loggrs.Info(fmt.Sprintf("%v (%v/%v)%v", fmt.Sprintf("__Begin_Data_Loading[%s]", newTable), currSizeM, totalSizeM, strings.Repeat("#", int(processRate/2))+rate+"%"))
					if ct.Count >= c.Count {
						return
					}
				}
			}
		}(doneC)

		execSQL := fmt.Sprintf("set query_timeout=7200;pipeline_dop=0;set exec_mem_limit=214748364800;insert into %s select * from %s", olapname, exlname)
		util.Loggrs.Info(cs.Add(color.FgHiYellow).Sprint(execSQL))
		row = db.Exec(execSQL)
		if row.Error != nil {
			util.Loggrs.Error(row.Error.Error())
			return
		}
	}
	util.Loggrs.Info(cs.Add(color.FgHiGreen).Sprint("insert into done!"))

	close(doneC)
	util.Loggrs.Info(fmt.Sprintf("%s (%s/%s)%s", fmt.Sprintf("__End__Data_Finished[%s]", newTable), totalSizeM, totalSizeM, strings.Repeat("#", 50)+"100%"))
	util.Loggrs.Info("数据写入成功")

	if !strings.Contains(util.P.Action, "alter") {
		return
	}

	result, c1, c2 := tools.CheckSum(db, olapname, exlname)
	if result == "Fail" {
		util.Loggrs.Error(fmt.Sprintf("[%s](%s)、[%s](%s)数据量不一致！熔断！", olapname, cs.Add(color.FgHiRed).Sprint(c1), exlname, cs.Add(color.FgHiGreen).Sprint(c2)))
		return
	}

	renameTable := fmt.Sprintf("buckets_swap_bts_%d", time.Now().Unix())
	util.Loggrs.Info("开始交换表名")
	sql = fmt.Sprintf("use %s;alter table %s rename %s", dbname, oldTable, renameTable)
	util.Loggrs.Info(sql)
	row = db.Exec(sql)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		return
	}
	sql = fmt.Sprintf("use %s;alter table %s rename %s", dbname, newTable, oldTable)
	util.Loggrs.Info(sql)
	row = db.Exec(sql)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		return
	}
	util.Loggrs.Info("数据交互完成, 分桶修改完成!")

	if !strings.Contains(util.P.Action, "drop") {
		return
	}
	util.Loggrs.Info("删除旧表")
	sql = fmt.Sprintf("drop table %s.%s", strings.Split(table, ".")[0], renameTable)
	util.Loggrs.Info(sql)
	row = db.Exec(sql)
	if row.Error != nil {
		util.Loggrs.Error(row.Error.Error())
		return
	}
	util.Loggrs.Info("删除成功!")
	util.Loggrs.Info("完成整改: " + table)
	state = "SUCCESSED"
}
