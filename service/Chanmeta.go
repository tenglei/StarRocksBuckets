/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    Chanmeta
 *@date    2025/4/17 10:49
 */

package service

import (
	"github.com/fatih/color"
	"setbuckets/metaload"
	"setbuckets/util"
)

func (j *Job) Chanmetadata() {
	return

	for {
		select {
		case meta, _ := <-t.MetaData:
			c := color.New()
			inProgress = 1
			var avg util.SrAvgs
			for _, m := range util.MetaLink {
				if m["app"].(string) == util.P.App {
					avg = util.SrAvgs{
						Host: m["feip"].(string),
						Port: int(m["feport"].(int32)),
						User: m["user"].(string),
						Pass: m["password"].(string),
					}
				}
			}
			if len(meta) == 0 {
				return
			}
			metaload.MetaStreamload(&avg, &meta, "ops.starrocks_devops_optimize_information")
			util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("MET:EXC > "), len(meta), " FINISHED")
			inProgress = 0
		}
	}
}
