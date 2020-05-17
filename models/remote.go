package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type RemoteModel struct {
	orm.ModelBase   `bson:"-",json:"-"`
	Id              string              `bson:"_id",json:"_id"`
	IdOp            string              `bson:"idop",json:"IdOp"`
	UserId          string              `bson:"userid",json:"UserId"`
	Name            string              `bson:"name",json:"Name"`
	Email           string              `bson:"email",json:"Email"`
	Location        string              `bson:"location",json:"Location"`
	Type            string              `bson:"type",json:"Type"`
	From            string              `bson:"from",json:"From"`
	To              string              `bson:"to",json:"To"`
	Projects        []Project           `bson:"projects",json:"Projects"`
	BranchManager   []BranchManagerList `bson:"branchmanager",json:"branchmanager"`
	SPVManager      []BranchManagerList `bson:"spvmanager,omitempty",json:"spvmanager,omitempty"`
	BAnalis         []BranchManagerList `bson:"banalis,omitempty",json:"banalis,omitempty"`
	Address         string              `bson:"address",json:"Address"`
	Contact         string              `bson:"contact",json:"Contact"`
	DateLeave       string              `bson:"dateleave",json:"DateLeave"`
	Reason          string              `bson:"reason",json:"Reason"`
	IsRequestChange bool                `bson:"isrequestchange",json:"IsRequestChange"`
	CreatedAt       time.Time           `bson:"createdat",json:"CreatedAt"`
	UpdatedAt       time.Time           `bson:"updatedat",json:"UpdatedAt"`
	IsDelete        bool                `bson:"isdelete",json:"IsDelete"`
	IsExpired       bool                `bson:"isexpired",json:"IsExpired"`
	ReasonAction    string              `bson:"ReasonAction,omitempty",json:"ReasonAction,omitempty"`
	ExpiredOn       string              `bson:"expiredon,omitempty",json:"ExpiredOn,omitempty"`
	ExpRemaining    string              `bson:"expremining,omitempty",json:"ExpRemaining,omitempty"`
}

type Project struct {
	Id                string `bson:"id",json:"Id"`
	ProjectName       string `bson:"projectname",json:"ProjectName"`
	ProjectLeader     User   `bson:"projectleader",json:"ProjectLeader"`
	ProjectManager    User   `bson:"projectmanager",json:"ProjectManager"`
	IsLeaderSend      bool   `bson:"isleadersend",json:"IsLeaderSend"`
	IsManagerSend     bool   `bson:"ismanagersend",json:"IsManagerSend"`
	IsApprovalLeader  bool   `bson:"isapprovalleader",json:"IsApprovalLeader"`
	IsApprovalManager bool   `bson:"isapprovalmanager",json:"IsApprovalManager"`
	NoteManager       string `bson:"notemanager",json:"NoteManager"`
	NoteLeader        string `bson:"noteleader",json:"NoteLeader"`
}

type User struct {
	IdEmp       string `bson:"idemp",json:"IdEmp"`
	UserId      string `bson:"userid",json:"UserId"`
	Name        string `bson:"name",json:"Name"`
	PhoneNumber string `bson:"phonenumber",json:"PhoneNumber"`
	Email       string `bson:"email",json:"Email"`
	Location    string `bson:"location",json:"Location"`
}

func NewRemoteModel() *RemoteModel {
	m := new(RemoteModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *RemoteModel) RecordID() interface{} {
	return e.Id
}

func (e *RemoteModel) TableName() string {
	return "remote"
}
