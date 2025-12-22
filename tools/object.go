/*
 *@author  chengkenli
 *@project setbuckets
 *@package tools
 *@file    object
 *@date    2024/7/29 14:38
 */

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"gorm.io/gorm"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"setbuckets/conn"
	"setbuckets/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Schemas struct {
	CreateTime        string
	Comment           string
	Size              string
	RowCount          string
	SchemaType        string
	Nil               string
	Nos               string
	Sum               int
	Enable            string
	Mv                []string
	MaxVisibleVersion string
}

func Other() Schemas {
	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		fmt.Println(err.Error())
		return Schemas{}
	}

	schema := strings.Split(util.P.Table, ".")

	//提取创建日期
	var m map[string]interface{}
	sql := fmt.Sprintf("select `TABLE_CATALOG`,`TABLE_SCHEMA`,`TABLE_NAME`,`TABLE_TYPE`,`ENGINE`,`CREATE_TIME`,`TABLE_COMMENT`,`TABLE_ROWS` from information_schema.tables where TABLE_SCHEMA='%s' and TABLE_NAME='%s'", schema[0], schema[1])
	r := db.Raw(sql).Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return Schemas{}
	}
	//提取表总容量大小和行数
	var s []map[string]interface{}
	r = db.Raw("show data from " + util.P.Table).Scan(&s)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return Schemas{}
	}

	//PARTITION BY RANGE
	var cqm []map[string]interface{}
	r = db.Raw(fmt.Sprintf("show create table %s", util.P.Table)).Scan(&cqm)
	if r.Error != nil {
		util.Loggrs.Error(err.Error())
		return Schemas{}
	}
	var createSQL string
	for _, m := range cqm {
		createSQL = fmt.Sprintf("%v", m["Create Table"])
	}

	var enable string
	ty := "full"
	for _, s2 := range strings.Split(createSQL, "\n") {
		if strings.Contains(s2, "PARTITION BY RANGE") {
			ty = "incre"
		}
		c := color.New()
		if strings.Contains(s2, "dynamic_partition.enable") {
			a := strings.NewReplacer(`"`, "", " ", "", ",", "").Replace(strings.Split(s2, "=")[1])
			if a == "true" {
				enable = c.Add(color.FgHiGreen).Sprint(a)
			} else if a == "false" {
				enable = c.Add(color.FgHiRed).Sprint(a)
			}
		}
	}
	//采集空分区和非空分区数量
	nil, nos, sum := IsPartition(db)
	mv := Materialized(db, util.P.Table)

	//获取最大的版本数
	var vv map[string]interface{}
	db.Raw(fmt.Sprintf("show partitions from %s order by VisibleVersion desc limit 1", util.P.Table)).Scan(&vv)

	return Schemas{
		CreateTime:        m["CREATE_TIME"].(time.Time).Format("2006-01-02 15:04:05"),
		Comment:           m["TABLE_COMMENT"].(string),
		Size:              s[0]["Size"].(string),
		RowCount:          s[0]["RowCount"].(string),
		SchemaType:        ty,
		Nil:               nil,
		Nos:               nos,
		Enable:            enable,
		Sum:               sum,
		Mv:                mv,
		MaxVisibleVersion: fmt.Sprintf("%s > %s", vv["PartitionName"].(string), vv["VisibleVersion"].(string)),
	}
}

// SumMapValues 定义一个辅助函数，用于计算map中所有int64值的总和
func SumMapValues(m map[string]int64) int64 {
	sum := int64(0)
	for _, v := range m {
		sum += v
	}
	return sum
}

// RmStrSlice /*数组去重*/
func RmStrSlice(strs []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复字符串
	for _, e := range strs {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// StrInSlice 检查切片中是否存在某个元素
func StrInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// 寻找元素位置
func FindKeyRank(slice []map[string]int64, searchKey string) (int, bool) {
	rank := 0 // 排行计数器
	// 遍历切片中的每个map
	for _, m := range slice {
		// 检查map中是否存在指定的键
		if _, ok := m[searchKey]; ok {
			// 如果找到了匹配的键，返回排行和true
			return rank, true // 排名从1开始计算
		}
		// 增加排行计数器，即使当前map中没有找到键
		rank += len(m)
	}
	// 如果没有找到，返回0和false
	return 0, false
}

// ScanCountDistinct 统计每个字段1000行内重复次数
func ScanCountDistinct(column, table string) int64 {
	db, err := conn.StarRocks(util.SrConfig)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	var m []map[string]interface{}
	sql := fmt.Sprintf(`select count(*),%s from (select DISTINCT %s from (select %s from %s limit 1000) a ) b group by %s`, column, column, column, table, column)
	r := db.Raw(sql).Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return 0
	}
	return r.RowsAffected
}

// CheckSum 计算两端数据流
func CheckSum(db *gorm.DB, t1, t2 string) (string, int64, int64) {
	fmt.Println("计算两端数据量")
	var c1, c2 map[string]interface{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		db.Raw("select count(*) as count from " + t1).Scan(&c1)
	}()
	go func() {
		defer wg.Done()
		db.Raw("select count(*) as count from " + t2).Scan(&c2)
	}()
	wg.Wait()

	if c1["count"] == nil || c2["count"] == nil {
		return "Fail", -1, -1
	}

	fmt.Println(fmt.Sprintf("[%d]: %s", c1["count"].(int64), t1))
	fmt.Println(fmt.Sprintf("[%d]: %s", c2["count"].(int64), t2))

	if c1["count"].(int64) == c2["count"].(int64) {
		return "Ok", c1["count"].(int64), c2["count"].(int64)
	} else {
		return "Fail", c1["count"].(int64), c2["count"].(int64)
	}
}

// CheckRow 计算表数据流
func CheckRow(db *gorm.DB, t1 string) int64 {
	fmt.Println("计算数据量")
	var c1 map[string]interface{}
	db.Raw("select count(*) as count from " + t1).Scan(&c1)

	fmt.Println(fmt.Sprintf("[%d]: %s", c1["count"].(int64), t1))
	return c1["count"].(int64)
}

// Submit 异步提交
func Submit(sql, taskname, tablename string, db *gorm.DB) error {
	type run struct {
		QUERYID      string `bson:"QUERY_ID"`
		TASKNAME     string `bson:"TASK_NAME"`
		CREATETIME   string `bson:"CREATE_TIME"`
		FINISHTIME   string `bson:"FINISH_TIME"`
		STATE        string `bson:"STATE"`
		DATABASE     string `bson:"DATABASE"`
		DEFINITION   string `bson:"DEFINITION"`
		EXPIRETIME   string `bson:"EXPIRE_TIME"`
		ERRORCODE    string `bson:"ERROR_CODE"`
		ERRORMESSAGE string `bson:"ERROR_MESSAGE"`
		PROGRESS     string `bson:"PROGRESS"`
		EXTRAMESSAGE string `bson:"EXTRA_MESSAGE"`
	}
	r := db.Exec(sql)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return r.Error
	}

	var i int
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			//c := color.New()
			s := fmt.Sprintf("select * from information_schema.task_runs where task_name ='%s'", taskname)
			var task run
			r := db.Raw(s).Scan(&task)
			if r.Error != nil {
				util.Loggrs.Error(r.Error.Error())
				return r.Error
			}

			if i%60 == 0 {
				fmt.Println(fmt.Sprintf("%s  %s  %s  %s", time.Now().Format("2006-01-02 15:04:05"), task.STATE, task.PROGRESS, tablename))
			}

			if s == "" {
				return errors.New("information_schema.task_runs is nil")
			}
			if task.STATE == "SUCCESS" {
				return nil
			}
			if task.STATE == "FAILED" {
				return errors.New("task staus is " + task.STATE)
			}
			i++
		}
	}
}

/*匹配starrocks版本*/
func Versions(db *gorm.DB) float64 {
	type Version struct {
		Version string `bson:"version"`
	}
	var v Version
	sqln := fmt.Sprintf("select current_version() as version")
	db.Raw(sqln).Scan(&v)
	version, _ := strconv.ParseFloat(fmt.Sprintf("%s.%s", strings.Split(strings.Split(v.Version, " ")[0], ".")[0], strings.Split(strings.Split(v.Version, " ")[0], ".")[1]), 64)
	return version
}

func Materialized(db *gorm.DB, table string) []string {
	var m map[string]interface{}
	r := db.Raw(fmt.Sprintf("select inspect_related_mv('%s') as mv", table)).Scan(&m)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
		return nil
	}
	if m == nil || m["mv"] == nil {
		fmt.Println("inspect_related_mv is nil")
		return nil
	}
	fmt.Println(m["mv"].(string))

	type Data []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var d Data
	err := json.Unmarshal([]byte(m["mv"].(string)), &d)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	c := color.New()
	var mv []string
	for i, s := range d {
		mv = append(mv, fmt.Sprintf("%s>%s", c.Add(color.FgHiGreen).Sprint(i), c.Add(color.FgHiMagenta).Sprint(s.Name)))
	}
	return mv
}

// PrintProgress 用于在一行内打印进度条
func PrintProgress(current, total int, comment ...string) {
	percent := current * 100 / total                                                  // 计算进度百分比
	length := 20                                                                      // 设定进度条的长度
	s := strings.Repeat("■", current*length/total)                                    // 根据完成度填充进度条
	e := strings.Repeat("□", length-len(s)/3)                                         // 填充剩余的空格(这里一个□占用了3个字符，所以除以3)
	bar := s + e                                                                      // 拼接
	fmt.Printf("\033[2K\r%d%% [%s](%d/%d) %v", percent, bar, current, total, comment) // 使用ANSI转义序列将光标移动到行首
}

func Size(s string) float64 {
	s = strings.ToLower(s)
	//正则
	re := regexp.MustCompile(`\d+\.?\d*`)

	float, _ := strconv.ParseFloat(re.FindString(s), 64)

	if strings.Contains(s, "kb") {
		return float * 1024
	}
	if strings.Contains(s, "mb") {
		return float * 1024 * 1024
	}
	if strings.Contains(s, "gb") {
		return float * 1024 * 1024 * 1024
	}
	if strings.Contains(s, "tb") {
		return float * 1024 * 1024 * 1024 * 1024
	}
	return 0
}

// SliceGroup
// 切片分组
func SliceGroup(data []string) [][]string {
	// Use a map to group the strings by the number after "^"
	groups := make(map[int][]string)
	for _, entry := range data {
		parts := strings.Split(entry, "^")
		if len(parts) != 2 {
			continue // Skip if format is incorrect
		}
		number, err := strconv.Atoi(parts[1])
		if err != nil {
			continue // Skip if the number conversion fails
		}
		groups[number] = append(groups[number], entry)
	}

	// Create a 2D slice to hold the partitioned data
	var result [][]string
	for _, group := range groups {
		// Append the group to the result as a new slice
		result = append(result, group)
	}
	return result
}

// SplitSlice splits the input slice into as many fragments as possible, each having 'fragmentSize' number of elements.
// If the total number of elements isn't a multiple of 'fragmentSize', the remaining elements will be added to the last fragment.
func SplitSlice(slice []string, fragmentSize int) [][]string {
	if fragmentSize <= 0 {
		return nil // Return nil if the fragment size is not valid
	}

	totalElements := len(slice)
	numFragments := totalElements / fragmentSize
	remainder := totalElements % fragmentSize

	// Initialize the result 2D slice
	var result [][]string

	for i := 0; i < numFragments; i++ {
		startIndex := i * fragmentSize
		endIndex := startIndex + fragmentSize
		// Append the current fragment to the result
		result = append(result, slice[startIndex:endIndex])
	}

	// If there is a remainder, add the remaining elements as the last fragment
	if remainder != 0 {
		startIndex := numFragments * fragmentSize
		result = append(result, slice[startIndex:])
	}

	return result
}

// SplitSlice
// splitSlice splits the input slice into 'numParts' parts as evenly as possible.
func SplitSlice2(slice []string, numParts int) [][]string {
	if numParts <= 0 {
		return nil // Return nil if the number of parts is not valid
	}
	totalElements := len(slice)
	partSize := totalElements / numParts
	remainder := totalElements % numParts

	// Initialize the result 2D slice
	var result [][]string
	startIndex := 0

	for i := 0; i < numParts; i++ {
		endIndex := startIndex + partSize
		// If there is a remainder, distribute it across the parts
		if i < remainder {
			endIndex++
		}
		// Append the current part to the result
		result = append(result, slice[startIndex:endIndex])
		// Update the startIndex for the next part
		startIndex = endIndex
	}
	return result
}

func GetSplitKey(db *gorm.DB, tablename string) string {
	var cmm []map[string]interface{}
	err := db.Raw(fmt.Sprintf("show create table %s", tablename)).Scan(&cmm).Error
	if err != nil {
		util.Loggrs.Error(err.Error())
		return ""
	}
	var CreateSQL string
	for _, m := range cmm {
		CreateSQL = fmt.Sprintf("%v", m["Create Table"])
	}

	//重新组合建表语句1
	var splitKey, skey string
	for _, s := range strings.Split(CreateSQL, "\n") {
		//获取分桶切割键
		if strings.Contains(s, "DISTRIBUTED BY") {
			matches := regexp.MustCompile(`\((.*?)\)`).FindAllStringSubmatch(s, -1)
			skey = matches[0][0]
			break
		}
	}
	if util.P.SplitKey != "" {
		splitKey = fmt.Sprintf("(%s)", util.P.SplitKey)
	} else {
		splitKey = skey
	}
	return splitKey
}
func GetTabletNum(tgr *gorm.DB, tablename string) int {
	c := color.New()
	//获取tablet分布百分比
	var tablet []map[string]interface{}
	r := tgr.Raw(fmt.Sprintf("ADMIN SHOW REPLICA DISTRIBUTION FROM %s", tablename)).Scan(&tablet)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
		return -1
	}
	for i, item := range tablet {
		c := color.New()
		msg := fmt.Sprintf("%-2d %-10s %-10s %-7s %-7s", i, item["BackendId"].(string), item["ReplicaNum"].(string), item["Graph"].(string), item["Percent"].(string))
		util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > "), c.Add(color.FgHiCyan).Sprint(msg))
	}

	var replica []map[string]interface{}
	//获取tablet分布详细情况
	r = tgr.Raw(fmt.Sprintf("ADMIN SHOW REPLICA STATUS FROM %s", tablename)).Scan(&replica)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
		return -1
	}
	util.Loggrs.Info(c.Add(color.FgHiWhite).Sprint("TOP:JOB > Tablet:"), c.Add(color.FgHiCyan).Sprint(len(replica)))
	return len(replica)
}

// Exec 需要执行的命令，日志存放文件，超时时间
func Exec(prog string, timeOut int) error {
	var (
		cmd *exec.Cmd
		f   *os.File
		err error
	)
	if timeOut == 0 {
		cmd = exec.Command("bash", "-c", prog)
	} else {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
		defer cancelFunc()
		cmd = exec.CommandContext(ctx, "bash", "-c", prog)
	}

	if f, err = os.OpenFile(filepath.Base(os.Args[0])+"-bash.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, _ = io.WriteString(f, fmt.Sprintf("Run: %s\nOutput_____________\n", prog))

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(f, os.Stdout)
	stderr := io.MultiWriter(f, os.Stderr)
	err = cmd.Start()
	if err != nil {
		errMsg := fmt.Sprintf("cmd.Start() failed with %s ", err)
		_, _ = io.WriteString(stdout, errMsg+"\n")
		return fmt.Errorf(errMsg)
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	err = cmd.Wait()
	if err != nil {
		errMsg := fmt.Sprintf("cmd.Run() failed with %s ", err)
		_, _ = io.WriteString(stderr, errMsg+"\n")
		return fmt.Errorf("%v", errMsg)
	}
	if errStdout != nil || errStderr != nil {
		errMsg := fmt.Sprintf("failed to capture stdout or stderr, errStdout:%s errStderr:%s", errStdout, errStderr)
		_, _ = io.WriteString(stderr, errMsg+"\n")
		return fmt.Errorf("%v", errMsg)
	}
	return nil
}
