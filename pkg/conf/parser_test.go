package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func TestNewParserNoLogger(t *testing.T) {
	_, err := NewParser(".json", nil)
	if err == nil {
		t.Errorf("Expected an error when logger is nil, but got nil")
	}
}

// 用于模拟文件内容
func mockFile(content string) afero.File {
	fs := afero.NewMemMapFs()

	// 创建一个新文件
	file, err := fs.Create("test")
	if err != nil {
		panic("Failed to create mock file: " + err.Error())
	}

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		panic("Failed to write to mock file: " + err.Error())
	}

	// 重置文件指针到文件开始位置
	_, err = file.Seek(0, 0)
	if err != nil {
		panic("Failed to seek mock file: " + err.Error())
	}

	return file
}

// TestNewParser 测试 NewParser 函数
func TestNewParser(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name           string
		fileExtension  string
		expectedParser string
		expectError    bool
	}{
		{"JSON Parser", ".json", "*config.JSONParser", false},
		{"YAML Parser", ".yaml", "*config.YAMLParser", false},
		{"Unsupported Extension", ".xml", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(tt.fileExtension, logger)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parser)
				assert.Equal(t, tt.expectedParser, fmt.Sprintf("%T", parser))
			}
		})
	}
}

// TestJSONParser_Parse 测试 JSONParser 的 Parse 方法
func TestJSONParser_Parse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	parser := &JSONParser{Logger: logger}

	validJSON := `{"key": "value"}`
	invalidJSON := `{key: "value"}`

	tests := []struct {
		name        string
		fileContent string
		expectError bool
	}{
		{"Valid JSON", validJSON, false},
		{"Invalid JSON", invalidJSON, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := mockFile(tt.fileContent)
			_, err := parser.Parse(file)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestYAMLParser_Parse 测试 YAMLParser 的 Parse 方法
func TestYAMLParser_Parse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	parser := &YAMLParser{Logger: logger}

	validYAML := `key: value`
	invalidYAML := `: invalidYAML`

	tests := []struct {
		name        string
		fileContent string
		expectError bool
	}{
		{"Valid YAML", validYAML, false},
		{"Invalid YAML", invalidYAML, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := mockFile(tt.fileContent)
			_, err := parser.Parse(file)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
