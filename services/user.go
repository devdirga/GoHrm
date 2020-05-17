package services

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"encoding/base64"
	"os"
	"path/filepath"

	knot "github.com/creativelab/knot/knot.v1"

	"github.com/creativelab/dbox"

	tk "github.com/creativelab/toolkit"
)

type UserService struct{}

func (s *UserService) GetByID(userid string) (SysUserModel, error) {
	chanApp := make(chan AprovalRequestLeaveModel)
	defer close(chanApp)

	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("_id", userid))
	filter.Set("where", dbox.And(dboxFilter...))
	user := SysUserModel{}

	go func(ch chan AprovalRequestLeaveModel) {
		approvalLeave := new(repositories.LeaveDboxRepo).GetLastLeaveDate()
		ch <- approvalLeave
	}(chanApp)

	userM := repositories.UserOrmRepo{}
	users, err := userM.GetByParam(filter)
	if len(users) > 0 {
		user = users[0]
	}
	approval := <-chanApp
	user.LastLeave = approval.DateLeave

	return user, err
}

func (s *UserService) GetProfileByID(userid string) (SysUserProfileModel, error) {
	chanApp := make(chan AprovalRequestLeaveModel)
	defer close(chanApp)

	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("_id", userid))
	filter.Set("where", dbox.And(dboxFilter...))
	user := SysUserProfileModel{}

	go func(ch chan AprovalRequestLeaveModel) {
		approvalLeave := new(repositories.LeaveDboxRepo).GetLastLeaveDate()
		ch <- approvalLeave
	}(chanApp)

	userM := repositories.UserOrmRepo{}
	users, err := userM.GetByParam(filter)
	if len(users) > 0 {
		user.EmpId = users[0].EmpId
		user.Designation = users[0].Designation
		user.Departement = users[0].Departement
		user.Username = users[0].Username
		user.Fullname = users[0].Fullname
		user.PhoneNumber = users[0].PhoneNumber
		user.Email = users[0].Email
		user.Address = users[0].Address
		user.Gender = users[0].Gender
		user.YearLeave = users[0].YearLeave
		user.PublicLeave = users[0].PublicLeave
		user.Location = users[0].Location
		user.Photo = users[0].Photo
		user.LastLeave = users[0].LastLeave
	}
	approval := <-chanApp
	user.LastLeave = approval.DateLeave

	return user, err
}

func (s *UserService) UploadImage(k *knot.WebContext) (string, error) {
	return helper.UploadImage(k)
}

func (s *UserService) CheckImageExist(filename string) string {
	config := helper.ReadConfig()
	filelocation := filepath.Join(config.GetString("UploadPath"), filename)
	_, err := os.Stat(filelocation)

	if err != nil {
		return helper.StaticImgPath + "default-user.png"
	}

	return helper.StaticImgPath + filename
}

//crud
func (s *UserService) SaveUserProfile(param SysUserModel) (SysUserModel, error) {
	olduser, err := s.GetById(param.Id)
	if err != nil {
		return olduser, err
	}

	if param.Password == "" {
		param.Password = olduser.Password
	} else {
		param.Password = helper.GetMD5Hash(param.Password)
	}

	return new(repositories.UserOrmRepo).Save(param)
}
func (s *UserService) SaveUserProfileByClient(param SysUserModel) (SysUserModel, error) {
	olduser, err := s.GetById(param.Id)
	if err != nil {
		return olduser, err
	}
	param.Enable = true
	param.Roles = olduser.Roles
	param.ProjectRuleID = olduser.ProjectRuleID
	param.IsChangePassword = true
	param.Username = olduser.Username
	param.YearLeave = olduser.YearLeave
	param.PublicLeave = olduser.PublicLeave
	param.JointDate = olduser.JointDate
	param.AddLeave = olduser.AddLeave
	param.DecYear = olduser.DecYear

	if param.Password == "" {
		param.Password = olduser.Password
	} else {
		passTemp, err := base64.StdEncoding.DecodeString(param.Password)
		if err != nil {
			tk.Println(err)
		}
		param.Password = helper.GetMD5Hash(string(passTemp))
	}

	return new(repositories.UserOrmRepo).Save(param)
}

func (s *UserService) Save(param SysUserModel) (SysUserModel, error) {
	olduser, err := s.GetById(param.Id)
	if err != nil {
		return olduser, err
	}

	if param.Password == "" {
		param.Password = olduser.Password
	}

	return new(repositories.UserOrmRepo).Save(param)
}

//end crud

func (s *UserService) GetAll() ([]SysUserModel, error) {
	return new(repositories.UserOrmRepo).GetByParam(tk.M{})
}

func (s *UserService) GetById(userid string) (SysUserModel, error) {
	return new(repositories.UserOrmRepo).GetByID(userid)
}
