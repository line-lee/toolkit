package apidoc

import "time"

// Doc 传统文档结构（向后兼容）
type Doc struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	APIs    []API  `json:"apis"`
}

// API 传统API结构（向后兼容）
type API struct {
	Path     string   `json:"path"`
	Method   string   `json:"method"`
	Summary  string   `json:"summary"`
	Handlers []string `json:"handlers"`
}

// APIDoc 完整的API文档结构
type APIDoc struct {
	Info     *Info      `json:"info"`
	BasePath string     `json:"basePath"`
	Schemes  []string   `json:"schemes"`
	APIs     []*APIInfo `json:"apis"`
	Models   []*Model   `json:"models,omitempty"`
}

// Info 文档基本信息
type Info struct {
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Contact     *Contact `json:"contact,omitempty"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// APIInfo 详细的API信息
type APIInfo struct {
	Path        string       `json:"path"`
	Method      string       `json:"method"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty"`
	Responses   []*Response  `json:"responses,omitempty"`
	HandlerFunc string       `json:"handlerFunc"`
	Deprecated  bool         `json:"deprecated,omitempty"`
}

// Parameter 请求参数
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, body, form
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
}

// Response 响应信息
type Response struct {
	Code        int                `json:"code"`
	Description string             `json:"description"`
	Schema      *Schema            `json:"schema,omitempty"`
	Headers     map[string]*Header `json:"headers,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
}

// Header 响应头信息
type Header struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Schema 数据模型定义
type Schema struct {
	Type       string             `json:"type,omitempty"`
	Format     string             `json:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
	Required   []string           `json:"required,omitempty"`
	Ref        string             `json:"$ref,omitempty"`
}

// Model 数据模型
type Model struct {
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Description string             `json:"description,omitempty"`
}

// RouteInfo 路由信息（从注释中解析）
type RouteInfo struct {
	// @Router 注解信息
	Path   string
	Method string

	// @Summary 注解信息
	Summary string

	// @Description 注解信息
	Description string

	// @Tags 注解信息
	Tags []string

	// @Param 注解信息
	Parameters []*Parameter

	// @Success/@Failure 注解信息
	Responses []*Response

	// @Deprecated 注解信息
	Deprecated bool

	// 处理函数名
	HandlerFunc string

	// 文件位置信息
	File string
	Line int
}

// NewAPIDoc 创建新的API文档
func NewAPIDoc(config *Config) *APIDoc {
	return &APIDoc{
		Info: &Info{
			Title:       config.Title,
			Version:     config.Version,
			Description: config.Description,
		},
		BasePath: config.BasePath,
		Schemes:  []string{"http", "https"},
		APIs:     make([]*APIInfo, 0),
		Models:   make([]*Model, 0),
	}
}

// AddAPI 添加API信息
func (doc *APIDoc) AddAPI(apiInfo *APIInfo) {
	doc.APIs = append(doc.APIs, apiInfo)
}

// AddModel 添加数据模型
func (doc *APIDoc) AddModel(model *Model) {
	doc.Models = append(doc.Models, model)
}

// GetGeneratedTime 获取生成时间
func (doc *APIDoc) GetGeneratedTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ToLegacyDoc 转换为旧版文档格式（向后兼容）
func (doc *APIDoc) ToLegacyDoc() *Doc {
	legacyDoc := &Doc{
		Title:   doc.Info.Title,
		Version: doc.Info.Version,
		APIs:    make([]API, 0, len(doc.APIs)),
	}

	for _, api := range doc.APIs {
		legacyAPI := API{
			Path:     api.Path,
			Method:   api.Method,
			Summary:  api.Summary,
			Handlers: []string{api.HandlerFunc},
		}
		legacyDoc.APIs = append(legacyDoc.APIs, legacyAPI)
	}

	return legacyDoc
}
