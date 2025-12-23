/*
 *@author  chengkenli
 *@project pjstack
 *@package permit
 *@file    Permit
 *@date    2025/1/14 15:39
 */

package permit

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/util"
	"strings"
)

// Permitgrants
// 重新授权
func Permitgrants(table string) {
	if !strings.Contains(util.P.Action, "drop") && !strings.Contains(util.P.Action, "alter") {
		return
	}
	
	// 检查srqa连接是否可用
	if srqa == nil {
		if util.Loggrs != nil {
			util.Loggrs.Info("跳过权限检查: 数据库连接未初始化")
		}
		return
	}
	
	util.Loggrs.Info("检索权限")
	
	// 检查audit数据库是否存在
	var dbExists int64
	checkDB := srqa.Raw("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = 'audit'").Scan(&dbExists)
	if checkDB.Error != nil || dbExists == 0 {
		if util.Loggrs != nil {
			util.Loggrs.Info("跳过权限检查: audit数据库不存在")
		}
		return
	}
	
	// 检查audit.ddlcatch表是否存在
	var tableExists int64
	checkTable := srqa.Raw("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'audit' AND TABLE_NAME = 'ddlcatch'").Scan(&tableExists)
	if checkTable.Error != nil || tableExists == 0 {
		if util.Loggrs != nil {
			util.Loggrs.Info("跳过权限检查: audit.ddlcatch表不存在")
		}
		return
	}
	
	var c util.Catch
	r := srqa.Raw(fmt.Sprintf("select distinct app,user,grants from audit.ddlcatch where app='%s' and grants like '%%%s%%'", util.SrConfig.Host, table)).Scan(&c)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return
	}
	var grants []string
	for i, item := range c {
		grants = append(grants, item.Grants)
		util.Loggrs.Info(fmt.Sprintf("#%03d %-8s %-20s %s", i, item.App, item.User, item.Grants))
	}

	if len(grants) <= 0 {
		return
	}

	util.Loggrs.Info("重新赋权")
	cc := color.New()
	for _, grant := range grants {
		r := srother.Exec(grant)
		if r.Error != nil {
			util.Loggrs.Error("Failed >", cc.Add(color.FgHiRed).Sprint(grant))
			util.Loggrs.Error("Failed >", cc.Add(color.FgHiRed).Sprint(r.Error.Error()))
			continue
		}
		util.Loggrs.Info("Ok     >", cc.Add(color.FgHiGreen).Sprint(grant))
	}
	util.Loggrs.Info("done.")
}
