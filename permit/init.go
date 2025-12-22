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

func init() {
	var err error
	srqa, err = conn.StarRocks("sr-qa")
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}

	srother, err = conn.StarRocks(util.SrConfig)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
}
