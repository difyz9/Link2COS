package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Result 下载结果
type Result struct {
	Link      string
	LocalPath string
	Size      int64
	Error     error
}

// Downloader 文件下载器
type Downloader struct {
	httpClient *http.Client
	outputDir  string
}

// NewDownloader 创建下载器
func NewDownloader(httpClient *http.Client, outputDir string) *Downloader {
	return &Downloader{
		httpClient: httpClient,
		outputDir:  outputDir,
	}
}

// DownloadFile 下载单个文件到本地
func (d *Downloader) DownloadFile(link string) (*Result, error) {
	result := &Result{Link: link}

	// 先获取文件大小
	fileSize, err := d.getRemoteFileSize(link)
	if err != nil {
		result.Error = fmt.Errorf("获取文件大小失败: %w", err)
		return result, result.Error
	}
	result.Size = fileSize

	fmt.Printf("  文件大小: %.2f MB\n", float64(fileSize)/(1024*1024))

	// 下载文件
	resp, err := d.httpClient.Get(link)
	if err != nil {
		result.Error = fmt.Errorf("下载失败: %w", err)
		return result, result.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
		return result, result.Error
	}

	// 确定本地保存路径
	localPath, err := d.getLocalPath(link)
	if err != nil {
		result.Error = fmt.Errorf("确定本地路径失败: %w", err)
		return result, result.Error
	}
	result.LocalPath = localPath

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		result.Error = fmt.Errorf("创建目录失败: %w", err)
		return result, result.Error
	}

	// 保存到本地文件
	file, err := os.Create(localPath)
	if err != nil {
		result.Error = fmt.Errorf("创建本地文件失败: %w", err)
		return result, result.Error
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("保存文件失败: %w", err)
		return result, result.Error
	}

	fmt.Printf("  保存路径: %s\n", localPath)
	return result, nil
}

// DownloadForUpload 下载文件用于上传（返回Reader和大小）
func (d *Downloader) DownloadForUpload(link string) (io.ReadCloser, int64, error) {
	// 先获取文件大小
	fileSize, err := d.getRemoteFileSize(link)
	if err != nil {
		return nil, 0, fmt.Errorf("获取文件大小失败: %w", err)
	}

	fmt.Printf("  文件大小: %.2f MB\n", float64(fileSize)/(1024*1024))

	// 下载文件
	resp, err := d.httpClient.Get(link)
	if err != nil {
		return nil, 0, fmt.Errorf("下载失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	return resp.Body, fileSize, nil
}

// getRemoteFileSize 获取远程文件大小
func (d *Downloader) getRemoteFileSize(url string) (int64, error) {
	resp, err := d.httpClient.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

// getLocalPath 根据URL确定本地保存路径
func (d *Downloader) getLocalPath(link string) (string, error) {
	// 从URL中提取文件名
	parts := strings.Split(link, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("无效的URL")
	}

	filename := parts[len(parts)-1]
	if filename == "" {
		filename = "downloaded_file"
	}

	// 如果有输出目录，使用输出目录
	if d.outputDir != "" {
		return filepath.Join(d.outputDir, filename), nil
	}

	// 否则使用当前目录
	return filename, nil
}
