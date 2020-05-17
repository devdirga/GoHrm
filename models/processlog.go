package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type ProcessLogModel struct {
	orm.ModelBase   `bson:"-",json:"-"`
	Id              bson.ObjectId ` bson:"_id" , json:"_id" `
	TradeDateInt    int
	TradeDateString string
	StartTime       time.Time
	EndTime         time.Time
	DetailProcess   []DetailProcessModel
	IsFinish        bool
	Reason          string
	CreateBy        string
	CreateDate      time.Time
	UpdateBy        string
	UpdateDate      time.Time
}

type DetailProcessModel struct {
	SequenceNo       int
	ProcessName      string
	Percentage       float64
	ProcessStartTime time.Time
	ProcessEndTime   time.Time
	Message          string
}

func NewProcessLogModel() *ProcessLogModel {
	m := new(ProcessLogModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *ProcessLogModel) RecordID() interface{} {
	return e.Id
}

func (m *ProcessLogModel) TableName() string {
	return "ProcessLog"
}
