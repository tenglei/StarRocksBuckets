package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
	"sync"
)

// distributionNode 每个节点的副本分布情况
func distributionNode(replica []map[string]interface{},backendId string)  {
	db, err := conn.StarRocks(util.P.App)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var n []map[string]interface{}
	r := db.Raw("show backends").Scan(&n)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
		return
	}
	var tablet []string
	ch := make(chan struct{},100)
	var wg sync.WaitGroup
	for _, m2 := range replica {
		wg.Add(1)
		go func(m2 map[string]interface{}) {
			defer func() {
				<- ch
				wg.Done()
			}()

			ch <- struct{}{}
			if m2["BackendId"].(string) == backendId {
				var cm map[string]interface{}
				r := db.Raw("show tablet "+m2["TabletId"].(string)).Scan(&cm)
				if r.Error != nil {
					fmt.Println(r.Error.Error())
					return
				}
				var tm []map[string]interface{}
				r = db.Raw(cm["DetailCmd"].(string)).Scan(&tm)
				if r.Error != nil {
					fmt.Println(r.Error.Error())
					return
				}
				var ds []string
				for _, m3 := range tm {
					ds= append(ds, bytesUnit(m3["DataSize"].(string)))
				}
				tablet = append(tablet, fmt.Sprintf("%s(%s)",m2["TabletId"].(string),strings.Join(ds,",")))
			}
		}(m2)
	}
	wg.Wait()
	fmt.Println(strings.Join(tablet," | "))
}

func bytesUnit(v string) string {
	c := color.New()
	parse, _ := strconv.ParseFloat(v, 64)
	if parse / 1024/1024/1024 >= 1 {
		return c.Add(color.FgHiRed).Sprint(fmt.Sprintf("%0.2fGB",parse / 1024/1024/1024))
	}
	if parse / 1024/1024 >= 700 {
		return c.Add(color.FgHiMagenta).Sprint(fmt.Sprintf("%0.2fMB",parse / 1024/1024))
	}
	if parse / 1024/1024 >= 200 && parse / 1024/1024 < 700 {
		return c.Add(color.FgHiGreen).Sprint(fmt.Sprintf("%0.2fMB",parse / 1024/1024))
	}
	if parse / 1024/1024 >= 1 {
		return c.Add(color.FgHiYellow).Sprint(fmt.Sprintf("%0.2fMB",parse / 1024/1024))
	}
	if parse / 1024 >= 1 {
		return c.Add(color.FgHiCyan).Sprint(fmt.Sprintf("%0.2fKB",parse / 1024))
	}
	if parse >= 1 {
		return c.Add(color.FgHiBlue).Sprint(fmt.Sprintf("%0.2fBytes",parse))
	}
	return c.Add(color.FgHiBlack).Sprint(v)
}