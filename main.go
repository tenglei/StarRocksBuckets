package main

import (
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
	fmt.Println(c.Add(color.FgHiCyan).Sprint("\n📋 程序启动模式：交互式输入集群连接信息"))
	fmt.Println(c.Add(color.FgHiWhite).Sprint("请按照提示输入StarRocks集群的连接参数..."))
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
	fmt.Println(c.Add(color.FgHiYellow).Sprint("🔌 正在尝试连接StarRocks集群..."))
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
	fmt.Println(c.Add(color.FgHiGreen).Sprint("✅ 集群信息已保存，后续操作将使用此连接配置"))
	fmt.Println()
	
	// 保存全局配置供后续使用
	util.SrConfig = cfg
	
	// ==========================================
	// 步骤2: 解析命令行参数（业务操作参数）
	// ==========================================
	util.Parm()
	
	// ==========================================
	// 步骤3: 执行业务逻辑
	// ==========================================
	service.Run()
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
