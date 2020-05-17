package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type LogleaveRemoteModel struct {
	orm.ModelBase     `bson:"-",json:"-"`
	Id                string `bson:"_id", json:"_id"`
	IdRequest         string `bson:"idrequest" , json:"idrequest"`
	TypeRequest       string
	Userid            string
	Name              string
	Email             string
	DateLogCreated    time.Time
	DateLogCreatedStr string
	ListLog           []ListLogModel
	StatusRequest     string `bson:"-" , json:"StatusRequest"`
	DateFrom          time.Time
	DateTo            time.Time
	Project           []string
	Location          string
}
type ListLogModel struct {
	RequestBy      string
	IdRequest      string
	DateLogStr     string
	DateLog        time.Time
	NameLogBy      string
	EmailNameLogBy string
	Description    string
	Status         string
}

func NewLogLeaveRemoteModel() *LogleaveRemoteModel {
	m := new(LogleaveRemoteModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *LogleaveRemoteModel) RecordID() interface{} {
	return e.Id
}

func (e *LogleaveRemoteModel) TableName() string {
	return "logLeaveRemote"
}

type LogleaveRemoteNewModel struct {
	orm.ModelBase     `bson:"-",json:"-"`
	Id                string `bson:"_id", json:"_id"`
	IdRequest         string `bson:"idrequest" , json:"idrequest"`
	TypeRequest       string
	Userid            string
	Name              string
	Email             string
	DateLogCreated    time.Time
	DateLogCreatedStr string
	ListLog           []ListLogModel
	StatusRequest     string `bson:"-" , json:"StatusRequest"`
	DateFrom          time.Time
	DateTo            time.Time
	Project           []string
	Location          string
	DataLeave         []RequestLeaveModel
}
