package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type SysUserModel struct {
	orm.ModelBase    `bson:"-",json:"-"`
	Id               string `bson:"_id" , json:"_id"`
	EmpId            string
	Designation      string
	Departement      string
	Username         string
	Fullname         string
	Enable           bool
	PhoneNumber      string
	Email            string
	Address          string
	Gender           string
	YearLeave        int
	PublicLeave      int
	Roles            string
	Password         string
	Location         string
	AccesRight       string
	IsChangePassword bool
	Photo            string
	IsProjectManager string
	IsProjectLeader  string
	LastLeave        string
	ProjectRuleID    string `bson:"projectruleid" , json:"projectruleid"`
	ProjectRuleName  string `bson:"-" , json:"ProjectRuleName"`
	JointDate        string
	AddLeave         string
	DecYear          float64
	TmpYear          int `bson:"tempyear,omitempty",json:"tempyear,omitempty"`
	// Remote           RemoteOption
}

// type RemoteOption struct {
// 	RemoteActive      bool
// 	ConditionalRemote int
// 	FullMonth         bool
// 	Monthly           bool
// }
type SysUserProfileModel struct {
	orm.ModelBase    `bson:"-",json:"-"`
	Id               string `bson:"_id" , json:"_id"`
	EmpId            string
	Designation      string
	Departement      string
	Username         string
	Fullname         string
	PhoneNumber      string
	Email            string
	Address          string
	Gender           string
	YearLeave        int
	PublicLeave      int
	Location         string
	Photo            string
	LastLeave        string
	Password         string `bson:"-" , json:"-"`
	ProjectRuleID    string
	IsChangePassword bool
}

func NewSysUserModel() *SysUserModel {
	m := new(SysUserModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *SysUserModel) RecordID() interface{} {
	return e.Id
}

func (m *SysUserModel) TableName() string {
	return "SysUsers"
}
