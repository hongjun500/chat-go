# Protocol Manager Demo

This demo showcases the new protocol management features implemented in Chat-Go.

## Features Demonstrated

1. **Dynamic Codec Switching**: Runtime switching between JSON and Protobuf encoders
2. **Message Creation**: Creating different types of messages (text, chat, ack, command)  
3. **Encoding/Decoding**: Message serialization and deserialization
4. **Version Control**: Version-specific handler registration and management
5. **Enhanced Error Handling**: Improved error messages with context
6. **Protocol Manager**: Unified API for all protocol operations

## Running the Demo

```bash
cd /path/to/chat-go
go run examples/protocol_demo/main.go
```

## Expected Output

The demo will show:
- Initial codec configuration
- Dynamic codec switching 
- Message creation and encoding/decoding
- Version control features
- Enhanced error messages
- Different message types

## Key Components Used

- `protocol.ProtocolManager`: Central management of protocol features
- `protocol.CodecConfig`: Dynamic codec configuration
- `protocol.VersionController`: Message version management
- `protocol.MessageFactory`: Standardized message creation
- Enhanced error handling in all codecs