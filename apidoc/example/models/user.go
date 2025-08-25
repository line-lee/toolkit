package models

import (
	"time"
)

// User 用户数据模型
type User struct {
	ID          int64     `json:"id" gorm:"primarykey"`
	Username    string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email       string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password    string    `json:"-" gorm:"size:255;not null"` // 密码不在JSON中暴露
	FullName    string    `json:"full_name" gorm:"size:100"`
	Avatar      string    `json:"avatar" gorm:"size:255"`
	Phone       string    `json:"phone" gorm:"size:20"`
	Status      string    `json:"status" gorm:"size:20;default:active"` // active, inactive, banned
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID     int64 `json:"id" gorm:"primarykey"`
	UserID int64 `json:"user_id" gorm:"not null"`
	RoleID int64 `json:"role_id" gorm:"not null"`
	User   User  `json:"user" gorm:"foreignKey:UserID"`
	Role   Role  `json:"role" gorm:"foreignKey:RoleID"`
}

// Role 角色模型
type Role struct {
	ID          int64        `json:"id" gorm:"primarykey"`
	Name        string       `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Description string       `json:"description" gorm:"size:255"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// Permission 权限模型
type Permission struct {
	ID          int64     `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Action      string    `json:"action" gorm:"size:50;not null"` // read, write, delete
	Resource    string    `json:"resource" gorm:"size:50;not null"` // user, role, etc.
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserWithRoles 包含角色信息的用户模型
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// UserWithPermissions 包含权限信息的用户模型
type UserWithPermissions struct {
	User
	Permissions []Permission `json:"permissions"`
}