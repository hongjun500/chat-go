package protocol

import (
	"fmt"
	"sync"
)

// ProtocolManager 协议管理器，集成动态配置、版本控制和消息工厂
type ProtocolManager struct {
	mu                sync.RWMutex
	config            CodecConfig
	versionController *VersionController
	messageFactory    *MessageFactory
	protocol          *VersionedProtocol
	defaultVersion    string
}

// NewProtocolManager 创建协议管理器
func NewProtocolManager(defaultCodec int, defaultVersion string) (*ProtocolManager, error) {
	config := NewDefaultCodecConfig(defaultCodec)
	versionController := NewVersionController(defaultVersion)
	protocol := NewVersionedProtocol(defaultCodec, defaultVersion)
	
	// 创建默认编码器
	codec, err := NewCodec(defaultCodec)
	if err != nil {
		return nil, fmt.Errorf("failed to create default codec: %w", err)
	}
	
	messageFactory := NewMessageFactoryWithVersionControl(codec, config, versionController)
	
	// 注册默认版本
	err = versionController.RegisterVersion(defaultVersion, CodecMapping{
		DefaultCodec: defaultCodec,
		Codecs:       make(map[MessageType]int),
	}, HandlerMapping{
		DefaultHandler: protocol.Protocol.textHandler,
		Handlers:       make(map[MessageType]MessageHandler),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to register default version: %w", err)
	}
	
	return &ProtocolManager{
		config:            config,
		versionController: versionController,
		messageFactory:    messageFactory,
		protocol:          protocol,
		defaultVersion:    defaultVersion,
	}, nil
}

// SetDefaultCodec 动态设置默认编码器
func (pm *ProtocolManager) SetDefaultCodec(codecType int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if !pm.config.IsCodecSupported(codecType) {
		return fmt.Errorf("ProtocolManager.SetDefaultCodec: unsupported codec type: %d", codecType)
	}
	
	// 更新配置
	if err := pm.config.SetDefaultCodec(codecType); err != nil {
		return fmt.Errorf("ProtocolManager.SetDefaultCodec: failed to update config: %w", err)
	}
	
	// 更新协议编码器
	if err := pm.protocol.SetCodec(codecType); err != nil {
		return fmt.Errorf("ProtocolManager.SetDefaultCodec: failed to update protocol codec: %w", err)
	}
	
	// 更新消息工厂的默认编码器
	codec, err := NewCodec(codecType)
	if err != nil {
		return fmt.Errorf("ProtocolManager.SetDefaultCodec: failed to create new codec: %w", err)
	}
	pm.messageFactory.SetDefaultCodec(codec)
	
	return nil
}

// SetDefaultCodecByName 根据名称动态设置默认编码器
func (pm *ProtocolManager) SetDefaultCodecByName(name string) error {
	codecType, err := pm.config.GetCodecByName(name)
	if err != nil {
		return fmt.Errorf("ProtocolManager.SetDefaultCodecByName: %w", err)
	}
	return pm.SetDefaultCodec(codecType)
}

// GetCurrentCodec 获取当前编码器信息
func (pm *ProtocolManager) GetCurrentCodec() (string, int) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.protocol.GetCurrentCodec()
}

// RegisterHandler 注册消息处理器
func (pm *ProtocolManager) RegisterHandler(msgType MessageType, handler MessageHandler) {
	pm.protocol.RegisterHandler(msgType, handler)
}

// RegisterVersionedHandler 注册版本特定的消息处理器
func (pm *ProtocolManager) RegisterVersionedHandler(version string, msgType MessageType, handler MessageHandler, codecType int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	return pm.versionController.UpdateVersionHandler(version, msgType, handler, codecType)
}

// Dispatch 分发消息
func (pm *ProtocolManager) Dispatch(env *Envelope) error {
	return pm.protocol.DispatchVersioned(env)
}

// CreateTextMessage 创建文本消息
func (pm *ProtocolManager) CreateTextMessage(text, from, to string) (*Envelope, error) {
	return pm.messageFactory.CreateTextMessage(text, from, to)
}

// CreateChatMessage 创建聊天消息
func (pm *ProtocolManager) CreateChatMessage(content, from string) (*Envelope, error) {
	return pm.messageFactory.CreateChatMessage(content, from)
}

// CreateAckMessage 创建确认消息
func (pm *ProtocolManager) CreateAckMessage(status, correlationID string) (*Envelope, error) {
	return pm.messageFactory.CreateAckMessage(status, correlationID)
}

// CreateCommandMessage 创建命令消息
func (pm *ProtocolManager) CreateCommandMessage(command, from string) (*Envelope, error) {
	return pm.messageFactory.CreateCommandMessage(command, from)
}

// EncodeMessage 编码消息
func (pm *ProtocolManager) EncodeMessage(env *Envelope) ([]byte, error) {
	return pm.messageFactory.EncodeMessage(env)
}

// DecodeMessage 解码消息
func (pm *ProtocolManager) DecodeMessage(data []byte, maxSize int) (*Envelope, error) {
	return pm.messageFactory.DecodeMessage(data, maxSize)
}

// DecodeMessageWithCodec 使用指定编码器解码消息
func (pm *ProtocolManager) DecodeMessageWithCodec(data []byte, codecType int, maxSize int) (*Envelope, error) {
	return pm.messageFactory.DecodeMessageWithCodec(data, codecType, maxSize)
}

// GetSupportedVersions 获取支持的版本列表
func (pm *ProtocolManager) GetSupportedVersions() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.versionController.GetSupportedVersions()
}

// SetDefaultVersion 设置默认版本
func (pm *ProtocolManager) SetDefaultVersion(version string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if err := pm.versionController.SetDefaultVersion(version); err != nil {
		return fmt.Errorf("ProtocolManager.SetDefaultVersion: %w", err)
	}
	
	pm.defaultVersion = version
	return nil
}

// GetDefaultVersion 获取默认版本
func (pm *ProtocolManager) GetDefaultVersion() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.versionController.GetDefaultVersion()
}

// IsVersionSupported 检查版本是否支持
func (pm *ProtocolManager) IsVersionSupported(version string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.versionController.IsVersionSupported(version)
}

// GetConfig 获取编码器配置
func (pm *ProtocolManager) GetConfig() CodecConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.config
}

// GetProtocol 获取协议实例
func (pm *ProtocolManager) GetProtocol() *VersionedProtocol {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.protocol
}

// GetMessageFactory 获取消息工厂
func (pm *ProtocolManager) GetMessageFactory() *MessageFactory {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.messageFactory
}

// GetVersionController 获取版本控制器
func (pm *ProtocolManager) GetVersionController() *VersionController {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.versionController
}