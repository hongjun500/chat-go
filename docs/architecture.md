# Chat-Go 分布式聊天系统架构文档

## 概述

Chat-Go 是一个支持多种传输协议和编码格式的分布式聊天系统。该系统的核心设计理念是**可插拔的传输层**和**可切换的编码格式**，为不同的应用场景提供灵活的选择。

## 架构设计

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

### 核心组件

#### 1. 传输层 (Transport Layer)

**接口定义** (`transport.go`):
- `Transport`: 统一的传输接口
- `Session`: 会话管理接口  
- `Gateway`: 协议无关的网关接口

**实现类**:
- `TCPServer`: TCP 传输实现，使用长度前缀帧格式
- `WebSocketServer`: WebSocket 传输实现，支持结构化消息

#### 2. 编码层 (Encoding Layer)

**接口定义** (`codec.go`):
- `MessageCodec`: 消息编解码接口

**实现类**:
- `JSONCodec`: JSON 格式编解码器
- `ProtobufCodec`: Protocol Buffers 格式编解码器

#### 3. 协议层 (Protocol Layer)

**消息格式** (`envelope.go`):
```go
type Envelope struct {
    Version     string      `json:"version"`     // 协议版本
    Type        MessageType `json:"type"`        // 消息类型
    Encoding    Encoding    `json:"encoding"`    // 编码格式
    Mid         string      `json:"mid"`         // 消息ID
    From        string      `json:"from"`        // 发送者
    To          string      `json:"to"`          // 接收者
    Ts          int64       `json:"ts"`          // 时间戳
    Payload     json.RawMessage `json:"payload"`  // JSON负载
    Data        []byte      `json:"data"`        // 二进制负载
}
```

**消息类型** (`payloads.go`):
- `TextPayload`: 纯文本消息
- `ChatPayload`: 聊天消息
- `SetNamePayload`: 设置昵称
- `CommandPayload`: 命令消息
- `AckPayload`: 确认消息
- `PingPayload`/`PongPayload`: 心跳消息

#### 4. 帧处理层 (Framing Layer)

**核心组件**:
- `FrameCodec`: 纯粹的帧处理器，处理长度前缀帧格式
- `FramedMessageCodec`: 组合帧处理和消息编解码

## 编码格式支持

### JSON 编码
- **优势**: 人类可读，调试友好，Web 兼容性好
- **适用场景**: WebSocket 连接，开发调试，轻量级应用
- **性能**: 解析速度中等，包体积较大

### Protobuf 编码  
- **优势**: 高效的二进制格式，包体积小，强类型验证
- **适用场景**: TCP 连接，高性能要求，移动端应用
- **性能**: 解析速度快，包体积小

## 传输协议支持

### TCP 传输
- **帧格式**: 4字节长度前缀 + 消息内容
- **特点**: 可靠传输，支持可插拔编码器
- **适用**: 服务器间通信，移动客户端

### WebSocket 传输
- **消息格式**: 支持结构化消息和向后兼容的纯文本
- **特点**: 双向通信，自动心跳，Web 兼容
- **适用**: Web 客户端，实时通信

## 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `CHAT_TCP_ADDR` | `:8080` | TCP 服务器地址 |
| `CHAT_WS_ADDR` | `:8081` | WebSocket 服务器地址 |
| `CHAT_TCP_CODEC` | `json` | TCP 编码格式 (json/protobuf) |
| `CHAT_WS_CODEC` | `json` | WebSocket 编码格式 (json/protobuf) |
| `CHAT_READ_TIMEOUT` | `60` | 读取超时时间(秒) |
| `CHAT_WRITE_TIMEOUT` | `15` | 写入超时时间(秒) |
| `CHAT_MAX_FRAME` | `1048576` | 最大帧大小(字节) |

### 使用示例

#### 启动 JSON + TCP 服务器
```bash
export CHAT_TCP_CODEC=json
export CHAT_TCP_ADDR=:8080
go run cmd/server/main.go
```

#### 启动 Protobuf + TCP 服务器
```bash
export CHAT_TCP_CODEC=protobuf  
export CHAT_TCP_ADDR=:8080
go run cmd/server/main.go
```

#### 混合模式 (TCP使用Protobuf，WebSocket使用JSON)
```bash
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
export CHAT_TCP_ADDR=:8080
export CHAT_WS_ADDR=:8081
go run cmd/server/main.go
```

## 扩展性

### 添加新的编码格式

1. 实现 `MessageCodec` 接口:
```go
type MyCodec struct{}

func (c *MyCodec) ContentType() string { return "application/my-format" }
func (c *MyCodec) Encode(w io.Writer, m *protocol.Envelope) error { ... }
func (c *MyCodec) Decode(r io.Reader, m *protocol.Envelope, maxSize int) error { ... }
```

2. 在 `codec.NewCodec()` 中添加支持:
```go
case "myformat":
    return &MyCodec{}, nil
```

### 添加新的传输协议

1. 实现 `Transport` 接口:
```go
type MyTransport struct{}

func (t *MyTransport) Name() string { return "mytransport" }
func (t *MyTransport) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error { ... }
```

2. 实现对应的 `Session` 接口

## 最佳实践

### 选择编码格式

- **开发阶段**: 使用 JSON 便于调试
- **生产环境**: 
  - Web 客户端使用 JSON
  - 移动客户端和服务器间使用 Protobuf
  - 高性能场景优先选择 Protobuf

### 选择传输协议

- **Web 应用**: WebSocket + JSON
- **移动应用**: TCP + Protobuf  
- **服务器间通信**: TCP + Protobuf
- **混合场景**: 根据客户端类型动态选择

### 性能优化

1. **帧大小**: 根据网络环境调整 `MaxFrameSize`
2. **缓冲区**: 根据并发数调整 `OutBuffer`
3. **超时设置**: 根据网络延迟调整读写超时
4. **连接池**: TCP 传输支持连接复用

## 错误处理

系统提供了完善的错误处理机制:

1. **编码错误**: 返回 `bad_payload` 状态
2. **认证错误**: 返回 `unauthorized` 状态  
3. **网络错误**: 自动断开连接并清理资源
4. **协议错误**: 返回 `unknown_type` 状态

## 监控和观测

- 通过 `observe.StartHTTP()` 提供 HTTP 监控端点
- 支持 Prometheus 指标收集
- 集成 Zap 日志框架，提供结构化日志

## 分布式支持

- 集成 Redis Stream 支持分布式消息传递
- 支持多实例负载均衡
- 支持集群间消息同步

---

本架构设计遵循了**单一职责原则**、**开闭原则**和**依赖倒置原则**，提供了高度的可扩展性和可维护性。