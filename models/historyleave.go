package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type HistoryLeaveModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id",json:"_id"`
	UserId        string
	Leavehistory  []HistoryDetails
}

type HistoryDetails struct {
	UserId         string `bson:"userid,omitempty"`
	Name           string `bson:"name,omitempty"`
	IdRequest      string
	DateFrom       string
	DateTo         string
	Description    string
	Status         string
	HistoryDate    string
	RequestType    string
	Reason         string
	ManagerApprove string
	IsEmergency    bool
	FileAttachment string
	// LeaderApprove  string
	StatusApproval string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewHistoryLeaveModel() *HistoryLeaveModel {
	m := new(HistoryLeaveModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *HistoryLeaveModel) RecordID() interface{} {
	return e.Id
}

func (e *HistoryLeaveModel) TableName() string {
	return "historyleave"
}
