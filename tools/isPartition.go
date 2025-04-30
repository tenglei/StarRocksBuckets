package tools

import (
	"fmt"
	"github.com/fatih/color"
	"gorm.io/gorm"
	"setbuckets/util"
)

func IsPartition(db *gorm.DB) (string, string, int) {
	/*获取分区信息*/
	var m []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", util.P.Table)).Scan(&m)
	if r.Error != nil {
		return "", "", 0
	}
	/*获取到有存储的分区*/
	var Nuls, Nos, Nds []string
	var NulsName, NosName []string
	for _, s := range m {
		if s["DataSize"].(string) == ".000 " {
			Nuls = append(Nuls, s["PartitionName"].(string))
			Nds = append(Nds, s["DataSize"].(string))
			NulsName = append(NulsName, fmt.Sprintf("[%s] [%s]", s["PartitionName"].(string), s["DataSize"].(string)))
			continue
		}
		if s["DataSize"].(string) == "0B" {
			Nuls = append(Nuls, s["PartitionName"].(string))
			Nds = append(Nds, s["DataSize"].(string))
			NulsName = append(NulsName, fmt.Sprintf("[%s] [%s]", s["PartitionName"].(string), s["DataSize"].(string)))
			continue
		}
		Nos = append(Nos, s["PartitionName"].(string))
		NosName = append(NosName, fmt.Sprintf("[%s] [%s]", s["PartitionName"].(string), s["DataSize"].(string)))
	}

	c := color.New()
	var nil, nos string
	nos = c.Add(color.FgHiGreen).Sprint(len(NosName))

	if len(NulsName) >= 1000 {
		nil = c.Add(color.FgHiRed).Sprint(len(NulsName))
		return nil, nos, len(NulsName) + len(NosName)
	} else if len(NulsName) >= 30 {
		nil = c.Add(color.FgHiYellow).Sprint(len(NulsName))
		return nil, nos, len(NulsName) + len(NosName)
	} else {
		nil = c.Add(color.FgHiGreen).Sprint(len(NulsName))
		return nil, nos, len(NulsName) + len(NosName)
	}
}
