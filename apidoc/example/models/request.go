package models

// 基础响应结构
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 通用响应结构
type Response struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// 错误响应结构
type ErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"`
}

// 分页请求
type PageRequest struct {
	Page  int    `json:"page" form:"page" validate:"min=1"`
	Size  int    `json:"size" form:"size" validate:"min=1,max=100"`
	Sort  string `json:"sort" form:"sort"`
	Order string `json:"order" form:"order" validate:"oneof=asc desc"`
}

// 分页响应
type PagedResponse struct {
	BaseResponse
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email,max=100" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	FullName string `json:"full_name" binding:"max=100" example:"John Doe"`
	Phone    string `json:"phone" binding:"max=20" example:"+1234567890"`
}

// 批量创建用户请求
type BatchCreateUserRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required,min=1,max=100"`
}

// 更新用户请求
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	FullName *string `json:"full_name,omitempty" binding:"omitempty,max=100"`
	Phone    *string `json:"phone,omitempty" binding:"omitempty,max=20"`
	Status   *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
}

// 用户过滤器
type UserFilter struct {
	Name   string `json:"name" form:"name"`
	Email  string `json:"email" form:"email"`
	Status string `json:"status" form:"status" validate:"omitempty,oneof=active inactive banned"`
}

// 用户查询选项
type UserQueryOptions struct {
	IncludeRoles       bool `json:"include_roles" form:"include_roles"`
	IncludePermissions bool `json:"include_permissions" form:"include_permissions"`
}

// 创建用户选项
type CreateUserOptions struct {
	SendNotification bool  `json:"send_notification"`
	CreatedBy        int64 `json:"created_by"`
}

// 用户响应
type UserResponse struct {
	BaseResponse
	Data User `json:"data"`
}

// 用户详情响应
type UserDetailResponse struct {
	BaseResponse
	Data UserWithRoles `json:"data"`
}

// 批量用户响应
type BatchUserResponse struct {
	BaseResponse
	Data struct {
		Successful []User   `json:"successful"`
		Failed     []string `json:"failed"`
		Total      int      `json:"total"`
		Success    int      `json:"success"`
		Fail       int      `json:"fail"`
	} `json:"data"`
}

// 消息响应
type MessageResponse struct {
	BaseResponse
}

// 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// 登录响应
type LoginResponse struct {
	BaseResponse
	Data struct {
		User        User   `json:"user"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	} `json:"data"`
}

// 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// 重置密码请求
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// 确认重置密码请求
type ConfirmResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}