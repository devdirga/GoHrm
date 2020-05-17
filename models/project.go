package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type ProjectModel struct {
	orm.ModelBase  `bson:"-",json:"-"`
	Id             string          `bson:"_id",json:"_id"`
	ProjectKey     string          `bson:"ProjectKey",json:"ProjectKey"`
	ProjectName    string          `bson:"ProjectName",json:"ProjectName"`
	ProjectManager OcupationData   `bson:"ProjectManager",json:"ProjectManager"`
	SPVManager     []OcupationData `bson:"SPVManager,omitempty",json:"SPVManager,omitempty"`
	// DeveloperManager OcupationData   `bson:"DeveloperManager",json:"DeveloperManager"`
	BusinessAnalist []OcupationData `bson:"BusinessAnalist",json:"BusinessAnalist"`
	ProjectLeader   OcupationData   `bson:"ProjectLeader",json:"ProjectLeader"`
	Developer       []OcupationData `bson:"Developer",json:"Developer"`
	Location        string          `bson:"Location",json:"Location"`
	Address         string          `bson:"Address",json:"Address"`
	Uri             string          `bson:"Uri",json:"Uri"`
	Active          bool            `bson:"Active",json:"Active"`
	Photo           string          `bson:"Photo",json:"Photo"`
}

type OcupationData struct {
	UserId      string `bson:"userid",json:"userid"`
	IdEmp       string `bson:"IdEmp",json:"IdEmp"`
	Name        string `bson:"Name",json:"Name"`
	Location    string `bson:"Location",json:"Location"`
	Email       string `bson:"Email",json:"Email"`
	PhoneNumber string `bson:"PhoneNumber",json:"PhoneNumber"`
}
type Location struct {
	Id   string
	Name string
}

type ListUnidentifiedName struct {
	ProjectName     string `bson:"ProjectName",json:"ProjectName"`
	ProjectManager  string `bson:"ProjectManager",json:"ProjectManager"`
	BusinessAnalist string `bson:"BusinessAnalist",json:"BusinessAnalist"`
	ProjectLeader   string `bson:"ProjectLeader",json:"ProjectLeader"`
	Developer       string `bson:"Developer",json:"Developer"`
}

func NewListProject() *ProjectModel {
	m := new(ProjectModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *ProjectModel) RecordID() interface{} {
	return e.Id
}

func (e *ProjectModel) TableName() string {
	return "ProjectProfile"
}
