package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/difyz9/Link2COS/config"
	"github.com/difyz9/Link2COS/internal/constants"
	"github.com/difyz9/Link2COS/internal/cos"
	"github.com/spf13/cobra"
)

var (
	localFile   string
	remotePath  string
	uploadConfig string
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "上传本地文件到腾讯云COS",
	Long:  `直接上传本地文件到腾讯云COS存储桶，可选择指定COS路径。`,
	RunE:  runLocalUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&localFile, "file", "f", "", "本地文件路径（必填）")
	uploadCmd.Flags().StringVarP(&remotePath, "path", "p", "", "COS存储路径（可选，默认使用文件名）")
	uploadCmd.Flags().StringVarP(&uploadConfig, "config", "c", constants.DefaultConfigFile, "配置文件路径（默认: config.yaml）")
	uploadCmd.MarkFlagRequired("file")
}

func runLocalUpload(cmd *cobra.Command, args []string) error {
	// 检查本地文件是否存在
	if _, err := os.Stat(localFile); os.IsNotExist(err) {
		return fmt.Errorf("本地文件不存在: %s", localFile)
	}

	// 加载配置
	cfg, err := config.LoadConfig(uploadConfig)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化COS客户端
	cosClient, err := cos.InitClient(cfg)
	if err != nil {
		return fmt.Errorf("初始化COS客户端失败: %w", err)
	}

	// 确定COS路径
	cosPath := remotePath
	if cosPath == "" {
		// 如果没有指定路径，使用文件名
		cosPath = filepath.Base(localFile)
		fmt.Printf("未指定COS路径，使用文件名: %s\n", cosPath)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	fmt.Printf("本地文件: %s\n", localFile)
	fmt.Printf("文件大小: %.2f MB\n", float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("COS路径: %s\n", cosPath)

	// 使用统一的上传器
	uploader := cos.NewUploader(cosClient)
	if err := uploader.UploadFile(localFile, cosPath); err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}

	fmt.Println("✓ 上传成功")
	return nil
}
