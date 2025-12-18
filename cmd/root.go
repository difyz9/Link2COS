package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
