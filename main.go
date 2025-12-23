package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"setbuckets/conn"
	"setbuckets/permit"
	"setbuckets/service"
	"setbuckets/util"
	"strings"
)

func main() {
	// 显示启动横幅
	printStartupBanner()
	
	// ==========================================
	// 步骤1: 交互式采集StarRocks连接参数
	// ==========================================
	c := color.New()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("\n📋 程序启动模式：交互式命令行界面"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("首先请输入StarRocks集群的连接信息..."))
	fmt.Println()
	
	cfg, err := util.CollectConnectionParams()
	if err != nil {
		fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 采集连接参数失败:"), err)
		return
	}
	
	// 验证连接参数
	if err := util.ValidateConnectionParams(cfg); err != nil {
		fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 参数验证失败:"), err)
		return
	}
	
	// 测试连接
	fmt.Println(c.Add(color.FgHiYellow).Sprint("🔌 正在测试连接StarRocks集群..."))
	db, err := conn.StarRocks(cfg)
	if err != nil {
		fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 连接StarRocks失败:"), err)
		fmt.Println(c.Add(color.FgHiYellow).Sprint("💡 请检查:"))
		fmt.Println("   - FE地址和端口是否正确")
		fmt.Println("   - 网络是否可达")
		fmt.Println("   - 用户名和密码是否正确")
		return
	}
	
	// 关闭测试连接
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	
	fmt.Println(c.Add(color.FgHiGreen).Sprint("✅ 连接成功！"))
	fmt.Println(c.Add(color.FgHiGreen).Sprint("✅ 集群连接配置已保存，可以开始执行命令"))
	
	// 保存全局配置供后续使用
	util.SrConfig = cfg
	
	// 初始化日志
	util.Logrus()
	
	// 初始化permit包的数据库连接
	if err := permit.InitConnections(); err != nil {
		fmt.Println(c.Add(color.FgHiYellow).Sprint("⚠️  警告: permit模块初始化失败:"), err)
		fmt.Println(c.Add(color.FgHiYellow).Sprint("⚠️  部分功能可能不可用，但程序将继续运行"))
	}
	
	// ==========================================
	// 步骤2: 进入交互式命令行循环
	// ==========================================
	startInteractiveCLI()
}

// startInteractiveCLI 启动交互式命令行界面
func startInteractiveCLI() {
	c := color.New()
	reader := bufio.NewReader(os.Stdin)
	
	// 显示帮助信息
	util.PrintCommandHelp()
	
	for {
		// 重置参数
		util.ResetParams()
		
		// 显示命令提示符
		fmt.Print(c.Add(color.FgHiCyan).Sprint("\n> "))
		
		// 读取用户输入
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 读取输入失败:"), err)
			continue
		}
		
		// 去除首尾空白
		input = strings.TrimSpace(input)
		
		// 处理空输入
		if input == "" {
			continue
		}
		
		// 处理退出命令
		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println(c.Add(color.FgHiYellow).Sprint("👋 感谢使用，再见！"))
			os.Exit(0)
		}
		
		// 处理帮助命令
		if input == "help" || input == "h" || input == "?" {
			util.PrintCommandHelp()
			continue
		}
		
		// 解析并执行命令
		args := util.ParseCommandLine(input)
		if len(args) == 0 {
			continue
		}
		
		// 解析参数
		if !util.ParseArgs(args) {
			continue
		}
		
		// 验证参数
		if !util.ValidateParams() {
			continue
		}
		
		// 执行业务逻辑
		fmt.Println(c.Add(color.FgHiGreen).Sprint("\n⚙️  正在执行命令..."))
		fmt.Println()
		
		// 执行service.Run()
		service.Run()
		
		fmt.Println()
		fmt.Println(c.Add(color.FgHiGreen).Sprint("✅ 命令执行完成"))
	}
}

// printStartupBanner 显示启动横幅
func printStartupBanner() {
	c := color.New()
	s := c.Add(color.FgHiGreen).Sprint(`
            ____  _             ____            _                      
           / ___|| |_ __ _ _ __|  _ \ ___   ___| | _____               
           \___ \| __/ __ | ___| |_) / _ \ / __| |/ / __|              
            ___) | || (_| | |  |  _ < (_) | (__|   <\__ \              
           |____/ \__\____|_|  |_| \_\___/ \___|_|\_\___/              `)
	v := fmt.Sprintf("%s %s %s\n%s %s %s\n%s %s %s\n%s %s %s\n%s %s %s",
		strings.Repeat("\t", 6),
		c.Add(color.FgHiMagenta).Sprint(filepath.Base(os.Args[0])),
		c.Add(color.FgHiYellow).Sprint(getVersion()),
		strings.Repeat("\t", 6),
		c.Add(color.FgHiWhite).Sprint("md5sum"),
		c.Add(color.FgHiYellow).Sprint("@"+getHash()),
		strings.Repeat("\t", 6),
		c.Add(color.FgHiWhite).Sprint("author"),
		c.Add(color.FgHiYellow).Sprint("@chengken li"),
		strings.Repeat("\t", 6),
		c.Add(color.FgHiWhite).Sprint("system"),
		c.Add(color.FgHiYellow).Sprint("@"+getUser()),
		strings.Repeat("\t", 6),
		c.Add(color.FgHiWhite).Sprint("update"),
		c.Add(color.FgHiYellow).Sprint("@"+getStat()),
	)
	fmt.Println(fmt.Sprintf("%s\n\n%s\n", s, v))
}

func getHash() string {
	var md5Sum string
	if v, err := os.Executable(); err == nil {
		file, _ := os.Open(v)
		defer file.Close()
		hash := md5.New()
		if _, err := io.Copy(hash, file); err == nil {
			md5Sum = hex.EncodeToString(hash.Sum(nil))
		}
	}
	return md5Sum
}

func getStat() string {
	var ts string
	if v, err := os.Executable(); err == nil {
		s, _ := os.Stat(v)
		ts = s.ModTime().Format("2006-01-02 15:04:05")
	}
	return ts
}

func getUser() string {
	u, _ := user.Current()
	return u.Username
}

func getVersion() string {
	file := "/tmp/.v." + filepath.Base(os.Args[0])
	v, _ := ioutil.ReadFile(file)
	if v != nil {
		return "v" + strings.NewReplacer("\n", "").Replace(string(v))
	}
	return ""
}
