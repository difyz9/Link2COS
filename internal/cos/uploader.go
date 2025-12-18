package cos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/difyz9/Link2COS/internal/constants"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// Uploader 文件上传器
type Uploader struct {
	client *cos.Client
}

// NewUploader 创建上传器
func NewUploader(client *cos.Client) *Uploader {
	return &Uploader{client: client}
}

// UploadFile 上传本地文件到COS（自动选择策略）
func (u *Uploader) UploadFile(localFile, cosPath string) error {
	// 获取文件信息
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	fileSize := fileInfo.Size()

	// 根据文件大小选择上传策略
	if fileSize < constants.SmallFileSizeThreshold {
		fmt.Printf("  策略: 内存上传 (%.2f MB)\n", float64(fileSize)/(1024*1024))
		return u.uploadFromMemory(localFile, cosPath)
	} else {
		fmt.Printf("  策略: 分块上传 (%.2f MB)\n", float64(fileSize)/(1024*1024))
		return u.uploadMultipart(localFile, cosPath, fileSize)
	}
}

// UploadFromReader 从Reader上传到COS（用于下载的文件）
func (u *Uploader) UploadFromReader(reader io.Reader, cosPath string, size int64) error {
	if size < constants.SmallFileSizeThreshold {
		// 小文件：读取到内存后上传
		data, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}

		return u.uploadBytes(data, cosPath)
	} else {
		// 大文件：先保存到临时文件，再分块上传
		tmpFile, err := u.saveTempFile(reader)
		if err != nil {
			return fmt.Errorf("保存临时文件失败: %w", err)
		}
		defer os.Remove(tmpFile)

		return u.uploadMultipart(tmpFile, cosPath, size)
	}
}

// uploadFromMemory 小文件：读取到内存后上传
func (u *Uploader) uploadFromMemory(localFile, cosPath string) error {
	data, err := os.ReadFile(localFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	return u.uploadBytes(data, cosPath)
}

// uploadBytes 从字节数组上传
func (u *Uploader) uploadBytes(data []byte, cosPath string) error {
	reader := bytes.NewReader(data)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: int64(len(data)),
		},
	}

	_, err := u.client.Object.Put(context.Background(), cosPath, reader, opt)
	return err
}

// uploadMultipart 大文件：使用并发分块上传
func (u *Uploader) uploadMultipart(localFile, cosPath string, fileSize int64) error {
	// 初始化分块上传
	initRes, _, err := u.client.Object.InitiateMultipartUpload(context.Background(), cosPath, nil)
	if err != nil {
		return fmt.Errorf("初始化分块上传失败: %w", err)
	}

	uploadID := initRes.UploadID

	// 计算分块数量
	totalParts := int((fileSize + constants.MultipartChunkSize - 1) / constants.MultipartChunkSize)
	fmt.Printf("  总分块数: %d\n", totalParts)

	// 并发上传分块
	type partResult struct {
		partNumber int
		etag       string
		err        error
	}

	resultChan := make(chan partResult, totalParts)
	semaphore := make(chan struct{}, constants.MaxConcurrentUploads)
	var wg sync.WaitGroup

	// 上传所有分块
	for partNumber := 1; partNumber <= totalParts; partNumber++ {
		wg.Add(1)
		go func(pn int) {
			defer wg.Done()

			// 限制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 读取分块数据
			offset := int64(pn-1) * constants.MultipartChunkSize
			size := constants.MultipartChunkSize
			if offset+int64(size) > fileSize {
				size = int(fileSize - offset)
			}

			data := make([]byte, size)
			file, err := os.Open(localFile)
			if err != nil {
				resultChan <- partResult{partNumber: pn, err: err}
				return
			}
			defer file.Close()

			_, err = file.ReadAt(data, offset)
			if err != nil && err != io.EOF {
				resultChan <- partResult{partNumber: pn, err: err}
				return
			}

			// 上传分块
			etag, err := u.uploadPart(cosPath, uploadID, pn, data)
			resultChan <- partResult{partNumber: pn, etag: etag, err: err}

			if err == nil {
				fmt.Printf("  已上传: %d/%d 块 (%.1f%%)\n", pn, totalParts, float64(pn)*100/float64(totalParts))
			}
		}(partNumber)
	}

	// 等待所有上传完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	parts := make([]cos.Object, 0, totalParts)
	var uploadErr error
	for result := range resultChan {
		if result.err != nil {
			uploadErr = result.err
			break
		}
		parts = append(parts, cos.Object{
			PartNumber: result.partNumber,
			ETag:       result.etag,
		})
	}

	// 如果有错误，终止上传
	if uploadErr != nil {
		u.client.Object.AbortMultipartUpload(context.Background(), cosPath, uploadID)
		return fmt.Errorf("上传分块失败: %w", uploadErr)
	}

	// 按 PartNumber 排序
	// COS需要按顺序提交parts
	sortedParts := make([]cos.Object, totalParts)
	for _, part := range parts {
		sortedParts[part.PartNumber-1] = part
	}

	// 完成分块上传
	completeOpt := &cos.CompleteMultipartUploadOptions{
		Parts: sortedParts,
	}

	_, _, err = u.client.Object.CompleteMultipartUpload(context.Background(), cosPath, uploadID, completeOpt)
	if err != nil {
		u.client.Object.AbortMultipartUpload(context.Background(), cosPath, uploadID)
		return fmt.Errorf("完成分块上传失败: %w", err)
	}

	return nil
}

// uploadPart 上传单个分块
func (u *Uploader) uploadPart(cosPath, uploadID string, partNumber int, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	resp, err := u.client.Object.UploadPart(
		context.Background(),
		cosPath,
		uploadID,
		partNumber,
		reader,
		nil,
	)
	if err != nil {
		return "", err
	}

	return resp.Header.Get("ETag"), nil
}

// saveTempFile 保存到临时文件
func (u *Uploader) saveTempFile(reader io.Reader) (string, error) {
	tmpFile, err := os.CreateTemp("", "link2cos-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, reader)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}
