package util

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
)
func PrintCommandHelp() {
	c := color.New()
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("  StarRocks Buckets 分桶修复工具"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println()
	fmt.Println(c.Add(color.FgHiWhite).Sprint("📖 可用命令选项："))
	fmt.Println()
	fmt.Println(c.Add(color.FgHiYellow).Sprint("  基础参数："))
	fmt.Printf("    %-20s %s\n", "-t <SCHEMA.TABLE>", "指定要操作的表（必需）")
	fmt.Printf("    %-20s %s\n", "-h", "显示帮助信息")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiYellow).Sprint("  分桶操作："))
	fmt.Printf("    %-20s %s\n", "-b <数字>", "设置分桶数")
	fmt.Printf("    %-20s %s\n", "-a", "自动覆写分桶数")
	fmt.Printf("    %-20s %s\n", "-splitkey <键名>", "设置分桶键")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiYellow).Sprint("  分区操作："))
	fmt.Printf("    %-20s %s\n", "-c", "过滤空分区")
	fmt.Printf("    %-20s %s\n", "-pset", "分区级别重置分桶数")
	fmt.Printf("    %-20s %s\n", "-p <分区名>", "指定分区名称")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiYellow).Sprint("  查询分析："))
	fmt.Printf("    %-20s %s\n", "-l", "显示排序键")
	fmt.Printf("    %-20s %s\n", "-id <后端ID>", "指定后端节点ID")
	fmt.Printf("    %-20s %s\n", "-n <MB数>", "拆分Tablet大小")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiYellow).Sprint("  其他选项："))
	fmt.Printf("    %-20s %s\n", "-m <操作>", "指定操作类型（默认：create,insert,alter,drop）")
	fmt.Printf("    %-20s %s\n", "-sortkey <键名>", "设置排序键")
	fmt.Printf("    %-20s %s\n", "-thread <数量>", "设置线程数（默认：3）")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiGreen).Sprint("💡 使用示例："))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("  查看表信息:        -t tpcds.store_sales"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("  查看排序键:        -t tpcds.store_sales -l"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("  设置分桶数:        -t tpcds.store_sales -b 10"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("  自动调整分桶:      -t tpcds.store_sales -a"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("  分区级重置:        -t tpcds.store_sales -pset -b 5"))
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("🚪 特殊命令："))
	fmt.Printf("    %-20s %s\n", "help / h / ?", "显示此帮助信息")
	fmt.Printf("    %-20s %s\n", "exit / quit / q", "退出程序")
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
}

// ResetParams 重置参数为默认值
func ResetParams() {
	P.Table = ""
	P.Buckets = 0
	P.Help = false
	P.List = false
	P.Auto = false
	P.ClearPartition = false
	P.PartitionSet = false
	P.PartitionName = ""
	P.BackendId = ""
	P.SplitKey = ""
	P.SortKey = ""
	P.TabletSize = 0
	P.Action = "create,insert,alter,drop"
	P.Thread = 3
}

// ParseCommandLine 解析命令行字符串为参数数组
func ParseCommandLine(input string) []string {
	// 简单的空格分割解析
	// 处理引号内的空格
	var args []string
	var current strings.Builder
	inQuotes := false
	
	for i := 0; i < len(input); i++ {
		ch := input[i]
		
		if ch == '"' || ch == '\'' {
			inQuotes = !inQuotes
			continue
		}
		
		if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		
		current.WriteByte(ch)
	}
	
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	
	return args
}

// ParseArgs 手动解析参数
func ParseArgs(args []string) bool {
	c := color.New()
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		switch arg {
		case "-t":
			if i+1 < len(args) {
				P.Table = args[i+1]
				i++
			} else {
				fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 错误: -t 需要指定表名"))
				return false
			}
		case "-b":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &P.Buckets)
				i++
			} else {
				fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 错误: -b 需要指定分桶数"))
				return false
			}
		case "-h":
			PrintCommandHelp()
			return false
		case "-l":
			P.List = true
		case "-a":
			P.Auto = true
		case "-c":
			P.ClearPartition = true
		case "-pset":
			P.PartitionSet = true
		case "-p":
			if i+1 < len(args) {
				P.PartitionName = args[i+1]
				i++
			}
		case "-id":
			if i+1 < len(args) {
				P.BackendId = args[i+1]
				i++
			}
		case "-splitkey":
			if i+1 < len(args) {
				P.SplitKey = args[i+1]
				i++
			}
		case "-sortkey":
			if i+1 < len(args) {
				P.SortKey = args[i+1]
				i++
			}
		case "-n":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &P.TabletSize)
				i++
			}
		case "-m":
			if i+1 < len(args) {
				P.Action = args[i+1]
				i++
			}
		case "-thread":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &P.Thread)
				i++
			}
		default:
			fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 未知参数:"), arg)
			fmt.Println(c.Add(color.FgHiYellow).Sprint("💡 输入 'help' 查看可用命令"))
			return false
		}
	}
	
	return true
}

// ValidateParams 验证参数有效性
func ValidateParams() bool {
	c := color.New()
	
	// 检查是否指定了表名
	if len(P.Table) == 0 {
		fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 错误: 必须使用 -t 指定表名"))
		fmt.Println(c.Add(color.FgHiYellow).Sprint("💡 示例: -t tpcds.store_sales"))
		return false
	}
	
	// 检查表名格式
	if !strings.Contains(P.Table, ".") {
		fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 错误: 表名格式必须为 <数据库>.<表名>"))
		fmt.Println(c.Add(color.FgHiYellow).Sprint("💡 示例: -t tpcds.store_sales"))
		return false
	}
	
	return true
}
