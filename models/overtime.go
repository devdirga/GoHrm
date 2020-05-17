package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type OvertimeFormModel struct {
	orm.ModelBase      `bson:"-",json:"-"`
	Id                 bson.ObjectId `bson:"_id", json:"_id"`
	IdRequest          string
	UserId             string
	Name               string
	EmpId              string
	Designation        string
	Location           string
	Departement        string
	Email              string
	Address            string
	Contact            string
	DateOvertimeString string
	DateCreated        time.Time
	DateOvertime       time.Time
	From               string
	To                 string
	OvertimeHours      int
	Reason             string
	Project            string
	ProjectLeader      UserEntityOvertime
	ProjectManager     UserEntityOvertime
	ApprovalManager    UserEntityOvertime
	BranchManagers     []BranchManagerList
	IsLeaderReceive    bool
	IsLeaderApprove    bool
	IsManagerReceive   bool
	IsManagerApprove   bool
	DeclineReason      string
	IsExpired          bool
	ExpiredOn          string
	ExpiredRemining    string
	IsDayOff           bool
	IsDelete           bool
	IsRequestChange    bool
}
type UserEntityOvertime struct {
	IdEmp       string
	Name        string
	Location    string
	Email       string
	PhoneNumber string
	UserId      string
}

func NewOvertimeFormModel() *OvertimeFormModel {
	m := new(OvertimeFormModel)
	m.Id = bson.NewObjectId()
	return m
}
func (e *OvertimeFormModel) RecordID() interface{} {
	return e.Id
}

func (m *OvertimeFormModel) TableName() string {
	return "Overtime"
}

// ======================== new overtime ====================

type OvertimeModel struct {
	orm.ModelBase   `bson:"-",json:"-"`
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
	DayList         []ApprovemenDay
	DayDuration     int
	DateCreated     string
	Reason          string
	Project         string
	ApprovalManager PMApprove
	ProjectLeader   UserEntityOvertime
	ProjectManager  UserEntityOvertime
	BranchManagers  []BranchManagerList
	MembersOvertime []UserOvertime
	DeclineReason   string
	ResultRequest   string
	IsExpired       bool
	ExpiredOn       string
	ExpiredRemining string
	IsDayOff        bool
	IsDelete        bool
	IsRequestChange bool
}

type ApprovemenDay struct {
	Date   string
	Result string
	Reason string
}

type ManagerApprove struct {
	IdEmp       string
	Name        string
	Location    string
	Email       string
	PhoneNumber string
	UserId      string
}

type UserOvertime struct {
	IdEmp        string
	Name         string
	Location     string
	Email        string
	PhoneNumber  string
	UserId       string
	TypeOvertime string
	Hours        int
	Result       string
}

type PMApprove struct {
	IdEmp       string
	Name        string
	Location    string
	Email       string
	PhoneNumber string
	UserId      string
	Result      string
	Reason      string
}

type ParameterURLUserModel struct {
	IdOvertime string
	UserId     string
	Name       string
	Project    string
	Result     string
}

type dtDetailsInput struct {
	Date     string
	Start    string
	End      string
	Type     string
	DeadLine string
	Task     string
	Hours    string
}
type DetailPayloadInput struct {
	Param        string
	Typeovertime string
	Hours        int
	DateDetails  []dtDetailsInput
}

func NewOvertimeModel() *OvertimeModel {
	m := new(OvertimeModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}
func (e *OvertimeModel) RecordID() interface{} {
	return e.Id
}

func (m *OvertimeModel) TableName() string {
	return "NewOvertime"
}
