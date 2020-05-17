package services

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/creativelab/dbox"

	"strings"

	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

type RemoteService struct {
	Data       RemoteModel
	DateList   []string
	RemoteList []RemoteModel
}

func (s *RemoteService) RequestLeave(timezone string) (tk.M, error) {
	origindata := s.Data
	//check if has request before
	hasRequest := s.VerifyIsAlreadyHasRequest(s.Data.UserId)

	if hasRequest {
		return nil, errors.New("Your Already Has Request Remote Before")
	}
	//send email to leader
	projectsByLeader := map[string][]Project{}
	projectsByManager := map[string][]Project{}
	isProjectLeader := false
	isProjectManager := false
	isStaff := false
	idsProject := []string{}

	//get user details
	user, err := new(repositories.UserOrmRepo).GetByID(s.Data.UserId)
	if err != nil {
		return nil, err
	}

	if user.Designation == "4" && user.Designation == "7" {
		isStaff = true
	}

	for k, project := range s.Data.Projects {
		idsProject = append(idsProject, project.Id)

		s.Data.Projects[k].IsLeaderSend = false
		emailLeader := project.ProjectLeader.Email

		if emailLeader != "" {
			if _, ok := projectsByLeader[emailLeader]; !ok {
				projectsByLeader[emailLeader] = []Project{}
			}
		}

		projectsByLeader[emailLeader] = append(projectsByLeader[emailLeader], project)

		if user.Id == project.ProjectLeader.UserId {
			isProjectLeader = true
		} else if project.ProjectLeader.UserId == "" {
			isProjectLeader = true
		}

		if user.Id == project.ProjectManager.UserId {
			isProjectManager = true
			s.Data.Projects[k].IsManagerSend = false
			emailManager := project.ProjectManager.Email

			if _, ok := projectsByManager[emailManager]; !ok {
				projectsByManager[emailManager] = []Project{}
			}

			projectsByManager[emailManager] = append(projectsByManager[emailManager], project)
		}

	}

	s.Data.IdOp = bson.NewObjectId().Hex()
	exp, err := helper.ExpiredDateTime(timezone, false)
	rm, err := helper.ExpiredRemaining(timezone, false)
	s.Data.ExpiredOn = exp
	s.Data.ExpRemaining = rm
	s.Data.IsExpired = false
	tk.Println("-------------- data ", s.Data.IsExpired)
	dataRespon := s.ProcessSaveWithDateList(s.Data)
	if dataRespon.Get("RemoteSuccess") == nil {
		return dataRespon, errors.New("Your Already Remote on Date Your Plan")
	} else if len(dataRespon.Get("RemoteSuccess").([]RemoteModel)) == 0 {
		if dataRespon.Get("Check") == false {
			return dataRespon, errors.New("Remote only can request 10 days/month")
		} else {
			return dataRespon, errors.New("Your Already Remote on Date Your Plan")
		}
	}

	remoteData := dataRespon.Get("RemoteSuccess").([]RemoteModel)
	if len(remoteData) > 0 {
		s.Data = remoteData[0]
	}

	param := tk.M{}.Set("IdRequest", s.Data.IdOp).Set("IdProject", strings.Join(idsProject, ",")).Set("Status", "true")

	notif := new(HistoryService)
	notif.IDRequest = s.Data.IdOp
	notif.RequestType = "remote"
	notif.Name = s.Data.Name
	notif.StatusApproval = ""
	notif.Status = "Pending"
	notif.Reason = s.Data.Reason
	notif.UserId = s.Data.UserId
	// log
	logdataRemote := s.Data
	logdataRemote.From = origindata.From
	logdataRemote.To = origindata.To
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&logdataRemote,
		"remote",
	}
	err = service.RequestLog()
	if err != nil {
		return dataRespon, err
	}
	//
	if isStaff {
		notif.Desc = "send email to hrd"
		notif.Push(false)
		s.SendEmailForHrd()
	} else if isProjectManager {
		notif.Desc = "request approved by manager"
		param.Set("Type", "manager")
		notif.Push(false)
		s.HandleMailResponseFromManager(param, true)
	} else if isProjectLeader {
		notif.Desc = "send mail to manager"
		param.Set("Type", "leader")
		notif.Push(false)
		s.HandleMailResponseFromLeader(param)
	} else {
		notif.Desc = "send email to leader"
		notif.Push(false)
		s.SendEmailForLeader(s.Data, projectsByLeader)
	}

	return dataRespon, nil

}

func (s *RemoteService) ProcessCancel(payload tk.M, level int, usr string) error {
	emailSender := MailService{}
	emailSender.Init()

	// tk.Println("--------- payload ", payload)
	userid := payload.GetString("UserId")
	request := payload.GetString("IDRequest")
	datelist := payload.Get("DateList").([]interface{})
	reason := payload.GetString("Reason")
	datesString := []string{}
	// tk.Println("---------------- level ", level)
	for _, date := range datelist {
		dateleave := date.(string)
		datesString = append(datesString, dateleave)
	}

	remotes, err := s.GetRemoteByIDRequestDateList(userid, request, datesString, level)
	// tk.Println("--------- remotes ", remotes)
	if err != nil {
		return err
	}
	// tk.Println("---------------- level1 ", level)
	user, err := new(repositories.UserOrmRepo).GetByID(userid)
	if err != nil {
		return err
	}
	// tk.Println("---------------- level2 ", level)
	dateCreate := ""
	typeRemote := ""
	repoOrm := new(repositories.RemoteOrmRepo)
	// tk.Println("---------------- remotes ", remotes[0].Name)
	for _, remote := range remotes {
		dateCreate = (remote.CreatedAt).Format("2006-01-02 15:04:05")
		typeRemote = remote.Type
		remote.IsRequestChange = true
		remote.ReasonAction = reason
		// tk.Println("---------------- level ", level)
		if level == 5 || level == 1 {
			remote.IsDelete = true
		} else if level == 6 {
			remote.IsDelete = true
		}
		err := repoOrm.Save(&remote)
		if err != nil {
			return err
		}

		layout := "2006-01-02"
		t, _ := time.Parse(layout, remote.DateLeave)

		//log
		emptyleave := new(RequestLeaveModel)
		service := LogService{
			emptyleave,
			&remote,
			"remote",
		}
		log := tk.M{}
		if level == 5 || level == 1 || level == 6 {
			log.Set("Status", "Cancel")
			log.Set("Desc", "Request Remote Canceled"+" for date "+t.Format("02-Jan-2006"))
		} else {
			log.Set("Status", "Request Cancel")
			log.Set("Desc", "Employee Request Cancel Remote"+" for date "+t.Format("02-Jan-2006"))
		}

		log.Set("NameLogBy", user.Fullname)
		log.Set("EmailNameLogBy", user.Email)
		err = service.CancelRequest(log)
		if err != nil {
			return err
		}
	}

	// fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	fmt.Println("---------------", dateCreate)

	paramForUrlDecline := map[string]string{
		"UserId":    userid,
		"IDRequest": request,
		"DateList":  strings.Join(datesString, ","),
		"Status":    "false",
		"Type":      "cancelremote",
		"Note":      "",
	}

	paramForUrlApproval := map[string]string{
		"UserId":    userid,
		"IDRequest": request,
		"DateList":  strings.Join(datesString, ","),
		"Status":    "true",
		"Type":      "cancelremote",
		"Note":      "",
	}
	tp := ""
	switch typeRemote {
	case "1":
		tp = "Conditional"
	case "2":
		tp = "Monthly"
	case "3":
		tp = "Full Monthly"
	}

	urlDecline, _ := s.HandleUrlDecline(paramForUrlDecline, "/remote/handleapproval")
	urlApproval, _ := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")

	fileParam := MailFileParamRequest{
		dateCreate,
		"",
		remotes[0].Name,
		reason,
		tp,
		[]string{},
		"",
		"",
		datesString,
		nil,
		urlDecline,
		urlApproval,
		usr,
		"",
		"",
		"",
	}

	conf := helper.ReadConfig()
	hrd := conf.GetString("HrdMail")
	emailSender.From = emailSender.Conf.EmailOperator
	if level == 2 || level == 3 {

		emailSender.MailSubject = "Employee Request Cancel Remote"
		emailSender.To = []string{hrd}
		emailSender.Filename = "remotecancel.html"
	} else {

		emailSender.MailSubject = "Request Remote Canceled"
		emailSender.To = []string{user.Email, hrd, remotes[0].Email}
		emailSender.Filename = "remotecancelbyadmin.html"

	}

	emailSender.FileParamRequest = fileParam
	err = emailSender.SendEmail()
	fmt.Println("--->>>> log email", err)

	return err
}

func (s *RemoteService) GetRemoteByIDRequestDateList(userid string, idrequest string, listdate []string, level int) ([]RemoteModel, error) {
	prematch := tk.M{}
	if level == 5 || level == 6 || level == 1 {
		prematch.Set("idop", idrequest).Set("dateleave", tk.M{}.Set("$in", listdate))
	} else {
		prematch.Set("idop", idrequest).Set("userid", userid).Set("dateleave", tk.M{}.Set("$in", listdate))
	}

	match := tk.M{}.Set("$match", prematch)

	return new(repositories.RemoteDboxRepo).GetByPipe([]tk.M{match})
}
func (s *RemoteService) SendEmailForLeader(remote RemoteModel, projectsData map[string][]Project) error {
	emailSender := MailService{}
	emailSender.Init()

	for k, projects := range projectsData {
		projectNames := []string{}
		nameLeader := ""
		nameManager := ""
		idsProject := []string{}
		// dateRequest := (remote.CreatedAt).Format("2006-01-02 15:04:05")
		for _, project := range projects {
			projectNames = append(projectNames, project.ProjectName)
			idsProject = append(idsProject, project.Id)
			nameLeader = project.ProjectLeader.Name
			nameManager = project.ProjectManager.Name
		}

		paramForUrlDecline := map[string]string{
			"IdRequest": remote.IdOp,
			"IdProject": strings.Join(idsProject, ","),
			"Type":      "leader",
			"Status":    "false",
		}

		paramForUrlApproval := map[string]string{
			"IdRequest": remote.IdOp,
			"IdProject": strings.Join(idsProject, ","),
			"Type":      "leader",
			"Status":    "true",
		}

		urlDecline, _ := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
		urlApproval, _ := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")

		fmt.Println("--------------- leader", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
		tyeRemote := s.convertTypetoTextString(remote.Type)
		fileParam := MailFileParamRequest{
			(remote.CreatedAt).Format("2006-01-02 15:04:05"),
			nameLeader,
			remote.Name,
			remote.Reason,
			tyeRemote,
			projectNames,
			remote.From,
			remote.To,
			s.DateList,
			nil,
			urlDecline,
			urlApproval,
			nameManager,
			"",
			"",
			"",
		}

		emailSender.MailSubject = remote.Name + " is Requesting for Remote Work " + time.Now().Format("January")
		emailSender.From = emailSender.Conf.EmailOperator
		emailSender.To = []string{k}
		emailSender.Filename = "remoterequest.html"
		emailSender.FileParamRequest = fileParam
		err := emailSender.SendEmail()
		fmt.Println("--->>>> log email", err)
	}

	return nil
}

func (s *RemoteService) ValidateParamMail(paramView tk.M, usr string) (tk.M, error) {
	paramenc := paramView.GetString("Param")
	param, err := s.DecodeParam(paramenc)
	if err != nil {
		return param, err
	}

	tk.Println("------------------- param validate ", param)
	param.Set("Note", paramView.GetString("Note"))
	if param.GetString("Type") == "leader" {
		remotes, err := s.HandleMailResponseFromLeader(param)
		if err != nil {
			return param, err
		}
		for _, each := range remotes {
			if each.IsExpired {
				param.Set("IsExpired", true)
				break
			}
		}
		if param.Get("IsExpired") == nil {
			param.Set("IsExpired", false)
			param.Set("ApprovalName", remotes[0].Projects[0].ProjectLeader.Name)
		}
	} else if param.GetString("Type") == "manager" {
		remotes, err := s.HandleMailResponseFromManager(param, false)
		if err != nil {
			return param, err
		}
		for _, each := range remotes {
			if each.IsExpired {
				param.Set("IsExpired", true)
				break
			}
		}
		if param.Get("IsExpired") == nil {
			param.Set("IsExpired", false)
			param.Set("ApprovalName", remotes[0].Projects[0].ProjectManager.Name)
		}
	} else if param.GetString("Type") == "hrd" {
		s.HandleEmailResponseFromHRD(param)
	} else if param.GetString("Type") == "cancelremote" {
		_, err := s.HandleMailCancelRemote(param, usr)
		if err != nil {
			return param, err
		}
	}

	return param, nil
}

func (s *RemoteService) ValidateParamFromApp(params []tk.M) (interface{}, error) {
	remotes := []RemoteModel{}

	for _, param := range params {
		typeOfLeave := param.GetString("Type")
		if typeOfLeave == "manager" {
			remote, err := s.HandleMailResponseFromManager(param, false)
			if err != nil {
				return remotes, err
			}

			remotes = append(remotes, remote...)
		} else if typeOfLeave == "leader" {
			remote, err := s.HandleMailResponseFromLeader(param)
			if err != nil {
				return remotes, err
			}

			remotes = append(remotes, remote...)
		} else if typeOfLeave == "cancel" {
			remote, err := s.HandleMailCancelRemoteFromApp(param)
			if err != nil {
				tk.Println("------------------- masuk eror ", err)
				return remotes, err
			}

			remotes = append(remotes, remote...)
		} else {
			return nil, errors.New("Type Invalid")
		}
	}
	//tk.Println("iiiii", remotes)
	return remotes, nil
}

func (s *RemoteService) ValidateParamFromAppDetail(action string, remotes []RemoteModel, dateList []string) error {
	// match := tk.M{}.Set("$match", tk.M{}.Set("idop", remotes[0].IdOp).Set("userid", remotes[0].UserId).Set("dateleave", tk.M{}.Set("$in", dateList)))
	// pipe := []tk.M{match}

	// remotesDB, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	// if err != nil {
	// 	return err
	// }

	// for k, remote := range remotes {
	// 	for _, remoteDB := range remotesDB {
	// 		fmt.Println("sarifff", remoteDB.Id)
	// 		if remote.UserId == remoteDB.UserId && remote.IdOp == remoteDB.IdOp && remote.DateLeave == remoteDB.DateLeave {
	// 			remotes[k].Id = remoteDB.Id
	// 			fmt.Println("renooon", remotes[k].Id)
	// 		}
	// 	}
	// }

	if action == "leader" {
		_, err := s.HandleApprovalDetailLeader(remotes)
		return err
	} else if action == "manager" {
		_, err := s.HandleApprovalDetailManager(remotes)
		if err != nil {
			return err
		}
	} else if action == "cancel" {
		_, err := s.HandleApprovalDetailCancel(remotes)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *RemoteService) HandleApprovalDetailLeader(remotes []RemoteModel) (interface{}, error) {
	repoM := new(repositories.RemoteOrmRepo)
	managerWithProject := map[string][]Project{}
	approvedManager := false
	//check request has approval by manager
	remoteManagers, _ := s.ValidateRequestHasSendManager(remotes[0].IdOp)
	if len(remoteManagers) > 0 {
		approvedManager = true
	}
	//end
	remoteTemp := RemoteModel{}
	dateList := []string{}
	leaderName := ""
	emailLeader := ""
	appLeader := false
	declined := true
	datelistDeline := []string{}
	branchmanager := []BranchManagerList{}
	note := ""
	for _, remote := range remotes {
		for _, project := range remote.Projects {
			leaderName = project.ProjectLeader.Name
			emailLeader = project.ProjectLeader.Email
			appLeader = project.IsApprovalLeader
			projectManagerEmail := project.ProjectManager.Email
			if _, ok := managerWithProject[projectManagerEmail]; !ok {
				managerWithProject[projectManagerEmail] = []Project{}
			}
			managerWithProject[projectManagerEmail] = append(managerWithProject[projectManagerEmail], project)
			if project.NoteLeader != "" {
				note = project.NoteLeader
			}
		}
		branchmanager = remote.BranchManager
		dateList = append(dateList, remote.DateLeave)

		if appLeader {
			declined = false
		} else {
			datelistDeline = append(datelistDeline, remote.DateLeave)
			remote.IsDelete = true
		}
		remoteTemp = remote
		err := repoM.Save(&remote)
		if err != nil {
			return remotes, err
		}
	}

	notif := new(HistoryService)
	notif.IDRequest = remoteTemp.IdOp
	notif.Name = remoteTemp.Name
	notif.RequestType = "remote"
	notif.Status = "Pending"
	notif.StatusApproval = ""
	notif.UserId = remoteTemp.UserId
	notif.Reason = remoteTemp.Reason

	s.RemoteList = remotes
	s.DateList = dateList

	if !approvedManager {
		if !declined {
			notif.Desc = "send to manager"
			err := s.SendEmailForManagerDetails(remoteTemp, managerWithProject, branchmanager)
			if err != nil {
				return remotes, err
			}
			notif.Push(false)
		} else {
			param := tk.M{
				"IdRequest": remoteTemp.IdOp,
				"IdProject": remoteTemp.Projects[0].Id,
				"Type":      "leader",
				"Status":    "false",
				"Note":      note,
			}

			notif.Desc = "Declined by leader " + leaderName
			err := s.SendEmailDeclineLeader(remoteTemp, param, leaderName, emailLeader)
			if err != nil {
				return remotes, err
			}
			notif.Push(false)
		}
	}

	//log
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&remoteTemp,
		"remote",
	}
	log := tk.M{}
	if !declined {
		log.Set("Status", "Approved")
		log.Set("Desc", "Request Approved by Leader")
		if len(datelistDeline) > 0 {
			log.Set("Desc", "Request Approved by Leader with decline date list : "+strings.Join(datelistDeline, ", "))
		}
	} else {
		log.Set("Status", "Declined")
		log.Set("Desc", "Request Declined by Leader")
	}
	log.Set("NameLogBy", leaderName)
	log.Set("EmailNameLogBy", emailLeader)
	err := service.ApproveDeclineLog(log)
	if err != nil {
		return remotes, err
	}

	return nil, nil
}

func (s *RemoteService) SendEmailForManagerDetails(remote RemoteModel, managers map[string][]Project, branchmanager []BranchManagerList) error {
	emailSender := MailService{}
	emailSender.Init()
	noteDecline := ""
	dateListObject := []DetailDate{}
	for _, remote := range s.RemoteList {
		dateRow := DetailDate{}
		dateRow.DateLeave = remote.DateLeave
		for _, project := range remote.Projects {
			dateRow.Note = project.NoteLeader
			if project.IsApprovalLeader {
				dateRow.Status = "Approval"
			} else {
				dateRow.Status = "Declined"
				if project.NoteLeader != "" {
					noteDecline = project.NoteLeader
				}
			}

			dateListObject = append(dateListObject, dateRow)
		}
	}

	// for manager, projects := range managers {
	for _, projects := range managers {
		projectNames := []string{}
		idsProject := []string{}
		projectName := ""
		// managerName := ""
		validForProject := map[string]int{}

		for _, project := range projects {
			projectName := project.ProjectName
			if _, ok := validForProject[projectName]; !ok {
				projectNames = append(projectNames, project.ProjectName)
				validForProject[projectName] = 1
			}

			idsProject = append(idsProject, project.Id)
			projectName = project.ProjectLeader.Name
			// managerName = project.ProjectManager.Name
		}
		for _, mgr := range branchmanager {
			paramForUrlDecline := map[string]string{
				"IdRequest":     remote.IdOp,
				"IdProject":     strings.Join(idsProject, ","),
				"Type":          "manager",
				"Status":        "false",
				"ManagerUserId": mgr.UserId,
			}

			paramForUrlApproval := map[string]string{
				"IdRequest":     remote.IdOp,
				"IdProject":     strings.Join(idsProject, ","),
				"Type":          "manager",
				"Status":        "true",
				"ManagerUserId": mgr.UserId,
			}

			urlDecline, err := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
			if err != nil {
				return err
			}

			urlApproval, err := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")
			if err != nil {
				return err
			}
			fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
			fileParam := MailFileParamRequest{
				(remote.CreatedAt).Format("2006-01-02 15:04:05"),
				projectName,
				remote.Name,
				remote.Reason,
				remote.Type,
				projectNames,
				remote.From,
				remote.To,
				s.DateList,
				dateListObject,
				urlDecline,
				urlApproval,
				mgr.Name,
				"",
				noteDecline,
				"",
			}

			emailSender.MailSubject = "Confirmation Remote From Leader"
			emailSender.From = emailSender.Conf.EmailOperator
			emailSender.To = []string{mgr.Email}
			emailSender.Filename = "remotemanagerdetail.html"
			emailSender.FileParamRequest = fileParam
			err = emailSender.SendEmail()

			if err != nil {
				fmt.Println("--->>>> log email", err)
				return err
			}
		}
	}

	return nil
}

func (s *RemoteService) HandleApprovalDetailManager(remotes []RemoteModel) (interface{}, error) {
	repoM := new(repositories.RemoteOrmRepo)
	projectLeaderList := map[string][]Project{}

	if s.ValidateIsManagerSend(remotes) == true {
		return remotes, errors.New("Request already processed")
	}

	remoteTemp := RemoteModel{}
	dateList := []string{}
	status := "false"
	managerName := ""
	emailManager := ""
	appManager := false
	declined := true
	datelistDeline := []string{}
	// branchmgrs := []BranchManagerList{}
	for _, remote := range remotes {
		// branchmgrs = remote.BranchManager

		dateList = append(dateList, remote.DateLeave)

		for _, project := range remote.Projects {

			projectLeaderEmail := project.ProjectLeader.Email
			if _, ok := projectLeaderList[projectLeaderEmail]; !ok {
				projectLeaderList[projectLeaderEmail] = []Project{}
			}

			projectLeaderList[projectLeaderEmail] = append(projectLeaderList[projectLeaderEmail], project)
			remoteTemp = remote
			appManager = project.IsApprovalManager
			if project.IsApprovalManager {
				status = "true"
			}
			managerName = project.ProjectManager.Name
			emailManager = project.ProjectManager.Email
		}

		if appManager {
			declined = false
		} else {
			datelistDeline = append(datelistDeline, remote.DateLeave)
			remote.IsDelete = true
		}
		err := repoM.Save(&remote)
		if err != nil {
			return remotes, err
		}
	}

	notif := new(HistoryService)
	notif.IDRequest = remoteTemp.IdOp
	notif.Name = remoteTemp.Name
	notif.RequestType = "remote"
	notif.UserId = remoteTemp.UserId
	notif.Reason = remoteTemp.Reason
	notif.ManagerApprove = managerName

	s.DateList = dateList
	s.RemoteList = remotes

	s.SendEmailAfterManagerDetail(remoteTemp, projectLeaderList)
	if status == "true" {
		notif.Status = "Approved"
		notif.StatusApproval = "Approved"
		notif.Desc = "already approved by manager"
	} else {
		notif.Status = "Decline"
		notif.StatusApproval = "Decline"
		notif.Desc = "already decline by manager"
	}

	notif.Push(false)

	//log
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&remoteTemp,
		"remote",
	}
	log := tk.M{}
	log.Set("Status", notif.StatusApproval)
	log.Set("Desc", notif.Desc)
	if !declined {
		log.Set("Status", "Approved")
		log.Set("Desc", "Request Approved by Manager")
		if len(datelistDeline) > 0 {
			log.Set("Desc", "Request Approved by Manager with decline date list : "+strings.Join(datelistDeline, ", "))
		}
	} else {
		log.Set("Status", "Declined")
		log.Set("Desc", "Request Declined by Manager")
	}
	log.Set("NameLogBy", managerName)
	log.Set("EmailNameLogBy", emailManager)
	err := service.ApproveDeclineLog(log)
	if err != nil {
		return remotes, err
	}

	return remotes, nil
}

func (s *RemoteService) HandleApprovalDetailCancel(remotes []RemoteModel) (interface{}, error) {
	repoM := new(repositories.RemoteOrmRepo)

	apprRemote := 0
	decRemote := 0
	typeRemoteNum := ""
	reasons := ""
	usrMail := ""
	usrName := ""
	dateCreate := ""
	datelist := []string{}
	for _, remote := range remotes {
		if remote.IsDelete {
			apprRemote++
			datelist = append(datelist, remote.DateLeave)
		} else {
			decRemote++
		}

		dateCreate = (remote.CreatedAt).Format("2006-01-02 15:04:05")
		typeRemoteNum = remote.Type
		reasons = remote.ReasonAction
		usrMail = remote.Email
		usrName = remote.Name

		err := repoM.Save(&remote)
		if err != nil {
			return remotes, err
		}

		layout := "2006-01-02"
		t, _ := time.Parse(layout, remote.DateLeave)

		//log
		emptyleave := new(RequestLeaveModel)
		service := LogService{
			emptyleave,
			&remote,
			"remote",
		}
		log := tk.M{}
		if remote.IsDelete {
			log.Set("Status", "Cancel")
			log.Set("Desc", "Request Remote is Canceled"+" for date "+t.Format("02-Jan-2006"))
		} else {
			log.Set("Status", "Cancel")
			log.Set("Desc", "Request Remote is Declined "+" for date "+t.Format("02-Jan-2006"))
		}

		err = service.CancelRequest(log)
		if err != nil {
			return remotes, err
		}
	}

	statusMail := ""
	if apprRemote >= decRemote {
		statusMail = "Approved"
	} else {
		statusMail = "Declined"
	}

	typeRemote := s.convertTypetoTextString(typeRemoteNum)
	fileParam := MailFileParamRequest{
		dateCreate,
		"",
		usrName,
		reasons,
		typeRemote,
		[]string{},
		"",
		"",
		datelist,
		nil,
		"",
		"",
		"",
		statusMail,
		"",
		"",
	}

	conf := helper.ReadConfig()
	emails := []string{usrMail, conf.GetString("HrdMail")}
	//this template is changged. This template now sending to 2 emails (hrd and user) =====
	err := s.SendEmailConfirmation(emails, "remotecanceluniversal.html", fileParam)
	if err != nil {
		return remotes, err
	}

	return remotes, nil
}

func (s *RemoteService) SendEmailAfterManagerDetail(remote RemoteModel, leaders map[string][]Project) error {
	emailSender := MailService{}
	emailSender.Init()

	//details user
	user, err := new(repositories.UserOrmRepo).GetByID(remote.UserId)
	if err != nil {
		return err
	}
	appr := 0
	noteDecline := ""
	datelist := []string{}
	dateListObject := []DetailDate{}
	for _, remote := range s.RemoteList {
		dateRow := DetailDate{}
		dateRow.DateLeave = remote.DateLeave
		for _, project := range remote.Projects {
			dateRow.Note = project.NoteLeader
			if project.IsApprovalManager {
				dateRow.Status = "Approval"
				dateRow.Note = project.NoteManager
				datelist = append(datelist, dateRow.DateLeave)

				appr = appr + 1
			} else {
				dateRow.Status = "Declined"
				dateRow.Note = project.NoteManager
				if project.NoteManager != "" {
					noteDecline = project.NoteManager
				}

			}

			dateListObject = append(dateListObject, dateRow)
		}
	}

	listSender := []string{}
	validForProject := map[string]int{}
	projectNames := []string{}
	managerName := ""
	// managerEmail := ""

	for leader, projects := range leaders {
		listSender = append(listSender, leader)
		for _, project := range projects {
			projectName := project.ProjectName
			if _, ok := validForProject[projectName]; !ok {
				projectNames = append(projectNames, project.ProjectName)
				validForProject[projectName] = 1
			}

			projectName = project.ProjectLeader.Name
			managerName = project.ProjectManager.Name
			// managerEmail = project.ProjectManager.Email
		}
	}

	conf := helper.ReadConfig()
	listSender = append(listSender, conf.GetString("HrdMail"))
	for _, mgr := range remote.BranchManager {
		listSender = append(listSender, mgr.Email)
	}
	// listSender = append(listSender, managerEmail)

	// hrd, err := s.GetHrOne()

	// if err == nil {
	// 	if hrd.AccountManager.Email != "" {
	// 		listSender = append(listSender, hrd.AccountManager.Email)
	// 	}
	// }

	if user.Email != "" {
		listSender = append(listSender, user.Email)
	}
	fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	strApp := strconv.Itoa(appr)
	fileParam := MailFileParamRequest{
		(remote.CreatedAt).Format("2006-01-02 15:04:05"),
		"",
		remote.Name,
		remote.Reason,
		remote.Type,
		projectNames,
		remote.From,
		remote.To,
		s.DateList,
		dateListObject,
		"",
		"",
		managerName,
		"",
		noteDecline,
		strApp,
	}

	tk.Println("jsonString...")
	tk.Println(tk.JsonString(fileParam.DetailsDate))
	tk.Println("jsonString...2")
	tk.Println(fileParam.DetailsDate)
	tk.Println("jsonString...3")
	tk.Println(tk.JsonString(fileParam))
	err = s.RemoveLeaveByDates(remote.UserId, datelist)
	if err != nil {
		return err
	}

	emailSender.MailSubject = "Confirmation Remote From Manager"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.Filename = "remoteresponmanagerdetail.html"
	emailSender.FileParamRequest = fileParam
	emailSender.To = listSender
	err = emailSender.SendEmail()
	if err != nil {
		return err
	}
	// for _, email := range listSender {
	// 	emailSender.To = []string{email}
	// 	err := emailSender.SendEmail()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (s *RemoteService) ValidateRequestHasSendManager(idRequest string) ([]RemoteModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("idop", idRequest))
	dboxFilter = append(dboxFilter, dbox.Eq("projects.ismanagersend", true))

	filter.Set("where", dbox.And(dboxFilter...))

	return new(repositories.RemoteOrmRepo).GetByParam(filter)
}

func (s *RemoteService) HandleMailResponseFromLeader(param tk.M) ([]RemoteModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	idTemp := param.GetString("IdRequest")
	approvedManager := false
	//check request has approval by manager
	remoteManagers, _ := s.ValidateRequestHasSendManager(idTemp)
	if len(remoteManagers) > 0 {
		approvedManager = true
	}
	//end

	dboxFilter = append(dboxFilter, dbox.Eq("idop", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	repoM := new(repositories.RemoteOrmRepo)
	remotes, err := repoM.GetByParam(filter)
	if err != nil {
		return remotes, err
	}
	for _, each := range remotes {
		if each.IsExpired {
			return remotes, nil
		}
	}
	status := param.GetString("Status")
	typeRespon := param.GetString("Type")
	idProjects := strings.Split(param.GetString("IdProject"), ",")
	note := param.GetString("Note")

	managerWithProject := map[string][]Project{}
	remoteTemp := RemoteModel{}

	dateList := []string{}
	leaderName := ""
	emailLeader := ""
	branchmanager := []BranchManagerList{}
	for _, remote := range remotes {
		for k, project := range remote.Projects {
			leaderName = project.ProjectLeader.Name
			emailLeader = project.ProjectLeader.Email
			for _, idproject := range idProjects {
				if remote.Projects[k].IsLeaderSend {
					param["Status"] = true
					return remotes, err
				}

				if project.Id == idproject {
					dateList = append(dateList, remote.DateLeave)

					projectManagerEmail := project.ProjectManager.Email
					if _, ok := managerWithProject[projectManagerEmail]; !ok {
						managerWithProject[projectManagerEmail] = []Project{}
					}

					managerWithProject[projectManagerEmail] = append(managerWithProject[projectManagerEmail], project)
					if typeRespon == "leader" {
						if status == "true" || approvedManager {
							remote.Projects[k].IsApprovalLeader = true
							remote.Projects[k].IsLeaderSend = true
						} else {
							remote.Projects[k].IsApprovalLeader = false
							remote.Projects[k].IsLeaderSend = true
							remote.Projects[k].NoteLeader = note
							remote.IsDelete = true
						}

					}

					remoteTemp = remote
					break
				}
			}
		}
		branchmanager = remote.BranchManager
		err = repoM.Save(&remote)
		if err != nil {
			return remotes, err
		}
	}

	notif := new(HistoryService)
	notif.IDRequest = remoteTemp.IdOp
	notif.Name = remoteTemp.Name
	notif.RequestType = "remote"
	notif.Status = "Pending"
	notif.StatusApproval = ""
	notif.UserId = remoteTemp.UserId
	notif.Reason = remoteTemp.Reason

	s.DateList = dateList
	if !approvedManager {
		if status == "true" {
			notif.Desc = "send to manager"
			err = s.SendEmailForManager(remoteTemp, param, managerWithProject, branchmanager)
			if err != nil {
				return remotes, err
			}
			notif.Push(false)
		} else {
			notif.Desc = "Declined by leader " + leaderName
			err = s.SendEmailDeclineLeader(remoteTemp, param, leaderName, emailLeader)
			if err != nil {
				return remotes, err
			}
			notif.Push(false)
		}
	}
	//log
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&remoteTemp,
		"remote",
	}
	log := tk.M{}
	if status == "true" {
		log.Set("Status", "Approved")
		log.Set("Desc", "Request Approved by Leader")
	} else {
		log.Set("Status", "Declined")
		log.Set("Desc", "Request Declined by Leader")
	}
	log.Set("NameLogBy", leaderName)
	log.Set("EmailNameLogBy", emailLeader)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		return remotes, err
	}
	param["Status"] = false
	return remotes, err
}

func (s *RemoteService) SendEmailForManager(remote RemoteModel, param tk.M, managers map[string][]Project, branchmanager []BranchManagerList) error {
	emailSender := MailService{}
	emailSender.Init()

	idRequest := param.GetString("IdRequest")
	status := param.GetString("Status")
	note := param.GetString("Note")

	if status == "true" {
		status = "Approve"
	} else {
		status = "Decline"
	}

	validForProject := map[string]int{}

	// for manager, projects := range managers {
	for _, projects := range managers {
		projectNames := []string{}
		idsProject := []string{}
		projectName := ""
		// managerName := ""
		for _, project := range projects {
			projectName := project.ProjectName
			if _, ok := validForProject[projectName]; !ok {
				projectNames = append(projectNames, project.ProjectName)
				validForProject[projectName] = 1
			}

			idsProject = append(idsProject, project.Id)
			projectName = project.ProjectLeader.Name
			// managerName = project.ProjectManager.Name
		}
		fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
		for _, mgr := range branchmanager {
			paramForUrlDecline := map[string]string{
				"IdRequest":     idRequest,
				"IdProject":     strings.Join(idsProject, ","),
				"Type":          "manager",
				"Status":        "false",
				"ManagerUserId": mgr.UserId,
			}

			paramForUrlApproval := map[string]string{
				"IdRequest":     idRequest,
				"IdProject":     strings.Join(idsProject, ","),
				"Type":          "manager",
				"Status":        "true",
				"ManagerUserId": mgr.UserId,
			}

			urlDecline, err := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
			if err != nil {
				return err
			}

			urlApproval, err := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")
			if err != nil {
				return err
			}

			fileParam := MailFileParamRequest{
				(remote.CreatedAt).Format("2006-01-02 15:04:05"),
				projectName,
				remote.Name,
				remote.Reason,
				remote.Type,
				projectNames,
				remote.From,
				remote.To,
				s.DateList,
				nil,
				urlDecline,
				urlApproval,
				mgr.Name,
				status,
				note,
				"",
			}

			emailSender.MailSubject = "Confirmation Remote From Leader"
			emailSender.From = emailSender.Conf.EmailOperator
			// emailSender.To = []string{manager}
			emailSender.To = []string{mgr.Email}
			emailSender.Filename = "remotemanager.html"
			emailSender.FileParamRequest = fileParam
			err = emailSender.SendEmail()

			if err != nil {
				fmt.Println("--->>>> log email", err)
				return err
			}
		}

	}

	return nil
}

func (s *RemoteService) SendEmailDeclineLeader(remote RemoteModel, param tk.M, leaderName string, leaderEmail string) error {
	emailSender := MailService{}
	emailSender.Init()

	// idRequest := param.GetString("IdRequest")
	status := param.GetString("Status")
	note := param.GetString("Note")

	if status == "true" {
		status = "Approve"
	} else {
		status = "Decline"
	}

	// projectNames := []string{}
	// idsProject := []string{}
	// projectName := ""
	// leaderName := ""
	// for _, project := range projects {
	// 	projectName := project.ProjectName
	// 	if _, ok := validForProject[projectName]; !ok {
	// 		projectNames = append(projectNames, project.ProjectName)
	// 		validForProject[projectName] = 1
	// 	}

	// 	idsProject = append(idsProject, project.Id)
	// 	projectName = project.ProjectLeader.Name
	// 	leaderName = project.ProjectManager.Name
	// }

	// paramForUrlDecline := map[string]string{
	// 	"IdRequest": idRequest,
	// 	"IdProject": strings.Join(idsProject, ","),
	// 	"Type":      "manager",
	// 	"Status":    "false",
	// }

	// paramForUrlApproval := map[string]string{
	// 	"IdRequest": idRequest,
	// 	"IdProject": strings.Join(idsProject, ","),
	// 	"Type":      "manager",
	// 	"Status":    "true",
	// }

	// urlDecline, err := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
	// if err != nil {
	// 	return err
	// }

	// urlApproval, err := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")
	// if err != nil {
	// 	return err
	// }
	remotetype := ""
	switch remote.Type {
	case "1":
		remotetype = "Conditional"
	case "2":
		remotetype = "Monthly"
	case "3":
		remotetype = "Full Monthly"
	}
	project := []string{""}
	fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	fileParam := MailFileParamRequest{
		(remote.CreatedAt).Format("2006-01-02 15:04:05"),
		leaderName,
		remote.Name,
		remote.Reason,
		remotetype,
		project,
		remote.From,
		remote.To,
		s.DateList,
		nil,
		"",
		"",
		"",
		status,
		note,
		"",
	}

	emailSender.MailSubject = "Request Decline by Leader"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.To = []string{leaderEmail}
	emailSender.Filename = "remotedeclineleader.html"
	emailSender.FileParamRequest = fileParam
	err := emailSender.SendEmail()

	if err != nil {
		fmt.Println("--->>>> log email", err)
		return err
	}
	return nil
}

func (s *RemoteService) HandleMailResponseFromManager(param tk.M, isProjectManager bool) ([]RemoteModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}

	idTemp := param.GetString("IdRequest")
	dboxFilter = append(dboxFilter, dbox.Eq("idop", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	remoteTemp := RemoteModel{}

	repoM := new(repositories.RemoteOrmRepo)
	remotes, err := repoM.GetByParam(filter)

	if s.ValidateIsManagerSend(remotes) == true {
		return remotes, errors.New("Request already processed")
	}

	if err != nil {
		return remotes, err
	}

	for _, each := range remotes {
		if each.IsExpired {
			tk.Println("--------------- masuk 123 expired ", each.IsExpired)
			return remotes, nil
		}
	}
	status := param.GetString("Status")
	typeRespon := param.GetString("Type")
	idProjects := strings.Split(param.GetString("IdProject"), ",")
	note := param.GetString("Note")
	managerName := ""
	emailManager := ""
	projectLeaderList := map[string][]Project{}

	dateList := []string{}
	branchmgr := BranchManagerList{}
	branchmgrs := []BranchManagerList{}
	for _, remote := range remotes {
		branchmgrs = remote.BranchManager
		for _, mgr := range remote.BranchManager {
			if mgr.UserId == param.GetString("ManagerUserId") {
				branchmgr = mgr
			}
		}
		for k, project := range remote.Projects {
			//set new manager
			if branchmgr.UserId != "" {
				remote.Projects[k].ProjectManager.UserId = branchmgr.UserId
				remote.Projects[k].ProjectManager.Email = branchmgr.Email
				remote.Projects[k].ProjectManager.Name = branchmgr.Name
				remote.Projects[k].ProjectManager.IdEmp = branchmgr.IdEmp
				remote.Projects[k].ProjectManager.PhoneNumber = branchmgr.PhoneNumber
				remote.Projects[k].ProjectManager.Location = branchmgr.Location
			}
			for _, idproject := range idProjects {
				if project.Id == idproject {
					dateList = append(dateList, remote.DateLeave)

					if remote.Projects[k].IsManagerSend {
						tk.Println("--------------- remote.Projects[k].IsManagerSend ", tk.JsonString(remote))
						return remotes, err
					}

					projectLeaderEmail := project.ProjectLeader.Email
					if projectLeaderEmail != "" {
						if _, ok := projectLeaderList[projectLeaderEmail]; !ok {
							projectLeaderList[projectLeaderEmail] = []Project{}
						}
					}

					managerName = remote.Projects[k].ProjectManager.Name
					emailManager = remote.Projects[k].ProjectManager.Email
					if projectLeaderEmail != "" {
						projectLeaderList[projectLeaderEmail] = append(projectLeaderList[projectLeaderEmail], remote.Projects[k])
					}
					// projectLeaderList[projectLeaderEmail] = append(projectLeaderList[projectLeaderEmail], remote.Projects[k])
					if typeRespon == "manager" {
						if status == "true" {
							remote.Projects[k].IsApprovalManager = true
							remote.Projects[k].IsManagerSend = true
							if isProjectManager {
								remote.Projects[k].IsApprovalLeader = true
								remote.Projects[k].IsLeaderSend = true
							}
						} else {
							remote.Projects[k].IsApprovalManager = false
							remote.Projects[k].IsManagerSend = true
							remote.Projects[k].NoteManager = note
							remote.IsDelete = true
						}

					}

					remoteTemp = remote
					break
				}
			}
		}

		err = repoM.Save(&remote)
		if err != nil {
			return remotes, err
		}
	}

	notif := new(HistoryService)
	notif.IDRequest = remoteTemp.IdOp
	notif.Name = remoteTemp.Name
	notif.RequestType = "remote"
	notif.UserId = remoteTemp.UserId
	notif.Reason = remoteTemp.Reason
	notif.ManagerApprove = managerName

	s.DateList = dateList
	s.RemoteList = remotes

	s.RemoveLeaveByDates(remoteTemp.UserId, dateList)
	s.SendEmailAfterManager(remoteTemp, param, projectLeaderList, branchmgrs)
	if status == "true" {
		notif.Status = "Approved"
		notif.StatusApproval = "Approved"
		notif.Desc = "already approved by manager"
	} else {
		notif.Status = "Decline"
		notif.StatusApproval = "Decline"
		notif.Desc = "already decline by manager"
	}

	notif.Push(false)
	//log
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&remoteTemp,
		"remote",
	}
	log := tk.M{}
	log.Set("Status", notif.StatusApproval)
	log.Set("Desc", notif.Desc)
	if notif.StatusApproval == "Approved" {
		log.Set("Desc", "Request Approved by Manager")
	} else {
		log.Set("Desc", "Request Declined by Manager")
	}
	log.Set("NameLogBy", managerName)
	log.Set("EmailNameLogBy", emailManager)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		return remotes, err
	}
	return remotes, err
}

func (s *RemoteService) SendEmailAfterManager(remote RemoteModel, param tk.M, leaders map[string][]Project, mgrs []BranchManagerList) error {
	emailSender := MailService{}
	emailSender.Init()

	//details user
	user, err := new(repositories.UserOrmRepo).GetByID(remote.UserId)
	if err != nil {
		return err
	}

	status := param.GetString("Status")
	note := param.GetString("Note")

	tk.Println("---------------- status ", param)
	if status == "true" {
		status = "Approve"
	} else {
		status = "Decline"
	}

	listSender := []string{}
	validForProject := map[string]int{}
	projectNames := []string{}
	managerName := ""
	// managerEmail := ""

	for leader, projects := range leaders {
		listSender = append(listSender, leader)
		for _, project := range projects {
			projectName := project.ProjectName
			if _, ok := validForProject[projectName]; !ok {
				projectNames = append(projectNames, project.ProjectName)
				validForProject[projectName] = 1
			}

			projectName = project.ProjectLeader.Name
			managerName = project.ProjectManager.Name
			// managerEmail = project.ProjectManager.Email
		}
	}

	appCount := 0
	datelist := []string{}
	dateListObject := []DetailDate{}
	for _, remote := range s.RemoteList {
		dateRow := DetailDate{}
		dateRow.DateLeave = remote.DateLeave
		for _, project := range remote.Projects {
			dateRow.Note = project.NoteLeader
			if project.IsApprovalManager && project.IsApprovalLeader {
				dateRow.Status = "Approval"
				datelist = append(datelist, dateRow.DateLeave)
				appCount++
			} else {
				dateRow.Status = "Declined"
				if project.NoteManager != "" {
					dateRow.Note = project.NoteManager
				}
				dateListObject = append(dateListObject, dateRow)
			}
		}
	}

	conf := helper.ReadConfig()
	listSender = append(listSender, conf.GetString("HrdMail"))
	for _, mgr := range mgrs {
		listSender = append(listSender, mgr.Email)
	}
	if remote.SPVManager != nil {
		if len(remote.SPVManager) > 0 {
			for _, spvmgr := range remote.SPVManager {
				listSender = append(listSender, spvmgr.Email)
			}
		}
	}

	if remote.BAnalis != nil {
		if len(remote.BAnalis) > 0 {
			for _, BA := range remote.BAnalis {
				listSender = append(listSender, BA.Email)
			}
		}
	}

	// listSender = append(listSender, managerEmail)
	// hrd, err := s.GetHrOne()

	// if err == nil {
	// 	if hrd.AccountManager.Email != "" {
	// 		listSender = append(listSender, hrd.AccountManager.Email)
	// 	}
	// }

	if user.Email != "" {
		listSender = append(listSender, user.Email)
	}
	dayDuration := strconv.Itoa(appCount)
	fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	fileParam := MailFileParamRequest{
		(remote.CreatedAt).Format("2006-01-02 15:04:05"),
		"",
		remote.Name,
		remote.Reason,
		remote.Type,
		projectNames,
		remote.From,
		remote.To,
		datelist,
		dateListObject,
		"",
		"",
		managerName,
		status,
		note,
		dayDuration,
	}

	emailSender.MailSubject = "Confirmation Remote From Manager"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.Filename = "remoteresponmanager.html"
	emailSender.FileParamRequest = fileParam
	emailSender.To = listSender
	err = emailSender.SendEmail()
	if err != nil {
		return err
	}
	// for _, email := range listSender {
	// 	emailSender.To = []string{email}
	// 	err := emailSender.SendEmail()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (s *RemoteService) SendEmailForHrd() error {
	remote := s.Data
	emailSender := MailService{}
	emailSender.Init()

	hr, err := s.GetHrOne()
	if err != nil {
		return err
	}

	projectNames := []string{"HRD"}
	nameLeader := ""
	idsProject := []string{}

	paramForUrlDecline := map[string]string{
		"IdRequest": remote.IdOp,
		"IdProject": strings.Join(idsProject, ","),
		"Type":      "hrd",
		"Status":    "false",
	}

	paramForUrlApproval := map[string]string{
		"IdRequest": remote.IdOp,
		"IdProject": strings.Join(idsProject, ","),
		"Type":      "hrd",
		"Status":    "true",
	}

	urlDecline, _ := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
	urlApproval, _ := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")
	fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))

	fileParam := MailFileParamRequest{
		(remote.CreatedAt).Format("2006-01-02 15:04:05"),
		nameLeader,
		remote.Name,
		remote.Reason,
		remote.Type,
		projectNames,
		remote.From,
		remote.To,
		s.DateList,
		nil,
		urlDecline,
		urlApproval,
		"",
		"",
		"",
		"",
	}

	emailSender.MailSubject = "Employee Request Remote"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.To = []string{hr.ManagingDirector.Email}
	emailSender.Filename = "remoterequest.html"
	emailSender.FileParamRequest = fileParam
	err = emailSender.SendEmail()
	return err
}

func (s *RemoteService) HandleMailCancelRemote(param tk.M, usr string) ([]RemoteModel, error) {
	idrequest := param.GetString("IDRequest")
	status := param.GetString("Status")
	userid := param.GetString("UserId")
	datelisttemp := param.Get("DateList").(string)
	datelist := strings.Split(datelisttemp, ",")

	user, err := new(repositories.UserOrmRepo).GetByID(userid)
	if err != nil {
		return nil, err
	}

	repoDbox := new(repositories.RemoteDboxRepo)
	repoOrm := new(repositories.RemoteOrmRepo)
	prematch := tk.M{}
	prematch.Set("idop", idrequest).Set("userid", userid).Set("dateleave", tk.M{}.Set("$in", datelist))
	match := tk.M{}.Set("$match", prematch)
	remotes, err := repoDbox.GetByPipe([]tk.M{match})
	if err != nil {
		return remotes, err
	}
	dateCreate := ""
	typeRemoteNum := ""
	reasons := ""
	isChange := false
	for _, remote := range remotes {
		dateCreate = (remote.CreatedAt).Format("2006-01-02 15:04:05")
		typeRemoteNum = remote.Type
		reasons = remote.ReasonAction
		tk.Println("---------------- isdelete ", remote.IsDelete)
		if remote.IsDelete == true {
			return remotes, errors.New("Request cancel remote already approved")
		}
		if remote.IsRequestChange == true {
			isChange = true
			if status == "true" {
				remote.IsDelete = true
			}

			remote.IsRequestChange = false
			err = repoOrm.Save(&remote)
			if err != nil {
				return remotes, err
			}

			layout := "2006-01-02"
			t, _ := time.Parse(layout, remote.DateLeave)

			//log
			emptyleave := new(RequestLeaveModel)
			service := LogService{
				emptyleave,
				&remote,
				"remote",
			}
			log := tk.M{}
			if status == "true" {
				log.Set("Status", "Cancel")
				log.Set("Desc", "Request Remote is Canceled"+" for date "+t.Format("02-Jan-2006"))
			} else {
				log.Set("Status", "Cancel")
				log.Set("Desc", "Request Remote is Declined "+" for date "+t.Format("02-Jan-2006"))
			}

			log.Set("NameLogBy", user.Fullname)
			log.Set("EmailNameLogBy", user.Email)
			err = service.CancelRequest(log)
			if err != nil {
				return remotes, err
			}
		}
	}

	statusMail := ""
	if status == "true" {
		statusMail = "Approved"
	} else {
		statusMail = "Declined"
	}

	fmt.Println("---------------", dateCreate)
	typeRemote := s.convertTypetoTextString(typeRemoteNum)
	fileParam := MailFileParamRequest{
		dateCreate,
		"",
		user.Fullname,
		reasons,
		typeRemote,
		[]string{},
		"",
		"",
		datelist,
		nil,
		"",
		"",
		usr,
		statusMail,
		"",
		"",
	}

	conf := helper.ReadConfig()
	emails := []string{user.Email, conf.GetString("HrdMail")}
	if isChange {
		//this template is changged. This template now sending to 2 emails (hrd and user) =====
		err = s.SendEmailConfirmation(emails, "remotecanceluniversal.html", fileParam)
		if err != nil {
			return remotes, err
		}

		// err = s.SendEmailConfirmation(conf.GetString("HrdMail"), "remotecancelforhrd.html", fileParam)
		// if err != nil {
		// 	return err
		// }
	}
	param["Status"] = false
	return remotes, nil
}

func (s *RemoteService) HandleMailCancelRemoteFromApp(param tk.M) ([]RemoteModel, error) {
	idrequest := param.Get("IdRequest").([]interface{})
	status := param.GetString("Status")
	userid := param.GetString("ManagerUserId")
	note := param.GetString("Note")

	user, err := new(repositories.UserOrmRepo).GetByID(userid)
	if err != nil {
		return nil, err
	}

	repoDbox := new(repositories.RemoteDboxRepo)
	repoOrm := new(repositories.RemoteOrmRepo)
	prematch := tk.M{}
	remotes := []RemoteModel{}

	for _, getid := range idrequest {
		prematch.Set("_id", getid)
		match := tk.M{}.Set("$match", prematch)
		//tk.Println("remeeeeee.....", match, getid)
		getremote, err := repoDbox.GetByPipe([]tk.M{match})
		if err != nil {
			return remotes, err
		}
		//append(remotes, getremote)
		remotes = append(remotes, getremote...)
	}

	//tk.Println("rem.....", remotes)
	// if err != nil {
	// 	return remotes, err
	// }
	dateCreate := ""
	typeRemoteNum := ""
	reasons := ""
	empName := ""
	empEmail := ""
	isChange := false
	datelist := []string{}
	for _, remote := range remotes {
		dateCreate = (remote.CreatedAt).Format("2006-01-02 15:04:05")
		typeRemoteNum = remote.Type
		//reasons = remote.ReasonAction
		reasons = note
		empName = remote.Name
		empEmail = remote.Email
		datelist = append(datelist, remote.DateLeave)
		if remote.IsRequestChange == true {
			isChange = true
			if status == "true" {
				remote.IsDelete = true
			}

			remote.IsRequestChange = false
			err = repoOrm.Save(&remote)
			if err != nil {
				return remotes, err
			}

			layout := "2006-01-02"
			t, _ := time.Parse(layout, remote.DateLeave)

			//log
			emptyleave := new(RequestLeaveModel)
			service := LogService{
				emptyleave,
				&remote,
				"remote",
			}
			log := tk.M{}
			if status == "true" {
				log.Set("Status", "Cancel")
				log.Set("Desc", "Request Remote is Canceled"+" for date "+t.Format("02-Jan-2006"))
			} else {
				log.Set("Status", "Cancel")
				log.Set("Desc", "Request Remote is Declined "+" for date "+t.Format("02-Jan-2006"))
			}

			log.Set("NameLogBy", user.Fullname)
			log.Set("EmailNameLogBy", user.Email)
			err = service.CancelRequest(log)
			if err != nil {
				return remotes, err
			}
		}
	}

	statusMail := ""
	if status == "true" {
		statusMail = "Approved"
	} else {
		statusMail = "Declined"
	}

	fmt.Println("---------------", dateCreate)
	typeRemote := s.convertTypetoTextString(typeRemoteNum)
	fileParam := MailFileParamRequest{
		dateCreate,
		"",
		empName,
		reasons,
		typeRemote,
		[]string{},
		"",
		"",
		datelist,
		nil,
		"",
		"",
		"",
		statusMail,
		"",
		"",
	}

	conf := helper.ReadConfig()
	emails := []string{empEmail, conf.GetString("HrdMail")}
	if isChange {
		//this template is changged. This template now sending to 2 emails (hrd and user) =====
		err = s.SendEmailConfirmation(emails, "remotecanceluniversal.html", fileParam)
		if err != nil {
			return remotes, err
		}

		// err = s.SendEmailConfirmation(conf.GetString("HrdMail"), "remotecancelforhrd.html", fileParam)
		// if err != nil {
		// 	return err
		// }
	}

	return remotes, nil
}

func (s *RemoteService) SendEmailConfirmation(emails []string, template string, mailparam MailFileParamRequest) error {

	emailSender := MailService{}
	emailSender.Init()

	emailSender.MailSubject = "Remote Work request has been turn down"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.To = emails
	emailSender.Filename = template
	emailSender.FileParamRequest = mailparam
	err := emailSender.SendEmail()

	return err
}
func (s *RemoteService) HandleEmailResponseFromHRD(param tk.M) error {

	emailSender := MailService{}
	emailSender.Init()

	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}

	idTemp := param.GetString("IdRequest")
	dboxFilter = append(dboxFilter, dbox.Eq("idop", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	remote := RemoteModel{}

	repoM := new(repositories.RemoteOrmRepo)
	remotes, err := repoM.GetByParam(filter)
	if err != nil {
		return err
	}

	if len(remotes) > 0 {
		remote = remotes[0]
	}

	//getuserdetail
	user, err := new(repositories.UserOrmRepo).GetByID(remote.UserId)
	if err != nil {
		return err
	}

	status := param.GetString("Status")
	note := param.GetString("Note")

	if status == "true" {
		status = "Approve"
	} else {
		status = "Decline"
	}

	datelist := []string{}
	dateListObject := []DetailDate{}
	for _, remote := range s.RemoteList {
		dateRow := DetailDate{}
		dateRow.DateLeave = remote.DateLeave
		for _, project := range remote.Projects {
			dateRow.Note = project.NoteLeader
			if project.IsApprovalManager && project.IsApprovalLeader {
				dateRow.Status = "Approval"
				datelist = append(datelist, dateRow.DateLeave)
			} else {
				dateRow.Status = "Declined"
				if project.NoteManager != "" {
					dateRow.Note = project.NoteManager
				}
				dateListObject = append(dateListObject, dateRow)
			}
		}
	}

	listSender := []string{user.Email}
	projectNames := []string{"HRD"}

	hrd, err := s.GetHrOne()

	if err == nil {
		if hrd.AccountManager.Email != "" {
			listSender = append(listSender, hrd.AccountManager.Email)
		}
	}

	managerName := hrd.ManagingDirector.Email
	fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	fileParam := MailFileParamRequest{
		(remote.CreatedAt).Format("2006-01-02 15:04:05"),
		"",
		remote.Name,
		remote.Reason,
		remote.Type,
		projectNames,
		remote.From,
		remote.To,
		datelist,
		dateListObject,
		"",
		remote.Projects[0].ProjectManager.Name,
		managerName,
		status,
		note,
		"",
	}

	emailSender.MailSubject = "Confirmation Remote From Manager"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.Filename = "remoteresponmanager.html"
	emailSender.FileParamRequest = fileParam
	for _, email := range listSender {
		emailSender.To = []string{email}
		emailSender.SendEmail()
	}
	//log
	emptyleave := new(RequestLeaveModel)
	service := LogService{
		emptyleave,
		&remote,
		"remote",
	}
	log := tk.M{}
	if status == "true" {
		log.Set("Status", "Approved")
		log.Set("Desc", "Request Approved by HDR")
	} else {
		log.Set("Status", "Declined")
		log.Set("Desc", "Request Declined by HDR")
	}
	log.Set("NameLogBy", hrd.AccountManager.Name)
	log.Set("EmailNameLogBy", hrd.AccountManager.Email)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		return err
	}
	return nil
}

func (s *RemoteService) HandleUrlDecline(param map[string]string, url string) (string, error) {
	conf := helper.ReadConfig()
	request, err := http.NewRequest("GET", (conf.GetString("BaseUrl") + url), nil)
	if err != nil {
		return "", err
	}

	queryParam := request.URL.Query()
	encparam, err := s.EncodeParam(param)
	if err != nil {
		return "", err
	}

	queryParam.Add("Param", encparam)

	request.URL.RawQuery = queryParam.Encode()

	return request.URL.String(), nil
}

func (s *RemoteService) HandleUrlApproval(param map[string]string, url string) (string, error) {
	conf := helper.ReadConfig()
	request, err := http.NewRequest("GET", (conf.GetString("BaseUrl") + url), nil)
	if err != nil {
		return "", err
	}

	queryParam := request.URL.Query()
	encparam, err := s.EncodeParam(param)
	if err != nil {
		return "", err
	}

	queryParam.Add("Param", encparam)

	request.URL.RawQuery = queryParam.Encode()

	return request.URL.String(), nil
}

func DatelistGroupBy(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// func (s *RemoteService) ProcessSaveWithDateList(remote RemoteModel) tk.M {
// 	listError := []tk.M{}
// 	remoteSkip := []RemoteModel{}
// 	remoteSuccess := []RemoteModel{}
// 	dateList := []string{}

//  monthYearCheck := time.Now().Format("200601")
// 	monthYear := time.Now().Format("012006")
// 	after := time.Now().AddDate(0, 1, 0).Format("012006")
// 	beginMonthAfter, _ := time.Parse("012006", after)
// 	beginMonth, _ := time.Parse("012006", monthYear)
// 	listdateInMonth := []string{}
// 	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
// 		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
// 	}
// 	listdateInMonthAfter := []string{}
// 	for daf := beginMonthAfter; daf.Month() == beginMonthAfter.Month(); daf = daf.AddDate(0, 0, 1) {
// 		listdateInMonthAfter = append(listdateInMonthAfter, daf.Format("2006-01-02"))
// 	}

// 	dtl := s.DateList
// 	gu := []string{}
// 	err := gubrak.Each(dtl, func(each string, i int) {
// 		s := strings.Split(each, "-")
// 		sa, sb := s[0], s[1]
// 		gu = append(gu, sa+sb)
// 	})

// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	result, err := gubrak.GroupBy(gu, func(each string) string {
// 		return each
// 	})
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	fg := result.(map[string][]string)

// 	fmt.Println(fg)

// 	uniqueSlice := DatelistGroupBy(gu)
// 	fmt.Println(uniqueSlice)

// 	fmt.Println("ppp", len(fg[uniqueSlice[0]]), after)

// 	pipe := []tk.M{}

// 	pipe = append(pipe,
// 		tk.M{
// 			"$match": tk.M{
// 				"$and": []tk.M{
// 					{"dateleave": tk.M{"$in": listdateInMonth}},
// 					{"userid": tk.M{"$eq": s.Data.UserId}},
// 					{"isdelete": tk.M{"$eq": false}},
// 					{
// 						"projects": tk.M{
// 							"$elemMatch": tk.M{
// 								"isapprovalmanager": tk.M{"$eq": true},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	)

// 	pipe2 := []tk.M{}

// 	pipe2 = append(pipe2,
// 		tk.M{
// 			"$match": tk.M{
// 				"$and": []tk.M{
// 					{"dateleave": tk.M{"$in": listdateInMonthAfter}},
// 					{"userid": tk.M{"$eq": s.Data.UserId}},
// 					{"isdelete": tk.M{"$eq": false}},
// 					{
// 						"projects": tk.M{
// 							"$elemMatch": tk.M{
// 								"isapprovalmanager": tk.M{"$eq": true},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	)

// 	CheckDate0, _ := new(repositories.RemoteOrmRepo).ChekDatelist(pipe)
// 	CheckDate1, _ := new(repositories.RemoteOrmRepo).ChekDatelist(pipe2)
// 	//tk.Println("datelist.......", s.DateList, len(CheckDate), s.Data.UserId)
// 	checkSuccess := true
// 	tk.Println("^^^^^^^^^^^^^^^^^^^^^^^^^ ", uniqueSlice)
// 	for y, dn := range uniqueSlice {
// 		getCdate := []RemoteModel{}
// 		if dn == monthYearCheck {
// 			getCdate = CheckDate0
// 		} else {
// 			getCdate = CheckDate1
// 		}
// 		if checkSuccess {
// 			if len(getCdate)+len(fg[uniqueSlice[y]]) <= 10 {
// 				checkSuccess = true
// 				if y == len(uniqueSlice)-1 {
// 					for _, date := range s.DateList {
// 						if remote.Type == "1" {
// 							remote.From = date
// 							remote.To = date
// 						}

// 						remote.DateLeave = date
// 						remote.Id = bson.NewObjectId().Hex()
// 						// remote.IsExpired = false

// 						skipR, _ := s.VerifyUserRemoteByDate(remote.UserId, date)

// 						if len(skipR) > 0 {
// 							remoteSkip = append(remoteSkip, skipR[0])
// 							continue
// 						}
// 						remote.CreatedAt = time.Now()
// 						// _, err := s.Save(remote)
// 						// if err != nil {
// 						// listError = append(listError, tk.M{}.Set("IdRequest", remote.Id).Set("Error", err.Error()).Set("Date", date))
// 						// } else {
// 						// 	dateList = append(dateList, remote.DateLeave)
// 						// 	remoteSuccess = append(remoteSuccess, remote)
// 						// }
// 					}
// 				}
// 			} else {
// 				checkSuccess = false
// 			}
// 		}

// 	}

// 	s.DateList = dateList
// 	return tk.M{}.Set("ListOfError", listError).Set("RemoteSkip", remoteSkip).Set("RemoteSuccess", remoteSuccess).Set("Check", checkSuccess)
// }

func (s *RemoteService) ProcessSaveWithDateList(remote RemoteModel) tk.M {
	listError := []tk.M{}
	remoteSkip := []RemoteModel{}
	remoteSuccess := []RemoteModel{}
	dateList := []string{}

	for _, date := range s.DateList {
		if remote.Type == "1" {
			remote.From = date
			remote.To = date
		}

		remote.DateLeave = date
		remote.Id = bson.NewObjectId().Hex()
		// remote.IsExpired = false

		skipR, _ := s.VerifyUserRemoteByDate(remote.UserId, date)

		if len(skipR) > 0 {
			remoteSkip = append(remoteSkip, skipR[0])
			continue
		}
		remote.CreatedAt = time.Now()
		_, err := s.Save(remote)
		if err != nil {
			listError = append(listError, tk.M{}.Set("IdRequest", remote.Id).Set("Error", err.Error()).Set("Date", date))
		} else {
			dateList = append(dateList, remote.DateLeave)
			remoteSuccess = append(remoteSuccess, remote)
		}
	}

	s.DateList = dateList
	return tk.M{}.Set("ListOfError", listError).Set("RemoteSkip", remoteSkip).Set("RemoteSuccess", remoteSuccess)
}

func (s *RemoteService) VerifyUserRemoteByDate(userid string, date string) ([]RemoteModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("userid", userid), dbox.Eq("dateleave", date), dbox.Eq("projects.isapprovalmanager", true), dbox.Eq("isdelete", false))
	filter.Set("where", dbox.And(dboxFilter...))
	return new(repositories.RemoteOrmRepo).GetByParam(filter)
}

func (s *RemoteService) Save(remote RemoteModel) (RemoteModel, error) {
	err := new(repositories.RemoteOrmRepo).Save(&remote)
	return remote, err
}

func (s *RemoteService) GetHrOne() (HRDAdminModel, error) {
	hrd := HRDAdminModel{}
	filter := tk.M{}.Set("limit", 1)
	hrds, err := new(repositories.HRDOrmRepo).GetByParam(filter)
	if err != nil {
		return hrd, err
	}

	if len(hrds) > 0 {
		hrd = hrds[0]
	}

	return hrd, nil
}

func (s *RemoteService) EncodeParam(param map[string]string) (string, error) {
	jsonString, err := json.Marshal(&param)
	if err != nil {
		return "", err
	}

	encode := helper.GCMEncrypter(string(jsonString))
	return encode, nil
}

func (s *RemoteService) DecodeParam(encode string) (tk.M, error) {
	decodeData := tk.M{}

	decodetext := helper.GCMDecrypter(encode)
	err := json.Unmarshal([]byte(decodetext), &decodeData)
	return decodeData, err
}

func (s *RemoteService) VerifyIsAlreadyHasRequest(userid string) bool {
	filter := tk.M{}.Set("where", dbox.And(dbox.Eq("projects.ismanagersend", false), dbox.Eq("userid", userid), dbox.Eq("isexpired", false)))
	rows, _ := new(repositories.RemoteOrmRepo).GetByParam(filter)
	if len(rows) > 0 {
		for _, row := range rows {
			for _, pro := range row.Projects {
				if pro.IsLeaderSend && !pro.IsApprovalLeader {
					return false
				}
				if pro.IsLeaderSend && pro.IsApprovalLeader {
					return true
				}
			}
		}
		return true
	}

	return false
}

func (s *RemoteService) GetByIdRequest(idreq string) ([]RemoteModel, error) {
	filter := tk.M{}.Set("where", dbox.And(dbox.Eq("idop", idreq)))
	return new(repositories.RemoteOrmRepo).GetByParam(filter)
}

func (s *RemoteService) GetRemoteNeedApproval(userid string, jobrole int) (interface{}, error) {
	// t := time.Now()
	// datenow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	pipe := []tk.M{}

	or := []tk.M{
		tk.M{}.Set("projects.projectleader.userid", userid),
		tk.M{}.Set("$and",
			[]tk.M{
				// tk.M{}.Set("projects.projectmanager.userid", userid),
				tk.M{}.Set("branchmanager.userid", userid),
				tk.M{}.Set("projects.isapprovalleader", true),
			},
		),
	}

	and := []tk.M{
		tk.M{}.Set("isexpired", false),
		tk.M{}.Set("projects.ismanagersend", false),
		tk.M{}.Set("userid", tk.M{}.Set("$ne", userid)),
		tk.M{}.Set("$or", or),
	}

	if jobrole == 6 {
		and = []tk.M{
			tk.M{}.Set("isexpired", false),
			tk.M{}.Set("projects.ismanagersend", false),
			tk.M{}.Set("$or", or),
		}
	}

	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("$and", and)))
	if jobrole == 5 {
		pipe = []tk.M{}
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("userid", tk.M{}.Set("$ne", userid))))
	}
	remotes, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	if err != nil {
		return remotes, err
	}
	newremotesfirst := []RemoteModel{}
	for _, each := range remotes {
		// date, _ := time.Parse("2006-01-02", each.DateLeave)
		// if date.After(datenow) || each.DateLeave == datenow.Format("2006-01-02") {
		newremotesfirst = append(newremotesfirst, each)
		// }
	}
	// tk.Println("--------------  0 ", newremotesfirst)
	newremotes := s.GroupRemoteByRequestID(newremotesfirst)
	statusApproval := tk.M{}.Set("Manager", []tk.M{}).Set("Leader", []tk.M{}).Set("Remotes", []RemoteModel{})
	// tk.Println("--------------  1 ", newremotes)
	for _, remote := range newremotes {
		remoterow := remote.(tk.M)

		newremote := RemoteModel{}
		newremote = remoterow.Get("Remote").(RemoteModel)
		row := tk.M{}.Set("Remote", newremote).Set("DateList", remoterow.Get("DateList").([]string))
		// tk.Println("--------------  2 ", newremote)
		for _, project := range newremote.Projects {
			statusProject := ""
			if project.ProjectLeader.UserId == userid && !project.IsLeaderSend {
				statusProject = "Leader"
				newremote.Projects = []Project{}
				newremote.Projects = append(newremote.Projects, project)
			}
			for _, each := range newremote.BranchManager {
				if each.UserId == userid && !project.IsManagerSend {
					statusProject = "Manager"
					newremote.Projects = []Project{}
					newremote.Projects = append(newremote.Projects, project)
					break
				}
			}

			if statusProject != "" {
				tempStatusRemote := statusApproval.Get(statusProject).([]tk.M)
				tempStatusRemote = append(tempStatusRemote, row)
				statusApproval.Set(statusProject, tempStatusRemote)
			}
		}
	}

	statusApproval.Set("Remotes", newremotesfirst)
	return statusApproval, err
}

func (s *RemoteService) GetRemoteCancelNeedApproval(userid string, jobrole int) (interface{}, error) {
	t := time.Now()
	datenow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	pipe := []tk.M{}

	or := []tk.M{
		tk.M{}.Set("projects.projectleader.userid", userid),
		tk.M{}.Set("$and",
			[]tk.M{
				tk.M{}.Set("branchmanager.userid", userid),
				tk.M{}.Set("projects.isapprovalleader", true),
			},
		),
	}

	and := []tk.M{
		tk.M{}.Set("isexpired", false),
		tk.M{}.Set("isdelete", false),
		tk.M{}.Set("isrequestchange", true),
		tk.M{}.Set("userid", tk.M{}.Set("$ne", userid)),
		tk.M{}.Set("$or", or),
	}

	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("$and", and)))
	if jobrole == 5 {
		pipe = []tk.M{}
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("userid", tk.M{}.Set("$ne", userid))))
	}
	remotes, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	if err != nil {
		return remotes, err
	}
	newremotesfirst := []RemoteModel{}
	for _, each := range remotes {
		date, _ := time.Parse("2006-01-02", each.DateLeave)
		if date.After(datenow) || each.DateLeave == datenow.Format("2006-01-02") {
			newremotesfirst = append(newremotesfirst, each)
		}
	}
	newremotes := s.GroupRemoteByRequestID(newremotesfirst)
	statusApproval := tk.M{}.Set("Remotes", []RemoteModel{}).Set("Cancel", []tk.M{})

	for _, remote := range newremotes {
		remoterow := remote.(tk.M)

		newremote := RemoteModel{}
		newremote = remoterow.Get("Remote").(RemoteModel)
		row := tk.M{}.Set("Remote", newremote).Set("DateList", remoterow.Get("DateList").([]string))

		for _, project := range newremote.Projects {
			statusProject := ""
			for _, each := range newremote.BranchManager {
				if each.UserId == userid && project.IsApprovalManager && newremote.IsRequestChange && !newremote.IsDelete {
					statusProject = "Cancel"
					newremote.Projects = []Project{}
					newremote.Projects = append(newremote.Projects, project)
					break
				}
			}
			if project.ProjectManager.UserId == userid && project.IsApprovalManager && newremote.IsRequestChange && !newremote.IsDelete {
				statusProject = "Cancel"
				newremote.Projects = []Project{}
				newremote.Projects = append(newremote.Projects, project)
			}

			if statusProject != "" {
				tempStatusRemote := statusApproval.Get(statusProject).([]tk.M)
				tempStatusRemote = append(tempStatusRemote, row)
				statusApproval.Set(statusProject, tempStatusRemote)
			}
		}
	}

	statusApproval.Set("Remotes", newremotesfirst)
	return statusApproval, err
}

// type timeSlice []time.Time

// func (s timeSlice) Less(i, j int) bool { return s[i].Before(s[j]) }
// func (s timeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
// func (s timeSlice) Len() int           { return len(s) }

func (s *RemoteService) GetRemoteApproved(userid string, jobrole int) (interface{}, error) {
	filter := tk.M{}
	if jobrole == 5 || jobrole == 1 {
		filter = filter.Set("where", dbox.And(dbox.Eq("projects.ismanagersend", true), dbox.Eq("isdelete", false), dbox.Eq("isrequestchange", false), dbox.Ne("projects.isapprovalmanager", false)))
	} else if jobrole == 6 {
		filter = filter.Set("where", dbox.And(dbox.Eq("projects.ismanagersend", true), dbox.Eq("isdelete", false), dbox.Eq("isrequestchange", false), dbox.Ne("projects.isapprovalmanager", false)))
	} else {
		filter = filter.Set("where", dbox.And(dbox.Eq("userid", userid), dbox.Eq("projects.ismanagersend", true), dbox.Eq("isdelete", false), dbox.Eq("isrequestchange", false), dbox.Ne("projects.isapprovalmanager", false)))
	}
	remotes, err := new(repositories.RemoteOrmRepo).GetByParam(filter)
	if err != nil {
		return nil, err
	}

	// var dateSlice timeSlice = []time.Time{}

	// for _, m := range remotes {
	// 	dateSlice = append(dateSlice, m.CreatedAt)
	// }

	// sort.Sort(sort.Reverse(dateSlice))

	// tk.Println("---------- data ", len(dateSlice))

	// data := []RemoteModel{}
	// // tk.Println("---------- ", dateSlice)
	// for i, _ := range remotes {
	// 	if remotes[i].CreatedAt == dateSlice[i] {
	// 		data = append(data, remotes[i])
	// 	}else{
	// 		if remotes[i].CreatedAt == dateSlice[i] {
	// 			data = append(data, remotes[i])
	// 		}
	// 	}
	// }

	// tk.Println("---------- data ", data)

	return s._groupRemoteByRequest(remotes), nil
}

func (s *RemoteService) GroupRemoteByRequestID(remotes []RemoteModel) tk.M {
	newremotes := tk.M{}
	for _, remote := range remotes {
		idRequest := remote.IdOp

		if _, ok := newremotes[idRequest]; !ok {
			newremotes[idRequest] = tk.M{}.Set("DateList", []string{}).Set("Remote", RemoteModel{})
		}

		rowdetail := newremotes.Get(idRequest).(tk.M)
		dateLists := rowdetail.Get("DateList").([]string)
		rowdetail.Set("Remote", remote)
		dateLists = append(dateLists, remote.DateLeave)
		rowdetail.Set("DateList", dateLists)
		newremotes.Set(idRequest, rowdetail)
	}

	return newremotes
}

func (s *RemoteService) _groupRemoteByRequest(remotes []RemoteModel) map[string]tk.M {
	newdata := map[string]tk.M{}
	for _, remote := range remotes {
		idrequest := remote.IdOp
		if _, ok := newdata[idrequest]; !ok {
			newdata[idrequest] = tk.M{}.Set("Remotes", []RemoteModel{}).Set("DateList", []string{})
		}

		tempremotes := newdata[idrequest].Get("Remotes").([]RemoteModel)
		tempdates := newdata[idrequest].Get("DateList").([]string)

		tempdates = append(tempdates, remote.DateLeave)
		tempremotes = append(tempremotes, remote)

		newdata[idrequest].Set("Remotes", tempremotes).Set("DateList", tempdates)
	}

	return newdata
}

func (s *RemoteService) RemoveLeaveByDates(userid string, dates []string) error {
	match := tk.M{}.Set("$match", tk.M{}.Set("userid", userid).Set("dateleave", tk.M{}.Set("$in", dates)))

	leaves, err := new(repositories.LeaveDboxRepo).GetByPipe([]tk.M{match})
	if err != nil {
		return err
	}

	repoOrm := new(repositories.LeaveOrmRepo)
	for _, leave := range leaves {
		leave.IsDelete = true
		err = repoOrm.Save(&leave)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *RemoteService) convertTypetoTextString(typeRemote string) string {
	switch typeRemote {
	case "1":
		return "Conditional"
	case "2":
		return "Monthly"
	case "3":
		return "Full Monthly"
	default:
		return ""
	}
}

// ValidateIsManagerSend is
func (s *RemoteService) ValidateIsManagerSend(remotes []RemoteModel) bool {
	remotes, err := s.GetByIdRequest(remotes[0].IdOp)
	if err != nil {
	}
	if remotes[0].Projects[0].IsManagerSend == true {
		return true
	} else {
		return false
	}
}

func (s *RemoteService) ValidateParamFromAppAll(params []tk.M) (interface{}, error) {
	remotes := []RemoteModel{}
	countManagerSend := 0
	for _, p := range params {
		rmts, errs := s.GetByIdRequest(p.GetString("IdRequest"))
		if errs != nil {
		}
		if rmts[0].Projects[0].IsManagerSend == true {
			countManagerSend++
		}
	}
	if countManagerSend > 0 {
		return remotes, errors.New("some requests have been approved / declined")
	}
	for _, param := range params {
		typeOfLeave := param.GetString("Type")
		if typeOfLeave == "manager" {
			remote, err := s.HandleMailResponseFromManager(param, false)
			if err != nil {
				return remotes, err
			}
			remotes = append(remotes, remote...)
		} else if typeOfLeave == "leader" {
			remote, err := s.HandleMailResponseFromLeader(param)
			if err != nil {
				return remotes, err
			}
			remotes = append(remotes, remote...)
		} else if typeOfLeave == "cancel" {
			remote, err := s.HandleMailCancelRemoteFromApp(param)
			if err != nil {
				tk.Println("------------------- masuk eror ", err)
				return remotes, err
			}
			remotes = append(remotes, remote...)
		} else {
			return nil, errors.New("Type Invalid")
		}
	}
	return remotes, nil
}
