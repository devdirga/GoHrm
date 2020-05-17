package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
	// toolkit "github.com/creativelab/toolkit"
	"fmt"
)

type HistoryModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId ` bson:"_id" , json:"_id" `
	Id_profile    string
	Name          string
	Tlp           string
	History       []HistoryDetail
	Address       string
	Email         string
	Dob           string
	Religion      string
	Image_file    string
}

type HistoryDetail struct {
	Date    string
	Time    string
	Status  string
	File    string
	Comment string
	Score   string
}

func NewHistoryModel() *HistoryModel {
	fmt.Println("masuk model")
	m := new(HistoryModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *HistoryModel) RecordID() interface{} {
	return e.Id
}

func (m *HistoryModel) TableName() string {
	return "History"
}
