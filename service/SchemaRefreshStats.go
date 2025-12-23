/*
 *@author  chengkenli
 *@project StarRocksBuckets
 *@package service
 *@file    SchemaRefreshStats
 *@date    2025/12/23
 */

package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
	"time"
)

// RefreshTableStats 刷新表的统计信息
// 在分桶修改后调用，确保后续查询能获取最新的统计信息
func RefreshTableStats(table string) error {
	if table == "" {
		return fmt.Errorf("表名不能为空")
	}
	
	c := color.New()
	
	// 解析数据库和表名
	parts := strings.Split(table, ".")
	if len(parts) != 2 {
		return fmt.Errorf("表名格式错误，应为: database.table")
	}
	
	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	
	// 关闭数据库连接
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()
	
	// 第一步：等待 tablet 物理创建完成
	if util.Loggrs != nil {
		util.Loggrs.Info(c.Add(color.FgHiYellow).Sprint("🔄 等待tablet物理创建完成..."))
	}
	
	// 检查 tablet 是否就绪（最多等待30秒）
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		var replicas []map[string]interface{}
		checkSQL := fmt.Sprintf("ADMIN SHOW REPLICA DISTRIBUTION FROM %s", table)
		result := db.Raw(checkSQL).Scan(&replicas)
		
		if result.Error == nil && len(replicas) > 0 {
			// 检查是否有副本数据
			hasReplicas := false
			for _, replica := range replicas {
				if replicaNum, ok := replica["ReplicaNum"].(string); ok {
					if num, err := strconv.ParseInt(replicaNum, 10, 64); err == nil && num > 0 {
						hasReplicas = true
						break
					}
				}
			}
			
			if hasReplicas {
				if util.Loggrs != nil {
					util.Loggrs.Info(c.Add(color.FgHiGreen).Sprint("✅ Tablet已就绪"))
				}
				break
			}
		}
		
		if i < maxRetries-1 {
			if util.Loggrs != nil {
				util.Loggrs.Info(fmt.Sprintf("⏳ 等待中... (%d/%d)", i+1, maxRetries))
			}
			time.Sleep(3 * time.Second)
		}
	}
	
	// 第二步：执行 ANALYZE TABLE 命令刷新统计信息
	analyzeSQL := fmt.Sprintf("ANALYZE TABLE %s", table)
	
	if util.Loggrs != nil {
		util.Loggrs.Info(c.Add(color.FgHiYellow).Sprint("🔄 刷新表统计信息..."))
		util.Loggrs.Info(fmt.Sprintf("执行: %s", analyzeSQL))
	}
	
	result := db.Exec(analyzeSQL)
	if result.Error != nil {
		// 记录错误但不中断流程，因为统计信息刷新失败不应影响主流程
		if util.Loggrs != nil {
			util.Loggrs.Warn(c.Add(color.FgHiYellow).Sprint("⚠️  统计信息刷新失败（不影响主流程）: "), result.Error.Error())
		}
		return result.Error
	}
	
	if util.Loggrs != nil {
		util.Loggrs.Info(c.Add(color.FgHiGreen).Sprint("✅ 表统计信息已刷新"))
	}
	
	return nil
}

// RefreshMultipleTablesStats 批量刷新多个表的统计信息
// 用于处理逗号分隔的表名列表
func RefreshMultipleTablesStats(tables string) {
	if tables == "" {
		return
	}
	
	tableList := strings.Split(tables, ",")
	for _, table := range tableList {
		table = strings.TrimSpace(table)
		if table != "" {
			// 即使单个表刷新失败，也继续处理下一个
			_ = RefreshTableStats(table)
		}
	}
}
