.PHONY: build build-linux build-darwin build-windows clean all

# 默认编译当前平台
build:
	go build -o link2cos

# 编译 Linux AMD64 (Ubuntu)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o link2cos-linux-amd64

# 编译 macOS (Darwin)
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o link2cos-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o link2cos-darwin-arm64

# 编译 Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o link2cos-windows-amd64.exe

# 编译所有平台
all: build-linux build-darwin build-windows
	@echo "所有平台编译完成"

# 清理编译产物
clean:
	rm -f link2cos link2cos-*

# 显示帮助
help:
	@echo "可用的编译命令:"
	@echo "  make build         - 编译当前平台"
	@echo "  make build-linux   - 编译 Linux AMD64"
	@echo "  make build-darwin  - 编译 macOS (AMD64 + ARM64)"
	@echo "  make build-windows - 编译 Windows AMD64"
	@echo "  make all           - 编译所有平台"
	@echo "  make clean         - 清理编译产物"
