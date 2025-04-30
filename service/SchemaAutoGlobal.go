/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaAutoGlobal
 *@date    2025/3/18 17:55
 */

package service

import (
	"github.com/rs/xid"
	"setbuckets/tools"
	"setbuckets/util"
	"strings"
	"sync"
)

func SetAutoGlobal() {
	ch := make(chan struct{}, util.P.Thread)
	var wg sync.WaitGroup
	for _, table := range strings.Split(util.P.Table, ",") {
		wg.Add(1)

		go func(table string) {
			defer func() {
				<-ch
				wg.Done()
			}()

			ch <- struct{}{}

			taskname := xid.New().String()
			bucketInfo := tools.GetBuckets(table)
			SetParGlobal(table, taskname, int64(bucketInfo.Best))
		}(table)
	}
	wg.Wait()

}
