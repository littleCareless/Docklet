# Docklet

Docklet 是一个基于 **pnpm workspace + turbo** monorepo 架构的 Web 应用程序，用于展示 Docker 服务信息。它包含一个 Go 后端和一个 Vue.js 前端。

## 🏗️ 项目架构

此项目采用现代化的 monorepo 架构：

- **pnpm workspace**: 统一的包管理和依赖管理
- **Turbo**: 高性能的构建系统和任务编排
- **多包结构**: 前端和后端作为独立的包进行管理

```
docklet-monorepo/
├── package.json          # 根 package.json
├── pnpm-workspace.yaml   # pnpm workspace 配置
├── turbo.json           # turbo 构建配置
├── frontend/            # @docklet/frontend 包
└── backend/             # @docklet/backend 包
```

## 📋 依赖要求

- **pnpm** >= 8.0.0 (包管理器)
- **Node.js** >= 18.0.0 (前端开发)
- **Go** >= 1.21 (后端开发) 
- **Docker** (容器化部署)

## 🚀 快速开始

### 1. 安装依赖

```bash
# 安装 pnpm（如果尚未安装）
npm install -g pnpm@8.15.0

# 安装项目依赖
pnpm install
```

### 2. 开发模式

```bash
# 启动所有服务的开发模式
pnpm dev

# 或者分别启动
pnpm dev --filter=@docklet/frontend   # 仅前端
pnpm dev --filter=@docklet/backend    # 仅后端
```

### 3. 构建项目

```bash
# 构建所有包
pnpm build

# 构建特定包
pnpm build --filter=@docklet/frontend
pnpm build --filter=@docklet/backend
```

## 🐳 Docker 部署

### 1. 构建 Docker 镜像

```bash
# 使用 pnpm 脚本
pnpm docker:build

# 或直接使用 docker 命令
docker build -t docklet-app .
```

### 2. 运行 Docker 容器

```bash
# 使用 pnpm 脚本
pnpm docker:run

# 或直接使用 docker 命令
docker run -d -p 8888:8888 -v /var/run/docker.sock:/var/run/docker.sock docklet-app
```

**重要**: 为了让应用程序能够访问宿主机的 Docker 服务并列出容器信息，您需要将宿主机的 Docker socket 文件挂载到容器内部。

### 3. 访问应用

容器成功运行后：

- **Web 界面**: `http://localhost:8888`
- **API 端点**:
  - 服务列表: `http://localhost:8888/api/services`
  - 系统服务: `http://localhost:8888/api/system-services`
  - 健康检查: `http://localhost:8888/api/health`

## 🛠️ 开发指南

### Turbo 任务

```bash
# 开发模式（带热重载）
pnpm dev

# 构建所有包
pnpm build

# 代码检查
pnpm lint

# 运行测试
pnpm test

# 清理构建产物
pnpm clean

# 代码格式化
pnpm format
```

### 单独开发包

#### 前端 (Vue.js)

```bash
cd frontend
pnpm dev     # 开发服务器
pnpm build   # 构建生产版本
pnpm lint    # 代码检查
```

#### 后端 (Go)

```bash
cd backend
pnpm dev          # 运行开发服务器
pnpm build        # 构建可执行文件
pnpm test         # 运行测试
pnpm mod-tidy     # 整理 Go 模块
```

## 📁 项目结构

```
docklet-monorepo/
├── package.json                    # 根配置，定义 workspace 和脚本
├── pnpm-workspace.yaml            # pnpm workspace 配置
├── turbo.json                     # turbo 构建管道配置
├── Dockerfile                     # 多阶段构建配置
├── .dockerignore                  # Docker 构建排除文件
├── frontend/                      # 前端 Vue.js 应用
│   ├── package.json              # 前端包配置
│   ├── src/                       # 前端源码
│   ├── vite.config.js            # Vite 构建配置
│   └── dist/                      # 构建输出（生成）
└── backend/                       # 后端 Go 应用
    ├── package.json              # 后端包配置（用于 turbo）
    ├── main.go                   # 后端入口文件
    ├── api/                      # API 处理器
    ├── docker_scanner/           # Docker 服务扫描
    ├── system_scanner/           # 系统服务扫描
    └── bin/                      # 构建输出（生成）
```

## 🔧 配置说明

### Turbo 构建系统

`turbo.json` 定义了构建管道和任务依赖关系：

- **build**: 构建所有包，支持缓存和并行构建
- **dev**: 开发模式，支持热重载
- **lint**: 代码检查，依赖构建完成
- **test**: 运行测试，支持缓存
- **clean**: 清理构建产物

### pnpm Workspace

`pnpm-workspace.yaml` 定义了 monorepo 的包结构，支持：

- 统一的依赖管理
- 包之间的依赖关系
- 高效的磁盘空间利用

## 🐛 故障排除

### 常见问题

1. **pnpm 安装失败**
   ```bash
   # 清理缓存重新安装
   pnpm store prune
   rm -rf node_modules pnpm-lock.yaml
   pnpm install
   ```

2. **Turbo 构建失败**
   ```bash
   # 清理 turbo 缓存
   pnpm clean
   rm -rf .turbo
   pnpm build
   ```

3. **Docker 构建问题**
   ```bash
   # 清理 Docker 缓存
   docker system prune -f
   docker build --no-cache -t docklet-app .
   ```

### 环境变量

- `DOCKLET_PORT`: 后端服务端口（默认: 8888）
- `DOCKLET_HOST_IP`: 主机 IP（用于日志显示）

## 📝 许可证

MIT License