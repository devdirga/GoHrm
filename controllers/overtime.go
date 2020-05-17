package controllers

import (
	"bytes"
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	db "github.com/creativelab/dbox"
	gomail "gopkg.in/gomail.v2"

	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type OvertimeController struct {
	*BaseController
}

func (d *OvertimeController) AdminOvertime(r *knot.WebContext) interface{} {
	d.LoadBase(r)
	viewData := tk.M{}
	if r.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", r.Session("jobrolename").(string))
		viewData.Set("UserId", r.Session("userid").(string))
		viewData.Set("JobRoleLevel", r.Session("jobrolelevel"))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
		viewData.Set("UserId", "")
	}

	DataAccess := d.SetViewData(r, viewData)

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.IncludeFiles = []string{
		"_loader.html",
	}

	return DataAccess
}

func (d *OvertimeController) UserOvertime(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *OvertimeController) ApprovedOvertime(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *OvertimeController) DeclinedOvertime(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *OvertimeController) ManagerOvertime(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}
func (c *OvertimeController) ResponseApproveManager(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		Param  string
		Reason string
	}{}

	err := r.GetPayload(&p)
	// fmt.Println("--------------- payload", p.IdRequest)
	if err != nil {
		return err.Error
	}

	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	mail := MailController(*c)

	payload := new(ParameterURLUserModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), payload)

	managerName := ""
	emailManager := ""

	data, err := c.GetOvertimeByID(r, payload.IdOvertime)

	// tk.Println("-------------- masuk sini 2s", data)

	tmpOvertimeid := ""
	tmpOveretimeUserid := ""
	tmpOvertimeDate := ""

	if payload.Result == "yes" {

		// tk.Println("-------------- masuk sini 1")

		if data.ProjectManager.UserId == payload.UserId {
			data.ApprovalManager.Result = "Approved"
			data.ApprovalManager.Name = data.ProjectManager.Name
			data.ApprovalManager.Location = data.ProjectManager.Location
			data.ApprovalManager.Email = data.ProjectManager.Email
			data.ApprovalManager.PhoneNumber = data.ProjectManager.PhoneNumber
			data.ApprovalManager.UserId = data.ProjectManager.UserId
			data.ApprovalManager.Reason = ""

			managerName = data.ProjectManager.Name
			emailManager = data.ProjectManager.Email
		}

		for _, br := range data.BranchManagers {
			if br.UserId == payload.UserId {
				data.ApprovalManager.Result = "Approved"
				data.ApprovalManager.Name = br.Name
				data.ApprovalManager.Location = br.Location
				data.ApprovalManager.Email = br.Email
				data.ApprovalManager.PhoneNumber = br.PhoneNumber
				data.ApprovalManager.UserId = br.UserId
				data.ApprovalManager.Reason = ""

				managerName = br.Name
				emailManager = br.Email
			}
		}

		data.ResultRequest = "Approved"
		for i, _ := range data.DayList {
			data.DayList[i].Result = "Approved"
		}

		for m, dev := range data.MembersOvertime {
			if data.ResultRequest == "Approved" {
				data.MembersOvertime[m].Result = "Approved"
			}

			to := []string{dev.Email}
			if c.VerifyIsAlreadyHasRequestOvertime(data.Id, dev.UserId, data.DayList[0].Date) {

			} else {

				for _, dt := range data.DayList {
					tmpOvertimeid = data.Id
					tmpOveretimeUserid = dev.UserId
					tmpOvertimeDate = dt.Date
					tk.Println("---------- ", dev.TypeOvertime)
					c.DateEmployeeOvertime(r, data.Id, dev.IdEmp, data.Project, dev.UserId, dev.Name, dev.Location, dev.Email, dev.PhoneNumber, dt.Date, dev.TypeOvertime, dev.Hours)
				}
				c.SendUsersOvertime(r, to, dev.UserId, data.Id, dev.Name, data.Project, data)

			}

		}

	} else {
		if data.ResultRequest == "Approved" {
			return c.SetResultInfo(true, "request Already Approved", nil)
		} else if data.ResultRequest == "Declined" {
			return c.SetResultInfo(true, "request Already Declined", nil)
		} else if data.ResultRequest == "Expired" {
			return c.SetResultInfo(true, "request Already Expired", nil)
		}
		if data.ProjectManager.UserId == payload.UserId {
			// tk.Println("-------------- masuk sini 1")
			data.ApprovalManager.Result = "Declined"
			data.ApprovalManager.Name = data.ProjectManager.Name
			data.ApprovalManager.Location = data.ProjectManager.Location
			data.ApprovalManager.Email = data.ProjectManager.Email
			data.ApprovalManager.PhoneNumber = data.ProjectManager.PhoneNumber
			data.ApprovalManager.UserId = data.ProjectManager.UserId
			data.ApprovalManager.Reason = p.Reason

			managerName = data.ProjectManager.Name
			emailManager = data.ProjectManager.Email
		}

		for _, br := range data.BranchManagers {
			if br.UserId == payload.UserId {
				// tk.Println("-------------- masuk sini 2s")
				data.ApprovalManager.Result = "Declined"
				data.ApprovalManager.Name = br.Name
				data.ApprovalManager.Location = br.Location
				data.ApprovalManager.Email = br.Email
				data.ApprovalManager.PhoneNumber = br.PhoneNumber
				data.ApprovalManager.UserId = br.UserId
				data.ApprovalManager.Reason = p.Reason

				managerName = br.Name
				emailManager = br.Email
			}
		}

		data.ResultRequest = "Declined"
		data.DeclineReason = p.Reason
		for m, _ := range data.MembersOvertime {
			if data.ResultRequest == "Declined" {
				data.MembersOvertime[m].Result = "Declined"
			}

		}

	}

	err = c.Ctx.Save(&data)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	inahrd := []string{hrd}

	leader := []string{data.Email}

	if c.VerifyIsAlreadyHasRequestOvertime(tmpOvertimeid, tmpOveretimeUserid, tmpOvertimeDate) {
		c.SendMailDeclineApproved(r, inahrd, "", data.Id, "INA HRD members", data.Project, data)
		mail.DelayProcess(5)

		c.SendMailDeclineApproved(r, leader, "", data.Id, data.Name, data.Project, data)

	} else {

	}

	// tk.Println("-------------- terus sini")

	// add log overtime
	service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: &data}

	log := tk.M{}
	log.Set("Status", data.ResultRequest)
	if data.ResultRequest == "Approved" {
		log.Set("Desc", "Request Approved by Manager")
	} else {
		log.Set("Desc", "Request Declined by Manager")
	}
	log.Set("NameLogBy", managerName)
	log.Set("EmailNameLogBy", emailManager)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		c.SetResultInfo(true, "Error occured in overtime when overtime approved", nil)
	}
	//

	notif := NotificationController(*c)
	getnotif := notif.GetDataNotification(r, data.Id)
	getnotif.Notif.ManagerApprove = data.ApprovalManager.Name
	getnotif.Notif.Status = data.ResultRequest
	getnotif.Notif.StatusApproval = data.ApprovalManager.Result
	if data.ResultRequest != "Approved" {
		getnotif.Notif.Description = p.Reason
	}
	notif.InsertNotification(getnotif)

	return c.SetResultInfo(false, "Request approved successfully", nil)

}

func (c *OvertimeController) ApproveByDate(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := []struct {
		Idrequest     string
		Date          string
		Result        string
		Reason        string
		ManagerUserid string
	}{}

	err := r.GetPayload(&p)

	if err != nil {
		return err.Error
	}

	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	mail := MailController(*c)

	ManagerUserid := p[0].ManagerUserid

	data, err := c.GetOvertimeByID(r, p[0].Idrequest)
	if err != nil {
		return err.Error
	}
	data.IsExpired = false
	im := 0
	mlist := []ApprovemenDay{}
	dlist := ApprovemenDay{}
	for i, _ := range data.DayList {

		data.DayList[i].Result = p[i].Result
		data.DayList[i].Reason = p[i].Reason
		fmt.Println("--------------- payload", p[i].Result)
		if p[i].Result == "Approved" {
			im = im + 1
			dlist.Date = p[i].Date
			dlist.Result = p[i].Result
			dlist.Reason = p[i].Reason
			mlist = append(mlist, dlist)

		}

	}
	// data.DayList = mlist
	data.DayDuration = len(mlist)

	fmt.Println("--------------- masuk sini ", im)

	if im > 0 {
		managerName := ""
		emailManager := ""
		data.ResultRequest = "Approved"
		tmpOvertimeid := ""
		tmpOveretimeUserid := ""
		tmpOvertimeDate := ""

		if data.ProjectManager.UserId == ManagerUserid {
			data.ApprovalManager.Result = "Approved"
			data.ApprovalManager.Name = data.ProjectManager.Name
			data.ApprovalManager.Location = data.ProjectManager.Location
			data.ApprovalManager.Email = data.ProjectManager.Email
			data.ApprovalManager.PhoneNumber = data.ProjectManager.PhoneNumber
			data.ApprovalManager.UserId = data.ProjectManager.UserId
			data.ApprovalManager.IdEmp = data.ProjectManager.IdEmp
			data.ApprovalManager.Reason = ""

			managerName = data.ProjectManager.Name
			emailManager = data.ProjectManager.Email
		}

		for _, br := range data.BranchManagers {
			if br.UserId == ManagerUserid {
				data.ApprovalManager.Result = "Approved"
				data.ApprovalManager.Name = br.Name
				data.ApprovalManager.Location = br.Location
				data.ApprovalManager.Email = br.Email
				data.ApprovalManager.PhoneNumber = br.PhoneNumber
				data.ApprovalManager.UserId = br.UserId
				data.ApprovalManager.IdEmp = br.IdEmp
				data.ApprovalManager.Reason = ""

				managerName = br.Name
				emailManager = br.Email
			}
		}

		for m, dev := range data.MembersOvertime {
			if data.ResultRequest == "Approved" {
				data.MembersOvertime[m].Result = "Approved"
			}

			to := []string{dev.Email}
			if c.VerifyIsAlreadyHasRequestOvertime(data.Id, dev.UserId, data.DayList[0].Date) {

			} else {

				for _, dt := range data.DayList {
					tmpOvertimeid = data.Id
					tmpOveretimeUserid = dev.UserId
					tmpOvertimeDate = dt.Date
					tk.Println("---------- ", dev.TypeOvertime)
					if dt.Result != "Declined" {
						c.DateEmployeeOvertime(r, data.Id, dev.IdEmp, data.Project, dev.UserId, dev.Name, dev.Location, dev.Email, dev.PhoneNumber, dt.Date, dev.TypeOvertime, dev.Hours)
					}

				}
				c.SendUsersOvertime(r, to, dev.UserId, data.Id, dev.Name, data.Project, data)

			}

		}

		inahrd := []string{hrd}

		leader := []string{data.Email}

		// leader := []string{data.ProjectLeader.Email}
		tk.Println("---------------- verify ", c.VerifyIsAlreadyHasRequestOvertime(tmpOvertimeid, tmpOveretimeUserid, tmpOvertimeDate))
		if !c.VerifyIsAlreadyApproveOvertime(tmpOvertimeid, tmpOveretimeUserid, tmpOvertimeDate) {
			tk.Println("---------------- masuk sini 1")
			c.SendMailDeclineApproved(r, inahrd, "", data.Id, "INA HRD members", data.Project, data)
			mail.DelayProcess(5)

			c.SendMailDeclineApproved(r, leader, "", data.Id, data.Name, data.Project, data)

			// c.SendMailDeclineApproved(r, leader, "", data.Id, data.ProjectLeader.Name, data.Project, data)
		} else {

		}

		err = c.Ctx.Save(&data)

		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		// tk.Println("-------------- terus sini")

		// add log overtime
		service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: &data}

		log := tk.M{}
		log.Set("Status", data.ResultRequest)
		if data.ResultRequest == "Approved" {
			log.Set("Desc", "Request Approved by Manager")
		} else {
			log.Set("Desc", "Request Declined by Manager")
		}
		log.Set("NameLogBy", managerName)
		log.Set("EmailNameLogBy", emailManager)
		err = service.ApproveDeclineLog(log)
		if err != nil {
			c.SetResultInfo(true, "Error occured in overtime when overtime approved", nil)
		}
		//

		notif := NotificationController(*c)
		getnotif := notif.GetDataNotification(r, data.Id)
		getnotif.Notif.ManagerApprove = data.ApprovalManager.Name
		getnotif.Notif.Status = data.ResultRequest
		getnotif.Notif.StatusApproval = data.ApprovalManager.Result
		if data.ResultRequest != "Approved" {
			getnotif.Notif.Description = ""
		}
		notif.InsertNotification(getnotif)

	}

	return c.SetResultInfo(false, "Request successfully saved", nil)

}

func (c *OvertimeController) ResponseApproveManagerApp(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := []struct {
		Idrequest     string
		Result        string
		ManagerUserId string
		Reason        string
	}{}

	err := r.GetPayload(&p)
	fmt.Println("--------------- payload", tk.JsonString(p))
	if err != nil {
		return err.Error
	}

	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	mail := MailController(*c)

	for _, pl := range p {
		managerName := ""
		emailManager := ""
		data, err := c.GetOvertimeByID(r, pl.Idrequest)
		tmpOvertimeid := ""
		tmpOveretimeUserid := ""
		tmpOvertimeDate := ""
		if data.IsExpired == true {
			data.IsExpired = false
		}
		if pl.Result == "Approved" {
			if data.ResultRequest == "Approved" {
				return c.SetResultInfo(true, "request Already Approved", nil)
			} else if data.ResultRequest == "Declined" {
				return c.SetResultInfo(true, "request Already Declined", nil)
			}
			if data.ProjectManager.UserId == pl.ManagerUserId {
				data.ApprovalManager.Result = "Approved"
				data.ApprovalManager.Name = data.ProjectManager.Name
				data.ApprovalManager.Location = data.ProjectManager.Location
				data.ApprovalManager.Email = data.ProjectManager.Email
				data.ApprovalManager.PhoneNumber = data.ProjectManager.PhoneNumber
				data.ApprovalManager.UserId = data.ProjectManager.UserId
				data.ApprovalManager.IdEmp = data.ProjectManager.IdEmp
				data.ApprovalManager.Reason = ""

				managerName = data.ProjectManager.Name
				emailManager = data.ProjectManager.Email
			}

			for _, br := range data.BranchManagers {
				if br.UserId == pl.ManagerUserId {
					data.ApprovalManager.Result = "Approved"
					data.ApprovalManager.Name = br.Name
					data.ApprovalManager.Location = br.Location
					data.ApprovalManager.Email = br.Email
					data.ApprovalManager.PhoneNumber = br.PhoneNumber
					data.ApprovalManager.UserId = br.UserId
					data.ApprovalManager.IdEmp = br.IdEmp
					data.ApprovalManager.Reason = ""

					managerName = br.Name
					emailManager = br.Email
				}
			}

			data.ResultRequest = "Approved"
			for i, _ := range data.DayList {
				data.DayList[i].Result = "Approved"
			}

			for m, dev := range data.MembersOvertime {
				if data.ResultRequest == "Approved" {
					data.MembersOvertime[m].Result = "Approved"
				}

				to := []string{dev.Email}
				if c.VerifyIsAlreadyHasRequestOvertime(data.Id, dev.UserId, data.DayList[0].Date) {

				} else {

					for _, dt := range data.DayList {
						tmpOvertimeid = data.Id
						tmpOveretimeUserid = dev.UserId
						tmpOvertimeDate = dt.Date
						tk.Println("---------- ", dev.TypeOvertime)
						c.DateEmployeeOvertime(r, data.Id, dev.IdEmp, data.Project, dev.UserId, dev.Name, dev.Location, dev.Email, dev.PhoneNumber, dt.Date, dev.TypeOvertime, dev.Hours)
					}
					c.SendUsersOvertime(r, to, dev.UserId, data.Id, dev.Name, data.Project, data)

				}

			}
		} else {
			if data.ResultRequest == "Approved" {
				return c.SetResultInfo(true, "request Already Approved", nil)
			} else if data.ResultRequest == "Declined" {
				return c.SetResultInfo(true, "request Already Declined", nil)
			}
			// else if data.ResultRequest == "Expired" {
			// 	return c.SetResultInfo(true, "request Already Expired", nil)
			// }
			if data.ProjectManager.UserId == pl.ManagerUserId {
				// tk.Println("-------------- masuk sini 1")
				data.ApprovalManager.Result = "Declined"
				data.ApprovalManager.Name = data.ProjectManager.Name
				data.ApprovalManager.Location = data.ProjectManager.Location
				data.ApprovalManager.Email = data.ProjectManager.Email
				data.ApprovalManager.PhoneNumber = data.ProjectManager.PhoneNumber
				data.ApprovalManager.UserId = data.ProjectManager.UserId
				data.ApprovalManager.IdEmp = data.ProjectManager.IdEmp
				data.ApprovalManager.Reason = pl.Reason

				managerName = data.ProjectManager.Name
				emailManager = data.ProjectManager.Email
			}

			for _, br := range data.BranchManagers {
				if br.UserId == pl.ManagerUserId {
					// tk.Println("-------------- masuk sini 2s")
					data.ApprovalManager.Result = "Declined"
					data.ApprovalManager.Name = br.Name
					data.ApprovalManager.Location = br.Location
					data.ApprovalManager.Email = br.Email
					data.ApprovalManager.PhoneNumber = br.PhoneNumber
					data.ApprovalManager.UserId = br.UserId
					data.ApprovalManager.IdEmp = br.IdEmp
					data.ApprovalManager.Reason = pl.Reason

					managerName = br.Name
					emailManager = br.Email
				}
			}

			data.ResultRequest = "Declined"
			data.DeclineReason = pl.Reason
			for i, _ := range data.DayList {
				data.DayList[i].Result = "Declined"
			}

			for m, _ := range data.MembersOvertime {
				if data.ResultRequest == "Declined" {
					data.MembersOvertime[m].Result = "Declined"
				}

			}

		}
		inahrd := []string{hrd}

		leader := []string{data.Email}

		// leader := []string{data.ProjectLeader.Email}
		tk.Println("---------------- verify ", c.VerifyIsAlreadyHasRequestOvertime(tmpOvertimeid, tmpOveretimeUserid, tmpOvertimeDate))
		if !c.VerifyIsAlreadyApproveOvertime(tmpOvertimeid, tmpOveretimeUserid, tmpOvertimeDate) {
			tk.Println("---------------- masuk sini 1")
			c.SendMailDeclineApproved(r, inahrd, "", data.Id, "INA HRD members", data.Project, data)
			mail.DelayProcess(5)

			c.SendMailDeclineApproved(r, leader, "", data.Id, data.Name, data.Project, data)

		} else {

		}

		err = c.Ctx.Save(&data)

		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		// tk.Println("-------------- terus sini")

		// add log overtime
		service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: &data}

		log := tk.M{}
		log.Set("Status", data.ResultRequest)
		if data.ResultRequest == "Approved" {
			log.Set("Desc", "Request Approved by Manager")
		} else {
			log.Set("Desc", "Request Declined by Manager")
		}
		log.Set("NameLogBy", managerName)
		log.Set("EmailNameLogBy", emailManager)
		err = service.ApproveDeclineLog(log)
		if err != nil {
			c.SetResultInfo(true, "Error occured in overtime when overtime approved", nil)
		}
		//

		notif := NotificationController(*c)
		getnotif := notif.GetDataNotification(r, data.Id)
		getnotif.Notif.ManagerApprove = data.ApprovalManager.Name
		getnotif.Notif.Status = data.ResultRequest
		getnotif.Notif.StatusApproval = data.ApprovalManager.Result
		if data.ResultRequest != "Approved" {
			getnotif.Notif.Description = pl.Reason
		}
		notif.InsertNotification(getnotif)

	}

	return c.SetResultInfo(false, "Request successfully saved", nil)
}

func (c *OvertimeController) VerifyIsAlreadyApproveOvertime(overtimeid string, userid string, overtimedate string) bool {
	res := []EmployeeOvertimeModel{}
	var dbFilter []*db.Filter
	query := tk.M{}
	dbFilter = append(dbFilter, db.Eq("_id", overtimeid))
	dbFilter = append(dbFilter, db.Eq("userid", userid))
	dbFilter = append(dbFilter, db.Eq("resultrequest", tk.M{"$ne": "Pending"}))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	crs, err := c.Ctx.Find(NewOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return false
	}
	if err != nil {
		return false
	}
	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return false
	}
	if len(res) > 0 {
		return true
	} else {
		return false
	}
}

func (c *OvertimeController) GetOvertimeByID(r *knot.WebContext, id string) (OvertimeModel, error) {
	r.Config.OutputType = knot.OutputJson
	res := OvertimeModel{}
	var dbFilter []*db.Filter
	query := tk.M{}

	dbFilter = append(dbFilter, db.Eq("_id", id))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return res, err
	}

	if err != nil {
		return res, err
	}

	err = crs.Fetch(&res, 1, false)
	if err != nil {
		return res, err
	}
	return res, err
}

func (c *OvertimeController) LeaderRequestOvertime(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := OvertimeModel{}

	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
	}

	level := k.Session("jobrolelevel").(int)

	dash := DashboardController(*c)
	tz, err := dash.GetTimeZone(k, p.Location)
	tl, err := helper.TimeLocation(tz)

	p.DateCreated = tl

	now, err := helper.ExpiredDateTime(tz, true)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	p.ExpiredOn = now

	rm, err := helper.ExpiredRemaining(tz, true)
	p.ExpiredRemining = rm
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	inahrd := []string{hrd}

	if level == 1 {
		p.ResultRequest = "Approved"
		for d, dtl := range p.DayList {
			p.DayList[d].Result = "Approved"
			for _, m := range p.MembersOvertime {
				idovertime := p.Id
				idemployee := m.IdEmp
				project := p.Project
				userid := m.UserId
				name := m.Name
				location := p.Location
				email := m.Email
				phonenumber := m.PhoneNumber
				date := dtl.Date
				tipe := ""
				hour := 0
				c.DateEmployeeOvertime(k, idovertime, idemployee, project, userid, name, location, email, phonenumber, date, tipe, hour)
			}
		}
		for l, _ := range p.MembersOvertime {
			p.MembersOvertime[l].Result = "Approved"
		}
		if p.ProjectManager.UserId == k.Session("userid").(string) {
			p.ApprovalManager.UserId = p.ProjectManager.UserId
			p.ApprovalManager.IdEmp = p.ProjectManager.IdEmp
			p.ApprovalManager.Location = p.ProjectManager.Location
			p.ApprovalManager.Name = p.ProjectManager.Name
			p.ApprovalManager.PhoneNumber = p.ProjectManager.PhoneNumber
			p.ApprovalManager.Email = p.ProjectManager.Email
		}

		p.ApprovalManager.Result = "Approved"
		err = c.Ctx.Save(&p)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		for _, dec := range p.MembersOvertime {
			c.SendUsersOvertime(k, []string{dec.Email}, dec.UserId, p.Id, dec.Name, p.Project, p)
		}

		c.SendMailDeclineApproved(k, inahrd, "", p.Id, "INA HRD members", p.Project, p)
	} else if level == 6 {
		p.ResultRequest = "Approved"
		for d, dtl := range p.DayList {
			p.DayList[d].Result = "Approved"
			for _, m := range p.MembersOvertime {
				idovertime := p.Id
				idemployee := m.IdEmp
				project := p.Project
				userid := m.UserId
				name := m.Name
				location := p.Location
				email := m.Email
				phonenumber := m.PhoneNumber
				date := dtl.Date
				tipe := ""
				hour := 0
				c.DateEmployeeOvertime(k, idovertime, idemployee, project, userid, name, location, email, phonenumber, date, tipe, hour)
			}
		}
		for l, _ := range p.MembersOvertime {
			p.MembersOvertime[l].Result = "Approved"
		}
		for _, br := range p.BranchManagers {
			if br.UserId == k.Session("userid").(string) {
				p.ApprovalManager.UserId = br.UserId
				p.ApprovalManager.IdEmp = br.IdEmp
				p.ApprovalManager.Location = br.Location
				p.ApprovalManager.Name = br.Name
				p.ApprovalManager.PhoneNumber = br.PhoneNumber
				p.ApprovalManager.Email = br.Email
			}
		}
		p.ApprovalManager.Result = "Approved"
		err = c.Ctx.Save(&p)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		for _, dec := range p.MembersOvertime {
			c.SendUsersOvertime(k, []string{dec.Email}, dec.UserId, p.Id, dec.Name, p.Project, p)
		}

		c.SendMailDeclineApproved(k, inahrd, "", p.Id, "INA HRD members", p.Project, p)

	} else {
		err = c.Ctx.Save(&p)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		c.SendManagerOvertime(k, []string{p.ProjectManager.Email}, p.ProjectManager.UserId, p.Id, p.ProjectManager.Name, p.Project, p)
		c.SendUserInfoOvertime(k, p.ProjectManager.UserId, p.Id, p.ProjectManager.Name, p.Project, p)

		for _, bm := range p.BranchManagers {
			c.SendManagerOvertime(k, []string{bm.Email}, bm.UserId, p.Id, bm.Name, p.Project, p)
		}

		// add log overtime
		service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: &p}
		err = service.RequestLog()
		if err != nil {
		}
		//

		tm := time.Now()
		notif := NotificationController(*c)
		dataNotif := NotificationModel{}
		dataNotif.Notif.CreatedAt = tm.Format("2006-01-02 15:04:05")
		dataNotif.Notif.UpdatedAt = tm.Format("2006-01-02 15:04:05")
		dataNotif.Id = ""
		dataNotif.UserId = p.UserId
		dataNotif.IsConfirmed = false
		dataNotif.Notif.Name = p.Name
		dataNotif.IdRequest = p.Id
		dataNotif.Notif.DateFrom = p.DayList[0].Date
		dataNotif.Notif.DateTo = p.DayList[len(p.DayList)-1].Date
		dataNotif.Notif.Description = "Create request Overtime"
		dataNotif.Notif.Status = p.ResultRequest
		dataNotif.Notif.RequestType = "Overtime"
		dataNotif.Notif.Reason = p.Reason
		dataNotif.Notif.ManagerApprove = p.ApprovalManager.Name
		dataNotif.Notif.StatusApproval = p.ApprovalManager.Result
		notif.InsertNotification(dataNotif)
	}

	return c.SetResultInfo(false, "success", nil)
}

func (c *OvertimeController) SaveOvertime(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Data     OvertimeFormModel
		ListDate []string
	}{}
	e := k.GetPayload(&payload)
	if e != nil {
		c.SetResultInfo(true, e.Error(), nil)
	}
	location := k.Session("location").(string)
	dash := DashboardController(*c)
	tz, err := dash.GetTimeZone(k, location)
	overtimeS := services.OvertimeService{
		payload.Data,
		payload.ListDate,
		[]OvertimeFormModel{},
	}

	respon, err := overtimeS.ProcessOvertime(tz)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), respon)
	}

	return c.SetResultInfo(false, "succes", respon)
}

func (c *OvertimeController) HandleDecline(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"

	param := tk.M{}
	param.Set("Param", k.Request.FormValue("Param"))
	param.Set("Note", k.Request.FormValue("Note"))

	return param
}
func (c *OvertimeController) HandleApproval(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}
	param.Set("Param", k.Request.FormValue("Param"))

	status := ""
	message := ""

	// userid := k.Session("userid").(string)
	// dash := DashboardController(*c)
	// usr := dash.GetDataSessionUser(k, userid)

	paramRes, err := new(services.OvertimeService).ValidateParamMail(param, "")
	if err != nil {
		res := c.SetResultError(err.Error(), paramRes)
		status = string(res.Status)
		message = res.Message
	}

	res := c.SetResultOK(paramRes)

	status = paramRes.GetString("Status")
	message = res.Message
	typeOfRequest := paramRes.GetString("Type")
	if paramRes.Get("IsExpired") != nil {
		if paramRes.Get("IsExpired").(bool) {
			message = "request remote is already expired"
		}
	}
	if typeOfRequest == "cancelremote" {
		if status == "true" {
			message = "you already approved cancel request remote"
		} else if status == "false" {
			message = "you already decline cancel request remote"
		}
	}

	http.Redirect(k.Writer, k.Request, "/overtime/responapproval?Status="+status+"&Message="+message, http.StatusTemporaryRedirect)

	return res
}

// func (c *OvertimeController) UserOvertime(k *knot.WebContext) interface{} {
// 	k.Config.OutputType = knot.OutputTemplate
// 	k.Config.LayoutTemplate = "_layoutEmail.html"

// 	dataView := tk.M{}
// 	dataView.Set("Status", k.Request.FormValue("Status"))
// 	dataView.Set("Message", k.Request.FormValue("Message"))

// 	return dataView
// }

func (c *OvertimeController) ResponApproval(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"

	dataView := tk.M{}
	dataView.Set("Status", k.Request.FormValue("Status"))
	dataView.Set("Message", k.Request.FormValue("Message"))

	return dataView
}
func (c *OvertimeController) HandleDeclineNote(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}
	err := k.GetPayload(&param)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	paramRes, err := new(services.OvertimeService).ValidateParamMail(param, "")
	if err != nil {
		return c.SetResultError(err.Error(), paramRes)
	}

	return c.SetResultOK(paramRes)
}
func (c *OvertimeController) HandleDeclineCancel(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}

	return c.SetResultOK(param)
}

type dataUserOvertime struct {
	Name        string
	Project     string
	Date        []string
	Purpose     string
	DayDuration int
	URL         string
}

func (c *OvertimeController) SendUsersOvertime(k *knot.WebContext, to []string, userid string, idovertime string, name string, project string, p OvertimeModel) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	urlConf := helper.ReadConfig()
	ret := ResultInfo{}
	datauser := dataUserOvertime{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", project+" - Overtime Duty")
	m := gomail.NewMessage()

	sUrl := urlConf.GetString("BaseUrlEmail")
	uriApprove := sUrl + "/overtime/userovertime"
	param := new(ParameterURLUserModel)
	param.Name = name
	param.IdOvertime = idovertime
	param.Project = project
	param.UserId = userid
	param.Result = "yes"
	paramApp, _ := json.Marshal(param)

	urlApprove, _ := http.NewRequest("GET", uriApprove, nil)
	urlA := urlApprove.URL.Query()
	urlA.Add("param", GCMEncrypter(string(paramApp)))
	urlApprove.URL.RawQuery = urlA.Encode()

	// fmt.Println("------------ leader masuk send manager", dec.UserId)s
	dayString := []string{}
	for _, d := range p.DayList {
		if d.Result != "Declined" {
			dy, _ := time.Parse("2006-01-02", d.Date)
			dayString = append(dayString, dy.Format("02-01-2006"))
		}

	}
	datauser.Name = name
	datauser.Project = project
	datauser.Date = dayString
	datauser.Purpose = p.Reason
	datauser.DayDuration = p.DayDuration
	datauser.URL = urlApprove.URL.String()

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileOvertimeTemplate("overtimeUser.html", datauser)

	if er != nil {
		tk.Println("----------- send email error ", er.Error())
		return c.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""

}

func (c *OvertimeController) SendUserInfoOvertime(k *knot.WebContext, userid string, idovertime string, name string, project string, p OvertimeModel) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	ret := ResultInfo{}
	datamanager := dataManagerOvertime{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", project+" - Info assigned overtime Duty")
	m := gomail.NewMessage()

	dev := []string{}
	for _, nm := range p.MembersOvertime {
		dev = append(dev, nm.Name)
	}

	dayString := []string{}
	for _, d := range p.DayList {
		dy, _ := time.Parse("2006-01-02", d.Date)
		dayString = append(dayString, dy.Format("02-01-2006"))
	}
	to := []string{}
	for _, usr := range p.MembersOvertime {
		to = append(to, usr.Email)
	}
	datamanager.ManagerName = name
	datamanager.Project = project
	datamanager.Date = dayString
	datamanager.Purpose = p.Reason
	datamanager.DayDuration = p.DayDuration
	datamanager.Developer = dev
	datamanager.LeaderName = p.Name

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileManagerTemplate("overtimeinfouser.html", datamanager)

	if er != nil {
		tk.Println("----------- send email error ", er.Error())
		return c.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""

}

func (c *OvertimeController) SendMailDeclineApproved(k *knot.WebContext, to []string, userid string, idovertime string, name string, project string, p OvertimeModel) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	// urlConf := helper.ReadConfig()
	ret := ResultInfo{}
	datamanager := dataManagerOvertime{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", project+" - Overtime Duty")
	if p.ResultRequest == "Declined" {
		mailsubj = tk.Sprintf("%v", "Rejection Notice from Manager")
	} else if p.ResultRequest == "Approved" {
		mailsubj = tk.Sprintf("%v", "Approval Notice from Manager")
		if name == "INA HRD members" {
			mailsubj = tk.Sprintf("%v", "Overtime Notification Project "+project)
		}
	}
	m := gomail.NewMessage()

	dev := []string{}
	for _, nm := range p.MembersOvertime {
		dev = append(dev, nm.Name)
	}

	dayString := []string{}
	dayDecline := []string{}
	// reasonDecline := []string{}

	for _, d := range p.DayList {
		if d.Result != "Declined" {
			dy, _ := time.Parse("2006-01-02", d.Date)
			dayString = append(dayString, dy.Format("02-01-2006"))
		} else {
			dy, _ := time.Parse("2006-01-02", d.Date)
			dayDecline = append(dayDecline, dy.Format("02-01-2006")+" - "+d.Reason)
			// reasonDecline = append(reasonDecline, )

		}

	}

	datamanager.ManagerName = name
	datamanager.Project = project
	datamanager.Date = dayString
	datamanager.Purpose = p.Reason
	datamanager.DayDuration = p.DayDuration
	datamanager.ManagerApprove = p.ApprovalManager.Name
	datamanager.URLApprove = ""
	datamanager.Reason = p.ApprovalManager.Reason
	datamanager.Developer = dev
	datamanager.URLDecline = ""
	datamanager.DateRequest = p.DateCreated
	datamanager.DateDeclined = dayDecline
	// datamanager.ReasonDeclined = reasonDecline

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	if p.ResultRequest == "Declined" {
		bd, er := FileManagerTemplate("overtimedeclined.html", datamanager)

		if er != nil {
			tk.Println("----------- send email error ", er.Error())
			return c.SetResultInfo(true, er.Error(), nil)
		}

		m.SetBody("text/html", string(bd))

	} else if p.ResultRequest == "Approved" {
		if name == "INA HRD members" {
			bd, er := FileManagerTemplate("overtimeapprovedhrd.html", datamanager)

			if er != nil {
				tk.Println("----------- send email error ", er.Error())
				return c.SetResultInfo(true, er.Error(), nil)
			}

			m.SetBody("text/html", string(bd))
		} else {
			bd, er := FileManagerTemplate("overtimeapproved.html", datamanager)

			if er != nil {
				tk.Println("----------- send email error ", er.Error())
				return c.SetResultInfo(true, er.Error(), nil)
			}

			m.SetBody("text/html", string(bd))
		}

	}

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}

	m.Reset()

	return ""

}

func (c *OvertimeController) SendManagerOvertime(k *knot.WebContext, to []string, userid string, idovertime string, name string, project string, p OvertimeModel) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	urlConf := helper.ReadConfig()
	ret := ResultInfo{}
	datamanager := dataManagerOvertime{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", "Request Overtime Project "+project)
	m := gomail.NewMessage()

	sUrl := urlConf.GetString("BaseUrlEmail")
	uriApprove := sUrl + "/overtime/approvedovertime"
	param := new(ParameterURLUserModel)
	param.Name = name
	param.IdOvertime = idovertime
	param.Project = project
	param.UserId = userid
	param.Result = "yes"
	paramApp, _ := json.Marshal(param)

	urlApprove, _ := http.NewRequest("GET", uriApprove, nil)
	urlA := urlApprove.URL.Query()
	urlA.Add("param", GCMEncrypter(string(paramApp)))
	urlApprove.URL.RawQuery = urlA.Encode()

	uriDecline := sUrl + "/overtime/declinedovertime"

	param.Result = "no"
	paramDec, _ := json.Marshal(param)

	urimgrpage := sUrl + "/overtime/managerovertime"

	urlmgrpage, _ := http.NewRequest("GET", urimgrpage, nil)
	urlpage := urlmgrpage.URL.Query()
	urlpage.Add("param", GCMEncrypter(string(paramApp)))
	urlmgrpage.URL.RawQuery = urlpage.Encode()

	// fmt.Println("------------ leader masuk send manager", dec.UserId)

	urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
	qDecline := urlDecline.URL.Query()
	qDecline.Add("param", GCMEncrypter(string(paramDec)))
	urlDecline.URL.RawQuery = qDecline.Encode()

	dev := []string{}
	for _, nm := range p.MembersOvertime {
		dev = append(dev, nm.Name)
	}

	dayString := []string{}
	for _, d := range p.DayList {
		dy, _ := time.Parse("2006-01-02", d.Date)
		dayString = append(dayString, dy.Format("02-01-2006"))
	}
	datamanager.ManagerName = name
	datamanager.Project = project
	datamanager.Date = dayString
	datamanager.Purpose = p.Reason
	datamanager.DayDuration = p.DayDuration
	datamanager.URLApprove = urlApprove.URL.String()
	datamanager.Developer = dev
	datamanager.URLDecline = urlDecline.URL.String()
	datamanager.URLPage = urlmgrpage.URL.String()
	datamanager.LeaderName = p.Name

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileManagerTemplate("overtimemanager.html", datamanager)

	if er != nil {
		tk.Println("----------- send email error ", er.Error())
		return c.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""

}

func FileOvertimeTemplate(filename string, data dataUserOvertime) ([]byte, error) {
	fmt.Println("------ masuk file")
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
	if err != nil {
		return body, err
	}
	// fmt.Println("------ masuk data ", data.Date)
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		return body, err
	}
	body = buffer.Bytes()

	return body, nil
}

type dataManagerOvertime struct {
	ManagerName    string
	LeaderName     string
	Project        string
	Date           []string
	Purpose        string
	DayDuration    int
	Developer      []string
	Reason         string
	ManagerApprove string
	URLApprove     string
	URLDecline     string
	DateRequest    string
	URLPage        string
	// DeclineState   []stateDecline
	DateDeclined   []string
	ReasonDeclined []string
}

type stateDecline struct {
	DateDeclined   string
	ReasonDeclined string
}

func FileManagerTemplate(filename string, data dataManagerOvertime) ([]byte, error) {
	fmt.Println("------ masuk file")
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
	if err != nil {
		return body, err
	}
	// fmt.Println("------ masuk data ", data.Date)
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		return body, err
	}
	body = buffer.Bytes()

	return body, nil
}

type dataPartialyApprove struct {
	LeaderName     string
	Project        string
	Date           []string
	Purpose        string
	DayDuration    int
	ReasonDeclined string
	ManagerApprove string
	DateRequest    string
	MemberApproved []string
	MemberDeclined []string
}

func PartiallyApprovedTemplate(filename string, data dataPartialyApprove) ([]byte, error) {
	// fmt.Println("------ masuk file")
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
	if err != nil {
		return body, err
	}
	// fmt.Println("------ masuk data ", data.Date)
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		return body, err
	}
	body = buffer.Bytes()

	return body, nil
}

func (c *OvertimeController) SendMailPartiallyApproved(k *knot.WebContext, to []string, leadername string, memberApproved []string, memberDeclined []string, project string, p OvertimeModel) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	// urlConf := helper.ReadConfig()
	ret := ResultInfo{}
	data := dataPartialyApprove{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", "Overtime announcement for "+project)
	m := gomail.NewMessage()

	dev := []string{}
	for _, nm := range p.MembersOvertime {
		dev = append(dev, nm.Name)
	}

	dayString := []string{}
	for _, d := range p.DayList {
		dy, _ := time.Parse("2006-01-02", d.Date)
		dayString = append(dayString, dy.Format("02-01-2006"))
	}

	data.LeaderName = leadername
	data.Project = project
	data.Date = dayString
	data.Purpose = p.Reason
	data.DayDuration = p.DayDuration
	data.ManagerApprove = p.ApprovalManager.Name
	data.ReasonDeclined = p.DeclineReason
	data.MemberApproved = memberApproved
	data.MemberDeclined = memberDeclined
	data.DateRequest = p.DateCreated

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := PartiallyApprovedTemplate("overtimedeclinedpartiallyleader.html", data)

	if er != nil {
		tk.Println("----------- send email error ", er.Error())
		return c.SetResultInfo(true, er.Error(), nil)
	}

	m.SetBody("text/html", string(bd))

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}

	m.Reset()

	return ""

}

func (c *OvertimeController) VerificationOvertimeUser(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	// dtDetails := struct {
	// 	Date  string
	// 	Start string
	// 	End   string
	// }{}
	// p := struct {
	// 	Param        string
	// 	Typeovertime string
	// 	Hours        int
	// 	DateDetails  []dtDetails
	// }{}
	p := DetailPayloadInput{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	// tk.Println("---------- ", tk.JsonString(p))

	getp := new(ParameterURLUserModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), getp)
	//tk.Println("thisssss.....", getp)
	pipe := []tk.M{}

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": getp.IdOvertime}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}

	getOvertime := []tk.M{}
	setOvertime := tk.M{}
	if err := csr.Fetch(&getOvertime, 0, false); err != nil {
		return nil
	}
	dateNow := time.Now()
	isconfirmed := false
	for _, each := range getOvertime {
		getmember := each.Get("membersovertime").([]interface{})
		setOvertime = each
		arrOvertime := []tk.M{}
		for _, gm := range getmember {
			ss := tk.M{}
			ss, err = tk.ToM(gm)
			if err != nil {
				return err
			}

			ovr, err := c.DateEmployeebyId(r, getp.IdOvertime, getp.UserId)
			datstr := []string{}
			for _, dateIn := range ovr {
				datstr = append(datstr, dateIn.DateClosed)
			}
			sort.Strings(datstr)

			lastIndex := len(ovr) - 1
			datestr, _ := time.Parse("2006-01-02", datstr[0])
			datend, _ := time.Parse("2006-01-02", datstr[lastIndex])
			// dtend := datend.AddDate(0, 0, 2)

			dateEndForm := datend.Format("2006-01-02")
			dateNowForm := dateNow.Format("2006-01-02")

			tk.Println("datestart ovr ======== ", datestr)
			tk.Println("datenow ovr ======== ", dateNowForm)
			if ss.Get("userid") == getp.UserId {
				//tk.Println("dataovertime...", ss.Get("userid"))
				if ss.Get("result") != "Confirmed" {
					ss.Set("typeovertime", p.Typeovertime)
					ss.Set("hours", p.Hours)
					if dateEndForm == dateNowForm {
						isconfirmed = true
						ss.Set("result", "Confirmed")
					}

				}
				//tk.Println("save....", ss)
				arrOvertime = append(arrOvertime, ss)
			} else {
				arrOvertime = append(arrOvertime, ss)
			}
			// // dtYesterday := dateNow.AddDate(0, 0, -1)
			// dtnow := dateNow.Format("2006-01-02")
			// dtn, _ := time.Parse("2006-01-02T15:04:05.000Z", dtnow)

			for _, over := range ovr {
				// if over.Hours == 0 {
				if dateNow.Year() == datestr.Year() && int(dateNow.Month()) == int(datend.Month()) && dateNow.Day() <= datend.Day() {
					// tk.Println("masuk sini 123 ======== ", dateNow)
					over.Hours = p.Hours
					over.TypeOvertime = p.Typeovertime

					for _, dtDate := range p.DateDetails {
						// tk.Println("---------- date ", dtDate.Task)
						// tk.Println("---------- date ", dtDate.DeadLine)

						if dtDate.Date == over.DateOvertime {

							over.TimeStart = dtDate.Start
							over.TimeEnd = dtDate.End
							over.TypeOvertime = dtDate.Type
							over.Task = dtDate.Task
							over.Deadline = dtDate.DeadLine
						}
					}

					err = c.Ctx.Save(&over)
					if err != nil {
						return c.SetResultInfo(true, err.Error(), nil)
					}
				} else {
					if dateNow.Year() == datestr.Year() && int(dateNow.Month()) == int(datend.Month()) && dateNow.Day() >= datend.Day() {
						return c.SetResultInfo(true, "your link Confimation has been expired ", nil)
					} else {
						// tk.Println("masuk sini ======== ", dateNow)
						over.Hours = p.Hours
						over.TypeOvertime = p.Typeovertime

						for _, dtDate := range p.DateDetails {
							// tk.Println("---------- date ", dtDate.Task)
							// tk.Println("---------- date ", dtDate.DeadLine)

							if dtDate.Date == over.DateOvertime {

								over.TimeStart = dtDate.Start
								over.TimeEnd = dtDate.End
								over.TypeOvertime = dtDate.Type
								over.Task = dtDate.Task
								over.Deadline = dtDate.DeadLine
							}
						}

						err = c.Ctx.Save(&over)
						if err != nil {
							return c.SetResultInfo(true, err.Error(), nil)
						}
					}

				}
				// }
			}

		}
		setOvertime.Set("membersovertime", arrOvertime)

	}

	// tk.Println("saveall....", setOvertime)

	q := c.Ctx.Connection.NewQuery().From("NewOvertime").SetConfig("multiexec", true).Save()
	defer q.Close()
	newdata := map[string]interface{}{"data": setOvertime}
	err = q.Exec(newdata)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	dataConf := tk.M{}
	dataConf.Set("IsConfirmed", isconfirmed)

	return c.SetResultInfo(false, "succesfully saved", dataConf)

}
func (c *OvertimeController) GetdataDev(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		Param string
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	getp := new(ParameterURLUserModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), getp)
	objid := getp.IdOvertime
	pipe := []tk.M{}

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": objid}},
				},
			},
		},
	)
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
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

func (c *OvertimeController) GetdataOvertime(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$ne": ""}},
				},
			},
		},
	)
	pipe = append(pipe,
		tk.M{
			"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": true},
		},
	)

	pipe = append(pipe, tk.M{
		"$project": tk.M{
			"_id":             1,
			"projectleader":   1,
			"project":         1,
			"resultrequest":   1,
			"approvalmanager": 1,
			"daylist":         1,
			"membersovertime": 1,
			"projleader":      "$name",
			"approvalman":     "$approvalmanager.name",
		},
	})

	pipe = append(pipe, tk.M{}.Set("$sort", tk.M{}.Set("daylist.date", -1)))
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
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

func (c *OvertimeController) GetdataOvertimePending(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$ne": ""}},
				},
			},
		},
	)
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{"resultrequest": tk.M{"$eq": "Pending"}},
		},
	)
	// pipe = append(pipe,
	// 	tk.M{
	// 		"$unwind": "$daylist",
	// 	},
	// )

	// pipe = append(pipe,
	// 	tk.M{
	// 		"$match": tk.M{"daylist.result": "Pending"},
	// 	},
	// )

	pipe = append(pipe, tk.M{
		"$project": tk.M{
			"_id":             1,
			"projectleader":   1,
			"project":         1,
			"resultrequest":   1,
			"approvalmanager": 1,
			"daylist":         1,
			"membersovertime": 1,
			"reason":          1,
			"projleader":      "$name",
			"approvalman":     "$approvalmanager.name",
		},
	})

	// pipe = append(pipe, tk.M{}.Set("$find", tk.M{}.Set("daylist.date", "Pending")))

	pipe = append(pipe, tk.M{}.Set("$sort", tk.M{}.Set("daylist.date", -1)))
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}
	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil
	}

	tk.Println("-------- pipe ", tk.JsonString(data))
	return data
}

func (c *OvertimeController) GetdataOvertimeExpired(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	today := time.Now()
	yesterday := today.AddDate(0, 0, -2)
	tk.Println("------------------------ yesterday ", yesterday)
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$ne": ""}},
				},
			},
		},
	)
	// pipe = append(pipe, tk.M{"$addFields": tk.M{"datecreated": tk.M{"$dateFromString": tk.M{"dateString": "$datecreated"}}}})
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{"resultrequest": tk.M{"$eq": "Expired"}},
		},
	)
	// pipe = append(pipe,
	// 	tk.M{
	// 		"$match": tk.M{"datecreated": tk.M{"$gt": yesterday, "$lte": today}},
	// 	},
	// )
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{"isdelete": false},
		},
	)

	// tk.M{"$addFields": tk.M{"datecreated": tk.M{"$dateFromString": tk.M{"dateString": "$datecreated"},},},},
	// pipe = append(pipe,
	// 	tk.M{
	// 		"$unwind": "$daylist",
	// 	},
	// )

	// pipe = append(pipe,
	// 	tk.M{
	// 		"$match": tk.M{"daylist.result": "Pending"},
	// 	},
	// )

	pipe = append(pipe, tk.M{
		"$project": tk.M{
			"_id":             1,
			"projectleader":   1,
			"project":         1,
			"resultrequest":   1,
			"approvalmanager": 1,
			"daylist":         1,
			"membersovertime": 1,
			"reason":          1,
			"datecreated":     1,
			"projleader":      "$name",
			"approvalman":     "$approvalmanager.name",
		},
	})

	// pipe = append(pipe, tk.M{}.Set("$find", tk.M{}.Set("daylist.date", "Pending")))

	pipe = append(pipe, tk.M{}.Set("$sort", tk.M{}.Set("daylist.date", -1)))
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
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

func (c *OvertimeController) VerificationOvertimeManager(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		Param   string
		DataDev []tk.M
		Decline string
		ContDev float64
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	getp := new(ParameterURLUserModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), getp)
	var getcount string
	if p.ContDev >= 1 {
		getcount = "Approved"
	} else {
		getcount = "Declined"
	}

	pipe := []tk.M{}
	//var objid string
	objid := getp.IdOvertime
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": objid}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}

	getOvertime := []tk.M{}
	setOvertime := tk.M{}
	managerName := ""
	emailManager := ""
	if err := csr.Fetch(&getOvertime, 0, false); err != nil {
		return nil
	}

	for _, each := range getOvertime {
		setOvertime = each
		getManager := each.Get("approvalmanager")
		setM := tk.M{}
		setM, err = tk.ToM(getManager)
		if err != nil {
			return err
		}

		userid := getp.UserId

		useridBranchManager := each.Get("branchmanagers")
		useridProjectManager := each.Get("projectmanager")
		getuseridBranchManager := ""
		getuseridProjectManager := ""
		gbmSave := tk.M{}

		getubm := useridBranchManager.([]interface{})

		for _, uibm := range getubm {
			tuibm := uibm.(tk.M)
			getuidf := tuibm.GetString("userid")
			if getuidf == userid {
				getuseridBranchManager = getuidf
				gbmSave = tuibm
			}
			// getuseridBranchManager = getuidf
			// gbmSave = tuibm
		}
		//tk.Println("iiii...", getuseridBranchManager)
		getupm := useridProjectManager.(tk.M)
		getuseridProjectManager = getupm.GetString("userid")
		//tk.Println("sss...", getuseridProjectManager)

		if userid == getuseridBranchManager {
			setM.Set("email", gbmSave.GetString("email"))
			setM.Set("idemp", gbmSave.GetString("idemp"))
			setM.Set("name", gbmSave.GetString("name"))
			setM.Set("phonenumber", gbmSave.GetString("phonenumber"))
			setM.Set("location", gbmSave.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", gbmSave.GetString("userid"))
			managerName = gbmSave.GetString("name")
			emailManager = gbmSave.GetString("email")
		} else if userid == getuseridProjectManager {
			setM.Set("email", getupm.GetString("email"))
			setM.Set("idemp", getupm.GetString("idemp"))
			setM.Set("name", getupm.GetString("name"))
			setM.Set("phonenumber", getupm.GetString("phonenumber"))
			setM.Set("location", getupm.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", getupm.GetString("userid"))
			managerName = getupm.GetString("name")
			emailManager = getupm.GetString("email")
		}

		setOvertime.Set("resultrequest", getcount)
		setOvertime.Set("declinereason", p.Decline)
		setOvertime.Set("approvalmanager", setM)
		setOvertime.Set("id", each.Get("_id"))
	}

	approval, _ := tk.ToM(setOvertime.Get("approvalmanager").(interface{}))
	setOvertime.Set("membersovertime", p.DataDev)
	getdaylist := []tk.M{}
	fgo := setOvertime.Get("daylist").([]interface{})
	userlistDecline := []string{}
	count := 0
	for _, dy := range fgo {
		dyt := dy.(tk.M)
		dyt.Set("result", getcount)
		dyt.Set("reason", p.Decline)
		getdaylist = append(getdaylist, dyt)
		for _, eo := range p.DataDev {
			getresult := eo.GetString("result")
			if getresult == "Approved" {
				//tk.Println("masuukkkk", getresult)
				idovertime := getp.IdOvertime
				idemployee := eo.GetString("idemp")
				project := getp.Project
				userid := eo.GetString("userid")
				name := eo.GetString("name")
				location := eo.GetString("location")
				email := eo.GetString("email")
				phonenumber := eo.GetString("phonenumber")
				dd := dy.(tk.M)
				date := dd.GetString("date")
				tipe := eo.GetString("typeovertime")
				hour := eo.GetInt("hours")
				c.DateEmployeeOvertime(r, idovertime, idemployee, project, userid, name, location, email, phonenumber, date, tipe, hour)
			} else {
				if count == 0 {
					userlistDecline = append(userlistDecline, eo.GetString("name"))
				}
			}
		}
		count++
	}
	setOvertime.Set("daylist", getdaylist)
	//tk.Println("dev....", setOvertime.Get("daylist"), reflect.TypeOf(setOvertime.Get("daylist")))

	qs := c.Ctx.Connection.NewQuery().From("NewOvertime").SetConfig("multiexec", true).Save()
	defer qs.Close()
	newdata := map[string]interface{}{"data": setOvertime}
	err = qs.Exec(newdata)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	c.SendMailApprovePartially(r, getp.IdOvertime)

	// add log overtime
	data := new(OvertimeModel)
	if err = tk.MtoStruct(setOvertime, &data); err != nil {
		return nil
	}
	service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: data}

	log := tk.M{}
	log.Set("Status", getcount)
	if getcount == "Approved" {
		log.Set("Desc", "Request Approved by Manager")
		if len(userlistDecline) > 0 {
			log.Set("Desc", "Request Approved by Manager with decline user list : "+strings.Join(userlistDecline, ", "))
		}
	} else {
		log.Set("Desc", "Request Declined by Manager")
	}
	log.Set("NameLogBy", managerName)
	log.Set("EmailNameLogBy", emailManager)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		c.SetResultInfo(true, "Error occured in overtime when overtime approved partially", nil)
	}
	//

	return c.SetResultInfo(false, "Success", approval)
}

func (c *OvertimeController) AddYearleaveUserOvertime(r *knot.WebContext, userid string) interface{} {
	dash := DashboardController(*c)
	data := dash.GetDataSessionUser(r, userid)

	if data[0].DecYear == 0 && data[0].YearLeave > 0 {
		data[0].DecYear = float64(data[0].YearLeave)
	}

	decYear := data[0].DecYear + 1.0

	data[0].DecYear = decYear
	data[0].YearLeave = int(decYear)

	err := c.Ctx.Save(data[0])
	if err != nil {
		c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "Success", nil)
}

func (c *OvertimeController) VerificationOvertimeManagerdb(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		ID      string
		DataDev []tk.M
		Decline string
		ContDev float64
		UserId  string
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	var getcount string
	if p.ContDev >= 1 {
		getcount = "Approved"
	} else {
		getcount = "Declined"
	}

	pipe := []tk.M{}
	var objid string
	objid = p.ID
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": objid}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}

	getOvertime := []tk.M{}
	setOvertime := tk.M{}
	managerName := ""
	emailManager := ""
	if err := csr.Fetch(&getOvertime, 0, false); err != nil {
		return nil
	}

	tmpOvertimeId := ""
	tmpOvertimeUserId := ""
	tmpOvertimeDate := ""

	for _, each := range getOvertime {

		tmpOvertimeId = each.GetString("_id")
		for _, m := range p.DataDev {
			tmpOvertimeUserId = m.GetString("userid")
		}
		do := each.Get("daylist").([]interface{})
		for _, d := range do {
			d1 := d.(tk.M)
			tmpOvertimeDate = d1.GetString("date")
		}

		setOvertime = each
		getManager := each.Get("approvalmanager")
		setM := tk.M{}
		setM, err = tk.ToM(getManager)
		if err != nil {
			return err
		}

		userid := p.UserId
		useridBranchManager := each.Get("branchmanagers")
		useridProjectManager := each.Get("projectmanager")
		getuseridBranchManager := ""
		getuseridProjectManager := ""
		gbmSave := tk.M{}

		getubm := useridBranchManager.([]interface{})

		for _, uibm := range getubm {
			tuibm := uibm.(tk.M)
			getuidf := tuibm.GetString("userid")
			if getuidf == userid {
				getuseridBranchManager = getuidf
				gbmSave = tuibm
			}
		}
		//tk.Println("iiii...", getuseridBranchManager)
		getupm := useridProjectManager.(tk.M)
		getuseridProjectManager = getupm.GetString("userid")
		//tk.Println("sss...", getuseridProjectManager)

		if userid == getuseridBranchManager {
			setM.Set("email", gbmSave.GetString("email"))
			setM.Set("idemp", gbmSave.GetString("idemp"))
			setM.Set("name", gbmSave.GetString("name"))
			setM.Set("phonenumber", gbmSave.GetString("phonenumber"))
			setM.Set("location", gbmSave.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", gbmSave.GetString("userid"))
			managerName = gbmSave.GetString("name")
			emailManager = gbmSave.GetString("email")
		} else if userid == getuseridProjectManager {
			setM.Set("email", getupm.GetString("email"))
			setM.Set("idemp", getupm.GetString("idemp"))
			setM.Set("name", getupm.GetString("name"))
			setM.Set("phonenumber", getupm.GetString("phonenumber"))
			setM.Set("location", getupm.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", getupm.GetString("userid"))
			managerName = getupm.GetString("name")
			emailManager = getupm.GetString("email")
		}

		//setM.Set("result ", getcount)
		setOvertime.Set("resultrequest", getcount)
		setOvertime.Set("declinereason", p.Decline)
		setOvertime.Set("approvalmanager", setM)
		setOvertime.Set("id", each.Get("_id"))
	}

	setOvertime.Set("membersovertime", p.DataDev)
	//tk.Println("dev....", setOvertime)
	getdaylist := []tk.M{}
	fgo := setOvertime.Get("daylist").([]interface{})
	userlistDecline := []string{}
	count := 0

	tk.Println("Data temporry ===========================>", tmpOvertimeId, tmpOvertimeUserId, tmpOvertimeDate)
	if c.VerifyIsAlreadyHasRequestOvertime(tmpOvertimeId, tmpOvertimeUserId, tmpOvertimeDate) {

	} else {

		for _, dy := range fgo {
			dyt := dy.(tk.M)
			dyt.Set("result", getcount)
			dyt.Set("reason", p.Decline)
			getdaylist = append(getdaylist, dyt)
			for _, eo := range p.DataDev {
				getresult := eo.GetString("result")
				if getresult == "Approved" {
					//tk.Println("masuukkkk", getresult)
					idovertime := setOvertime.GetString("_id")
					idemployee := eo.GetString("idemp")
					project := setOvertime.GetString("project")
					userid := eo.GetString("userid")
					name := eo.GetString("name")
					location := eo.GetString("location")
					email := eo.GetString("email")
					phonenumber := eo.GetString("phonenumber")
					dd := dy.(tk.M)
					date := dd.GetString("date")
					tipe := eo.GetString("typeovertime")
					hour := eo.GetInt("hours")
					c.DateEmployeeOvertime(r, idovertime, idemployee, project, userid, name, location, email, phonenumber, date, tipe, hour)
				} else {
					if count == 0 {
						userlistDecline = append(userlistDecline, eo.GetString("name"))
					}
				}
			}
			count++
		}
		setOvertime.Set("daylist", getdaylist)
		qs := c.Ctx.Connection.NewQuery().From("NewOvertime").SetConfig("multiexec", true).Save()
		defer qs.Close()
		newdata := map[string]interface{}{"data": setOvertime}
		err = qs.Exec(newdata)

		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		c.SendMailApprovePartially(r, p.ID)

		// add log overtime
		data := new(OvertimeModel)
		if err = tk.MtoStruct(setOvertime, &data); err != nil {
			return nil
		}
		service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: data}

		log := tk.M{}
		log.Set("Status", getcount)
		if getcount == "Approved" {
			log.Set("Desc", "Request Approved by Manager")
			if len(userlistDecline) > 0 {
				log.Set("Desc", "Request Approved by Manager with decline user list : "+strings.Join(userlistDecline, ", "))
			}
		} else {
			log.Set("Desc", "Request Declined by Manager")
		}
		log.Set("NameLogBy", managerName)
		log.Set("EmailNameLogBy", emailManager)
		err = service.ApproveDeclineLog(log)
		if err != nil {
			c.SetResultInfo(true, "Error occured in overtime when overtime approved partially", nil)
		}
		//

		notif := NotificationController(*c)
		getnotif := notif.GetDataNotification(r, data.Id)
		getnotif.Notif.ManagerApprove = data.ApprovalManager.Name
		getnotif.Notif.Status = data.ResultRequest
		getnotif.Notif.StatusApproval = data.ApprovalManager.Result
		if getcount != "Approved" {
			getnotif.Notif.Description = p.Decline
		}
		notif.InsertNotification(getnotif)

	}

	return c.SetResultInfo(false, "Success", nil)
}

func (c *OvertimeController) SendMailApprovePartially(r *knot.WebContext, idovertime string) interface{} {
	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")

	usrDecline := []string{}
	usrApproved := []string{}
	data, err := c.GetOvertimeByID(r, idovertime)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	for _, dec := range data.MembersOvertime {
		if dec.Result == "Approved" {
			c.SendUsersOvertime(r, []string{dec.Email}, dec.UserId, idovertime, dec.Name, data.Project, data)
			usrApproved = append(usrApproved, dec.Name)
		} else if dec.Result == "Declined" {
			usrDecline = append(usrDecline, dec.Name)
		}
	}

	c.SendMailPartiallyApproved(r, []string{hrd}, "INA HRD members", usrApproved, usrDecline, data.Project, data) //--------- hrd

	c.SendMailPartiallyApproved(r, []string{data.ProjectLeader.Email}, data.ProjectLeader.Name, usrApproved, usrDecline, data.Project, data) //--------- leader
	return ""

}

func (c *OvertimeController) DateEmployeeOvertime(r *knot.WebContext, idovertime string, idemployee string, project string, userid string, name string, location string, email string, phonenumber string, date string, tipe string, hour int) interface{} {
	c.LoadBase(r)
	r.Config.OutputType = knot.OutputJson
	p := EmployeeOvertimeModel{}

	tk.Println("--------------- idemp ", idemployee)
	p.Id = bson.NewObjectId().Hex()
	p.IdOvertime = idovertime
	p.UserId = userid
	p.IdEmployee = idemployee
	p.Name = name
	p.Location = location
	p.Email = email
	p.PhoneNumber = phonenumber
	p.DateOvertime = date
	p.TypeOvertime = tipe
	p.Hours = hour
	p.TrackHour = 0
	p.ResultMatch = "empty"
	p.DateAdminCheck = ""
	p.IsCheck = false
	p.Project = project
	dtutc, _ := time.Parse("2006-01-02", date)
	clsDate := dtutc.AddDate(0, 0, 2)
	p.DateClosed = clsDate.Format("2006-01-02")
	now := time.Now()
	p.DateApprove = now.Format("2006-01-02")

	dateOvertime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	tk.Println("--------------- day ", dateOvertime.Day())

	p.Day = dateOvertime.Day()
	p.Month = int(dateOvertime.Month())
	p.Year = dateOvertime.Year()

	// err = c.Ctx.Save(&p)
	if c.VerifyIsAlreadyHasRequestOvertime(idovertime, userid, date) {
		err = nil
	} else {
		err = c.Ctx.Save(&p)
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "success", nil)
}

func (c *OvertimeController) DateEmployeebyId(r *knot.WebContext, idovertime string, userid string) ([]EmployeeOvertimeModel, error) {
	res := []EmployeeOvertimeModel{}

	var dbFilter []*db.Filter
	query := tk.M{}
	// sort := "-" + sort
	dbFilter = append(dbFilter, db.Eq("idovertime", idovertime))
	dbFilter = append(dbFilter, db.Eq("userid", userid))

	if len(dbFilter) > 0 {
		// query.Set("order", []string{sort})
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return res, err
	}

	if err != nil {
		return res, err
	}

	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return res, err
	}
	return res, err
}

func (c *OvertimeController) UnCheckOvertime(r *knot.WebContext) ([]EmployeeOvertimeModel, error) {
	r.Config.OutputType = knot.OutputJson
	res := []EmployeeOvertimeModel{}

	var dbFilter []*db.Filter
	query := tk.M{}

	dbFilter = append(dbFilter, db.Eq("ischeck", false))
	dbFilter = append(dbFilter, db.Gt("hours", 0))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return res, err
	}

	if err != nil {
		return res, err
	}

	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return res, err
	}

	data := []EmployeeOvertimeModel{}

	for _, rs := range res {
		dateOvertime, err := time.Parse("2006-01-02", rs.DateOvertime)
		if err != nil {
			return res, err
		}
		dayOvertime := dateOvertime.Day()

		now := time.Now()
		dayNow := now.Day()

		if dayNow > dayOvertime {
			data = append(data, rs)
		}
	}

	// tk.Println("------------- data ", data)

	return res, nil

}

func (c *OvertimeController) GetOvertimeNotif(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	tm := time.Now()
	pipe = append(pipe, tk.M{"$match": tk.M{"ischeck": tk.M{"$eq": false}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"day": tk.M{"$ne": tm.Day()}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"hours": tk.M{"$ne": 0}}})
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"project": "$project", "dateovertime": "$dateovertime"}, "date": tk.M{"$push": "$dateovertime"}, "data": tk.M{"$push": "$$ROOT"}}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("EmployeeOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return c.SetResultInfo(true, "bad query", nil)
	}
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	result := []tk.M{}
	for _, each := range data {
		date, _ := time.Parse("2006-01-02", each.GetString("dateovertime"))
		if date.Before(tm) {
			result = append(result, each)
		}
	}

	return c.SetResultInfo(false, "success", len(result))
}

func (c *OvertimeController) GetAdminOvertime(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	tm := time.Now()
	pipe = append(pipe, tk.M{"$match": tk.M{"ischeck": tk.M{"$eq": false}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"day": tk.M{"$ne": tm.Day()}}})
	// pipe = append(pipe, tk.M{"$match": tk.M{"day": tk.M{"$eq": tm.Day()}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"hours": tk.M{"$ne": 0}}})
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"project": "$project", "dateovertime": "$dateovertime"}, "date": tk.M{"$push": "$dateovertime"}, "data": tk.M{"$push": "$$ROOT"}}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("EmployeeOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return c.SetResultInfo(true, "bad query", nil)
	}
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "success", data)
}

func (c *OvertimeController) GetOvertimeAdminDate(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	tm := time.Now()
	// pipe = append(pipe, tk.M{"$match": tk.M{"ischeck": tk.M{"$eq": false}}})
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$or": []tk.M{
					{"trackhour": tk.M{"$ne": 0}},
					tk.M{
						"$and": []tk.M{
							{"hours": tk.M{"$ne": 0}},
							{"ischeck": tk.M{"$eq": false}},
						},
					},
				},
			},
		},
	)
	// pipe = append(pipe, tk.M{"$match": tk.M{"hours": tk.M{"$ne": 0}}})
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"project": "$project", "dateovertime": "$dateovertime", "idovertime": "$idovertime"}, "date": tk.M{"$push": "$dateovertime"}, "data": tk.M{"$push": "$$ROOT"}}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("EmployeeOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	} else {
		return c.SetResultInfo(true, "bad query", nil)
	}
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("--------------- data overtime ", data)

	result := []tk.M{}

	for _, each := range data {
		date, _ := time.Parse("2006-01-02", each.GetString("dateovertime"))
		if date.Before(tm) {
			result = append(result, each)
		}
	}

	return c.SetResultInfo(false, "success", result)
}
func (c *OvertimeController) SaveEmployeeOvertime(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := EmployeeOvertimeModel{}

	err := r.GetPayload(&p)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if p.ResultMatch == "Match" {
		c.AddYearleaveUserOvertime(r, p.UserId)
	}

	err = c.Ctx.Save(&p)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "success", nil)
}

func (c *OvertimeController) GetNationalHoliday(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	data := []NationalHolidaysModel{}
	crs, err := c.Ctx.Find(NewNationalHolidaysModel(), nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if err = crs.Fetch(&data, 0, false); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	now := time.Now()
	// datnow := now.Format("2006-01-02")
	listday := []string{}
	for _, listnat := range data {
		for _, listdat := range listnat.ListDate {
			month := int(listdat.Month())
			dt := listdat.Format("2006-01-02")
			if int(now.Month()) == month {
				listday = append(listday, dt)
			}
		}
	}

	return c.SetResultInfo(false, "success", listday)
}

func (c *OvertimeController) GetdataDev2(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		Param string
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	getp := new(ParameterURLUserModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), getp)
	objid := getp.IdOvertime
	// userid := getp.UserId

	pipe := []tk.M{}

	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": objid}},
				},
			},
		},
		tk.M{
			"$unwind": "$membersovertime",
		},
		tk.M{
			"$match": tk.M{
				"membersovertime.userid": getp.UserId,
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}
	data := []tk.M{}
	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil
	}

	DateEmployee, err := c.DateEmployeebyId(r, getp.IdOvertime, getp.UserId)

	returnData := tk.M{}
	returnData.Set("newovertime", data)
	returnData.Set("employeeovertime", DateEmployee)

	return returnData
}

func (c *OvertimeController) CancelOvertimeManagerdb(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	p := struct {
		ID      string
		Date    string
		Decline string
		UserId  string
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return err.Error
	}

	getcount := "Cancelled"

	pipe := []tk.M{}
	var objid string
	objid = p.ID
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$eq": objid}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("NewOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}

	getOvertime := []tk.M{}
	setOvertime := tk.M{}
	managerName := ""
	emailManager := ""
	if err := csr.Fetch(&getOvertime, 0, false); err != nil {
		return nil
	}

	for _, each := range getOvertime {
		setOvertime = each
		getManager := each.Get("approvalmanager")
		setM := tk.M{}
		setM, err = tk.ToM(getManager)
		if err != nil {
			return err
		}

		userid := p.UserId
		dateOvertime := p.Date
		useridBranchManager := each.Get("branchmanagers")
		useridProjectManager := each.Get("projectmanager")
		daylistOvertime := each.Get("daylist")
		getuseridBranchManager := ""
		getuseridProjectManager := ""
		gbmSave := tk.M{}
		gbmSaveDaylist := []tk.M{}

		getubm := useridBranchManager.([]interface{})
		getdaylist := daylistOvertime.([]interface{})

		for _, uibm := range getubm {
			tuibm := uibm.(tk.M)
			getuidf := tuibm.GetString("userid")
			if getuidf == userid {
				getuseridBranchManager = getuidf
				gbmSave = tuibm
			}
		}

		for _, uiday := range getdaylist {
			tuiday := uiday.(tk.M)
			getuiday := tuiday.GetString("date")
			if getuiday == dateOvertime {
				tuiday.Set("result", getcount)
				tuiday.Set("reason", p.Decline)
			}
			gbmSaveDaylist = append(gbmSaveDaylist, tuiday)
		}
		//tk.Println("iiii...", getuseridBranchManager)
		getupm := useridProjectManager.(tk.M)
		getuseridProjectManager = getupm.GetString("userid")
		//tk.Println("sss...", getuseridProjectManager)

		if userid == getuseridBranchManager {
			setM.Set("email", gbmSave.GetString("email"))
			setM.Set("idemp", gbmSave.GetString("idemp"))
			setM.Set("name", gbmSave.GetString("name"))
			setM.Set("phonenumber", gbmSave.GetString("phonenumber"))
			setM.Set("location", gbmSave.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", gbmSave.GetString("userid"))
			managerName = gbmSave.GetString("name")
			emailManager = gbmSave.GetString("email")
		} else if userid == getuseridProjectManager {
			setM.Set("email", getupm.GetString("email"))
			setM.Set("idemp", getupm.GetString("idemp"))
			setM.Set("name", getupm.GetString("name"))
			setM.Set("phonenumber", getupm.GetString("phonenumber"))
			setM.Set("location", getupm.GetString("location"))
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", getupm.GetString("userid"))
			managerName = getupm.GetString("name")
			emailManager = getupm.GetString("email")
		} else {
			dash := DashboardController(*c)
			user := dash.GetDataSessionUser(r, p.UserId)

			setM.Set("email", user[0].Email)
			setM.Set("idemp", user[0].EmpId)
			setM.Set("name", user[0].Fullname)
			setM.Set("phonenumber", user[0].PhoneNumber)
			setM.Set("location", user[0].Location)
			setM.Set("reason", p.Decline)
			setM.Set("result", getcount)
			setM.Set("userid", p.UserId)
			managerName = user[0].Fullname
			emailManager = user[0].Email
		}

		// setOvertime.Set("resultrequest", getcount)
		// setOvertime.Set("declinereason", p.Decline)
		setOvertime.Set("daylist", gbmSaveDaylist)
		setOvertime.Set("approvalmanager", setM)
		setOvertime.Set("id", each.Get("_id"))
	}

	qs := c.Ctx.Connection.NewQuery().From("NewOvertime").SetConfig("multiexec", true).Save()
	defer qs.Close()
	newdata := map[string]interface{}{"data": setOvertime}
	err = qs.Exec(newdata)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	//update employeeovertime
	pipe = []tk.M{}
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"idovertime": tk.M{"$eq": p.ID}},
					{"dateovertime": tk.M{"$eq": p.Date}},
				},
			},
		},
	)

	csr, err = c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("EmployeeOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil
	}

	getOvertime = []tk.M{}
	setOvertime = tk.M{}
	if err := csr.Fetch(&getOvertime, 0, false); err != nil {
		return nil
	}
	for _, each := range getOvertime {
		setOvertime = each
		setOvertime.Set("resultmatch", getcount)
		setOvertime.Set("ischeck", true)

		qs := c.Ctx.Connection.NewQuery().From("EmployeeOvertime").SetConfig("multiexec", true).Save()
		defer qs.Close()
		newdata := map[string]interface{}{"data": setOvertime}
		err = qs.Exec(newdata)

		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
	}
	//update employeeovertime

	c.SendMailCancelled(r, p.ID, p.Date)

	// add log overtime
	data := new(OvertimeModel)
	if err = tk.MtoStruct(setOvertime, &data); err != nil {
		return nil
	}
	service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: data}

	log := tk.M{}
	log.Set("Status", getcount)
	log.Set("Desc", "Request Cancelled by Manager")
	log.Set("NameLogBy", managerName)
	log.Set("EmailNameLogBy", emailManager)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		c.SetResultInfo(true, "Error occured in overtime when overtime cancelled", nil)
	}
	//

	notif := NotificationController(*c)
	getnotif := notif.GetDataNotification(r, data.Id)
	getnotif.Notif.ManagerApprove = data.ApprovalManager.Name
	getnotif.Notif.Status = data.ApprovalManager.Result
	getnotif.Notif.StatusApproval = data.ApprovalManager.Result
	getnotif.Notif.Description = p.Decline + " on date " + p.Date
	notif.InsertNotification(getnotif)

	return c.SetResultInfo(false, "Success", nil)
}

func (c *OvertimeController) SendMailCancelled(k *knot.WebContext, idovertime string, date string) interface{} {
	// c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")

	p, err := c.GetOvertimeByID(k, idovertime)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	to := []string{hrd, p.ProjectLeader.Email}
	ret := ResultInfo{}
	data := dataPartialyApprove{}
	mail := MailController(*c)
	conf, emailAddress := mail.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", "Overtime announcement for "+p.Project)
	m := gomail.NewMessage()

	dev := []string{}
	for _, nm := range p.MembersOvertime {
		dev = append(dev, nm.Name)
		to = append(to, nm.Email)
	}

	dayString := []string{}
	for _, d := range p.DayList {
		if d.Date == date {
			dayString = append(dayString, d.Date)
		}
	}

	data.LeaderName = p.ProjectLeader.Name
	data.Project = p.Project
	data.Date = dayString
	data.Purpose = p.Reason
	data.DayDuration = p.DayDuration
	data.ManagerApprove = p.ApprovalManager.Name
	data.ReasonDeclined = p.ApprovalManager.Reason
	data.MemberApproved = dev
	data.DateRequest = p.DateCreated

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := PartiallyApprovedTemplate("overtimecancelled.html", data)

	if er != nil {
		tk.Println("----------- send email error ", er.Error())
		return c.SetResultInfo(true, er.Error(), nil)
	}

	m.SetBody("text/html", string(bd))

	tk.Println("----------- send email")

	mail.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}

	m.Reset()

	return ""

}

func (c *OvertimeController) VerifyIsAlreadyHasRequestOvertime(overtimeid string, userid string, overtimedate string) bool {
	res := []EmployeeOvertimeModel{}
	var dbFilter []*db.Filter
	query := tk.M{}
	dbFilter = append(dbFilter, db.Eq("idovertime", overtimeid))
	dbFilter = append(dbFilter, db.Eq("userid", userid))
	dbFilter = append(dbFilter, db.Eq("dateovertime", overtimedate))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	crs, err := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return false
	}
	if err != nil {
		return false
	}
	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return false
	}
	if len(res) > 0 {
		return true
	} else {
		return false
	}
}

func (c *OvertimeController) Getdataemployeeovertime(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Requestids []string
	}{}
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	var rs []interface{}
	for _, v := range p.Requestids {
		rs = append(rs, v)
	}

	res := []EmployeeOvertimeModel{}
	var dbFilter []*db.Filter
	query := tk.M{}
	dbFilter = append(dbFilter, db.In("idovertime", rs...))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	crs, err := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultError(err.Error(), nil)
	}
	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultInfo(false, "success", res)
}

func (c *OvertimeController) Saveverificationtime(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	var newovertimeid string
	payload := struct {
		Timestart string
		Timeend   string
		Task      string
		Deadline  string
		Date      string
		Userid    string
		Hour      int
		Type      string
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	query := tk.M{}
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("dateovertime", payload.Date), db.Eq("userid", payload.Userid))
	employeeovertimes := EmployeeOvertimeModel{}
	query.Set("where", db.And(dbFilter...))
	crs, _ := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	}
	err = crs.Fetch(&employeeovertimes, 1, false)
	if (EmployeeOvertimeModel{}) != employeeovertimes {
		employeeovertimes.TimeStart = payload.Timestart
		employeeovertimes.TimeEnd = payload.Timeend
		employeeovertimes.Task = payload.Task
		employeeovertimes.Deadline = payload.Deadline
		employeeovertimes.Hours = payload.Hour
		employeeovertimes.TypeOvertime = payload.Type
		err = c.Ctx.Save(&employeeovertimes)
		if err != nil {
			return c.SetResultError(err.Error(), nil)
		}
		newovertimeid = employeeovertimes.IdOvertime
	}

	// update if memberovertime of userid result is pending -> confirmed
	q := tk.M{}
	var dbFiltero []*db.Filter
	dbFiltero = append(dbFiltero, db.Eq("_id", newovertimeid))
	overtime := OvertimeModel{}
	q.Set("where", db.And(dbFiltero...))
	crs, _ = c.Ctx.Find(NewOvertimeModel(), q)
	if crs != nil {
		defer crs.Close()
	}
	err = crs.Fetch(&overtime, 1, false)
	a := 0
	b := 0
	d := false
	dataEmpl := c.GetdateOvertimeEmp(k, employeeovertimes.IdOvertime, payload.Userid)
	tk.Println("-------------- data emp ", tk.JsonString(dataEmpl))

	idl := 0
	for _, dtemp := range dataEmpl {
		if dtemp.Hours > 0 {
			idl++
		}
	}

	for _, each := range overtime.MembersOvertime {
		if each.UserId == payload.Userid && each.Result != "Confirmed" {
			if idl == len(dataEmpl) {
				d = true
			}

			b = a
		}
		a++
	}
	dateNow := time.Now()
	// datestr, _ := time.Parse("2006-01-02", overtime.DayList[0].Date)
	datend, _ := time.Parse("2006-01-02", overtime.DayList[len(overtime.DayList)-1].Date)
	dtend := datend.AddDate(0, 0, 2)

	dateEndForm := dtend.Format("2006-01-02")
	dateNowForm := dateNow.Format("2006-01-02")
	if d {
		if dateEndForm == dateNowForm {
			overtime.MembersOvertime[b].Result = "Confirmed"
		}

		overtime.MembersOvertime[b].TypeOvertime = payload.Type
	}
	err = c.Ctx.Save(&overtime)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	return c.SetResultInfo(false, "success", nil)
}

func (c *OvertimeController) GetdateOvertimeEmp(k *knot.WebContext, idrequest string, userid string) []EmployeeOvertimeModel {
	k.Config.OutputType = knot.OutputJson

	res := []EmployeeOvertimeModel{}
	var dbFilter []*db.Filter
	query := tk.M{}
	dbFilter = append(dbFilter, db.Eq("idovertime", idrequest))
	dbFilter = append(dbFilter, db.Eq("userid", userid))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	crs, err := c.Ctx.Find(NewEmployeeOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return res
	}
	err = crs.Fetch(&res, 0, false)
	if err != nil {
		return res
	}

	return res
}

// GetDataRequestOvertime ...
func (c *OvertimeController) GetDataRequestOvertime(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Designations []string
		Location     string
		MonthYear    string
		Project      string
		Search       string
	}{}
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	t, _ := time.Parse("012006", p.MonthYear)
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", t.Format("2006-01")+".*"))))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("isexpired", false)))
	csr, e := c.Ctx.Connection.NewQuery().Select().From("NewOvertime").Command("pipe", pipe).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	overtimedata := []CNewOvertimeModel{}
	e = csr.Fetch(&overtimedata, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	return c.SetResultInfo(false, "success", overtimedata)
}
