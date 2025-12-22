package util

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ValidateConnectionParams 验证连接参数的格式和有效性
func ValidateConnectionParams(cfg SrAvgs) error {
	// 验证主机地址不为空
	if strings.TrimSpace(cfg.Host) == "" {
		return errors.New("主机地址不能为空")
	}

	// 验证端口号范围（1-65535）
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内，当前值: %d", cfg.Port)
	}

	// 简单的主机地址格式验证（IP地址或域名）
	if err := validateHost(cfg.Host); err != nil {
		return err
	}

	// 用户名和密码允许为空
	return nil
}

// validateHost 验证主机地址格式
func validateHost(host string) error {
	// 尝试解析为IP地址
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// 如果不是IP地址，检查是否为有效域名
	// 简单验证：域名不应包含非法字符
	if strings.Contains(host, " ") {
		return errors.New("主机地址格式无效：不能包含空格")
	}

	// localhost 或其他域名格式
	if len(host) > 0 && len(host) <= 253 {
		return nil
	}

	return errors.New("主机地址格式无效")
}

// ParsePort 解析端口号字符串
func ParsePort(portStr string) (int, error) {
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		return 0, errors.New("端口号不能为空")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, errors.New("端口号必须为有效数字")
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口号必须在1-65535范围内")
	}

	return port, nil
}
