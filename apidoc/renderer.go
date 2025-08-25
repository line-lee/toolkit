package apidoc

import (
	"html/template"
	"io"
	"strings"
)

// Renderer 模板渲染器
type Renderer struct {
	template *template.Template
}

// NewRenderer 创建新的渲染器
func NewRenderer() *Renderer {
	tmpl := template.Must(template.New("apidoc").Funcs(template.FuncMap{
		"len": func(v interface{}) int {
			switch s := v.(type) {
			case []*APIInfo:
				return len(s)
			case []*Model:
				return len(s)
			case []string:
				return len(s)
			default:
				return 0
			}
		},
		"GetMethodCount": func(doc *APIDoc, method string) int {
			count := 0
			for _, api := range doc.APIs {
				if strings.ToUpper(api.Method) == strings.ToUpper(method) {
					count++
				}
			}
			return count
		},
	}).Parse(HTMLTemplate))

	return &Renderer{
		template: tmpl,
	}
}

// RenderHTML 渲染HTML文档
func (r *Renderer) RenderHTML(doc *APIDoc, writer io.Writer) error {
	return r.template.Execute(writer, doc)
}

// RenderHTMLString 渲染HTML字符串
func (r *Renderer) RenderHTMLString(doc *APIDoc) (string, error) {
	var buf strings.Builder
	err := r.RenderHTML(doc, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
