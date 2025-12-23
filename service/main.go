package service

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/xid"
	"setbuckets/permit"
	"setbuckets/tools"
	"setbuckets/util"
	"strings"
	"time"
)

type bucketStr struct {
	App          string `json:"app"`
	Best         int    `json:"best"`
	Buckets      string `json:"buckets"`
	Client       string `json:"client"`
	Conservative int    `json:"conservative"`
	Datasize     string `json:"datasize"`
	Msg          string `json:"msg"`
	Normal       bool   `json:"normal"`
	Table        string `json:"table"`
}

var fix util.Fix

func Run() {
	defer func() {
		if util.P.PartitionSet {
			return
		}
		if util.P.Buckets < 1 {
			return
		}
		if !util.P.ClearPartition {
			return
		}
		if !strings.Contains(util.P.Action, "drop") || !strings.Contains(util.P.Action, "alter") {
			return
		}
		//所有处理已经完成了，发送一个信号告诉chan可以进行数据记录入表了。
		t.MetaDone <- struct{}{}
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
	}()

	//////////////////////////////////
	//////////////////////////////////
	//////////////////////////////////

	if len(util.P.Table) != 0 && util.P.ClearPartition && !util.P.Auto && util.P.Buckets == 0 {
		AutoSetPartitions()
		
		// 刷新表统计信息
		RefreshMultipleTablesStats(util.P.Table)
		
		// 如果指定了 -id 参数，则显示tablet分布信息
		if util.P.BackendId != "" {
			fmt.Println()
			Distribution(util.P.Table)
			fmt.Println()
		}
		return
	}

	if len(util.P.Table) != 0 && util.P.PartitionSet && util.P.Auto {
		SetAutoGlobal()
		permit.Permitgrants(util.P.Table)
		
		// 刷新表统计信息
		RefreshMultipleTablesStats(util.P.Table)
		
		// 如果指定了 -id 参数，则显示tablet分布信息
		if util.P.BackendId != "" {
			fmt.Println()
			Distribution(util.P.Table)
			fmt.Println()
		}
		return
	}

	if len(util.P.Table) != 0 && util.P.PartitionSet && util.P.Buckets != 0 && len(util.P.PartitionName) == 0 {
		SetParGlobal(util.P.Table, xid.New().String(), util.P.Buckets)
		util.Loggrs.Info("TOP:JOB > done.")
		
		// 刷新表统计信息
		RefreshMultipleTablesStats(util.P.Table)
		
		// 如果指定了 -id 参数，则显示tablet分布信息
		if util.P.BackendId != "" {
			fmt.Println()
			Distribution(util.P.Table)
			fmt.Println()
		}
		return
	}

	if len(util.P.Table) != 0 && util.P.PartitionSet {
		SetParBuckets()
		util.Loggrs.Info("TOP:JOB > done.")
		
		// 刷新表统计信息
		RefreshMultipleTablesStats(util.P.Table)
		
		// 如果指定了 -id 参数，则显示tablet分布信息
		if util.P.BackendId != "" {
			fmt.Println()
			Distribution(util.P.Table)
			fmt.Println()
		}
		return
	}

	if len(util.P.Table) != 0 {
		if util.P.Auto || util.P.Buckets > 0 {
			ScanSchemaAutoSetBuckets()
			util.Loggrs.Info("done.")
			
			// 刷新表统计信息
			RefreshMultipleTablesStats(util.P.Table)
			
			// 如果指定了 -id 参数，则在分桶修改后显示tablet分布信息
			if util.P.BackendId != "" {
				fmt.Println()
				Distribution(util.P.Table)
				fmt.Println()
			}
			return
		}
	}

	schema := tools.Other()

	if len(util.P.Table) != 0 {
		b := getbuckets(util.P.Table)
		c := color.New()
		fmt.Println("\nOptions:\nstarrocks buckets分桶体检1.0")
		fmt.Println()
		if b.Normal {
			fmt.Println(c.Add(color.FgHiGreen).Sprint("正常"))
			fmt.Println(fmt.Sprintf("集群: %s", c.Add(color.FgHiCyan).Sprint(b.App)))
			fmt.Println(fmt.Sprintf("表名: %s", c.Add(color.FgHiCyan).Sprint(b.Table)))
			fmt.Println(fmt.Sprintf("分桶: %s", c.Add(color.FgHiCyan).Sprint(b.Buckets)))
			fmt.Println(fmt.Sprintf("容量: %s", c.Add(color.FgHiCyan).Sprint(b.Datasize)))
			fmt.Println(fmt.Sprintf("类型: %s", c.Add(color.FgHiWhite).Sprint(schema.SchemaType)))
			fmt.Println(fmt.Sprintf("大小: %s", c.Add(color.FgHiGreen).Sprint(schema.Size)))
			fmt.Println(fmt.Sprintf("数量: %s", c.Add(color.FgHiMagenta).Sprint(schema.RowCount)))
			fmt.Println(fmt.Sprintf("时间: %s", c.Add(color.FgHiBlue).Sprint(schema.CreateTime)))
			fmt.Println(fmt.Sprintf("版本: %s", c.Add(color.FgHiCyan).Sprint(schema.MaxVisibleVersion)))
			fmt.Println(fmt.Sprintf("注释: %s", c.Add(color.FgHiYellow).Sprint(schema.Comment)))
		} else {
			fmt.Println(c.Add(color.FgHiYellow).Sprint("倾斜"))
			fmt.Println(fmt.Sprintf("集群: %s", c.Add(color.FgHiCyan).Sprint(b.App)))
			fmt.Println(fmt.Sprintf("表名: %s", c.Add(color.FgHiCyan).Sprint(b.Table)))
			fmt.Println(fmt.Sprintf("分桶: %s", c.Add(color.FgHiMagenta).Sprint(b.Buckets)))
			fmt.Println(fmt.Sprintf("容量: %s", c.Add(color.FgHiCyan).Sprint(b.Datasize)))
			fmt.Println(fmt.Sprintf("类型: %s", c.Add(color.FgHiWhite).Sprint(schema.SchemaType)))
			fmt.Println(fmt.Sprintf("保底: %s", c.Add(color.FgHiYellow).Sprint(b.Conservative)))
			fmt.Println(fmt.Sprintf("建议: %s", c.Add(color.FgHiGreen).Sprint(b.Best)))
			fmt.Println(fmt.Sprintf("大小: %s", c.Add(color.FgHiGreen).Sprint(schema.Size)))
			fmt.Println(fmt.Sprintf("数量: %s", c.Add(color.FgHiMagenta).Sprint(schema.RowCount)))
			fmt.Println(fmt.Sprintf("时间: %s", c.Add(color.FgHiBlue).Sprint(schema.CreateTime)))
			fmt.Println(fmt.Sprintf("版本: %s", c.Add(color.FgHiCyan).Sprint(schema.MaxVisibleVersion)))
			fmt.Println(fmt.Sprintf("注释: %s", c.Add(color.FgHiYellow).Sprint(schema.Comment)))
		}
		if schema.Sum >= 2 {
			fmt.Println(fmt.Sprintf("分区: 动态分区:%s,空分区:[%s],非空分区:[%s]", schema.Enable, schema.Nil, schema.Nos))
		}
		if schema.Mv != nil {
			fmt.Println(fmt.Sprintf("物化: %s", strings.Join(schema.Mv, ",")))
		}

		if util.P.List {
			ScanSchemaSortKey()
		}

		if util.P.TabletSize > 0 {
			fmt.Println()
			TabletRepSize(b.Datasize)
		}
		fmt.Println()
		Distribution(util.P.Table)
		fmt.Println()
		return
	}
}
