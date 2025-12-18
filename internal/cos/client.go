package cos

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/difyz9/Link2COS/config"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// InitClient 初始化COS客户端（上传到国内COS不需要代理）
func InitClient(cfg *config.Config) (*cos.Client, error) {
	u, err := url.Parse(cfg.COS.BucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析Bucket URL失败: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}

	// 上传到腾讯云 COS 不使用代理，直连更快
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COS.SecretID,
			SecretKey: cfg.COS.SecretKey,
		},
		Timeout: 300 * time.Second,
	})

	return client, nil
}
