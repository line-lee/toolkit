package apidoc

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册API文档路由
func RegisterRoutes(r *gin.Engine) {
	RegisterRoutesWithConfig(r, nil)
}

// RegisterRoutesWithConfig 使用指定配置注册API文档路由
func RegisterRoutesWithConfig(r *gin.Engine, config *Config) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建渲染器
	renderer := NewRenderer()

	// HTML 文档界面
	r.GET("/docs", func(c *gin.Context) {
		// 转换为完整的API文档
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		err = renderer.RenderHTML(apiDoc, c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	// JSON 格式文档
	r.GET("/docs/json", func(c *gin.Context) {
		_, err := GenerateDocsWithConfig(config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 转换为完整的API文档
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiDoc.ToLegacyDoc())
	})

	// 完整的API文档 JSON 格式
	r.GET("/docs/api", func(c *gin.Context) {
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiDoc)
	})

	// 健康检查
	r.GET("/docs/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API文档服务运行正常",
		})
	})
}

// RegisterRoutesWithPrefix 使用指定前缀注册路由
func RegisterRoutesWithPrefix(r *gin.Engine, prefix string, config *Config) {
	if config == nil {
		config = DefaultConfig()
	}

	// 确保前缀以 / 开头但不以 / 结尾
	prefix = strings.TrimRight(prefix, "/")
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	renderer := NewRenderer()

	r.GET(prefix+"/docs", func(c *gin.Context) {
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		err = renderer.RenderHTML(apiDoc, c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	r.GET(prefix+"/docs/json", func(c *gin.Context) {
		_, err := GenerateDocsWithConfig(config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 转换为完整的API文档
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiDoc.ToLegacyDoc())
	})

	r.GET(prefix+"/docs/api", func(c *gin.Context) {
		parser := NewParser(config)
		apiDoc, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiDoc)
	})
}

// SetupDocsServer 设置文档服务器（独立服务）
func SetupDocsServer(port string, config *Config) *gin.Engine {
	if config == nil {
		config = DefaultConfig()
	}

	r := gin.Default()
	RegisterRoutesWithConfig(r, config)

	return r
}
