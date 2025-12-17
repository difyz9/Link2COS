package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置文件结构
type Config struct {
	COS COSConfig `yaml:"cos"`
}

// COSConfig 腾讯云COS配置
type COSConfig struct {
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	BucketName string `yaml:"bucket_name"` // 例如: examplebucket-1250000000
	Region     string `yaml:"region"`      // 例如: ap-guangzhou
	BucketURL  string `yaml:"-"`           // 自动拼接: https://bucketname.cos.region.myqcloud.com
	URLPrefix  string `yaml:"url_prefix"`  // 例如: https://huggingface.co/Comfy-Org/Wan_2.2_ComfyUI_Repackaged/resolve/main/
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证必填字段
	if config.COS.SecretID == "" || config.COS.SecretKey == "" {
		return nil, fmt.Errorf("配置文件缺少必填字段: secret_id 或 secret_key")
	}
	if config.COS.BucketName == "" {
		return nil, fmt.Errorf("配置文件缺少必填字段: bucket_name")
	}
	if config.COS.Region == "" {
		return nil, fmt.Errorf("配置文件缺少必填字段: region")
	}
	if config.COS.URLPrefix == "" {
		return nil, fmt.Errorf("配置文件缺少必填字段: url_prefix")
	}

	// 根据 BucketName 和 Region 拼接 BucketURL
	config.COS.BucketURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.COS.BucketName, config.COS.Region)

	return &config, nil
}
