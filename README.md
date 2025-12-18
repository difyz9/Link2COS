**Link2COS** 是一款基于 Go 语言开发的轻量级命令行工具，专注于高效、智能地将文件上传到腾讯云对象存储（COS）。支持从 URL 批量下载上传和本地文件直传，自动优化上传策略，完美解决跨境文件存储难题。

## 🎯 核心特性

### 🚀 智能上传策略
- **小文件（< 100MB）**：内存直传，零磁盘 I/O，速度更快
- **大文件（≥ 100MB）**：并发分块上传（10MB/块），支持断点续传
- **自动选择**：根据文件大小自动选择最优上传方式

### 📦 三种模式支持
- **`sync` 模式**：从 URL 批量下载并上传到 COS，支持路径映射和链接去重
- **`download` 模式**：纯下载模式，将链接文件下载到本地，支持链接去重
- **`upload` 模式**：直接上传本地文件，可指定 COS 存储路径

### 🔄 链接去重功能
- 自动记录已下载的链接到本地文件
- 重复运行时自动跳过已下载的链接
- 避免重复下载，节省时间和带宽

### ⚡ 性能优化
- 并发分块上传（最多 5 个分块同时上传）
- 实时进度显示，任务状态一目了然
- 上传失败自动清理，避免产生碎片

### 🛡️ 安全可靠
- 配置文件管理密钥，支持多环境切换
- 完整的错误处理和日志追踪
- 自动路径计算，保持原始目录结构

## 💼 适用场景

- 🌐 **跨境文件同步**：将 HuggingFace、GitHub 等海外资源同步到国内 COS
- 📊 **数据备份迁移**：批量迁移服务器日志、备份文件到云端
- 🔄 **CI/CD 集成**：自动化构建产物上传到对象存储
- 📦 **模型文件管理**：AI 模型文件、大数据集的云端存储
- 🔁 **断点续传场景**：利用链接去重功能，中断后重新运行自动跳过已下载文件
- 💾 **本地归档**：使用 download 命令批量下载远程资源到本地存储

## 📥 快速开始

### 安装

**方式一：从源码编译**

```bash
# 克隆仓库
git clone https://github.com/difyz9/Link2COS.git
cd Link2COS

# 编译
make build
```

**方式二：跨平台编译**

```bash
# Linux AMD64 (Ubuntu/Debian)
make build-linux

# macOS (Intel)
GOOS=darwin GOARCH=amd64 make build

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 make build

# Windows
make build-windows
```

### 快速体验

```bash
# 1. 创建配置文件
cat > config.yaml << EOF
cos:
  secret_id: "YOUR_SECRET_ID"
  secret_key: "YOUR_SECRET_KEY"
  bucket_name: "mybucket-1234567890"
  region: "ap-guangzhou"
  url_prefix: "https://example.com/files/"
EOF

# 2. 上传本地文件
./link2cos upload -f myfile.txt

# 3. 从 URL 批量下载上传
echo "https://example.com/files/data.bin" > links.txt
./link2cos sync -i links.txt

# 4. 批量下载到本地（不上传）
./link2cos download -i links.txt -o downloads
```

## ⚙️ 配置

创建 `config.yaml` 配置文件：

```yaml
cos:
  # 腾讯云访问密钥
  secret_id: "AKIDxxxxxxxxxxxxxxxx"
  secret_key: "xxxxxxxxxxxxxxxx"
  
  # COS 存储桶信息
  bucket_name: "mybucket-1234567890"
  region: "ap-guangzhou"
  
  # URL 前缀（用于 sync 命令）
  url_prefix: "https://huggingface.co/Comfy-Org/Wan_2.2_ComfyUI_Repackaged/resolve/main/"
```

### 配置项说明

| 配置项 | 必填 | 说明 | 示例值 |
|--------|------|------|--------|
| `secret_id` | ✅ | 腾讯云 SecretId（[获取方式](https://console.cloud.tencent.com/cam/capi)） | `AKID...` |
| `secret_key` | ✅ | 腾讯云 SecretKey | `xxxxx` |
| `bucket_name` | ✅ | 存储桶名称（格式：name-appid） | `mybucket-1234567890` |
| `region` | ✅ | COS 地域（[地域列表](https://cloud.tencent.com/document/product/436/6224)） | `ap-guangzhou` |
| `url_prefix` | ⚠️ | URL 前缀（仅 sync 命令需要） | `https://example.com/` |
| `proxy` | ⚠️ | 代理地址（下载时使用，可选） | `http://127.0.0.1:7890` |

**常用地域代码：**
- `ap-guangzhou`（广州）
- `ap-shanghai`（上海）
- `ap-beijing`（北京）
- `ap-chengdu`（成都）

> 💡 提示：`bucket_url` 会自动由 `bucket_name` 和 `region` 拼接，无需手动配置

## 📖 使用指南

### 命令总览

```bash
link2cos [command] [flags]

可用命令：
  sync        批量下载 URL 并上传到 COS（支持链接去重）
  download    批量下载 URL 到本地目录（支持链接去重）
  upload      上传本地文件到 COS
  help        查看帮助信息

全局参数：
  -c, --config    配置文件路径（默认：config.yaml）
  -h, --help      显示帮助信息
```

### 1. sync - 批量下载上传

从文件中读取 URL 列表，下载后上传到 COS（支持链接去重）：

```bash
# 使用默认配置文件
./link2cos sync -i links.txt

# 指定配置文件
./link2cos sync -i links.txt -c /path/to/config.yaml
```

**参数说明：**
- `-i, --input`：输入文件路径（必填）
- `-c, --config`：配置文件路径（可选，默认 `config.yaml`）

**链接去重机制：**
- 所有成功下载的链接会记录在 `.link2cos_downloaded.txt` 文件中
- 再次运行时，已下载的链接会自动跳过
- 输出统计会显示：成功、失败、跳过的数量

**输出示例：**
```
已下载链接数: 15
共找到 20 个链接
[1/20] 处理: https://example.com/file1.bin
  ⊘ 跳过（已下载）
[2/20] 处理: https://example.com/file2.bin
  文件大小: 256.50 MB
  策略: 分块上传
  ✓ 成功

完成: 成功 5, 失败 0, 跳过 15
```

**输入文件格式：**

```txt
# 注释行以 # 开头
https://example.com/files/file1.bin

# 支持空行
https://example.com/files/file2.bin
https://example.com/files/file3.bin
```

**路径映射规则：**

配置中的 `url_prefix` 会被自动移除，保留相对路径：

| URL 前缀配置 | 完整链接 | COS 存储路径 |
|-------------|---------|-------------|
| `https://huggingface.co/models/main/` | `https://huggingface.co/models/main/weights/model.safetensors` | `weights/model.safetensors` |
| `https://example.com/data/` | `https://example.com/data/2024/file.bin` | `2024/file.bin` |

### 2. download - 批量下载到本地

从文件中读取 URL 列表，下载到本地目录（支持链接去重）：

```bash
# 使用默认配置和输出目录
./link2cos download -i links.txt

# 指定输出目录
./link2cos download -i links.txt -o /path/to/output

# 指定配置文件
./link2cos download -i links.txt -o downloads -c /path/to/config.yaml
```

**参数说明：**
- `-i, --input`：输入文件路径（必填）
- `-o, --output`：下载文件保存目录（可选，默认 `downloads`）
- `-c, --config`：配置文件路径（可选，默认 `config.yaml`）

**特点：**
- 纯下载模式，不上传到 COS
- 支持链接去重，避免重复下载
- 自动创建输出目录
- 文件名从 URL 中自动提取

**输出示例：**
```
已下载链接数: 10
共找到 15 个链接
下载目录: downloads
[1/15] 下载: https://example.com/model.bin
  文件大小: 128.50 MB
  保存路径: downloads/model.bin
  ✓ 成功

完成: 成功 5, 失败 0, 跳过 10
```

---

### 3. upload - 上传本地文件

直接上传本地文件到 COS：

```bash
# 上传文件（使用文件名作为 COS 路径）
./link2cos upload -f /path/to/file.bin

# 指定 COS 存储路径
./link2cos upload -f /path/to/file.bin -p models/mymodel.bin

# 使用自定义配置
./link2cos upload -f /path/to/file.bin -p models/mymodel.bin -c config.yaml
```

**参数说明：**
- `-f, --file`：本地文件路径（必填）
- `-p, --path`：COS 存储路径（可选，默认使用文件名）
- `-c, --config`：配置文件路径（可选，默认 `config.yaml`）

## 📊 上传策略

### 小文件上传（< 100MB）

```
下载到内存 → 直接上传
           ↓
     减少磁盘 I/O
     速度更快
```

### 大文件上传（≥ 100MB）

```
下载到临时文件 → 分块上传（10MB/块）
                  ↓
            并发上传（5个并发）
                  ↓
              断点续传支持
                  ↓
            失败自动清理
```

**进度显示：**
```
[2/5] 处理: https://example.com/large-file.bin
  文件大小: 512.50 MB
  策略: 分块上传
  总分块数: 52
  已上传: 10/52 块 (19.2%)
  已上传: 20/52 块 (38.5%)
  ...
  ✓ 成功
```

## 🔧 开发构建

### Makefile 命令

| 命令 | 说明 | 输出文件 |
|------|------|----------|
| `make build` | 编译当前平台 | `link2cos` |
| `make build-linux` | 编译 Linux AMD64 | `link2cos-linux-amd64` |
| `make build-darwin` | 编译 macOS | `link2cos-darwin-*` |
| `make build-windows` | 编译 Windows | `link2cos-windows-amd64.exe` |
| `make all` | 编译所有平台 | 多个二进制文件 |
| `make clean` | 清理编译产物 | - |

### 手动编译示例

```bash
# Linux AMD64（静态链接，适合容器部署）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o link2cos-linux-amd64

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o link2cos-darwin-arm64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o link2cos-windows-amd64.exe
```

**编译参数说明：**
- `CGO_ENABLED=0`：禁用 CGO，生成纯静态二进制
- `-ldflags="-s -w"`：去除调试信息，减小文件体积

## 💡 使用技巧

### 1. 批量下载并上传到 COS

准备 `links.txt`：
```txt
https://huggingface.co/models/main/model.safetensors
https://huggingface.co/models/main/config.json
https://huggingface.co/models/main/tokenizer.json
```

执行上传：
```bash
./link2cos sync -i links.txt
```

重复运行会自动跳过已下载的文件。

---

### 2. 纯下载模式（不上传到 COS）

```bash
# 下载到默认目录 downloads
./link2cos download -i links.txt

# 下载到指定目录
./link2cos download -i links.txt -o /data/models
```

适用场景：
- 只需要下载文件到本地，不需要上传
- 批量下载模型文件到服务器
- 定期同步远程资源

---

### 3. 本地文件上传到指定路径

```bash
# 上传到根目录
./link2cos upload -f model.bin

# 上传到子目录
./link2cos upload -f model.bin -p models/v1.0/model.bin

# 批量上传多个文件
for file in *.bin; do
  ./link2cos upload -f "$file" -p "models/$file"
done
```

---

### 4. 使用代理下载海外资源

在 `config.yaml` 中配置代理：

```yaml
cos:
  # ... 其他配置
  proxy: "http://127.0.0.1:7890"  # 或 socks5://127.0.0.1:1080
```

然后正常使用 sync 或 download 命令，下载时会自动使用代理。

---

### 5. 使用不同的配置文件

```bash
# 生产环境
./link2cos sync -i links.txt -c config-prod.yaml

# 测试环境
./link2cos download -i links.txt -c config-test.yaml

# 开发环境
./link2cos upload -f file.bin -c config-dev.yaml
```

---

### 6. 清除下载记录

如果需要重新下载所有文件：

```bash
# 删除下载记录文件
rm .link2cos_downloaded.txt

# 然后重新运行命令
./link2cos sync -i links.txt
```

## 🐛 常见问题

### Q1: 上传失败怎么办？

**检查清单：**
1. ✅ 配置文件中的 `secret_id` 和 `secret_key` 是否正确
2. ✅ `bucket_name` 格式是否为 `name-appid`（如 `mybucket-1234567890`）
3. ✅ `region` 是否与存储桶实际地域一致
4. ✅ 密钥是否有 COS 写入权限
5. ✅ 网络是否能访问腾讯云（检查：`ping cos.ap-guangzhou.myqcloud.com`）

**查看详细错误：**
程序会输出详细的错误信息，根据提示排查问题。

---

### Q2: 分块上传中断后会留下碎片吗？

不会。程序在上传失败或中断时会自动调用 `AbortMultipartUpload` 清理未完成的分块，不会产生存储费用。

---

### Q3: 支持的文件大小限制？

- **最小**：无限制
- **最大**：5TB（腾讯云 COS 限制）
- **建议**：大文件（≥100MB）自动使用分块上传

---

### Q4: 如何提高上传速度？

1. 选择离你更近的 COS 地域（`region`）
2. 大文件会自动启用并发分块上传（5 个并发）
3. 网络带宽是主要瓶颈，确保网络稳定

---

### Q5: sync 或 download 命令下载失败？

- 检查 URL 是否可访问（浏览器测试）
- 某些海外链接需要配置代理（在 `config.yaml` 中设置 `proxy`）
- 确保链接以配置的 `url_prefix` 开头（仅 sync 命令）

---

### Q6: 链接去重记录存储在哪里？

- 记录文件：`.link2cos_downloaded.txt`（项目根目录）
- 格式：每行一个链接
- 可手动编辑或删除来管理已下载记录

---

### Q7: sync 和 download 的区别？

| 命令 | 功能 | 适用场景 |
|------|------|---------|
| `sync` | 下载 + 上传到 COS | 需要将文件同步到云端存储 |
| `download` | 仅下载到本地 | 只需要本地存储，不需要云端 |

两个命令都支持链接去重，共享同一个下载记录文件。

---

### Q8: 如何配置代理？

在 `config.yaml` 中添加 `proxy` 配置：

```yaml
cos:
  # ... 其他配置
  proxy: "http://127.0.0.1:7890"     # HTTP 代理
  # 或
  proxy: "socks5://127.0.0.1:1080"   # SOCKS5 代理
```

代理仅用于文件下载，上传到 COS 不使用代理（直连更快）。

## 🏗️ 项目结构

```
Link2COS/
├── cmd/                          # 命令行接口层
│   ├── root.go                  # 根命令
│   ├── sync.go                  # sync 命令：下载并上传到 COS
│   ├── download.go              # download 命令：纯下载
│   └── upload.go                # upload 命令：上传本地文件
│
├── internal/                     # 内部业务逻辑（不对外暴露）
│   ├── constants/               # 常量定义
│   │   └── constants.go
│   │
│   ├── cos/                     # COS 相关功能
│   │   ├── client.go            # COS 客户端初始化
│   │   └── uploader.go          # 文件上传逻辑（分块/普通）
│   │
│   ├── download/                # 下载相关功能
│   │   ├── client.go            # HTTP 客户端创建（支持代理）
│   │   └── downloader.go        # 文件下载逻辑
│   │
│   ├── tracker/                 # 链接追踪
│   │   └── link_tracker.go      # 已下载链接管理（去重）
│   │
│   └── util/                    # 通用工具
│       └── file.go              # 文件读取工具
│
├── config/                       # 配置管理
│   └── config.go                # 配置文件解析
│
├── main.go                       # 程序入口
├── config.yaml                   # 配置文件示例
├── Makefile                      # 构建脚本
└── README.md                     # 项目文档
```

**设计特点：**
- **职责分离**：命令层、业务逻辑层、配置层各司其职
- **模块化**：每个功能独立成包，便于维护和测试
- **可复用**：内部包可被多个命令复用
- **Go 规范**：遵循 Go 项目标准布局

---

## 📚 依赖项

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行框架
- [github.com/tencentyun/cos-go-sdk-v5](https://github.com/tencentyun/cos-go-sdk-v5) - 腾讯云 COS SDK
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML 配置解析

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**Made with ❤️ by difyz9**
