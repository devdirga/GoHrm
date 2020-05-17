package models

import (
	// "time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type RequestLeaveModel struct {
	orm.ModelBase            `bson:"-",json:"-"`
	Id                       string `bson:"_id", json:"_id"`
	UserId                   string
	Name                     string
	EmpId                    string
	Designation              string
	Location                 string
	Departement              string
	Reason                   string
	Email                    string
	Address                  string
	Contact                  string
	LeaveFrom                string
	LeaveTo                  string
	NoOfDays                 int
	Project                  []string
	YearLeave                int
	PublicLeave              int
	ProjectManagerList       []ProjectManager
	StatusManagerProject     AprovalManagerProject
	BranchManager            []BranchManagerList
	StatusProjectCoordinator []AprovalProjectCoordinator2
	StatusBusinesAnalyst     []AprovalBusinessAnalyst
	StatusProjectLeader      []AprovalProjectLeader
	ResultRequest            string
	IsEmergency              bool
	DateCreateLeave          string
	DetailsLeave             []DetailAprovalLeader
	LeaveDateList            []string
	ExpiredOn                string
	ExpRemaining             string
	IsSpecials               bool
	IsReset                  bool
	FileLocation             string
	IsAttach                 bool
}

type ProjectManager struct {
	IdEmp         string
	Name          string
	Location      string
	Email         string
	PhoneNumber   string
	StatusRequest string
	Reason        string
	ProjectName   string
	UserId        string
}

type BranchManagerList struct {
	Email       string
	IdEmp       string
	Location    string
	Name        string
	PhoneNumber string
	UserId      string
}

type AprovalProjectCoordinator2 struct {
	IdEmp         string
	Name          string
	Location      string
	Email         string
	PhoneNumber   string
	StatusRequest string
	Reason        string
	ProjectName   string
	UserId        string
}

type DetailAprovalLeader struct {
	UserId     string
	IdRequest  string
	LeaderName string
	DateLeave  string
	IsApproved bool
	Reason     string
}

type AprovalProjectLeader struct {
	IdEmp         string
	Name          string
	Location      string
	Email         string
	PhoneNumber   string
	StatusRequest string
	Reason        string
	ProjectName   string
	UserId        string
}
type AprovalManagerProject struct {
	IdEmp         string
	Name          string
	Location      string
	Email         string
	PhoneNumber   string
	StatusRequest string
	Reason        string
	ProjectName   string
	UserId        string
}
type AprovalBusinessAnalyst struct {
	IdEmp         string
	Name          string
	Location      string
	Email         string
	PhoneNumber   string
	StatusRequest string
	Reason        string
	ProjectName   string
	UserId        string
}

type ParameterURLModel struct {
	IdRequest     string
	UserId        string
	IdLeader      string
	ApproveLeader string
	Reason        string
	Level         int
}

type ParameterURLManagerModel struct {
	IdRequest      string
	UserId         string
	IdManager      string
	UserIdManager  string
	ApproveManager string
	Reason         string
	Level          int
}

type ParameterDecriptmodel struct {
	IdRequest     string
	IdLeader      string
	ApproveLeader string
	Reson         string
}

type ParameterResetPassword struct {
	UserId string
}

type AprovalRequestLeaveModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id", json:"_id"`
	IdRequest     string
	Name          string
	EmpId         string
	Designation   string
	Location      string
	Departement   string
	Reason        string
	Email         string
	Address       string
	Contact       string
	Project       []string
	YearLeave     int
	PublicLeave   int
	DateLeave     string
	DayVal        int
	MonthVal      int
	YearVal       int
	UserId        string
	IsEmergency   bool
	StsByLeader   string
	StsByManager  string
	TypeOfLeave   string
	IsDelete      bool
	RequestDelete bool
	ReasonAction  string `bson:"reasonaction,omitempty",json:"reasonaction,omitempty"`
	IsCutOff      bool
	IsReset       bool
	IsPaidLeave   bool
}

func NewAprovalRequestLeaveModel() *AprovalRequestLeaveModel {
	m := new(AprovalRequestLeaveModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *AprovalRequestLeaveModel) RecordID() interface{} {
	return e.Id
}

func (e *AprovalRequestLeaveModel) TableName() string {
	return "requestLeaveByDate"
}

// ==========================================================

func NewRequestLeave() *RequestLeaveModel {
	m := new(RequestLeaveModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *RequestLeaveModel) RecordID() interface{} {
	return e.Id
}

func (e *RequestLeaveModel) TableName() string {
	return "requestLeave"
}

// type DataResult struct {
// 	Data    interface{}
// 	Message string
// 	Summary interface{}
// }

// func (a *Dashboard) TableName() string {
// 	return "Folder"
// }

// type Dashboard struct {
// 	Id            string `bson:"_id",json:"_id"`
// 	FileName      string
// 	FileExtension string
// 	FileSize      string
// 	FileSizeStr   string
// 	FileType      string
// 	PathLocation  string
// 	// Tag           []interface{}
// 	IsFolder    string
// 	Note        string
// 	CreatedDate string
// 	CreatedBy   string
// 	UpdatedDate string
// 	UpdatedBy   string
// 	ParentId    string
// 	// SharedWith    []interface{}
// }

// type Share struct {
// 	Id            string `bson:"_id",json:"_id"`
// 	Title         string
// 	Name          string
// 	Type          string
// 	Parent        string
// 	Status        bool
// 	CreatedNameBy string
// 	CreatedIdBy   string
// 	CreatedTime   time.Time
// }

// func (a *Share) TableName() string {
// 	return "ShareLink"
// }
