package download

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/difyz9/Link2COS/config"
)

// CreateHTTPClient 创建支持代理的 HTTP 客户端（用于下载文件）
func CreateHTTPClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// 如果配置了代理，设置代理（用于下载海外文件）
	if cfg.COS.Proxy != "" {
		proxy, err := url.Parse(cfg.COS.Proxy)
		if err != nil {
			fmt.Printf("警告: 代理URL解析失败: %v\n", err)
		} else {
			transport.Proxy = http.ProxyURL(proxy)
			fmt.Printf("✓ 下载使用代理: %s\n", cfg.COS.Proxy)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   300 * time.Second, // 5分钟超时
	}
}
