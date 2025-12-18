package tracker

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// LinkTracker 用于跟踪已下载的链接
type LinkTracker struct {
	filePath string
	links    map[string]bool
	mu       sync.RWMutex
}

// NewLinkTracker 创建新的链接跟踪器
func NewLinkTracker(filePath string) (*LinkTracker, error) {
	tracker := &LinkTracker{
		filePath: filePath,
		links:    make(map[string]bool),
	}

	// 尝试从文件加载已有的链接记录
	if err := tracker.load(); err != nil {
		// 如果文件不存在，这不是错误
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载链接记录失败: %w", err)
		}
	}

	return tracker, nil
}

// load 从文件加载已下载的链接
func (lt *LinkTracker) load() error {
	file, err := os.Open(lt.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := scanner.Text()
		if link != "" {
			lt.links[link] = true
		}
	}

	return scanner.Err()
}

// IsDownloaded 检查链接是否已下载
func (lt *LinkTracker) IsDownloaded(link string) bool {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.links[link]
}

// MarkDownloaded 标记链接已下载
func (lt *LinkTracker) MarkDownloaded(link string) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	// 如果已经存在，不需要重复添加
	if lt.links[link] {
		return nil
	}

	// 添加到内存中
	lt.links[link] = true

	// 追加到文件
	file, err := os.OpenFile(lt.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开链接记录文件失败: %w", err)
	}
	defer file.Close()

	if _, err := fmt.Fprintln(file, link); err != nil {
		return fmt.Errorf("写入链接记录失败: %w", err)
	}

	return nil
}

// GetDownloadedCount 获取已下载链接数量
func (lt *LinkTracker) GetDownloadedCount() int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return len(lt.links)
}
