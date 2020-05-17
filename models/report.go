package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type ReportAnnualGrid struct {
	Name          string
	EmpId         string
	TotalLeave    int
	TotalEleave   int
	TotalRemote   int
	TotalOvertime int
	TotalDecline  int
	ActiveDays    int
}

type ReportLeaveGrid struct {
	Name          string
	EmpId         string
	TotalLeave    int
	TotalEleave   int
	TotalDecline  int
	TotalOvertime int
	Summary       int
	ActiveDays    int
}

type ReportLeaveBetweenGrid struct {
	Name        string
	EmpId       string
	PhoneNumber string
	Project     string
	DateLeave   []DateLeaveBetween
	Isvisible   bool
}

type DateLeaveBetween struct {
	Date    string
	Leave   bool
	Holiday bool
}

type ReportRemoteGrid struct {
	Name              string
	EmpId             string
	TotalRemote       int
	FullRemote        int
	ConditionalRemote int
	RemotePerformance int
	UserId            string
}

type ReportChart struct {
	Name string
	Data []int
}

type RequestLeaveByDateModel struct {
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
	RequestLeave  []RequestLeaves
}

type RequestLeaves struct {
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

type LogCronModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id", json:"_id"`
	Typelog       string
	Date          string
	Detail        []LogCronDetailModel
}

type LogCronDetailModel struct {
	Empid   string
	Message string
	Date    string
	Err     string
}

func NewLogCronModel() *LogCronModel {
	m := new(LogCronModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *LogCronModel) RecordID() interface{} {
	return e.Id
}

func (e *LogCronModel) TableName() string {
	return "ActivityLogs"
}

func (e *LogCronModel) setM(m orm.ModelBase) {
	e.ModelBase = m
}

type Overtimereport struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id", json:"_id"`
	Empid         string
	Fullname      string
	NewOvertime   CNewOvertimeModel
	Project       string
	Sign          int
}
type CNewOvertimeModel struct {
	Id              string `bson:"_id", json:"_id"`
	UserId          string
	Name            string
	EmpId           string
	Designation     string
	Location        string
	Departement     string
	Email           string
	Address         string
	Contact         string
	DayList         ApprovemenDay
	DayDuration     int
	DateCreated     string
	Reason          string
	Project         string
	ApprovalManager PMApprove
	ProjectLeader   UserEntityOvertime
	ProjectManager  UserEntityOvertime
	BranchManagers  []BranchManagerList
	MembersOvertime UserOvertime
	DeclineReason   string
	ResultRequest   string
	IsExpired       bool
	ExpiredOn       string
	ExpiredRemining string
	IsDayOff        bool
	IsDelete        bool
	IsRequestChange bool
}
type Overtimeresult struct {
	Id           string
	Name         string
	EmpId        string
	TotalApprove int
	TotalDecline int
}

type OvertimeResultDetail struct {
	Id            string
	Name          string
	Date          string
	Expectedhours int
	Actualhours   int
	Reason        string
}

type OvertimeDetail struct {
	Id               string `bson:"_id", json:"_id"`
	DayList          ApprovemenDay
	Reason           string
	MembersOvertime  UserOvertime
	EmployeeOvertime EmployeeOvertimeModel
}

//

type OvertimeResultExport struct {
	Name    string
	EmpId   string
	Date    string
	Project string
	Reason  string
}
