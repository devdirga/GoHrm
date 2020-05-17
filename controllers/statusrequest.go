package controllers

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"io"
	"os"
	"path/filepath"

	// "creativelab/ecleave-dev/services"
	"fmt"
	"time"

	// "fmt"

	// "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"
	"errors"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type StatusRequestController struct {
	*BaseController
}

func (c *StatusRequestController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	viewData := tk.M{}

	userset := UserSettingController(*c)
	dataUser, err := userset.GetOptionUser(k, k.Session("userid").(string))
	if err != nil {
	}

	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel").(int))

		// viewData.Set("RemoteActive", k.Session("onremote"))
		// viewData.Set("ConditionalRemote", k.Session("conditionalremote"))
		// viewData.Set("FullMonth", k.Session("fullmonth"))
		// viewData.Set("Monthly", k.Session("monthly"))
		if len(dataUser) > 0 {
			viewData.Set("RemoteActive", dataUser[0].Remote.RemoteActive)
			viewData.Set("ConditionalRemote", dataUser[0].Remote.ConditionalRemote)
			viewData.Set("FullMonth", dataUser[0].Remote.FullMonth)
			viewData.Set("Monthly", dataUser[0].Remote.Monthly)
		}

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

// db.getCollection('remote').aggregate([
//     {"$match":{"userid": "5bab38ef25149720381cdaf4", "dateleave": {"$regex": "2018"}}},
//     {"$group":{"_id":{"result":"$projects.isapprovalmanager"}, "count":{"$sum": 1}}}
// ])

func (c *StatusRequestController) GetStatusRequest(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	fmt.Println("---------------------", k.Session("jobrolelevel"))
	Dashboard := DashboardController(*c)
	userid := tk.ToString(k.Session("userid"))
	dataHistory := Dashboard.GetHistoryLeave(k, userid, k.Session("jobrolelevel").(int))

	return dataHistory

}

func (c *StatusRequestController) GetRemoteThisYear(k *knot.WebContext, userid string, time string) (interface{}, error) {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "dateleave": tk.M{"$regex": time}}})
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"idop": "$idop"}, "count": tk.M{"$sum": 1}}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("remote").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return nil, errors.New("bad query")
	}
	if err != nil {
		return nil, err
	}

	data := []tk.M{}
	if err = csr.Fetch(&data, 0, false); err != nil {
		return nil, err
	}

	return data, nil

}

func (c *StatusRequestController) ResetCountLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Id    string
		Reset int
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	dataLeave, er := c.GetLeaveUnReset(k, payload.Id)

	if er != "" {
		return c.SetResultInfo(true, er, nil)
	}

	// tk.Println("--------- dataleave ", dataLeave)

	if len(dataLeave) > 0 {

		if dataLeave[0].IsReset == true {
			return c.SetResultInfo(true, "Already reset", nil)
		}

		if dataLeave[0].NoOfDays < payload.Reset {
			return c.SetResultInfo(true, "No of days less than Days cut off", nil)
		}

		if payload.Reset == 0 {
			return c.SetResultInfo(true, "Please select Days cut off", nil)
		}

		dataLeave[0].IsReset = true
		err := c.Ctx.Save(dataLeave[0])
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
		// noOfDays := dataLeave[0].NoOfDays

		dash := DashboardController(*c)
		user := dash.GetDataSessionUser(k, dataLeave[0].UserId)

		if user[0].DecYear == 0 && user[0].YearLeave > 0 {
			user[0].DecYear = float64(user[0].YearLeave)
		}

		decYear := user[0].DecYear + float64(payload.Reset)

		user[0].DecYear = decYear
		user[0].YearLeave = int(decYear)

		err = c.Ctx.Save(user[0])
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
		return c.SetResultInfo(false, "Success save", nil)
	}

	return c.SetResultInfo(true, "No Data", nil)

}

func (c *StatusRequestController) GetLeaveUnReset(k *knot.WebContext, id string) ([]*RequestLeaveModel, string) {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	data := make([]*RequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("_id", id))
	// dbFilter = append(dbFilter, db.Eq("isreset", false))

	if len(dbFilter) > 0 {

		query.Set("where", db.And(dbFilter...))

	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return data, ""
	}
	// defer crs.Close()
	if err != nil {
		return data, err.Error()
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data, err.Error()
	}

	return data, ""
}

func (c *StatusRequestController) GetLeaveRemoteThisYear(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	pipeLeave := []tk.M{}
	res := tk.M{}
	// var dbFilter []*db.Filter
	// query := tk.M{}
	curtime := time.Now()
	userid := k.Session("userid")
	level := k.Session("jobrolelevel").(int)

	if userid != nil {
		if level == 1 || level == 5 || level == 6 {
			pipeLeave = append(pipeLeave, tk.M{"$match": tk.M{"datecreateleave": tk.M{"$regex": tk.ToString(curtime.Year())}, "resultrequest": tk.M{"$ne": "Pending"}}})
			pipeLeave = append(pipeLeave, tk.M{"$group": tk.M{"_id": tk.M{"resulrequest": "$resultrequest", "isemergency": "$isemergency"}, "count": tk.M{"$sum": 1}}})
		} else {
			pipeLeave = append(pipeLeave, tk.M{"$match": tk.M{"userid": tk.ToString(userid), "datecreateleave": tk.M{"$regex": tk.ToString(curtime.Year())}, "resultrequest": tk.M{"$ne": "Pending"}}})
			pipeLeave = append(pipeLeave, tk.M{"$group": tk.M{"_id": tk.M{"resulrequest": "$resultrequest", "isemergency": "$isemergency"}, "count": tk.M{"$sum": 1}}})
		}

		csr, err := c.Ctx.Connection.
			NewQuery().
			Command("pipe", pipeLeave).
			From("requestLeave").
			Cursor(nil)

		if csr != nil {
			defer csr.Close()
		} else {
			return c.SetResultInfo(true, "bad  query", nil)
		}
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		data := []tk.M{}
		if err = csr.Fetch(&data, 0, false); err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		res.Set("leave", data)

		remote, err := c.GetRemoteThisYear(k, tk.ToString(userid), tk.ToString(curtime.Year()))

		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		res.Set("remote", remote)

		return c.SetResultInfo(false, "success", res)
	}

	return c.SetResultInfo(true, "no userid", nil)

}

func (c *StatusRequestController) GetPendingRequestByUser(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	Dashboard := DashboardController(*c)
	userid := tk.ToString(k.Session("userid"))
	// fmt.Println("--------------- userid", userid)
	dataHistory, err := Dashboard.PendingRequest(k, userid)

	if err != nil {
		return err
	}

	return dataHistory
}

func (c *StatusRequestController) GetRemoteLogByUserId(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	history := HistoryLeaveModel{}
	var userid interface{}

	if userid = k.Session("userid"); userid == nil {
		return c.SetResultError(errors.New("Session expired").Error(), history)
	}
	jobrolelevel := k.Session("jobrolelevel")
	// if level = k.Session("jobrolelevel"); level == nil {
	// 	return c.SetResultError(errors.New("Session expired").Error(), history)
	// }
	if jobrolelevel.(int) == 5 {
		historys := []HistoryLeaveModel{}
		historys, err := new(services.HistoryService).GetHistoryForAdmin("", jobrolelevel.(int))
		if err != nil {
			return c.SetResultError(err.Error(), history)
		}
		return c.SetResultOK(historys)
	} else {
		history, err := new(services.HistoryService).GetHistoryByUserId(userid.(string))
		if err != nil {
			return c.SetResultError(err.Error(), history)
		}
		return c.SetResultOK(history)
	}

	return c.SetResultOK(history)
}

func (c *StatusRequestController) GetDataHistory(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		DateByMonth string
		UserId      string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	service := new(services.LogUserService)
	service.DateMonth = payload.DateByMonth
	service.UserId = payload.UserId

	leaves, err := service.ConstructDashboardLogUser()
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(leaves)
}
func (c *StatusRequestController) GetDataDetailLeaveByID(k *knot.WebContext, idLeaveByDate string) *AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	data := make([]*AprovalRequestLeaveModel, 0)
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("_id", idLeaveByDate))

	query := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data[0]
}

func (c *StatusRequestController) GetDateDetailLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		UserId string
	}{}

	// tk.Println("---------------- ", k.Session("jobrolelevel"))
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	level := k.Session("jobrolelevel")
	data := make([]*AprovalRequestLeaveModel, 0)
	if level == 2 || level == 3 {
		var dbFilter []*db.Filter
		dbFilter = append(dbFilter, db.Eq("userid", payload.UserId))
		// dbFilter = append(dbFilter, db.Eq("isemergency", false))

		query := tk.M{}

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
		defer crs.Close()
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}

	} else if level == 1 || level == 5 || level == 6 {
		crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), nil)
		defer crs.Close()
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
	}

	return c.SetResultInfo(false, "sukses", data)
}

func (c *StatusRequestController) GetDataDetailLeaveByDate(k *knot.WebContext, userId string, yearVal int, monthVal int, dayVal int) *AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	data := make([]*AprovalRequestLeaveModel, 0)
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("userid", userId))
	dbFilter = append(dbFilter, db.Eq("yearval", yearVal))
	dbFilter = append(dbFilter, db.Eq("monthval", monthVal))
	dbFilter = append(dbFilter, db.Eq("dayval", dayVal))

	query := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data[0]
}

func (c *StatusRequestController) GetLeaveDateByIdReq(k *knot.WebContext, idrequest string) []*AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	data := make([]*AprovalRequestLeaveModel, 0)
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("idrequest", idrequest))

	query := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data
}

func (c *StatusRequestController) CancelRequest(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		IdRequest string
		Name      string
		Email     string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("----------- payload ", payload.IdRequest)

	data := make([]*RequestLeaveModel, 0)
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("_id", payload.IdRequest))

	query := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewRequestLeave(), query)
	defer crs.Close()
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}

	level := k.Session("jobrolelevel").(int)

	// tk.Println("----------- 0 ", data)
	dash := DashboardController(*c)
	mail := MailController(*c)

	noOfDays := 0
	t := time.Now()
	datenow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if len(data) > 0 {
		dataDate := c.GetLeaveDateByIdReq(k, payload.IdRequest)
		for _, dt := range dataDate {
			dt.IsDelete = true

			date, _ := time.Parse("2006-01-02", dt.DateLeave)
			if date.Before(datenow) || dt.DateLeave == datenow.Format("2006-01-02") {
				if dt.StsByManager == "Approved" {
					noOfDays++
				}
			}

			err = c.Ctx.Save(dt)
			if err != nil {
				return c.SetResultInfo(true, errdata.Error(), nil)
			}
		}
	}

	if level == 5 || level == 6 || level == 1 {
		if data[0].IsReset != true {
			dataUser := dash.GetDataSessionUser(k, data[0].UserId)
			dataUser[0].YearLeave = dataUser[0].YearLeave + noOfDays
			dataUser[0].DecYear = dataUser[0].DecYear + float64(noOfDays)
			errUser := c.Ctx.Save(dataUser[0])
			if errUser != nil {
				return c.SetResultInfo(true, errUser.Error(), nil)
			}
		}
	} else {
		if len(data) > 0 {
			if data[0].ResultRequest != "Pending" {
				return c.SetResultInfo(true, "Your request already "+data[0].ResultRequest+" by Manager", nil)
			} else {
				// nu := 0
				for _, dt := range data[0].StatusProjectLeader {
					if dt.StatusRequest == "Approved" {
						return c.SetResultInfo(true, "Your request already "+dt.StatusRequest+" by Leader", nil)
					}
				}
			}
		}
	}

	nameSes := dash.GetDataSessionUser(k, k.Session("userid").(string))

	if len(data) > 0 {
		switch level {
		case 5:
			tk.Println("---------- masuk admin")
			data[0].ResultRequest = "Canceled by Admin"
			dash.SetHistoryLeave(k, data[0].UserId, data[0].Id, data[0].LeaveFrom, data[0].LeaveTo, "Request Canceled by Admin", "Canceled", data[0])
			mail.SendMailCancelLeave(k, data[0], 5, nameSes[0].Fullname)
		case 1:
			data[0].ResultRequest = "Canceled by Manager"
			dash.SetHistoryLeave(k, data[0].UserId, data[0].Id, data[0].LeaveFrom, data[0].LeaveTo, "Request Canceled by Manager", "Canceled", data[0])
			mail.SendMailCancelLeave(k, data[0], 1, nameSes[0].Fullname)
		case 6:
			data[0].ResultRequest = "Canceled by Manager"
			dash.SetHistoryLeave(k, data[0].UserId, data[0].Id, data[0].LeaveFrom, data[0].LeaveTo, "Request Canceled by Manager", "Canceled", data[0])
			mail.SendMailCancelLeave(k, data[0], 6, nameSes[0].Fullname)
		default:
			data[0].ResultRequest = "Canceled by User"
			dash.SetHistoryLeave(k, data[0].UserId, data[0].Id, data[0].LeaveFrom, data[0].LeaveTo, "Request Canceled by User", "Canceled", data[0])
		}

		// tk.Println("-----------", data[0].ResultRequest)

		err = c.Ctx.Save(data[0])
		if err != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}

		//log
		emptyremote := new(RemoteModel)
		service := services.LogService{
			data[0],
			emptyremote,
			"leave",
		}
		log := tk.M{}
		log.Set("Status", "Cancel")
		log.Set("Desc", "Request "+data[0].ResultRequest)
		log.Set("NameLogBy", payload.Name)
		log.Set("EmailNameLogBy", payload.Email)
		err = service.CancelRequest(log)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
	} else {
		return c.SetResultInfo(true, "data empty", nil)
	}

	return c.SetResultInfo(false, "Your Request has been Cancel successfully", nil)
}

func (c *StatusRequestController) GetDataRequest(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := struct {
		Date string
	}{}

	err := k.GetPayload(&p)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	leave := c.GetdateLeave(k, p.Date)
	eleave := c.GetdataELeave(k, p.Date)
	remote := c.GetdateRemote(k, p.Date)

	data := tk.M{}
	data.Set("Leave", leave).Set("ELeave", eleave).Set("Remote", remote)

	return c.SetResultInfo(false, "success", data)
	//
}

func (c *StatusRequestController) GetdataELeave(k *knot.WebContext, date string) []AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]AprovalRequestLeaveModel, 0)

	dbFilter = append(dbFilter, db.Eq("dateleave", date))
	dbFilter = append(dbFilter, db.Eq("stsbymanager", "Approved"))
	dbFilter = append(dbFilter, db.Eq("isdelete", false))
	dbFilter = append(dbFilter, db.Eq("isemergency", true))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		tk.Println("-------- query ", query)
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	// eleaveData := c.IsSpecialsLeave(k, data)
	return data
}

func (c *StatusRequestController) GetdateLeave(k *knot.WebContext, date string) []AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]AprovalRequestLeaveModel, 0)

	dbFilter = append(dbFilter, db.Eq("dateleave", date))
	dbFilter = append(dbFilter, db.Eq("stsbymanager", "Approved"))
	dbFilter = append(dbFilter, db.Eq("isdelete", false))
	dbFilter = append(dbFilter, db.Eq("isemergency", false))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		tk.Println("-------- query ", query)
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	leaveData := c.IsSpecialsLeave(k, data)
	return leaveData
}

func (c *StatusRequestController) IsSpecialsLeave(k *knot.WebContext, leave []AprovalRequestLeaveModel) []AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	tk.Println("------------- len", len(leave))
	data := make([]*RequestLeaveModel, 0)

	crs, errdata := c.Ctx.Find(NewRequestLeave(), nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	dataLeave := []AprovalRequestLeaveModel{}

	for _, dt := range data {
		for _, l := range leave {
			if l.IdRequest == dt.Id {
				// if dt.IsSpecials != nil{
				if dt.IsSpecials == false {
					dataLeave = append(dataLeave, l)
				}
				// }

			}
		}
	}

	return dataLeave

}

func (c *StatusRequestController) GetdateRemote(k *knot.WebContext, date string) []*RemoteModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*RemoteModel, 0)

	dbFilter = append(dbFilter, db.Eq("dateleave", date))
	dbFilter = append(dbFilter, db.Eq("isdelete", false))
	dbFilter = append(dbFilter, db.Eq("projects.isapprovalmanager", true))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewRemoteModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	return data

}

func (c *StatusRequestController) GetLeaveDateById(k *knot.WebContext, Ids string) *AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	// data := AprovalRequestLeaveModel{}
	data := new(AprovalRequestLeaveModel)
	query := tk.M{}
	var dbFilter []*db.Filter

	// tk.Println("-------------- id ", Ids)
	dbFilter = append(dbFilter, db.Eq("_id", Ids))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 1, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data
}

func (c *StatusRequestController) GetRemoteDateById(k *knot.WebContext, Ids string) *RemoteModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	// data := AprovalRequestLeaveModel{}
	data := new(RemoteModel)
	query := tk.M{}
	var dbFilter []*db.Filter

	// tk.Println("-------------- id ", Ids)
	dbFilter = append(dbFilter, db.Eq("_id", Ids))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewRemoteModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 1, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data
}

func (c *StatusRequestController) CancelByDateId(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := struct {
		Ids    []string
		Reason string
		Date   string
		Type   string
	}{}

	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// to := []string{}

	by := ""

	if k.Session("jobrolelevel").(int) == 5 {
		by = "Admin"
	} else {
		by = "Manager"
	}

	mail := MailController(*c)

	data := make([]*RequestLeaveModel, 0)
	crs, errdata := c.Ctx.Find(NewRequestLeave(), nil)
	defer crs.Close()
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return c.SetResultInfo(true, errdata.Error(), nil)
	}

	dash := DashboardController(*c)
	t := time.Now()
	datenow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	for _, id := range p.Ids {

		if p.Type == "leave" || p.Type == "eleave" {
			dtl := c.GetLeaveDateById(k, id)

			usr := dash.GetDataSessionUser(k, dtl.UserId)

			dtl.ReasonAction = p.Reason
			dtl.IsDelete = true

			date, _ := time.Parse("2006-01-02", dtl.DateLeave)
			if date.Before(datenow) || dtl.DateLeave == datenow.Format("2006-01-02") {
				if dtl.StsByManager == "Approved" {
					if dtl.IsCutOff == true {
						usr[0].YearLeave = usr[0].YearLeave + 1
						usr[0].DecYear = usr[0].DecYear + 1.0
					}

				}
			}

			err = c.Ctx.Save(dtl)

			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

			err = c.Ctx.Save(usr[0])

			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

			tol := []string{dtl.Email}
			for _, dl := range data {
				if dl.Id == dtl.IdRequest {
					mail.SendMailUserCancelByDate(k, tol, p.Date, p.Reason, by, p.Type, dl.StatusManagerProject.Name, dl.StatusProjectLeader[0].Name, dl.Name)
				}
			}
		} else {
			dtr := c.GetRemoteDateById(k, id)

			dtr.ReasonAction = p.Reason
			dtr.IsDelete = true

			err = c.Ctx.Save(dtr)

			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

			dash := DashboardController(*c)
			usr := dash.GetDataSessionUser(k, dtr.UserId)

			tor := []string{usr[0].Email}

			mail.SendMailUserCancelByDate(k, tor, p.Date, p.Reason, by, p.Type, dtr.Projects[0].ProjectManager.Name, dtr.Projects[0].ProjectLeader.Name, dtr.Name)
		}

	}

	tk.Println("------------ to ", p)

	return c.SetResultInfo(false, "success", nil)
}
func (c *StatusRequestController) GetDataLeaveById(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		IdRequest string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	dataLeave := make([]*RequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	fmt.Println("----------- idrequest ", payload.IdRequest)

	dbFilter = append(dbFilter, db.Eq("_id", payload.IdRequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "success", dataLeave)
}
func (c *StatusRequestController) CancelRequestByDate(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Data  []AprovalRequestLeaveModel
		Name  string
		Email string
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	csr, e := c.Ctx.Connection.NewQuery().Select().From(NewRequestLeave().TableName()).Where(db.Eq("_id", p.Data[0].IdRequest)).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	results := RequestLeaveModel{}
	e = csr.Fetch(&results, 1, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	m := 0

	for _, pr := range results.StatusProjectLeader {
		if pr.StatusRequest == "Pending" {
			m++
		}

	}
	level := k.Session("jobrolelevel").(int)
	dash := DashboardController(*c)
	if len(results.Project) == m && results.StatusManagerProject.StatusRequest == "Pending" || level == 5 || level == 1 {
		// dateCancel := []string{}
		for _, each := range p.Data {
			// dateCancel = append(dateCancel, each.DateLeave)
			e = c.Ctx.Save(&each)
			if e != nil {
				return c.SetResultInfo(true, e.Error(), nil)
			}
		}
		// change parent
		list := results.LeaveDateList
		for _, ne := range p.Data {
			for i, ls := range list {
				if ne.DateLeave == ls {
					list = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
		results.LeaveDateList = list
		results.LeaveFrom = list[0]
		results.LeaveTo = list[len(list)-1]
		results.NoOfDays = len(list)
		e = c.Ctx.Save(&results)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}

		dash.SetHistoryLeave(k, results.UserId, results.Id, results.LeaveFrom, results.LeaveTo, "some date cancelled", results.ResultRequest, &results)
		emptyremote := new(RemoteModel)
		service := services.LogService{
			&results,
			emptyremote,
			"leave",
		}
		log := tk.M{}
		log.Set("Status", "Canceled")
		log.Set("Desc", "Request "+results.ResultRequest)
		log.Set("NameLogBy", results.Name)
		log.Set("EmailNameLogBy", results.Email)
		e = service.CancelRequest(log)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}

		if level != 5 || level != 1 {
			mail := MailController(*c)
			mail.SendMailLeaderCancel(k, &results)
		}

	}
	// level := k.Session("jobrolelevel").(int)
	//log
	emptyremote := new(RemoteModel)
	service := services.LogService{
		&results,
		emptyremote,
		"leave",
	}
	log := tk.M{}
	log.Set("Status", "CancelByDate")
	log.Set("Desc", "Request Cancel By Date")
	log.Set("NameLogBy", p.Name)
	log.Set("EmailNameLogBy", p.Email)
	e = service.CancelRequest(log)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "Success", nil)
}

func (c *StatusRequestController) CheckLeaveDetails(k *knot.WebContext, m *RequestLeaveModel) (bool, error) {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	pipe := []tk.M{}

	pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
	pipe = append(pipe, tk.M{"$match": tk.M{"_id": m.Id}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeave").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return false, err
	}

	if err != nil {
		return false, err
	}

	// tk.Println("-------------------- masuk")

	data := make([]*RequestLeaveModel, 0)
	if err = csr.Fetch(&data, 0, false); err != nil {
		return false, err
	}

	level := k.Session("jobrolelevel").(int)

	if level == 1 || level == 6 {
		if data[0].StatusManagerProject.StatusRequest != "Pending" {
			return true, nil
		}
	} else if level == 2 {
		for _, d := range data[0].StatusProjectLeader {
			// tk.Println("-------------------- masuk", d)
			if d.UserId == k.Session("userid").(string) && data[0].StatusManagerProject.StatusRequest != "Pending" {
				return true, nil
			}
		}
	} else {
		for _, d := range data[0].StatusProjectLeader {
			// tk.Println("-------------------- masuk", d)
			if d.StatusRequest != "Pending" {
				return true, nil
			}
		}
	}

	return false, nil
}

func (c *StatusRequestController) LeaveByDateId(k *knot.WebContext, idRequest string) []*AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	data := make([]*AprovalRequestLeaveModel, 0)
	var dbFilter []*db.Filter
	// dbFilter = append(dbFilter, db.Eq("dateleave", dateLeave))
	dbFilter = append(dbFilter, db.Eq("idrequest", idRequest))
	// dbFilter = append(dbFilter, db.Eq("stsbyleader", "Pending"))

	query := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	defer crs.Close()
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}
	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		//return *models.AprovalRequestLeaveModel{}
	}

	return data
}

func (c *StatusRequestController) EditRequestLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := new(RequestLeaveModel)
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	dash := DashboardController(*c)

	det, err := c.CheckLeaveDetails(k, p)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if det == true {
		return c.SetResultInfo(true, "Leader already give decision to your request", nil)
	}

	level := k.Session("jobrolelevel").(int)

	res := c.CheckLeaveEdit(k, p, level)

	// if err != nil {
	// 	return err
	// }

	if res != nil {
		return res
	}

	pend := c.CheckLeavePendingEdit(k, p, level)

	tk.Println("------------------- pend ", pend)

	// if err != nil {
	// 	return err
	// }

	if pend != nil {
		return pend
	}

	rem := dash.CheckRemote(k, p)

	tk.Println("------------------- rem ", rem)

	if rem != nil {
		return rem
	}

	err = c.Ctx.Save(p)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	bydate := c.LeaveByDateId(k, p.Id)
	tk.Println("------- ", bydate)
	if len(bydate) > 0 {
		for _, indate := range bydate {
			if level == 4 || level == 5 || level == 3 {
				tk.Println("------------- masuk not leader", level)
				if indate.StsByLeader != "Pending" {
					return c.SetResultInfo(true, "Request already has approvement by Leader", nil)
				}
			} else {
				if indate.StsByManager != "Pending" {
					return c.SetResultInfo(true, "Request already has approvement by manager", nil)
				}
			}

			err = c.Ctx.Delete(indate)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

		}
	} else {
		return c.SetResultInfo(true, "Leader already give decision to your request", nil)
	}

	dash.SetHistoryLeave(k, p.UserId, p.Id, p.LeaveFrom, p.LeaveTo, "Send To Leader", "Pending", p)

	emptyremote := new(RemoteModel)
	service := services.LogService{
		p,
		emptyremote,
		"leave",
	}
	err = service.RequestLog()
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	mailContro := MailController(*c)
	mailContro.RequestLeaveOnDateV2(p, p.UserId)
	if level == 1 || level == 6 {
		mailContro.SendMailManagerEdited(k, p.UserId, p.Id, level)
	} else if level == 2 {
		for _, lead := range p.StatusProjectLeader {
			if p.UserId == lead.UserId {
				mailContro.SendMailManagerEdited(k, p.UserId, p.Id, level)
			} else {
				mailContro.SendMailLeaderEdited(k, p, level)
			}
		}
	}

	return c.SetResultInfo(false, "data success saved", nil)
}

func (c *StatusRequestController) CheckLeaveEdit(k *knot.WebContext, m *RequestLeaveModel, level int) interface{} {
	k.Config.OutputType = knot.OutputJson
	// lvDate := new(AprovalRequestLeaveModel)

	pipe := []tk.M{}

	if level == 2 || level == 6 || level == 1 {
		pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbyleader": "Approved"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Pending"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})
	} else {
		pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
		// pipe = append(pipe, tk.M{"$match": tk.M{"stsbyleader": "Pending"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Approved"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})
	}

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeaveByDate").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return c.SetResultInfo(true, "invalid query1", nil)
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("-------------------- masuk")

	data := []*AprovalRequestLeaveModel{}
	if err = csr.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	for _, d := range m.LeaveDateList {
		// tk.Println("-------------------- masuk", d)
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}

		// pend := c.CheckRequestPending(k, dt)

		// if pend != nil {
		// 	return pend
		// }
		for _, dm := range data {
			if dm.DateLeave == string(dt.Format("2006-01-02")) {
				if dm.IdRequest != m.Id {
					return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken and status is "+dm.StsByManager, nil)
				}
			}

		}
	}
	return nil
}

func (c *StatusRequestController) CheckLeavePendingEdit(k *knot.WebContext, m *RequestLeaveModel, level int) interface{} {
	k.Config.OutputType = knot.OutputJson
	// lvDate := new(AprovalRequestLeaveModel)

	pipe := []tk.M{}

	if level == 2 || level == 6 || level == 1 {
		pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbyleader": "Approved"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Pending"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})
	} else {
		pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbyleader": "Pending"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Pending"}})
		pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})
	}

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeaveByDate").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return c.SetResultInfo(true, "invalid query1", nil)
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("-------------------- masuk")

	data := []*AprovalRequestLeaveModel{}
	if err = csr.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	for _, d := range m.LeaveDateList {
		// tk.Println("-------------------- masuk", d)
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}

		// pend := c.CheckRequestPending(k, dt)

		// if pend != nil {
		// 	return pend
		// }
		for _, dm := range data {
			if dm.DateLeave == string(dt.Format("2006-01-02")) {
				if dm.IdRequest != m.Id {
					return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken and status is "+dm.StsByManager, nil)
				}
			}

		}
	}

	return nil
}

func (c *StatusRequestController) UploadELAttachment(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	reader, err := k.Request.MultipartReader()

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	var fileLocation string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		fileLocation = filepath.Join("assets/doc/", part.FileName())
		dst, err := os.Create(fileLocation)
		if dst != nil {
			defer dst.Close()
		}
		// tk.Pri
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		if _, err := io.Copy(dst, part); err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
	}

	return c.SetResultInfo(false, "success", nil)
}

// var (
// 	wd = func() string {
// 		d, _ := os.Getwd()
// 		return d + "/"
// 	}()
// )

func (c *StatusRequestController) SaveDataUpload(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)
	payload := struct {
		Id       string
		FileName string
	}{}

	err := k.GetPayload(&payload)

	if err != nil {
		tk.Println("--------- payload ", payload)
		return c.SetResultInfo(true, err.Error(), nil)
	}

	mail := MailController(*c)
	dash := DashboardController(*c)
	dataleave := mail.getDataLeave(payload.Id)
	tk.Println("--------- dataleave ", dataleave)
	urlConf := helper.ReadConfig()
	sUrl := urlConf.GetString("BaseUrl")
	dataleave.FileLocation = sUrl + "/static/doc/" + payload.FileName
	dataleave.IsAttach = true

	err = c.Ctx.Save(dataleave)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	userid := k.Session("userid").(string)
	level := k.Session("jobrolelevel").(int)
	hist := dash.GetHistoryLeave(k, userid, level)
	for i, _ := range hist[0].Leavehistory {
		if hist[0].Leavehistory[i].IdRequest == payload.Id {
			hist[0].Leavehistory[i].FileAttachment = sUrl + "/static/doc/" + payload.FileName
		}
	}
	err = c.Ctx.Save(hist[0])
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "Success", nil)
}

func (c *StatusRequestController) Datedetails(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)

	p := struct {
		IdRequest string
		Tipe      string
	}{}

	err := k.GetPayload(&p)

	if err != nil {
		tk.Println("--------- payload 123 tipe ", err.Error())
		return c.SetResultInfo(true, err.Error(), nil)
	}
	tk.Println("--------- payload 123 tipe ", p.Tipe)
	switch p.Tipe {
	case "LEAVE":
		dl := c.GetdateLeaveId(k, p.IdRequest)
		return dl
	case "EMERGENCY LEAVE":
		de := c.GetdateLeaveId(k, p.IdRequest)
		return de
	case "REMOTE":
		dr := c.GetRemoteId(k, p.IdRequest)
		return dr
	case "OVERTIME":
		ov := OvertimeController(*c)
		ovd, err := ov.GetOvertimeByID(k, p.IdRequest)

		if err != nil {
			// tk.Println("--------- payload ", payload)
			return c.SetResultInfo(true, err.Error(), nil)
		}
		return ovd
	}

	return "success"
}

func (c *StatusRequestController) GetdateLeaveId(k *knot.WebContext, idreequest string) []AprovalRequestLeaveModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]AprovalRequestLeaveModel, 0)

	dbFilter = append(dbFilter, db.Eq("idrequest", idreequest))
	// dbFilter = append(dbFilter, db.Eq("isdelete", false))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		tk.Println("-------- query ", query)
	}

	crs, errdata := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	leaveData := c.IsSpecialsLeave(k, data)
	return leaveData
}

func (c *StatusRequestController) GetRemoteId(k *knot.WebContext, id string) []*RemoteModel {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*RemoteModel, 0)

	dbFilter = append(dbFilter, db.Eq("idop", id))
	// dbFilter = append(dbFilter, db.Eq("isdelete", false))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := c.Ctx.Find(NewRemoteModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	return data

}

func (c *StatusRequestController) GetLeavebyID(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Id string
	}{}
	err := k.GetPayload(&p)
	if err != nil {
		tk.Println(err)
	}
	tk.Println(p)
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*RequestLeaveModel, 0)
	if p.Id != "nil" {

		dbFilter = append(dbFilter, db.Eq("_id", p.Id))
		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewRequestLeave(), query)
		if crs != nil {
			defer crs.Close()
		} else {
			return nil
		}
		// defer crs.Close()
		if errdata != nil {
			return nil
		}

		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return nil
		}

	} else {
		return nil
	}

	return data
}
