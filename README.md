# MiniMax Status Dashboard

用于 dashy 监控的 MiniMax 套餐使用状态页面，Go 语言实现，Docker 部署。

## 快速开始

### 1. 配置环境变量

```bash
export MINIMAX_API_KEY=your_api_key_here
export MINIMAX_API_URL=https://www.minimaxi.com
```

### 2. 本地运行

```bash
go build -o minimax-status .
./minimax-status
```

访问 http://localhost:8080

### 3. Docker 部署

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

或手动构建：

```bash
docker build -t minimax-status .
docker run -d -p 8080:8080 \
  -e MINIMAX_API_KEY=your_key \
  minimax-status
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MINIMAX_API_KEY` | MiniMax API 密钥 | 必填 |
| `MINIMAX_API_URL` | API 服务器地址 | `https://www.minimaxi.com` |
| `PORT` | HTTP 监听端口 | `8080` |

## 嵌入 dashy

在 dashy 的 `widgets` 配置中添加：

```yaml
widgets:
  - type: iframe
    options:
      url: http://your-server:8080
```

## 主题支持

页面自动适配浏览器主题（浅色/深色模式）。

## License

MIT
