# TVBox Video Source

Go 重写的 TVBox 配置管理服务，单个 exe 部署，零依赖。

## 功能

- JSON 配置读写（default.json 等）
- 基于配置的 URL 重定向
- 文件上传（/raw/ 路径访问）
- 页面管理（CRUD + 模板生成 HTML）
- 静态文件服务（/mod-ce/）
- Swagger UI（/mod-api/）

## 构建

### Windows

```bash
go build -ldflags="-s -w" -o release/tvbox-video-source.exe
go build -ldflags="-s -w" -o release/tvbox-video-source.exe && upx.exe -9 release/tvbox-video-source.exe
```

### Linux amd64

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o release/tvbox-video-source && upx.exe -9 release/tvbox-video-source
```

## 运行

```bash
./tvbox-video-source.exe -port 5000
```

默认端口 `5000`，可通过 `-port` 参数修改。

## 部署

### 直接运行

```bash
# 将 exe 和 wwwroot 目录放到目标服务器（wwwroot 与 exe 同级，不打包进 exe）
scp tvbox-video-source.exe user@server:/app/
scp -r wwwroot user@server:/app/

# 运行
ssh user@server
cd /app && ./tvbox-video-source.exe -port 5000
```

### nginx 反向代理

```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态文件可直接由 nginx 托管（可选）
    location /mod-ce/ {
        alias /app/wwwroot/mod-ce/;
        expires 7d;
    }

    location /raw/ {
        alias /app/wwwroot/uploads/;
        expires 30d;
    }
}
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 健康检查，返回 "success" |
| GET | `/{key}` | 根据 default.json 重定向 |
| GET | `/{file}/{key}` | 根据 {file}.json 重定向 |
| GET | `/read?file=default` | 读取 JSON 配置 |
| POST | `/save?file=default` | 保存 JSON 配置 |
| POST | `/upload` | 上传文件 |
| GET | `/page/list` | 页面列表 |
| POST | `/page/create` | 创建/更新页面 |
| DELETE | `/page/delete?app=xxx` | 删除页面 |
| GET | `/mod-api/` | Swagger UI |
| GET | `/mod-api/openapi.json` | OpenAPI 文档 |

## 项目结构

```
.
├── main.go              # 入口 + 路由
├── handlers/api.go      # API 处理逻辑
├── models/models.go     # 数据模型
├── wwwroot/             # 静态资源
│   ├── default.json
│   ├── pages.json
│   ├── mod-api/          # Swagger UI
│   ├── mod-ce/          # 前端页面
│   └── uploads/         # 上传文件存储
├── go.mod
```

`wwwroot/` 不打包进 exe，始终从 **exe 同级目录** 读取。
