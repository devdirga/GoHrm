package models

import (
	"github.com/creativelab/orm"
)

type FtpModel struct {
	orm.ModelBase       `bson:"-",json:"-"`
	Id             int `bson:"_id" , json:"_id" `
	Pathupload     	string
	Pathprocess     string
	Pathdone     	string
}

func NewFtpModel() *FtpModel {
	m := new(FtpModel)
	return m
}
func (e *FtpModel) RecordID() interface{} {
	return e.Id
}

func (m *FtpModel) TableName() string {
	return "FtpSetting"
}
