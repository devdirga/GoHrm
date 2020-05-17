package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type NotificationModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id",json:"_id"`
	UserId        string
	IdRequest     string
	Notif         NotificationDetails
	IsConfirmed   bool
}

type NotificationDetails struct {
	Name string `bson:"name,omitempty"`

	DateFrom    string
	DateTo      string
	Description string
	Status      string
	// HistoryDate    string
	RequestType    string
	Reason         string
	ManagerApprove string
	// IsEmergency    bool
	StatusApproval string
	CreatedAt      string
	UpdatedAt      string
}

func NewNotificationModel() *NotificationModel {
	m := new(NotificationModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *NotificationModel) RecordID() interface{} {
	return e.Id
}

func (e *NotificationModel) TableName() string {
	return "notification"
}
