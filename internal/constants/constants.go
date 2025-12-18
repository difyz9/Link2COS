package constants

const (
	// DownloadedLinksFile 已下载链接记录文件名
	DownloadedLinksFile = ".link2cos_downloaded.txt"

	// DefaultOutputDir 默认下载输出目录
	DefaultOutputDir = "downloads"

	// DefaultConfigFile 默认配置文件路径
	DefaultConfigFile = "config.yaml"

	// SmallFileSizeThreshold 小文件阈值：100MB
	SmallFileSizeThreshold = 100 * 1024 * 1024

	// MultipartChunkSize 分块上传的块大小：10MB
	MultipartChunkSize = 10 * 1024 * 1024

	// MaxConcurrentUploads 并发上传的最大数量
	MaxConcurrentUploads = 5
)
