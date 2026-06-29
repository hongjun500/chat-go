# Chat-Go 分层架构重构文档

## 概述

本文档描述了 Chat-Go 项目的分层架构重构，重点解决了协议层与传输层职责混淆的问题。通过清晰的职责分离，提高了代码的可维护性、可扩展性和可测试性。

## 🎯 重构目标

### 问题分析
原有架构存在以下问题：
1. **职责混淆**：协议层和传输层职责不明确
2. **耦合过紧**：传输层直接处理编解码逻辑
3. **扩展困难**：添加新的传输协议或编码格式困难
4. **测试复杂**：各层耦合导致单元测试困难

### 解决方案
通过分层设计原则，重新定义各层职责：

```
┌─────────────────────────────────────────┐
│           应用业务层 (Chat/Hub)           │ ← 聊天逻辑、用户管理
├─────────────────────────────────────────┤
│           协议层 (Protocol)             │ ← 消息定义、编解码、路由
│  - MessageFactory (消息工厂)             │
│  - ProtocolManager (协议管理器)          │
│  - MessageCodec (编解码接口)             │
│  - MessageRouter (消息路由器)            │
├─────────────────────────────────────────┤
│           传输层 (Transport)            │ ← 连接管理、会话管理
│  - Transport Interface (传输接口)       │
│  - Session Management (会话管理)        │
│  - Gateway (网关)                       │
│  - TCP/WebSocket Implementations       │
├─────────────────────────────────────────┤
│           网络层 (Network)              │ ← 底层网络操作
└─────────────────────────────────────────┘
```

## 🏗️ 架构设计

### 协议层 (Protocol Layer)

**核心职责**：
- 消息格式定义和验证
- 编解码算法实现
- 消息路由和分发
- 协议版本管理

**主要组件**：

#### 1. MessageFactory (消息工厂)
```go
type MessageFactory struct {
    version string
}

// 统一的消息创建接口
func (f *MessageFactory) CreateTextMessage(text string) *Envelope
func (f *MessageFactory) CreateCommandMessage(command string) *Envelope
func (f *MessageFactory) CreateAckMessage(status, correlationID string) *Envelope
```

#### 2. ProtocolManager (协议管理器)
```go
type ProtocolManager struct {
    codec   MessageCodec
    factory *MessageFactory
    router  *MessageRouter
}

// 协议层的统一入口
func (p *ProtocolManager) EncodeMessage(w io.Writer, envelope *Envelope) error
func (p *ProtocolManager) DecodeMessage(r io.Reader, envelope *Envelope, maxSize int) error
func (p *ProtocolManager) ProcessMessage(r io.Reader, maxSize int) error
```

#### 3. MessageCodec (编解码接口)
```go
type MessageCodec interface {
    Name() string
    Encode(w io.Writer, m *Envelope) error
    Decode(r io.Reader, m *Envelope, maxSize int) error
}

// 支持多种编码格式
- JSONCodec: JSON 格式编解码
- ProtobufCodec: Protocol Buffers 格式编解码
```

#### 4. MessageRouter (消息路由器)
```go
type MessageRouter struct {
    handlers       map[MessageType]MessageHandler
    defaultHandler MessageHandler
}

// 消息分发和路由
func (r *MessageRouter) RegisterHandler(msgType MessageType, handler MessageHandler)
func (r *MessageRouter) Dispatch(env *Envelope) error
```

### 传输层 (Transport Layer)

**核心职责**：
- 网络连接管理
- 会话生命周期管理
- 数据传输抽象
- 协议无关的消息传递

**主要组件**：

#### 1. Transport Interface (传输接口)
```go
type Transport interface {
    Name() string
    Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}

// 具体实现
- TCPServer: TCP 传输实现
- WebSocketServer: WebSocket 传输实现
```

#### 2. Session Interface (会话接口)
```go
type Session interface {
    ID() string
    RemoteAddr() string
    SendEnvelope(*protocol.Envelope) error
    Close() error
}

// 统一的会话抽象，屏蔽底层传输差异
```

#### 3. Gateway Interface (网关接口)
```go
type Gateway interface {
    OnSessionOpen(sess Session)
    OnEnvelope(sess Session, msg *protocol.Envelope)
    OnSessionClose(sess Session)
}

// 传输层与业务层的桥梁
```

#### 4. SimpleGateway (简单网关实现)
```go
type SimpleGateway struct {
    sessionManager    *SessionManager
    messageHandlers   map[string]handlerFunc
    protocolManager   *protocol.ProtocolManager
}

// 提供基础的消息转发和会话管理
```

## 🔄 职责分离

### 协议层职责
✅ **负责**：
- 消息格式定义 (Envelope 结构)
- 消息编解码 (JSON/Protobuf)
- 消息路由分发
- 协议版本控制
- 消息验证和处理

❌ **不负责**：
- 网络连接管理
- 会话状态管理
- 传输层错误处理
- 具体传输协议实现

### 传输层职责
✅ **负责**：
- 网络连接建立和维护
- 会话生命周期管理
- 数据帧处理 (TCP 帧格式)
- 传输层错误处理
- 连接超时和心跳

❌ **不负责**：
- 消息内容理解
- 消息编解码逻辑
- 业务逻辑处理
- 消息路由决策

## 📊 接口设计

### 层间接口

#### 1. 协议层向传输层提供
```go
type MessageCodecProvider interface {
    GetCodec() protocol.MessageCodec
}

// 传输层通过此接口获取编解码能力
```

#### 2. 传输层向协议层提供
```go
type Session interface {
    SendEnvelope(*protocol.Envelope) error
    // 统一的消息发送接口
}
```

#### 3. 网关接口桥接两层
```go
type Gateway interface {
    OnSessionOpen(sess Session)          // 会话管理
    OnEnvelope(sess Session, msg *Envelope)  // 消息处理
    OnSessionClose(sess Session)         // 清理资源
}
```

## 🚀 使用示例

### 创建分层服务器

```go
// 1. 创建协议管理器
tcpProtocolManager := protocol.NewProtocolManager(protocol.CodecJson)
wsProtocolManager := protocol.NewProtocolManager(protocol.CodecProtobuf)

// 2. 创建网关
tcpGateway := transport.NewSimpleGateway(tcpProtocolManager)
wsGateway := transport.NewSimpleGateway(wsProtocolManager)

// 3. 注册消息处理器
tcpGateway.RegisterProtocolHandler(protocol.MsgText, func(env *protocol.Envelope) error {
    log.Printf("Received: %s", string(env.Data))
    return nil
})

// 4. 启动传输服务器
tcpServer := transport.NewTCPServer(":8080")
go tcpServer.Start(ctx, ":8080", tcpGateway, transport.Options{
    TCPProtocolManager: tcpProtocolManager,
    ReadTimeout:        60 * time.Second,
    MaxFrameSize:       1024 * 1024,
})

wsServer := transport.NewWebSocketServer("/ws")
go wsServer.Start(ctx, ":8081", wsGateway, transport.Options{
    WSProtocolManager: wsProtocolManager,
})
```

### 添加新的编码格式

```go
// 1. 实现 MessageCodec 接口
type MyCustomCodec struct{}

func (c *MyCustomCodec) Name() string { return "custom" }
func (c *MyCustomCodec) Encode(w io.Writer, m *protocol.Envelope) error { /* 实现 */ }
func (c *MyCustomCodec) Decode(r io.Reader, m *protocol.Envelope, maxSize int) error { /* 实现 */ }

// 2. 注册到工厂
const CodecCustom = 2
protocol.CodecFactories[CodecCustom] = func() protocol.MessageCodec { 
    return &MyCustomCodec{} 
}

// 3. 使用自定义编码
pm := protocol.NewProtocolManager(CodecCustom)
```

### 添加新的传输协议

```go
// 1. 实现 Transport 接口
type UDPServer struct {
    addr string
}

func (s *UDPServer) Name() string { return "udp" }
func (s *UDPServer) Start(ctx context.Context, addr string, gateway transport.Gateway, opt transport.Options) error {
    // UDP 传输实现
}

// 2. 实现对应的 Session
type udpSession struct {
    // UDP 会话实现
}

// 3. 使用新传输协议
udpServer := &UDPServer{}
go udpServer.Start(ctx, ":8082", gateway, options)
```

## ✅ 向后兼容性

### 保持现有接口可用

```go
// 旧的 Protocol 结构仍然可用
protocol := protocol.NewProtocol(protocol.CodecJson)  // Deprecated 但仍可用
codec := protocol.GetCodec()
textMsg := protocol.CreateTextMessage("Hello")

// 旧的 Gateway 构造函数
gateway := transport.NewSimpleGatewayWithCodec(protocol.CodecJson)  // 内部转换为 ProtocolManager
```

### 渐进式迁移路径

1. **第一阶段**：使用新接口，保留旧接口
2. **第二阶段**：标记旧接口为 Deprecated
3. **第三阶段**：移除旧接口（未来版本）

## 🧪 测试策略

### 单元测试覆盖

#### 协议层测试
```go
func TestMessageFactory_CreateTextMessage(t *testing.T)     // 消息工厂
func TestProtocolManager_MessageHandlers(t *testing.T)     // 协议管理器
func TestMessageRouter_Dispatch(t *testing.T)              // 消息路由
```

#### 传输层测试
```go
func TestSimpleGateway_Creation(t *testing.T)               // 网关创建
func TestSessionManager_Operations(t *testing.T)           // 会话管理
func TestTransportOptions_ProtocolManagerGetters(t *testing.T) // 选项配置
```

### 集成测试
- TCP 传输 + JSON 编码
- WebSocket 传输 + Protobuf 编码
- 混合场景测试

## 🎉 重构收益

### 代码质量提升
- ✅ **职责单一**：每个组件职责明确
- ✅ **松耦合**：层间依赖最小化
- ✅ **高内聚**：相关功能集中管理

### 可维护性增强
- ✅ **易于理解**：清晰的分层结构
- ✅ **易于修改**：影响范围可控
- ✅ **易于测试**：独立的单元测试

### 可扩展性改善
- ✅ **新编码格式**：插件式添加
- ✅ **新传输协议**：接口统一
- ✅ **新业务逻辑**：处理器模式

### 性能优化
- ✅ **减少耦合**：避免不必要的依赖
- ✅ **缓存优化**：协议对象可复用
- ✅ **并发安全**：线程安全的设计

## 📈 迁移指南

### 对现有代码的影响

**无需立即修改**：
- 现有的 `protocol.NewProtocol()` 调用
- 现有的 `transport.NewSimpleGateway(codecType)` 调用
- 现有的消息处理逻辑

**建议逐步迁移**：
1. 使用 `protocol.NewProtocolManager()` 替代 `protocol.NewProtocol()`
2. 使用 `transport.NewSimpleGateway(protocolManager)` 替代编解码器类型
3. 使用新的消息工厂和路由器API

### 最佳实践

1. **分层原则**：严格按照层次职责开发
2. **接口编程**：依赖接口而非具体实现
3. **单元测试**：每层独立测试
4. **文档同步**：及时更新文档

---

本重构方案成功解决了原有架构的职责混淆问题，提供了清晰的分层设计，为系统的长期维护和扩展奠定了坚实基础。