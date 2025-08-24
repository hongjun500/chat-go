# Chat-Go 使用指南

## 快速开始

### 启动服务器

#### 使用 JSON 编码 (默认)
```bash
# 启动服务器
go run cmd/server/main.go

# 或者使用环境变量明确指定
export CHAT_TCP_CODEC=json
export CHAT_WS_CODEC=json
go run cmd/server/main.go
```

#### 使用 Protobuf 编码
```bash
# TCP 使用 Protobuf，WebSocket 使用 JSON
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
go run cmd/server/main.go

# 全部使用 Protobuf
export CHAT_TCP_CODEC=protobuf  
export CHAT_WS_CODEC=protobuf
go run cmd/server/main.go
```

#### 自定义端口和配置
```bash
export CHAT_TCP_ADDR=:9090
export CHAT_WS_ADDR=:9091
export CHAT_HTTP_ADDR=:9092
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
export CHAT_READ_TIMEOUT=30
export CHAT_WRITE_TIMEOUT=10
export CHAT_MAX_FRAME=2097152
go run cmd/server/main.go
```

### 使用测试客户端

#### 连接到 JSON 服务器
```bash
# 构建客户端
go build cmd/client/main.go -o client

# 连接到服务器
./client -addr localhost:8080 -codec json -user alice
```

#### 连接到 Protobuf 服务器  
```bash
# 如果服务器使用 Protobuf 编码
./client -addr localhost:8080 -codec protobuf -user bob
```

## 架构验证

### 验证不同编码格式工作正常

1. **启动混合编码服务器**:
```bash
export CHAT_TCP_CODEC=protobuf
export CHAT_WS_CODEC=json
go run cmd/server/main.go
```

2. **测试 TCP + Protobuf**:
```bash
./client -addr localhost:8080 -codec protobuf -user alice
```

3. **测试 WebSocket + JSON** (在另一个终端):
```bash
# 使用 curl 或 WebSocket 客户端连接到 :8081/ws
```

### 性能对比测试

#### JSON 编码性能
```bash
# 启动 JSON 服务器
export CHAT_TCP_CODEC=json
go run cmd/server/main.go

# 运行多个客户端测试
for i in {1..5}; do
  ./client -addr localhost:8080 -codec json -user user$i &
done
wait
```

#### Protobuf 编码性能
```bash
# 启动 Protobuf 服务器
export CHAT_TCP_CODEC=protobuf
go run cmd/server/main.go

# 运行多个客户端测试
for i in {1..5}; do
  ./client -addr localhost:8080 -codec protobuf -user user$i &
done
wait
```

## 消息格式示例

### JSON 格式消息

#### 设置昵称
```json
{
  "type": "set_name",
  "mid": "set-name-001",
  "ts": 1697123456789,
  "payload": {
    "name": "alice"
  }
}
```

#### 聊天消息
```json
{
  "type": "chat", 
  "mid": "chat-001",
  "from": "alice",
  "ts": 1697123456789,
  "payload": {
    "content": "Hello everyone!"
  }
}
```

#### 命令消息
```json
{
  "type": "command",
  "mid": "cmd-001", 
  "ts": 1697123456789,
  "payload": {
    "raw": "/help"
  }
}
```

#### 心跳消息
```json
{
  "type": "ping",
  "mid": "ping-001",
  "ts": 1697123456789,
  "payload": {
    "seq": 1,
    "timestamp": 1697123456789
  }
}
```

### Protobuf 格式

Protobuf 消息使用二进制格式，无法直接以文本形式显示，但包含相同的字段信息。

## 开发指南

### 添加新的消息类型

1. **在 `protocol/payloads.go` 中定义负载结构**:
```go
type MyCustomPayload struct {
    CustomField string `json:"custom_field"`
    Data        []byte `json:"data"`
}
```

2. **在 `protocol/envelope.go` 中添加消息类型常量**:
```go
const (
    // ... 现有类型
    MsgCustom MessageType = "custom"
)
```

3. **在 `transport/gateway.go` 中处理新消息类型**:
```go
case "custom":
    var p protocol.MyCustomPayload
    if err := json.Unmarshal(m.Payload, &p); err != nil {
        // 错误处理
        return
    }
    // 处理自定义逻辑
```

### 添加新的编码格式

1. **实现 `MessageCodec` 接口**:
```go
type MyCodec struct{}

func (c *MyCodec) ContentType() string { 
    return "application/my-format" 
}

func (c *MyCodec) Encode(w io.Writer, m *protocol.Envelope) error {
    // 实现编码逻辑
}

func (c *MyCodec) Decode(r io.Reader, m *protocol.Envelope, maxSize int) error {
    // 实现解码逻辑
}
```

2. **在 `codec/codec.go` 中注册新编码器**:
```go
func NewCodec(codecType string) (MessageCodec, error) {
    switch codecType {
    case "json":
        return &JSONCodec{}, nil
    case "protobuf":
        return &ProtobufCodec{}, nil
    case "myformat":
        return &MyCodec{}, nil
    default:
        return nil, fmt.Errorf("unsupported codec type: %s", codecType)
    }
}
```

## 故障排除

### 常见问题

#### 连接被拒绝
```
Error: dial tcp: connect: connection refused
```
**解决方案**: 确保服务器正在运行，检查端口配置。

#### 编码格式不匹配
```
Error: json decode: invalid character
```
**解决方案**: 确保客户端和服务器使用相同的编码格式。

#### 帧大小超限
```
Error: frame too large
```
**解决方案**: 增加 `CHAT_MAX_FRAME` 环境变量的值。

### 调试技巧

1. **启用详细日志**:
```bash
export CHAT_LOG_LEVEL=debug
go run cmd/server/main.go
```

2. **使用网络工具监控**:
```bash
# 监控 TCP 连接
netstat -an | grep :8080

# 使用 tcpdump 捕获数据包
sudo tcpdump -i lo port 8080
```

3. **测试不同编码格式的性能**:
```bash
# 运行性能测试
go test -bench=. ./internal/transport/
```

## 最佳实践

### 生产环境配置

1. **设置合适的超时时间**:
```bash
export CHAT_READ_TIMEOUT=60    # 60秒读取超时
export CHAT_WRITE_TIMEOUT=30   # 30秒写入超时
```

2. **设置合适的帧大小**:
```bash
export CHAT_MAX_FRAME=4194304  # 4MB 最大帧大小
```

3. **启用 Redis 分布式支持**:
```bash
export CHAT_REDIS_ENABLE=true
export CHAT_REDIS_ADDR=redis:6379
export CHAT_REDIS_STREAM=chat_prod
export CHAT_REDIS_GROUP=chat_group_prod
```

### 编码格式选择建议

- **Web 应用**: WebSocket + JSON
- **移动应用**: TCP + Protobuf
- **服务器间通信**: TCP + Protobuf
- **开发调试**: JSON (便于调试)
- **高性能场景**: Protobuf (更小的包体积，更快的解析)

### 监控和运维

使用内置的 HTTP 监控端点:
```bash
# 访问监控指标
curl http://localhost:8082/metrics

# 访问健康检查
curl http://localhost:8082/health
```