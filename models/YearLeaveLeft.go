package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
	// toolkit "github.com/creativelab/toolkit"
)

type YearLastLeave struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId ` bson:"_id" , json:"_id" `
	IdEmp         string
	Name          string
	JoinDate      string
	Leave         int
	LastLeaveLeft float32
}

func NewLastLeaveModel() *YearLastLeave {
	m := new(YearLastLeave)
	m.Id = bson.NewObjectId()
	return m
}

func (e *YearLastLeave) RecordID() interface{} {
	return e.Id
}

func (m *YearLastLeave) TableName() string {
	return "SysRoles"
}
