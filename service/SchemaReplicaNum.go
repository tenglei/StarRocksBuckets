package service

import (
	"fmt"
	"github.com/fatih/color"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
)

// Distribution 副本分布情况(百分比)
func Distribution(stable string)  {
	c := color.New()

	db, err := conn.StarRocks(util.SrCfg)
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
	var backendsIds []string
	for _, m3 := range n {
		backendsIds = append(backendsIds,m3["BackendId"].(string) )
	}
	//获取tablet分布百分比
	var m []map[string]interface{}
	r = db.Raw(fmt.Sprintf("ADMIN SHOW REPLICA DISTRIBUTION FROM %s",stable)).Scan(&m)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
		return
	}

	var replica []map[string]interface{}
	if util.P.BackendId != "" {
		//获取tablet分布详细情况
			r = db.Raw(fmt.Sprintf("ADMIN SHOW REPLICA STATUS FROM %s",stable)).Scan(&replica)
			if r.Error != nil {
				fmt.Println(r.Error.Error())
				return
			}
	}

	fmt.Println("副本分布情况(百分比)")
	fmt.Println(fmt.Sprintf("%-18s %-15s %-15s %-10s %-10s","BE","BackendId","ReplicaNum","Graph","Percent"))
	var rNum int64

	for _, m2 := range m {
			c := color.New()
			var BackendHost string
			for _, m3 := range n {
				if m2["BackendId"].(string) == m3["BackendId"].(string) {
					BackendHost = m3["IP"].(string)
				}
			}

			replicaNum, _ := strconv.ParseInt(m2["ReplicaNum"].(string), 10, 64)
			rNum = rNum + replicaNum
			if m2["Graph"].(string) == "" {
				fmt.Println(fmt.Sprintf("%-18s %-15s %-15s %-10s %-10s",BackendHost,m2["BackendId"].(string),m2["ReplicaNum"].(string),m2["Graph"].(string),m2["Percent"].(string)))
			}else {
				fmt.Println(c.Add(color.FgHiBlue).Sprint(fmt.Sprintf("%-18s %-15s %-15s %-10s %-10s",BackendHost,m2["BackendId"].(string),m2["ReplicaNum"].(string),m2["Graph"].(string),m2["Percent"].(string))))
			}
			if util.P.BackendId == "all" {
				util.P.BackendId = strings.Join(backendsIds,",")
			}
			for _, backendId := range strings.Split(util.P.BackendId,",") {
				if backendId == m2["BackendId"].(string) {
					//获取tablet分布详细情况
					distributionNode(replica,backendId)
				}
			}
	}
	fmt.Println(fmt.Sprintf("ReplicaCount:(%s)",c.Add(color.FgHiGreen).Sprint(rNum)))
	fmt.Println()
	fmt.Println()
}
