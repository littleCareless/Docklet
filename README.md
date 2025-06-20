# Docklet

Docklet 是一个 Web 应用程序，用于展示 Docker 服务信息。它包含一个 Go 后端和一个 Vue.js 前端。

## 依赖

*   Docker (用于构建和运行应用)
*   Node.js (仅在本地开发前端时需要)
*   Go (仅在本地开发后端时需要)

## 如何构建和运行

本项目被设计为通过 Docker 运行。

### 1. 构建 Docker 镜像

在项目的根目录下，执行以下命令来构建 Docker 镜像。请将 `docklet-app` 替换为您希望的镜像名称。

```bash
docker build -t docklet-app .
```

### 2. 运行 Docker 容器

构建成功后，使用以下命令来运行 Docker 容器。

**重要**: 为了让应用程序能够访问宿主机的 Docker 服务并列出容器信息，您需要将宿主机的 Docker socket 文件挂载到容器内部。

```bash
docker run -p 8888:8888 -v /var/run/docker.sock:/var/run/docker.sock docklet-app
```

参数说明:
*   `-p 8888:8888`: 将主机的 8888 端口映射到容器的 8888 端口。如果需要，您可以更改主机端口 (例如 `-p <your_port>:8888`)。
*   `-v /var/run/docker.sock:/var/run/docker.sock`: 将宿主机的 Docker socket 挂载到容器中，允许应用与 Docker 守护进程通信。
*   `docklet-app`: 您在构建步骤中为镜像指定的名称。

### 3. 访问应用

容器成功运行后：

*   **Web 界面**: 打开浏览器并访问 `http://localhost:8888`
*   **API 端点**:
    *   服务列表: `http://localhost:8888/api/services`
    *   健康检查: `http://localhost:8888/api/health`

## 开发

### 前端 (Vue.js)

前端代码位于 `frontend` 目录。

```bash
cd frontend
npm install
npm run dev
```

### 后端 (Go)

后端代码位于 `backend` 目录。

```bash
cd backend
go run main.go
```
默认情况下，后端服务会在 `8888` 端口启动。您可以通过设置 `DOCKLET_PORT` 环境变量来更改端口。

## Dockerfile 说明

根目录下的 `Dockerfile` 使用多阶段构建：
1.  **frontend-builder**: 构建 Vue.js 前端静态文件。
2.  **backend-builder**: 构建 Go 后端可执行文件。
3.  **Final stage**: 将构建好的前端静态文件和后端可执行文件复制到一个轻量级的 Alpine 镜像中。

`.dockerignore` 文件用于排除不必要的文件，以优化 Docker 构建过程。