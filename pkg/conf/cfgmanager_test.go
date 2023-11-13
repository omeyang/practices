package config

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omeyang/practices/internal/entity"
	mocks "github.com/omeyang/practices/mocks/conf"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewConfigManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建模拟对象
	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	// 定义重试策略
	retryPolicy := RetryPolicy{
		MaxAttempts: 3,
		Timeout:     2 * time.Second,
	}

	// 调用 NewConfigManager
	cm := NewConfigManager(mockLoader, mockWatcher, logger, retryPolicy)

	// 验证返回的 CfgManager 是否正确
	assert.NotNil(t, cm, "ConfigManager should not be nil")
	assert.Equal(t, mockLoader, cm.loader, "Loader should be set correctly")
	assert.Equal(t, mockWatcher, cm.watcher, "Watcher should be set correctly")
	assert.Equal(t, logger, cm.logger, "Logger should be set correctly")
	assert.Equal(t, retryPolicy, cm.retryPolicy, "RetryPolicy should be set correctly")
	assert.NotNil(t, cm.configChan, "Config channel should be initialized")
	assert.NotNil(t, cm.errorChan, "Error channel should be initialized")
}

// TestCfgManager_Init 测试配置管理器的初始化
func TestCfgManager_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	// 创建用于模拟事件和错误的通道
	mockEvents := make(chan fsnotify.Event, 1)
	mockErrors := make(chan error, 1)

	// 设置模拟对象的期望行为
	mockLoader.EXPECT().LoadConfig(gomock.Any()).Return(&entity.AppConf{}, nil).Times(1)
	mockLoader.EXPECT().GetConfigPath().Return("/path/to/config").AnyTimes()
	mockWatcher.EXPECT().Add("/path/to/config").Return(nil).Times(1)
	mockWatcher.EXPECT().Events().Return(mockEvents).AnyTimes()
	mockWatcher.EXPECT().Errors().Return(mockErrors).AnyTimes()

	cm := NewConfigManager(mockLoader, mockWatcher, logger, RetryPolicy{
		MaxAttempts: 3,
		Timeout:     2 * time.Second,
	})

	// 执行测试
	err := cm.Init(context.Background())

	// 验证结果
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}

}

func TestCfgManager_AddWatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	cm := NewConfigManager(mockLoader, mockWatcher, logger, RetryPolicy{
		MaxAttempts: 3,
		Timeout:     2 * time.Second,
	})

	mockWatcher.EXPECT().Add("/additional/path").Return(nil).Times(1)

	err := cm.AddWatcher("/additional/path")
	if err != nil {
		t.Fatalf("AddWatcher failed: %v", err)
	}
}

func TestCfgManager_RemoveWatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	cm := NewConfigManager(mockLoader, mockWatcher, logger, RetryPolicy{
		MaxAttempts: 3,
		Timeout:     2 * time.Second,
	})

	mockWatcher.EXPECT().Remove("/additional/path").Return(nil).Times(1)

	err := cm.RemoveWatcher("/additional/path")
	if err != nil {
		t.Fatalf("RemoveWatcher failed: %v", err)
	}
}

func TestCfgManager_ListenForConfigErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	cm := NewConfigManager(mockLoader, mockWatcher, logger, RetryPolicy{
		MaxAttempts: 3,
		Timeout:     2 * time.Second,
	})

	errorChan := cm.ListenForConfigErrors()
	if errorChan == nil {
		t.Fatal("Expected non-nil error channel")
	}
}

func TestCfgManager_reloadConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoader := mocks.NewMockCfgLoader(ctrl)
	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	logger, _ := zap.NewDevelopment()

	configPath := "/path/to/config"
	mockLoader.EXPECT().GetConfigPath().Return(configPath).AnyTimes()

	cm := NewConfigManager(mockLoader, mockWatcher, logger, RetryPolicy{
		MaxAttempts: 3,
		Timeout:     50 * time.Millisecond,
	})

	ctx := context.Background()

	// 测试成功加载配置
	mockLoader.EXPECT().LoadConfig(ctx).Return(&entity.AppConf{}, nil).Times(1)
	cm.reloadConfig(ctx)

	// 测试加载配置失败，但在重试中成功
	mockLoader.EXPECT().LoadConfig(ctx).Return(nil, errors.New("load error")).Times(2)
	mockLoader.EXPECT().LoadConfig(ctx).Return(&entity.AppConf{}, nil).Times(1)
	cm.reloadConfig(ctx)

	// 测试加载配置失败，重试次数耗尽
	mockLoader.EXPECT().LoadConfig(ctx).Return(nil, errors.New("load error")).Times(3)
	cm.reloadConfig(ctx)
}
