# FPK Compose Builder Action

一个 GitHub Action，用于将 Docker Compose 文件转换为 fnOS FPK 应用包。

## 功能特性

- 自动解析 `compose.yaml` 文件中的 `x-fnpack` 扩展配置
- 生成完整的 fnOS 应用包结构（manifest、scripts、ui 等）
- 支持自定义安装向导（wizard）
- 支持应用图标配置
- 输出可直接安装到 fnOS 的 `.fpk` 文件

## 快速开始

### 基本用法

在你的仓库中创建 `.github/workflows/build.yml`：

```yaml
name: Build FPK Package

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Build FPK Package
        uses: tf4fun/fpk-compose-builder@main
        id: build
        with:
          input-dir: ./my-app
          output-dir: ./dist

      - name: Upload Artifact
        uses: actions/upload-artifact@v4
        with:
          name: fpk-package
          path: ${{ steps.build.outputs.fpk-file }}
```

### 输入参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `input-dir` | ✅ | - | 包含 `compose.yaml` 和 `icon.png` 的目录 |
| `output-dir` | ❌ | `./dist` | FPK 文件输出目录 |

### 输出参数

| 参数 | 说明 |
|------|------|
| `fpk-file` | 生成的 FPK 文件路径 |
| `app-name` | 应用名称（来自 manifest） |
| `app-version` | 应用版本（来自 manifest） |

## 输入目录结构

你的输入目录需要包含以下文件：

```
my-app/
├── compose.yaml    # Docker Compose 文件（必需）
└── icon.png        # 应用图标（可选，推荐 256x256）
```

## Compose 文件格式

在标准的 Docker Compose 文件中添加 `x-fnpack` 扩展来配置 fnOS 应用信息：

```yaml
x-fnpack:
  # 应用清单配置
  manifest:
    appname: "my-app"           # 应用标识名（必需）
    version: "1.0.0"            # 版本号（必需）
    desc: "应用描述"             # 应用描述
    arch: "noarch"              # 架构：noarch, x86_64, aarch64
    maintainer: "作者名"         # 维护者
    maintainer_url: "https://example.com"
    distributor: "发布者"
    distributor_url: "https://example.com"
    os_min_ver: "7.2-64555"     # 最低 fnOS 版本
    beta: "no"                  # 是否为测试版
    reloadui: "yes"             # 安装后是否刷新 UI
    display_name: "显示名称"     # 在 fnOS 中显示的名称
    changelog: "更新日志"
    source: thirdparty          # 来源：thirdparty

  # 安装向导配置（可选）
  wizard/install: |
    [
      {
        "stepTitle": "配置步骤",
        "items": [
          {
            "type": "text",
            "field": "wizard_username",
            "label": "用户名",
            "rules": [{ "required": true, "min": 3, "max": 50 }]
          }
        ]
      }
    ]

  # UI 配置（可选）
  app/ui/config: |
    {
      ".url": {
        "my-app.Application": {
          "title": "我的应用",
          "desc": "应用描述",
          "icon": "images/icon-{0}.PNG",
          "type": "url",
          "url": "/my-app/",
          "allUsers": false
        }
      }
    }

services:
  my-service:
    image: my-image:latest
    container_name: my-app
    ports:
      - 8080:8080
    environment:
      - USERNAME=${wizard_username}
    volumes:
      - /var/apps/my-app/data:/data
    restart: unless-stopped
    networks:
      - trim-default

networks:
  trim-default:
    external: true
```

## 向导字段类型

在 `wizard/install` 中可以使用以下字段类型：

| 类型 | 说明 | 示例 |
|------|------|------|
| `text` | 文本输入框 | 用户名、密码、API Key |
| `tips` | 提示信息 | 显示帮助文本 |
| `number` | 数字输入 | 端口号、数量 |
| `select` | 下拉选择 | 选项列表 |
| `checkbox` | 复选框 | 开关选项 |

向导中定义的字段可以在 compose 文件中通过 `${wizard_fieldname}` 引用。

## 完整示例

### 示例 1：简单应用

```yaml
x-fnpack:
  manifest:
    appname: "hello-world"
    version: "1.0.0"
    desc: "Hello World 示例应用"
    display_name: "Hello World"

services:
  hello:
    image: nginx:alpine
    ports:
      - 8080:80
    restart: unless-stopped
    networks:
      - trim-default

networks:
  trim-default:
    external: true
```

### 示例 2：带配置向导的应用

```yaml
x-fnpack:
  manifest:
    appname: "my-chat"
    version: "2.0.0"
    desc: "AI 聊天应用"
    display_name: "AI Chat"
    os_min_ver: "7.2-64555"

  wizard/install: |
    [
      {
        "stepTitle": "API 配置",
        "items": [
          {
            "type": "tips",
            "helpText": "请输入您的 API Key"
          },
          {
            "type": "text",
            "field": "wizard_api_key",
            "label": "API Key",
            "rules": [{ "required": true, "min": 10 }]
          }
        ]
      }
    ]

services:
  chat:
    image: my-chat:latest
    environment:
      - API_KEY=${wizard_api_key}
    ports:
      - 3000:3000
    restart: unless-stopped
    networks:
      - trim-default

networks:
  trim-default:
    external: true
```

## 多应用构建

使用 matrix 策略构建多个应用：

```yaml
name: Build Multiple FPK Packages

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        app: [app1, app2, app3]
    
    steps:
      - uses: actions/checkout@v4

      - name: Build FPK
        uses: tf4fun/fpk-compose-builder@main
        with:
          input-dir: apps/${{ matrix.app }}
          output-dir: dist/${{ matrix.app }}

      - name: Upload Artifact
        uses: actions/upload-artifact@v4
        with:
          name: fpk-${{ matrix.app }}
          path: dist/${{ matrix.app }}/*.fpk
```

## 发布到 Release

构建并发布到 GitHub Release：

```yaml
name: Build and Release

on:
  release:
    types: [created]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build FPK
        uses: tf4fun/fpk-compose-builder@main
        id: build
        with:
          input-dir: ./my-app
          output-dir: ./dist

      - name: Upload to Release
        uses: softprops/action-gh-release@v1
        with:
          files: ${{ steps.build.outputs.fpk-file }}
```

## 本地开发

如果你想在本地测试构建：

```bash
# 克隆仓库
git clone https://github.com/tf4fun/fpk-compose-builder.git
cd fpk-compose-builder

# 构建 Docker 镜像
docker build -t fpk-builder .

# 运行构建
docker run --rm -v $(pwd)/examples/chromium:/input -v $(pwd)/dist:/output \
  fpk-builder build -i /input -o /output
```

## 注意事项

1. **网络配置**：建议使用 `trim-default` 外部网络，这是 fnOS 的默认 Docker 网络
2. **数据持久化**：卷挂载路径建议使用 `/var/apps/<appname>/` 前缀
3. **环境变量**：可以使用 `${TRIM_UID}` 和 `${TRIM_GID}` 获取 fnOS 用户 ID
4. **图标要求**：推荐使用 256x256 的 PNG 图片

## 许可证

MIT License
