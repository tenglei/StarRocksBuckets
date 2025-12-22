package service

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"gorm.io/gorm"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
	"sync"
)

// ScanSchemaPartition 空分区与非空分区组装
func ScanSchemaPartition(stable string) (string, error) {
	db, err := conn.StarRocks(util.P.App)
	if err != nil {
		return "", err
	}
	/*获取分区信息*/
	var m []map[string]interface{}
	r := db.Raw(fmt.Sprintf("show partitions from %s", stable)).Scan(&m)
	if r.Error != nil {
		return "", err
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
	//util.Loggrs.Info(fmt.Sprintf("[%s]空分区:%s", c.Add(color.FgHiRed).Sprint(len(NulsName)), strings.Join(NulsName, ",")))
	//util.Loggrs.Info(fmt.Sprintf("[%s]非空分区:%s", c.Add(color.FgHiGreen).Sprint(len(NosName)), strings.Join(NosName, ",")))
	util.Loggrs.Info(fmt.Sprintf("[%d]空历史分区，[%d]非空分区", len(NulsName), len(NosName)))
	/*获取建表信息*/
	var tb map[string]interface{}
	r = db.Raw("show create table " + stable).Scan(&tb)
	if r.Error != nil {
		return "", r.Error
	}
	source := strings.Split(tb["Create Table"].(string), "\n")
	/*组合字段*/
	var target []string
	var once sync.Once
	for _, s := range source {
		/*过滤有存储容量分区*/
		if strings.Contains(strings.ToUpper(s), "PARTITION") && strings.Contains(strings.ToUpper(s), "VALUES [(") {
			once.Do(func() {
				target = append(target, "(STARROCKS_PARTITION_VALUES_LONG_LENGTH)")
			})
			continue
		}
		target = append(target, s)
	}

	partitionRange, err := ScanSchemaGetPartitionRange(stable)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return "", err
	}
	var STARROCKS_PARTITION_VALUES_LONG_LENGTH []string
	for _, nos := range Nos {
		for _, m2 := range partitionRange {
			if name, ok := m2[nos]; ok {
				STARROCKS_PARTITION_VALUES_LONG_LENGTH = append(STARROCKS_PARTITION_VALUES_LONG_LENGTH, name)
			}
		}
	}
	create := strings.NewReplacer("STARROCKS_PARTITION_VALUES_LONG_LENGTH", strings.Join(STARROCKS_PARTITION_VALUES_LONG_LENGTH, ",\n")).Replace(strings.Join(target, "\n"))

	if len(STARROCKS_PARTITION_VALUES_LONG_LENGTH) != len(Nos) {
		return create, errors.New(fmt.Sprintf("统计后的实分区与原实分区不一致！新：%d,旧：%d", len(STARROCKS_PARTITION_VALUES_LONG_LENGTH), len(Nos)))
	}
	return create, nil
}

func dynamic(schema string, createStruct []string, db *gorm.DB) {
	c := color.New()
	for _, s := range createStruct {
		if !strings.Contains(strings.ToLower(s), "dynamic_partition.end") {
			continue
		}
		end := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.Split(s, "=")[1], `"`, ""), ",", ""), " ", "")
		atoi, _ := strconv.Atoi(end)
		if atoi >= 35 {
			util.Loggrs.Info(`被动触发设置 "dynamic_partition.end" = "30", 原设置: ` + s)
			r := db.Exec(fmt.Sprintf(`ALTER TABLE %s SET ("dynamic_partition.end" = "30")`, schema))
			if r.Error != nil {
				util.Loggrs.Error(c.Add(color.FgHiRed).Sprint(r.Error.Error()))
			}
		}
		return
	}
}
