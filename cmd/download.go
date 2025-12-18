package cmd

import (
	"fmt"
	"os"

	"github.com/difyz9/Link2COS/config"
	"github.com/difyz9/Link2COS/internal/constants"
	"github.com/difyz9/Link2COS/internal/download"
	"github.com/difyz9/Link2COS/internal/tracker"
	"github.com/difyz9/Link2COS/internal/util"
	"github.com/spf13/cobra"
)

var (
	downloadInputFile  string
	downloadConfigFile string
	downloadOutputDir  string
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "从链接下载文件到本地",
	Long:  `从输入文件中读取链接，下载文件到本地目录，支持链接去重。`,
	RunE:  runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&downloadInputFile, "input", "i", "", "输入文件路径（必填）")
	downloadCmd.Flags().StringVarP(&downloadOutputDir, "output", "o", constants.DefaultOutputDir, "下载文件保存目录（默认: downloads）")
	downloadCmd.Flags().StringVarP(&downloadConfigFile, "config", "c", constants.DefaultConfigFile, "配置文件路径（默认: config.yaml）")
	downloadCmd.MarkFlagRequired("input")
}

func runDownload(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig(downloadConfigFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化链接跟踪器
	linkTracker, err := tracker.NewLinkTracker(constants.DownloadedLinksFile)
	if err != nil {
		return fmt.Errorf("初始化链接跟踪器失败: %w", err)
	}
	fmt.Printf("已下载链接数: %d\n", linkTracker.GetDownloadedCount())

	// 创建HTTP客户端和下载器
	httpClient := download.CreateHTTPClient(cfg)
	downloader := download.NewDownloader(httpClient, downloadOutputDir)

	// 读取输入文件中的链接
	links, err := util.ReadLinksFromFile(downloadInputFile)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	fmt.Printf("共找到 %d 个链接\n", len(links))
	fmt.Printf("下载目录: %s\n", downloadOutputDir)

	// 确保输出目录存在
	if err := os.MkdirAll(downloadOutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 处理每个链接
	successCount := 0
	failCount := 0
	skipCount := 0

	for i, link := range links {
		fmt.Printf("[%d/%d] 下载: %s\n", i+1, len(links), link)

		// 检查链接是否已下载
		if linkTracker.IsDownloaded(link) {
			fmt.Println("  ⊘ 跳过（已下载）")
			skipCount++
			continue
		}

		// 下载文件
		_, err := downloader.DownloadFile(link)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ 失败: %v\n", err)
			failCount++
			continue
		}

		// 标记为已下载
		if err := linkTracker.MarkDownloaded(link); err != nil {
			fmt.Fprintf(os.Stderr, "  警告: 记录链接失败: %v\n", err)
		}

		fmt.Println("  ✓ 成功")
		successCount++
	}

	fmt.Printf("\n完成: 成功 %d, 失败 %d, 跳过 %d\n", successCount, failCount, skipCount)
	return nil
}
