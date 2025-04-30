/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    SchemaAutoSetPartitions
 *@date    2025/4/16 16:54
 */

package service

import (
	"setbuckets/permit"
	"setbuckets/util"
	"strings"
	"sync"
)

func AutoSetPartitions() {
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
			SetPartitions(table)
			permit.Permitgrants(table)
		}(table)
	}
	wg.Wait()
}
