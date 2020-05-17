package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type UploadLogModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId ` bson:"_id" , json:"_id" `
	UploadType    string
	FileType      string
	TradeDate     time.Time
	FileCount     int
	Remark        string
	Details       []UploadDetailModel
	CreateBy      string
	CreateDate    time.Time
	UpdateBy      string
	UpdateDate    time.Time
}
type UploadDetailModel struct {
	FileNumber           int
	FileName             string
	FilePath             string
	FileSize             float64
	FileType             string
	Status               string
	Remark               string
	UploadPercentage     float64
	ProcessingPercentage float64
	Message              string
	SkipRow              int
	TotalRowFlatFile     int
	TotalInserted        int
	CutOffInfo           []UploadCutOffInfoModel
	IsValid              bool
}
type UploadCutOffInfoModel struct {
	ProductId      string
	ContractCode   string
	ContractExpiry time.Time
	Count          int
}

//func NewUploadDetailModel() *UploadDetailModel {
//	m := new(UploadDetailModel)
//	m.FileNumber = bson.NewObjectId()
//	return m
//}
func NewUploadLogModel() *UploadLogModel {
	m := new(UploadLogModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *UploadLogModel) RecordID() interface{} {
	return e.Id
}

func (m *UploadLogModel) TableName() string {
	return "UploadLog"
}
