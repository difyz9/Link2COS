package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/difyz9/Link2COS/config"
	"github.com/difyz9/Link2COS/internal/constants"
	"github.com/difyz9/Link2COS/internal/cos"
	"github.com/difyz9/Link2COS/internal/download"
	"github.com/difyz9/Link2COS/internal/tracker"
	"github.com/difyz9/Link2COS/internal/util"
	"github.com/spf13/cobra"
	cosSDK "github.com/tencentyun/cos-go-sdk-v5"
)

var (
	syncInputFile  string
	syncConfigFile string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "从链接下载文件并上传到COS",
	Long:  `从输入文件中读取链接，下载文件并上传到腾讯云COS存储桶。`,
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&syncInputFile, "input", "i", "", "输入文件路径（必填）")
	syncCmd.Flags().StringVarP(&syncConfigFile, "config", "c", constants.DefaultConfigFile, "配置文件路径（默认: config.yaml）")
	syncCmd.MarkFlagRequired("input")
}

func runSync(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig(syncConfigFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化COS客户端
	cosClient, err := cos.InitClient(cfg)
	if err != nil {
		return fmt.Errorf("初始化COS客户端失败: %w", err)
	}

	// 初始化链接跟踪器
	linkTracker, err := tracker.NewLinkTracker(constants.DownloadedLinksFile)
	if err != nil {
		return fmt.Errorf("初始化链接跟踪器失败: %w", err)
	}
	fmt.Printf("已下载链接数: %d\n", linkTracker.GetDownloadedCount())

	// 读取输入文件中的链接
	links, err := util.ReadLinksFromFile(syncInputFile)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	fmt.Printf("共找到 %d 个链接\n", len(links))

	// 处理每个链接
	successCount := 0
	failCount := 0
	skipCount := 0
	for i, link := range links {
		fmt.Printf("[%d/%d] 处理: %s\n", i+1, len(links), link)

		// 检查链接是否已下载
		if linkTracker.IsDownloaded(link) {
			fmt.Println("  ⊘ 跳过（已下载）")
			skipCount++
			continue
		}

		if err := processLink(cosClient, cfg, link, linkTracker); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ 失败: %v\n", err)
			failCount++
		} else {
			fmt.Println("  ✓ 成功")
			successCount++
		}
	}

	fmt.Printf("\n完成: 成功 %d, 失败 %d, 跳过 %d\n", successCount, failCount, skipCount)
	return nil
}

// processLink 处理单个链接：下载并上传到COS
func processLink(client *cosSDK.Client, cfg *config.Config, link string, linkTracker *tracker.LinkTracker) error {
	// 计算COS存储路径
	cosPath, err := getCOSPath(cfg.COS.URLPrefix, link)
	if err != nil {
		return err
	}

	// 创建HTTP客户端和下载器
	httpClient := download.CreateHTTPClient(cfg)
	downloader := download.NewDownloader(httpClient, "")

	// 下载文件（用于上传）
	reader, fileSize, err := downloader.DownloadForUpload(link)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 使用统一的上传器
	uploader := cos.NewUploader(client)
	if err := uploader.UploadFromReader(reader, cosPath, fileSize); err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}

	// 上传成功后，记录该链接
	if err := linkTracker.MarkDownloaded(link); err != nil {
		fmt.Fprintf(os.Stderr, "  警告: 记录链接失败: %v\n", err)
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
