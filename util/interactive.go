package util

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// CollectConnectionParams 交互式采集StarRocks连接参数
// 返回：连接配置结构体和可能的错误
func CollectConnectionParams() (SrAvgs, error) {
	c := color.New()
	
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("    StarRocks 集群连接配置"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var cfg SrAvgs

	// 输入FE地址
	fmt.Printf("请输入FE地址 [默认: %s]: ", c.Add(color.FgHiYellow).Sprint("localhost"))
	host, err := reader.ReadString('\n')
	if err != nil {
		return cfg, fmt.Errorf("读取FE地址失败: %v", err)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	cfg.Host = host

	// 输入FE端口
	for {
		fmt.Printf("请输入FE端口 [默认: %s]: ", c.Add(color.FgHiYellow).Sprint("9030"))
		portStr, err := reader.ReadString('\n')
		if err != nil {
			return cfg, fmt.Errorf("读取FE端口失败: %v", err)
		}
		portStr = strings.TrimSpace(portStr)
		if portStr == "" {
			cfg.Port = 9030
			break
		}
		
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 端口号必须为有效数字，请重新输入"))
			continue
		}
		if port < 1 || port > 65535 {
			fmt.Println(c.Add(color.FgHiRed).Sprint("❌ 端口号必须在1-65535范围内，请重新输入"))
			continue
		}
		cfg.Port = port
		break
	}

	// 输入用户名
	fmt.Printf("请输入用户名 [默认: %s，直接回车使用默认]: ", c.Add(color.FgHiYellow).Sprint("root"))
	username, err := reader.ReadString('\n')
	if err != nil {
		return cfg, fmt.Errorf("读取用户名失败: %v", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		username = "root"
	}
	cfg.User = username

	// 输入密码（隐藏输入）
	password, err := ReadPassword("请输入密码（输入时不显示）: ")
	if err != nil {
		return cfg, fmt.Errorf("读取密码失败: %v", err)
	}
	cfg.Pass = password

	// 显示配置确认
	fmt.Println()
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("    配置确认"))
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Printf("FE地址: %s\n", c.Add(color.FgHiWhite).Sprint(cfg.Host))
	fmt.Printf("FE端口: %s\n", c.Add(color.FgHiWhite).Sprint(cfg.Port))
	fmt.Printf("用户名: %s\n", c.Add(color.FgHiWhite).Sprint(cfg.User))
	if cfg.Pass != "" {
		fmt.Printf("密  码: %s\n", c.Add(color.FgHiWhite).Sprint("******"))
	} else {
		fmt.Printf("密  码: %s\n", c.Add(color.FgHiWhite).Sprint("(空)"))
	}
	fmt.Println(c.Add(color.FgHiCyan).Sprint("========================================"))
	fmt.Println()

	return cfg, nil
}

// ReadPassword 安全读取密码（不显示明文）
// 参数：prompt 提示文本
// 返回：密码明文和可能的错误
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// 检查是否为TTY终端
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		// 非TTY环境，降级为普通输入并警告
		c := color.New()
		fmt.Println()
		fmt.Println(c.Add(color.FgHiYellow).Sprint("⚠️  警告：非交互式终端环境，密码将以明文显示"))
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	// TTY环境，使用隐藏输入
	passwordBytes, err := term.ReadPassword(fd)
	fmt.Println() // 输入完成后换行
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}
