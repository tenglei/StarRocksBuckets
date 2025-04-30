package service

import (
	"fmt"
	"github.com/fatih/color"
	"gorm.io/gorm"
	"regexp"
	"setbuckets/util"
	"strings"
	"sync"
)

type scanSchema struct {
	Connect   *gorm.DB
	Table     string
	SchemaSql []string
	Model     string
	ModelSql  string
}

// ScanSchemaIsPrimary 非明细模型表，自定义分桶键或排序键时，重新排序字段顺序
func ScanSchemaIsPrimary(schema *scanSchema) string {
	//获取字段和注释,从 information_schema.columns 里面拿
	tb := strings.Split(schema.Table, ".")
	var cm []map[string]interface{}
	r := schema.Connect.Raw(fmt.Sprintf("select COLUMN_NAME,COLUMN_TYPE,IS_NULLABLE,COLUMN_COMMENT from information_schema.columns where TABLE_SCHEMA='%s' and TABLE_NAME='%s' order by ORDINAL_POSITION", tb[0], tb[1])).Scan(&cm)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return ""
	}
	var fieldlist []string
	for _, m := range cm {
		var isnil string
		if m["IS_NULLABLE"].(string) == "NO" {
			isnil = "NOT NULL"
		} else {
			isnil = "NULL"
		}
		fieldlist = append(fieldlist, fmt.Sprintf("  `%s` %s %s COMMENT \"%s\"", m["COLUMN_NAME"].(string), m["COLUMN_TYPE"].(string), isnil, m["COLUMN_COMMENT"].(string)))
	}
	//判断表是分区表还是全量表
	var rangerKey string
	for _, s := range schema.SchemaSql {
		for _, s2 := range strings.Split(s, "\n") {
			if strings.Contains(s2, "PARTITION BY RANGE") {
				re1 := regexp.MustCompile(`\((.*?)\)`)
				match := re1.FindAllStringSubmatch(s, -1)
				rangerKey = strings.NewReplacer("(", "", ")", "", "`", "").Replace(match[0][0])
			}
		}
	}
	var splitkey []string
	if rangerKey != "" {
		splitkey = append(splitkey, rangerKey)
	}
	splitkey = append(splitkey, strings.Split(util.P.SortKey, ",")...)
	//重新组合建表语句2，只有当自定义切割键时，非明细模型表，需要重置字段顺序
	var newColumns []string
	var oldSortkey, newSortkey string
	//重置排序键
	var sk []string
	for _, s2 := range splitkey {
		sk = append(sk, fmt.Sprintf("`%s`", s2))
	}
	re1 := regexp.MustCompile(`\((.*?)\)`)
	matches2 := re1.FindAllStringSubmatch(schema.ModelSql, -1)
	oldSortkey = matches2[0][0]
	newSortkey = fmt.Sprintf("(%s)", strings.Join(sk, ","))
	//重置字段顺序
	cs := color.New()
	util.Loggrs.Info(cs.Add(color.FgHiYellow).Sprint("重置字段顺序"))
	var columns, rcolumns []string
label:
	for _, s2 := range fieldlist {
		for _, s := range splitkey {
			if strings.Contains(s2, fmt.Sprintf("`%s`", s)) {
				continue label
			}
		}
		rcolumns = append(rcolumns, s2)
	}
	//新的
label2:
	for _, s := range splitkey {
		for _, s2 := range fieldlist {
			if strings.Contains(s2, fmt.Sprintf("`%s`", s)) {
				if !strings.Contains(s2, "NOT NULL") {
					s2 = strings.NewReplacer("NULL", "NOT NULL").Replace(s2)
				}
				columns = append(columns, s2)
				continue label2
			}
		}
	}
	// 创建一个新的切片，长度为A和B长度之和
	newColumns = make([]string, len(columns)+len(rcolumns))
	// 将A的元素复制到新切片的前面
	copy(newColumns, columns)
	// 将B的元素复制到新切片的后面
	copy(newColumns[len(columns):], rcolumns)
	//获取旧表的字段数量信息
	var mn []map[string]interface{}
	r = schema.Connect.Raw("desc " + schema.Table).Scan(&mn)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return ""
	}
	if len(mn) != len(newColumns) {
		util.Loggrs.Warn(fmt.Sprintf("排序键重置后与原字段数不一致 原字段数：%d，新排序后字段数：%d，熔断！", len(mn), len(newColumns)))
		util.Loggrs.Warn("旧的:\n" + strings.Join(fieldlist, "\n"))
		util.Loggrs.Warn("新的:\n" + strings.Join(newColumns, "\n"))
		return ""
	}
	util.Loggrs.Info(fmt.Sprintf("排序键变更: %s -> %s", cs.Add(color.FgHiYellow).Sprint(oldSortkey), cs.Add(color.FgHiGreen).Sprint(newSortkey)))
	NewSQL := strings.NewReplacer(oldSortkey, newSortkey).Replace(strings.Join(schema.SchemaSql, "\n"))

	var once sync.Once
	var stmt []string
	for _, s := range strings.Split(NewSQL, "\n") {
		if regexp.MustCompile(`NULL\s*COMMENT|(COMMENT\s*")`).MatchString(s) {
			once.Do(func() {
				stmt = append(stmt, "STARROCKS_OLAP_FIELD")
			})
			continue
		}
		stmt = append(stmt, s)
	}
	NewSQL = strings.NewReplacer("STARROCKS_OLAP_FIELD", strings.Join(newColumns, ",\n")).Replace(strings.Join(stmt, "\n"))
	return NewSQL
}
