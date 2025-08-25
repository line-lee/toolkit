package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/line-lee/toolkit/apidoc"
)

// User 用户数据模型
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

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// @Summary 获取用户列表
// @Description 获取所有用户的列表，支持分页和查询
// @Tags 用户管理
// @Router /users [GET]
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param name query string false "用户名称过滤"
// @Success 200 {object} "成功获取用户列表"
// @Failure 400 {object} "请求参数错误"
func getUserList(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	name := c.Query("name")

	// 模拟数据
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25, IsActive: true},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30, IsActive: true},
	}

	// 如果有名称过滤，进行简单过滤
	if name != "" {
		filteredUsers := make([]User, 0)
		for _, user := range users {
			if user.Name == name {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成功",
		Data: gin.H{
			"users": users,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": len(users),
			},
		},
	})
}

// @Summary 获取用户详情
// @Description 根据ID获取单个用户的详细信息
// @Tags 用户管理
// @Router /users/{id} [GET]
// @Param id path int true "用户ID"
// @Success 200 {object} "成功获取用户信息"
// @Failure 400 {object} "无效的用户ID"
// @Failure 404 {object} "用户不存在"
func getUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 模拟查找用户
	if id == 1 {
		user := User{
			ID:       1,
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Age:      25,
			IsActive: true,
		}
		c.JSON(http.StatusOK, Response{
			Code:    200,
			Message: "成功",
			Data:    user,
		})
	} else {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
		})
	}
}

// @Summary 创建用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Router /users [POST]
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} "用户创建成功"
// @Failure 400 {object} "请求参数错误"
func createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 模拟创建用户
	newUser := User{
		ID:       99, // 模拟生成的ID
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		IsActive: true,
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "用户创建成功",
		Data:    newUser,
	})
}

// @Summary 更新用户
// @Description 更新指定用户的信息
// @Tags 用户管理
// @Router /users/{id} [PUT]
// @Param id path int true "用户ID"
// @Param user body CreateUserRequest true "更新的用户信息"
// @Success 200 {object} "用户更新成功"
// @Failure 400 {object} "请求参数错误"
// @Failure 404 {object} "用户不存在"
func updateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 模拟更新用户
	updatedUser := User{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		IsActive: true,
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "用户更新成功",
		Data:    updatedUser,
	})
}

// @Summary 删除用户
// @Description 根据ID删除指定的用户
// @Tags 用户管理
// @Router /users/{id} [DELETE]
// @Param id path int true "用户ID"
// @Success 200 {object} "用户删除成功"
// @Failure 400 {object} "无效的用户ID"
// @Failure 404 {object} "用户不存在"
func deleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	// 模拟删除用户
	if id > 0 {
		c.JSON(http.StatusOK, Response{
			Code:    200,
			Message: "用户删除成功",
		})
	} else {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
		})
	}
}

// @Summary 系统状态
// @Description 获取系统运行状态
// @Tags 系统
// @Router /health [GET]
// @Success 200 {object} "系统状态正常"
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "系统运行正常",
		Data: gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"version":   "1.0.0",
		},
	})
}

func main() {
	r := gin.Default()

	// 配置API文档生成器
	config := apidoc.NewConfig().
		SetTitle("用户管理 API").
		SetVersion("1.0.0").
		SetDescription("这是一个用户管理系统的API文档示例").
		SetBasePath("/api/v1").
		SetScanDirs(".").
		EnableDebug()

	// 注册业务路由
	api := r.Group("/api/v1")
	{
		// 用户管理
		api.GET("/users", getUserList)
		api.GET("/users/:id", getUserByID)
		api.POST("/users", createUser)
		api.PUT("/users/:id", updateUser)
		api.DELETE("/users/:id", deleteUser)

		// 系统状态
		api.GET("/health", healthCheck)
	}

	// 注册文档路由
	apidoc.RegisterRoutesWithConfig(r, config)

	println("服务启动成功!")
	println("文档地址: http://localhost:8080/docs")
	println("JSON文档: http://localhost:8080/docs/json")
	println("完整API文档: http://localhost:8080/docs/api")

	r.Run(":8080")
}
