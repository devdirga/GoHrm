package controllers

import (

	// "fmt"
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
)

type UserProfileController struct {
	*BaseController
}

func (c *UserProfileController) SaveProfile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	userdata, _ := c.GetUserByIDSession(k)
	p := new(UserProfileModel)
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
	}
	passTemp, errDecr := base64.StdEncoding.DecodeString(p.Password)
	if err != nil {
		return c.SetResultInfo(true, errDecr.Error(), nil)
	}

	p.Password = helper.GetMD5Hash(string(passTemp))

	err = c.Ctx.Save(p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	userdata[0].Designation = p.Designation
	userdata[0].Departement = p.Departement
	userdata[0].Fullname = p.FirstName + " " + p.LastName
	userdata[0].PhoneNumber = p.PhoneNo
	userdata[0].Email = p.Email
	userdata[0].Gender = p.Gender
	userdata[0].Password = helper.GetMD5Hash(string(passTemp))
	userdata[0].Location = p.Location
	userdata[0].IsChangePassword = true
	userdata[0].EmpId = p.EmployeeID
	userdata[0].Address = p.Address

	err = c.Ctx.Save(userdata[0])

	return c.SetResultInfo(false, "data save successfully", nil)
}

func (c *UserProfileController) GetUserByIDSession(k *knot.WebContext) ([]*SysUserModel, error) {
	k.Config.OutputType = knot.OutputJson
	userid := k.Session("userid")
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserModel, 0)
	if userid != "nil" {

		dbFilter = append(dbFilter, db.Eq("_id", userid))

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewSysUserModel(), query)
		if crs != nil {
			defer crs.Close()
		} else {
			return nil, nil
		}
		defer crs.Close()
		if errdata != nil {
			return nil, errdata
		}

		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return nil, errdata
		}

	} else {
		fmt.Println("no userid")
	}

	return data, nil
}

func (c *UserProfileController) GetAllUser(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	data := make([]*SysUserModel, 0)
	crs, errdata := c.Ctx.Find(NewSysUserModel(), nil)

	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	defer crs.Close()
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}

	return data
}

func (c *UserProfileController) ImportDataUser(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	t, err := os.Getwd()
	if err != nil {
		return err.Error()
	}

	pathFile := filepath.Join(t, "assets", "doc", "data_employee.xlsx")
	xlFile, err := xlsx.OpenFile(pathFile)
	if err != nil {
		return err.Error()
	}

	p := new(SysUserModel)

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {

			p.Id = bson.NewObjectId().Hex()
			for i, cell := range row.Cells {
				text := cell.String()
				// fmt.Printf("%s", i)
				// fmt.Printf("%s\n", text)
				switch i {
				case 0:
					p.EmpId = text
				case 1:
					p.Designation = text
				case 2:
					p.Departement = text
				case 3:
					p.Username = text
				case 4:
					p.Fullname = text
				case 5:
					p.Enable = true
				case 6:
					p.PhoneNumber = text
				case 7:
					p.Email = text
				case 8:
					p.YearLeave = 12
				case 9:
					p.PublicLeave = 12
				case 10:
					p.Roles = text
				case 11:
					p.Password = text
				case 12:
					if text == "EC-SBY" {
						p.Location = "Indonesia"
					} else {
						p.Location = "Singapore"
					}
				case 14:
					p.IsChangePassword = false
				}
			}
			err := c.Ctx.Save(p)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

		}
	}

	return ""
}

func (c *UserProfileController) RemainingLeaveLeft(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	t, err := os.Getwd()
	if err != nil {
		return err.Error()
	}

	pathFile := filepath.Join(t, "assets", "doc", "sisa_cuti_2018.xlsx")
	xlFile, err := xlsx.OpenFile(pathFile)
	if err != nil {
		return err.Error()
	}

	p := new(SysUserModel)

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {

			p.Id = bson.NewObjectId().Hex()
			for i, cell := range row.Cells {
				text := cell.String()
				// fmt.Printf("%s", i)
				// fmt.Printf("%s\n", text)
				switch i {
				case 0:
					p.EmpId = text
				case 1:
					p.Designation = text
				case 2:
					p.Departement = text
				case 3:
					p.Username = text
				case 4:
					p.Fullname = text
				case 5:
					p.Enable = true
				case 6:
					p.PhoneNumber = text
				case 7:
					p.Email = text
				case 8:
					p.YearLeave = 12
				case 9:
					p.PublicLeave = 12
				case 10:
					p.Roles = text
				case 11:
					p.Password = text
				case 12:
					if text == "EC-SBY" {
						p.Location = "Indonesia"
					} else {
						p.Location = "Singapore"
					}
				case 14:
					p.IsChangePassword = false
				}
			}
			err := c.Ctx.Save(p)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

		}
	}

	return ""
}
