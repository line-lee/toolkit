package apidoc

const HTMLTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Info.Title}} - API文档</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f8f9fa;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 20px;
            text-align: center;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.1);
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            font-weight: 700;
        }
        
        .header .version {
            background: rgba(255,255,255,0.2);
            padding: 5px 15px;
            border-radius: 20px;
            display: inline-block;
            margin-top: 10px;
        }
        
        .info-section {
            background: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        
        .info-section h2 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 1.5em;
        }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
            border-left: 4px solid #667eea;
        }
        
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 5px;
        }
        
        .api-section {
            background: white;
            border-radius: 10px;
            overflow: hidden;
            margin-bottom: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        
        .api-header {
            background: #f8f9fa;
            padding: 20px;
            border-bottom: 1px solid #dee2e6;
            cursor: pointer;
            transition: background-color 0.3s;
        }
        
        .api-header:hover {
            background: #e9ecef;
        }
        
        .api-method {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: bold;
            margin-right: 15px;
            min-width: 60px;
            text-align: center;
        }
        
        .method-GET { background: #28a745; color: white; }
        .method-POST { background: #007bff; color: white; }
        .method-PUT { background: #ffc107; color: #212529; }
        .method-DELETE { background: #dc3545; color: white; }
        .method-PATCH { background: #6f42c1; color: white; }
        .method-HEAD { background: #6c757d; color: white; }
        .method-OPTIONS { background: #17a2b8; color: white; }
        
        .api-path {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 1.1em;
            font-weight: 500;
        }
        
        .api-summary {
            color: #6c757d;
            margin-top: 8px;
        }
        
        .api-content {
            padding: 20px;
            display: none;
        }
        
        .api-content.active {
            display: block;
        }
        
        .section {
            margin-bottom: 25px;
        }
        
        .section h4 {
            color: #495057;
            margin-bottom: 10px;
            font-size: 1.1em;
        }
        
        .param-table, .response-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        
        .param-table th,
        .param-table td,
        .response-table th,
        .response-table td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #dee2e6;
        }
        
        .param-table th,
        .response-table th {
            background: #f8f9fa;
            font-weight: 600;
            color: #495057;
        }
        
        .required {
            color: #dc3545;
            font-weight: bold;
        }
        
        .code {
            background: #f8f9fa;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9em;
        }
        
        .response-code {
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 0.9em;
        }
        
        .code-200 { background: #d4edda; color: #155724; }
        .code-400 { background: #f8d7da; color: #721c24; }
        .code-500 { background: #f5c6cb; color: #721c24; }
        
        .tags {
            margin-top: 10px;
        }
        
        .tag {
            display: inline-block;
            background: #e9ecef;
            color: #495057;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            margin-right: 8px;
        }
        
        .deprecated {
            opacity: 0.6;
            text-decoration: line-through;
        }
        
        .search-box {
            background: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        
        .search-input {
            width: 100%;
            padding: 12px;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            font-size: 1em;
        }
        
        .no-results {
            text-align: center;
            padding: 40px;
            color: #6c757d;
        }
        
        .footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            color: #6c757d;
            font-size: 0.9em;
        }
        
        @media (max-width: 768px) {
            .stats {
                grid-template-columns: 1fr;
            }
            
            .api-method {
                margin-bottom: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Info.Title}}</h1>
            <p>{{.Info.Description}}</p>
            <span class="version">版本 {{.Info.Version}}</span>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">{{len .APIs}}</div>
                <div>接口总数</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.GetMethodCount "GET"}}</div>
                <div>GET 接口</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.GetMethodCount "POST"}}</div>
                <div>POST 接口</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{len .Models}}</div>
                <div>数据模型</div>
            </div>
        </div>

        <div class="search-box">
            <input type="text" class="search-input" placeholder="搜索接口路径、方法或描述..." onkeyup="searchAPIs(this.value)">
        </div>
        
        <div id="api-list">
            {{range $index, $api := .APIs}}
            <div class="api-section" data-search="{{$api.Method}} {{$api.Path}} {{$api.Summary}} {{$api.Description}}">
                <div class="api-header" onclick="toggleAPI({{$index}})">
                    <span class="api-method method-{{$api.Method}}">{{$api.Method}}</span>
                    <span class="api-path {{if $api.Deprecated}}deprecated{{end}}">{{$api.Path}}</span>
                    {{if $api.Summary}}
                    <div class="api-summary">{{$api.Summary}}</div>
                    {{end}}
                    {{if $api.Tags}}
                    <div class="tags">
                        {{range $api.Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    {{end}}
                </div>
                <div class="api-content" id="api-content-{{$index}}">
                    {{if $api.Description}}
                    <div class="section">
                        <h4>描述</h4>
                        <p>{{$api.Description}}</p>
                    </div>
                    {{end}}
                    
                    {{if $api.Parameters}}
                    <div class="section">
                        <h4>请求参数</h4>
                        <table class="param-table">
                            <thead>
                                <tr>
                                    <th>参数名</th>
                                    <th>位置</th>
                                    <th>类型</th>
                                    <th>必需</th>
                                    <th>描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $api.Parameters}}
                                <tr>
                                    <td><span class="code">{{.Name}}</span></td>
                                    <td>{{.In}}</td>
                                    <td>{{.Type}}</td>
                                    <td>{{if .Required}}<span class="required">是</span>{{else}}否{{end}}</td>
                                    <td>{{.Description}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{end}}
                    
                    {{if $api.Responses}}
                    <div class="section">
                        <h4>响应</h4>
                        <table class="response-table">
                            <thead>
                                <tr>
                                    <th>状态码</th>
                                    <th>描述</th>
                                    <th>数据类型</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $api.Responses}}
                                <tr>
                                    <td><span class="response-code code-{{.Code}}">{{.Code}}</span></td>
                                    <td>{{.Description}}</td>
                                    <td>{{if .Schema}}{{.Schema.Type}}{{end}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{end}}
                    
                    <div class="section">
                        <h4>处理函数</h4>
                        <p><span class="code">{{$api.HandlerFunc}}</span></p>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        
        <div class="no-results" id="no-results" style="display: none;">
            <p>没有找到匹配的接口</p>
        </div>
        
        <div class="footer">
            <p>文档生成时间: <span id="generated-time"></span></p>
            <p>由 <strong>API Doc Generator</strong> 自动生成</p>
        </div>
    </div>

    <script>
        function toggleAPI(index) {
            const content = document.getElementById('api-content-' + index);
            content.classList.toggle('active');
        }
        
        function searchAPIs(query) {
            const apiSections = document.querySelectorAll('.api-section');
            const noResults = document.getElementById('no-results');
            let hasResults = false;
            
            apiSections.forEach(section => {
                const searchText = section.dataset.search.toLowerCase();
                const matches = query.toLowerCase().split(' ').every(term => 
                    searchText.includes(term)
                );
                
                if (matches || query === '') {
                    section.style.display = 'block';
                    hasResults = true;
                } else {
                    section.style.display = 'none';
                }
            });
            
            noResults.style.display = hasResults ? 'none' : 'block';
        }
        
        // 设置生成时间
        document.getElementById('generated-time').textContent = new Date().toLocaleString('zh-CN');
        
        // 键盘快捷键
        document.addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.key === 'f') {
                e.preventDefault();
                document.querySelector('.search-input').focus();
            }
        });
    </script>
</body>
</html>
`
