package apidoc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Parser AST解析器
type Parser struct {
	config   *Config
	fileSet  *token.FileSet
	routeMap map[string]*RouteInfo
	models   map[string]*Model
}

// NewParser 创建新的解析器
func NewParser(config *Config) *Parser {
	return &Parser{
		config:   config,
		fileSet:  token.NewFileSet(),
		routeMap: make(map[string]*RouteInfo),
		models:   make(map[string]*Model),
	}
}

// GenerateDocs 生成接口文档
func GenerateDocs() (*Doc, error) {
	config := DefaultConfig()
	return GenerateDocsWithConfig(config)
}

// GenerateDocsWithConfig 使用指定配置生成接口文档
func GenerateDocsWithConfig(config *Config) (*Doc, error) {
	parser := NewParser(config)
	apiDoc, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return apiDoc.ToLegacyDoc(), nil
}

// Parse 解析项目并生成文档
func (p *Parser) Parse() (*APIDoc, error) {
	doc := NewAPIDoc(p.config)

	// 扫描所有指定的目录
	for _, scanDir := range p.config.ScanDirs {
		absDir := p.config.getAbsolutePath(scanDir)
		if err := p.scanDirectory(absDir); err != nil {
			if p.config.Debug {
				log.Printf("扫描目录 %s 失败: %v", absDir, err)
			}
			continue
		}
	}

	// 转换路由信息为API文档
	for _, routeInfo := range p.routeMap {
		apiInfo := p.convertRouteToAPI(routeInfo)
		doc.AddAPI(apiInfo)
	}

	// 添加模型信息
	for _, model := range p.models {
		doc.AddModel(model)
	}

	return doc, nil
}

// scanDirectory 扫描目录中的Go文件
func (p *Parser) scanDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件是否应该被包含
		if !p.config.shouldIncludeFile(path) {
			return nil
		}

		return p.parseFile(path)
	})
}

// parseFile 解析单个Go文件
func (p *Parser) parseFile(filePath string) error {
	file, err := parser.ParseFile(p.fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		if p.config.Debug {
			log.Printf("解析文件 %s 失败: %v", filePath, err)
		}
		return err
	}

	// 解析路由注册
	p.parseRouteRegistrations(file, filePath)

	// 解析处理函数
	p.parseHandlerFunctions(file, filePath)

	// 解析数据模型
	p.parseModels(file, filePath)

	return nil
}

// parseRouteRegistrations 解析路由注册语句
func (p *Parser) parseRouteRegistrations(file *ast.File, filePath string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// 检查是否是Gin路由注册调用
			if p.isGinRouteCall(node) {
				routeInfo := p.extractRouteFromCall(node, filePath)
				if routeInfo != nil {
					key := fmt.Sprintf("%s:%s", routeInfo.Method, routeInfo.Path)
					p.routeMap[key] = routeInfo
				}
			}
		}
		return true
	})
}

// parseHandlerFunctions 解析处理函数
func (p *Parser) parseHandlerFunctions(file *ast.File, filePath string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if p.isHandlerFunc(node) {
				routeInfo := p.extractRouteFromHandler(node, filePath)
				if routeInfo != nil {
					// 尝试匹配已存在的路由或创建新的
					for key, existingRoute := range p.routeMap {
						if existingRoute.HandlerFunc == routeInfo.HandlerFunc {
							// 合并信息
							p.mergeRouteInfo(p.routeMap[key], routeInfo)
							return true
						}
					}
					// 如果没有找到匹配的路由，则创建一个默认的
					if routeInfo.Path != "" && routeInfo.Method != "" {
						key := fmt.Sprintf("%s:%s", routeInfo.Method, routeInfo.Path)
						p.routeMap[key] = routeInfo
					}
				}
			}
		}
		return true
	})
}

// parseModels 解析数据模型
func (p *Parser) parseModels(file *ast.File, filePath string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := node.Type.(*ast.StructType); ok {
				model := p.convertStructToModel(node.Name.Name, structType)
				if model != nil {
					p.models[model.Name] = model
				}
			}
		}
		return true
	})
}

// isGinRouteCall 检查是否为Gin路由调用
func (p *Parser) isGinRouteCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		methodName := sel.Sel.Name
		// 检查是否为 HTTP 方法
		httpMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		for _, method := range httpMethods {
			if methodName == method {
				return true
			}
		}
		// 检查其他可能的路由方法
		if methodName == "Handle" || methodName == "Any" {
			return true
		}
	}
	return false
}

// extractRouteFromCall 从调用表达式中提取路由信息
func (p *Parser) extractRouteFromCall(call *ast.CallExpr, filePath string) *RouteInfo {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) < 2 {
		return nil
	}

	method := sel.Sel.Name
	if method == "Handle" && len(call.Args) >= 3 {
		// Handle(method, path, handler)
		if methodLit, ok := call.Args[0].(*ast.BasicLit); ok {
			method = strings.Trim(methodLit.Value, `"`)
		}
	}

	// 提取路径
	var path string
	if pathLit, ok := call.Args[0].(*ast.BasicLit); ok {
		path = strings.Trim(pathLit.Value, `"`)
	} else if method == "Handle" {
		if pathLit, ok := call.Args[1].(*ast.BasicLit); ok {
			path = strings.Trim(pathLit.Value, `"`)
		}
	}

	// 提取处理函数名
	var handlerFunc string
	handlerArgIndex := 1
	if method == "Handle" {
		handlerArgIndex = 2
	}
	if len(call.Args) > handlerArgIndex {
		if ident, ok := call.Args[handlerArgIndex].(*ast.Ident); ok {
			handlerFunc = ident.Name
		}
	}

	position := p.fileSet.Position(call.Pos())
	return &RouteInfo{
		Path:        path,
		Method:      strings.ToUpper(method),
		HandlerFunc: handlerFunc,
		File:        filePath,
		Line:        position.Line,
	}
}

// isHandlerFunc 检查函数是否为Gin处理函数
func (p *Parser) isHandlerFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}

	for _, param := range fn.Type.Params.List {
		if len(param.Names) > 0 {
			// 检查参数类型是否为 *gin.Context
			if starExpr, ok := param.Type.(*ast.StarExpr); ok {
				if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
					if selExpr.Sel.Name == "Context" {
						// 进一步检查包名
						if ident, ok := selExpr.X.(*ast.Ident); ok {
							if ident.Name == "gin" {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

// extractRouteFromHandler 从处理函数中提取路由信息
func (p *Parser) extractRouteFromHandler(fn *ast.FuncDecl, filePath string) *RouteInfo {
	if fn.Doc == nil {
		return nil
	}

	routeInfo := &RouteInfo{
		HandlerFunc: fn.Name.Name,
		File:        filePath,
		Line:        p.fileSet.Position(fn.Pos()).Line,
	}

	// 解析注释
	for _, comment := range fn.Doc.List {
		p.parseComment(comment.Text, routeInfo)
	}

	return routeInfo
}

// parseComment 解析注释中的文档信息
func (p *Parser) parseComment(comment string, routeInfo *RouteInfo) {
	comment = strings.TrimSpace(comment)

	// @Router 路由信息
	if match := regexp.MustCompile(`@Router\s+(\S+)\s+\[(\w+)\]`).FindStringSubmatch(comment); len(match) == 3 {
		routeInfo.Path = match[1]
		routeInfo.Method = strings.ToUpper(match[2])
		return
	}

	// @route 简化路由信息
	if match := regexp.MustCompile(`@route\s+(\w+)\s+(\S+)`).FindStringSubmatch(comment); len(match) == 3 {
		routeInfo.Method = strings.ToUpper(match[1])
		routeInfo.Path = match[2]
		return
	}

	// @Summary 摘要
	if match := regexp.MustCompile(`@Summary\s+(.+)`).FindStringSubmatch(comment); len(match) == 2 {
		routeInfo.Summary = strings.TrimSpace(match[1])
		return
	}

	// @Description 描述
	if match := regexp.MustCompile(`@Description\s+(.+)`).FindStringSubmatch(comment); len(match) == 2 {
		routeInfo.Description = strings.TrimSpace(match[1])
		return
	}

	// @Tags 标签
	if match := regexp.MustCompile(`@Tags\s+(.+)`).FindStringSubmatch(comment); len(match) == 2 {
		tags := strings.Split(match[1], ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		routeInfo.Tags = tags
		return
	}

	// @Param 参数
	if match := regexp.MustCompile(`@Param\s+(\w+)\s+(\w+)\s+(\w+)\s+(\w+)\s+"([^"]*)"`).FindStringSubmatch(comment); len(match) == 6 {
		param := &Parameter{
			Name:        match[1],
			In:          match[2],
			Type:        match[3],
			Required:    match[4] == "true",
			Description: match[5],
		}
		routeInfo.Parameters = append(routeInfo.Parameters, param)
		return
	}

	// @Success/@Failure 响应
	if match := regexp.MustCompile(`@(Success|Failure)\s+(\d+)\s+{(\w+)}\s+"([^"]*)"`).FindStringSubmatch(comment); len(match) == 5 {
		code, _ := strconv.Atoi(match[2])
		response := &Response{
			Code:        code,
			Description: match[4],
			Schema: &Schema{
				Type: match[3],
			},
		}
		routeInfo.Responses = append(routeInfo.Responses, response)
		return
	}

	// @Deprecated 已废弃
	if strings.Contains(comment, "@Deprecated") {
		routeInfo.Deprecated = true
		return
	}
}

// convertStructToModel 将结构体转换为模型
func (p *Parser) convertStructToModel(name string, structType *ast.StructType) *Model {
	model := &Model{
		Name:       name,
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   make([]string, 0),
	}

	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			if fieldName.IsExported() {
				schema := p.convertTypeToSchema(field.Type)
				fieldNameStr := fieldName.Name

				// 检查tag
				if field.Tag != nil {
					tagValue := strings.Trim(field.Tag.Value, "`")
					if jsonTag := p.extractJSONTag(tagValue); jsonTag != "" {
						if jsonTag != "-" {
							fieldNameStr = strings.Split(jsonTag, ",")[0]
						}
					}
				}

				model.Properties[fieldNameStr] = schema
			}
		}
	}

	return model
}

// convertTypeToSchema 将Go类型转换为Schema
func (p *Parser) convertTypeToSchema(expr ast.Expr) *Schema {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.convertBasicTypeToSchema(t.Name)
	case *ast.StarExpr:
		return p.convertTypeToSchema(t.X)
	case *ast.ArrayType:
		return &Schema{
			Type:  "array",
			Items: p.convertTypeToSchema(t.Elt),
		}
	case *ast.MapType:
		return &Schema{
			Type: "object",
		}
	case *ast.SelectorExpr:
		return &Schema{
			Ref: fmt.Sprintf("#/definitions/%s", t.Sel.Name),
		}
	default:
		return &Schema{Type: "string"}
	}
}

// convertBasicTypeToSchema 将基本类型转换为Schema
func (p *Parser) convertBasicTypeToSchema(typeName string) *Schema {
	switch typeName {
	case "string":
		return &Schema{Type: "string"}
	case "int", "int8", "int16", "int32", "int64":
		return &Schema{Type: "integer"}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return &Schema{Type: "integer"}
	case "float32", "float64":
		return &Schema{Type: "number"}
	case "bool":
		return &Schema{Type: "boolean"}
	default:
		return &Schema{Type: "string"}
	}
}

// extractJSONTag 提取JSON标签
func (p *Parser) extractJSONTag(tag string) string {
	if match := regexp.MustCompile(`json:"([^"]*)"`).FindStringSubmatch(tag); len(match) == 2 {
		return match[1]
	}
	return ""
}

// mergeRouteInfo 合并路由信息
func (p *Parser) mergeRouteInfo(existing, new *RouteInfo) {
	if new.Summary != "" {
		existing.Summary = new.Summary
	}
	if new.Description != "" {
		existing.Description = new.Description
	}
	if len(new.Tags) > 0 {
		existing.Tags = new.Tags
	}
	if len(new.Parameters) > 0 {
		existing.Parameters = new.Parameters
	}
	if len(new.Responses) > 0 {
		existing.Responses = new.Responses
	}
	if new.Deprecated {
		existing.Deprecated = new.Deprecated
	}
}

// convertRouteToAPI 将路由信息转换为API信息
func (p *Parser) convertRouteToAPI(routeInfo *RouteInfo) *APIInfo {
	return &APIInfo{
		Path:        routeInfo.Path,
		Method:      routeInfo.Method,
		Summary:     routeInfo.Summary,
		Description: routeInfo.Description,
		Tags:        routeInfo.Tags,
		Parameters:  routeInfo.Parameters,
		Responses:   routeInfo.Responses,
		HandlerFunc: routeInfo.HandlerFunc,
		Deprecated:  routeInfo.Deprecated,
	}
}
