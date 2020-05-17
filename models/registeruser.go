package models

import (
	"github.com/creativelab/orm"
)

type RegisterUserModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id" , json:"_id"`
	UserID        string `bson:"UserID" , json:"UserID"`
	IsRegister    bool   `bson:"IsRegister" , json:"IsRegister"`
	EmpID         string `bson:"EmpID" , json:"EmpID"`
	Name          string `bson:"Name" , json:"Name"`
	RoleID        string `bson:"RoleID" , json:"RoleID"`
	RoleName      string `bson:"-" , json:"RoleName"`
}

func NewRegisterUserModel() *RegisterUserModel {
	m := new(RegisterUserModel)

	return m
}
func (e *RegisterUserModel) RecordID() interface{} {
	return e.Id
}
func (m *RegisterUserModel) TableName() string {
	return "RegisterUser"
}
