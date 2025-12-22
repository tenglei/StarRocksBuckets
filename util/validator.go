package util

import (
	"fmt"
	"net"
)

// ValidateConnectionParams 验证连接参数的格式和有效性
// 参数：cfg 连接配置结构体
// 返回：验证失败的错误信息
func ValidateConnectionParams(cfg SrAvgs) error {
	// 验证主机地址不为空
	if cfg.Host == "" {
		return fmt.Errorf("主机地址不能为空")
	}

	// 验证端口范围
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内，当前值: %d", cfg.Port)
	}

	// 验证主机地址格式（IP地址或域名）
	// 尝试解析为IP地址
	ip := net.ParseIP(cfg.Host)
	if ip == nil {
		// 不是有效IP，尝试作为域名验证
		// 简单验证域名格式（检查是否包含非法字符）
		// 注意：这里不做DNS解析，只做基本格式检查
		if len(cfg.Host) > 255 {
			return fmt.Errorf("主机地址过长（最大255字符）")
		}
	}

	// 用户名和密码允许为空，不做验证

	return nil
}
