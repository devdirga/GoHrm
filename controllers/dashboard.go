package controllers

import (
	"creativelab/ecleave-dev/helper"
	"creativelab/ecleave-dev/services"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	// "fmt"

	"time"

	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type DashboardController struct {
	*BaseController
}

func (c *DashboardController) Default(k *knot.WebContext) interface{} {

	c.LoadBase(k)
	viewData := tk.M{}

	userset := UserSettingController(*c)
	dataUser, err := userset.GetOptionUser(k, k.Session("userid").(string))
	if err != nil {
	}

	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("UserId", k.Session("userid").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel"))
		if len(dataUser) > 0 {
			viewData.Set("RemoteActive", dataUser[0].Remote.RemoteActive)
			viewData.Set("ConditionalRemote", dataUser[0].Remote.ConditionalRemote)
			viewData.Set("FullMonth", dataUser[0].Remote.FullMonth)
			viewData.Set("Monthly", dataUser[0].Remote.Monthly)
		}

	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
		viewData.Set("UserId", "")
	}

	DataAccess := c.SetViewData(k, viewData)

	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = []string{
		"_modal.html",
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

func (c *DashboardController) GetCurrentDate(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	dateNow := time.Now().Format("2006-01-02")
	timeNow := time.Now().Format("15:04:05")
	ret.Data = tk.M{}.Set("CurrentDate", dateNow).Set("CurrentTime", timeNow)
	return ret
}

func (c *DashboardController) GetDataUser(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)

	userid := k.Session("userid")
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserProfileModel, 0)
	if userid != "nil" {

		dbFilter = append(dbFilter, db.Eq("_id", userid))

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewSysUserModel(), query)
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

func (c *DashboardController) GetDataSessionUser(k *knot.WebContext, id string) []*SysUserModel {
	k.Config.OutputType = knot.OutputJson

	userid := id
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
func (c *DashboardController) GetLeavebyDateFilterIdTransc(k *knot.WebContext, id string) []*AprovalRequestLeaveModel {
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*AprovalRequestLeaveModel, 0)
	if id != "nil" {

		dbFilter = append(dbFilter, db.Eq("idrequest", id))

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
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

	} else {
		return nil
	}

	return data
}
func (c *DashboardController) GetDetailLeaveByIdTransc(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := new(RequestLeaveModel)
	err := k.GetPayload(&p)
	if err != nil {
		tk.Println(err)
	}
	tk.Println(p)
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*AprovalRequestLeaveModel, 0)
	if p.Id != "nil" {

		dbFilter = append(dbFilter, db.Eq("idrequest", p.Id))
		dbFilter = append(dbFilter, db.Eq("isdelete", false))
		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
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

	} else {
		return nil
	}

	return data
}

func (c *DashboardController) GetAllUser(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	pipe := []tk.M{}

	pipe = append(pipe, tk.M{"$match": tk.M{"designation": tk.M{"$ne": "Junior"}}})
	pipe = append(pipe, tk.M{"$project": tk.M{
		"_id":           1,
		"empid":         1,
		"designation":   1,
		"departement":   1,
		"fullname":      1,
		"phonenumber":   1,
		"email":         1,
		"yearleave":     1,
		"publicleave":   1,
		"location":      1,
		"photo":         1,
		"projectruleid": 1,
	}})
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("SysUsers").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else if csr == nil {
		return errors.New("error when build query")
	}

	if err != nil {
		return nil
	}
	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil
	}

	return data
}

func (c *DashboardController) GetTimeZone(k *knot.WebContext, location string) (string, error) {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	data := LocationModel{}
	var filter []*db.Filter
	query := tk.M{}
	filter = append(filter, db.Eq("Location", location))
	if len(filter) > 0 {
		query.Set("where", db.And(filter...))
	}
	crs, err := c.Ctx.Find(NewLocationModel(), query)

	if err != nil {
		return "", err
	}

	err = crs.Fetch(&data, 1, false)
	if err != nil {
		return "", err
	}

	return data.TimeZone, nil
}

func (c *DashboardController) RequestLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := new(RequestLeaveModel)
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	res := c.CheckLeave(k, p)

	if err != nil {
		return err
	}

	if res != nil {
		return res
	}

	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
	}

	userid := k.Session("userid")
	p.UserId = tk.ToString(userid)
	p.ResultRequest = "Pending"
	p.IsEmergency = false
	p.IsReset = false
	p.IsAttach = false

	tz, err := c.GetTimeZone(k, p.Location)
	tl, err := helper.TimeLocation(tz)

	rm, err := helper.ExpiredRemaining(tz, false)
	p.ExpRemaining = rm
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	p.DateCreateLeave = tl
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	now, err := helper.ExpiredDateTime(tz, false)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	p.ExpiredOn = now

	rem := c.CheckRemote(k, p)

	if rem != nil {
		return rem
	}

	err = c.Ctx.Save(p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// dataUser := c.GetDataSessionUser(k, tk.ToString(userid))

	// if len(dataUser) > 0 && dataUser != nil {
	// 	// fmt.Println("------------- masuk")
	// 	dataUser[0].YearLeave = p.YearLeave

	// 	err = c.Ctx.Save(dataUser[0])
	// 	if err != nil {
	// 		return c.SetResultInfo(true, err.Error(), nil)
	// 	}
	// }

	mailContro := MailController(*c)
	level := k.Session("jobrolelevel").(int)

	mailContro.RequestLeaveOnDateV2(p, tk.ToString(userid))

	if level == 1 || level == 6 {
		tk.Println("------------------ masuk sini1")
		mailContro.SendMailManager(k, tk.ToString(userid), p.Id, level)
	} else {
		tk.Println("------------------ masuk sini2")
		mailContro.SendMailLeader(k, p, level)
	}

	fmt.Println(tk.ToString(userid))
	fmt.Println(p.Id)
	fmt.Println(p.LeaveFrom)
	fmt.Println(p.LeaveTo)
	fmt.Println("Create request Leave")
	fmt.Println("New Request")

	c.SetHistoryLeave(k, tk.ToString(userid), p.Id, p.LeaveFrom, p.LeaveTo, "Send To Leader", "Pending", p)
	// log
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
	notif := NotificationController(*c)
	dataNotif := NotificationModel{}
	dataNotif.Id = ""
	dataNotif.UserId = p.UserId
	dataNotif.IsConfirmed = false
	dataNotif.Notif.Name = p.Name
	dataNotif.IdRequest = p.Id
	dataNotif.Notif.DateFrom = p.LeaveFrom
	dataNotif.Notif.DateTo = p.LeaveTo
	dataNotif.Notif.Description = "Create request Leave"
	dataNotif.Notif.Status = p.ResultRequest
	dataNotif.Notif.CreatedAt = tl
	dataNotif.Notif.UpdatedAt = tl
	dataNotif.Notif.RequestType = "Leave"
	dataNotif.Notif.Reason = p.Reason
	dataNotif.Notif.ManagerApprove = p.StatusManagerProject.Name
	// dataNotif.Notif.IdRequest = p.Id
	dataNotif.Notif.StatusApproval = p.StatusManagerProject.StatusRequest

	notif.InsertNotification(dataNotif)
	return c.SetResultInfo(false, "Request succesfully sent", nil)

}

// func (c *DashboardController) CancelRemote(k *knot.WebContext, m *RequestLeaveModel) interface{} {
// 	k.Config.OutputType = knot.OutputJson

// 	datas, _ := c.GetRemote(k, m)
// 	datas.IsDelete = true
// 	err := c.Ctx.Save(datas)

// 	if err != nil {
// 		return c.SetResultInfo(true, err.Error(), nil)
// 	}

// 	mail := MailController()

// 	return c.SetResultInfo(false, "success", nil)
// }

// func (c *DashboardController) GetRemote(k *knot.WebContext, m *RequestLeaveModel) (*RemoteModel, error) {
// 	k.Config.OutputType = knot.OutputJson
// 	pipe := []tk.M{}

// 	pipe = append(pipe, tk.M{"$match": tk.M{"userid": m.UserId}})
// 	datas := []*RemoteModel{}
// 	crs, err := c.Ctx.Connection.NewQuery().From("remote").Command("pipe", pipe).Cursor(nil)
// 	if crs != nil {
// 		defer crs.Close()
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	err = crs.Fetch(&datas, 0, false)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// tk.Println("-------- datas", m.UserId)

// 	for _, dt := range datas {
// 		for _, l := range m.LeaveDateList {
// 			if dt.DateLeave == l {
// 				return dt, nil
// 			}
// 		}
// 	}

// 	return nil, nil
// }

func (c *DashboardController) CheckRemote(k *knot.WebContext, m *RequestLeaveModel) interface{} {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	pipe = append(pipe, tk.M{"$match": tk.M{"userid": m.UserId}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})

	datas := []RemoteModel{}
	crs, err := c.Ctx.Connection.NewQuery().From("remote").Command("pipe", pipe).Cursor(nil)
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

	// tk.Println("-------- datas", m.UserId)

	for _, dt := range datas {
		for _, l := range m.LeaveDateList {
			if dt.DateLeave == l {
				return c.SetResultInfo(true, "you already get remote in "+l+" please cancel remote date first, for info call your admin", dt)
			}
		}
	}

	return nil
}

func (c *DashboardController) CheckLeave(k *knot.WebContext, m *RequestLeaveModel) interface{} {
	k.Config.OutputType = knot.OutputJson
	// lvDate := new(AprovalRequestLeaveModel)

	pipe := []tk.M{}
	// pipeTemp := []tk.M{}
	// start, _ := time.Parse("2006-1-2", m.LeaveFrom)
	// end, _ := time.Parse("2006-1-2", m.LeaveTo)
	// // fmt.Println(end)

	// for d := rangeDate(start, end); ; {
	// 	dt := d()
	// 	if dt.IsZero() {
	// 		break
	// 	}
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
	pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Approved"}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})

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
	pending := c.CheckPendingDate(k, m)

	if pending != nil {
		return pending
	}

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

		pend := c.CheckRequestPending(k, dt)

		if pend != nil {
			return pend
		}
		for _, dm := range data {
			if dm.DateLeave == string(dt.Format("2006-01-02")) {
				if dm.IsDelete == false && dm.StsByManager == "Approved" {
					return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken on your leave", nil)
				}
			}

		}
	}
	return nil
}

func (c *DashboardController) CheckLeaveByUser(k *knot.WebContext, m *RequestLeaveModel, empid string) interface{} {
	k.Config.OutputType = knot.OutputJson
	// lvDate := new(AprovalRequestLeaveModel)

	pipe := []tk.M{}
	// pipeTemp := []tk.M{}
	// start, _ := time.Parse("2006-1-2", m.LeaveFrom)
	// end, _ := time.Parse("2006-1-2", m.LeaveTo)
	// // fmt.Println(end)

	// for d := rangeDate(start, end); ; {
	// 	dt := d()
	// 	if dt.IsZero() {
	// 		break
	// 	}
	pipe = append(pipe, tk.M{"$match": tk.M{"empid": empid}})
	pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Approved"}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})

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
	pending := c.CheckPendingDate(k, m)

	if pending != nil {
		return pending
	}

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

		pend := c.CheckRequestPending(k, dt)

		if pend != nil {
			return pend
		}
		for _, dm := range data {
			if dm.DateLeave == string(dt.Format("2006-01-02")) {
				if dm.IsDelete == false && dm.StsByManager == "Approved" {
					return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken on your leave", nil)
				}
			}

		}
	}
	return nil
}

func (c *DashboardController) CheckPendingDate(k *knot.WebContext, m *RequestLeaveModel) interface{} {
	k.Config.OutputType = knot.OutputJson
	// lvDate := new(AprovalRequestLeaveModel)

	pipe := []tk.M{}
	// pipeTemp := []tk.M{}
	// start, _ := time.Parse("2006-1-2", m.LeaveFrom)
	// end, _ := time.Parse("2006-1-2", m.LeaveTo)
	// // fmt.Println(end)

	// for d := rangeDate(start, end); ; {
	// 	dt := d()
	// 	if dt.IsZero() {
	// 		break
	// 	}
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": k.Session("userid")}})
	pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": "Pending"}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})

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
		pend := c.CheckRequestPending(k, dt)

		if pend != nil {
			return pend
		}
		for _, dm := range data {
			if dm.DateLeave == string(dt.Format("2006-01-02")) {
				if dm.IsDelete == false && dm.StsByManager == "Pending" {
					return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken on your leave and status is 'Pending'", nil)
				}
			}

		}
	}
	return nil
}

func (c *DashboardController) CheckRequestPending(k *knot.WebContext, dt time.Time) interface{} {
	requestPending, err := c.PendingRequest(k, tk.ToString(k.Session("userid")))
	if err != nil {
		return nil
	}
	if len(requestPending) == 0 {
		return nil
	}
	for _, pend := range requestPending {
		// tk.Println("--------------------", pend.LeaveFrom)
		onstart, _ := time.Parse("2006-1-2", pend.LeaveFrom)
		onend, _ := time.Parse("2006-1-2", pend.LeaveTo)

		for pendL := rangeDate(onstart, onend); ; {
			pendt := pendL()

			if pendt.IsZero() {

				break
			}

			// tk.Println("--------------------", string(dt.Format("2006-01-02")) == string(pendt.Format("2006-01-02")))

			if string(dt.Format("2006-01-02")) == string(pendt.Format("2006-01-02")) {
				return c.SetResultInfo(true, "date of "+string(dt.Format("2-1-2006"))+" already taken on your leave", nil)
			}
		}
	}
	return nil
}

func (c *DashboardController) RequestEmergency(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	p := new(RequestLeaveModel)
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	tz, err := c.GetTimeZone(k, p.Location)
	tl, err := helper.TimeLocation(tz)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	// userid := k.Session("userid")

	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
	}

	// tk.Println("--------------------  to ", p.Leave)
	i := 0
	for _, d := range p.LeaveDateList {
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}

		onday := dt.Weekday()

		if onday.String() == "Sunday" || onday.String() == "Saturday" {

		} else {
			i = i + 1

			// p.LeaveTo = d
		}
	}
	// tk.Println("-------------------- ", i)
	p.UserId = tk.ToString(k.Session("userid"))
	p.ResultRequest = "Pending"
	p.IsEmergency = true
	p.DateCreateLeave = tl
	p.NoOfDays = i
	p.IsReset = false
	p.IsAttach = false
	// p.IsCutOff = false
	// p.YearLeave = p.YearLeave - i
	// exp := time.Now().Add(6 * time.Hour)
	// p.ExpiredOn = exp.Format("2006-01-02 15:04")

	rem := c.CheckRemote(k, p)

	if rem != nil {
		return rem
	}

	res := c.CheckLeave(k, p)

	if res != nil {
		return res
	}

	err = c.Ctx.Save(p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	level := k.Session("jobrolelevel").(int)

	mailContro := MailController(*c)
	mailContro.EmergencyLeave(k, p, level)

	c.SetHistoryLeave(k, p.UserId, p.Id, p.LeaveFrom, p.LeaveTo, "Request Emergency Leave", "Pending", p)

	//log
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

	notif := NotificationController(*c)
	dataNotif := NotificationModel{}
	dataNotif.Id = ""
	dataNotif.UserId = p.UserId
	dataNotif.IsConfirmed = false
	dataNotif.Notif.Name = p.Name
	dataNotif.IdRequest = p.Id
	dataNotif.Notif.DateFrom = p.LeaveFrom
	dataNotif.Notif.DateTo = p.LeaveTo
	dataNotif.Notif.Description = "Create request ELeave"
	dataNotif.Notif.Status = p.ResultRequest
	dataNotif.Notif.CreatedAt = tl
	dataNotif.Notif.UpdatedAt = tl
	dataNotif.Notif.RequestType = "ELeave"
	dataNotif.Notif.Reason = p.Reason
	dataNotif.Notif.ManagerApprove = p.StatusManagerProject.Name
	// dataNotif.Notif.IdRequest = p.Id
	dataNotif.Notif.StatusApproval = p.StatusManagerProject.StatusRequest

	notif.InsertNotification(dataNotif)

	return c.SetResultInfo(false, "Request succesfully sent", nil)

}

func (c *DashboardController) SetHistoryLeave(k *knot.WebContext, userid string, idreq string, dateFrom string, dateTo string, description string, status string, leave *RequestLeaveModel) bool {
	c.LoadBase(k)
	notif := new(services.HistoryService)
	notif.Name = leave.Name
	notif.DateFrom = dateFrom
	notif.DateTo = dateTo
	notif.IDRequest = idreq
	notif.UserId = leave.UserId
	notif.RequestType = "leave"
	notif.Desc = description
	notif.StatusApproval = ""
	notif.Status = status
	notif.Reason = leave.Reason
	notif.ManagerApprove = leave.StatusManagerProject.Name

	err := notif.Push(leave.IsEmergency)
	if err != nil {
		return false
	}
	// k.Config.OutputType = knot.OutputJson
	// history := c.GetHistoryLeave(k, userid)
	// // fmt.Println("------------------ history", history)
	// if len(history) == 0 {
	// 	p := new(HistoryLeaveModel)
	// 	// userid := k.Session("userid")

	// 	if p.Id == "" {
	// 		p.Id = bson.NewObjectId().Hex()
	// 	}

	// 	// fmt.Println("------ userid", userid)

	// 	p.UserId = userid

	// 	hist := HistoryDetails{}
	// 	hist.IdRequest = idreq
	// 	hist.DateFrom = dateFrom
	// 	hist.DateTo = dateTo
	// 	hist.Description = description
	// 	hist.Status = status
	// 	hist.ManagerApprove = leave.StatusManagerProject.Name
	// 	hist.Reason = leave.Reason
	// 	hist.IsEmergency = leave.IsEmergency
	// 	t := time.Now()
	// 	hist.HistoryDate = t.Format("2006-01-02 15:04:05")
	// 	p.Leavehistory = append(p.Leavehistory, hist)

	// 	err := c.Ctx.Save(p)
	// 	if err != nil {
	// 		return false
	// 	}
	// } else {
	// 	p := new(HistoryLeaveModel)

	// 	p.Id = history[0].Id
	// 	p.UserId = history[0].UserId

	// 	hist := HistoryDetails{}
	// 	hist.IdRequest = idreq
	// 	hist.DateFrom = dateFrom
	// 	hist.DateTo = dateTo
	// 	hist.Description = description
	// 	hist.Reason = leave.Reason
	// 	hist.ManagerApprove = leave.StatusManagerProject.Name
	// 	hist.IsEmergency = leave.IsEmergency
	// 	hist.Status = status
	// 	t := time.Now()
	// 	hist.HistoryDate = t.Format("2006-01-02 15:04:05")

	// 	p.Leavehistory = history[0].Leavehistory

	// 	p.Leavehistory = append(p.Leavehistory, hist)

	// 	err := c.Ctx.Save(p)
	// 	if err != nil {
	// 		return false
	// 	}
	// }

	return true
}

func (c *DashboardController) GetHistoryLeave(k *knot.WebContext, userid string, level int) []*HistoryLeaveModel {
	k.Config.OutputType = knot.OutputJson
	// userid := k.Session("userid")
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*HistoryLeaveModel, 0)
	// fmt.Println("---------------- userid", userid)
	dbFilter = append(dbFilter, db.Eq("userid", userid))

	if len(dbFilter) > 0 {
		if level == 2 || level == 3 {
			query.Set("where", db.And(dbFilter...))
		}

	}

	crs, errdata := c.Ctx.Find(NewHistoryLeaveModel(), query)
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

	// fmt.Println("---------------- data", tk.JsonString(data))

	return data
}

func (c *DashboardController) FetchRequestLeave(k *knot.WebContext, userid string, status string) ([]tk.M, error) {
	// k.Config.OutputType = knot.OutputJson
	pipeld := []tk.M{}
	pipemg := []tk.M{}
	pipeEm := []tk.M{}
	res := []tk.M{}

	pipeld = append(pipeld, tk.M{"$unwind": "$statusprojectleader"})
	pipeld = append(pipeld, tk.M{"$match": tk.M{"statusprojectleader.userid": userid, "statusprojectleader.statusrequest": status, "isemergency": tk.M{"$eq": false}, "resultrequest": tk.M{"$eq": "Pending"}}}) // "resultrequest":             status
	pipeld = append(pipeld, tk.M{"$project": tk.M{
		"userid":          1,
		"name":            1,
		"empid":           1,
		"designation":     1,
		"location":        1,
		"departement":     1,
		"reason":          1,
		"email":           1,
		"address":         1,
		"contact":         1,
		"leavefrom":       1,
		"leaveto":         1,
		"noofdays":        1,
		"project":         1,
		"yearleave":       1,
		"publicleave":     1,
		"isemergency":     1,
		"datecreateleave": 1,
		"statusprojectleader": tk.M{
			"idemp":         1,
			"name":          1,
			"location":      1,
			"email":         1,
			"phonenumber":   1,
			"statusrequest": 1,
			"reason":        1,
			"projectname":   1,
			"userid":        1,
		},
	}}) // "resultrequest":

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeld).
		From("requestLeave").
		Cursor(nil)
	if csr != nil {
		defer csr.Close()
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// defer csr.Close()
	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, err
	}

	for _, ld := range data {
		res = append(res, ld)
	}

	level := k.Session("jobrolelevel").(int)

	if level == 1 || level == 5 || level == 6 {
		// tk.Println("----------- masuk")
		pipemg = append(pipemg, tk.M{"$unwind": "$statusmanagerproject"})
		pipemg = append(pipemg, tk.M{"$match": tk.M{"statusmanagerproject.statusrequest": tk.M{"$eq": "Pending"}, "statusprojectleader.statusrequest": tk.M{"$ne": "Pending"}, "isemergency": tk.M{"$eq": false}, "resultrequest": tk.M{"$eq": "Pending"}}}) // "resultrequest":             status,

		pipemg = append(pipemg, tk.M{"$project": tk.M{
			"userid":          1,
			"name":            1,
			"empid":           1,
			"designation":     1,
			"location":        1,
			"departement":     1,
			"reason":          1,
			"email":           1,
			"address":         1,
			"contact":         1,
			"leavefrom":       1,
			"leaveto":         1,
			"noofdays":        1,
			"project":         1,
			"yearleave":       1,
			"publicleave":     1,
			"isemergency":     1,
			"datecreateleave": 1,
			"statusmanagerproject": tk.M{
				"idemp":         1,
				"name":          1,
				"location":      1,
				"email":         1,
				"phonenumber":   1,
				"statusrequest": 1,
				"reason":        1,
				"projectname":   1,
				"userid":        1,
			},
			"statusprojectleader": 1,
		}}) // "resultrequest":

		csrm, err := c.Ctx.Connection.
			NewQuery().
			Command("pipe", pipemg).
			From("requestLeave").
			Cursor(nil)

		if csrm != nil {
			defer csrm.Close()
		} else {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		// defer csrm.Close()

		datamg := []tk.M{}
		if err := csrm.Fetch(&datamg, 0, false); err != nil {
			return nil, err
		}

		// fmt.Println("-------------- datamg", datamg)

		for _, mg := range datamg {
			res = append(res, mg)
		}

		// pipeEm = append(pipeEm, tk.M{"$unwind": "$statusprojectleader"})
		pipeEm = append(pipeEm, tk.M{"$unwind": "$statusmanagerproject"})
		pipeEm = append(pipeEm, tk.M{"$match": tk.M{"statusmanagerproject.statusrequest": "Pending", "isemergency": tk.M{"$eq": true}, "resultrequest": tk.M{"$eq": "Pending"}}}) // "resultrequest":             status,
		pipeEm = append(pipeEm, tk.M{"$project": tk.M{
			"userid":          1,
			"name":            1,
			"empid":           1,
			"designation":     1,
			"location":        1,
			"departement":     1,
			"reason":          1,
			"email":           1,
			"address":         1,
			"contact":         1,
			"leavefrom":       1,
			"leaveto":         1,
			"noofdays":        1,
			"project":         1,
			"yearleave":       1,
			"publicleave":     1,
			"isemergency":     1,
			"datecreateleave": 1,
			"statusmanagerproject": tk.M{
				"idemp":         1,
				"name":          1,
				"location":      1,
				"email":         1,
				"phonenumber":   1,
				"statusrequest": 1,
				"reason":        1,
				"projectname":   1,
				"userid":        1,
			},
		}}) // "resultrequest":
		// pipeEm = append(pipeEm, tk.M{"$project": tk.M{}}) // "resultrequest":             status,

		csrEm, err := c.Ctx.Connection.
			NewQuery().
			Command("pipe", pipeEm).
			From("requestLeave").
			Cursor(nil)

		if csrEm != nil {
			defer csrEm.Close()
		} else {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		// defer csrEm.Close()

		dataEm := []tk.M{}
		if err := csrEm.Fetch(&dataEm, 0, false); err != nil {
			return nil, err
		}

		for _, Em := range dataEm {
			res = append(res, Em)
		}
	}

	return res, nil

}

func (c *DashboardController) FetchRequestCount(k *knot.WebContext, userid string, status string) ([]tk.M, error) {
	// k.Config.OutputType = knot.OutputJson
	pipeld := []tk.M{}
	pipemg := []tk.M{}
	res := []tk.M{}

	pipeld = append(pipeld, tk.M{"$unwind": "$statusprojectleader"})
	pipeld = append(pipeld, tk.M{"$match": tk.M{"statusprojectleader.userid": userid, "statusprojectleader.statusrequest": status, "isemergency": tk.M{"$eq": false}, "resultrequest": tk.M{"$eq": "Pending"}}}) // "resultrequest":             status
	pipeld = append(pipeld, tk.M{"$project": tk.M{
		"userid":          1,
		"name":            1,
		"empid":           1,
		"designation":     1,
		"location":        1,
		"departement":     1,
		"reason":          1,
		"email":           1,
		"address":         1,
		"contact":         1,
		"leavefrom":       1,
		"leaveto":         1,
		"noofdays":        1,
		"project":         1,
		"yearleave":       1,
		"publicleave":     1,
		"datecreateleave": 1,
		"statusprojectleader": tk.M{
			"idemp":         1,
			"name":          1,
			"location":      1,
			"email":         1,
			"phonenumber":   1,
			"statusrequest": 1,
			"reason":        1,
			"projectname":   1,
			"userid":        1,
		},
	}}) // "resultrequest":

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeld).
		From("requestLeave").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// defer csr.Close()
	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, err
	}

	for _, ld := range data {
		res = append(res, ld)
	}
	pipemg = append(pipemg, tk.M{"$unwind": "$statusprojectleader"})
	pipemg = append(pipemg, tk.M{"$unwind": "$statusmanagerproject"})
	pipemg = append(pipemg, tk.M{"$match": tk.M{"statusmanagerproject.userid": userid, "statusmanagerproject.statusrequest": tk.M{"$eq": "Pending"}, "statusprojectleader.statusrequest": tk.M{"$ne": "Pending"}, "isemergency": tk.M{"$eq": false}, "resultrequest": tk.M{"$eq": "Pending"}}}) // "resultrequest":             status,
	pipemg = append(pipemg, tk.M{"$project": tk.M{
		"userid":          1,
		"name":            1,
		"empid":           1,
		"designation":     1,
		"location":        1,
		"departement":     1,
		"reason":          1,
		"email":           1,
		"address":         1,
		"contact":         1,
		"leavefrom":       1,
		"leaveto":         1,
		"noofdays":        1,
		"project":         1,
		"yearleave":       1,
		"publicleave":     1,
		"datecreateleave": 1,
		"statusmanagerproject": tk.M{
			"idemp":         1,
			"name":          1,
			"location":      1,
			"email":         1,
			"phonenumber":   1,
			"statusrequest": 1,
			"reason":        1,
			"projectname":   1,
			"userid":        1,
		},
	}}) // "resultrequest":
	// pipemg = append(pipemg, tk.M{"$project": tk.M{}}) // "resultrequest":             status,

	csrm, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipemg).
		From("requestLeave").
		Cursor(nil)

	if csrm != nil {
		defer csrm.Close()
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// defer csrm.Close()

	datamg := []tk.M{}
	if err := csrm.Fetch(&datamg, 0, false); err != nil {
		return nil, err
	}

	for _, mg := range datamg {
		res = append(res, mg)
	}

	return res, nil

}

func (c *DashboardController) PenAppDecRequest(k *knot.WebContext) interface{} {
	now := time.Now()
	defer func() {
		fmt.Println("time elapse", time.Since(now))
	}()

	k.Config.OutputType = knot.OutputJson
	p := struct {
		UserId string
	}{}

	err := k.GetPayload(&p)

	if err != nil {
		return nil
	}

	userid := p.UserId

	pipeP := []tk.M{}
	pipeA := []tk.M{}
	pipeD := []tk.M{}
	pipeEm := []tk.M{}
	res := tk.M{}

	pipeP = append(pipeP, tk.M{"$match": tk.M{"userid": userid, "resultrequest": "Pending"}}) // "resultrequest":             status,

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeP).
		From("requestLeave").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return nil
	}
	if err != nil {
		return nil
	}
	// defer csr.Close()

	dataP := []tk.M{}
	if err := csr.Fetch(&dataP, 0, false); err != nil {
		return nil
	}

	pipeA = append(pipeA, tk.M{"$match": tk.M{"userid": userid, "resultrequest": "Approved"}}) // "resultrequest":             status,

	csrA, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeA).
		From("requestLeave").
		Cursor(nil)

	if csrA != nil {
		defer csrA.Close()
	} else {
		return nil
	}
	if err != nil {
		return nil
	}
	// defer csrA.Close()

	dataA := []tk.M{}
	if err := csrA.Fetch(&dataA, 0, false); err != nil {
		return nil
	}

	// fmt.Println("----------- userid", userid)
	pipeD = append(pipeD, tk.M{"$match": tk.M{"userid": userid, "resultrequest": "Declined"}}) // "resultrequest":             status,

	csrD, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeD).
		From("requestLeave").
		Cursor(nil)

	if csrD != nil {
		defer csrD.Close()
	} else {
		return nil
	}
	if err != nil {
		return nil
	}
	// defer csrD.Close()

	dataD := []tk.M{}
	if err := csrD.Fetch(&dataD, 0, false); err != nil {
		return nil
	}

	pipeEm = append(pipeEm, tk.M{"$unwind": "$statusmanagerproject"})
	pipeEm = append(pipeEm, tk.M{"$match": tk.M{"statusmanagerproject.userid": userid, "isemergency": true, "resultrequest": "Pending"}}) // "resultrequest":
	csrEm, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipeEm).
		From("requestLeave").
		Cursor(nil)

	if csrEm != nil {
		defer csrEm.Close()
	} else {
		return nil
	}

	if err != nil {
		return nil
	}
	// defer csrEm.Close()

	dataEm := []tk.M{}
	if err := csrEm.Fetch(&dataEm, 0, false); err != nil {
		return nil
	}

	dataReq, err := c.FetchRequestCount(k, userid, "Pending")
	if err != nil {
		return nil
	}

	// count overtime Pending , Approved, Declined
	pipe := []tk.M{}
	dOvertimePending := []tk.M{}
	pipe = append(pipe, tk.M{"$unwind": "$membersovertime"})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", userid)))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("resultrequest", "Pending")))
	csr, err = c.Ctx.Connection.NewQuery().Command("pipe", pipe).From(NewOvertimeModel().TableName()).Cursor(nil)
	if csr != nil {
		defer csr.Close()
	}
	if err = csr.Fetch(&dOvertimePending, 0, false); err != nil {
		return nil
	}
	pipe = []tk.M{}
	dOvertimeApproved := []tk.M{}
	pipe = append(pipe, tk.M{"$unwind": "$membersovertime"})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", userid)))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("resultrequest", "Approved")))
	csr, err = c.Ctx.Connection.NewQuery().Command("pipe", pipe).From(NewOvertimeModel().TableName()).Cursor(nil)
	if csr != nil {
		defer csr.Close()
	}
	if err = csr.Fetch(&dOvertimeApproved, 0, false); err != nil {
		return nil
	}
	pipe = []tk.M{}
	dOvertimeDeclined := []tk.M{}
	pipe = append(pipe, tk.M{"$unwind": "$membersovertime"})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", userid)))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("resultrequest", "Declined")))
	csr, err = c.Ctx.Connection.NewQuery().Command("pipe", pipe).From(NewOvertimeModel().TableName()).Cursor(nil)
	if csr != nil {
		defer csr.Close()
	}
	if err = csr.Fetch(&dOvertimeDeclined, 0, false); err != nil {
		return nil
	}

	// res.Set("Pending", len(dataP))
	// res.Set("Approved", len(dataA))
	// res.Set("Decline", len(dataD))
	// res.Set("Emergency", len(dataEm))
	// res.Set("Request", len(dataReq))
	// res.Set("Total", len(dataReq)+len(dataEm)+len(dataP))

	res.Set("Pending", len(dataP)+len(dOvertimePending))
	res.Set("Approved", len(dataA)+len(dOvertimeApproved))
	res.Set("Decline", len(dataD)+len(dOvertimeDeclined))
	res.Set("Emergency", len(dataEm))
	res.Set("Request", len(dataReq))
	res.Set("Total", len(dataReq)+len(dataEm)+len(dataP)+len(dOvertimePending)+len(dOvertimeApproved)+len(dOvertimeDeclined))

	// fmt.Println("------------ data", res)

	return c.SetResultInfo(false, "success", res)
	// return "nanana"

}

func (c *DashboardController) PendingRequest(k *knot.WebContext, userid string) ([]RequestLeaveModel, error) {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	// pipe = append(pipe, tk.M{"$sort": tk.M{"datecreateleave": -1}})
	// fmt.Println("----------------- notif ", userid)
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "resultrequest": "Pending"}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeave").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// defer csr.Close()

	data := []RequestLeaveModel{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, err
	}

	return data, nil

}

func (c *DashboardController) NotYetCutLeave(k *knot.WebContext, userid string) ([]AprovalRequestLeaveModel, error) {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	// pipe = append(pipe, tk.M{"$sort": tk.M{"datecreateleave": -1}})
	// fmt.Println("----------------- notif ", userid)
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "stsbymanager": "Approved"}})
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "iscutoff": false}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeaveByDate").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// defer csr.Close()

	data := []AprovalRequestLeaveModel{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, err
	}

	return data, nil

}
func (c *DashboardController) GetAllUserRAWData(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	results, err := new(services.UserService).GetAll()
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	return c.SetResultOK(results)
}

func (c *DashboardController) AdminDashboard(k *knot.WebContext) interface{} {
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

func (c *DashboardController) GetDashboardForAdmin(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := services.PayloadDashboardService{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	remote, err := new(services.DashboardService).ConstructDashboardDataForAdminReport(payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(remote)
}

func (c *DashboardController) WriteExcelAdminReport(k *knot.WebContext) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	fileName := "admin-dashboard-report_" + time.Now().Format("2006-01-02_15-04-05") + ".xlsx"

	payload := services.PayloadDashboardService{}
	payload.DateByMonth = k.Request.FormValue("DateByMonth")

	_, err := new(services.DashboardService).WriteExcelForAdmin(payload, fileName)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	// f, err := os.Open(helper.RenderPathDoc(fileName))
	// if err != nil {
	// 	return c.SetResultError(err.Error(), nil)
	// }

	// contentDisposition := fmt.Sprintf("attachment; filename=%s", fileName)
	// k.Writer.Header().Set("Content-Disposition", contentDisposition)
	// if _, err := io.Copy(k.Writer, f); err != nil {
	// 	return c.SetResultError(err.Error(), nil)
	// }

	return fileName
}
func (c *DashboardController) GetNationalHolidays(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Year     int
		Location string
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	location := []string{"Global"}
	location = append(location, payload.Location)
	// tk.Println(location)
	filter := []*db.Filter{}
	filter = append(filter, db.Eq("year", payload.Year))
	filter = append(filter, db.Or(db.Eq("location", "Global"), db.Eq("location", payload.Location)))
	csr, err := c.Ctx.Connection.NewQuery().Select().From("NationalHolidays").Where(db.And(filter...)).Cursor(nil)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	defer csr.Close()
	results := []NationalHolidaysModel{}
	err = csr.Fetch(&results, 0, false)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	// tk.Println(results)
	data := []string{}
	for _, each := range results {
		for _, each := range each.ListDate {
			data = append(data, each.Format("02012006"))
		}
	}
	return c.SetResultInfo(false, "Success", data)
}

func (c *DashboardController) GetDocCert(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	// pipe = append(pipe, tk.M{"$sort": tk.M{"datecreateleave": -1}})
	// fmt.Println("----------------- notif ", userid)
	// pipe = append(pipe,
	// 	tk.M{"$match": tk.M{"resultrequest": "Approved", "resultrequest": "Pending"}})

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"resultrequest": tk.M{"$eq": "Approved"}},
					{"isemergency": tk.M{"$eq": true}},
					{"isreset": tk.M{"$eq": false}},
					{"isattach": tk.M{"$eq": true}},
					{"filelocation": tk.M{"$nin": []interface{}{nil, ""}}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeave").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}
	// defer csr.Close()

	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil
	}
	return data
}

func (c *DashboardController) GetUserTkm(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)
	payload := struct {
		EmpId string
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	userid := payload.EmpId

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserProfileModel, 0)
	if userid != "nil" {

		dbFilter = append(dbFilter, db.Eq("empid", userid))

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewSysUserModel(), query)
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

func inTime(time, check time.Time) bool {
	return check.After(time)
}

func (c *DashboardController) RequestAdmin(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	p := new(RequestLeaveModel)
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	tz, err := c.GetTimeZone(k, p.Location)
	tl, err := helper.TimeLocation(tz)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
	}
	i := 0
	tnow := time.Now()
	for _, d := range p.LeaveDateList {
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}
		onday := dt.Weekday()
		if onday.String() == "Sunday" || onday.String() == "Saturday" {

		} else {
			i = i + 1
		}
	}

	p.ResultRequest = "Approved"
	p.IsEmergency = true
	p.DateCreateLeave = tl
	p.NoOfDays = i
	p.IsReset = false
	p.IsAttach = false
	rem := c.CheckRemote(k, p)
	if rem != nil {
		return rem
	} else if rem == nil {
		res := c.CheckLeaveByUser(k, p, p.EmpId)
		if res != nil {
			return res
		} else if res == nil {
			for _, d := range p.LeaveDateList {
				//get user by emp id
				dt, _ := time.Parse("2006-01-02", d)
				getSysUser, _ := c.GetByEmpId(k, p.EmpId)
				if inTime(dt, tnow) {
					//fmt.Println("ok...", c.inTime(dt, tnow))
					getSysUser.DecYear = getSysUser.DecYear - 1
					getSysUser.YearLeave = getSysUser.YearLeave - 1
					//update by date leave when date less than now
					err = c.Ctx.Save(&getSysUser)
					if err != nil {
						return c.SetResultInfo(true, err.Error(), nil)
					}
				}
			}
		}
	}

	err = c.Ctx.Save(p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	level := k.Session("jobrolelevel").(int)
	mailContro := MailController(*c)
	mailContro.EmergencyLeaveAdminRequest(k, p, level, p.UserId)
	c.SetHistoryLeave(k, p.UserId, p.Id, p.LeaveFrom, p.LeaveTo, "Request Emergency Leave", "Approved", p)
	return c.SetResultInfo(false, "Request succesfully sent", nil)
}

func (c *DashboardController) GetByEmpId(k *knot.WebContext, empid string) (SysUserModel, error) {
	var dbFilter []*db.Filter
	query := tk.M{}
	data := SysUserModel{}
	dbFilter = append(dbFilter, db.Eq("empid", empid))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return data, nil
	}

	err = crs.Fetch(&data, 1, false)
	if err != nil {
		return data, err
	}
	return data, nil
}

// ValidateLeave is ...
func (c *DashboardController) ValidateLeave(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)
	p := struct {
		Dates int
	}{}
	err := k.GetPayload(&p)
	if err != nil {
		tk.Println(err)
	}
	userid := k.Session("userid").(string)
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
			return c.SetResultInfo(true, "user not found", nil)
		}
		if errdata != nil {
			return c.SetResultInfo(true, "user not found", nil)
		}
		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return c.SetResultInfo(true, "user not found", nil)
		}
		//validate year leave
		notCut, _ := c.NotYetCutLeave(k, userid)

		tempY := data[0].YearLeave - len(notCut) - p.Dates

		if tempY < 0 {
			return c.SetResultInfo(true, "You don't have any annual leave left, additional leave taken will be charged from your monthly salary.", nil)
		}
		return c.SetResultInfo(false, "success", nil)
	}
	return c.SetResultInfo(true, "user not found", nil)
}
