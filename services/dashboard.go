package services

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
	gu "github.com/novalagung/gubrak"
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
)

type PayloadDashboardService struct {
	Project      string
	Location     string
	Designations []string
	DateByMonth  string
	UserId       string
	EmpId        string
	Level        int
	MonthYear    string
	Search       string
	UserOnly     bool
}

type DashboardService struct{}

func (s *DashboardService) ConstructDashboardData(payload PayloadDashboardService) (interface{}, error) {
	dataRemote, err := s.GetRemoteForDashboard(payload)
	// tk.Println("error", err)
	if err != nil {
		return nil, err
	}

	dataLeave, err := s.GetRequestLeave(payload)
	if err != nil {
		return nil, err
	}

	dataLeaveEmergency, err := s.GetRequestEmergencyLeave(payload)
	if err != nil {
		return nil, err
	}
	dataSpecialLeave, err := s.GetRequestSpecialLeave(payload)
	if err != nil {
		return nil, err
	}
	dataOvertime, err := s.GetRequestOvertime(payload)
	if err != nil {
		return nil, err
	}
	return tk.M{}.Set("DataRemote", dataRemote).Set("DataLeave", dataLeave).Set("DataEmergency", dataLeaveEmergency).Set("DataSpecial", dataSpecialLeave).Set("DataOvertime", dataOvertime), err
}

func (s DashboardService) ConstructDashboardDataForAdminReport(payload PayloadDashboardService) (tk.M, error) {
	dateByMonth := payload.DateByMonth
	location := payload.Location

	or := []tk.M{
		tk.M{}.Set("projects.isleadersend", true).Set("projects.isapprovalleader", false),
		tk.M{}.Set("projects.isapprovalleader", true).Set("projects.ismanagersend", true).Set("projects.isapprovalmanager", false),
		tk.M{}.Set("projects.isapprovalmanager", true),
	}

	and := []tk.M{
		tk.M{}.Set("dateleave", tk.M{}.Set("$regex", ".*"+dateByMonth+".*")),
		tk.M{}.Set("location", location),
		tk.M{}.Set("$or", or),
	}

	leaveFilter := tk.M{}.Set("dateleave", tk.M{}.Set("$regex", ".*"+dateByMonth+".*")).Set("location", location).Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))

	if location == "" || location == "Global" {
		and = []tk.M{
			tk.M{}.Set("dateleave", tk.M{}.Set("$regex", ".*"+dateByMonth+".*")),
			tk.M{}.Set("$or", or),
		}

		leaveFilter = tk.M{}.Set("dateleave", tk.M{}.Set("$regex", ".*"+dateByMonth+".*")).Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))
	}

	paramRemote := tk.M{}.Set("$match", tk.M{}.Set("$and", and))
	pipe := []tk.M{paramRemote}

	paramLeave := tk.M{}.Set("$match", leaveFilter)
	pipeLeave := []tk.M{paramLeave}

	remotes, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	if err != nil {
		return nil, err
	}

	leaves, err := new(repositories.LeaveDboxRepo).GetByPipe(pipeLeave)
	if err != nil {
		return nil, err
	}

	leavesNotApproved, err := s.GetLeaveNotApproved(dateByMonth, location)
	// tk.Println("data leave not approved : ", len(leavesNotApproved))
	if err != nil {
		return nil, err
	}

	overtimeData, err := s.GetOvertimeData(dateByMonth, location)
	// tk.Println("data overtime : ", tk.JsonStringIndent(overtimeData, "\n >>>>>>"))
	// tk.Println("data overtime : ", len(overtimeData))
	if err != nil {
		return nil, err
	}

	countAllLeave := 0
	countAllLeave = countAllLeave + len(remotes) + len(leaves) + len(overtimeData)

	countRemoteApproved := 0
	countRemoteDecline := 0
	countRemoteRequest := 0
	countRemoteCancelled := 0

	countLeaveApproved := 0
	countLeaveDecline := 0
	countLeaveRequest := 0
	countLeaveCancelled := 0

	countELeaveApproved := 0
	countELeaveDecline := 0
	countELeaveRequest := 0
	countELeaveCancelled := 0

	countOvertimeApproved := 0
	countOvertimeDecline := 0

	// remoteAlreadyCheck := map[string]RemoteModel{}

	listLeaveWeek := map[int]tk.M{
		1: tk.M{}.Set("Remote", 0).Set("ELeave", 0).Set("Leave", 0),
		2: tk.M{}.Set("Remote", 0).Set("ELeave", 0).Set("Leave", 0),
		3: tk.M{}.Set("Remote", 0).Set("ELeave", 0).Set("Leave", 0),
		4: tk.M{}.Set("Remote", 0).Set("ELeave", 0).Set("Leave", 0),
	}

	topfiveuserremote := []tk.M{
		0: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		1: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		2: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		3: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		4: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
	}

	topfiveusereleave := []tk.M{
		0: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		1: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		2: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		3: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		4: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
	}

	topfiveuserleave := []tk.M{
		0: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		1: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		2: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		3: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		4: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
	}

	topfiveuserovertime := []tk.M{
		0: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		1: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		2: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		3: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
		4: tk.M{}.Set("Name", "").Set("UserId", "").Set("Count", 0),
	}

	countLeaveByStatus := tk.M{}
	projectGroupByName := map[string]tk.M{}
	locationGroup := map[string]tk.M{}
	userGroupRemote := map[string]tk.M{}
	userGroupELeave := map[string]tk.M{}
	userGroupLeave := map[string]tk.M{}
	userGroupOvertime := map[string]tk.M{}
	declineList := []tk.M{}
	leaveByPerson := map[string]tk.M{}

	dateByMonthDay := dateByMonth + "-10"
	currentDate, err := time.Parse("2006-1-2", dateByMonthDay)
	if err != nil {
		return nil, err
	}

	nameList := tk.M{}.Set("Remote", []tk.M{}).Set("Leave", []tk.M{}).Set("ELeave", []tk.M{}).Set("Overtime", []tk.M{})

	for _, remote := range remotes {
		// idRequest := remote.IdOp
		isApproval := false
		isDecline := false
		isRequest := false

		for _, project := range remote.Projects {
			if project.IsApprovalManager {
				isApproval = true
				isDecline = false
				isRequest = false
				break
			} else if !project.IsApprovalLeader {
				isApproval = false
				isDecline = true
				isRequest = false
				break
			} else if project.IsManagerSend && !project.IsApprovalManager {
				isApproval = false
				isDecline = true
				isRequest = false
			} else {
				isApproval = false
				isDecline = false
				isRequest = true
			}
		}

		dateLeaveParse, err := time.Parse("2006-1-2", remote.DateLeave)
		if err != nil {
			return nil, err
		}
		if currentDate.Month() == dateLeaveParse.Month() {
			day := dateLeaveParse.Day()
			weekTemp := float64(day / 7)
			week := int(math.Ceil(weekTemp))
			week++

			if week > 4 {
				week = 4
			}

			if _, ok := listLeaveWeek[week]; !ok {
				listLeaveWeek[week] = tk.M{}
				listLeaveWeek[week].Set("Remote", 0)
			}

			countRemoteByWeek := listLeaveWeek[week].GetInt("Remote")
			listLeaveWeek[week].Set("Remote", countRemoteByWeek+1)
		}

		for _, project := range remote.Projects {
			projectname := project.ProjectName
			if projectname == "" {
				projectname = "creativelab"
			}
			if _, ok := projectGroupByName[projectname]; !ok {
				projectGroupByName[projectname] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
			}

			remoteList := nameList.Get("Remote").([]tk.M)
			result, _ := gu.Find(remoteList, func(each tk.M, i int) bool {
				return strings.Contains(each.GetString("Project"), projectname)
			}, 0)

			if result == nil {
				newProject := tk.M{}.Set("Project", projectname).Set("EmpName", []string{remote.Name})
				remoteList = append(remoteList, newProject)
				nameList.Set("Remote", remoteList)

				countprojectremote := projectGroupByName[projectname].GetInt("Remote")

				projectGroupByName[projectname].Set("Remote", countprojectremote+1)
			} else {
				projectList := result.(tk.M)
				empList := projectList.Get("EmpName")

				result2, _ := gu.Find(empList, func(each string, i int) bool {
					return strings.Contains(each, remote.Name)
				}, 0)

				if result2 == nil {
					projectList.Set("EmpName", append(empList.([]string), remote.Name))
					countprojectremote := projectGroupByName[projectname].GetInt("Remote")

					projectGroupByName[projectname].Set("Remote", countprojectremote+1)
				}
			}

		}

		//start by person
		employeeName := remote.Name
		if _, ok := leaveByPerson[employeeName]; !ok {
			leaveByPerson[employeeName] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		countpersonremote := leaveByPerson[employeeName].GetInt("Remote")
		leaveByPerson[employeeName].Set("Remote", countpersonremote+1)
		//end by person

		if isApproval {
			if remote.IsRequestChange {
				countRemoteCancelled++
			} else {
				countRemoteApproved++
			}
		} else if isDecline {
			countRemoteDecline++
			declineReason := remote.Projects[0].NoteLeader
			if remote.Projects[0].NoteManager != "" {
				declineReason = remote.Projects[0].NoteManager
			}
			declineList = append(declineList, tk.M{}.Set("Name", remote.Name).Set("Reason", declineReason))
		} else if isRequest {
			countRemoteRequest++
		}

		locationUser := remote.Location
		if locationUser == "" {
			locationUser = "Indonesia"
		}

		if _, ok := locationGroup[locationUser]; !ok {
			locationGroup[locationUser] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		locationUserCount := locationGroup[locationUser].GetInt("Remote")
		locationGroup[locationUser].Set("Remote", locationUserCount+1)

		if _, ok := userGroupRemote[remote.UserId]; !ok {
			userGroupRemote[remote.UserId] = tk.M{}.Set("UserId", remote.UserId).Set("Name", remote.Name).Set("Count", 0)
		}

		countremotebyuserid := userGroupRemote[remote.UserId].GetInt("Count")
		userGroupRemote[remote.UserId].Set("Count", countremotebyuserid+1)
	}

	for _, user := range userGroupRemote {
		mintop := user.GetInt("Count")
		username := user.GetString("Name")
		userid := user.GetString("UserId")

		for k, top := range topfiveuserremote {
			if mintop > top.GetInt("Count") {
				tempmin := top.GetInt("Count")

				tempusername := top.GetString("Name")
				tempuserid := top.GetString("UserId")

				topfiveuserremote[k].Set("Name", username).Set("UserId", userid).Set("Count", mintop)

				mintop = tempmin
				username = tempusername
				userid = tempuserid
			}
		}
	}

	for _, leave := range leaves {
		dateLeaveParse, err := time.Parse("2006-1-2", leave.DateLeave)
		if err != nil {
			return nil, err
		}

		statusLeave := ""
		if leave.IsEmergency {
			statusLeave = "ELeave"
			countELeaveRequest++
			if leave.StsByManager == "Approved" {
				if leave.IsDelete {
					countELeaveCancelled++
				} else {
					countELeaveApproved++
				}
			}

			if _, ok := userGroupELeave[leave.UserId]; !ok {
				userGroupELeave[leave.UserId] = tk.M{}.Set("UserId", leave.UserId).Set("Name", leave.Name).Set("Count", 0)
			}

			countremotebyuserid := userGroupELeave[leave.UserId].GetInt("Count")
			userGroupELeave[leave.UserId].Set("Count", countremotebyuserid+1)

		} else {
			statusLeave = "Leave"
			countLeaveRequest++
			if leave.StsByManager == "Approved" {
				if leave.IsDelete {
					countLeaveCancelled++
				} else {
					countLeaveApproved++
				}
			}

			if _, ok := userGroupLeave[leave.UserId]; !ok {
				userGroupLeave[leave.UserId] = tk.M{}.Set("UserId", leave.UserId).Set("Name", leave.Name).Set("Count", 0)
			}

			countremotebyuserid := userGroupLeave[leave.UserId].GetInt("Count")
			userGroupLeave[leave.UserId].Set("Count", countremotebyuserid+1)
		}

		for _, projectname := range leave.Project {
			if projectname == "" {
				projectname = "creativelab"
			}
			if _, ok := projectGroupByName[projectname]; !ok {
				projectGroupByName[projectname] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
			}

			remoteList := nameList.Get(statusLeave).([]tk.M)
			result, _ := gu.Find(remoteList, func(each tk.M, i int) bool {
				return strings.Contains(each.GetString("Project"), projectname)
			}, 0)

			if result == nil {
				newProject := tk.M{}.Set("Project", projectname).Set("EmpName", []string{leave.Name})
				remoteList = append(remoteList, newProject)
				nameList.Set(statusLeave, remoteList)

				countprojectleave := projectGroupByName[projectname].GetInt(statusLeave)

				projectGroupByName[projectname].Set(statusLeave, countprojectleave+1)
			} else {
				projectList := result.(tk.M)
				empList := projectList.Get("EmpName")

				result2, _ := gu.Find(empList, func(each string, i int) bool {
					return strings.Contains(each, leave.Name)
				}, 0)

				if result2 == nil {
					projectList.Set("EmpName", append(empList.([]string), leave.Name))
					countprojectleave := projectGroupByName[projectname].GetInt(statusLeave)

					projectGroupByName[projectname].Set(statusLeave, countprojectleave+1)
				}
			}

		}

		//start by person
		employeeName := leave.Name
		if _, ok := leaveByPerson[employeeName]; !ok {
			leaveByPerson[employeeName] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		countpersonleave := leaveByPerson[employeeName].GetInt(statusLeave)
		leaveByPerson[employeeName].Set(statusLeave, countpersonleave+1)
		//end by person

		locationUser := leave.Location
		if locationUser == "" {
			locationUser = "Indonesia"
		}

		if _, ok := locationGroup[locationUser]; !ok {
			locationGroup[locationUser] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		locationUserCount := locationGroup[locationUser].GetInt(statusLeave)
		locationGroup[locationUser].Set(statusLeave, locationUserCount+1)

		if currentDate.Month() == dateLeaveParse.Month() {
			day := dateLeaveParse.Day()
			weekTemp := float64(day / 7)
			week := int(math.Ceil(weekTemp))
			week++

			if week > 4 {
				week = 4
			}

			statusLeave := ""
			if leave.IsEmergency {
				statusLeave = "ELeave"
			} else {
				statusLeave = "Leave"
			}

			if _, ok := listLeaveWeek[week]; !ok {
				listLeaveWeek[week] = tk.M{}
				listLeaveWeek[week].Set(statusLeave, 0)
			}

			countLeaveByWeek := listLeaveWeek[week].GetInt(statusLeave)
			listLeaveWeek[week].Set(statusLeave, countLeaveByWeek+1)

		}
	}

	for _, leave := range leavesNotApproved {
		if leave.StatusManagerProject.StatusRequest == "Declined" {
			if leave.IsEmergency {
				countELeaveDecline++
			} else {
				countLeaveDecline++
			}

			declineReason := leave.StatusProjectLeader[0].Reason
			if leave.StatusManagerProject.Reason != "" {
				declineReason = leave.StatusManagerProject.Reason
			}
			declineList = append(declineList, tk.M{}.Set("Name", leave.Name).Set("Reason", declineReason))
		} //else if leave.StatusManagerProject.StatusRequest != "Approved" {
		// 	if leave.IsEmergency {
		// 		countELeaveRequest++
		// 	} else {
		// 		countLeaveRequest++
		// 	}
		// }
	}

	for _, user := range userGroupELeave {
		mintop := user.GetInt("Count")
		username := user.GetString("Name")
		userid := user.GetString("UserId")

		for k, top := range topfiveusereleave {
			if mintop > top.GetInt("Count") {
				tempmin := top.GetInt("Count")

				tempusername := top.GetString("Name")
				tempuserid := top.GetString("UserId")

				topfiveusereleave[k].Set("Name", username).Set("UserId", userid).Set("Count", mintop)

				mintop = tempmin
				username = tempusername
				userid = tempuserid
			}
		}
	}

	for _, user := range userGroupLeave {
		mintop := user.GetInt("Count")
		username := user.GetString("Name")
		userid := user.GetString("UserId")

		for k, top := range topfiveuserleave {
			if mintop > top.GetInt("Count") {
				tempmin := top.GetInt("Count")

				tempusername := top.GetString("Name")
				tempuserid := top.GetString("UserId")

				topfiveuserleave[k].Set("Name", username).Set("UserId", userid).Set("Count", mintop)

				mintop = tempmin
				username = tempusername
				userid = tempuserid
			}
		}
	}

	//overtime
	for _, overtime := range overtimeData {
		//request by week
		dayList := overtime.Get("daylist")
		dateOvertimeParse, err := time.Parse("2006-1-2", dayList.(tk.M).GetString("date"))
		if err != nil {
			return nil, err
		}

		if currentDate.Month() == dateOvertimeParse.Month() {
			day := dateOvertimeParse.Day()
			weekTemp := float64(day / 7)
			week := int(math.Ceil(weekTemp))
			week++

			if week > 4 {
				week = 4
			}

			if _, ok := listLeaveWeek[week]; !ok {
				listLeaveWeek[week] = tk.M{}
				listLeaveWeek[week].Set("Overtime", 0)
			}

			countOvertimeByWeek := listLeaveWeek[week].GetInt("Overtime")
			listLeaveWeek[week].Set("Overtime", countOvertimeByWeek+1)
		}

		//request by status
		membersOvertime := overtime.Get("membersovertime")
		result := membersOvertime.(tk.M).GetString("result")
		nameEmp := membersOvertime.(tk.M).GetString("name")

		if result == "Confirmed" {
			countOvertimeApproved++
		} else {
			countOvertimeDecline++
			declineReason := overtime.Get("declinereason")
			declineList = append(declineList, tk.M{}.Set("Name", nameEmp).Set("Reason", declineReason))
		}

		//request by project
		{
			projectname := overtime.GetString("project")
			if projectname == "" {
				projectname = "creativelab"
			}
			if _, ok := projectGroupByName[projectname]; !ok {
				projectGroupByName[projectname] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
			}

			overtimeList := nameList.Get("Overtime").([]tk.M)
			result, _ := gu.Find(overtimeList, func(each tk.M, i int) bool {
				return strings.Contains(each.GetString("Project"), projectname)
			}, 0)

			if result == nil {
				newProject := tk.M{}.Set("Project", projectname).Set("EmpName", []string{nameEmp})
				overtimeList = append(overtimeList, newProject)
				nameList.Set("Overtime", overtimeList)

				countprojectovertime := projectGroupByName[projectname].GetInt("Overtime")

				projectGroupByName[projectname].Set("Overtime", countprojectovertime+1)
			} else {
				projectList := result.(tk.M)
				empList := projectList.Get("EmpName")

				result2, _ := gu.Find(empList, func(each string, i int) bool {
					return strings.Contains(each, nameEmp)
				}, 0)

				if result2 == nil {
					projectList.Set("EmpName", append(empList.([]string), nameEmp))
					countprojectovertime := projectGroupByName[projectname].GetInt("Overtime")

					projectGroupByName[projectname].Set("Overtime", countprojectovertime+1)
				}
			}

		}

		//request by location
		locationUser := overtime.GetString("location")
		if locationUser == "" {
			locationUser = "Indonesia"
		}

		if _, ok := locationGroup[locationUser]; !ok {
			locationGroup[locationUser] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		locationUserCount := locationGroup[locationUser].GetInt("Overtime")
		locationGroup[locationUser].Set("Overtime", locationUserCount+1)

		//top five data
		userid := membersOvertime.(tk.M).GetString("userid")
		if _, ok := userGroupOvertime[userid]; !ok {
			userGroupOvertime[userid] = tk.M{}.Set("UserId", userid).Set("Name", membersOvertime.(tk.M).GetString("name")).Set("Count", 0)
		}

		countovertimebyuserid := userGroupOvertime[userid].GetInt("Count")
		userGroupOvertime[userid].Set("Count", countovertimebyuserid+1)

		//start by person
		employeeName := nameEmp
		if _, ok := leaveByPerson[employeeName]; !ok {
			leaveByPerson[employeeName] = tk.M{}.Set("Remote", 0).Set("Leave", 0).Set("ELeave", 0).Set("Overtime", 0)
		}

		countpersonovertime := leaveByPerson[employeeName].GetInt("Overtime")
		leaveByPerson[employeeName].Set("Overtime", countpersonovertime+1)
		//end by person
	}

	for _, user := range userGroupOvertime {
		mintop := user.GetInt("Count")
		username := user.GetString("Name")
		userid := user.GetString("UserId")

		for k, top := range topfiveuserovertime {
			if mintop > top.GetInt("Count") {
				tempmin := top.GetInt("Count")

				tempusername := top.GetString("Name")
				tempuserid := top.GetString("UserId")

				topfiveuserovertime[k].Set("Name", username).Set("UserId", userid).Set("Count", mintop)

				mintop = tempmin
				username = tempusername
				userid = tempuserid
			}
		}
	}

	listCountLeaveApproved := tk.M{}.Set("Leave", countLeaveRequest).Set("ELeave", countELeaveRequest).Set("Remote", len(remotes)).Set("Overtime", len(overtimeData))
	countLeaveByStatus.
		Set("Approved", countRemoteApproved+countLeaveApproved+countELeaveApproved+countOvertimeApproved).
		Set("Decline", countRemoteDecline+countELeaveDecline+countLeaveDecline+countOvertimeDecline).
		Set("Cancelled", countRemoteCancelled+countELeaveCancelled+countLeaveCancelled)
		//Set("Request", countRemoteRequest+countELeaveRequest+countLeaveRequest)

	topFive := tk.M{}.Set("TopRemote", topfiveuserremote).Set("TopELeave", topfiveusereleave).Set("TopLeave", topfiveuserleave).Set("TopOvertime", topfiveuserovertime)
	return tk.M{}.
		Set("CountAll", countAllLeave).
		Set("CountLeaveApproved", listCountLeaveApproved).
		Set("CountByStatus", countLeaveByStatus).
		Set("DeclineList", declineList).
		Set("LeaveByWeek", listLeaveWeek).
		Set("LeaveByProject", projectGroupByName).
		Set("LeaveByPerson", leaveByPerson).
		Set("NameByProject", nameList).
		Set("LeaveByLocation", locationGroup).
		Set("TopFive", topFive), nil
}

func (s *DashboardService) GetRemoteForDashboard(payload PayloadDashboardService) (tk.M, error) {
	pipeUser := []tk.M{}
	preMatchUser := tk.M{}
	userIds := []string{}
	location := payload.Location
	if location != "" {
		preMatchUser.Set("location", location)
	}

	// designations := payload.Designations

	// if len(designations) > 0 {
	// 	preMatchUser.Set("designation", tk.M{}.Set("$in", designations))
	// }
	// role an rules by project
	if payload.UserId != "" {
		userIds = append(userIds, payload.UserId)
	}
	monthyear := payload.MonthYear
	if monthyear == "" {
		monthyear = time.Now().Format("012006")
	}
	{
		// sample/
		// monthyear = "112018"
		//===
	}
	beginMonth, _ := time.Parse("012006", monthyear)
	listdateInMonth := []string{}
	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	}
	if !payload.UserOnly {
		pipePr := []tk.M{}
		switch payload.Level {
		case 1:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		case 2, 3:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":         bson.NewObjectId(),
				"projectname": "$ProjectName",
				"useriddev":   "$Developer.userid",
				"empiddev":    "$Developer.IdEmp",
				"namedev":     "$Developer.Name",
			}))
		case 5, 6:
			// pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		}
		// tk.Println("-------------- payload ", payload.Level)
		if payload.Level != 0 {
			switch payload.Level {
			case 2:
				// tk.Println("-------------- masuk 1")
				projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
				//tk.Println("resmote", projectDatas, payload.Level, err)
				if err != nil {
					return nil, err
				}

				proj := []string{}
				for _, each := range projectDatas {

					proj = append(proj, each.GetString("projectname"))

				}
				piperj := []tk.M{}
				if payload.Level == 2 {
					piperj = append(piperj, tk.M{"$unwind": "$projects"})
					piperj = append(piperj, tk.M{"$match": tk.M{"projects.projectname": tk.M{"$in": proj}}})
					piperj = append(piperj, tk.M{"$match": tk.M{"dateleave": tk.M{"$in": listdateInMonth}}})
					piperj = append(piperj, tk.M{"$group": tk.M{"_id": "$userid"}})
					datarj, err := new(repositories.ProjectOrmRepo).GetByDateRemote(piperj)
					if err != nil {
						return nil, err
					}

					for _, usrprj := range datarj {
						// tk.Println("----------------- datarj ", usrprj.GetString("_id"))
						// if user == usrprj.GetString("userid") {
						userIds = append(userIds, usrprj.GetString("_id"))
						// }
					}
				}
			case 3:
				// tk.Println("-------------- masuk 2")
				preMatchUser.Set("_id", tk.M{}.Set("$eq", payload.UserId))
			case 5, 6, 1:
				// tk.Println("-------------- masuk 3")
				piperj := []tk.M{}
				// piperj = append(piperj, tk.M{"$unwind": "$projects"})
				// piperj = append(piperj, tk.M{"$match": tk.M{"projects.projectname": tk.M{"$in": proj}}})
				piperj = append(piperj, tk.M{"$match": tk.M{"dateleave": tk.M{"$in": listdateInMonth}}})
				piperj = append(piperj, tk.M{"$group": tk.M{"_id": "$userid"}})
				datarj, err := new(repositories.ProjectOrmRepo).GetByDateRemote(piperj)
				if err != nil {
					return nil, err
				}

				for _, usrprj := range datarj {
					// tk.Println("----------------- datarj ", usrprj.GetString("_id"))
					// if user == usrprj.GetString("userid") {
					userIds = append(userIds, usrprj.GetString("_id"))
					// }
				}
			}

			// for _, each := range projectDatas {
			// 	if payload.Level == 2 {
			// 		userIds = append(userIds, each.GetString("useridleader"))
			// 		for _, each := range projectDatas {

			// 			proj = append(proj, each.GetString("projectname"))

			// 		}
			// 	}
			// 	skip := false
			// 	for _, user := range userIds {
			// 		if user == each.GetString("useriddev") {
			// 			skip = true
			// 			break
			// 		}
			// 	}
			// 	if !skip {
			// 		userIds = append(userIds, each.GetString("useriddev"))
			// 	}
			// }

			// tk.Println("------------- userid ", newUserIds)

			// preMatchUser.Set("$unwind", "$projects")
			// preMatchUser.Set("projects.projectname", tk.M{}.Set("$in", proj))
		}
		//===================================== end condition ============================================================
	} else {
		preMatchUser.Set("_id", tk.M{}.Set("$eq", payload.UserId))
	}

	if payload.Search == "" {
		newUserIds := helper.DistincValue(userIds)
		preMatchUser.Set("userid", tk.M{}.Set("$in", newUserIds))
	}

	if payload.UserId != "" {
		tk.Println("----------- users 2 ", payload.UserId)
		// preMatchUser.Set("_id", payload.UserId)
		// preMatchUser.Set("userid", tk.M{}.Set("$in", []string{payload.UserId}))
	}
	if payload.Search != "" {

		preMatchUser.Set("fullname", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
	}

	matchUser := tk.M{}.Set("$match", preMatchUser)
	pipeUser = append(pipeUser, matchUser)
	tk.Println("----------- users ", pipeUser)
	users, err := new(repositories.UserOrmRepo).GetByPipe(pipeUser)

	if err != nil {
		return nil, err
	}

	tk.Println("--------- user 3 ", users)
	idsUser := []string{}

	if len(users) <= 0 {
		if payload.Search == "" {
			for _, u := range userIds {
				idsUser = append(idsUser, u)
			}
		} else {
			return nil, errors.New("user not found")
		}

	} else {
		for _, user1 := range users {
			for _, usr1 := range userIds {
				if usr1 == user1.Id {
					idsUser = append(idsUser, user1.Id)
				}
			}

		}
	}

	project := payload.Project
	preMatch := tk.M{}.Set("userid", tk.M{}.Set("$in", idsUser))
	if project != "" {
		preMatch.Set("projects.projectname", project)
	}
	preMatch.Set("projects.isapprovalmanager", true)
	preMatch.Set("isdelete", false)
	if location == "" || location == "Global" {
		preMatch.Set("location", tk.M{}.Set("$ne", ""))
	} else {
		preMatch.Set("location", location)
	}
	//get date in month
	// monthyear := payload.MonthYear
	// if monthyear == "" {
	// 	monthyear = time.Now().Format("012006")
	// }
	// {
	// 	// sample/
	// 	// monthyear = "112018"
	// 	//===
	// }
	// beginMonth, _ := time.Parse("012006", monthyear)
	// listdateInMonth := []string{}
	// for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
	// 	listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	// }
	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))
	match := tk.M{}.Set("$match", preMatch)
	pipe := []tk.M{}
	pipe = append(pipe, match)
	remotes, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)

	if err != nil {
		return nil, err
	}

	datas := map[string]tk.M{}
	dataUsers := map[string]tk.M{}

	for _, remote := range remotes {
		dateLeave := remote.DateLeave
		if _, ok := datas[dateLeave]; !ok {
			datas[dateLeave] = tk.M{}.Set("Count", 0).
				Set("Rows", []RemoteModel{}).
				Set("Date", remote.DateLeave).
				Set("From", remote.DateLeave).
				Set("To", remote.DateLeave).
				Set("Projects", map[string][]RemoteModel{})
		}

		if _, ok := datas[remote.UserId]; !ok {
			dataUsers[remote.UserId] = tk.M{}.Set("Detail", RemoteModel{}).Set("ListOfRemote", map[string]string{})
		}

		countRemote := datas[dateLeave].GetInt("Count") + 1
		datas[dateLeave].Set("Count", countRemote)
		rows := datas[dateLeave].Get("Rows").([]RemoteModel)
		rows = append(rows, remote)
		datas[dateLeave].Set("Rows", rows)

		projects := datas[dateLeave].Get("Projects").(map[string][]RemoteModel)
		for _, project := range remote.Projects {
			projectName := project.ProjectName
			if _, ok := projects[projectName]; !ok {
				projects[projectName] = []RemoteModel{}
			}

			projects[projectName] = append(projects[projectName], remote)
		}

		datas[dateLeave].Set("Projects", projects)

		dataUsers[remote.UserId].Set("Detail", remote)
		listDateUsers := dataUsers[remote.UserId].Get("ListOfRemote").(map[string]string)
		listDateUsers[remote.DateLeave] = remote.DateLeave
		dataUsers[remote.UserId].Set("ListOfRemote", listDateUsers)
	}

	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
}

// func (s *DashboardService) GetRemoteForDashboard(payload PayloadDashboardService) (tk.M, error) {
// 	pipeUser := []tk.M{}
// 	preMatchUser := tk.M{}
// 	userIds := []string{}
// 	location := payload.Location
// 	// if location != "" {
// 	// 	preMatchUser.Set("location", location)
// 	// }

// 	designations := payload.Designations

// 	if len(designations) > 0 {
// 		preMatchUser.Set("designation", tk.M{}.Set("$in", designations))
// 	}
// 	// role an rules by project
// 	if payload.UserId != "" {
// 		userIds = append(userIds, payload.UserId)
// 	}
// 	if !payload.UserOnly {
// 		pipePr := []tk.M{}
// 		switch payload.Level {
// 		case 1:
// 			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
// 			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
// 			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
// 				// "_id":          bson.NewObjectId(),
// 				"projectname":  "$ProjectName",
// 				"useridleader": "$ProjectLeader.userid",
// 				"empidleader":  "$ProjectLeader.IdEmp",
// 				"nameleader":   "$ProjectLeader.Name",
// 				"useriddev":    "$Developer.userid",
// 				"empiddev":     "$Developer.IdEmp",
// 				"namedev":      "$Developer.Name",
// 			}))
// 		case 2, 3:
// 			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
// 			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
// 			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
// 				// "_id":         bson.NewObjectId(),
// 				"projectname": "$ProjectName",
// 				"useriddev":   "$Developer.userid",
// 				"empiddev":    "$Developer.IdEmp",
// 				"namedev":     "$Developer.Name",
// 			}))
// 		}

// 		if payload.UserId != "" && payload.Level != 5 && payload.Level != 6 && payload.Level != 1 {
// 			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
// 			//tk.Println("resmote", projectDatas, payload.Level, err)
// 			if err != nil {
// 				return nil, err
// 			}

// 			for _, each := range projectDatas {
// 				if payload.Level == 2 {
// 					userIds = append(userIds, each.GetString("useridleader"))
// 				}
// 				skip := false
// 				for _, user := range userIds {
// 					if user == each.GetString("useriddev") {
// 						skip = true
// 						break
// 					}
// 				}
// 				if !skip {
// 					userIds = append(userIds, each.GetString("useriddev"))
// 				}
// 			}
// 			newUserIds := helper.DistincValue(userIds)
// 			// tk.Println(newUserIds)
// 			preMatchUser.Set("_id", tk.M{}.Set("$in", newUserIds))
// 		}
// 		//===================================== end condition ============================================================
// 	} else {
// 		preMatchUser.Set("_id", tk.M{}.Set("$eq", payload.UserId))
// 	}

// 	// if payload.UserId != "" {
// 	// 	preMatchUser.Set("_id", payload.UserId)
// 	// }
// 	if payload.Search != "" {
// 		preMatchUser.Set("fullname", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
// 	}
// 	matchUser := tk.M{}.Set("$match", preMatchUser)
// 	pipeUser = append(pipeUser, matchUser)

// 	users, err := new(repositories.UserOrmRepo).GetByPipe(pipeUser)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(users) <= 0 {
// 		return nil, errors.New("user not found")
// 	}
// 	idsUser := []string{}
// 	for _, user := range users {
// 		idsUser = append(idsUser, user.Id)
// 	}

// 	project := payload.Project
// 	preMatch := tk.M{}.Set("userid", tk.M{}.Set("$in", idsUser))
// 	if project != "" {
// 		preMatch.Set("projects.projectname", project)
// 	}
// 	preMatch.Set("projects.isapprovalmanager", true)
// 	preMatch.Set("isdelete", false)
// 	if location == "" || location == "Global" {
// 		preMatch.Set("location", tk.M{}.Set("$ne", ""))
// 	} else {
// 		preMatch.Set("location", location)
// 	}
// 	//get date in month
// 	monthyear := payload.MonthYear
// 	if monthyear == "" {
// 		monthyear = time.Now().Format("012006")
// 	}
// 	{
// 		// sample/
// 		// monthyear = "112018"
// 		//===
// 	}
// 	beginMonth, _ := time.Parse("012006", monthyear)
// 	listdateInMonth := []string{}
// 	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
// 		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
// 	}
// 	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))
// 	match := tk.M{}.Set("$match", preMatch)
// 	pipe := []tk.M{}
// 	pipe = append(pipe, match)
// 	remotes, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)

// 	if err != nil {
// 		return nil, err
// 	}

// 	datas := map[string]tk.M{}
// 	dataUsers := map[string]tk.M{}

// 	for _, remote := range remotes {
// 		dateLeave := remote.DateLeave
// 		if _, ok := datas[dateLeave]; !ok {
// 			datas[dateLeave] = tk.M{}.Set("Count", 0).
// 				Set("Rows", []RemoteModel{}).
// 				Set("Date", remote.DateLeave).
// 				Set("From", remote.DateLeave).
// 				Set("To", remote.DateLeave).
// 				Set("Projects", map[string][]RemoteModel{})
// 		}

// 		if _, ok := datas[remote.UserId]; !ok {
// 			dataUsers[remote.UserId] = tk.M{}.Set("Detail", RemoteModel{}).Set("ListOfRemote", map[string]string{})
// 		}

// 		countRemote := datas[dateLeave].GetInt("Count") + 1
// 		datas[dateLeave].Set("Count", countRemote)
// 		rows := datas[dateLeave].Get("Rows").([]RemoteModel)
// 		rows = append(rows, remote)
// 		datas[dateLeave].Set("Rows", rows)

// 		projects := datas[dateLeave].Get("Projects").(map[string][]RemoteModel)
// 		for _, project := range remote.Projects {
// 			projectName := project.ProjectName
// 			if _, ok := projects[projectName]; !ok {
// 				projects[projectName] = []RemoteModel{}
// 			}

// 			projects[projectName] = append(projects[projectName], remote)
// 		}

// 		datas[dateLeave].Set("Projects", projects)

// 		dataUsers[remote.UserId].Set("Detail", remote)
// 		listDateUsers := dataUsers[remote.UserId].Get("ListOfRemote").(map[string]string)
// 		listDateUsers[remote.DateLeave] = remote.DateLeave
// 		dataUsers[remote.UserId].Set("ListOfRemote", listDateUsers)
// 	}

// 	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
// }

func (s *DashboardService) GetRequestEmergencyLeave(payload PayloadDashboardService) (tk.M, error) {
	location := payload.Location
	projects := []string{}
	projectTemp := payload.Project
	designations := payload.Designations

	userIds := []string{}

	users := []SysUserModel{}
	var err error

	if len(designations) > 0 {
		pipeUser := []tk.M{}
		match := tk.M{}.Set("$match", tk.M{}.Set("designation", tk.M{}.Set("$in", designations)))
		pipeUser = append(pipeUser, match)

		users, err = new(repositories.UserOrmRepo).GetByPipe(pipeUser)
	}

	for _, user := range users {
		userIds = append(userIds, user.Id)
	}

	pipe := []tk.M{}

	preMatch := tk.M{}.Set("location", location)
	preMatch.Set("stsbymanager", tk.M{}.Set("$eq", "Approved"))
	// preMatch.Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))
	preMatch.Set("isemergency", tk.M{}.Set("$eq", true))
	if location == "" || location == "Global" {
		preMatch.Set("location", tk.M{}.Set("$ne", ""))
	} else {
		preMatch.Set("location", location)
	}
	if projectTemp != "" {
		projects = append(projects, projectTemp)
	}

	if len(projects) > 0 {
		preMatch.Set("project", tk.M{}.Set("$in", projects))
	}
	// role an rules by project
	if payload.UserId != "" {
		userIds = append(userIds, payload.UserId)
	}

	monthyear := payload.MonthYear
	if monthyear == "" {
		monthyear = time.Now().Format("012006")
	}
	{
		// sample/
		// monthyear = "112018"
		//===
	}
	beginMonth, _ := time.Parse("012006", monthyear)
	listdateInMonth := []string{}
	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	}

	if !payload.UserOnly {
		pipePr := []tk.M{}
		switch payload.Level {
		case 1:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		case 2, 3:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":         bson.NewObjectId(),
				"projectname": "$ProjectName",
				"useriddev":   "$Developer.userid",
				"empiddev":    "$Developer.IdEmp",
				"namedev":     "$Developer.Name",
			}))
		}

		if payload.UserId != "" && payload.Level != 5 && payload.Level != 6 && payload.Level != 1 {
			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
			if err != nil {
				return nil, err
			}
			pipelv := []tk.M{}
			proj := []string{}

			for _, each := range projectDatas {

				proj = append(proj, each.GetString("projectname"))

			}
			// tk.Println("Projjjjjj ---- ", proj)
			pipelv = append(pipelv, tk.M{"$match": tk.M{"project": tk.M{"$in": proj}}})
			pipelv = append(pipelv, tk.M{"$match": tk.M{"dateleave": tk.M{"$in": listdateInMonth}}})
			pipelv = append(pipelv, tk.M{"$group": tk.M{"_id": "$userid"}})
			datalv, err := new(repositories.ProjectOrmRepo).GetByDateleave(pipelv)
			if err != nil {
				return nil, err
			}

			for _, usrlv := range datalv {
				// if user == usrlv.GetString("_id") {
				userIds = append(userIds, usrlv.GetString("_id"))
				// }
			}

			newUserIds := helper.DistincValue(userIds)
			if payload.Level == 2 {
				preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
				// preMatch.Set("project", tk.M{}.Set("$in", proj))
			} else {
				preMatch.Set("userid", tk.M{}.Set("$in", userIds))
			}
		}
		//===================================== end condition ============================================================
	} else {
		preMatch.Set("userid", tk.M{}.Set("$eq", payload.UserId))
	}

	// if payload.UserId != "" {
	// 	preMatch.Set("userid", payload.UserId)
	// } else if len(userIds) > 0 {
	// 	preMatch.Set("userid", tk.M{}.Set("$in", userIds))
	// }
	if payload.Search != "" {
		preMatch.Set("name", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
	}

	preMatch.Set("isdelete", false)
	// monthyear := payload.MonthYear
	// if monthyear == "" {
	// 	monthyear = time.Now().Format("012006")
	// }
	// {
	// 	// sample/
	// 	// monthyear = "112018"
	// 	//===
	// }
	// beginMonth, _ := time.Parse("012006", monthyear)
	// listdateInMonth := []string{}
	// for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
	// 	listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	// }
	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))
	match := tk.M{}.Set("$match", preMatch)
	pipe = append(pipe, match)

	leaves, err := new(repositories.LeaveDboxRepo).GetByPipe(pipe)
	if err != nil {
		return nil, err
	}
	datas := map[string]tk.M{}
	dataUsers := map[string]tk.M{}

	for _, leave := range leaves {
		if _, ok := datas[leave.DateLeave]; !ok {
			datas[leave.DateLeave] = tk.M{}.
				Set("Count", 0).
				Set("Rows", []AprovalRequestLeaveModel{}).
				Set("Date", leave.DateLeave).
				Set("From", leave.DateLeave).
				Set("To", leave.DateLeave).
				Set("Projects", tk.M{})
		}

		if _, ok := dataUsers[leave.EmpId]; !ok {
			dataUsers[leave.EmpId] = tk.M{}.Set("Detail", AprovalRequestLeaveModel{}).Set("DateOfLeave", map[string]string{})
		}

		countLeave := datas[leave.DateLeave].GetInt("Count") + 1
		rowLeaves := datas[leave.DateLeave].Get("Rows").([]AprovalRequestLeaveModel)
		rowLeaves = append(rowLeaves, leave)
		datas[leave.DateLeave].Set("Count", countLeave)
		datas[leave.DateLeave].Set("Rows", rowLeaves)

		projects := datas[leave.DateLeave].Get("Projects").(tk.M)
		for _, project := range leave.Project {

			projects := datas[leave.DateLeave].Get("Projects").(tk.M)
			if _, ok := projects[project]; !ok {
				projects.Set(project, []AprovalRequestLeaveModel{})
			}

			rowProjects := projects.Get(project).([]AprovalRequestLeaveModel)
			rowProjects = append(rowProjects, leave)

			projects.Set(project, rowProjects)
		}
		datas[leave.DateLeave].Set("Projects", projects)

		dataUsers[leave.EmpId].Set("Detail", leave)
		datesByUser := dataUsers[leave.EmpId].Get("DateOfLeave").(map[string]string)
		datesByUser[leave.DateLeave] = leave.DateLeave
		dataUsers[leave.EmpId].Set("DateOfLeave", datesByUser)
	}

	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
}

// func (s *DashboardService) GetRequestSpecialLeave(payload PayloadDashboardService) (tk.M, error) {
// 	location := payload.Location
// 	projects := []string{}
// 	projectTemp := payload.Project
// 	designations := payload.Designations
// 	userIds := []string{}

// 	users := []SysUserModel{}
// 	var err error

// 	if len(designations) > 0 {
// 		pipeUser := []tk.M{}
// 		match := tk.M{}.Set("$match", tk.M{}.Set("designation", tk.M{}.Set("$in", designations)))
// 		pipeUser = append(pipeUser, match)

// 		users, err = new(repositories.UserOrmRepo).GetByPipe(pipeUser)
// 	}

// 	for _, user := range users {
// 		userIds = append(userIds, user.Id)
// 	}
// 	datas := map[string]tk.M{}
// 	dataUsers := map[string]tk.M{}

// 	pipe := []tk.M{}

// 	preMatch := tk.M{}.Set("isemergency", false)
// 	preMatch.Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))
// 	preMatch.Set("stsbymanager", tk.M{}.Set("$eq", "Approved"))

// 	if location == "" || location == "Global" {
// 		preMatch.Set("location", tk.M{}.Set("$ne", ""))
// 	} else {
// 		preMatch.Set("location", location)
// 	}
// 	if projectTemp != "" {
// 		projects = append(projects, projectTemp)
// 	}

// 	if len(projects) > 0 {
// 		preMatch.Set("project", tk.M{}.Set("$in", projects))
// 	}
// 	// role an rules by project
// 	monthyear := payload.MonthYear
// 	if monthyear == "" {
// 		monthyear = time.Now().Format("012006")
// 	}
// 	{
// 		// sample/
// 		// monthyear = "112018"
// 		//===
// 	}
// 	beginMonth, _ := time.Parse("012006", monthyear)
// 	listdateInMonth := []string{}
// 	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
// 		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
// 	}
// 	if payload.UserId != "" {
// 		userIds = append(userIds, payload.UserId)
// 	}
// 	// tk.Println("----------------->", payload.Level)
// 	if !payload.UserOnly {
// 		pipePr := []tk.M{}
// 		switch payload.Level {
// 		case 1:
// 			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
// 			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
// 			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
// 				// "_id":          bson.NewObjectId(),
// 				"projectname":  "$ProjectName",
// 				"useridleader": "$ProjectLeader.userid",
// 				"empidleader":  "$ProjectLeader.IdEmp",
// 				"nameleader":   "$ProjectLeader.Name",
// 				"useriddev":    "$Developer.userid",
// 				"empiddev":     "$Developer.IdEmp",
// 				"namedev":      "$Developer.Name",
// 			}))
// 		case 2, 3:
// 			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
// 			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
// 			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
// 				// "_id":         bson.NewObjectId(),
// 				"projectname": "$ProjectName",
// 				"useriddev":   "$Developer.userid",
// 				"empiddev":    "$Developer.IdEmp",
// 				"namedev":     "$Developer.Name",
// 			}))
// 		}

// 		if payload.UserId != "" && payload.Level != 5 && payload.Level != 6 && payload.Level != 1 {
// 			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
// 			if err != nil {
// 				return nil, err
// 			}
// 			pipelv := []tk.M{}
// 			proj := []string{}

// 			for _, each := range projectDatas {

// 				proj = append(proj, each.GetString("projectname"))

// 			}
// 			// tk.Println("Projjjjjj ---- ", proj)
// 			pipelv = append(pipelv, tk.M{"$match": tk.M{"project": tk.M{"$in": proj}}})
// 			pipelv = append(pipelv, tk.M{"$match": tk.M{"dateleave": tk.M{"$in": listdateInMonth}}})
// 			pipelv = append(pipelv, tk.M{"$group": tk.M{"_id": "$userid"}})
// 			datalv, err := new(repositories.ProjectOrmRepo).GetByDateleave(pipelv)
// 			if err != nil {
// 				return nil, err
// 			}

// 			for _, usrlv := range datalv {
// 				// if user == usrlv.GetString("_id") {
// 				userIds = append(userIds, usrlv.GetString("_id"))
// 				// }
// 			}

// 			newUserIds := helper.DistincValue(userIds)
// 			preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
// 			preMatch.Set("project", tk.M{}.Set("$in", proj))
// 		}
// 		//===================================== end condition ============================================================
// 	} else {
// 		preMatch.Set("userid", tk.M{}.Set("$eq", payload.UserId))
// 	}

// 	// if payload.UserId != "" {
// 	// 	preMatch.Set("userid", payload.UserId)
// 	// } else if len(userIds) > 0 {
// 	// preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
// 	// }
// 	if payload.Search != "" {
// 		preMatch.Set("name", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
// 	}

// 	preMatch.Set("isdelete", false)
// 	// monthyear := payload.MonthYear
// 	// if monthyear == "" {
// 	// 	monthyear = time.Now().Format("012006")
// 	// }
// 	// {
// 	// 	// sample/
// 	// 	// monthyear = "112018"
// 	// 	//===
// 	// }
// 	// beginMonth, _ := time.Parse("012006", monthyear)
// 	// listdateInMonth := []string{}
// 	// for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
// 	// 	listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
// 	// }
// 	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))

// 	match := tk.M{}.Set("$match", preMatch)
// 	pipe = append(pipe, match)

// 	// fmt.Println("pipe", string(tk.Jsonify(pipe)))

// 	leaves, err := new(repositories.LeaveDboxRepo).GetByPipeSpecial(pipe, true)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tk.Println("[]=== array spesial", leaves)
// 	for _, leave := range leaves {
// 		// tk.Println("leave", leave)
// 		if _, ok := datas[leave.DateLeave]; !ok {
// 			datas[leave.DateLeave] = tk.M{}.
// 				Set("Count", 0).
// 				Set("Rows", []AprovalRequestLeaveModel{}).
// 				Set("Date", leave.DateLeave).
// 				Set("From", leave.DateLeave).
// 				Set("To", leave.DateLeave).
// 				Set("Projects", tk.M{})
// 		}

// 		if _, ok := dataUsers[leave.EmpId]; !ok {
// 			dataUsers[leave.EmpId] = tk.M{}.Set("Detail", AprovalRequestLeaveModel{}).Set("DateOfLeave", map[string]string{})
// 		}

// 		countLeave := datas[leave.DateLeave].GetInt("Count") + 1
// 		rowLeaves := datas[leave.DateLeave].Get("Rows").([]AprovalRequestLeaveModel)
// 		rowLeaves = append(rowLeaves, leave)
// 		datas[leave.DateLeave].Set("Count", countLeave)
// 		datas[leave.DateLeave].Set("Rows", rowLeaves)

// 		projects := datas[leave.DateLeave].Get("Projects").(tk.M)
// 		for _, project := range leave.Project {
// 			// tk.Println("pro", project)
// 			projects := datas[leave.DateLeave].Get("Projects").(tk.M)
// 			if _, ok := projects[project]; !ok {
// 				projects.Set(project, []AprovalRequestLeaveModel{})
// 			}

// 			rowProjects := projects.Get(project).([]AprovalRequestLeaveModel)
// 			rowProjects = append(rowProjects, leave)

// 			projects.Set(project, rowProjects)
// 		}
// 		datas[leave.DateLeave].Set("Projects", projects)

// 		dataUsers[leave.EmpId].Set("Detail", leave)
// 		datesByUser := dataUsers[leave.EmpId].Get("DateOfLeave").(map[string]string)
// 		datesByUser[leave.DateLeave] = leave.DateLeave
// 		dataUsers[leave.EmpId].Set("DateOfLeave", datesByUser)
// 	}

// 	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
// }

func (s *DashboardService) GetRequestSpecialLeave(payload PayloadDashboardService) (tk.M, error) {
	location := payload.Location
	projects := []string{}
	projectTemp := payload.Project
	designations := payload.Designations
	userIds := []string{}

	users := []SysUserModel{}
	var err error
	monthyear := payload.MonthYear
	beginMonth, _ := time.Parse("012006", monthyear)
	listdateInMonth := []string{}
	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	}

	if len(designations) > 0 {
		pipeUser := []tk.M{}
		match := tk.M{}.Set("$match", tk.M{}.Set("designation", tk.M{}.Set("$in", designations)))
		pipeUser = append(pipeUser, match)

		users, err = new(repositories.UserOrmRepo).GetByPipe(pipeUser)
	}

	for _, user := range users {
		userIds = append(userIds, user.Id)
	}
	datas := map[string]tk.M{}
	dataUsers := map[string]tk.M{}

	pipe := []tk.M{}

	preMatch := tk.M{}.Set("isemergency", false)
	preMatch.Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))
	preMatch.Set("stsbymanager", tk.M{}.Set("$eq", "Approved"))

	if location == "" || location == "Global" {
		preMatch.Set("location", tk.M{}.Set("$ne", ""))
	} else {
		preMatch.Set("location", location)
	}
	if projectTemp != "" {
		projects = append(projects, projectTemp)
	}

	if len(projects) > 0 {
		preMatch.Set("project", tk.M{}.Set("$in", projects))
	}
	// role an rules by project
	if payload.UserId != "" {
		userIds = append(userIds, payload.UserId)
	}
	// tk.Println("----------------->", payload.Level)
	// tk.Println("rrr----------------->", payload.UserOnly)
	if !payload.UserOnly {
		pipePr := []tk.M{}
		switch payload.Level {
		case 1:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		case 2, 3:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			// pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
			// 	// "_id":         bson.NewObjectId(),
			// 	"projectname": "$ProjectName",
			// 	"useriddev":   "$Developer.userid",
			// 	"empiddev":    "$Developer.IdEmp",
			// 	"namedev":     "$Developer.Name",
			// }))
		}

		if payload.UserId != "" && payload.Level == 2 {
			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
			//tk.Println("ffff", projectDatas)
			if err != nil {
				return nil, err
			}
			// for _, each := range projectDatas {
			// 	// tk.Println(each)
			// 	if payload.Level == 2 {
			// 		userIds = append(userIds, each.GetString("useridleader"))
			// 	}
			// 	skip := false
			// 	for _, user := range userIds {
			// 		if user == each.GetString("useriddev") {
			// 			skip = true
			// 			break
			// 		}
			// 	}
			// 	if !skip {
			// 		userIds = append(userIds, each.GetString("useriddev"))
			// 	}
			// }
			pipeOv := []tk.M{}
			getproj := []string{}
			for _, each := range projectDatas {

				getproj = append(getproj, each.GetString("ProjectName"))

			}
			//tk.Println("ffff", getproj)
			//pipeOv = append(pipePr, tk.M{}.Set("$unwind": "$project"))
			pipeOv = append(pipeOv,
				tk.M{"$unwind": "$project"},
			)

			pipeOv = append(pipeOv,
				tk.M{
					"$match": tk.M{
						"$and": []tk.M{
							{"project": tk.M{"$in": getproj}},
							{"dateleave": tk.M{"$in": listdateInMonth}},
						},
					},
				},
			)
			//leaves, err := new(repositories.LeaveDboxRepo).GetByPipeSpecial(pipe, true)
			dataOv, err := new(repositories.LeaveDboxRepo).GetBySL(pipeOv)
			if err != nil {
				return nil, err
			}
			//tk.Println("tda....", dataOv)
			for _, df := range dataOv {
				if payload.Level == 2 {
					userIds = append(userIds, payload.UserId)
				}
				userIds = append(userIds, df.GetString("userid"))
			}

			newUserIds := helper.DistincValue(userIds)
			preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
		} else if payload.Level != 1 && payload.Level != 5 && payload.Level != 6 {
			preMatch.Set("userid", tk.M{}.Set("$eq", payload.UserId))
		}
		//===================================== end condition ============================================================
	} else {
		preMatch.Set("userid", tk.M{}.Set("$eq", payload.UserId))
	}

	// if payload.UserId != "" {
	// 	preMatch.Set("userid", payload.UserId)
	// } else if len(userIds) > 0 {
	// preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
	// }
	if payload.Search != "" {
		preMatch.Set("name", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
	}

	preMatch.Set("isdelete", false)

	if monthyear == "" {
		monthyear = time.Now().Format("012006")
	}
	{
		// sample/
		// monthyear = "112018"
		//===
	}

	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))

	match := tk.M{}.Set("$match", preMatch)
	pipe = append(pipe, match)

	// fmt.Println("pipe", string(tk.Jsonify(pipe)))

	leaves, err := new(repositories.LeaveDboxRepo).GetByPipeSpecial(pipe, true)
	if err != nil {
		return nil, err
	}
	// tk.Println("[]=== array", leaves)
	for _, leave := range leaves {
		// tk.Println("leave", leave)
		if _, ok := datas[leave.DateLeave]; !ok {
			datas[leave.DateLeave] = tk.M{}.
				Set("Count", 0).
				Set("Rows", []AprovalRequestLeaveModel{}).
				Set("Date", leave.DateLeave).
				Set("From", leave.DateLeave).
				Set("To", leave.DateLeave).
				Set("Projects", tk.M{})
		}

		if _, ok := dataUsers[leave.EmpId]; !ok {
			dataUsers[leave.EmpId] = tk.M{}.Set("Detail", AprovalRequestLeaveModel{}).Set("DateOfLeave", map[string]string{})
		}

		countLeave := datas[leave.DateLeave].GetInt("Count") + 1
		rowLeaves := datas[leave.DateLeave].Get("Rows").([]AprovalRequestLeaveModel)
		rowLeaves = append(rowLeaves, leave)
		datas[leave.DateLeave].Set("Count", countLeave)
		datas[leave.DateLeave].Set("Rows", rowLeaves)

		projects := datas[leave.DateLeave].Get("Projects").(tk.M)
		for _, project := range leave.Project {
			// tk.Println("pro", project)
			projects := datas[leave.DateLeave].Get("Projects").(tk.M)
			if _, ok := projects[project]; !ok {
				projects.Set(project, []AprovalRequestLeaveModel{})
			}

			rowProjects := projects.Get(project).([]AprovalRequestLeaveModel)
			rowProjects = append(rowProjects, leave)

			projects.Set(project, rowProjects)
		}
		datas[leave.DateLeave].Set("Projects", projects)

		dataUsers[leave.EmpId].Set("Detail", leave)
		datesByUser := dataUsers[leave.EmpId].Get("DateOfLeave").(map[string]string)
		datesByUser[leave.DateLeave] = leave.DateLeave
		dataUsers[leave.EmpId].Set("DateOfLeave", datesByUser)
	}

	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
}

func (s *DashboardService) GetRequestOvertime(payload PayloadDashboardService) (tk.M, error) {
	monthYear := payload.MonthYear
	getSearch := payload.Search
	searchReg := bson.RegEx{getSearch, "i"}
	searchLoc := payload.Location
	searchProj := payload.Project
	userIds := []string{}
	if monthYear == "" {
		monthYear = time.Now().Format("012006")
	}

	beginMonth, _ := time.Parse("012006", monthYear)
	listdateInMonth := []string{}
	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	}
	designations := payload.Designations
	preMatchUser := tk.M{}
	if len(designations) > 0 {
		preMatchUser.Set("designation", tk.M{}.Set("$in", designations))
	}
	pipe := []tk.M{}
	//add custome for user role
	preMatch := tk.M{}

	if !payload.UserOnly {
		//tk.Println("m22222222")
		pipePr := []tk.M{}
		switch payload.Level {
		case 1:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		case 2, 3:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
		}

		if payload.UserId != "" && payload.Level == 2 {
			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
			// tk.Println("resovertime", projectDatas, payload.Level, payload.EmpId, pipePr)
			if err != nil {
				return nil, err
			}

			pipeOv := []tk.M{}
			for _, each := range projectDatas {
				pipeOv = append(pipeOv,
					tk.M{
						"$match": tk.M{
							"$and": []tk.M{
								{"project": tk.M{"$eq": each.GetString("ProjectName")}},
								{"dateovertime": tk.M{"$in": listdateInMonth}},
							},
						},
					},
				)
				dataOv, err := new(repositories.ProjectOrmRepo).GetByOvertime(pipeOv)
				if err != nil {
					return nil, err
				}
				for _, df := range dataOv {
					if payload.Level == 2 {
						userIds = append(userIds, payload.UserId)
					}
					userIds = append(userIds, df.GetString("userid"))
				}
			}
			newUserIds := helper.DistincValue(userIds)
			//tk.Println("m22222", newUserIds, projectDatas)
			pipe = append(pipe,
				tk.M{
					"$match": tk.M{
						"$and": []tk.M{
							{"userid": tk.M{"$in": newUserIds}},
						},
					},
				},
			)
		} else if payload.Level != 1 && payload.Level != 5 && payload.Level != 6 {
			//tk.Println("thu..", payload.UserId, payload.Level, preMatchUser)
			pipe = append(pipe,
				tk.M{
					"$match": tk.M{
						"$and": []tk.M{
							{"userid": tk.M{"$eq": payload.UserId}},
						},
					},
				},
			)
		}
		//===================================== end condition =========================================
	} else {
		pipe = append(pipe,
			tk.M{
				"$match": tk.M{
					"$and": []tk.M{
						{"userid": tk.M{"$eq": payload.UserId}},
					},
				},
			},
		)
	}
	//032019
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$ne": ""}},
					{"resultmatch": tk.M{"$ne": "Cancelled"}},
				},
			},
		},
	)

	if getSearch != "" {
		pipe = append(pipe,
			tk.M{
				"$match": tk.M{
					"$and": []tk.M{
						{"name": tk.M{"$regex": searchReg}},
					},
				},
			},
		)
	}

	if searchLoc != "" {
		if searchLoc != "Global" {
			pipe = append(pipe,
				tk.M{
					"$match": tk.M{
						"$and": []tk.M{
							{"location": tk.M{"$eq": searchLoc}},
						},
					},
				},
			)
		}
	}

	if searchProj != "" {
		pipe = append(pipe,
			tk.M{
				"$match": tk.M{
					"$and": []tk.M{
						{"project": tk.M{"$eq": searchProj}},
					},
				},
			},
		)
	}

	pipe = append(pipe, tk.M{
		"$lookup": tk.M{
			"from":         "NewOvertime",
			"localField":   "idovertime",
			"foreignField": "_id",
			"as":           "getreason",
		},
	})
	//tk.Println("thu..", payload.UserId, payload.Level, preMatchUser)
	//preMatch := tk.M{}
	preMatch.Set("dateovertime", tk.M{}.Set("$in", listdateInMonth))

	match := tk.M{}.Set("$match", preMatch)

	pipe = append(pipe, match)
	data, err := new(repositories.ProjectOrmRepo).GetByOvertime(pipe)
	if err != nil {
		return nil, err
	}

	return tk.M{}.Set("Data", data), err
}

func (s *DashboardService) GetRequestLeave(payload PayloadDashboardService) (tk.M, error) {
	location := payload.Location
	projects := []string{}
	projectTemp := payload.Project
	designations := payload.Designations
	userIds := []string{}

	users := []SysUserModel{}
	var err error

	if len(designations) > 0 {
		pipeUser := []tk.M{}
		match := tk.M{}.Set("$match", tk.M{}.Set("designation", tk.M{}.Set("$in", designations)))
		pipeUser = append(pipeUser, match)

		users, err = new(repositories.UserOrmRepo).GetByPipe(pipeUser)
	}

	for _, user := range users {
		userIds = append(userIds, user.Id)
	}
	datas := map[string]tk.M{}
	dataUsers := map[string]tk.M{}

	pipe := []tk.M{}

	preMatch := tk.M{}.Set("isemergency", false)
	preMatch.Set("stsbymanager", tk.M{}.Set("$ne", "Pending"))
	preMatch.Set("stsbymanager", tk.M{}.Set("$eq", "Approved"))

	if location == "" || location == "Global" {
		preMatch.Set("location", tk.M{}.Set("$ne", ""))
	} else {
		preMatch.Set("location", location)
	}
	if projectTemp != "" {
		projects = append(projects, projectTemp)
	}

	if len(projects) > 0 {
		preMatch.Set("project", tk.M{}.Set("$in", projects))
	}
	// role an rules by project
	if payload.UserId != "" {
		userIds = append(userIds, payload.UserId)
	}
	// tk.Println("----------------->", payload.Level)
	monthyear := payload.MonthYear
	if monthyear == "" {
		monthyear = time.Now().Format("012006")
	}
	{
		// sample/
		// monthyear = "112018"
		//===
	}
	beginMonth, _ := time.Parse("012006", monthyear)
	listdateInMonth := []string{}
	for d := beginMonth; d.Month() == beginMonth.Month(); d = d.AddDate(0, 0, 1) {
		listdateInMonth = append(listdateInMonth, d.Format("2006-01-02"))
	}
	if !payload.UserOnly {
		pipePr := []tk.M{}
		switch payload.Level {
		case 1:
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectManager.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":          bson.NewObjectId(),
				"projectname":  "$ProjectName",
				"useridleader": "$ProjectLeader.userid",
				"empidleader":  "$ProjectLeader.IdEmp",
				"nameleader":   "$ProjectLeader.Name",
				"useriddev":    "$Developer.userid",
				"empiddev":     "$Developer.IdEmp",
				"namedev":      "$Developer.Name",
			}))
		case 2, 3:
			// tk.Println("---------------------- empid ", payload.EmpId)
			pipePr = append(pipePr, tk.M{}.Set("$match", tk.M{}.Set("ProjectLeader.IdEmp", payload.EmpId)))
			// pipePr = append(pipePr, tk.M{}.Set("$unwind", "$Developer"))
			pipePr = append(pipePr, tk.M{}.Set("$project", tk.M{
				// "_id":         bson.NewObjectId(),
				"projectname": "$ProjectName",
				"useriddev":   "$Developer.userid",
				"empiddev":    "$Developer.IdEmp",
				"namedev":     "$Developer.Name",
			}))
		}

		if payload.UserId != "" && payload.Level != 5 && payload.Level != 6 && payload.Level != 1 {
			projectDatas, err := new(repositories.ProjectOrmRepo).GetByPipeProjection(pipePr)
			// tk.Println("----------- pipepr ", projectDatas)
			if err != nil {
				return nil, err
			}
			pipelv := []tk.M{}
			proj := []string{}

			for _, each := range projectDatas {

				proj = append(proj, each.GetString("projectname"))

			}

			// proj = append(proj, "SIRS")
			// tk.Println("Projjjjjj ---- ", proj)
			pipelv = append(pipelv, tk.M{"$match": tk.M{"project": tk.M{"$in": proj}}})
			pipelv = append(pipelv, tk.M{"$match": tk.M{"dateleave": tk.M{"$in": listdateInMonth}}})
			pipelv = append(pipelv, tk.M{"$group": tk.M{"_id": "$userid"}})
			datalv, err := new(repositories.ProjectOrmRepo).GetByDateleave(pipelv)
			if err != nil {
				return nil, err
			}

			for _, usrlv := range datalv {
				// if user == usrlv.GetString("_id") {
				userIds = append(userIds, usrlv.GetString("_id"))
				// }
			}

			newUserIds := helper.DistincValue(userIds)
			if payload.Level == 2 {
				preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
				// preMatch.Set("project", tk.M{}.Set("$in", proj))
			} else {
				preMatch.Set("userid", tk.M{}.Set("$in", userIds))
			}

		}
		//===================================== end condition ============================================================
	} else {
		preMatch.Set("userid", tk.M{}.Set("$eq", payload.UserId))
	}

	// if payload.UserId != "" {
	// 	preMatch.Set("userid", payload.UserId)
	// } else if len(userIds) > 0 {
	// preMatch.Set("userid", tk.M{}.Set("$in", newUserIds))
	// }
	if payload.Search != "" {
		preMatch.Set("name", tk.M{}.Set("$regex", bson.RegEx{Pattern: payload.Search, Options: "i"}))
	}

	preMatch.Set("isdelete", false)
	// pipelv = append(pipelv, tk.M{"$match": tk.M{"project": tk.M{"$in": proj}}})
	preMatch.Set("dateleave", tk.M{}.Set("$in", listdateInMonth))

	match := tk.M{}.Set("$match", preMatch)
	pipe = append(pipe, match)

	// fmt.Println("pipe", string(tk.Jsonify(pipe)))

	leaves, err := new(repositories.LeaveDboxRepo).GetByPipeSpecial(pipe, false)
	if err != nil {
		return nil, err
	}
	// tk.Println("[]=== array ", tk.JsonString(leaves))
	for _, leave := range leaves {
		// tk.Println("leave", leave)
		if _, ok := datas[leave.DateLeave]; !ok {
			datas[leave.DateLeave] = tk.M{}.
				Set("Count", 0).
				Set("Rows", []AprovalRequestLeaveModel{}).
				Set("Date", leave.DateLeave).
				Set("From", leave.DateLeave).
				Set("To", leave.DateLeave).
				Set("Projects", tk.M{})
		}

		if _, ok := dataUsers[leave.EmpId]; !ok {
			dataUsers[leave.EmpId] = tk.M{}.Set("Detail", AprovalRequestLeaveModel{}).Set("DateOfLeave", map[string]string{})
		}

		countLeave := datas[leave.DateLeave].GetInt("Count") + 1
		rowLeaves := datas[leave.DateLeave].Get("Rows").([]AprovalRequestLeaveModel)
		rowLeaves = append(rowLeaves, leave)
		datas[leave.DateLeave].Set("Count", countLeave)
		datas[leave.DateLeave].Set("Rows", rowLeaves)

		projects := datas[leave.DateLeave].Get("Projects").(tk.M)
		for _, project := range leave.Project {
			// tk.Println("pro", project)
			projects := datas[leave.DateLeave].Get("Projects").(tk.M)
			if _, ok := projects[project]; !ok {
				projects.Set(project, []AprovalRequestLeaveModel{})
			}

			rowProjects := projects.Get(project).([]AprovalRequestLeaveModel)
			rowProjects = append(rowProjects, leave)

			projects.Set(project, rowProjects)
		}
		datas[leave.DateLeave].Set("Projects", projects)

		dataUsers[leave.EmpId].Set("Detail", leave)
		datesByUser := dataUsers[leave.EmpId].Get("DateOfLeave").(map[string]string)
		datesByUser[leave.DateLeave] = leave.DateLeave
		dataUsers[leave.EmpId].Set("DateOfLeave", datesByUser)
	}

	return tk.M{}.Set("DataByDate", datas).Set("DataByUser", dataUsers), err
}

func (s *DashboardService) GetLeaveNotApproved(dateByMonth, location string) ([]RequestLeaveModel, error) {
	rowsDate := []RequestLeaveModel{}
	monthFilter := dateByMonth + "-10"
	// tk.Println(monthFilter)
	currentDate, err := time.Parse("2006-1-2", monthFilter)
	if err != nil {
		return rowsDate, err
	}

	filter := tk.M{}.Set("where", dbox.And(dbox.Ne("statusmanagerproject.statusrequest", "Approved"), dbox.Eq("location", location)))
	if location == "" || location == "Global" {
		filter = tk.M{}.Set("where", dbox.And(dbox.Ne("statusmanagerproject.statusrequest", "Approved")))
	}
	leaves, err := new(repositories.LeaveOrmRepo).GetByParamMasterLeave(filter)
	if err != nil {
		return rowsDate, err
	}

	if len(leaves) > 0 {
		for _, leave := range leaves {
			leavefrom := leave.LeaveFrom
			if leave.LeaveFrom == "" {
				leavefrom = "0001-1-1"
			}
			startdate, err := time.Parse("2006-1-2", leavefrom)
			if err != nil {
				startdate, err = time.Parse("02-01-2006", leavefrom)
				if err != nil {
					return rowsDate, err
				}
			}
			leaveto := leave.LeaveTo
			if leave.LeaveTo == "" {
				leaveto = "0001-1-1"
			}
			enddate, err := time.Parse("2006-1-2", leaveto)
			// tk.Println("enddate", leave.LeaveTo, startdate, err)
			if err != nil {
				enddate, err = time.Parse("02-01-2006", leaveto)
				if err != nil {
					return rowsDate, err
				}
			}

			for !startdate.After(enddate) {
				if startdate.Month() != currentDate.Month() {
					break
				}

				leave.DateCreateLeave = startdate.Format("2016-01-02")
				rowsDate = append(rowsDate, leave)
				startdate = startdate.AddDate(0, 0, 1)
			}
		}
	}

	return rowsDate, nil
}

func (s *DashboardService) WriteExcelForAdmin(payload PayloadDashboardService, fileName string) (string, error) {
	fileLocation := "assets/doc/" + fileName
	// os.Remove(fileLocation)

	var file *xlsx.File
	var sheet, sheet2, sheet3, sheet4, sheet5 *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error
	styleTitle := *xlsx.NewStyle()
	style := *xlsx.NewStyle()
	styleHeader := *xlsx.NewStyle()

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("All Request")
	if err != nil {
		return "", err
	}

	sheet2, err = file.AddSheet("By Status")
	if err != nil {
		return "", err
	}

	sheet3, err = file.AddSheet("By Person")
	if err != nil {
		return "", err
	}

	sheet4, err = file.AddSheet("By Project")
	if err != nil {
		return "", err
	}

	sheet5, err = file.AddSheet("By Location")
	if err != nil {
		return "", err
	}

	datas, err := s.ConstructDashboardDataForAdminReport(payload)
	if err != nil {
		return "", err
	}

	monthNow, err := time.Parse("2006-01-02 ", payload.DateByMonth+"-01")
	if err != nil {
		return "", err
	}

	reportTitle := "Report " + monthNow.Format("January 2006")

	styleTitle.Font.Size = 18
	styleTitle.Font.Bold = true

	style.Border.Bottom = "thin"
	style.Border.Top = "thin"
	style.Border.Right = "thin"
	style.Border.Left = "thin"

	styleHeader.Border.Bottom = "thin"
	styleHeader.Border.Top = "thin"
	styleHeader.Border.Right = "thin"
	styleHeader.Border.Left = "thin"
	styleHeader.Font.Bold = true

	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = reportTitle
	cell.Merge(3, 0)
	cell.SetStyle(&styleTitle)

	row = sheet.AddRow()
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "All Request"
	cell.SetStyle(&styleHeader)
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.SetInt(datas.GetInt("CountAll"))
	cell.SetStyle(&style)

	row = sheet.AddRow()
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "Request Type Summary"
	cell.Merge(2, 0)
	row = sheet.AddRow()

	countApproved := datas.Get("CountLeaveApproved").(tk.M)
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)
	row = sheet.AddRow()

	cell = row.AddCell()
	cell.SetInt(countApproved.GetInt("Leave"))
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countApproved.GetInt("ELeave"))
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countApproved.GetInt("Remote"))
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countApproved.GetInt("Overtime"))
	cell.SetStyle(&style)

	row = sheet.AddRow()
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "Leave By Week"
	cell.Merge(3, 0)
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "Week"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)

	listWeek := datas.Get("LeaveByWeek").(map[int]tk.M)

	for k, v := range listWeek {
		row = sheet.AddRow()
		cell = row.AddCell()
		cell.Value = "week " + strconv.Itoa(k)
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Leave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("ELeave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Remote"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Overtime"))
		cell.SetStyle(&style)
	}

	countLeaveByStatus := datas.Get("CountByStatus").(tk.M)

	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = reportTitle
	cell.Merge(1, 0)
	cell.SetStyle(&styleTitle)

	row = sheet2.AddRow()
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Total Request By Status"
	cell.Merge(1, 0)
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Status"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Total"
	cell.SetStyle(&styleHeader)
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Approved"
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countLeaveByStatus.GetInt("Approved"))
	cell.SetStyle(&style)
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Decline"
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countLeaveByStatus.GetInt("Decline"))
	cell.SetStyle(&style)
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Cancelled"
	cell.SetStyle(&style)
	cell = row.AddCell()
	cell.SetInt(countLeaveByStatus.GetInt("Cancelled"))
	cell.SetStyle(&style)

	declineList := datas.Get("DeclineList").([]tk.M)

	row = sheet2.AddRow()
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Detail Decline"
	cell.Merge(1, 0)
	row = sheet2.AddRow()
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Decline Reason"
	cell.SetStyle(&styleHeader)

	for _, v := range declineList {
		row = sheet2.AddRow()
		cell = row.AddCell()
		cell.Value = v.GetString("Name")
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.Value = v.GetString("Reason")
		cell.SetStyle(&style)
	}

	countLeaveByPerson := datas.Get("LeaveByPerson").(map[string]tk.M)

	row = sheet3.AddRow()
	cell = row.AddCell()
	cell.Value = reportTitle
	cell.Merge(3, 0)
	cell.SetStyle(&styleTitle)

	row = sheet3.AddRow()
	row = sheet3.AddRow()
	cell = row.AddCell()
	cell.Value = "Total Request By Person"
	cell.Merge(3, 0)
	row = sheet3.AddRow()
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)

	for k, v := range countLeaveByPerson {
		row = sheet3.AddRow()
		cell = row.AddCell()
		cell.Value = k
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Leave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("ELeave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Remote"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Overtime"))
		cell.SetStyle(&style)
	}

	countLeaveByProject := datas.Get("LeaveByProject").(map[string]tk.M)

	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = reportTitle
	cell.Merge(3, 0)
	cell.SetStyle(&styleTitle)

	row = sheet4.AddRow()
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Total Request By Project"
	cell.Merge(3, 0)
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Project Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)

	for k, v := range countLeaveByProject {
		row = sheet4.AddRow()
		cell = row.AddCell()
		cell.Value = k
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Leave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("ELeave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Remote"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Overtime"))
		cell.SetStyle(&style)
	}

	countNameByProject := datas.Get("NameByProject").(tk.M)

	row = sheet4.AddRow()
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Detail Leave"
	cell.Merge(1, 0)
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Project Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)

	leaveList := countNameByProject.Get("Leave").([]tk.M)

	for _, v := range leaveList {
		proj := v.GetString("Project")
		empName := v.Get("EmpName").([]string)

		for _, y := range empName {
			row = sheet4.AddRow()
			cell = row.AddCell()
			cell.Value = proj
			cell.SetStyle(&style)
			cell = row.AddCell()
			cell.Value = y
			cell.SetStyle(&style)
		}
	}

	row = sheet4.AddRow()
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Detail E. Leave"
	cell.Merge(1, 0)
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Project Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)

	eleaveList := countNameByProject.Get("ELeave").([]tk.M)

	for _, v := range eleaveList {
		proj := v.GetString("Project")
		empName := v.Get("EmpName").([]string)

		for _, y := range empName {
			row = sheet4.AddRow()
			cell = row.AddCell()
			cell.Value = proj
			cell.SetStyle(&style)
			cell = row.AddCell()
			cell.Value = y
			cell.SetStyle(&style)
		}
	}

	row = sheet4.AddRow()
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Detail Remote"
	cell.Merge(1, 0)
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Project Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)

	remoteList := countNameByProject.Get("Remote").([]tk.M)

	for _, v := range remoteList {
		proj := v.GetString("Project")
		empName := v.Get("EmpName").([]string)

		for _, y := range empName {
			row = sheet4.AddRow()
			cell = row.AddCell()
			cell.Value = proj
			cell.SetStyle(&style)
			cell = row.AddCell()
			cell.Value = y
			cell.SetStyle(&style)
		}
	}

	row = sheet4.AddRow()
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Detail Overtime"
	cell.Merge(1, 0)
	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "Project Name"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Employee Name"
	cell.SetStyle(&styleHeader)

	overtimeList := countNameByProject.Get("Overtime").([]tk.M)

	for _, v := range overtimeList {
		proj := v.GetString("Project")
		empName := v.Get("EmpName").([]string)

		for _, y := range empName {
			row = sheet4.AddRow()
			cell = row.AddCell()
			cell.Value = proj
			cell.SetStyle(&style)
			cell = row.AddCell()
			cell.Value = y
			cell.SetStyle(&style)
		}
	}

	countLeaveByLocation := datas.Get("LeaveByLocation").(map[string]tk.M)

	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = reportTitle
	cell.Merge(3, 0)
	cell.SetStyle(&styleTitle)

	row = sheet5.AddRow()
	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = "Total Request By Location"
	cell.Merge(3, 0)
	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = "Location"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)

	for k, v := range countLeaveByLocation {
		row = sheet5.AddRow()
		cell = row.AddCell()
		cell.Value = k
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Leave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("ELeave"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Remote"))
		cell.SetStyle(&style)
		cell = row.AddCell()
		cell.SetInt(v.GetInt("Overtime"))
		cell.SetStyle(&style)
	}

	topFive := datas.Get("TopFive").(tk.M)
	topFiveLeave := topFive.Get("TopLeave").([]tk.M)
	topFiveELeave := topFive.Get("TopELeave").([]tk.M)
	topFiveRemote := topFive.Get("TopRemote").([]tk.M)
	topFiveOvertime := topFive.Get("TopOvertime").([]tk.M)

	row = sheet5.AddRow()
	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = "Top Five User Leave"
	cell.Merge(2, 0)
	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = "Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "E. Leave"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Remote"
	cell.SetStyle(&styleHeader)
	cell = row.AddCell()
	cell.Value = "Overtime"
	cell.SetStyle(&styleHeader)

	for k, v := range topFiveLeave {
		row = sheet5.AddRow()
		cell = row.AddCell()
		cell.Value = v.GetString("Name")
		cell.SetStyle(&style)

		cell = row.AddCell()
		cell.Value = topFiveELeave[k].GetString("Name")
		cell.SetStyle(&style)

		cell = row.AddCell()
		cell.Value = topFiveRemote[k].GetString("Name")
		cell.SetStyle(&style)

		cell = row.AddCell()
		cell.Value = topFiveOvertime[k].GetString("Name")
		cell.SetStyle(&style)
	}

	err = file.Save(fileLocation)
	if err != nil {
		return "", err
	}

	return "/static/doc/" + fileName, nil
}

func (s *DashboardService) GetOvertimeData(dateByMonth, location string) ([]tk.M, error) {
	pipe := []tk.M{}
	match := tk.M{}

	or := []tk.M{
		tk.M{}.Set("membersovertime.result", "Confirmed"),
		tk.M{}.Set("membersovertime.result", "Declined"),
	}

	match.Set("isexpired", false)
	match.Set("daylist.date", tk.M{}.Set("$regex", ".*"+dateByMonth+".*"))
	if location != "" && location != "Global" {
		match.Set("location", location)
	}
	match.Set("$or", or)

	pipe = append(pipe, tk.M{"$unwind": "$membersovertime"})
	pipe = append(pipe, tk.M{"$unwind": "$daylist"})
	pipe = append(pipe, tk.M{}.Set("$match", match))

	overtime, err := new(repositories.OvertimeDboxRepo).GetNewOvertimeByPipe(pipe)
	if err != nil {
		return nil, err
	}

	return overtime, nil
}
