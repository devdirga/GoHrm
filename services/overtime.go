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
	"strings"
	"time"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

type OvertimeService struct {
	Data       OvertimeFormModel
	DateList   []string
	RemoteList []OvertimeFormModel
}

func (s *OvertimeService) ProcessOvertime(timezone string) (tk.M, error) {
	// origindata := s.Data
	hasRequest := s.VerifyIsAlreadyHasRequest(s.Data.UserId)
	if hasRequest {
		return nil, errors.New("Your Already Has Request Overtime Before")
	}
	isProjectLeader := false
	isProjectManager := false
	isStaff := false
	//get user details
	user, err := new(repositories.UserOrmRepo).GetByID(s.Data.UserId)
	if err != nil {
		return nil, err
	}

	if user.Designation == "4" && user.Designation == "7" {
		isStaff = true
	}
	if user.Id == s.Data.ProjectLeader.UserId {
		isProjectLeader = true
	}
	if user.Id == s.Data.ProjectManager.UserId {
		isProjectManager = true
	}
	s.Data.IdRequest = bson.NewObjectId().Hex()
	exp, err := helper.ExpiredDateTime(timezone, true)
	rm, err := helper.ExpiredRemaining(timezone, true)
	s.Data.ExpiredOn = exp
	s.Data.ExpiredRemining = rm
	s.Data.IsExpired = false
	dataRespon := s.ProcessSaveWithDateList(s.Data)
	if dataRespon.Get("OvertimeSuccess") == nil {
		return dataRespon, errors.New("Your Already Overtime on Date Your Plan")
	} else if len(dataRespon.Get("OvertimeSuccess").([]OvertimeFormModel)) == 0 {
		return dataRespon, errors.New("Your Already Overtime on Date Your Plan")
	}
	overtimeData := dataRespon.Get("OvertimeSuccess").([]OvertimeFormModel)
	if len(overtimeData) > 0 {
		s.Data = overtimeData[0]
	}
	param := tk.M{}.Set("IdRequest", s.Data.IdRequest).Set("Status", "true")
	if isStaff {
		s.SendEmailForHrd()
	} else if isProjectManager {
		param.Set("Type", "manager")
		s.HandleMailResponseFromManager(param, true)
	} else if isProjectLeader {
		param.Set("Type", "leader")
		s.HandleMailResponseFromLeader(param)
	} else {
		s.SendEmailForLeader(s.Data)
	}
	return nil, nil
}
func (s *OvertimeService) VerifyIsAlreadyHasRequest(userid string) bool {
	filter := tk.M{}.Set("where", dbox.And(dbox.Eq("ismanagerreceive", false), dbox.Eq("userid", userid), dbox.Eq("isexpired", false)))
	rows, _ := new(repositories.OvertimeOrmRepo).GetByParam(filter)
	if len(rows) > 0 {
		for _, row := range rows {
			if row.IsLeaderReceive && !row.IsLeaderApprove {
				return false
			}
			if row.IsLeaderReceive && row.IsLeaderApprove {
				return true
			}
		}
		return true
	}

	return false
}
func (s *OvertimeService) ProcessSaveWithDateList(overtime OvertimeFormModel) tk.M {
	listError := []tk.M{}
	overtimeSkip := []OvertimeFormModel{}
	overtimeSuccess := []OvertimeFormModel{}
	dateList := []string{}

	for _, date := range s.DateList {

		overtime.DateOvertimeString = date
		overtime.DateOvertime, _ = time.Parse("2009-01-02", date)
		weekday := overtime.DateOvertime.Weekday()
		if int(weekday) == 0 || int(weekday) == 6 {
			overtime.IsDayOff = true
		}
		filterNH := tk.M{}.Set("where", dbox.And(dbox.Eq("year", overtime.DateOvertime.Year()), dbox.Eq("month", int(overtime.DateOvertime.Month()))))
		NationalHolidays, _ := new(repositories.NationalHolidayOrmRepo).GetByParam(filterNH)
		if len(NationalHolidays) > 0 {
			for _, each := range NationalHolidays {
				for _, da := range each.ListDate {
					str := da.Format("2009-01-02")
					if str == overtime.DateOvertimeString {
						overtime.IsDayOff = true
						break
					}
				}
			}
		}
		overtime.Id = bson.NewObjectId()

		skipR, _ := s.VerifyUserOvertimeByDate(overtime.UserId, date)

		if len(skipR) > 0 {
			overtimeSkip = append(overtimeSkip, skipR[0])
			continue
		}
		overtime.DateCreated = time.Now()
		_, err := s.Save(overtime)
		if err != nil {
			listError = append(listError, tk.M{}.Set("IdRequest", overtime.Id).Set("Error", err.Error()).Set("Date", date))
		} else {
			dateList = append(dateList, overtime.DateOvertimeString)
			overtimeSuccess = append(overtimeSuccess, overtime)
		}
	}

	s.DateList = dateList
	return tk.M{}.Set("ListOfError", listError).Set("OvertimeSkip", overtimeSkip).Set("OvertimeSuccess", overtimeSuccess)
}
func (s *OvertimeService) VerifyUserOvertimeByDate(userid string, date string) ([]OvertimeFormModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("userid", userid), dbox.Eq("dateovertimestring", date), dbox.Eq("ismanagerapprove", true), dbox.Eq("isdelete", false))
	filter.Set("where", dbox.And(dboxFilter...))
	return new(repositories.OvertimeOrmRepo).GetByParam(filter)
}
func (s *OvertimeService) Save(overtime OvertimeFormModel) (OvertimeFormModel, error) {
	err := new(repositories.OvertimeOrmRepo).Save(&overtime)
	return overtime, err
}
func (s *OvertimeService) SendEmailForLeader(overtime OvertimeFormModel) error {
	projectnames := []string{overtime.Project}
	emailSender := MailService{}
	emailSender.Init()
	paramForUrlDecline := map[string]string{
		"IdRequest": overtime.IdRequest,
		"Type":      "leader",
		"Status":    "false",
	}
	paramForUrlApproval := map[string]string{
		"IdRequest": overtime.IdRequest,
		"Type":      "leader",
		"Status":    "true",
	}
	urlDecline, _ := s.HandleUrlDecline(paramForUrlDecline, "/overtime/handledecline")
	urlApproval, _ := s.HandleUrlApproval(paramForUrlApproval, "/overtime/handleapproval")
	fileParam := MailFileParamRequest{
		(overtime.DateCreated).Format("2006-01-02 15:04:05"),
		overtime.ProjectLeader.Name,
		overtime.Name,
		overtime.Reason,
		"overtime",
		projectnames,
		overtime.From,
		overtime.To,
		s.DateList,
		nil,
		urlDecline,
		urlApproval,
		overtime.ProjectManager.Name,
		"",
		"",
		"",
	}
	emailSender.MailSubject = overtime.Name + " is Requesting for Overtime Work " + time.Now().Format("January")
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.To = []string{overtime.ProjectLeader.Email}
	emailSender.Filename = "overtimerequest.html"
	emailSender.FileParamRequest = fileParam
	err := emailSender.SendEmail()
	fmt.Println("--->>>> log email", err)

	return nil
}

func (s *OvertimeService) HandleUrlDecline(param map[string]string, url string) (string, error) {
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

func (s *OvertimeService) HandleUrlApproval(param map[string]string, url string) (string, error) {
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
func (s *OvertimeService) EncodeParam(param map[string]string) (string, error) {
	jsonString, err := json.Marshal(&param)
	if err != nil {
		return "", err
	}

	encode := helper.GCMEncrypter(string(jsonString))
	return encode, nil
}

func (s *OvertimeService) DecodeParam(encode string) (tk.M, error) {
	decodeData := tk.M{}

	decodetext := helper.GCMDecrypter(encode)
	err := json.Unmarshal([]byte(decodetext), &decodeData)
	return decodeData, err
}
func (s *OvertimeService) ValidateParamMail(paramView tk.M, usr string) (tk.M, error) {
	paramenc := paramView.GetString("Param")
	param, err := s.DecodeParam(paramenc)
	if err != nil {
		return param, err
	}

	param.Set("Note", paramView.GetString("Note"))
	if param.GetString("Type") == "leader" {
		overtimes, err := s.HandleMailResponseFromLeader(param)
		if err != nil {
			return param, err
		}
		for _, each := range overtimes {
			if each.IsExpired {
				param.Set("IsExpired", true)
				break
			}
		}
	} else if param.GetString("Type") == "manager" {
		s.HandleMailResponseFromManager(param, false)
	} else if param.GetString("Type") == "hrd" {
		s.HandleEmailResponseFromHRD(param)
	} else if param.GetString("Type") == "cancelremote" {
		s.HandleMailCancelRemote(param, usr)
	}

	return param, nil
}
func (s *OvertimeService) HandleMailResponseFromLeader(param tk.M) ([]OvertimeFormModel, error) {
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

	dboxFilter = append(dboxFilter, dbox.Eq("idrequest", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	repoM := new(repositories.OvertimeOrmRepo)
	overtimes, err := repoM.GetByParam(filter)
	if err != nil {
		return overtimes, err
	}
	for _, each := range overtimes {
		if each.IsExpired {
			return overtimes, nil
		}
	}
	status := param.GetString("Status")
	typeRespon := param.GetString("Type")
	// idProjects := strings.Split(param.GetString("IdProject"), ",")
	note := param.GetString("Note")

	// managerWithProject := map[string][]Project{}
	overtimeTemp := OvertimeFormModel{}

	dateList := []string{}
	branchmanager := []BranchManagerList{}
	for _, overtime := range overtimes {
		dateList = append(dateList, overtime.DateOvertimeString)
		if typeRespon == "leader" {
			if status == "true" || approvedManager {
				overtime.IsLeaderReceive = true
				overtime.IsLeaderApprove = true
			} else {
				overtime.IsLeaderApprove = false
				overtime.IsLeaderReceive = true
				overtime.DeclineReason = note
			}

		}
		overtimeTemp = overtime

		branchmanager = overtime.BranchManagers
		err = repoM.Save(&overtime)
		if err != nil {
			return overtimes, err
		}
	}

	s.DateList = dateList
	if !approvedManager {
		if status == "true" {
			err = s.SendEmailForManager(overtimeTemp, param, branchmanager)
			if err != nil {
				return overtimes, err
			}
		} else {
			err = s.SendEmailDeclineLeader(overtimeTemp, param, overtimeTemp.ProjectLeader.Name, overtimeTemp.ProjectLeader.Email)
			if err != nil {
				return overtimes, err
			}
		}
	}

	return overtimes, err
}
func (s *OvertimeService) ValidateRequestHasSendManager(idRequest string) ([]OvertimeFormModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("idrequest", idRequest))
	dboxFilter = append(dboxFilter, dbox.Eq("ismanagerreceive", true))

	filter.Set("where", dbox.And(dboxFilter...))

	return new(repositories.OvertimeOrmRepo).GetByParam(filter)
}
func (s *OvertimeService) SendEmailForManager(overtime OvertimeFormModel, param tk.M, branchmanager []BranchManagerList) error {
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
	for _, mgr := range branchmanager {
		paramForUrlDecline := map[string]string{
			"IdRequest":     idRequest,
			"Type":          "manager",
			"Status":        "false",
			"ManagerUserId": mgr.UserId,
		}

		paramForUrlApproval := map[string]string{
			"IdRequest":     idRequest,
			"Type":          "manager",
			"Status":        "true",
			"ManagerUserId": mgr.UserId,
		}

		urlDecline, err := s.HandleUrlDecline(paramForUrlDecline, "/overtime/handledecline")
		if err != nil {
			return err
		}

		urlApproval, err := s.HandleUrlApproval(paramForUrlApproval, "/overtime/handleapproval")
		if err != nil {
			return err
		}

		fileParam := MailFileParamRequest{
			(overtime.DateCreated).Format("2006-01-02 15:04:05"),
			overtime.ProjectLeader.Name,
			overtime.Name,
			overtime.Reason,
			"overtime",
			[]string{overtime.Project},
			overtime.From,
			overtime.To,
			s.DateList,
			nil,
			urlDecline,
			urlApproval,
			mgr.Name,
			status,
			note,
			"",
		}

		emailSender.MailSubject = "Confirmation Overtime From Leader"
		emailSender.From = emailSender.Conf.EmailOperator
		// emailSender.To = []string{manager}
		emailSender.To = []string{mgr.Email}
		emailSender.Filename = "overtimemanager.html"
		emailSender.FileParamRequest = fileParam
		err = emailSender.SendEmail()

		if err != nil {
			fmt.Println("--->>>> log email", err)
			return err
		}
	}

	return nil
}
func (s *OvertimeService) SendEmailDeclineLeader(overtime OvertimeFormModel, param tk.M, leaderName string, leaderEmail string) error {
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

	project := []string{""}
	fileParam := MailFileParamRequest{
		(overtime.DateCreated).Format("2006-01-02 15:04:05"),
		leaderName,
		overtime.Name,
		overtime.Reason,
		"overtime",
		project,
		overtime.From,
		overtime.To,
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
	emailSender.Filename = "overtimedeclineleader.html"
	emailSender.FileParamRequest = fileParam
	err := emailSender.SendEmail()

	if err != nil {
		fmt.Println("--->>>> log email", err)
		return err
	}
	return nil
}
func (s *OvertimeService) HandleMailResponseFromManager(param tk.M, isProjectManager bool) ([]OvertimeFormModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}

	idTemp := param.GetString("IdRequest")
	dboxFilter = append(dboxFilter, dbox.Eq("idrequest", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	overtimeTemp := OvertimeFormModel{}

	repoM := new(repositories.OvertimeOrmRepo)
	overtimes, err := repoM.GetByParam(filter)
	if err != nil {
		return overtimes, err
	}

	status := param.GetString("Status")
	typeRespon := param.GetString("Type")
	note := param.GetString("Note")

	dateList := []string{}
	branchmgr := BranchManagerList{}
	branchmgrs := []BranchManagerList{}
	for _, overtime := range overtimes {
		branchmgrs = overtime.BranchManagers
		for _, mgr := range overtime.BranchManagers {
			if mgr.UserId == param.GetString("ManagerUserId") {
				branchmgr = mgr
			}
		}
		overtime.ApprovalManager.UserId = branchmgr.UserId
		overtime.ApprovalManager.Email = branchmgr.Email
		overtime.ApprovalManager.Name = branchmgr.Name
		overtime.ApprovalManager.IdEmp = branchmgr.IdEmp
		overtime.ApprovalManager.PhoneNumber = branchmgr.PhoneNumber
		overtime.ApprovalManager.Location = branchmgr.Location
		dateList = append(dateList, overtime.DateOvertimeString)
		if typeRespon == "manager" {
			if status == "true" {
				overtime.IsManagerApprove = true
				overtime.IsManagerReceive = true
				if isProjectManager {
					overtime.IsLeaderApprove = true
					overtime.IsLeaderReceive = true
				}
			} else {
				overtime.IsManagerApprove = false
				overtime.IsManagerReceive = true
				overtime.DeclineReason = note
			}

		}

		overtimeTemp = overtime

		err = repoM.Save(&overtime)
		if err != nil {
			return overtimes, err
		}
	}

	s.DateList = dateList

	s.SendEmailAfterManager(overtimeTemp, param, branchmgrs)

	return overtimes, err
}
func (s *OvertimeService) SendEmailAfterManager(overtime OvertimeFormModel, param tk.M, mgrs []BranchManagerList) error {
	emailSender := MailService{}
	emailSender.Init()

	//details user
	user, err := new(repositories.UserOrmRepo).GetByID(overtime.UserId)
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

	listSender := []string{}
	projectNames := []string{overtime.Project}
	managerName := overtime.ApprovalManager.Name
	listSender = append(listSender, overtime.ProjectLeader.Email)
	// managerEmail := ""

	conf := helper.ReadConfig()
	listSender = append(listSender, conf.GetString("HrdMail"))
	for _, mgr := range mgrs {
		listSender = append(listSender, mgr.Email)
	}

	if user.Email != "" {
		listSender = append(listSender, user.Email)
	}
	dayDuration := strconv.Itoa(len(s.DateList))
	// fmt.Println("---------------", (remote.CreatedAt).Format("2006-01-02 15:04:05"))
	fileParam := MailFileParamRequest{
		(overtime.DateCreated).Format("2006-01-02 15:04:05"),
		"",
		overtime.Name,
		overtime.Reason,
		"overtime",
		projectNames,
		overtime.From,
		overtime.To,
		s.DateList,
		nil,
		"",
		"",
		managerName,
		status,
		note,
		dayDuration,
	}

	emailSender.MailSubject = "Confirmation Remote From Manager"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.Filename = "overtimeresponmanager.html"
	emailSender.FileParamRequest = fileParam
	emailSender.To = listSender
	err = emailSender.SendEmail()
	if err != nil {
		return err
	}

	return nil
}
func (s *OvertimeService) SendEmailForHrd() error {
	overtime := s.Data
	emailSender := MailService{}
	emailSender.Init()

	hr, err := s.GetHrOne()
	if err != nil {
		return err
	}

	projectNames := []string{"HRD"}
	nameLeader := ""
	// idsProject := []string{}

	paramForUrlDecline := map[string]string{
		"IdRequest": overtime.IdRequest,
		"Type":      "hrd",
		"Status":    "false",
	}

	paramForUrlApproval := map[string]string{
		"IdRequest": overtime.IdRequest,
		"Type":      "hrd",
		"Status":    "true",
	}

	urlDecline, _ := s.HandleUrlDecline(paramForUrlDecline, "/remote/handledecline")
	urlApproval, _ := s.HandleUrlApproval(paramForUrlApproval, "/remote/handleapproval")
	// fmt.Println("---------------", (overtime.CreatedAt).Format("2006-01-02 15:04:05"))

	fileParam := MailFileParamRequest{
		(overtime.DateCreated).Format("2006-01-02 15:04:05"),
		nameLeader,
		overtime.Name,
		overtime.Reason,
		"overtime",
		projectNames,
		overtime.From,
		overtime.To,
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
	emailSender.Filename = "overtimerequest.html"
	emailSender.FileParamRequest = fileParam
	err = emailSender.SendEmail()
	return err
}
func (s *OvertimeService) GetHrOne() (HRDAdminModel, error) {
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
func (s *OvertimeService) HandleEmailResponseFromHRD(param tk.M) error {

	emailSender := MailService{}
	emailSender.Init()

	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}

	idTemp := param.GetString("IdRequest")
	dboxFilter = append(dboxFilter, dbox.Eq("idrequest", idTemp))
	filter.Set("where", dbox.And(dboxFilter...))

	overtime := OvertimeFormModel{}

	repoM := new(repositories.OvertimeOrmRepo)
	overtimes, err := repoM.GetByParam(filter)
	if err != nil {
		return err
	}

	if len(overtimes) > 0 {
		overtime = overtimes[0]
	}

	//getuserdetail
	user, err := new(repositories.UserOrmRepo).GetByID(overtime.UserId)
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

	listSender := []string{user.Email}
	projectNames := []string{"HRD"}

	hrd, err := s.GetHrOne()

	if err == nil {
		if hrd.AccountManager.Email != "" {
			listSender = append(listSender, hrd.AccountManager.Email)
		}
	}

	managerName := hrd.ManagingDirector.Email
	fmt.Println("---------------", (overtime.DateCreated).Format("2006-01-02 15:04:05"))
	fileParam := MailFileParamRequest{
		(overtime.DateCreated).Format("2006-01-02 15:04:05"),
		"",
		overtime.Name,
		overtime.Reason,
		"overtime",
		projectNames,
		overtime.From,
		overtime.To,
		s.DateList,
		nil,
		"",
		"",
		managerName,
		status,
		note,
		"",
	}

	emailSender.MailSubject = "Confirmation Remote From Manager"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.Filename = "overtimeresponmanager.html"
	emailSender.FileParamRequest = fileParam
	for _, email := range listSender {
		emailSender.To = []string{email}
		emailSender.SendEmail()
	}
	return nil
}
func (s *OvertimeService) HandleMailCancelRemote(param tk.M, usr string) error {
	idrequest := param.GetString("IDRequest")
	status := param.GetString("Status")
	userid := param.GetString("UserId")
	datelisttemp := param.Get("DateList").(string)
	datelist := strings.Split(datelisttemp, ",")

	user, err := new(repositories.UserOrmRepo).GetByID(userid)
	if err != nil {
		return err
	}

	repoDbox := new(repositories.OvertimeDboxRepo)
	repoOrm := new(repositories.OvertimeOrmRepo)
	prematch := tk.M{}
	prematch.Set("idop", idrequest).Set("userid", userid).Set("dateleave", tk.M{}.Set("$in", datelist))
	match := tk.M{}.Set("$match", prematch)
	overtimes, err := repoDbox.GetByPipe([]tk.M{match})
	if err != nil {
		return err
	}
	dateCreate := ""
	reasons := ""
	isChange := false
	for _, overtime := range overtimes {
		dateCreate = (overtime.DateCreated).Format("2006-01-02 15:04:05")
		reasons = overtime.Reason
		if overtime.IsRequestChange == true {
			isChange = true
			if status == "true" {
				overtime.IsDelete = true
			}

			overtime.IsRequestChange = false
			err = repoOrm.Save(&overtime)
			if err != nil {
				return err
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
	fileParam := MailFileParamRequest{
		dateCreate,
		"",
		user.Fullname,
		reasons,
		"overtime",
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
		err = s.SendEmailConfirmation(emails, "overtimecanceluniversal.html", fileParam)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *OvertimeService) SendEmailConfirmation(emails []string, template string, mailparam MailFileParamRequest) error {

	emailSender := MailService{}
	emailSender.Init()

	emailSender.MailSubject = "Overtime Work request has been turn down"
	emailSender.From = emailSender.Conf.EmailOperator
	emailSender.To = emails
	emailSender.Filename = template
	emailSender.FileParamRequest = mailparam
	err := emailSender.SendEmail()

	return err
}
