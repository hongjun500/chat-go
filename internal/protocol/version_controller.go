package protocol

import (
	"fmt"
	"sync"
)

// VersionController 版本控制器，根据消息版本选择编码器和处理器
type VersionController struct {
	mu             sync.RWMutex
	codecMappings  map[string]CodecMapping  // 版本到编码器的映射
	handlerMappings map[string]HandlerMapping // 版本到处理器的映射
	defaultVersion string                     // 默认版本
}

// CodecMapping 版本对应的编码器映射
type CodecMapping struct {
	DefaultCodec int                     // 默认编码器
	Codecs       map[MessageType]int     // 消息类型特定编码器
}

// HandlerMapping 版本对应的处理器映射
type HandlerMapping struct {
	DefaultHandler MessageHandler                // 默认处理器
	Handlers       map[MessageType]MessageHandler // 消息类型特定处理器
}

// NewVersionController 创建版本控制器
func NewVersionController(defaultVersion string) *VersionController {
	return &VersionController{
		codecMappings:   make(map[string]CodecMapping),
		handlerMappings: make(map[string]HandlerMapping),
		defaultVersion:  defaultVersion,
	}
}

// RegisterVersion 注册版本及其编码器和处理器映射
func (vc *VersionController) RegisterVersion(version string, codecMapping CodecMapping, handlerMapping HandlerMapping) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	
	// 验证编码器映射
	if !vc.isCodecSupported(codecMapping.DefaultCodec) {
		return fmt.Errorf("unsupported default codec %d for version %s", codecMapping.DefaultCodec, version)
	}
	
	for msgType, codecType := range codecMapping.Codecs {
		if !vc.isCodecSupported(codecType) {
			return fmt.Errorf("unsupported codec %d for message type %s in version %s", codecType, msgType, version)
		}
	}
	
	vc.codecMappings[version] = codecMapping
	vc.handlerMappings[version] = handlerMapping
	
	return nil
}

// UpdateVersionHandler 更新版本特定的消息处理器
func (vc *VersionController) UpdateVersionHandler(version string, msgType MessageType, handler MessageHandler, codecType int) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	
	// 检查版本是否存在
	codecMapping, codecExists := vc.codecMappings[version]
	handlerMapping, handlerExists := vc.handlerMappings[version]
	
	if !codecExists || !handlerExists {
		// 版本不存在，创建新的映射
		codecMapping = CodecMapping{
			DefaultCodec: codecType,
			Codecs:       make(map[MessageType]int),
		}
		handlerMapping = HandlerMapping{
			DefaultHandler: nil,
			Handlers:       make(map[MessageType]MessageHandler),
		}
	}
	
	// 验证编码器类型
	if !vc.isCodecSupported(codecType) {
		return fmt.Errorf("unsupported codec %d for message type %s in version %s", codecType, msgType, version)
	}
	
	// 更新映射
	codecMapping.Codecs[msgType] = codecType
	handlerMapping.Handlers[msgType] = handler
	
	// 保存更新的映射
	vc.codecMappings[version] = codecMapping
	vc.handlerMappings[version] = handlerMapping
	
	return nil
}

// GetCodecForMessage 根据消息版本和类型获取编码器
func (vc *VersionController) GetCodecForMessage(env *Envelope) (MessageCodec, error) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	
	version := env.Version
	if version == "" {
		version = vc.defaultVersion
	}
	
	codecMapping, exists := vc.codecMappings[version]
	if !exists {
		return nil, fmt.Errorf("no codec mapping found for version %s", version)
	}
	
	// 首先检查是否有特定消息类型的编码器
	if codecType, exists := codecMapping.Codecs[env.Type]; exists {
		return NewCodec(codecType)
	}
	
	// 使用默认编码器
	return NewCodec(codecMapping.DefaultCodec)
}

// GetHandlerForMessage 根据消息版本和类型获取处理器
func (vc *VersionController) GetHandlerForMessage(env *Envelope) (MessageHandler, error) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	
	version := env.Version
	if version == "" {
		version = vc.defaultVersion
	}
	
	handlerMapping, exists := vc.handlerMappings[version]
	if !exists {
		return nil, fmt.Errorf("no handler mapping found for version %s", version)
	}
	
	// 首先检查是否有特定消息类型的处理器
	if handler, exists := handlerMapping.Handlers[env.Type]; exists {
		return handler, nil
	}
	
	// 使用默认处理器
	if handlerMapping.DefaultHandler != nil {
		return handlerMapping.DefaultHandler, nil
	}
	
	return nil, fmt.Errorf("no handler found for message type %s in version %s", env.Type, version)
}

// GetSupportedVersions 获取支持的版本列表
func (vc *VersionController) GetSupportedVersions() []string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	
	versions := make([]string, 0, len(vc.codecMappings))
	for version := range vc.codecMappings {
		versions = append(versions, version)
	}
	return versions
}

// SetDefaultVersion 设置默认版本
func (vc *VersionController) SetDefaultVersion(version string) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	
	if _, exists := vc.codecMappings[version]; !exists {
		return fmt.Errorf("version %s is not registered", version)
	}
	
	vc.defaultVersion = version
	return nil
}

// GetDefaultVersion 获取默认版本
func (vc *VersionController) GetDefaultVersion() string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.defaultVersion
}

// IsVersionSupported 检查版本是否支持
func (vc *VersionController) IsVersionSupported(version string) bool {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	_, exists := vc.codecMappings[version]
	return exists
}

// isCodecSupported 检查编码器是否支持（内部方法）
func (vc *VersionController) isCodecSupported(codecType int) bool {
	_, exists := CodecFactories[codecType]
	return exists
}

// VersionedProtocol 支持版本控制的协议
type VersionedProtocol struct {
	*Protocol
	versionController *VersionController
}

// NewVersionedProtocol 创建支持版本控制的协议
func NewVersionedProtocol(defaultCodec int, defaultVersion string) *VersionedProtocol {
	protocol := NewProtocol(defaultCodec)
	versionController := NewVersionController(defaultVersion)
	
	return &VersionedProtocol{
		Protocol:          protocol,
		versionController: versionController,
	}
}

// DispatchVersioned 根据版本分发消息
func (vp *VersionedProtocol) DispatchVersioned(env *Envelope) error {
	// 尝试使用版本特定的处理器
	handler, err := vp.versionController.GetHandlerForMessage(env)
	if err == nil {
		return handler(env)
	}
	
	// 回退到默认的分发逻辑
	return vp.Protocol.Dispatch(env)
}

// RegisterVersionMapping 注册版本映射
func (vp *VersionedProtocol) RegisterVersionMapping(version string, codecMapping CodecMapping, handlerMapping HandlerMapping) error {
	return vp.versionController.RegisterVersion(version, codecMapping, handlerMapping)
}

// GetVersionController 获取版本控制器
func (vp *VersionedProtocol) GetVersionController() *VersionController {
	return vp.versionController
}