package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/difyz9/Link2COS/config"
	"github.com/spf13/cobra"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	inputFile  string
	configFile string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "link2cos",
	Short: "下载链接文件并上传到腾讯云COS",
	Long:  `从输入文件中读取链接，下载文件并上传到腾讯云COS存储桶。`,
}

// Execute 执行命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 添加 sync 子命令（从链接下载并上传的功能）
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "从链接下载文件并上传到COS",
		Long:  `从输入文件中读取链接，下载文件并上传到腾讯云COS存储桶。`,
		RunE:  runUpload,
	}
	syncCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入文件路径（必填）")
	syncCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "配置文件路径（默认: config.yaml）")
	syncCmd.MarkFlagRequired("input")
	
	rootCmd.AddCommand(syncCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化COS客户端
	cosClient, err := initCOSClient(cfg)
	if err != nil {
		return fmt.Errorf("初始化COS客户端失败: %w", err)
	}

	// 读取输入文件中的链接
	links, err := readLinksFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	fmt.Printf("共找到 %d 个链接\n", len(links))

	// 处理每个链接
	successCount := 0
	failCount := 0
	for i, link := range links {
		fmt.Printf("[%d/%d] 处理: %s\n", i+1, len(links), link)
		
		if err := processLink(cosClient, cfg, link); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ 失败: %v\n", err)
			failCount++
		} else {
			fmt.Println("  ✓ 成功")
			successCount++
		}
	}

	fmt.Printf("\n完成: 成功 %d, 失败 %d\n", successCount, failCount)
	return nil
}

// initCOSClient 初始化COS客户端
func initCOSClient(cfg *config.Config) (*cos.Client, error) {
	u, err := url.Parse(cfg.COS.BucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析Bucket URL失败: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COS.SecretID,
			SecretKey: cfg.COS.SecretKey,
		},
	})

	return client, nil
}

// readLinksFromFile 从文件中读取链接
func readLinksFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释行
		if line != "" && !strings.HasPrefix(line, "#") {
			links = append(links, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

// processLink 处理单个链接：下载并上传到COS
func processLink(client *cos.Client, cfg *config.Config, link string) error {
	// 计算COS存储路径
	cosPath, err := getCOSPath(cfg.COS.URLPrefix, link)
	if err != nil {
		return err
	}

	// 先获取文件大小
	fileSize, err := getRemoteFileSize(link)
	if err != nil {
		return fmt.Errorf("获取文件大小失败: %w", err)
	}

	fmt.Printf("  文件大小: %.2f MB\n", float64(fileSize)/(1024*1024))

	// 下载文件
	resp, err := http.Get(link)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 使用统一的上传器
	uploader := NewUploader(client)
	if err := uploader.UploadFromReader(resp.Body, cosPath, fileSize); err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}

	return nil
}

// getCOSPath 根据URL前缀计算COS存储路径
func getCOSPath(prefix, link string) (string, error) {
	if !strings.HasPrefix(link, prefix) {
		return "", fmt.Errorf("链接不匹配配置的前缀")
	}

	// 移除前缀，得到相对路径
	relativePath := strings.TrimPrefix(link, prefix)
	return relativePath, nil
}

// getRemoteFileSize 获取远程文件大小
func getRemoteFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}
