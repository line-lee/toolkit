package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/line-lee/toolkit/apidoc/example/models"
	"github.com/line-lee/toolkit/apidoc/example/services"
)

// UserHandler 用户处理器结构体
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// @Summary 获取用户分页列表
// @Description 获取用户列表，支持分页、排序和多条件过滤
// @Tags 用户管理
// @Router /api/v1/users [GET]
// @Param page query int false "页码，从1开始" default(1)
// @Param size query int false "每页大小" default(10)
// @Param sort query string false "排序字段" default("id")
// @Param order query string false "排序方向：asc或desc" default("asc")
// @Param name query string false "用户名模糊匹配"
// @Param email query string false "邮箱模糊匹配"
// @Param status query string false "用户状态：active, inactive, banned"
// @Success 200 {object} models.PagedResponse{data=[]models.User} "用户列表"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "asc")

	// 解析过滤参数
	filters := &models.UserFilter{
		Name:   c.Query("name"),
		Email:  c.Query("email"),
		Status: c.Query("status"),
	}

	// 构建分页请求
	pageReq := &models.PageRequest{
		Page:  page,
		Size:  size,
		Sort:  sort,
		Order: order,
	}

	// 调用服务层
	result, err := h.userService.GetUsers(pageReq, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "获取用户列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary 根据ID获取用户详情
// @Description 根据用户ID获取用户的详细信息，包括关联的角色和权限
// @Tags 用户管理
// @Router /api/v1/users/{id} [GET]
// @Param id path int true "用户ID" minimum(1)
// @Param include query string false "包含的关联数据：roles,permissions" example("roles,permissions")
// @Success 200 {object} models.UserDetailResponse "用户详情"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
			Error:   "用户ID必须是正整数",
		})
		return
	}

	// 解析包含参数
	include := c.Query("include")
	options := &models.UserQueryOptions{
		IncludeRoles:       contains(include, "roles"),
		IncludePermissions: contains(include, "permissions"),
	}

	user, err := h.userService.GetUserByID(id, options)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    404,
				Message: "用户不存在",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "获取用户信息失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserDetailResponse{
		Code:    200,
		Message: "获取成功",
		Data:    user,
	})
}

// @Summary 创建新用户
// @Description 创建一个新的用户账户，支持批量创建和邮件通知
// @Tags 用户管理
// @Router /api/v1/users [POST]
// @Param user body models.CreateUserRequest true "用户创建请求"
// @Param notify query bool false "是否发送邮件通知" default(false)
// @Success 201 {object} models.UserResponse "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 409 {object} models.ErrorResponse "用户已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 解析查询参数
	notify, _ := strconv.ParseBool(c.Query("notify"))

	// 设置创建选项
	options := &models.CreateUserOptions{
		SendNotification: notify,
		CreatedBy:        getUserIDFromContext(c),
	}

	user, err := h.userService.CreateUser(&req, options)
	if err != nil {
		if err == services.ErrUserExists {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:    409,
				Message: "用户已存在",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "创建用户失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		Code:    201,
		Message: "用户创建成功",
		Data:    user,
	})
}

// @Summary 批量创建用户
// @Description 批量创建多个用户账户
// @Tags 用户管理
// @Router /api/v1/users/batch [POST]
// @Param users body models.BatchCreateUserRequest true "批量用户创建请求"
// @Success 201 {object} models.BatchUserResponse "批量创建结果"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) BatchCreateUsers(c *gin.Context) {
	var req models.BatchCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.userService.BatchCreateUsers(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "批量创建用户失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// @Summary 更新用户信息
// @Description 更新指定用户的信息，支持部分更新
// @Tags 用户管理
// @Router /api/v1/users/{id} [PUT]
// @Param id path int true "用户ID" minimum(1)
// @Param user body models.UpdateUserRequest true "用户更新请求"
// @Success 200 {object} models.UserResponse "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
			Error:   "用户ID必须是正整数",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(id, &req)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    404,
				Message: "用户不存在",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "更新用户失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Code:    200,
		Message: "用户更新成功",
		Data:    user,
	})
}

// @Summary 删除用户
// @Description 删除指定的用户账户（软删除）
// @Tags 用户管理
// @Router /api/v1/users/{id} [DELETE]
// @Param id path int true "用户ID" minimum(1)
// @Param hard query bool false "是否硬删除" default(false)
// @Success 200 {object} models.MessageResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "无效的用户ID",
			Error:   "用户ID必须是正整数",
		})
		return
	}

	// 检查是否硬删除
	hard, _ := strconv.ParseBool(c.Query("hard"))

	err = h.userService.DeleteUser(id, hard)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    404,
				Message: "用户不存在",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "删除用户失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Code:    200,
		Message: "用户删除成功",
	})
}

// 辅助函数：检查字符串是否包含指定值
func contains(str, substr string) bool {
	if str == "" {
		return false
	}
	for _, s := range []string{str} {
		if s == substr {
			return true
		}
	}
	return false
}

// 辅助函数：从上下文获取用户ID
func getUserIDFromContext(c *gin.Context) int64 {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}
