package apidoc

import (
	"os"
	"path/filepath"
	"strings"
)

// Config 配置结构体
type Config struct {
	// 项目根路径
	ProjectRoot string `json:"project_root"`
	// 扫描的目录列表
	ScanDirs []string `json:"scan_dirs"`
	// 需要包含的文件模式
	IncludePatterns []string `json:"include_patterns"`
	// 需要排除的文件模式
	ExcludePatterns []string `json:"exclude_patterns"`
	// 文档标题
	Title string `json:"title"`
	// 文档版本
	Version string `json:"version"`
	// 文档描述
	Description string `json:"description"`
	// 基础路径
	BasePath string `json:"base_path"`
	// 是否启用调试模式
	Debug bool `json:"debug"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	workDir, _ := os.Getwd()
	return &Config{
		ProjectRoot:     workDir,
		ScanDirs:        []string{".", "./handlers", "./controllers", "./api"},
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"*_test.go", "vendor/*", ".git/*"},
		Title:           "API Documentation",
		Version:         "1.0.0",
		Description:     "自动生成的API接口文档",
		BasePath:        "/api/v1",
		Debug:           false,
	}
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return DefaultConfig()
}

// SetProjectRoot 设置项目根路径
func (c *Config) SetProjectRoot(root string) *Config {
	c.ProjectRoot = root
	return c
}

// SetScanDirs 设置扫描目录
func (c *Config) SetScanDirs(dirs ...string) *Config {
	c.ScanDirs = dirs
	return c
}

// AddScanDir 添加扫描目录
func (c *Config) AddScanDir(dir string) *Config {
	c.ScanDirs = append(c.ScanDirs, dir)
	return c
}

// SetTitle 设置文档标题
func (c *Config) SetTitle(title string) *Config {
	c.Title = title
	return c
}

// SetVersion 设置文档版本
func (c *Config) SetVersion(version string) *Config {
	c.Version = version
	return c
}

// SetDescription 设置文档描述
func (c *Config) SetDescription(desc string) *Config {
	c.Description = desc
	return c
}

// SetBasePath 设置基础路径
func (c *Config) SetBasePath(basePath string) *Config {
	c.BasePath = basePath
	return c
}

// EnableDebug 启用调试模式
func (c *Config) EnableDebug() *Config {
	c.Debug = true
	return c
}

// shouldIncludeFile 判断文件是否应该被包含
func (c *Config) shouldIncludeFile(filePath string) bool {
	// 检查排除模式
	for _, pattern := range c.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return false
		}
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "*")) {
			return false
		}
	}

	// 检查包含模式
	for _, pattern := range c.IncludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
	}

	return false
}

// getAbsolutePath 获取绝对路径
func (c *Config) getAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.ProjectRoot, path)
}