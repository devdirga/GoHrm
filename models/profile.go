package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
	// toolkit "github.com/creativelab/toollkit"
	"fmt"
)

type IdentityModel struct {
	orm.ModelBase    `bson:"-",json:"-"`
	Id               bson.ObjectId ` bson:"_id" , json:"_id" `
	Id_card          string
	Name             string
	Dob              string
	Religion         string
	Nationality      string
	Bank_account     string
	Account_number   string
	Gender           string
	Marital_status   string
	Height           string
	Weight           string
	Phone            string
	Email            string
	Blood            string
	Residence_status string
	Address          string
	Father           string
	Mother           string
	Brother_sister   []string
	Husband_wife     string
	Child            []string
	Education        []DetailEducation
	Language         []LanguageSkill
	Why_join         string
	Name_friend      string
	When_join        string
	Status           string
	Date_invite      string
	Time_invite      string
	Experience       []ExperienceDetails
	Image_file       string
}

type DetailEducation struct {
	Level string
	Name  string
	// City  string
	Start string
	End   string
}

type LanguageSkill struct {
	Language string
	Write    string
	Read     string
	Speaking string
}

type ExperienceDetails struct {
	Company_name   string
	Start          string
	End            string
	Employee_count string
	Resign_reason  string
	Salary         string
}

func NewProfileModel() *IdentityModel {
	fmt.Println("masuk model")
	m := new(IdentityModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *IdentityModel) RecordID() interface{} {
	return e.Id
}

func (m *IdentityModel) TableName() string {
	return "profile"
}
