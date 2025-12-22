package util

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

// CollectConnectionParams 采集用户输入的所有连接参数
func CollectConnectionParams() (SrAvgs, error) {
	c := color.New()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
		fmt.Println(c.Add(color.FgHiCyan).Sprint("StarRocks 集群连接配置"))
		fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
		fmt.Println()

		// 获取FE地址
		fmt.Print(c.Add(color.FgHiWhite).Sprint("请输入FE地址 [默认: localhost]: "))
		hostInput, _ := reader.ReadString('\n')
		host := strings.TrimSpace(hostInput)
		if host == "" {
			host = "localhost"
		}

		// 获取FE端口
		fmt.Print(c.Add(color.FgHiWhite).Sprint("请输入FE端口 [默认: 9030]: "))
		portInput, _ := reader.ReadString('\n')
		portStr := strings.TrimSpace(portInput)
		if portStr == "" {
			portStr = "9030"
		}

		port, err := ParsePort(portStr)
		if err != nil {
			fmt.Println(c.Add(color.FgHiRed).Sprint("错误: "), err)
			fmt.Println()
			continue
		}

		// 获取用户名
		fmt.Print(c.Add(color.FgHiWhite).Sprint("请输入用户名 [默认: root，直接回车使用默认]: "))
		userInput, _ := reader.ReadString('\n')
		user := strings.TrimSpace(userInput)
		if user == "" {
			user = "root"
		}

		// 获取密码（隐藏输入）
		password, err := ReadPassword(c.Add(color.FgHiWhite).Sprint("请输入密码（输入时不显示）: "))
		if err != nil {
			fmt.Println(c.Add(color.FgHiRed).Sprint("\n错误: "), "读取密码失败: ", err)
			continue
		}

		// 构建配置
		cfg := SrAvgs{
			Host: host,
			Port: port,
			User: user,
			Pass: password,
		}

		// 验证参数
		if err := ValidateConnectionParams(cfg); err != nil {
			fmt.Println(c.Add(color.FgHiRed).Sprint("错误: "), err)
			fmt.Println()
			continue
		}

		// 显示配置确认
		fmt.Println()
		fmt.Println(c.Add(color.FgHiYellow).Sprint("========================================"))
		fmt.Println(c.Add(color.FgHiYellow).Sprint("配置确认"))
		fmt.Println(c.Add(color.FgHiYellow).Sprint("========================================"))
		fmt.Printf("FE地址: %s\n", c.Add(color.FgHiCyan).Sprint(cfg.Host))
		fmt.Printf("FE端口: %s\n", c.Add(color.FgHiCyan).Sprint(cfg.Port))
		fmt.Printf("用户名: %s\n", c.Add(color.FgHiCyan).Sprint(cfg.User))
		if cfg.Pass == "" {
			fmt.Printf("密码: %s\n", c.Add(color.FgHiGreen).Sprint("(空)"))
		} else {
			fmt.Printf("密码: %s\n", c.Add(color.FgHiGreen).Sprint("******"))
		}
		fmt.Println()

		// 请求确认
		fmt.Print(c.Add(color.FgHiWhite).Sprint("确认连接配置？(y/n) [默认: y]: "))
		confirmInput, _ := reader.ReadString('\n')
		confirm := strings.TrimSpace(strings.ToLower(confirmInput))
		if confirm == "" || confirm == "y" || confirm == "yes" {
			return cfg, nil
		}

		fmt.Println(c.Add(color.FgHiYellow).Sprint("重新输入配置..."))
	}
}

// ReadPassword 安全采集密码输入而不显示明文
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// 检查是否为TTY终端
	if !term.IsTerminal(int(syscall.Stdin)) {
		// 非TTY环境，降级为普通输入并警告
		fmt.Println()
		c := color.New()
		fmt.Println(c.Add(color.FgHiYellow).Sprint("警告: 非TTY终端环境，密码将以明文显示"))
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	// 使用term库读取密码
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println() // 密码输入后换行
	return string(passwordBytes), nil
}
