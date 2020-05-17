package controllers

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"
	"errors"
	"time"

	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type RemoteController struct {
	*BaseController
}

func (c *RemoteController) ProcessRemote(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Data     RemoteModel
		DateList []string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	location := k.Session("location").(string)
	dash := DashboardController(*c)
	tz, err := dash.GetTimeZone(k, location)
	userid := k.Session("userid").(string)
	usr := dash.GetDataSessionUser(k, userid)
	payload.Data.Email = usr[0].Email
	// payload.Data.IsExpiredon = false
	//get branchmanager by location
	loc := LocationController(*c)
	dataloc := loc.GetLocationParam(k, location)
	branchmgr := []BranchManagerList{}
	for _, mgr := range dataloc.PC {
		branch := BranchManagerList{}
		branch.UserId = mgr.UserId
		branch.IdEmp = mgr.IdEmp
		branch.Email = mgr.Email
		branch.Location = mgr.Location
		branch.PhoneNumber = mgr.PhoneNumber
		branch.Name = mgr.Name
		branchmgr = append(branchmgr, branch)
	}
	payload.Data.BranchManager = branchmgr
	remoteS := services.RemoteService{
		payload.Data,
		payload.DateList,
		[]RemoteModel{},
	}

	respon, err := remoteS.RequestLeave(tz)
	if err != nil {
		return c.SetResultError(err.Error(), respon)
	}

	//add notif for mobile
	remoteData := respon.Get("RemoteSuccess").([]RemoteModel)
	tl, err := helper.TimeLocation(tz)

	notif := NotificationController(*c)
	dataNotif := NotificationModel{}
	dataNotif.Id = ""
	dataNotif.UserId = payload.Data.UserId
	dataNotif.IsConfirmed = false
	dataNotif.Notif.Name = payload.Data.Name
	dataNotif.IdRequest = remoteData[0].IdOp
	dataNotif.Notif.DateFrom = payload.Data.From
	dataNotif.Notif.DateTo = payload.Data.To
	dataNotif.Notif.Description = "Create request Remote"
	dataNotif.Notif.Status = "Pending"
	dataNotif.Notif.RequestType = "Remote"
	dataNotif.Notif.Reason = payload.Data.Reason
	dataNotif.Notif.ManagerApprove = payload.Data.Projects[0].ProjectManager.Name
	dataNotif.Notif.StatusApproval = "Pending"
	dataNotif.Notif.CreatedAt = tl
	dataNotif.Notif.UpdatedAt = tl

	notif.InsertNotification(dataNotif)
	//add notif for mobile

	return c.SetResultOK(respon)
}

func (c *RemoteController) HandleDecline(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"

	param := tk.M{}
	param.Set("Param", k.Request.FormValue("Param"))
	param.Set("Note", k.Request.FormValue("Note"))

	return param
}

func (c *RemoteController) HandleDeclineNote(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}
	err := k.GetPayload(&param)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	//userid := k.Session("userid").(string)
	//dash := DashboardController(*c)
	//usr := dash.GetDataSessionUser(k, userid)

	paramRes, err := new(services.RemoteService).ValidateParamMail(param, "")
	if err != nil {
		return c.SetResultError(err.Error(), paramRes)
	}

	//update notif for mobile

	p2, errr := new(services.RemoteService).DecodeParam(param.GetString("Param"))
	if errr != nil {
	}
	remotes, errRemote := new(services.RemoteService).GetByIdRequest(p2.GetString("IdRequest"))
	if errRemote != nil {
	}
	location := remotes[0].Location

	dash := DashboardController(*c)
	tz, err := dash.GetTimeZone(k, location)
	tl, err := helper.TimeLocation(tz)

	notif := NotificationController(*c)
	getnotif := notif.GetDataNotification(k, paramRes.GetString("IdRequest"))
	getnotif.Notif.Status = "Declined"
	getnotif.Notif.StatusApproval = "Declined"
	getnotif.Notif.Description = "Declined by " + paramRes.GetString("Type") + " from mail : " + paramRes.GetString("Note")
	getnotif.Notif.UpdatedAt = tl
	getnotif.Notif.ManagerApprove = paramRes.GetString("ApprovalName")

	notif.InsertNotification(getnotif)
	//update notif for mobile

	return c.SetResultOK(paramRes)
}

func (c *RemoteController) HandleDeclineCancel(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}

	return c.SetResultOK(param)
}

func (c *RemoteController) HandleApproval(k *knot.WebContext) interface{} {
	// k.Config.OutputType = knot.OutputJson

	// param := tk.M{}
	// param.Set("Param", k.Request.FormValue("Param"))

	// status := ""
	// message := ""

	// // userid := k.Session("userid").(string)
	// // dash := DashboardController(*c)
	// // usr := dash.GetDataSessionUser(k, userid)

	// paramRes, err := new(services.RemoteService).ValidateParamMail(param, "")
	// if err != nil {
	// 	res := c.SetResultError(err.Error(), paramRes)
	// 	status = string(res.Status)
	// 	message = res.Message

	// 	status = "false"
	// 	message = err.Error()
	// }

	// res := c.SetResultOK(paramRes)

	// if err != nil {
	// 	status = "false"
	// 	message = err.Error()
	// } else {
	// 	status = paramRes.GetString("Status")
	// 	message = res.Message
	// }

	// typeOfRequest := paramRes.GetString("Type")
	// if paramRes.Get("IsExpired") != nil {
	// 	if paramRes.Get("IsExpired").(bool) {
	// 		message = "request remote is already expired"
	// 	}
	// }
	// if typeOfRequest == "cancelremote" {
	// 	if status == "true" {
	// 		message = "you already approved cancel request remote"
	// 	} else if status == "false" {
	// 		message = "you already decline cancel request remote"
	// 	}
	// }

	// if paramRes.Get("IsExpired") == nil {
	// 	//update notif for mobile

	// 	p2, errr := new(services.RemoteService).DecodeParam(param.GetString("Param"))
	// 	if errr != nil {
	// 	}
	// 	remotes, errRemote := new(services.RemoteService).GetByIdRequest(p2.GetString("IdRequest"))
	// 	if errRemote != nil {
	// 	}
	// 	location := remotes[0].Location

	// 	dash := DashboardController(*c)
	// 	tz, _ := dash.GetTimeZone(k, location)
	// 	tl, _ := helper.TimeLocation(tz)

	// 	notif := NotificationController(*c)
	// 	getnotif := notif.GetDataNotification(k, paramRes.GetString("IdRequest"))
	// 	getnotif.Notif.Description = "Approved by " + typeOfRequest + " from mail"
	// 	getnotif.Notif.UpdatedAt = tl

	// 	if typeOfRequest == "manager" {
	// 		getnotif.Notif.Status = "Approved"
	// 		getnotif.Notif.StatusApproval = "Approved"
	// 		getnotif.Notif.ManagerApprove = paramRes.GetString("ApprovalName")
	// 	}

	// 	notif.InsertNotification(getnotif)
	// 	//update notif for mobile
	// }

	// http.Redirect(k.Writer, k.Request, "/remote/responapproval?Status="+status+"&Message="+message, http.StatusTemporaryRedirect)

	// return res
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"

	dataView := tk.M{}
	dataView.Set("Status", k.Request.FormValue("Status"))
	dataView.Set("Message", k.Request.FormValue("Message"))

	return dataView
}

func (c *RemoteController) ResponApproval(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	param := tk.M{}
	err := k.GetPayload(&param)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	paramRes, err := new(services.RemoteService).ValidateParamMail(param, "")
	tk.Println("----------------- paramRes expired ", paramRes)
	if err != nil {

		return c.SetResultInfo(true, err.Error(), paramRes)
	}

	typeOfRequest := paramRes.GetString("Type")
	if paramRes.Get("IsExpired") == true {
		if paramRes.Get("IsExpired").(bool) {
			return c.SetResultInfo(true, "request remote is already expired", paramRes)
		}
	}
	if typeOfRequest == "cancelremote" {
		if paramRes.GetString("Status") == "true" {
			return c.SetResultInfo(true, "request already approved request cancel remote", paramRes)
		} else if paramRes.GetString("Status") == "false" {
			return c.SetResultInfo(true, "request already decline request cancel remote", paramRes)
		}
	} else if typeOfRequest == "leader" {
		if paramRes.GetString("Status") == "true" {
			return c.SetResultInfo(true, "request already approved", paramRes)
		}
	} else if typeOfRequest == "manager" {
		if paramRes.GetString("Status") == "true" {
			return c.SetResultInfo(true, "request already approved", paramRes)
		}
	}

	if paramRes.Get("IsExpired") == nil {
		//update notif for mobile

		p2, errr := new(services.RemoteService).DecodeParam(param.GetString("Param"))
		if errr != nil {
			return c.SetResultInfo(true, errr.Error(), nil)
		}
		remotes, errRemote := new(services.RemoteService).GetByIdRequest(p2.GetString("IdRequest"))
		if errRemote != nil {
			return c.SetResultInfo(true, errRemote.Error(), nil)
		}
		location := remotes[0].Location

		dash := DashboardController(*c)
		tz, _ := dash.GetTimeZone(k, location)
		tl, _ := helper.TimeLocation(tz)

		notif := NotificationController(*c)
		getnotif := notif.GetDataNotification(k, paramRes.GetString("IdRequest"))
		getnotif.Notif.Description = "Approved by " + typeOfRequest + " from mail"
		getnotif.Notif.UpdatedAt = tl

		if typeOfRequest == "manager" {
			getnotif.Notif.Status = "Approved"
			getnotif.Notif.StatusApproval = "Approved"
			getnotif.Notif.ManagerApprove = paramRes.GetString("ApprovalName")
		}

		notif.InsertNotification(getnotif)
		//update notif for mobile
	}

	// // http.Redirect(k.Writer, k.Request, "/remote/responapproval?Status="+status+"&Message="+message, http.StatusTemporaryRedirect)
	// tk.Println(message)

	return c.SetResultInfo(false, "request approved successfully", paramRes)
}

func (c *RemoteController) GetRemoteByRequest(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		IDRequest string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	remotes, err := new(services.RemoteService).GetByIdRequest(payload.IDRequest)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(remotes)
}

func (c *RemoteController) GetRemoteNeedApproval(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		UserId string
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	jobrole := k.Session("jobrolelevel")
	remotes, err := new(services.RemoteService).GetRemoteNeedApproval(payload.UserId, jobrole.(int))
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(remotes)
}

func (c *RemoteController) GetRemoteCancelNeedApproval(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		UserId string
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	jobrole := k.Session("jobrolelevel")
	remotes, err := new(services.RemoteService).GetRemoteCancelNeedApproval(payload.UserId, jobrole.(int))
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(remotes)
}

func (c *RemoteController) HandleApprovalFromApp(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Param []tk.M
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	// userIdMgr := k.Session("userid").(string)
	// payload.Param.Set("ManagerUserId", userIdMgr)
	remotes, err := new(services.RemoteService).ValidateParamFromApp(payload.Param)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	// userApproval := payload.Param[0].GetString("Type")
	// statusApproval := payload.Param[0].GetString("Status")

	getP := payload.Param
	for _, dt := range getP {
		userApproval := dt.GetString("Type")
		statusApproval := dt.GetString("Status")
		//tk.Println("not.....", dt.GetString("Type"))
		if userApproval == "leader" || userApproval == "manager" {
			//update notif for mobile
			location := k.Session("location").(string)
			dash := DashboardController(*c)
			tz, _ := dash.GetTimeZone(k, location)
			tl, _ := helper.TimeLocation(tz)
			statusReq := "Approved"
			declineReason := ""
			if statusApproval == "false" {
				statusReq = "Declined"
				declineReason = " : " + dt.GetString("Note")
			}

			remote := remotes.([]RemoteModel)
			approvalName := remote[0].Projects[0].ProjectLeader.Name
			if userApproval == "manager" {
				approvalName = remote[0].Projects[0].ProjectManager.Name
			}

			notif := NotificationController(*c)
			getnotif := notif.GetDataNotification(k, dt.GetString("IdRequest"))
			getnotif.Notif.Description = statusReq + " by " + userApproval + " from app" + declineReason
			getnotif.Notif.UpdatedAt = tl

			if (userApproval == "leader" && statusApproval == "false") || userApproval == "manager" {
				getnotif.Notif.Status = statusReq
				getnotif.Notif.StatusApproval = statusReq
				getnotif.Notif.ManagerApprove = approvalName
			}

			notif.InsertNotification(getnotif)
			//update notif for mobile
		}
	}

	// if userApproval == "leader" || userApproval == "manager" {
	// 	//update notif for mobile
	// 	location := k.Session("location").(string)
	// 	dash := DashboardController(*c)
	// 	tz, _ := dash.GetTimeZone(k, location)
	// 	tl, _ := helper.TimeLocation(tz)
	// 	statusReq := "Approved"
	// 	declineReason := ""
	// 	if statusApproval == "false" {
	// 		statusReq = "Declined"
	// 		declineReason = " : " + payload.Param[0].GetString("Note")
	// 	}

	// 	remote := remotes.([]RemoteModel)
	// 	approvalName := remote[0].Projects[0].ProjectLeader.Name
	// 	if userApproval == "manager" {
	// 		approvalName = remote[0].Projects[0].ProjectManager.Name
	// 	}

	// 	notif := NotificationController(*c)
	// 	getnotif := notif.GetDataNotification(k, payload.Param[0].GetString("IdRequest"))
	// 	getnotif.Notif.Description = statusReq + " by " + userApproval + " from app" + declineReason
	// 	getnotif.Notif.UpdatedAt = tl

	// 	if (userApproval == "leader" && statusApproval == "false") || userApproval == "manager" {
	// 		getnotif.Notif.Status = statusReq
	// 		getnotif.Notif.StatusApproval = statusReq
	// 		getnotif.Notif.ManagerApprove = approvalName
	// 	}

	// 	notif.InsertNotification(getnotif)
	// 	//update notif for mobile
	// }

	return c.SetResultOK(remotes)
}

func (c *RemoteController) HandleApprovalDetails(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Remotes       []RemoteModel
		DateLIst      []string
		TypeOfRequest string
		SendNotif     bool
		AppList       int
		DecList       int
	}{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	userIdMgr := k.Session("userid").(string)
	for i, remote := range payload.Remotes {
		for _, each := range remote.BranchManager {
			if each.UserId == userIdMgr {
				for j, _ := range remote.Projects {
					payload.Remotes[i].Projects[j].ProjectManager.UserId = each.UserId
					payload.Remotes[i].Projects[j].ProjectManager.Email = each.Email
					payload.Remotes[i].Projects[j].ProjectManager.Name = each.Name
					payload.Remotes[i].Projects[j].ProjectManager.IdEmp = each.IdEmp
					payload.Remotes[i].Projects[j].ProjectManager.PhoneNumber = each.PhoneNumber
					payload.Remotes[i].Projects[j].ProjectManager.Location = each.Location
				}
			}
		}
	}
	err = new(services.RemoteService).ValidateParamFromAppDetail(payload.TypeOfRequest, payload.Remotes, payload.DateLIst)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	//update notif for mobile
	location := k.Session("location").(string)
	dash := DashboardController(*c)
	tz, _ := dash.GetTimeZone(k, location)
	tl, _ := helper.TimeLocation(tz)
	statusReq := "Declined"
	if payload.AppList >= payload.DecList {
		statusReq = "Approved"
	}

	approvalName := payload.Remotes[0].Projects[0].ProjectLeader.Name
	if payload.TypeOfRequest == "manager" {
		approvalName = payload.Remotes[0].Projects[0].ProjectManager.Name
	}

	notif := NotificationController(*c)
	getnotif := notif.GetDataNotification(k, payload.Remotes[0].IdOp)
	getnotif.Notif.Description = statusReq + " by " + payload.TypeOfRequest + " from app by details"
	getnotif.Notif.UpdatedAt = tl

	if payload.SendNotif {
		getnotif.Notif.Status = statusReq
		getnotif.Notif.StatusApproval = statusReq
		getnotif.Notif.ManagerApprove = approvalName
	}

	notif.InsertNotification(getnotif)
	//update notif for mobile

	return c.SetResultOK(payload.Remotes)
}

func (c *RemoteController) GetRemoteApproved(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		UserId string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	jobrole := k.Session("jobrolelevel")

	remotes, err := new(services.RemoteService).GetRemoteApproved(payload.UserId, jobrole.(int))
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(remotes)
}

func (c *RemoteController) ProcessCancel(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := tk.M{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	level := (k.Session("jobrolelevel")).(int)
	dash := DashboardController(*c)
	userid := k.Session("userid").(string)
	user := dash.GetDataSessionUser(k, userid)
	err = new(services.RemoteService).ProcessCancel(payload, level, user[0].Fullname)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	//update notif for mobile
	location := k.Session("location").(string)
	tz, err := dash.GetTimeZone(k, location)
	tl, err := helper.TimeLocation(tz)

	notif := NotificationController(*c)
	getnotif := notif.GetDataNotification(k, payload.GetString("IDRequest"))
	getnotif.Notif.Status = "Cancelled"
	getnotif.Notif.StatusApproval = "Cancelled"
	getnotif.Notif.Description = "Cancelled by " + user[0].Fullname + " from app : " + payload.GetString("Reason")
	getnotif.Notif.UpdatedAt = tl
	getnotif.Notif.ManagerApprove = user[0].Fullname

	notif.InsertNotification(getnotif)
	//update notif for mobile

	return c.SetResultOK(nil)
}

func (c *RemoteController) CheckMonthRemote(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	tm := time.Now()
	dateNow := tm.Format("2006-01")

	nm := tm.AddDate(0, 1, 0)
	next := nm.Format("2006-01")

	pipe := []tk.M{}
	userid := k.Session("userid").(string)
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid}})
	pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalmanager": true}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isdelete": false}})
	pipe = append(pipe, tk.M{"$match": tk.M{"isexpired": false}})
	// pipe = append(pipe, tk.M{"$match": tk.M{"dateleave": tk.M{"$regex": dateNow, "$options": "g"}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"$or": []tk.M{tk.M{"dateleave": tk.M{"$regex": dateNow, "$options": "g"}}, tk.M{"dateleave": tk.M{"$regex": next, "$options": "g"}}}}})
	tk.Println("-------------- ", next)
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"$substr": []interface{}{"$dateleave", 5, 2}}, "count": tk.M{"$sum": 1}}})

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("remote").
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

	// tk.Println("-------------- ", tk.JsonString(data))

	return data
}

// HandleApprovalFromAppAll is
func (c *RemoteController) HandleApprovalFromAppAll(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Param []tk.M
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	// userIdMgr := k.Session("userid").(string)
	// payload.Param.Set("ManagerUserId", userIdMgr)
	remotes, err := new(services.RemoteService).ValidateParamFromAppAll(payload.Param)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	getP := payload.Param
	for _, dt := range getP {
		userApproval := dt.GetString("Type")
		statusApproval := dt.GetString("Status")
		//tk.Println("not.....", dt.GetString("Type"))
		if userApproval == "leader" || userApproval == "manager" {
			//update notif for mobile
			location := k.Session("location").(string)
			dash := DashboardController(*c)
			tz, _ := dash.GetTimeZone(k, location)
			tl, _ := helper.TimeLocation(tz)
			statusReq := "Approved"
			declineReason := ""
			if statusApproval == "false" {
				statusReq = "Declined"
				declineReason = " : " + dt.GetString("Note")
			}

			remote := remotes.([]RemoteModel)
			approvalName := remote[0].Projects[0].ProjectLeader.Name
			if userApproval == "manager" {
				approvalName = remote[0].Projects[0].ProjectManager.Name
			}

			notif := NotificationController(*c)
			getnotif := notif.GetDataNotification(k, dt.GetString("IdRequest"))
			getnotif.Notif.Description = statusReq + " by " + userApproval + " from app" + declineReason
			getnotif.Notif.UpdatedAt = tl

			if (userApproval == "leader" && statusApproval == "false") || userApproval == "manager" {
				getnotif.Notif.Status = statusReq
				getnotif.Notif.StatusApproval = statusReq
				getnotif.Notif.ManagerApprove = approvalName
			}

			notif.InsertNotification(getnotif)
			//update notif for mobile
		}
	}

	// userApproval := payload.Param[0].GetString("Type")
	// statusApproval := payload.Param[0].GetString("Status")

	// if userApproval == "leader" || userApproval == "manager" {
	// 	//update notif for mobile
	// 	location := k.Session("location").(string)
	// 	dash := DashboardController(*c)
	// 	tz, _ := dash.GetTimeZone(k, location)
	// 	tl, _ := helper.TimeLocation(tz)
	// 	statusReq := "Approved"
	// 	declineReason := ""
	// 	if statusApproval == "false" {
	// 		statusReq = "Declined"
	// 		declineReason = " : " + payload.Param[0].GetString("Note")
	// 	}

	// 	remote := remotes.([]RemoteModel)
	// 	approvalName := remote[0].Projects[0].ProjectLeader.Name
	// 	if userApproval == "manager" {
	// 		approvalName = remote[0].Projects[0].ProjectManager.Name
	// 	}

	// 	notif := NotificationController(*c)
	// 	getnotif := notif.GetDataNotification(k, payload.Param[0].GetString("IdRequest"))
	// 	getnotif.Notif.Description = statusReq + " by " + userApproval + " from app" + declineReason
	// 	getnotif.Notif.UpdatedAt = tl

	// 	if (userApproval == "leader" && statusApproval == "false") || userApproval == "manager" {
	// 		getnotif.Notif.Status = statusReq
	// 		getnotif.Notif.StatusApproval = statusReq
	// 		getnotif.Notif.ManagerApprove = approvalName
	// 	}

	// 	notif.InsertNotification(getnotif)
	// 	//update notif for mobile
	// }

	return c.SetResultOK(remotes)
}
