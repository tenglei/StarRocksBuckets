/*
 *@author  chengkenli
 *@project setbuckets
 *@package service
 *@file    init
 *@date    2025/3/12 18:51
 */

package service

func init() {
	go t.ChanTicker()
	go t.SyncPartition()
	go t.Chanmetadata()
}
