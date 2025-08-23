# Transport Layer Refactoring Documentation

## 概述 (Overview)

本次重构主要解决了 `protocol.go` 文件的职责混乱问题，通过分离关注点提升了代码的可维护性和可扩展性。

## 问题分析 (Problem Analysis)

### 原有问题 (Original Issues)

1. **职责混乱**: `protocol.go` 文件混合了多个不同层次的职责
   - 协议定义 (`Envelope` 结构体)
   - 传输模式定义 (`TCPMode`)
   - 编解码实现 (`FrameCodec`)
   - 网络工具函数 (`SafeDeadline`)

2. **依赖混乱**: `FrameCodec` 硬编码了 JSON 编解码逻辑
   - 无法支持其他编码格式 (如 Protobuf)
   - 违反了开闭原则

3. **可扩展性差**: 架构设计不利于添加新的编码格式

### 根本原因 (Root Causes)

- 违反了单一职责原则 (Single Responsibility Principle)
- 缺乏适当的抽象层次
- 硬编码依赖，缺乏依赖注入

## 重构方案 (Refactoring Solution)

### 新的文件结构 (New File Structure)

```
internal/transport/
├── envelope.go           # 消息协议定义
├── frame.go             # 纯粹的帧处理
├── framed_codec.go      # 组合帧处理和消息编解码
├── protocol.go          # 向后兼容的遗留接口
├── codec.go             # 编解码接口定义
├── json_codec.go        # JSON 编解码实现
├── proto_codec.go       # Protobuf 编解码实现
└── tcp.go               # TCP 传输实现
```

### 关键改进 (Key Improvements)

1. **分离关注点**:
   - `envelope.go`: 只包含协议定义和传输模式
   - `frame.go`: 只包含底层帧处理逻辑
   - `framed_codec.go`: 组合帧处理和消息编解码

2. **移除硬编码依赖**:
   - `FrameCodec` 现在是纯粹的帧处理器
   - 通过 `FramedMessageCodec` 组合不同的编解码器

3. **提升可扩展性**:
   - 支持插拔式编解码器 (JSON, Protobuf)
   - 易于添加新的编码格式

4. **保持向后兼容**:
   - 提供 `LegacyFrameCodec` 包装器
   - 现有代码无需修改

## 架构设计 (Architecture Design)

### 分层架构 (Layered Architecture)

```
┌─────────────────────────────────────┐
│        Business Layer              │  <- Gateway, Session
├─────────────────────────────────────┤
│     Transport Abstraction Layer    │  <- Transport, Options
├─────────────────────────────────────┤
│      Message Codec Layer           │  <- MessageCodec, JSONCodec, ProtobufCodec
├─────────────────────────────────────┤
│      Frame Processing Layer        │  <- FrameCodec, FramedMessageCodec
├─────────────────────────────────────┤
│        Network Layer               │  <- TCP, WebSocket
└─────────────────────────────────────┘
```

### 组件关系 (Component Relationships)

```
FramedMessageCodec
├── FrameCodec (处理帧)
└── MessageCodec (处理消息编解码)
    ├── JSONCodec
    └── ProtobufCodec
```

## 使用示例 (Usage Examples)

### 新的方式 (New Approach)

```go
// 创建帧编解码器
frameCodec := NewFrameCodec(conn)

// 选择消息编解码器
var messageCodec MessageCodec
if useProtobuf {
    messageCodec = &ProtobufCodec{}
} else {
    messageCodec = &JSONCodec{}
}

// 组合成完整的编解码器
framedCodec := NewFramedMessageCodec(frameCodec, messageCodec)

// 编码和发送消息
envelope := &Envelope{Type: "text", Payload: payload}
err := framedCodec.Encode(envelope)

// 接收和解码消息
var envelope Envelope
err := framedCodec.Decode(&envelope, maxSize)
```

### 向后兼容方式 (Backward Compatible)

```go
// 旧代码继续工作
frameCodec := NewFrameCodec(conn)
legacyCodec := NewLegacyFrameCodec(frameCodec)

// 使用旧的接口
err := legacyCodec.Encode(envelope)
err := legacyCodec.Decode(&envelope, maxSize)
```

## 测试覆盖 (Test Coverage)

### 新增测试 (New Tests)

1. `TestFramedMessageCodec_JSONEncoding`: 测试 JSON 编解码
2. `TestFramedMessageCodec_ProtobufEncoding`: 测试 Protobuf 编解码  
3. `TestFramedMessageCodec_ContentType`: 测试内容类型正确性

### 现有测试 (Existing Tests)

- 保持所有现有测试通过
- 更新测试使用新的 `LegacyFrameCodec`

## 性能影响 (Performance Impact)

### 优势 (Benefits)

1. **更好的内存管理**: 分离了帧处理和消息处理
2. **减少重复编码**: 避免了双重序列化
3. **支持流式处理**: 适合大消息处理

### 开销 (Overhead)

- 增加了一层抽象，但开销微乎其微
- 通过接口调用的轻微性能损失

## 迁移指南 (Migration Guide)

### 对现有代码的影响 (Impact on Existing Code)

1. **无需立即修改**: 通过 `LegacyFrameCodec` 保持兼容
2. **推荐迁移路径**: 
   - 第一步: 使用 `FramedMessageCodec` 替换直接使用 `FrameCodec`
   - 第二步: 根据需要选择具体的 `MessageCodec` 实现

### 废弃计划 (Deprecation Plan)

- `LegacyFrameCodec` 标记为废弃，但会保留一段时间
- 建议新代码使用 `FramedMessageCodec`

## 总结 (Summary)

本次重构成功解决了以下问题：

1. ✅ **分离关注点**: 每个文件职责明确
2. ✅ **移除硬编码**: 支持多种编解码格式
3. ✅ **提升可扩展性**: 易于添加新功能
4. ✅ **保持兼容性**: 现有代码无需修改
5. ✅ **完善测试**: 确保代码质量

重构后的代码更加模块化，更容易维护和扩展，同时保持了良好的性能和向后兼容性。