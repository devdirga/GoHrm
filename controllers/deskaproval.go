package controllers

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"
	"strings"

	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type DeskAprovalController struct {
	*BaseController
}

func (c *DeskAprovalController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputTemplate
	DataAccess := c.SetViewData(k, nil)

	return DataAccess
}

func (c *DeskAprovalController) RequestPending(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	dashboard := DashboardController(*c)
	data, _ := dashboard.FetchRequestLeave(k, tk.ToString(k.Session("userid")), "Pending")
	return data
}

func (c *DeskAprovalController) DecisionManager(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Id        string
		Result    string
		Ismanager string
		Reason    string
	}{}

	userid := tk.ToString(k.Session("userid"))

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	log := tk.M{}
	level := k.Session("jobrolelevel").(int)
	mail := MailController(*c)
	dataLeave := mail.getDataLeave(payload.Id)
	// tk.Println("============================ masuk sini ya ", tk.JsonString(dataLeave))
	if dataLeave.ResultRequest == "Approved" {
		return c.SetResultInfo(true, "Request has been "+dataLeave.ResultRequest+" Manager by "+dataLeave.StatusManagerProject.Name, nil)
	} else if dataLeave.ResultRequest == "Declined" {
		return c.SetResultInfo(true, "Request has been "+dataLeave.ResultRequest+" Manager by "+dataLeave.StatusManagerProject.Name, nil)
	}
	dash := DashboardController(*c)
	usr := dash.GetDataSessionUser(k, dataLeave.UserId)
	mgr := dash.GetDataSessionUser(k, userid)
	count := 0
	for _, ky := range dataLeave.BranchManager {
		if ky.UserId == userid {
			dataLeave.StatusManagerProject.IdEmp = ky.IdEmp
			dataLeave.StatusManagerProject.Name = ky.Name
			dataLeave.StatusManagerProject.Location = ky.Location
			dataLeave.StatusManagerProject.Email = ky.Email
			dataLeave.StatusManagerProject.PhoneNumber = ky.PhoneNumber
			dataLeave.StatusManagerProject.UserId = ky.UserId
			count = count + 1
		} else if level == 5 {
			dataLeave.StatusManagerProject.IdEmp = usr[0].EmpId
			dataLeave.StatusManagerProject.Name = usr[0].Fullname
			dataLeave.StatusManagerProject.Location = usr[0].Location
			dataLeave.StatusManagerProject.Email = usr[0].Email
			dataLeave.StatusManagerProject.PhoneNumber = usr[0].PhoneNumber
			dataLeave.StatusManagerProject.UserId = usr[0].Id
		}

	}
	if level == 5 || level == 6 {
		dataLeave.StatusManagerProject.IdEmp = mgr[0].EmpId
		dataLeave.StatusManagerProject.Name = mgr[0].Fullname
		dataLeave.StatusManagerProject.Location = mgr[0].Location
		dataLeave.StatusManagerProject.Email = mgr[0].Email
		dataLeave.StatusManagerProject.PhoneNumber = mgr[0].PhoneNumber
		dataLeave.StatusManagerProject.UserId = mgr[0].Id
	}
	if payload.Ismanager == "true" {
		dataLeave.StatusManagerProject.StatusRequest = payload.Result
		dataLeave.ResultRequest = payload.Result

		if payload.Result == "Approved" {
			dash.SetHistoryLeave(k, userid, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Approved Manager by "+dataLeave.StatusManagerProject.Name, "Approved", dataLeave)
			if dataLeave.IsSpecials == false {

				if dataLeave.NoOfDays > usr[0].YearLeave {
					tk.Println("--------------- masuk1")
					return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
				} else {
					if dataLeave.IsEmergency == true {
						usr[0].YearLeave = usr[0].YearLeave - 1
						usr[0].DecYear = usr[0].DecYear - 1.0
						//update tmpyear
						usr[0].TmpYear = usr[0].TmpYear - dataLeave.NoOfDays
					} else {
						//update tmpYear
						usr[0].TmpYear = usr[0].TmpYear - dataLeave.NoOfDays
					}

					if dataLeave.IsSpecials == false {

						err = c.Ctx.Save(usr[0])
						if err != nil {
							return c.SetResultInfo(true, err.Error(), nil)
						}
					}

				}

			}
			emptyremote := new(RemoteModel)
			service := services.LogService{
				dataLeave,
				emptyremote,
				"leave",
			}
			desc := "Request approved by manager"
			var stsReq = "Approved"
			log.Set("Status", stsReq)
			log.Set("Desc", desc)
			log.Set("NameLogBy", dataLeave.StatusManagerProject.Name)
			log.Set("EmailNameLogBy", dataLeave.StatusManagerProject.Email)
			err = service.ApproveDeclineLog(log)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}
			// manager project list
			log2 := log
			log2.Set("Desc", "Request approved by PM")
			if len(dataLeave.ProjectManagerList) > 0 {
				log2.Set("NameLogBy", dataLeave.ProjectManagerList[0].Name)
				log2.Set("EmailNameLogBy", dataLeave.ProjectManagerList[0].Email)
			}
			err = service.ApproveDeclineLog(log2)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

		} else {
			//reason input
			dataLeave.StatusManagerProject.Reason = payload.Reason
			dash.SetHistoryLeave(k, userid, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Declined Manager by "+dataLeave.StatusManagerProject.Name, "Declined", dataLeave)
			emptyremote := new(RemoteModel)
			service := services.LogService{
				dataLeave,
				emptyremote,
				"leave",
			}
			desc := "Request declined by manager"
			var stsReq = "Declined"
			log.Set("Status", stsReq)
			log.Set("Desc", desc)
			log.Set("NameLogBy", dataLeave.StatusManagerProject.Name)
			log.Set("EmailNameLogBy", dataLeave.StatusManagerProject.Email)
			err = service.ApproveDeclineLog(log)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}
			// manager project list
			log2 := log
			log2.Set("Desc", "Request declined by PM")
			if len(dataLeave.ProjectManagerList) > 0 {
				log2.Set("NameLogBy", dataLeave.ProjectManagerList[0].Name)
				log2.Set("EmailNameLogBy", dataLeave.ProjectManagerList[0].Email)
			}
			err = service.ApproveDeclineLog(log2)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}
		}

		user := []string{dataLeave.StatusManagerProject.Email, dataLeave.Email}

		if len(dataLeave.ProjectManagerList) > 0 {
			for _, mg := range dataLeave.ProjectManagerList {
				user = append(user, mg.Email)
			}
		}

		if len(dataLeave.StatusBusinesAnalyst) > 0 {
			for _, ba := range dataLeave.StatusBusinesAnalyst {
				user = append(user, ba.Email)
			}
		}

		if len(dataLeave.StatusProjectLeader) > 0 {
			for _, ld := range dataLeave.StatusProjectLeader {
				if ld.Email != "" {
					user = append(user, ld.Email)
				}

			}
		}

		if dataLeave.Location == "Indonesia" {
			dataHR, err := mail.GetDataHR(k)
			if err != nil {
				return err
			}

			if len(dataHR) > 0 {
				user = append(user, dataHR[0].ManagingDirector.Email)
				user = append(user, dataHR[0].AccountManager.Email)
				for _, staf := range dataHR[0].Staf {
					user = append(user, staf.Email)
				}
			}
		}

		// if dataLeave.NoOfDays > usr[0].YearLeave {
		// 	return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
		// }

		err = c.Ctx.Save(dataLeave)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		mail.SendMailUser(k, payload.Id, user)
		// mail.RequestLeaveOnDate(dataLeave, dataLeave.UserId)
		listDetailLeave := dash.GetLeavebyDateFilterIdTransc(k, dataLeave.Id)
		for i, ireqByDate := range listDetailLeave {
			if ireqByDate.DateLeave == dataLeave.LeaveDateList[0] && dataLeave.IsEmergency == true {
				if i == 0 {
					ireqByDate.IsCutOff = true
				}
			}
			if payload.Result == "Approved" {
				ireqByDate.StsByManager = payload.Result
			} else {
				ireqByDate.StsByManager = payload.Result
			}
			err = c.Ctx.Save(ireqByDate)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

		}
		notif := NotificationController(*c)

		getnotif := notif.GetDataNotification(k, dataLeave.Id)
		getnotif.Notif.ManagerApprove = dataLeave.StatusManagerProject.Name
		getnotif.Notif.Status = dataLeave.ResultRequest
		getnotif.Notif.StatusApproval = dataLeave.ResultRequest
		getnotif.Notif.Description = payload.Reason

		notif.InsertNotification(getnotif)
	} else {
		if dataLeave.ResultRequest == "Canceled by User" {
			return c.SetResultInfo(true, "Request has been canceled by user", nil)
		}
		listDetailLeave := dash.GetLeavebyDateFilterIdTransc(k, dataLeave.Id)
		for _, ireqByDate := range listDetailLeave {
			ireqByDate.StsByLeader = payload.Result
			err = c.Ctx.Save(ireqByDate)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}
		}

		for i, data := range dataLeave.StatusProjectLeader {
			if data.UserId == userid {
				dataLeave.StatusProjectLeader[i].StatusRequest = payload.Result
				// fmt.Printf("---------------- dataleve.id", dataLeave.Id)
			}

			if payload.Result == "Approved" {
				emptyremote := new(RemoteModel)
				service := services.LogService{
					dataLeave,
					emptyremote,
					"leave",
				}
				desc := "Request approved by leader"
				var stsReq = "Approved"
				log.Set("Status", stsReq)
				log.Set("Desc", desc)
				log.Set("NameLogBy", data.Name)
				log.Set("EmailNameLogBy", data.Email)
				err = service.ApproveDeclineLog(log)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
				// dash.SetHistoryLeave(k, userid, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Approved Leader by "+data.Name, "Approved")
			} else {
				//input reason
				dataLeave.StatusProjectLeader[i].Reason = payload.Reason
				emptyremote := new(RemoteModel)
				service := services.LogService{
					dataLeave,
					emptyremote,
					"leave",
				}
				desc := "Request declined by leader"
				var stsReq = "Declined"
				log.Set("Status", stsReq)
				log.Set("Desc", desc)
				log.Set("NameLogBy", data.Name)
				log.Set("EmailNameLogBy", data.Email)
				err = service.ApproveDeclineLog(log)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}

				// dash.SetHistoryLeave(k, userid, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Declined Leader by "+data.Name, "Declined")

			}

		}

		err = c.Ctx.Save(dataLeave)
		// mail.SendMailManager(k, userid, dataLeave.Id)
		mail.CheckLeaderDeclined(k, dataLeave, level)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

	}

	return c.SetResultInfo(false, "save successfully", nil)
}

func (c *DeskAprovalController) ApproveAll(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)
	payload := []struct {
		Id        string
		Result    string
		Ismanager bool
	}{}

	userid := tk.ToString(k.Session("userid"))
	level := k.Session("jobrolelevel").(int)

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	stsReq := StatusRequestController(*c)
	dash := DashboardController(*c)
	dtusr := dash.GetDataSessionUser(k, userid)
	mail := MailController(*c)
	notif := NotificationController(*c)
	for _, dt := range payload {
		dataLeave := mail.getDataLeave(dt.Id)
		databyDate := stsReq.GetLeaveDateByIdReq(k, dataLeave.Id)

		if dt.Ismanager == true {
			for _, ky := range dataLeave.BranchManager {
				if ky.UserId == userid {
					dataLeave.StatusManagerProject.IdEmp = ky.IdEmp
					dataLeave.StatusManagerProject.Name = ky.Name
					dataLeave.StatusManagerProject.Location = ky.Location
					dataLeave.StatusManagerProject.Email = ky.Email
					dataLeave.StatusManagerProject.PhoneNumber = ky.PhoneNumber
					dataLeave.StatusManagerProject.UserId = ky.UserId
				} else if level == 5 {
					dataLeave.StatusManagerProject.IdEmp = dtusr[0].EmpId
					dataLeave.StatusManagerProject.Name = dtusr[0].Fullname
					dataLeave.StatusManagerProject.Location = dtusr[0].Location
					dataLeave.StatusManagerProject.Email = dtusr[0].Email
					dataLeave.StatusManagerProject.PhoneNumber = dtusr[0].PhoneNumber
					dataLeave.StatusManagerProject.UserId = dtusr[0].Id
				}
			}
			dataLeave.StatusManagerProject.StatusRequest = dt.Result
			dataLeave.ResultRequest = dt.Result
			err = c.Ctx.Save(dataLeave)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}

			for _, dtDateManager := range databyDate {
				dtDateManager.StsByManager = "Approved"
				err = c.Ctx.Save(dtDateManager)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
			}

			if dataLeave.IsSpecials == false {
				// usr := dash.GetDataSessionUser(k, dataLeave.UserId)
				if len(dataLeave.DetailsLeave) > 0 {
					for _, dtl := range dataLeave.DetailsLeave {
						if dtl.IsApproved == true {
							// usr[0].YearLeave = usr[0].YearLeave - 1
						}
					}
				}
				// err = c.Ctx.Save(usr[0])
				// if err != nil {
				// 	return c.SetResultInfo(true, err.Error(), nil)
				// }
			}

			// mail.RequestLeaveOnDate(dataLeave, dataLeave.UserId)
			user := []string{dataLeave.Email}
			for _, mgr := range dataLeave.ProjectManagerList {
				if mgr.Email != "" {
					user = append(user, mgr.Email)
				}

			}
			for _, ba := range dataLeave.StatusBusinesAnalyst {
				if ba.Email != "" {
					user = append(user, ba.Email)
				}

			}
			for _, ld := range dataLeave.StatusProjectLeader {
				if ld.Email != "" {
					user = append(user, ld.Email)
				}

			}
			mail.SendMailUser(k, dt.Id, user)
			tk.Println("----------- masuk")
			dash.SetHistoryLeave(k, dataLeave.UserId, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Approved Manager by "+dataLeave.StatusManagerProject.Name, "Approved", dataLeave)
			// mail.RequestLeaveOnDateV2(dataLeave, dataLeave.UserId)

			getnotif := notif.GetDataNotification(k, dataLeave.Id)
			getnotif.Notif.ManagerApprove = dataLeave.StatusManagerProject.Name
			getnotif.Notif.Status = dataLeave.ResultRequest
			getnotif.Notif.StatusApproval = dataLeave.ResultRequest

			notif.InsertNotification(getnotif)

		} else {
			for i, data := range dataLeave.StatusProjectLeader {
				if data.UserId == userid {
					dataLeave.StatusProjectLeader[i].StatusRequest = dt.Result
					err = c.Ctx.Save(dataLeave)
					if err != nil {
						return c.SetResultInfo(true, err.Error(), nil)
					}
					for _, dtDateLeader := range databyDate {
						dtDateLeader.StsByLeader = "Approved"
						err = c.Ctx.Save(dtDateLeader)
						if err != nil {
							return c.SetResultInfo(true, err.Error(), nil)
						}
					}
					mail.SendMailManager(k, dataLeave.UserId, dataLeave.Id, 2)
					dash.SetHistoryLeave(k, dataLeave.UserId, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Approved Leader by "+data.Name, "Pending", dataLeave)

					getnotif := notif.GetDataNotification(k, dataLeave.Id)
					getnotif.Notif.ManagerApprove = dataLeave.StatusManagerProject.Name
					getnotif.Notif.Status = dataLeave.ResultRequest
					getnotif.Notif.StatusApproval = dataLeave.ResultRequest

					notif.InsertNotification(getnotif)
					// mail.RequestLeaveOnDateV2(dataLeave, dataLeave.UserId)

				}
			}
		}
		////save in log
		tk.Println("=============================== data manager ", dataLeave)
		emptyremote := new(RemoteModel)
		log := tk.M{}
		service := services.LogService{
			dataLeave,
			emptyremote,
			"leave",
		}
		desc := "Request approved by manager"
		var stsReq = "Approved"
		log.Set("Status", stsReq)
		log.Set("Desc", desc)
		log.Set("NameLogBy", dataLeave.StatusManagerProject.Name)
		log.Set("EmailNameLogBy", dataLeave.StatusManagerProject.Email)
		err = service.ApproveDeclineLog(log)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
		// manager project list
		log2 := log
		log2.Set("Desc", "Request approved by PM")
		if len(dataLeave.ProjectManagerList) > 0 {
			log2.Set("NameLogBy", dataLeave.ProjectManagerList[0].Name)
			log2.Set("EmailNameLogBy", dataLeave.ProjectManagerList[0].Email)
		}
		err = service.ApproveDeclineLog(log2)

	}

	return c.SetResultInfo(false, "save successfully", nil)
}

func (c *DeskAprovalController) ApprovebyDetail(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := []struct {
		Id           string
		IdRequest    string
		StsByLeader  string
		StsByManager string
		DateLeave    string
		ReasonAction string
	}{}

	jobrolelevel := k.Session("jobrolelevel").(int)
	userid := k.Session("userid").(string)
	// UserName := k.Session("username").(string)
	// tk.Println("------------------- username", UserName)
	dash := DashboardController(*c)
	usr := dash.GetDataSessionUser(k, userid)
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	tk.Println("-------- ", payload[0].ReasonAction)
	mail := MailController(*c)
	dataLeave := mail.getDataLeave(payload[0].IdRequest)
	if dataLeave.ResultRequest == "Approved" {
		return c.SetResultInfo(true, "Request has been "+dataLeave.ResultRequest+" Manager by "+dataLeave.StatusManagerProject.Name, nil)
	} else if dataLeave.ResultRequest == "Declined" {
		return c.SetResultInfo(true, "Request has been "+dataLeave.ResultRequest+" Manager by "+dataLeave.StatusManagerProject.Name, nil)
	}
	res := 0
	log := tk.M{}
	// var idReq string
	ArrApp := []DetailAprovalLeader{}
	dtApp := DetailAprovalLeader{}
	mailCtrler := MailController(*c)
	leaveHeader := mailCtrler.getDataLeave(payload[0].IdRequest)
	datelistDeclined := []string{}
	// dash := DashboardController(*c)
	usrt := dash.GetDataSessionUser(k, dataLeave.UserId)
	ArrApp = leaveHeader.DetailsLeave
	if jobrolelevel == 6 || jobrolelevel == 1 {
		if dataLeave.NoOfDays > usrt[0].YearLeave {
			tk.Println("--------------- masuk1hahahahaha")
			return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
		}
	}

	for i, iVal := range payload {
		// idReq = iVal.IdRequest

		stsReq := StatusRequestController(*c)
		detailIVAl := stsReq.GetDataDetailLeaveByID(k, iVal.Id)

		if jobrolelevel == 2 {
			tk.Println("--------- masuk sini")
			if len(leaveHeader.DetailsLeave) > 0 {

				if len(ArrApp) > 0 {

					dtApp.LeaderName = usr[0].Fullname
					dtApp.DateLeave = iVal.DateLeave
					dtApp.Reason = ""
					dtApp.IdRequest = iVal.IdRequest
					dtApp.UserId = userid
					dtApp.Reason = iVal.ReasonAction
					detailIVAl.StsByLeader = iVal.StsByLeader
					if iVal.StsByLeader == "Approved" {
						dtApp.IsApproved = true
						res = res + 1
					} else if iVal.StsByLeader == "Declined" {
						dtApp.IsApproved = false
						res = res + 0
						datelistDeclined = append(datelistDeclined, dtApp.DateLeave)
					}
					ArrApp = append(ArrApp, dtApp)
				}

			} else {
				dtApp.LeaderName = usr[0].Fullname
				dtApp.DateLeave = iVal.DateLeave
				dtApp.Reason = ""
				dtApp.IdRequest = iVal.IdRequest
				dtApp.UserId = userid
				dtApp.Reason = iVal.ReasonAction
				detailIVAl.StsByLeader = iVal.StsByLeader
				if iVal.StsByLeader == "Approved" {
					dtApp.IsApproved = true
					res = res + 1
				} else if iVal.StsByLeader == "Declined" {
					dtApp.IsApproved = false
					res = res + 0
					datelistDeclined = append(datelistDeclined, dtApp.DateLeave)
				}
				ArrApp = append(ArrApp, dtApp)
			}

		} else if jobrolelevel == 1 {
			// if dataLeave.NoOfDays > usrt[0].YearLeave {
			// 	tk.Println("--------------- masuk1")
			// 	return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
			// }
			detailIVAl.StsByManager = iVal.StsByManager
			dtApp.LeaderName = usr[0].Fullname
			dtApp.DateLeave = iVal.DateLeave
			dtApp.Reason = iVal.ReasonAction
			dtApp.IdRequest = iVal.IdRequest
			dtApp.UserId = userid
			detailIVAl.StsByLeader = iVal.StsByLeader
			detailIVAl.StsByManager = iVal.StsByManager
			if iVal.StsByManager == "Approved" {
				// detailIVAl.IsDelete = false
				dtApp.IsApproved = true
				res = res + 1
				if dataLeave.IsSpecials == false {
					if res > usrt[0].YearLeave {
						return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
					} else if dataLeave.IsEmergency == true {
						if i == 0 {
							usrt[0].YearLeave = usrt[0].YearLeave - 1
							//update tmpyear
							usrt[0].TmpYear = usrt[0].TmpYear - 1
							err = c.Ctx.Save(usrt[0])
							if err != nil {
								return c.SetResultInfo(true, err.Error(), nil)
							}
						}

					} else if dataLeave.IsEmergency == false {
						if i == 0 {
							//update tmpyear
							usrt[0].TmpYear = usrt[0].TmpYear - 1
							err = c.Ctx.Save(usrt[0])
							if err != nil {
								return c.SetResultInfo(true, err.Error(), nil)
							}
						}
					}

				}

			} else if iVal.StsByManager == "Declined" {
				// detailIVAl.IsDelete = true
				dtApp.IsApproved = false
				res = res + 0
				datelistDeclined = append(datelistDeclined, dtApp.DateLeave)
			}
			ArrApp = append(ArrApp, dtApp)
		} else if jobrolelevel == 6 {
			detailIVAl.StsByManager = iVal.StsByManager
			dtApp.LeaderName = usr[0].Fullname
			dtApp.DateLeave = iVal.DateLeave
			dtApp.Reason = iVal.ReasonAction
			dtApp.IdRequest = iVal.IdRequest
			dtApp.UserId = userid
			detailIVAl.StsByLeader = iVal.StsByLeader
			detailIVAl.StsByManager = iVal.StsByManager
			if iVal.StsByManager == "Approved" {
				// detailIVAl.IsDelete = false
				dtApp.IsApproved = true
				res = res + 1
				if res > usrt[0].YearLeave {
					return c.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
				} else if dataLeave.IsEmergency == true {
					if i == 0 {
						usrt[0].YearLeave = usrt[0].YearLeave - 1
						err = c.Ctx.Save(usrt[0])
						if err != nil {
							return c.SetResultInfo(true, err.Error(), nil)
						}
					}

				}
			} else if iVal.StsByManager == "Declined" {
				// detailIVAl.IsDelete = true
				dtApp.IsApproved = false
				res = res + 0
				datelistDeclined = append(datelistDeclined, dtApp.DateLeave)
			}
			ArrApp = append(ArrApp, dtApp)
		}

		err = c.Ctx.Save(detailIVAl)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
	}

	if jobrolelevel == 1 || jobrolelevel == 6 {
		to := []string{}
		to = append(to, leaveHeader.Email)

		// if jobrolelevel == 6 {
		for _, ky := range leaveHeader.BranchManager {
			if ky.UserId == userid {
				leaveHeader.StatusManagerProject.IdEmp = ky.IdEmp
				leaveHeader.StatusManagerProject.Name = ky.Name
				leaveHeader.StatusManagerProject.Location = ky.Location
				leaveHeader.StatusManagerProject.Email = ky.Email
				leaveHeader.StatusManagerProject.PhoneNumber = ky.PhoneNumber
				leaveHeader.StatusManagerProject.UserId = ky.UserId
			}
		}
		// }

		if userid == leaveHeader.StatusManagerProject.UserId {
			// tk.Println("------------- masuk sini")
			for _, vall := range leaveHeader.StatusProjectLeader {
				if vall.Email != "" {
					to = append(to, vall.Email)
				}

			}
			for _, mgr := range leaveHeader.ProjectManagerList {
				if mgr.Email != "" {
					to = append(to, mgr.Email)
				}

			}
			for _, ba := range leaveHeader.StatusBusinesAnalyst {
				if ba.Email != "" {
					to = append(to, ba.Email)
				}

			}
			to = append(to, leaveHeader.Email)
			if res >= 1 {
				leaveHeader.StatusManagerProject.StatusRequest = "Approved"
				leaveHeader.ResultRequest = "Approved"
				dash.SetHistoryLeave(k, leaveHeader.UserId, leaveHeader.Id, leaveHeader.LeaveFrom, leaveHeader.LeaveTo, "already approved by manager", "Approved", leaveHeader)

				emptyremote := new(RemoteModel)
				service := services.LogService{
					leaveHeader,
					emptyremote,
					"leave",
				}
				desc := "Request approved by manager"
				if len(datelistDeclined) > 0 {
					desc = "Request approved by manager with decline date list: " + strings.Join(datelistDeclined, ", ")
				}
				var stsReq = "Approved"
				log.Set("Status", stsReq)
				log.Set("Desc", desc)
				log.Set("NameLogBy", leaveHeader.StatusManagerProject.Name)
				log.Set("EmailNameLogBy", leaveHeader.StatusManagerProject.Email)
				err = service.ApproveDeclineLog(log)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
				// manager project list
				log2 := log
				log2.Set("Desc", "Request approved by PM")
				if len(dataLeave.ProjectManagerList) > 0 {
					log2.Set("NameLogBy", dataLeave.ProjectManagerList[0].Name)
					log2.Set("EmailNameLogBy", dataLeave.ProjectManagerList[0].Email)
				}
				err = service.ApproveDeclineLog(log2)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
			} else {
				leaveHeader.StatusManagerProject.StatusRequest = "Declined"
				leaveHeader.ResultRequest = "Declined"
				dash.SetHistoryLeave(k, leaveHeader.UserId, leaveHeader.Id, leaveHeader.LeaveFrom, leaveHeader.LeaveTo, "already declined by manager", "Declined", leaveHeader)

				emptyremote := new(RemoteModel)
				service := services.LogService{
					leaveHeader,
					emptyremote,
					"leave",
				}
				desc := "Request declined by manager"
				var stsReq = "Declined"
				log.Set("Status", stsReq)
				log.Set("Desc", desc)
				log.Set("NameLogBy", leaveHeader.StatusManagerProject.Name)
				log.Set("EmailNameLogBy", leaveHeader.StatusManagerProject.Email)
				err = service.ApproveDeclineLog(log)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
				// manager project list
				log2 := log
				log2.Set("Desc", "Request declined by PM")
				if len(dataLeave.ProjectManagerList) > 0 {
					log2.Set("NameLogBy", dataLeave.ProjectManagerList[0].Name)
					log2.Set("EmailNameLogBy", dataLeave.ProjectManagerList[0].Email)
				}
				err = service.ApproveDeclineLog(log2)
				if err != nil {
					return c.SetResultInfo(true, err.Error(), nil)
				}
			}
		}
		leaveHeader.DetailsLeave = ArrApp
		err = c.Ctx.Save(leaveHeader)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		// mail.SendMailUser(k, leaveHeader.Id, to)
		mail.SendMailLeaveDetails(k, to, leaveHeader.Name, leaveHeader.Reason, leaveHeader)
		// mail.RequestLeaveOnDateV2(leaveHeader, leaveHeader.UserId)
	} else if jobrolelevel == 2 {
		leaveHeader.DetailsLeave = ArrApp
		rest := 0
		for idx, val := range leaveHeader.StatusProjectLeader {
			if userid == val.UserId {
				tk.Println("----------- res ", res)
				if res >= 1 {
					leaveHeader.StatusProjectLeader[idx].StatusRequest = "Approved"
					dash.SetHistoryLeave(k, leaveHeader.UserId, leaveHeader.Id, leaveHeader.LeaveFrom, leaveHeader.LeaveTo, "already approved by Leader", "Pending", leaveHeader)
					emptyremote := new(RemoteModel)
					service := services.LogService{
						leaveHeader,
						emptyremote,
						"leave",
					}
					desc := "Request approved by Leader " + leaveHeader.StatusProjectLeader[idx].Name
					if len(datelistDeclined) > 0 {
						desc = "Request approved by Leader " + leaveHeader.StatusProjectLeader[idx].Name + " with decline date list: " + strings.Join(datelistDeclined, ", ")
					}
					var stsReq = "Approved"
					log.Set("Status", stsReq)
					log.Set("Desc", desc)
					log.Set("NameLogBy", leaveHeader.StatusProjectLeader[idx].Name)
					log.Set("EmailNameLogBy", leaveHeader.StatusProjectLeader[idx].Email)
					err = service.ApproveDeclineLog(log)
					if err != nil {
						return c.SetResultInfo(true, err.Error(), nil)
					}
				} else {
					leaveHeader.StatusProjectLeader[idx].StatusRequest = "Declined"
					dash.SetHistoryLeave(k, leaveHeader.UserId, leaveHeader.Id, leaveHeader.LeaveFrom, leaveHeader.LeaveTo, "already declined by Leader", "Pending", leaveHeader)
					emptyremote := new(RemoteModel)
					service := services.LogService{
						leaveHeader,
						emptyremote,
						"leave",
					}
					desc := "Request declined by Leader " + leaveHeader.StatusProjectLeader[idx].Name
					var stsReq = "Declined"
					log.Set("Status", stsReq)
					log.Set("Desc", desc)
					log.Set("NameLogBy", leaveHeader.StatusProjectLeader[idx].Name)
					log.Set("EmailNameLogBy", leaveHeader.StatusProjectLeader[idx].Email)
					err = service.ApproveDeclineLog(log)
					if err != nil {
						return c.SetResultInfo(true, err.Error(), nil)
					}
				}
			}
		}
		for _, led := range leaveHeader.StatusProjectLeader {
			if led.StatusRequest != "Pending" {
				rest = rest + 1
			}
		}

		if len(leaveHeader.StatusProjectLeader) == rest {
			mail.SendMailManagerDetails(k, leaveHeader.Name, leaveHeader.Reason, leaveHeader)
		}
		err = c.Ctx.Save(leaveHeader)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		// mail.SendMailManager(k, userid, leaveHeader.Id, 2)
		// mail.RequestLeaveOnDateV2(leaveHeader, leaveHeader.UserId)
	}
	// mail.RequestLeaveOnDateV2(leaveHeader, leaveHeader.UserId)

	return c.SetResultInfo(false, "save successfully", nil)
}

func (c *DeskAprovalController) ChangeStsDetailLeave(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payload := []struct {
		UserId   string
		YearVal  int
		MonthVal int
		DayVal   int
	}{}

	// jobrolelevel := k.Session("jobrolelevel").(int)
	// userid := k.Session("userid").(string)

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	for _, iVal := range payload {
		stsReq := StatusRequestController(*c)
		detailIVAl := stsReq.GetDataDetailLeaveByDate(k, iVal.UserId, iVal.YearVal, iVal.MonthVal, iVal.DayVal)
		detailIVAl.IsDelete = true
		err = c.Ctx.Save(detailIVAl)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

	}

	return c.SetResultInfo(false, "save successfully", nil)
}
