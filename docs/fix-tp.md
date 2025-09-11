# 协议层与传输层分层设计（最小可用版）

## 目标与规则

- 目标：客户端 TCP 建连后，服务端发送一条“欢迎消息”。
- 硬性规则：
  - 仅 `protocol.MessageCodec` 负责把内存中的消息（`Envelope`）序列化/反序列化为线上的字节流（JSON/Protobuf 等）。
  - 其他任何地方（Gateway/Transport/业务）不得直接使用 JSON/Protobuf API。
  - 业务层对 `Envelope.Data` 只看作“无格式字节序列”。是否有更高层“payload 编解码”，由 `protocol` 包内部提供，不向外泄露实现细节。

## 分层概览

- protocol（协议层）
  - 核心：`Envelope`、`MessageType`、`Encoding`。
  - 接口：`MessageCodec`（`Encode/Decode`），实现：`JSONCodec`、`ProtobufCodec`。
  - 分发/便捷：`Protocol`（如 `Welcome()`），后续可扩展 handler 注册。
  - 关键点：只有这里可以 import `encoding/json` 或 protobuf，外部一律禁止。

- transport（传输层）
  - 接口：`Session`（抽象连接）、`Gateway`（业务入口）、`Transport`（TCP/WS 实现）。
  - TCP：`FrameCodec` 做长度前缀帧；读取到一帧字节后，交给 `MessageCodec.Decode` 得到 `Envelope`；回调 `Gateway.OnEnvelope`。
  - 发送：`Gateway` 产出 `Envelope` → `Session.SendEnvelope` → `MessageCodec.Encode` → `FrameCodec.WriteFrame`。
  - 关键点：Transport 不处理 JSON/Proto，不理解 `Envelope.Data` 的语义。

- gateway（业务入口）
  - 生命周期：`OnSessionOpen` / `OnEnvelope` / `OnSessionClose`。
  - 最小欢迎：`OnSessionOpen` 调 `dispatcher.Welcome(sess)`，由 dispatcher 使用 protocol 生成欢迎 `Envelope` 并通过 `sess.SendEnvelope` 发出。
  - 关键点：Gateway 不做 payload 解析（不使用 JSON/Proto），对 `Data` 保持“无知”。

## 数据模型（最小可用约定）

- `Envelope`（protocol/envelope.go）：
  - 元字段：`Version`、`Type`、`Encoding`、`Mid/Correlation/From/To/Ts`。
  - 负载：`Data []byte`（对上层是字节序列，不解释格式）。
- 欢迎消息：`Type = text`，`Data = []byte("请输入昵称并回车：")`（UTF-8 文本字节）。

## 欢迎消息工作流

1. TCP 建连后：`TCPServer.serveConn` 创建 `sess`，调用 `gateway.OnSessionOpen(sess)`。
2. Gateway：`GatewayHandler.OnSessionOpen` 调 `dispatcher.Welcome(sess)`。
3. Dispatcher：`EnvelopeDispatcher.Welcome(sess)` 使用 `protocol.DefaultProtocol.Welcome("请输入昵称并回车：")` 构造 `Envelope`。
4. 发送：`sess.SendEnvelope(envelope)` → `MessageCodec.Encode` → `FrameCodec.WriteFrame`。

注意：全流程只有 `MessageCodec` 使用 JSON/Protobuf；其他层均以 `[]byte` 视之。

## 本次最小修改

仅为保证“欢迎消息发出后连接保持存活”，进行两处最小变更：

1) 初始化 `closeChan`

- 文件：`internal/transport/tcp.go`
- 位置：`serveConn` 构造 `sess` 时加入 `closeChan: make(chan struct{})`，避免 `Close()` 时对 nil 通道 `close()` 导致 panic。

2) 不要在 writer goroutine 里主动关闭连接

- 文件：`internal/transport/tcp.go`
- 位置：writer goroutine 末尾将 `_ = conn.Close()` 改为阻塞等待 `<-sess.closeChan`，让连接保持存活，交由生命周期统一关闭。

以上变更不触碰 JSON/Protobuf，不改变分层职责，仅完善连接生命周期以满足“最小可用欢迎消息”。

## 线协议选择：编码整包 vs 仅编码 Data

结论：编码“整包 Envelope”。

- 理由：
  - Envelope 承载必要的元信息（类型、时间戳、路由字段等），只编码 `Data` 会丢失类型/时间等上下文，破坏前后兼容与扩展性。
  - `MessageCodec` 的职责就是对“整包”做序列化/反序列化；传输层只处理字节帧，不窥视内部结构。
  - Protobuf/JSON 均可对整包建模；`Data` 作为字节字段在两种编码下都能被一致承载。

- JSONCodec 下为何看到 Base64：
  - Go `encoding/json` 对 `[]byte` 字段默认以 base64 字符串表示。这就是你在 telnet 里看到的：
    `{"version":"","type":"text",...,"data":"6K+36L6T5YWl5pi156ew5bm25Zue6L2m77ya"}`。
  - 该 base64 解码后就是原始 UTF-8 文本“请输入昵称并回车：”。
  - 这对“协议感知”的客户端完全合理：客户端应先用同一 `MessageCodec` 解出 Envelope，再把 `Data` 当字节处理/展示。

- 兼顾可读性的可选方案（仍然只在 MessageCodec 层实现）：
  - 方案 A：新增一个 JSON 文本友好编解码器（如 `JSONTextCodec`），在 JSON 表示时将 `Data` 以普通字符串字段（如 `data_text`）呈现，而不是 base64；解码时反向还原到 `[]byte`。这不改变分层约束。
  - 方案 B：提供一个简单 CLI 客户端/调试器，使用相同 `MessageCodec` 解码并“人类可读”地打印 Envelope（同时展示 `Data` 的文本预览）。
  - 默认与推荐：保持“整包编码”的标准做法，业务/测试侧使用 Codec 感知的客户端进行交互；仅在需要人工调试时使用 A/B 之一。

## 未来演进（仍遵守“编码仅在 MessageCodec”）

- 实现 TCP 读循环：`ReadFrame` → `MessageCodec.Decode` → `Gateway.OnEnvelope`；不在外层解析 `Data`。
- 若需结构化 payload：在 `protocol` 包内部提供 payload 级编解码（根据 `Envelope.Type` 对 `Data` 进行 encode/decode），外部只见结构体，不见 JSON/Proto。
- 通过配置在 `main` 注入具体 `MessageCodec` 到 `Protocol`/`Transport`，保持可替换性与统一性。

## 验证建议

- 运行：`go run ./cmd/server`
- 客户端：使用与服务端一致的 `MessageCodec` 解码首帧 Envelope，再展示 `Data` 为文本，即可看到“请输入昵称并回车：”。
- 直接 telnet 仅作“看原始字节”用途；JSONCodec 下 `data` 为 base64 是预期行为。

