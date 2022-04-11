package model

import (
	"api/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	// 用户名
	Username string `bson:"username" json:"username"`

	// 密码
	Password string `bson:"password" json:"password,omitempty"`

	// 所属部门
	Departments []primitive.ObjectID `bson:"departments" json:"-"`

	// 权限组
	Roles []primitive.ObjectID `bson:"roles" json:"roles,omitempty"`

	// 称呼
	Name string `bson:"name" json:"name"`

	// 头像
	Avatar string `bson:"avatar" json:"avatar"`

	// 地区
	Region string `bson:"region" json:"region"`

	// 城市
	City string `bson:"city" json:"city"`

	// 地址
	Address string `bson:"address" json:"address"`

	// 个人简介
	Introduction string `bson:"introduction" json:"introduction"`

	// 联系电话
	Phone string `bson:"phone" json:"phone"`

	// 电子邮件
	Email string `bson:"email" json:"email"`

	// 标记
	Labels []string `bson:"labels" json:"labels"`

	// 状态
	Status *bool `bson:"status" json:"status"`

	// 创建时间
	CreateTime time.Time `bson:"create_time" json:"-"`

	// 更新时间
	UpdateTime time.Time `bson:"update_time" json:"-"`
}

func NewUser(username string, password string) *User {
	return &User{
		Username:   username,
		Password:   password,
		Roles:      []primitive.ObjectID{},
		Labels:     []string{},
		Status:     common.BoolToP(true),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}

func (x *User) SetEmail(v string) *User {
	x.Email = v
	return x
}

func (x *User) SetRoles(v []primitive.ObjectID) *User {
	x.Roles = v
	return x
}

func (x *User) SetLabel(v string) *User {
	x.Labels = append(x.Labels, v)
	return x
}