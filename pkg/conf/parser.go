package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/omeyang/practices/internal/entity"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// CfgParser 配置解析器
type CfgParser interface {
	Parse(file afero.File) (*entity.AppConf, error)
}

// JSONParser JSON配置解析器
type JSONParser struct {
	Logger *zap.Logger
}

// YAMLParser YAML配置解析器
type YAMLParser struct {
	Logger *zap.Logger
}

// NewParser 创建新的配置解析器
func NewParser(fileExtension string, logger *zap.Logger) (CfgParser, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	// Normalize file extension
	if !strings.HasPrefix(fileExtension, ".") {
		fileExtension = "." + fileExtension
	}

	switch fileExtension {
	case ".json":
		return &JSONParser{Logger: logger}, nil
	case ".yaml", ".yml":
		return &YAMLParser{Logger: logger}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", fileExtension)
	}
}

// Parse 解析json配置文件
func (j *JSONParser) Parse(file afero.File) (*entity.AppConf, error) {
	var config entity.AppConf
	err := json.NewDecoder(file).Decode(&config)
	if err != nil {
		j.Logger.Error("Failed to parse JSON config", zap.Error(err))
		return nil, fmt.Errorf("json parsing error: %w", err)
	}
	j.Logger.Info("Successfully parsed JSON config")
	return &config, nil
}

// Parse 解析yaml配置文件
func (y *YAMLParser) Parse(file afero.File) (*entity.AppConf, error) {
	var config entity.AppConf
	err := yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		y.Logger.Error("Failed to parse YAML config", zap.Error(err))
		return nil, fmt.Errorf("yaml parsing error: %w", err)
	}
	y.Logger.Info("Successfully parsed YAML config")
	return &config, nil
}
