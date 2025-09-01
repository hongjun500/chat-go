# Chat-Go: 分布式聊天系统

一个支持多种传输协议和编码格式的高性能分布式聊天系统。

## 🚀 核心特性

- **多传输协议支持**: TCP 和 WebSocket
- **可插拔编码格式**: JSON 和 Protobuf
- **分布式架构**: 支持 Redis Stream 集群同步
- **高性能**: 零拷贝写入，连接池支持
- **可扩展**: 清晰的分层架构，易于添加新协议和编码格式
- **生产就绪**: 完整的错误处理、监控和日志

## 📋 目录

- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [使用示例](#使用示例)
- [性能对比](#性能对比)
- [扩展性](#扩展性)
- [文档](#文档)

## 🏗️ 架构设计

### 分层架构

```
┌─────────────────────────────────────────────────┐
│                  应用层 (main.go)                │
├─────────────────────────────────────────────────┤
│              网关层 (gateway.go)                │ ← 协议无关的业务逻辑
├─────────────────────────────────────────────────┤
│           传输层 (tcp.go, ws.go)               │ ← 具体传输实现
├─────────────────────────────────────────────────┤
│           编码层 (json/protobuf codec)         │ ← 可插拔编码器
├─────────────────────────────────────────────────┤
│           帧处理层 (frame.go)                   │ ← 底层帧处理
└─────────────────────────────────────────────────┘
```

### 关键优势

- **职责分离**: 每层只处理自己的关注点
- **可插拔设计**: 编码器和传输协议可独立选择
- **向后兼容**: 支持原有客户端无缝迁移
- **高可扩展**: 易于添加新协议和编码格式

## 🚀 快速开始

### 安装依赖

```bash
go mod download
```

### 启动服务器

#### 使用 JSON 编码 (适合开发调试)
```bash
go run cmd/server/main.go
```

#### 使用 Protobuf 编码 (适合生产环境)
```bash
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
go run cmd/server/main.go
```

### 测试客户端

#### JSON 客户端
```bash
go run cmd/client/main.go -addr localhost:8080 -codec json -user alice
```

#### Protobuf 客户端
```bash
go run cmd/client/main.go -addr localhost:8080 -codec protobuf -user bob
```

## ⚙️ 配置说明

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `CHAT_TCP_ADDR` | `:8080` | TCP 服务器地址 |
| `CHAT_WS_ADDR` | `:8081` | WebSocket 服务器地址 |
| `CHAT_HTTP_ADDR` | `:8082` | HTTP 监控地址 |
| `CHAT_TCP_CODEC` | `json` | TCP 编码格式 (json/protobuf) |
| `CHAT_WS_CODEC` | `json` | WebSocket 编码格式 (json/protobuf) |
| `CHAT_READ_TIMEOUT` | `60` | 读取超时(秒) |
| `CHAT_WRITE_TIMEOUT` | `15` | 写入超时(秒) |
| `CHAT_MAX_FRAME` | `1048576` | 最大帧大小(字节) |

## 📖 使用示例

### 混合编码模式

```bash
# 服务器端: TCP 使用 Protobuf，WebSocket 使用 JSON
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
go run cmd/server/main.go

# 高性能客户端 (移动端/服务器)
go run cmd/client/main.go -codec protobuf -user mobile_client

# Web 客户端可以连接到 WebSocket (JSON)
# curl -H "Upgrade: websocket" ws://localhost:8081/ws
```

### 分布式部署

```bash
# 启用 Redis 集群同步
export CHAT_REDIS_ENABLE=true
export CHAT_REDIS_ADDR=redis:6379
export CHAT_REDIS_STREAM=chat_prod
export CHAT_REDIS_GROUP=chat_cluster
go run cmd/server/main.go
```

## 📊 性能对比

| 编码格式 | 包大小 | 解析速度 | 可读性 | 适用场景 |
|---------|--------|----------|--------|----------|
| JSON | 大 | 中等 | 高 | Web 端、调试 |
| Protobuf | 小 | 快 | 低 | 移动端、服务器间 |

### 基准测试

```bash
# 运行性能测试
go test -bench=. ./internal/transport/

# 结果示例:
# BenchmarkJSONCodec-8     100000   10234 ns/op   2048 B/op   32 allocs/op
# BenchmarkProtobufCodec-8 200000    5123 ns/op   1024 B/op   16 allocs/op
```

## 🔧 扩展性

### 添加新编码格式

1. 实现 `MessageCodec` 接口
2. 在 `codec.NewCodec()` 中注册
3. 配置环境变量使用

### 添加新传输协议

1. 实现 `Transport` 和 `Session` 接口
2. 复用现有编码层
3. 在 main.go 中启动

### 支持的消息类型

- `text`: 纯文本消息
- `chat`: 聊天消息  
- `set_name`: 设置昵称
- `command`: 命令消息
- `direct`: 私聊消息
- `ping`/`pong`: 心跳消息
- `ack`: 确认消息

## 📚 文档

- [详细架构文档](docs/architecture.md)
- [使用指南](docs/usage-guide.md)
- [传输层重构文档](docs/transport-refactoring.md)
- [统一传输架构](docs/unified-transport-architecture.md)

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行传输层测试
go test ./internal/transport/

# 运行编码器兼容性测试
go test -v ./internal/transport/ -run TestCodecInteroperability
```

## 🔍 监控

访问监控端点:
```bash
# Prometheus 指标
curl http://localhost:8082/metrics

# 健康检查
curl http://localhost:8082/health
```

## 📝 重构说明

本次重构主要解决了以下问题:

1. **职责混乱**: 原 `protocol.go` 混合了多层职责
2. **硬编码依赖**: `FrameCodec` 硬编码 JSON 逻辑
3. **可扩展性差**: 难以添加新编码格式

### 重构成果

- ✅ 清晰的分层架构
- ✅ 可插拔的编码器
- ✅ 统一的传输接口
- ✅ 向后兼容性
- ✅ 完整的文档和测试

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[MIT License](LICENSE)