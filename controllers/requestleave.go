package controllers

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"

	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type RequestLeaveController struct {
	*BaseController
}

func (c *RequestLeaveController) GetDataRequestLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := services.PayloadDashboardService{}

	err := k.GetPayload(&payload)

	if err != nil {
		return err.Error
	}

	level := k.Session("jobrolelevel").(int)
	payload.Level = level

	if level != 5 && level != 6 {
		tk.Println("--------------- masuk ", level)
		payload.UserId = k.Session("userid").(string)
		if k.Session("empid") != nil {
			payload.EmpId = k.Session("empid").(string)
		}
	}
	datas, err := new(services.DashboardService).ConstructDashboardData(payload)
	// tk.Println(err)
	if err != nil {
		return err
	}

	return datas
}

func (c *RequestLeaveController) GetDataUserProject(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	level := k.Session("jobrolelevel").(int)
	userid := k.Session("userid").(string)
	locat := k.Session("location").(string)
	datas := []ProjectModel{}
	var dbFilter []*db.Filter
	query := tk.M{}
	if level != 5 {
		dash := DashboardController(*c)
		dataUser := dash.GetDataSessionUser(k, userid)

		pipe := []tk.M{}

		switch level {
		case 2:
			pipe = append(pipe, tk.M{"$match": tk.M{"ProjectLeader.userid": userid}})
		case 1:
			pipe = append(pipe, tk.M{"$match": tk.M{"ProjectManager.userid": userid}})
		case 6:
			pipe = append(pipe, tk.M{"$match": tk.M{"Location": dataUser[0].Location}})
		}

		crs, err := c.Ctx.Connection.NewQuery().From("ProjectProfile").Command("pipe", pipe).Cursor(nil)
		if crs != nil {
			defer crs.Close()
		}

		if err != nil {
			return c.SetResultInfo(true, "", nil)
		}

		err = crs.Fetch(&datas, 0, false)
		if err != nil {
			return c.SetResultInfo(true, "", nil)
		}

		return c.SetResultInfo(false, "success", datas)
	}

	if locat != "Global" {
		dbFilter = append(dbFilter, db.Eq("Location", locat))
	}

	if len(dbFilter) > 0 {

		query.Set("where", db.And(dbFilter...))

	}

	crs, errdata := c.Ctx.Find(NewListProject(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "", nil)

	}
	// defer crs.Close()
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)

	}

	errdata = crs.Fetch(&datas, 0, false)
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)

	}

	return c.SetResultInfo(false, "success", datas)
}
