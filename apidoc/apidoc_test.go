package apidoc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "API Documentation", config.Title)
	assert.Equal(t, "1.0.0", config.Version)
	assert.NotEmpty(t, config.ScanDirs)
	assert.NotEmpty(t, config.IncludePatterns)
	assert.NotEmpty(t, config.ExcludePatterns)
}

func TestConfigBuilder(t *testing.T) {
	config := NewConfig().
		SetTitle("Test API").
		SetVersion("2.0.0").
		SetDescription("Test Description").
		SetBasePath("/api/v2").
		EnableDebug()

	assert.Equal(t, "Test API", config.Title)
	assert.Equal(t, "2.0.0", config.Version)
	assert.Equal(t, "Test Description", config.Description)
	assert.Equal(t, "/api/v2", config.BasePath)
	assert.True(t, config.Debug)
}

func TestShouldIncludeFile(t *testing.T) {
	config := DefaultConfig()

	// 应该包含的文件
	assert.True(t, config.shouldIncludeFile("main.go"))
	assert.True(t, config.shouldIncludeFile("handler.go"))

	// 应该排除的文件
	assert.False(t, config.shouldIncludeFile("main_test.go"))
	assert.False(t, config.shouldIncludeFile("vendor/package.go"))
}

func TestNewParser(t *testing.T) {
	config := DefaultConfig()
	parser := NewParser(config)

	assert.NotNil(t, parser)
	assert.Equal(t, config, parser.config)
	assert.NotNil(t, parser.fileSet)
	assert.NotNil(t, parser.routeMap)
	assert.NotNil(t, parser.models)
}

func TestNewAPIDoc(t *testing.T) {
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	assert.NotNil(t, doc)
	assert.Equal(t, config.Title, doc.Info.Title)
	assert.Equal(t, config.Version, doc.Info.Version)
	assert.Equal(t, config.Description, doc.Info.Description)
	assert.Equal(t, config.BasePath, doc.BasePath)
	assert.NotNil(t, doc.APIs)
	assert.NotNil(t, doc.Models)
}

func TestAddAPI(t *testing.T) {
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	api := &APIInfo{
		Path:        "/test",
		Method:      "GET",
		Summary:     "Test API",
		HandlerFunc: "testHandler",
	}

	doc.AddAPI(api)

	assert.Len(t, doc.APIs, 1)
	assert.Equal(t, api, doc.APIs[0])
}

func TestAddModel(t *testing.T) {
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	model := &Model{
		Name: "User",
		Type: "object",
		Properties: map[string]*Schema{
			"id":   {Type: "integer"},
			"name": {Type: "string"},
		},
	}

	doc.AddModel(model)

	assert.Len(t, doc.Models, 1)
	assert.Equal(t, model, doc.Models[0])
}

func TestToLegacyDoc(t *testing.T) {
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	api := &APIInfo{
		Path:        "/test",
		Method:      "GET",
		Summary:     "Test API",
		HandlerFunc: "testHandler",
	}
	doc.AddAPI(api)

	legacyDoc := doc.ToLegacyDoc()

	assert.NotNil(t, legacyDoc)
	assert.Equal(t, doc.Info.Title, legacyDoc.Title)
	assert.Equal(t, doc.Info.Version, legacyDoc.Version)
	assert.Len(t, legacyDoc.APIs, 1)
	assert.Equal(t, api.Path, legacyDoc.APIs[0].Path)
	assert.Equal(t, api.Method, legacyDoc.APIs[0].Method)
	assert.Equal(t, api.Summary, legacyDoc.APIs[0].Summary)
	assert.Equal(t, []string{api.HandlerFunc}, legacyDoc.APIs[0].Handlers)
}

func TestConvertBasicTypeToSchema(t *testing.T) {
	parser := NewParser(DefaultConfig())

	tests := []struct {
		typeName string
		expected string
	}{
		{"string", "string"},
		{"int", "integer"},
		{"int32", "integer"},
		{"int64", "integer"},
		{"uint", "integer"},
		{"float32", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"unknown", "string"},
	}

	for _, test := range tests {
		t.Run(test.typeName, func(t *testing.T) {
			schema := parser.convertBasicTypeToSchema(test.typeName)
			assert.Equal(t, test.expected, schema.Type)
		})
	}
}

func TestExtractJSONTag(t *testing.T) {
	parser := NewParser(DefaultConfig())

	tests := []struct {
		tag      string
		expected string
	}{
		{`json:"name"`, "name"},
		{`json:"email,omitempty"`, "email,omitempty"},
		{`json:"-"`, "-"},
		{`binding:"required"`, ""},
		{`json:"id" binding:"required"`, "id"},
	}

	for _, test := range tests {
		t.Run(test.tag, func(t *testing.T) {
			result := parser.extractJSONTag(test.tag)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseCommentRouteInfo(t *testing.T) {
	parser := NewParser(DefaultConfig())
	routeInfo := &RouteInfo{}

	// 测试 @Router 注解
	parser.parseComment("// @Router /users/{id} [GET]", routeInfo)
	assert.Equal(t, "/users/{id}", routeInfo.Path)
	assert.Equal(t, "GET", routeInfo.Method)

	// 测试 @Summary 注解
	routeInfo = &RouteInfo{}
	parser.parseComment("// @Summary 获取用户信息", routeInfo)
	assert.Equal(t, "获取用户信息", routeInfo.Summary)

	// 测试 @Description 注解
	routeInfo = &RouteInfo{}
	parser.parseComment("// @Description 根据用户ID获取详细信息", routeInfo)
	assert.Equal(t, "根据用户ID获取详细信息", routeInfo.Description)

	// 测试 @Tags 注解
	routeInfo = &RouteInfo{}
	parser.parseComment("// @Tags 用户管理,API", routeInfo)
	assert.Equal(t, []string{"用户管理", "API"}, routeInfo.Tags)

	// 测试 @Deprecated 注解
	routeInfo = &RouteInfo{}
	parser.parseComment("// @Deprecated", routeInfo)
	assert.True(t, routeInfo.Deprecated)
}

func TestParseCommentParam(t *testing.T) {
	parser := NewParser(DefaultConfig())
	routeInfo := &RouteInfo{}

	// 测试 @Param 注解
	parser.parseComment(`// @Param id path int true "用户ID"`, routeInfo)

	assert.Len(t, routeInfo.Parameters, 1)
	param := routeInfo.Parameters[0]
	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "path", param.In)
	assert.Equal(t, "int", param.Type)
	assert.True(t, param.Required)
	assert.Equal(t, "用户ID", param.Description)
}

func TestParseCommentResponse(t *testing.T) {
	parser := NewParser(DefaultConfig())
	routeInfo := &RouteInfo{}

	// 测试 @Success 注解
	parser.parseComment(`// @Success 200 {object} "成功响应"`, routeInfo)

	assert.Len(t, routeInfo.Responses, 1)
	response := routeInfo.Responses[0]
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "成功响应", response.Description)
	assert.Equal(t, "object", response.Schema.Type)
}

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer()
	assert.NotNil(t, renderer)
	assert.NotNil(t, renderer.template)
}

func TestRenderHTMLString(t *testing.T) {
	renderer := NewRenderer()
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	// 添加一个测试API
	api := &APIInfo{
		Path:        "/test",
		Method:      "GET",
		Summary:     "Test API",
		HandlerFunc: "testHandler",
	}
	doc.AddAPI(api)

	html, err := renderer.RenderHTMLString(doc)

	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, config.Title)
	assert.Contains(t, html, "/test")
	assert.Contains(t, html, "GET")
}

// 创建临时测试文件的辅助函数
func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}

func TestParseFileWithRoutes(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "apidoc_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFileContent := `package main

import (
	"github.com/gin-gonic/gin"
)

// @Summary 测试接口
// @Description 这是一个测试接口
// @Tags 测试
// @Router /test [GET]
// @Success 200 {string} "成功"
func testHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "test"})
}

func main() {
	r := gin.Default()
	r.GET("/test", testHandler)
}
`

	testFile := createTestFile(t, tempDir, "main.go", testFileContent)

	config := NewConfig().SetProjectRoot(tempDir).SetScanDirs(tempDir)
	parser := NewParser(config)

	err = parser.parseFile(testFile)
	assert.NoError(t, err)

	// 检查是否正确解析了路由信息
	assert.NotEmpty(t, parser.routeMap)
}

func TestGenerateDocsIntegration(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "apidoc_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFileContent := `package main

import (
	"github.com/gin-gonic/gin"
)

// @Summary 获取用户列表
// @Description 获取所有用户的列表
// @Tags 用户
// @Router /users [GET]
// @Success 200 {array} "用户列表"
func getUsers(c *gin.Context) {
	c.JSON(200, []string{"user1", "user2"})
}

func main() {
	r := gin.Default()
	r.GET("/users", getUsers)
}
`

	createTestFile(t, tempDir, "main.go", testFileContent)

	// 配置
	config := NewConfig().
		SetProjectRoot(tempDir).
		SetScanDirs(".").
		SetTitle("Test API").
		SetVersion("1.0.0")

	// 更新扫描目录为绝对路径
	config.ScanDirs = []string{tempDir}

	// 生成文档
	doc, err := GenerateDocsWithConfig(config)

	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "Test API", doc.Title)
	assert.Equal(t, "1.0.0", doc.Version)
}

func TestMethodCountInTemplate(t *testing.T) {
	config := DefaultConfig()
	doc := NewAPIDoc(config)

	// 添加不同方法的API
	apis := []*APIInfo{
		{Method: "GET", Path: "/users", HandlerFunc: "getUsers"},
		{Method: "POST", Path: "/users", HandlerFunc: "createUser"},
		{Method: "GET", Path: "/health", HandlerFunc: "health"},
		{Method: "DELETE", Path: "/users/1", HandlerFunc: "deleteUser"},
	}

	for _, api := range apis {
		doc.AddAPI(api)
	}

	renderer := NewRenderer()
	html, err := renderer.RenderHTMLString(doc)

	assert.NoError(t, err)
	assert.NotEmpty(t, html)

	// 检查HTML中是否包含正确的统计信息
	assert.Contains(t, html, ">4<") // 总API数
	// 注意：模板函数的测试在实际渲染中才能体现
}

func TestConfigScanDirs(t *testing.T) {
	config := NewConfig()

	// 测试添加扫描目录
	config.AddScanDir("./handlers")
	config.AddScanDir("./controllers")

	assert.Contains(t, config.ScanDirs, "./handlers")
	assert.Contains(t, config.ScanDirs, "./controllers")

	// 测试设置扫描目录
	config.SetScanDirs("./api", "./routes")
	assert.Equal(t, []string{"./api", "./routes"}, config.ScanDirs)
}

func TestGetAbsolutePath(t *testing.T) {
	config := NewConfig().SetProjectRoot("/home/user/project")

	// 测试相对路径
	absPath := config.getAbsolutePath("./handlers")
	assert.Equal(t, "/home/user/project/handlers", absPath)

	// 测试绝对路径
	absPath = config.getAbsolutePath("/absolute/path")
	assert.Equal(t, "/absolute/path", absPath)
}
