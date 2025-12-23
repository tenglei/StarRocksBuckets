/*
 *@author  chengkenli
 *@project setbuckets
 *@package permit
 *@file    init
 *@date    2025/4/16 14:29
 */

package permit

import (
	"gorm.io/gorm"
	"setbuckets/conn"
	"setbuckets/util"
)

var srqa, srother *gorm.DB

// InitConnections 初始化数据库连接
// 在用户输入集群连接信息并验证成功后调用
func InitConnections() error {
	var err error
	
	// 使用全局配置连接StarRocks
	srqa, err = conn.StarRocks(util.SrConfig)
	if err != nil {
		if util.Loggrs != nil {
			util.Loggrs.Error("初始化srqa连接失败: ", err.Error())
		}
		return err
	}
	
	srother, err = conn.StarRocks(util.SrConfig)
	if err != nil {
		if util.Loggrs != nil {
			util.Loggrs.Error("初始化srother连接失败: ", err.Error())
		}
		return err
	}
	
	return nil
}
