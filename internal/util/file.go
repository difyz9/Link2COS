package util

import (
	"bufio"
	"os"
	"strings"
)

// ReadLinksFromFile 从文件中读取链接，并跳过空行和注释
func ReadLinksFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和以 # 开头的注释行
		if line != "" && !strings.HasPrefix(line, "#") {
			links = append(links, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
