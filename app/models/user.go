package models

import "time"

type User struct {
	UserID    int64      `gorm:"column:id;primaryKey;not null;autoIncrement;comment:自增ID" json:"user_id,string" db:"user_id"` // 指定json序列化/反序列化时使用小写user_id
	UserName  string     `gorm:"column:username" json:"username" db:"username"`
	Password  string     `gorm:"column:password" json:"password" db:"password"`
	Name      string     `gorm:"column:name" json:"name" db:"name"`
	CreatedAt *time.Time `gorm:"column:created_at;not null;comment:创建时间"`
	UpdatedAt *time.Time `gorm:"column:updated_at;not null;comment:更新时间"`
}
type RegisterForm struct {
	Name     string `json:"name" form:"name" binding:"required"`
	UserName string `json:"username" form:"username" binding:"required,alphanum"`
	Password string `json:"password" form:"password" binding:"required,alphanum"`
	//ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}
type LoginForm struct {
	UserName string `json:"username" form:"username" binding:"required,alphanum"`
	Password string `json:"password" form:"password" binding:"required,alphanum"`
}
