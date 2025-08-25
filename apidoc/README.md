# API Doc Generator

[![Go Version](https://img.shields.io/github/go-mod/go-version/line-lee/toolkit)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

一个基于 Go 抽象语法树（AST）技术的 Gin Web 框架接口文档自动生成工具。通过解析 Go 源代码中的注释和路由定义，自动生成美观的 API 文档。

## ✨ 特性

- 🚀 **零配置启动** - 开箱即用，无需复杂配置
- 🎯 **AST 技术** - 基于 Go 抽象语法树，精确解析代码结构
- 🎨 **美观界面** - 响应式设计，支持搜索和交互
- 📝 **丰富注释** - 支持 Swagger 风格的注释语法
- 🔧 **灵活配置** - 支持自定义扫描路径、文档样式等
- 📱 **多种格式** - 同时支持 HTML 和 JSON 格式输出
- 🔍 **实时搜索** - 支持按路径、方法、描述等条件搜索
- 🏷️ **标签分类** - 支持 API 标签分类管理

## 📦 安装

```bash
go get github.com/line-lee/toolkit/apidoc
```

## 🚀 快速开始

### 1. 基础用法

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/line-lee/toolkit/apidoc"
)

// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Router /users/{id} [GET]
// @Param id path int true "用户ID"
// @Success 200 {object} "用户信息"
// @Failure 404 {object} "用户不存在"
func getUser(c *gin.Context) {
    // 处理逻辑
}

func main() {
    r := gin.Default()
    
    // 注册业务路由
    r.GET("/users/:id", getUser)
    
    // 注册文档路由
    apidoc.RegisterRoutes(r)
    
    r.Run(":8080")
}
```

访问 `http://localhost:8080/docs` 查看生成的文档。

### 2. 高级配置

```go
func main() {
    r := gin.Default()
    
    // 自定义配置
    config := apidoc.NewConfig().
        SetTitle("我的 API 文档").
        SetVersion("1.0.0").
        SetDescription("这是一个示例 API 服务").
        SetBasePath("/api/v1").
        SetScanDirs("./handlers", "./controllers").
        EnableDebug()
    
    // 使用自定义配置注册路由
    apidoc.RegisterRoutesWithConfig(r, config)
    
    r.Run(":8080")
}
```

## 📖 注释语法

支持以下 Swagger 风格的注释：

### 基本信息

```go
// @Summary API 摘要
// @Description API 详细描述
// @Tags 标签1,标签2
// @Router /path/{param} [METHOD]
// @Deprecated （标记为已废弃）
```

### 请求参数

```go
// @Param name位置 type required "描述"
// @Param id path int true "用户ID"
// @Param name query string false "用户名称"
// @Param user body UserRequest true "用户信息"
```

参数位置选项：
- `path` - 路径参数
- `query` - 查询参数  
- `header` - 请求头参数
- `body` - 请求体
- `form` - 表单参数

### 响应信息

```go
// @Success 200 {type} "描述"
// @Failure 400 {object} "错误信息"
// @Success 200 {array} "数组响应"
// @Success 201 {string} "创建成功"
```

### 完整示例

```go
// @Summary 创建用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Router /users [POST]
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} User "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func createUser(c *gin.Context) {
    // 实现逻辑
}
```

## 🛠️ 配置选项

### Config 配置项

```go
type Config struct {
    ProjectRoot     string   // 项目根路径
    ScanDirs        []string // 扫描目录列表
    IncludePatterns []string // 包含文件模式
    ExcludePatterns []string // 排除文件模式
    Title           string   // 文档标题
    Version         string   // 文档版本
    Description     string   // 文档描述
    BasePath        string   // API 基础路径
    Debug           bool     // 调试模式
}
```

### 配置方法

```go
config := apidoc.NewConfig().
    SetTitle("API 文档").                    // 设置标题
    SetVersion("1.0.0").                     // 设置版本
    SetDescription("API 接口文档").           // 设置描述
    SetBasePath("/api/v1").                  // 设置基础路径
    SetProjectRoot("/path/to/project").      // 设置项目根路径
    SetScanDirs("./handlers", "./api").      // 设置扫描目录
    AddScanDir("./controllers").             // 添加扫描目录
    EnableDebug()                            // 启用调试模式
```

## 🌐 路由注册

### 基础注册

```go
// 使用默认配置
apidoc.RegisterRoutes(r)
```

### 自定义配置注册

```go
config := apidoc.NewConfig().SetTitle("My API")
apidoc.RegisterRoutesWithConfig(r, config)
```

### 带前缀注册

```go
// 文档路由将为 /api/docs, /api/docs/json 等
apidoc.RegisterRoutesWithPrefix(r, "/api", config)
```

### 独立文档服务器

```go
// 创建独立的文档服务器
docsServer := apidoc.SetupDocsServer(":8081", config)
go docsServer.Run(":8081")
```

## 📋 可用端点

| 端点 | 描述 | 格式 |
|------|------|------|
| `/docs` | HTML 文档界面 | HTML |
| `/docs/json` | 简化的 JSON 文档 | JSON |
| `/docs/api` | 完整的 API 文档 | JSON |
| `/docs/health` | 服务健康检查 | JSON |

## 🎯 数据模型

自动解析 Go 结构体并生成文档：

```go
// User 用户信息
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Age      int    `json:"age,omitempty"`
    IsActive bool   `json:"is_active"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age,omitempty"`
}
```

支持的 JSON 标签：
- `json:"field_name"` - 字段名称
- `json:"field_name,omitempty"` - 可选字段
- `json:"-"` - 忽略字段

## 🔧 编程式使用

### 生成文档对象

```go
// 使用默认配置
doc, err := apidoc.GenerateDocs()

// 使用自定义配置
config := apidoc.NewConfig().SetTitle("My API")
doc, err := apidoc.GenerateDocsWithConfig(config)
```

### 自定义渲染

```go
parser := apidoc.NewParser(config)
apiDoc, err := parser.Parse()

renderer := apidoc.NewRenderer()
html, err := renderer.RenderHTMLString(apiDoc)
```

## 📂 项目结构示例

```
your-project/
├── main.go
├── handlers/
│   ├── user.go      # 用户相关处理函数
│   └── auth.go      # 认证相关处理函数
├── models/
│   └── user.go      # 数据模型定义
└── docs/            # 生成的文档（可选）
```

## 🧪 测试

运行测试：

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v ./apidoc

# 运行带覆盖率的测试
go test -v -cover ./apidoc
```

## 📝 示例项目

查看 `example/` 目录中的完整示例：

```bash
cd example
go run main.go
```

然后访问：
- 文档界面：http://localhost:8080/docs
- JSON 文档：http://localhost:8080/docs/json
- 完整 API：http://localhost:8080/docs/api

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 🆘 常见问题

### Q: 为什么我的注释没有被解析？

A: 请确保：
1. 注释格式正确（使用 `//` 开头）
2. 函数参数包含 `*gin.Context`
3. 文件在配置的扫描目录中
4. 文件符合包含模式且不在排除模式中

### Q: 如何自定义文档样式？

A: 你可以：
1. 修改 `template.go` 中的 HTML 模板
2. 创建自定义渲染器
3. 使用 JSON 格式自行渲染

### Q: 支持哪些 HTTP 方法？

A: 支持所有标准 HTTP 方法：GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

### Q: 如何处理嵌套结构体？

A: 工具会自动解析嵌套结构体，并在文档中生成相应的模型定义。

## 📞 联系方式

- 项目地址：https://github.com/line-lee/toolkit
- 问题报告：https://github.com/line-lee/toolkit/issues

---

**⭐ 如果这个项目对你有帮助，请给它一个 Star！**