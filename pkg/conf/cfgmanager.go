package config

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/omeyang/practices/internal/entity"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// CfgLoader 接口的定义需要根据你的实际需求来实现。
type CfgLoader interface {
	LoadConfig(ctx context.Context) (*entity.AppConf, error)
	GetConfigPath() string
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           // 最大重试次数
	Timeout     time.Duration // 超时时间
}

// CfgManager 管理配置加载和监听配置变化，以及通知其他部分应用程序的错误。
type CfgManager struct {
	loader      CfgLoader            // 配置加载器
	config      atomic.Value         // 原子值用来存储配置
	configChan  chan *entity.AppConf // 配置通道
	errorChan   chan error           // 错误通道
	watcher     WatcherInterface     // 配置监听器 使用接口
	rwMutex     sync.RWMutex         // 读写锁 用于保护配置在加载和更新时的并发访问
	once        sync.Once            // 用于确保只初始化一次
	logger      *zap.Logger          // 日志
	retryPolicy RetryPolicy          // 重试策略
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(loader CfgLoader, watcher WatcherInterface, logger *zap.Logger, retryPolicy RetryPolicy) *CfgManager {
	return &CfgManager{
		loader:      loader,
		configChan:  make(chan *entity.AppConf, 1),
		errorChan:   make(chan error, 1),
		watcher:     watcher,
		logger:      logger,
		retryPolicy: retryPolicy,
	}
}

// GetConfig 获取当前的配置
func (cm *CfgManager) GetConfig() *entity.AppConf {
	cm.rwMutex.RLock()
	defer cm.rwMutex.RUnlock()
	return cm.config.Load().(*entity.AppConf)
}

// Init 初始化配置加载和更新机制
func (cm *CfgManager) Init(ctx context.Context) error {
	var initErr error
	cm.once.Do(func() {
		initErr = cm.loadAndWatchConfig(ctx)
	})
	return initErr
}

// loadAndWatchConfig 加载并监听配置的变化
func (cm *CfgManager) loadAndWatchConfig(ctx context.Context) error {
	newConfig, err := cm.loader.LoadConfig(ctx)
	if err != nil {
		cm.logger.Error("Failed to load initial config", zap.Error(err))
		return err
	}
	cm.config.Store(newConfig)

	configPath := cm.loader.GetConfigPath()
	if err := cm.watcher.Add(configPath); err != nil {
		cm.logger.Error("Failed to watch config file", zap.String("path", configPath), zap.Error(err))
		return err
	}

	go cm.handleFSNotify(ctx)

	return nil
}

// handleFSNotify 处理配置系统通知事件
func (cm *CfgManager) handleFSNotify(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			cm.cleanupWatcher()
			return
		case event, ok := <-cm.watcher.Events():
			if !ok {
				cm.logger.Info("Config watcher events channel closed")
				return
			}
			cm.processFSNotifyEvent(ctx, event)
		case err, ok := <-cm.watcher.Errors():
			if !ok {
				cm.logger.Info("Config watcher errors channel closed")
				return
			}
			cm.logger.Error("Watcher error", zap.Error(err))
		}
	}
}

// processFSNotifyEvent 处理配置系统通知事件
func (cm *CfgManager) processFSNotifyEvent(ctx context.Context, event fsnotify.Event) {
	if event.Op&fsnotify.Write == fsnotify.Write {
		cm.reloadConfig(ctx)
	}
}

// reloadConfig 重新加载配置
func (cm *CfgManager) reloadConfig(ctx context.Context) {
	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()

	var err error
	for attempt := 1; attempt <= cm.retryPolicy.MaxAttempts; attempt++ {
		newConfig, loadErr := cm.loader.LoadConfig(ctx)
		if loadErr == nil {
			cm.config.Store(newConfig)
			cm.logger.Info("Config reloaded", zap.String("configPath", cm.loader.GetConfigPath()))
			return
		}
		err = loadErr
		cm.logger.Error("Error reloading config, retrying...", zap.Error(err), zap.Int("attempt", attempt), zap.String("configPath", cm.loader.GetConfigPath()))
		time.Sleep(cm.retryPolicy.Timeout)
	}

	cm.errorChan <- err // Notify other parts of the application
	cm.logger.Error("Failed to reload config after retries", zap.Error(err), zap.String("configPath", cm.loader.GetConfigPath()))
}

// cleanupWatcher 清理配置监听器
func (cm *CfgManager) cleanupWatcher() {
	cm.logger.Info("Config watcher stopped", zap.String("configPath", cm.loader.GetConfigPath()))
	if cm.watcher != nil {
		err := cm.watcher.Close()
		if err != nil {
			cm.logger.Error("Failed to close watcher", zap.Error(err))
		}
		cm.watcher = nil
	}
}

// AddWatcher 添加配置监听器
func (cm *CfgManager) AddWatcher(filePath string) error {
	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()
	if cm.watcher == nil {
		return errors.New("watcher not initialized")
	}
	return cm.watcher.Add(filePath)
}

// RemoveWatcher 移除监听器
func (cm *CfgManager) RemoveWatcher(filePath string) error {
	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()
	if cm.watcher == nil {
		return errors.New("watcher not initialized")
	}
	return cm.watcher.Remove(filePath)
}

// ListenForConfigErrors 监听配置错误
func (cm *CfgManager) ListenForConfigErrors() <-chan error {
	return cm.errorChan
}
