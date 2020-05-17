package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type EmployeeOvertimeModel struct {
	orm.ModelBase  `bson:"-",json:"-"`
	Id             string `bson:"_id", json:"_id"`
	IdOvertime     string
	UserId         string
	IdEmployee     string
	Project        string
	Name           string
	Location       string
	Email          string
	PhoneNumber    string
	DateOvertime   string
	TimeStart      string
	TimeEnd        string
	DateClosed     string
	TypeOvertime   string
	Hours          int
	TrackHour      int
	ResultMatch    string
	DateAdminCheck string
	Day            int
	Month          int
	Year           int
	IsCheck        bool
	Task           string
	Deadline       string
	DateApprove    string
}

func NewEmployeeOvertimeModel() *EmployeeOvertimeModel {
	m := new(EmployeeOvertimeModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}
func (e *EmployeeOvertimeModel) RecordID() interface{} {
	return e.Id
}

func (m *EmployeeOvertimeModel) TableName() string {
	return "EmployeeOvertime"
}
