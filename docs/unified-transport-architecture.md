# 统一传输层架构文档 (Unified Transport Layer Architecture)

## 概述 (Overview)

本次重构实现了分布式聊天系统的统一传输层架构，解决了 TCP 和 WebSocket 传输实现不一致的问题，并支持了可插拔的编码格式（JSON 和 Protobuf）。

## 架构设计 (Architecture Design)

### 分层架构 (Layered Architecture)

```
┌─────────────────────────────────────┐
│        Business Layer              │  <- Gateway, Session, Hub
├─────────────────────────────────────┤
│     Transport Abstraction Layer    │  <- Transport Interface
│     ├── TCPServer                   │  <- TCP implementation
│     └── WebSocketServer             │  <- WebSocket implementation
├─────────────────────────────────────┤
│      Message Codec Layer           │  <- MessageCodec Interface
│     ├── JSONCodec                   │  <- JSON encoding/decoding
│     └── ProtobufCodec               │  <- Protobuf encoding/decoding
├─────────────────────────────────────┤
│      Frame Processing Layer        │  <- FrameCodec (TCP only)
│     └── FramedMessageCodec          │  <- Combines frame + message codec
├─────────────────────────────────────┤
│        Network Layer               │  <- net.Conn, websocket.Conn
└─────────────────────────────────────┘
```

### 核心接口 (Core Interfaces)

#### Transport 接口
```go
type Transport interface {
    Name() string
    Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}
```

#### Session 接口
```go
type Session interface {
    ID() string
    RemoteAddr() string
    SendEnvelope(*Envelope) error
    Close() error
}
```

#### Gateway 接口
```go
type Gateway interface {
    OnSessionOpen(sess Session)
    OnEnvelope(sess Session, msg *Envelope)
    OnSessionClose(sess Session, err error)
}
```

#### MessageCodec 接口
```go
type MessageCodec interface {
    ContentType() string
    Encode(w io.Writer, m *Envelope) error
    Decode(r io.Reader, m *Envelope, maxSize int) error
}
```

## 实现详情 (Implementation Details)

### TCP 传输 (TCP Transport)

- 使用 length-prefixed framing (4字节长度前缀)
- 支持可插拔的 MessageCodec (JSON/Protobuf)
- 零拷贝写入优化
- 连接池支持

### WebSocket 传输 (WebSocket Transport)

- 支持结构化消息 (Envelope 格式)
- 向后兼容纯文本消息
- 自动心跳检测
- 支持可插拔的 MessageCodec

### 编码层 (Encoding Layer)

#### JSON Codec
- 基于 `encoding/json`
- 流式编解码
- 支持大小限制
- 格式验证

#### Protobuf Codec
- 基于 `google.golang.org/protobuf`
- 高效的二进制编码
- 强类型验证
- 向前/向后兼容性

## 使用方式 (Usage)

### 基本设置 (Basic Setup)

```go
// 创建 Hub 和命令注册表
hub := chat.NewHub()
cmdReg := command.NewRegistry()
gateway := &transport.GatewayHandler{Hub: hub, Commands: cmdReg}

// TCP 服务器 - JSON 编码
tcpSrv := &transport.TCPServer{Codec: &transport.JSONCodec{}}
go tcpSrv.Start(ctx, ":8080", gateway, options)

// WebSocket 服务器 - JSON 编码  
wsSrv := &transport.WebSocketServer{Codec: &transport.JSONCodec{}}
go wsSrv.Start(ctx, ":8081", gateway, options)

// TCP 服务器 - Protobuf 编码
tcpSrvPb := &transport.TCPServer{Codec: &transport.ProtobufCodec{}}
go tcpSrvPb.Start(ctx, ":8082", gateway, options)
```

### 消息格式 (Message Format)

#### Envelope 结构
```go
type Envelope struct {
    Version     string `json:"version,omitempty"`
    Type        string `json:"type"`                // 消息类型
    From        string `json:"from,omitempty"`      // 发送者
    To          []string `json:"to,omitempty"`      // 接收者
    Payload     json.RawMessage `json:"payload,omitempty"`    // JSON 载荷
    Data        []byte `json:"data,omitempty"`      // 二进制载荷
    Ts          int64  `json:"ts,omitempty"`        // 时间戳
    // ... 其他字段
}
```

#### 消息类型 (Message Types)
- `text`: 纯文本消息
- `set_name`: 设置用户昵称
- `chat`: 聊天消息
- `direct`: 私聊消息  
- `command`: 命令消息
- `ack`: 确认消息
- `ping`/`pong`: 心跳消息

## 分布式聊天系统支持 (Distributed Chat System Support)

### 特性 (Features)

1. **传输层统一**: TCP 和 WebSocket 使用相同的业务逻辑处理
2. **编码格式可插拔**: 支持 JSON 和 Protobuf，易于扩展其他格式
3. **会话管理**: 统一的 Session 接口管理连接生命周期
4. **消息路由**: Gateway 模式处理消息路由和业务逻辑
5. **向后兼容**: WebSocket 支持传统的纯文本客户端

### 扩展性 (Extensibility)

#### 添加新的传输类型
```go
type UDPServer struct {
    Codec MessageCodec
}

func (u *UDPServer) Name() string { return "udp" }
func (u *UDPServer) Start(ctx context.Context, addr string, gateway Gateway, opt Options) error {
    // UDP 传输实现
}
```

#### 添加新的编码格式
```go
type XMLCodec struct{}

func (x *XMLCodec) ContentType() string { return "application/xml" }
func (x *XMLCodec) Encode(w io.Writer, m *Envelope) error { /* XML编码 */ }
func (x *XMLCodec) Decode(r io.Reader, m *Envelope, maxSize int) error { /* XML解码 */ }
```

## 测试覆盖 (Test Coverage)

- 单元测试: 各个编码器的独立测试
- 集成测试: 传输层和编码层的组合测试
- 接口测试: Transport 接口合规性测试
- 性能测试: 编码/解码性能基准测试

## 性能优化 (Performance Optimizations)

1. **零拷贝写入**: TCP 使用 `net.Buffers` 避免额外的内存拷贝
2. **缓冲区池**: 帧编码器使用缓冲区池减少内存分配
3. **流式处理**: JSON 编码器支持流式读写
4. **连接复用**: 支持长连接和连接池

## 迁移指南 (Migration Guide)

### 从旧的 WebSocket 实现迁移

1. **客户端兼容性**: 新实现向后兼容纯文本消息
2. **结构化消息**: 建议客户端迁移到 Envelope 格式
3. **配置更改**: 在 main.go 中使用新的 WebSocketServer

### 从旧的 TCP 实现迁移

1. **帧格式不变**: 保持 4 字节长度前缀的帧格式
2. **编码格式**: 默认使用 JSON，可选择 Protobuf
3. **API 变更**: 使用新的 Transport 接口

## 最佳实践 (Best Practices)

1. **选择合适的编码**: 
   - JSON: 适合调试和人类可读的场景
   - Protobuf: 适合性能要求高的生产环境

2. **错误处理**: 
   - 使用 Gateway 的错误回调处理连接错误
   - 实现重连机制

3. **安全考虑**:
   - 设置合理的消息大小限制
   - 实现消息签名验证
   - 使用 TLS 加密传输

4. **监控指标**:
   - 连接数统计
   - 消息吞吐量
   - 编码/解码延迟

## 未来扩展 (Future Extensions)

1. **QUIC 支持**: 添加基于 QUIC 的传输层
2. **消息压缩**: 添加 gzip/zstd 压缩支持  
3. **负载均衡**: 添加传输层负载均衡
4. **多路复用**: 支持单连接多流的传输
5. **集群支持**: 添加节点间的传输抽象