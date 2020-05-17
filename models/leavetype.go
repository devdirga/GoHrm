package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type LeaveTypeModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id",json:"_id"`
	Name          string        `bson:"Name",json:"Name"`
	CodeName      string        `bson:"CodeName",json:"CodeName"`
}

func NewLeaveTypeModel() *LeaveTypeModel {
	m := new(LeaveTypeModel)
	return m
}

func (m *LeaveTypeModel) TableName() string {
	return "LeaveTypeMaster"
}
