package controllers

import (
	. "creativelab/ecleave-dev/models"
	"time"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type LogLeaveRemoteController struct {
	*BaseController
}

func (c *LogLeaveRemoteController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	viewData := tk.M{}
	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("UserId", k.Session("userid").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel"))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
		viewData.Set("UserId", "")
	}

	DataAccess := c.SetViewData(k, viewData)

	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = []string{
		"_loader.html",
	}
	// if access != nil {
	// 	tk.Println("sdsfsdf")
	// 	e := tk.Serde(access, DataAccess, "json")
	// 	if e != nil {
	// 		tk.Println(e.Error(), "<<")
	// 	}
	// }
	// fmt.Println("---------------------", k.Session("userid").(string))
	return DataAccess
}
func (c *LogLeaveRemoteController) GetDataLog(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	p := struct {
		MonthYear string
		Type      string
		Level     string
		UserId    string
		Location  string
		Name      []interface{}
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	startDate, _ := time.Parse("02-Jan-2006", "01-"+p.MonthYear)
	endDate := startDate.AddDate(0, 1, 0)
	andfilter := []*db.Filter{}
	// andfilter = append(andfilter, db.Ne("typerequest", "overtime"))
	if p.Type != "All" {
		andfilter = append(andfilter, db.Contains("typerequest", p.Type))
	}
	andfilter = append(andfilter, db.Gte("datelogcreated", startDate))
	andfilter = append(andfilter, db.Lt("datelogcreated", endDate))
	if p.Level != "" {
		if p.Level != "5" && p.Level != "6" {
			andfilter = append(andfilter, db.Eq("userid", p.UserId))
		}
		if p.Level != "5" {
			andfilter = append(andfilter, db.Eq("location", p.Location))
		}
	}
	if len(p.Name) > 0 {
		andfilter = append(andfilter, db.In("name", p.Name...))
	}

	filter := db.And(andfilter...)

	csr, e := c.Ctx.Connection.NewQuery().Select().From(NewLogLeaveRemoteModel().TableName()).Where(filter).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	results := []LogleaveRemoteModel{}
	e = csr.Fetch(&results, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	tk.Println("------------------------- datatemp ", results)
	arrTempt := []LogleaveRemoteNewModel{}
	for _, dt := range results {
		datatempt := LogleaveRemoteNewModel{}
		datatempt.Id = dt.Id
		datatempt.IdRequest = dt.IdRequest
		datatempt.TypeRequest = dt.TypeRequest
		datatempt.Userid = dt.Userid
		datatempt.Name = dt.Name
		datatempt.Email = dt.Email
		datatempt.DateLogCreated = dt.DateLogCreated
		datatempt.ListLog = dt.ListLog
		datatempt.StatusRequest = dt.StatusRequest
		datatempt.DateFrom = dt.DateFrom
		datatempt.DateTo = dt.DateTo
		datatempt.Project = dt.Project
		datatempt.Location = dt.Location
		datatempt.DataLeave = c.GetDataLeaveById(k, dt.IdRequest)
		arrTempt = append(arrTempt, datatempt)

	}

	return c.SetResultInfo(false, "Success", arrTempt)
}

func (c *LogLeaveRemoteController) GetDataLeaveById(k *knot.WebContext, IdRequest string) []RequestLeaveModel {
	k.Config.OutputType = knot.OutputJson

	dataLeave := []RequestLeaveModel{}
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("_id", IdRequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return dataLeave
	}
	if err != nil {
		return dataLeave
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		return dataLeave
	}

	return dataLeave
}
