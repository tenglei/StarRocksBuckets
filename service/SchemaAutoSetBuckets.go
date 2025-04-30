/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    ScanSchemaAutoSetBuckets
 *@date    2024/8/2 9:40
 */

package service

import (
	"setbuckets/permit"
	"setbuckets/tools"
	"setbuckets/util"
	"strings"
	"sync"
)

// ScanSchemaAutoSetBuckets 自动分桶设置
func ScanSchemaAutoSetBuckets() {

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

			var bucket int64
			if util.P.Auto {
				bucketInfo := tools.GetBuckets(table)
				bucket = int64(bucketInfo.Best)
			} else {
				bucket = util.P.Buckets
			}

			ScanSchemaSetBuckets(table, bucket)
			permit.Permitgrants(table)
		}(table)
	}
	wg.Wait()
}
