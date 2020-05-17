package controllers

import (
	. "creativelab/ecleave-dev/models"
	"time"

	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

type NationalHolidayController struct {
	*BaseController
}

func (c *NationalHolidayController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	viewData := tk.M{}
	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel").(int))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
	}

	DataAccess := c.SetViewData(k, viewData)
	// k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = []string{
		"_modal.html",
		"_loader.html",
	}
	// DataAccess := Previlege{}

	// for _, o := range access {
	// 	DataAccess.Create = o["Create"].(bool)
	// 	DataAccess.View = o["View"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Process = o["Process"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Edit = o["Edit"].(bool)
	// 	DataAccess.Menuid = o["Menuid"].(string)
	// 	DataAccess.Menuname = o["Menuname"].(string)
	// 	DataAccess.Approve = o["Approve"].(bool)
	// 	DataAccess.Username = o["Username"].(string)
	// }

	return DataAccess
}
func (c *NationalHolidayController) GetDataNH(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Year int
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	csr, e := c.Ctx.Connection.NewQuery().Select().From("NationalHolidays").Where(db.Eq("year", p.Year)).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	results := []NationalHolidaysModel{}
	e = csr.Fetch(&results, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "Success", results)
}
func (c *NationalHolidayController) DeleteHoliday(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Id bson.ObjectId
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	result := new(NationalHolidaysModel)
	e = c.Ctx.GetById(result, p.Id)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	e = c.Ctx.Delete(result)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "success", nil)
}
func (c *NationalHolidayController) SaveHoliday(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Data    NationalHolidaysModel
		ListStr []string
		DateStr string
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	newlistStr := []time.Time{}
	for _, each := range p.ListStr {
		date, _ := time.Parse("2006 01 02", each)
		dt, err := c.CheckDateNH(k, p.Data.Location, date)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		if len(dt) > 0 {
			return c.SetResultInfo(true, "Some date Has been filled", nil)
		}
		newlistStr = append(newlistStr, date)
	}
	newDate, _ := time.Parse("2006 01 02", p.DateStr)
	data := p.Data
	if data.Id == "" {
		data.Id = bson.NewObjectId()
	}
	data.Date = newDate
	data.ListDate = newlistStr
	e = c.Ctx.Save(&data)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "Success", nil)
}
func (c *NationalHolidayController) CheckDateNH(k *knot.WebContext, location string, date time.Time) ([]*NationalHolidaysModel, error) {
	k.Config.OutputType = knot.OutputJson
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*NationalHolidaysModel, 0)
	// fmt.Println("---------------- date", date)
	dbFilter = append(dbFilter, db.Eq("location", location))
	dbFilter = append(dbFilter, db.Eq("date", date))

	if len(dbFilter) > 0 {

		query.Set("where", db.And(dbFilter...))

	}

	crs, errdata := c.Ctx.Find(NewNationalHolidaysModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil, nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil, nil
	}
	// fmt.Println("---------------- data", data)
	return data, nil

}
func (c *NationalHolidayController) GetListLocation(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	csr, e := c.Ctx.Connection.NewQuery().Select().From(NewLocationModel().TableName()).Cursor(nil)
	defer csr.Close()
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	results := []LocationModel{}
	e = csr.Fetch(&results, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "Success", results)
}
