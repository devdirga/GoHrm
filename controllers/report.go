package controllers

import (
	. "creativelab/ecleave-dev/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type ReportController struct {
	*BaseController
}
type PayloadRemote struct {
	Yearly   bool
	Month    int
	Year     int
	Day      int
	Project  string
	Name     []string
	Location string
	Date     string
}

type PayloadLeave struct {
	Project  string
	Name     []string
	Location string
	DateFrom string
	DateTo   string
}

type PayloadOvertime struct {
	Yearly   bool
	Month    int
	Day      string
	Year     int
	Project  string
	Name     []string
	Location string
}

type PayloadOvertimeDetail struct {
	Id     string
	Yearly bool
	Month  int
	Year   int
	Type   string
}

type PayloadAnnual struct {
	Yearly   bool
	Month    int
	Year     int
	Day      string
	Project  string
	Name     []string
	Location string
	Date     string
}

func (c *ReportController) Default(k *knot.WebContext) interface{} {
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
		"_modal.html",
		"_loader.html",
		"/report/leavereport.html",
		"/report/remotereport.html",
		"/report/overtimereport.html",
	}
	return DataAccess
}

func (c *ReportController) AnnualGridreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadAnnual{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := map[string]*ReportAnnualGrid{}

	start, _ := time.Parse("02-01-2006", "01-01-2019")
	end, _ := time.Parse("02-01-2006", "31-12-2019")

	// leave
	{
		pipe := []tk.M{}
		match := tk.M{}
		yearFilter := tk.M{
			"$and": []tk.M{
				tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
				tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
			},
		}
		if len(p.Name) > 0 {
			match.Set("fullname", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("detailLeave.project", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("detailLeave.location", p.Location)
		}
		if p.Day != "32" {
			match.Set("detailLeave.dateleave", p.Day)
		}
		if p.Yearly {
			match.Set("yearval", p.Year)
			start, _ = time.Parse("02-01-2006", "01-01-"+strconv.Itoa(p.Year))
			end, _ = time.Parse("02-01-2006", "31-12-"+strconv.Itoa(p.Year))
		} else {
			match.Set("monthval", p.Month)
			match.Set("yearval", p.Year)
			start = time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.Local)
			nextMonth := start.AddDate(0, 1, 0)
			end = nextMonth.AddDate(0, 0, -1)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
					tk.M{"$eq": []interface{}{"$detailLeave.monthval", p.Month}},
					tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
				},
			}
		}
		pipe = append(pipe, tk.M{
			"$lookup": tk.M{
				"from":         "requestLeaveByDate",
				"localField":   "empid",
				"foreignField": "empid",
				"as":           "detailLeave",
			},
		})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailLeave", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname": "$fullname",
				"empid":    "$empid",
				"detailLeave": tk.M{
					"$cond": []interface{}{
						yearFilter, "$detailLeave", tk.M{},
					},
				},
				"yearval": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
							},
						}, "$detailLeave.yearval", p.Year,
					},
				},
				"monthval": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$detailLeave.monthval", p.Month}},
							},
						}, "$detailLeave.monthval", p.Month,
					},
				},
			},
		})
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"Name":         "$fullname",
				"EmpId":        "$empid",
				"Isemergency":  "$detailLeave.isemergency",
				"Stsbymanager": "$detailLeave.stsbymanager",
			},
			"Count": tk.M{"$sum": 1},
		}})

		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		leavedata := []tk.M{}
		e = csr.Fetch(&leavedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		activeDays, _ := c.CalcBussinesDays(&start, &end)

		for _, each := range leavedata {

			d := each.Get("_id").(tk.M)

			name := d.GetString("Name")
			empid := d.GetString("EmpId")

			if datas[empid] == nil {
				datas[empid] = new(ReportAnnualGrid)
			}
			datas[empid].Name = strings.ToUpper(name)
			datas[empid].EmpId = empid
			if d.GetString("Stsbymanager") == "Approved" {
				if d.Get("Isemergency").(bool) {
					datas[empid].TotalEleave = int(each.GetFloat64("Count"))
				} else {
					datas[empid].TotalLeave = int(each.GetFloat64("Count"))
				}
			} else if d.GetString("Stsbymanager") == "Declined" {
				datas[empid].TotalDecline = datas[empid].TotalDecline + int(each.GetFloat64("Count"))
			}
			datas[empid].ActiveDays = activeDays - (datas[empid].TotalEleave + datas[empid].TotalLeave)
		}
	}
	//remote
	{
		pipe := []tk.M{}
		match := tk.M{}
		monthStr := strconv.Itoa(p.Month)
		match.Set("detailRemote.projects.isleadersend", true)
		match.Set("detailRemote.isdelete", false)

		if p.Day != "32" {
			match.Set("detailRemote.dateleave", p.Day)
		}

		if p.Month < 10 {
			monthStr = "0" + monthStr
		}
		yearStr := strconv.Itoa(p.Year)
		yearFilter := tk.M{
			"$and": []tk.M{
				tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
			},
		}

		if len(p.Name) > 0 {
			match.Set("fullname", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("detailRemote.projects.projectname", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("detailRemote.location", p.Location)
		}
		if p.Yearly {
			match.Set("yearRemote", yearStr)
		} else {
			match.Set("monthRemote", monthStr)
			match.Set("yearRemote", yearStr)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
					tk.M{"$eq": []interface{}{"$monthRemote", monthStr}},
				},
			}
		}
		pipe = append(pipe, tk.M{
			"$lookup": tk.M{
				"from":         "remote",
				"localField":   "_id",
				"foreignField": "userid",
				"as":           "detailRemote",
			},
		})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote.projects", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname":     "$fullname",
				"empid":        "$empid",
				"detailRemote": "$detailRemote",
				"monthRemote":  tk.M{"$ifNull": []interface{}{"$detailRemote.dateleave", "2016-" + monthStr + "01"}},
				"yearRemote":   tk.M{"$ifNull": []interface{}{"$detailRemote.dateleave", yearStr + "-01-01"}},
			},
		})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname":     "$fullname",
				"empid":        "$empid",
				"detailRemote": "$detailRemote",
				"monthRemote":  tk.M{"$substr": []interface{}{"$monthRemote", 5, 2}},
				"yearRemote":   tk.M{"$substr": []interface{}{"$yearRemote", 0, 4}},
			},
		})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname": "$fullname",
				"empid":    "$empid",
				"detailRemote": tk.M{
					"$cond": []interface{}{
						yearFilter, "$detailRemote", tk.M{},
					},
				},
				"monthRemote": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$monthRemote", monthStr}},
							},
						}, "$monthRemote", monthStr,
					},
				},
				"yearRemote": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
							},
						}, "$yearRemote", yearStr,
					},
				},
			},
		})
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"Name":           "$fullname",
				"EmpId":          "$empid",
				"SendLeader":     "$detailRemote.projects.isleadersend",
				"SendManager":    "$detailRemote.projects.ismanagersend",
				"ApproveLeader":  "$detailRemote.projects.isapprovalleader",
				"ApproveManager": "$detailRemote.projects.isapprovalmanager",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		remotedata := []tk.M{}
		e = csr.Fetch(&remotedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, each := range remotedata {
			d := each.Get("_id").(tk.M)
			name := d.GetString("Name")
			empid := d.GetString("EmpId")
			if datas[empid] == nil {
				datas[empid] = new(ReportAnnualGrid)
			}
			datas[empid].Name = strings.ToUpper(name)
			datas[empid].EmpId = empid
			if d.Get("ApproveManager") != nil {
				if d.Get("ApproveManager").(bool) {
					datas[empid].TotalRemote = int(each.GetFloat64("Count"))
				} else {
					if d.Get("SendLeader").(bool) && !d.Get("ApproveLeader").(bool) {
						datas[empid].TotalDecline = datas[empid].TotalDecline + int(each.GetFloat64("Count"))
					} else if d.Get("ApproveLeader").(bool) && d.Get("SendManager").(bool) && !d.Get("ApproveManager").(bool) {
						datas[empid].TotalDecline = datas[empid].TotalDecline + int(each.GetFloat64("Count"))
					}
				}
			}

			if datas[empid].ActiveDays == 0 {
				activeDays, _ := c.CalcBussinesDays(&start, &end)
				datas[empid].ActiveDays = activeDays
			}
		}
	}

	//Overtime
	{
		pipe := []tk.M{}
		var matchdate string

		if p.Day == "32" {
			if p.Yearly {
				matchdate = strconv.Itoa(p.Year)
			} else {
				matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
			}
		} else {
			matchdate = p.Day
		}
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
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
		for _, each := range overtimedata {
			if datas[each.MembersOvertime.IdEmp] == nil {
				datas[each.MembersOvertime.IdEmp] = new(ReportAnnualGrid)
				datas[each.MembersOvertime.IdEmp].Name = strings.ToUpper(each.MembersOvertime.Name)
				datas[each.MembersOvertime.IdEmp].EmpId = each.MembersOvertime.IdEmp
			}
			if each.MembersOvertime.Result == "Confirmed" {
				datas[each.MembersOvertime.IdEmp].TotalOvertime = datas[each.MembersOvertime.IdEmp].TotalOvertime + 1
			} else if each.MembersOvertime.Result == "Declined" {
				datas[each.MembersOvertime.IdEmp].TotalDecline = datas[each.MembersOvertime.IdEmp].TotalDecline + 1
			}
		}

	}
	//data
	results := []*ReportAnnualGrid{}
	for _, data := range datas {
		results = append(results, data)
	}

	return c.SetResultInfo(false, "Success", results)
}

func (c *ReportController) LeaveGridreport(k *knot.WebContext) interface{} {
	//sm
	k.Config.OutputType = knot.OutputJson
	p := PayloadRemote{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := map[string]*ReportLeaveGrid{}

	start, _ := time.Parse("02-01-2006", "01-01-2019")
	end, _ := time.Parse("02-01-2006", "31-12-2019")

	// leave
	{
		pipe := []tk.M{}
		match := tk.M{}
		yearFilter := tk.M{
			"$and": []tk.M{
				tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
				tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
			},
		}
		if len(p.Name) > 0 {
			match.Set("fullname", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("detailLeave.project", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("detailLeave.location", p.Location)
		}
		if p.Yearly {
			match.Set("yearval", p.Year)
			start, _ = time.Parse("02-01-2006", "01-01-"+strconv.Itoa(p.Year))
			end, _ = time.Parse("02-01-2006", "31-12-"+strconv.Itoa(p.Year))
		} else if p.Day == 32 {
			match.Set("monthval", p.Month)
			match.Set("yearval", p.Year)
			start = time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.Local)
			nextMonth := start.AddDate(0, 1, 0)
			end = nextMonth.AddDate(0, 0, -1)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
					tk.M{"$eq": []interface{}{"$detailLeave.monthval", p.Month}},
					tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
				},
			}
		} else {
			match.Set("monthval", p.Month)
			match.Set("yearval", p.Year)
			match.Set("dayval", p.Day)
			match.Set("detailLeave.isdelete", false)
			start = time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.Local)
			nextMonth := start.AddDate(0, 1, 0)
			end = nextMonth.AddDate(0, 0, -1)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
					tk.M{"$eq": []interface{}{"$detailLeave.monthval", p.Month}},
					tk.M{"$eq": []interface{}{"$detailLeave.dayval", p.Day}},
					tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
				},
			}
		}
		pipe = append(pipe, tk.M{
			"$lookup": tk.M{
				"from":         "requestLeaveByDate",
				"localField":   "empid",
				"foreignField": "empid",
				"as":           "detailLeave",
			},
		})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailLeave", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname": "$fullname",
				"empid":    "$empid",
				"detailLeave": tk.M{
					"$cond": []interface{}{
						yearFilter, "$detailLeave", tk.M{},
					},
				},
				"yearval": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$detailLeave.yearval", p.Year}},
							},
						}, "$detailLeave.yearval", p.Year,
					},
				},
				"monthval": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$detailLeave.monthval", p.Month}},
							},
						}, "$detailLeave.monthval", p.Month,
					},
				},
				"dayval": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$detailLeave.dayval", p.Day}},
							},
						}, "$detailLeave.dayval", p.Day,
					},
				},
			},
		})
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"Name":         "$fullname",
				"EmpId":        "$empid",
				"Isemergency":  "$detailLeave.isemergency",
				"Stsbymanager": "$detailLeave.stsbymanager",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		leavedata := []tk.M{}
		e = csr.Fetch(&leavedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		activeDays, _ := c.CalcBussinesDays(&start, &end)

		for _, each := range leavedata {
			d := each.Get("_id").(tk.M)
			name := d.GetString("Name")
			empid := d.GetString("EmpId")
			if datas[empid] == nil {
				datas[empid] = new(ReportLeaveGrid)
			}
			datas[empid].Name = strings.ToUpper(name)
			datas[empid].EmpId = empid
			if d.GetString("Stsbymanager") == "Approved" {
				if d.Get("Isemergency").(bool) {
					datas[empid].TotalEleave = int(each.GetFloat64("Count"))
				} else {
					datas[empid].TotalLeave = int(each.GetFloat64("Count"))
				}
			} else if d.GetString("Stsbymanager") == "Declined" {
				datas[empid].TotalDecline = datas[empid].TotalDecline + int(each.GetFloat64("Count"))
			}
			datas[empid].Summary = (datas[empid].TotalLeave + datas[empid].TotalEleave) - datas[empid].TotalDecline
			if datas[empid].Summary < 0 {
				datas[empid].Summary = 0
			}
			datas[empid].ActiveDays = activeDays - (datas[empid].TotalEleave + datas[empid].TotalLeave)
		}
	}
	//data
	results := []*ReportLeaveGrid{}
	for _, data := range datas {
		results = append(results, data)
	}
	return c.SetResultInfo(false, "Success", results)
}

func (c *ReportController) LeaveGridBetweenreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadLeave{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := map[string]*ReportLeaveBetweenGrid{}

	start := p.DateFrom
	end := p.DateTo

	startDate, _ := time.Parse("2006-01-02", start)
	endDate, _ := time.Parse("2006-01-02", end)

	runesStart := []rune(start)
	runesEnd := []rune(end)

	startSubstring := string(runesStart[0:7])
	endSubstring := string(runesEnd[0:7])

	// leave
	{
		pipe := []tk.M{}
		match := tk.M{}
		match.Set("$or", []tk.M{
			tk.M{}.Set("detailLeave.dateleave", tk.M{}.Set("$regex", ".*"+startSubstring+".*")),
			tk.M{}.Set("detailLeave.dateleave", tk.M{}.Set("$regex", ".*"+endSubstring+".*")),
		})
		yearFilter := tk.M{
			"$and": []tk.M{
				tk.M{"$eq": []interface{}{"$detailLeave.isdelete", false}},
				tk.M{"$eq": []interface{}{"$detailLeave.isemergency", false}},
			},
		}
		if len(p.Name) > 0 {
			match.Set("fullname", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("detailLeave.project", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("detailLeave.location", p.Location)
		}
		pipe = append(pipe, tk.M{
			"$lookup": tk.M{
				"from":         "requestLeaveByDate",
				"localField":   "empid",
				"foreignField": "empid",
				"as":           "detailLeave",
			},
		})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailLeave", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname":    "$fullname",
				"empid":       "$empid",
				"phonenumber": "$phonenumber",
				"detailLeave": tk.M{
					"$cond": []interface{}{
						yearFilter, "$detailLeave", tk.M{},
					},
				},
			},
		})
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"Name":         "$fullname",
				"EmpId":        "$empid",
				"PhoneNumber":  "$phonenumber",
				"Project":      "$detailLeave.project",
				"DateLeave":    "$detailLeave.dateleave",
				"Stsbymanager": "$detailLeave.stsbymanager",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		leavedata := []tk.M{}
		e = csr.Fetch(&leavedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}

		dateLeaveList := []DateLeaveBetween{}
		for d := startDate; d.Before(endDate.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
			dateLeave := DateLeaveBetween{}
			dateLeave.Date = d.Format("2006-01-02")
			dateLeave.Leave = false

			holiday := NationalHolidayController(*c)
			holiData, _ := holiday.CheckDateNH(k, "Indonesia", d)
			if len(holiData) > 0 {
				dateLeave.Holiday = true
			}
			if d.Weekday() == 0 || d.Weekday() == 6 {
				dateLeave.Holiday = true
			}

			dateLeaveList = append(dateLeaveList, dateLeave)
		}

		for _, each := range leavedata {
			d := each.Get("_id").(tk.M)
			name := d.GetString("Name")
			empid := d.GetString("EmpId")
			phonenumber := d.GetString("PhoneNumber")
			project := d.GetString("Project")
			dateleave := d.GetString("DateLeave")
			if datas[empid] == nil {
				datas[empid] = new(ReportLeaveBetweenGrid)
				dateLeaveList2 := []DateLeaveBetween{}
				for _, each2 := range dateLeaveList {
					dateLeaveList2 = append(dateLeaveList2, each2)
				}
				datas[empid].DateLeave = dateLeaveList2
			}
			datas[empid].Name = strings.ToUpper(name)
			datas[empid].EmpId = empid
			datas[empid].PhoneNumber = phonenumber
			datas[empid].Project = project
			index := 0
			for _, item := range datas[empid].DateLeave {
				if item.Date == dateleave {
					if d.GetString("Stsbymanager") == "Approved" {
						datas[empid].DateLeave[index].Leave = true
						datas[empid].Isvisible = true
					}
				}
				index++
			}
		}
	}
	//data
	results := []*ReportLeaveBetweenGrid{}
	for _, data := range datas {
		if data.Isvisible {
			results = append(results, data)
		}
	}
	return c.SetResultInfo(false, "Success", results)
}

func (c *ReportController) RemoteGridreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadRemote{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := map[string]*ReportRemoteGrid{}

	start, _ := time.Parse("02-01-2006", "01-01-2019")
	end, _ := time.Parse("02-01-2006", "31-12-2019")

	//remote
	{
		pipe := []tk.M{}
		match := tk.M{}
		monthStr := strconv.Itoa(p.Month)
		if p.Month < 10 {
			monthStr = "0" + monthStr
		}
		yearStr := strconv.Itoa(p.Year)
		yearFilter := tk.M{
			"$and": []tk.M{
				tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
				tk.M{"$eq": []interface{}{"$detailRemote.isdelete", false}},
			},
		}

		if len(p.Name) > 0 {
			match.Set("fullname", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("detailRemote.projects.projectname", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("detailRemote.location", p.Location)
		}
		if p.Yearly {
			match.Set("yearRemote", yearStr)
			start, _ = time.Parse("02-01-2006", "01-01-"+yearStr)
			end, _ = time.Parse("02-01-2006", "31-12-"+yearStr)
		} else if p.Date == "32" {
			match.Set("monthRemote", monthStr)
			match.Set("yearRemote", yearStr)
			start = time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.Local)
			nextMonth := start.AddDate(0, 1, 0)
			end = nextMonth.AddDate(0, 0, -1)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
					tk.M{"$eq": []interface{}{"$monthRemote", monthStr}},
					tk.M{"$eq": []interface{}{"$detailRemote.isdelete", false}},
				},
			}
		} else {
			tk.Println("innnnnn.....", p.Date)
			match.Set("monthRemote", monthStr)
			match.Set("yearRemote", yearStr)
			match.Set("dateLeave", p.Date)
			match.Set("detailRemote.isdelete", false).Set("detailRemote.isexpired", false).Set("detailRemote.projects.isapprovalmanager", true)
			start = time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.Local)
			nextMonth := start.AddDate(0, 1, 0)
			end = nextMonth.AddDate(0, 0, -1)
			yearFilter = tk.M{
				"$and": []tk.M{
					tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
					tk.M{"$eq": []interface{}{"$monthRemote", monthStr}},
					tk.M{"$eq": []interface{}{"$dateLeave", p.Date}},
					tk.M{"$eq": []interface{}{"$detailRemote.isdelete", false}},
				},
			}
		}
		pipe = append(pipe, tk.M{
			"$lookup": tk.M{
				"from":         "remote",
				"localField":   "_id",
				"foreignField": "userid",
				"as":           "detailRemote",
			},
		})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote.projects", "preserveNullAndEmptyArrays": true}})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"fullname":     "$fullname",
				"empid":        "$empid",
				"_id":          "$_id",
				"detailRemote": "$detailRemote",
				"monthRemote":  tk.M{"$ifNull": []interface{}{"$detailRemote.dateleave", "2016-" + monthStr + "01"}},
				"yearRemote":   tk.M{"$ifNull": []interface{}{"$detailRemote.dateleave", yearStr + "-01-01"}},
				"dateLeave":    "$detailRemote.dateleave",
			},
		})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"dateLeave":    1,
				"fullname":     "$fullname",
				"empid":        "$empid",
				"_id":          "$_id",
				"detailRemote": "$detailRemote",
				"monthRemote":  tk.M{"$substr": []interface{}{"$monthRemote", 5, 2}},
				"yearRemote":   tk.M{"$substr": []interface{}{"$yearRemote", 0, 4}},
			},
		})
		pipe = append(pipe, tk.M{
			"$project": tk.M{
				"dateLeave": 1,
				"fullname":  "$fullname",
				"empid":     "$empid",
				"_id":       "$_id",
				"detailRemote": tk.M{
					"$cond": []interface{}{
						yearFilter, "$detailRemote", tk.M{},
					},
				},
				"monthRemote": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$monthRemote", monthStr}},
							},
						}, "$monthRemote", monthStr,
					},
				},
				"yearRemote": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$yearRemote", yearStr}},
							},
						}, "$yearRemote", yearStr,
					},
				},
			},
		})
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"Name":           "$fullname",
				"EmpId":          "$empid",
				"UserId":         "$_id",
				"Type":           "$detailRemote.type",
				"ApproveManager": "$detailRemote.projects.isapprovalmanager",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		remotedata := []tk.M{}
		e = csr.Fetch(&remotedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		activeDays, _ := c.CalcBussinesDays(&start, &end)

		for _, each := range remotedata {
			d := each.Get("_id").(tk.M)
			name := d.GetString("Name")
			empid := d.GetString("EmpId")
			userid := d.GetString("UserId")
			if datas[empid] == nil {
				datas[empid] = new(ReportRemoteGrid)
			}
			datas[empid].Name = strings.ToUpper(name)
			datas[empid].EmpId = empid
			datas[empid].UserId = userid
			if d.Get("ApproveManager") != nil {
				if d.Get("ApproveManager").(bool) {
					if d.Get("Type") == "1" {
						datas[empid].ConditionalRemote = int(each.GetFloat64("Count"))
					} else {
						datas[empid].FullRemote = int(each.GetFloat64("Count"))
					}
				}
			}
			datas[empid].TotalRemote = datas[empid].ConditionalRemote + datas[empid].FullRemote
			datas[empid].RemotePerformance = (datas[empid].TotalRemote * 100) / activeDays
		}
	}
	//data
	results := []*ReportRemoteGrid{}
	for _, data := range datas {
		results = append(results, data)
	}
	return c.SetResultInfo(false, "Success", results)
}

func (c *ReportController) AnnualChartreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Yearly   bool
		Month    int
		Year     int
		Day      string
		Project  string
		Name     []string
		Location string
		Date     string
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	datas := []*ReportChart{}

	datas = append(datas, &ReportChart{
		Name: "Leave",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Emergency Leave",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Remote",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Overtime",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})

	// leave
	{
		pipe := []tk.M{}
		match := tk.M{}
		match.Set("stsbymanager", "Approved")
		match.Set("yearval", p.Year)
		if len(p.Name) > 0 {
			match.Set("name", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("project", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("location", p.Location)
		}
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"MonthVal":    "$monthval",
				"Isemergency": "$isemergency",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewAprovalRequestLeaveModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		leavedata := []tk.M{}
		e = csr.Fetch(&leavedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		//tk.Println(leavedata)

		for _, each := range leavedata {
			for index := 0; index < 12; index++ {
				d := each.Get("_id").(tk.M)

				if d.Get("Isemergency").(bool) {
					if (index + 1) == d.Get("MonthVal") {
						datas[1].Data[index] = int(each.GetFloat64("Count"))
					}
				} else {
					if (index + 1) == d.Get("MonthVal") {
						datas[0].Data[index] = int(each.GetFloat64("Count"))
					}
				}
			}
		}
	}

	//remote
	{
		pipe := []tk.M{}
		match := tk.M{}
		pipe = append(pipe, tk.M{"$unwind": "$projects"})
		match.Set("projects.isleadersend", tk.M{}.Set("$eq", true))
		match.Set("projects.ismanagersend", tk.M{}.Set("$eq", true))
		match.Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true))
		YearStr := strconv.Itoa(p.Year)
		match.Set("dateleave", tk.M{}.Set("$regex", ".*"+YearStr+".*"))
		if len(p.Name) > 0 {
			match.Set("name", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("projects.projectname", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("location", p.Location)
		}
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"MonthLeave": tk.M{"$substr": []interface{}{"$dateleave", 5, 2}},
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewRemoteModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		remotedata := []tk.M{}
		e = csr.Fetch(&remotedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, each := range remotedata {
			for index := 0; index < 12; index++ {
				d := each.Get("_id").(tk.M)

				if (index + 1) == d.GetInt("MonthLeave") {
					datas[2].Data[index] = int(each.GetFloat64("Count"))
				}
			}
		}
	}

	//overtime
	{
		pipe := []tk.M{}
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.result", "Confirmed")))
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}})
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
		for _, each := range overtimedata {
			for index := 0; index < 12; index++ {
				m := strings.Split(each.DayList.Date, "-")
				mInt, _ := strconv.Atoi(m[1])
				if (index + 1) == mInt {
					datas[3].Data[index] = datas[3].Data[index] + 1
				}
			}
		}

	}

	return c.SetResultInfo(false, "Success", datas)
}

func (c *ReportController) LeaveChartreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadRemote{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	datas := []*ReportChart{}

	datas = append(datas, &ReportChart{
		Name: "Leave",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Emergency Leave",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})

	// leave
	{
		pipe := []tk.M{}
		match := tk.M{}
		match.Set("stsbymanager", "Approved")
		match.Set("yearval", p.Year)
		if len(p.Name) > 0 {
			match.Set("name", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("project", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("location", p.Location)
		}
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"MonthVal":    "$monthval",
				"Isemergency": "$isemergency",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewAprovalRequestLeaveModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		leavedata := []tk.M{}
		e = csr.Fetch(&leavedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}

		for _, each := range leavedata {
			for index := 0; index < 12; index++ {
				d := each.Get("_id").(tk.M)

				if d.Get("Isemergency").(bool) {
					if (index + 1) == d.Get("MonthVal") {
						datas[1].Data[index] = int(each.GetFloat64("Count"))
					}
				} else {
					if (index + 1) == d.Get("MonthVal") {
						datas[0].Data[index] = int(each.GetFloat64("Count"))
					}
				}
			}
		}
	}

	return c.SetResultInfo(false, "Success", datas)
}

func (c *ReportController) RemoteChartreport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadRemote{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	datas := []*ReportChart{}
	datas = append(datas, &ReportChart{
		Name: "Full Remote",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Conditional Remote",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})

	//remote
	{
		pipe := []tk.M{}
		match := tk.M{}
		pipe = append(pipe, tk.M{"$unwind": "$projects"})
		match.Set("projects.isleadersend", tk.M{}.Set("$eq", true))
		match.Set("projects.ismanagersend", tk.M{}.Set("$eq", true))
		match.Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true))
		YearStr := strconv.Itoa(p.Year)
		match.Set("dateleave", tk.M{}.Set("$regex", ".*"+YearStr+".*"))
		if len(p.Name) > 0 {
			match.Set("name", tk.M{}.Set("$in", p.Name))
		}
		if p.Project != "" {
			match.Set("projects.projectname", p.Project)
		}
		if p.Location != "" && p.Location != "Global" {
			match.Set("location", p.Location)
		}
		pipe = append(pipe, tk.M{}.Set("$match", match))
		pipe = append(pipe, tk.M{"$group": tk.M{
			"_id": tk.M{
				"MonthLeave": tk.M{"$substr": []interface{}{"$dateleave", 5, 2}},
				"Type":       "$type",
			},
			"Count": tk.M{"$sum": 1},
		}})
		csr, e := c.Ctx.Connection.NewQuery().Select().From(NewRemoteModel().TableName()).Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		remotedata := []tk.M{}
		e = csr.Fetch(&remotedata, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, each := range remotedata {
			for index := 0; index < 12; index++ {
				d := each.Get("_id").(tk.M)

				if d.Get("Type") == "1" {
					if (index + 1) == d.GetInt("MonthLeave") {
						datas[1].Data[index] = int(each.GetFloat64("Count"))
					}
				} else {
					if (index + 1) == d.GetInt("MonthLeave") {
						datas[0].Data[index] = int(each.GetFloat64("Count"))
					}
				}
			}
		}
	}

	return c.SetResultInfo(false, "Success", datas)
}

func (c *ReportController) CalcBussinesDays(from, to *time.Time) (int, error) {
	totalDays := float32(to.Sub(*from) / (24 * time.Hour))
	weekDays := float32(from.Weekday()) - float32(to.Weekday())
	businessDays := int(1 + (totalDays*5-weekDays*2)/7)
	if to.Weekday() == time.Saturday {
		businessDays--
	}
	if from.Weekday() == time.Sunday {
		businessDays--
	}

	//get national holidays
	nationalHolidays := 0
	andfilter := []*db.Filter{}
	andfilter = append(andfilter, db.Gte("date", from))
	andfilter = append(andfilter, db.Lt("date", to))

	filter := db.And(andfilter...)

	csr, e := c.Ctx.Connection.NewQuery().Select().From("NationalHolidays").Where(filter).Cursor(nil)
	if e != nil {
		nationalHolidays = 0
	}
	results := []NationalHolidaysModel{}
	e = csr.Fetch(&results, 0, false)
	if e != nil {
		nationalHolidays = 0
	}
	nationalHolidays = len(results)

	businessDays = businessDays - nationalHolidays

	return businessDays, nil
}

func (c *ReportController) DataDetail(k *knot.WebContext) interface{} {
	//sm
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Id       string
		Type     string
		Empid    string
		ArrEmpid []string
		Name     string
		Yearly   bool
		Month    int
		Year     int
		Day      int
		Date     string
	}{}
	if e := k.GetPayload(&p); e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	tk.Println("payload Datadetail....")
	tk.Println(p)

	//get userid from empid
	tk.Println("---------------  empid ", p.Empid)
	var userid string
	if p.Empid != "" {
		pipe := []tk.M{}
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("empid", p.Empid)))
		csrm, e := c.Ctx.Connection.NewQuery().Select().From("SysUsers").Command("pipe", pipe).Cursor(nil)
		defer csrm.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dUser := []SysUserModel{}
		if e = csrm.Fetch(&dUser, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dUser {
			userid = dec.Id
		}
	}

	if p.Type == "leave" {
		pipeLeave := []tk.M{}
		matchLeave := tk.M{}
		matchLeave.Set("empid", p.Empid).Set("yearval", p.Year).Set("stsbymanager", "Approved").Set("isemergency", false).Set("isdelete", false)
		if p.Yearly == false && p.Day == 32 {
			matchLeave.Set("monthval", p.Month)
		} else if p.Yearly == false && p.Day != 32 {
			matchLeave.Set("monthval", p.Month)
			matchLeave.Set("dayval", p.Day)
		}
		pipeLeave = append(pipeLeave, tk.M{}.Set("$match", matchLeave))
		csrLeave, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", pipeLeave).Cursor(nil)
		defer csrLeave.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dLeave := []RequestLeaveByDateModel{}
		leaveData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrLeave.Fetch(&dLeave, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dLeave {
			leaveData = append(leaveData, Leave{
				EmpId:       dec.EmpId,
				Name:        dec.Name,
				Project:     dec.Project,
				RequestType: "Leave",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return leaveData

	} else if p.Type == "eleave" {
		pipeEleave := []tk.M{}
		matchEleave := tk.M{}
		matchEleave.Set("empid", p.Empid).Set("yearval", p.Year).Set("stsbymanager", "Approved").Set("isemergency", true).Set("isdelete", false)
		// if p.Yearly == false {
		// 	matchEleave.Set("monthval", p.Month)
		// }
		if p.Yearly == false && p.Day == 32 {
			matchEleave.Set("monthval", p.Month)
		} else if p.Yearly == false && p.Day != 32 {
			matchEleave.Set("monthval", p.Month)
			matchEleave.Set("dayval", p.Day)
		}
		pipeEleave = append(pipeEleave, tk.M{}.Set("$match", matchEleave))
		csrEleave, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", pipeEleave).Cursor(nil)
		defer csrEleave.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dEleave := []RequestLeaveByDateModel{}
		eleaveData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrEleave.Fetch(&dEleave, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dEleave {
			eleaveData = append(eleaveData, Leave{
				EmpId:       dec.EmpId,
				Name:        dec.Name,
				Project:     dec.Project,
				RequestType: "Eleave",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return eleaveData

	} else if p.Type == "remote" {
		pipeRemote := []tk.M{}
		matchRemote := tk.M{}
		matchRemote.Set("userid", userid).Set("projects.isleadersend", tk.M{}.Set("$eq", true)).Set("projects.ismanagersend", tk.M{}.Set("$eq", true)).Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true)).Set("isdelete", false).Set("isexpired", false)
		if p.Yearly == true {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		} else if p.Date == "32" {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		} else {
			//tk.Println("iiii 32")
			matchRemote.Set("dateleave", p.Date)
		}
		pipeRemote = append(pipeRemote, tk.M{}.Set("$match", matchRemote))
		csrRemote, e := c.Ctx.Connection.NewQuery().Select().From("remote").Command("pipe", pipeRemote).Cursor(nil)
		defer csrRemote.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dRemote := []RemoteModel{}
		remoteData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrRemote.Fetch(&dRemote, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dRemote {
			remoteData = append(remoteData, Leave{
				EmpId:       "",
				Name:        dec.Name,
				Project:     []string{dec.Projects[0].ProjectName},
				RequestType: "Remote",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return remoteData

	} else if p.Type == "decline" {

		pipeLeave := []tk.M{}
		matchLeave := tk.M{}
		matchLeave.Set("empid", p.Empid).Set("yearval", p.Year).Set("$or", []tk.M{
			tk.M{}.Set("stsbyleader", "Declined"),
			tk.M{}.Set("stsbymanager", "Declined"),
		})
		if p.Yearly == false {
			matchLeave.Set("monthval", p.Month)
		}
		pipeLeave = append(pipeLeave, tk.M{}.Set("$match", matchLeave), tk.M{
			"$lookup": tk.M{
				"from":         "requestLeave",
				"localField":   "idrequest",
				"foreignField": "_id",
				"as":           "requestleave",
			},
		})
		csrLeave, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", pipeLeave).Cursor(nil)
		defer csrLeave.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		declineData := []struct {
			RequestType   string
			Date          string
			LeaveReason   string
			DeclineReason string
		}{}
		decLeave := []RequestLeaveByDateModel{}
		if e = csrLeave.Fetch(&decLeave, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		var typerequest, declinereason string
		for _, dec := range decLeave {
			if dec.IsEmergency {
				typerequest = "Eleave"
			} else {
				typerequest = "Leave"
			}
			declinereason = dec.RequestLeave[0].StatusProjectLeader[0].Reason
			if dec.RequestLeave[0].StatusManagerProject.Reason != "" {
				declinereason = dec.RequestLeave[0].StatusManagerProject.Reason
			}
			declineData = append(declineData, Decline{
				RequestType:   typerequest,
				Date:          dec.DateLeave,
				LeaveReason:   dec.Reason,
				DeclineReason: declinereason,
			})
		}

		pipeRemote := []tk.M{}
		matchRemote := tk.M{}
		matchRemote.Set("userid", userid)
		matchRemote.Set("$or", []tk.M{
			tk.M{"$and": []interface{}{tk.M{}.Set("projects.isleadersend", true), tk.M{}.Set("projects.isapprovalleader", false)}},
			tk.M{"$and": []interface{}{tk.M{}.Set("projects.ismanagersend", true), tk.M{}.Set("projects.isapprovalleader", true), tk.M{}.Set("projects.isapprovalmanager", false)}},
		})
		if p.Yearly == true {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		} else {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		}
		pipeRemote = append(pipeRemote, tk.M{}.Set("$match", matchRemote))
		csrRemote, e := c.Ctx.Connection.NewQuery().Select().From("remote").Command("pipe", pipeRemote).Cursor(nil)
		defer csrRemote.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		decRemote := []RemoteModel{}
		if e = csrRemote.Fetch(&decRemote, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range decRemote {
			remoteDecline := dec.Projects[0].NoteLeader
			if dec.Projects[0].NoteManager != "" {
				remoteDecline = dec.Projects[0].NoteManager
			}
			declineData = append(declineData, Decline{
				RequestType:   "Remote",
				Date:          dec.DateLeave,
				LeaveReason:   dec.Reason,
				DeclineReason: remoteDecline,
			})
		}

		pipe := []tk.M{}
		var matchdate string
		if p.Yearly {
			matchdate = strconv.Itoa(p.Year)
		} else {
			matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
		}
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.result", "Declined")))
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", userid)))
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
		for _, each := range overtimedata {
			declineData = append(declineData, Decline{
				RequestType:   "Overtime",
				Date:          each.DayList.Date,
				LeaveReason:   each.Reason,
				DeclineReason: each.ApprovalManager.Reason,
			})
		}

		return declineData
	} else if p.Type == "declineleave" {

		pipeLeave := []tk.M{}
		matchLeave := tk.M{}
		matchLeave.Set("empid", p.Empid).Set("yearval", p.Year).Set("$or", []tk.M{
			tk.M{}.Set("stsbyleader", "Declined"),
			tk.M{}.Set("stsbymanager", "Declined"),
		})
		// if p.Yearly == false {
		// 	matchLeave.Set("monthval", p.Month)
		// }
		if p.Yearly == false && p.Day == 32 {
			matchLeave.Set("monthval", p.Month)
		} else if p.Yearly == false && p.Day != 32 {
			matchLeave.Set("monthval", p.Month)
			matchLeave.Set("dayval", p.Day)
		}
		pipeLeave = append(pipeLeave, tk.M{}.Set("$match", matchLeave), tk.M{
			"$lookup": tk.M{
				"from":         "requestLeave",
				"localField":   "idrequest",
				"foreignField": "_id",
				"as":           "requestleave",
			},
		})
		csrLeave, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", pipeLeave).Cursor(nil)
		defer csrLeave.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		declineData := []struct {
			RequestType   string
			Date          string
			LeaveReason   string
			DeclineReason string
		}{}
		decLeave := []RequestLeaveByDateModel{}
		if e = csrLeave.Fetch(&decLeave, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		var typerequest, declinereason string
		for _, dec := range decLeave {
			if dec.IsEmergency {
				typerequest = "Eleave"
			} else {
				typerequest = "Leave"
			}
			declinereason = dec.RequestLeave[0].StatusProjectLeader[0].Reason
			if dec.RequestLeave[0].StatusManagerProject.Reason != "" {
				declinereason = dec.RequestLeave[0].StatusManagerProject.Reason
			}
			declineData = append(declineData, Decline{
				RequestType:   typerequest,
				Date:          dec.DateLeave,
				LeaveReason:   dec.Reason,
				DeclineReason: declinereason,
			})
		}

		return declineData

	} else if p.Type == "fullremote" {

		pipeRemote := []tk.M{}
		matchRemote := tk.M{}
		matchRemote.Set("userid", userid).Set("type", "2").Set("projects.isleadersend", tk.M{}.Set("$eq", true)).Set("projects.ismanagersend", tk.M{}.Set("$eq", true)).Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true)).Set("isdelete", false).Set("isexpired", false)
		// if p.Yearly == true {
		// 	matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		// } else {
		// 	matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		// }
		if p.Yearly == true {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		} else if p.Date == "32" {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		} else {
			matchRemote.Set("dateleave", p.Date)
		}
		pipeRemote = append(pipeRemote, tk.M{}.Set("$match", matchRemote))
		csrRemote, e := c.Ctx.Connection.NewQuery().Select().From("remote").Command("pipe", pipeRemote).Cursor(nil)
		defer csrRemote.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dRemote := []RemoteModel{}
		remoteData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrRemote.Fetch(&dRemote, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dRemote {
			remoteData = append(remoteData, Leave{
				EmpId:       "",
				Name:        dec.Name,
				Project:     []string{dec.Projects[0].ProjectName},
				RequestType: "Full Remote",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return remoteData

	} else if p.Type == "conditionalremote" {

		pipeRemote := []tk.M{}
		matchRemote := tk.M{}
		matchRemote.Set("userid", userid).Set("type", "1").Set("projects.isleadersend", tk.M{}.Set("$eq", true)).Set("projects.ismanagersend", tk.M{}.Set("$eq", true)).Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true)).Set("isdelete", false).Set("isexpired", false)
		// if p.Yearly == true {
		// 	matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		// } else {
		// 	matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		// }
		if p.Yearly == true {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		} else if p.Date == "32" {
			matchRemote.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		} else {
			matchRemote.Set("dateleave", p.Date)
		}
		pipeRemote = append(pipeRemote, tk.M{}.Set("$match", matchRemote))
		csrRemote, e := c.Ctx.Connection.NewQuery().Select().From("remote").Command("pipe", pipeRemote).Cursor(nil)
		defer csrRemote.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dRemote := []RemoteModel{}
		remoteData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrRemote.Fetch(&dRemote, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dRemote {
			remoteData = append(remoteData, Leave{
				EmpId:       "",
				Name:        dec.Name,
				Project:     []string{dec.Projects[0].ProjectName},
				RequestType: "Conditional Remote",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return remoteData

	} else if p.Type == "overtime" {
		pipe := []tk.M{}
		var matchdate string

		if p.Date == "" {
			if p.Yearly {
				matchdate = strconv.Itoa(p.Year)
			} else {
				matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
			}
		} else {
			matchdate = p.Date
		}

		tk.Println("MATCHDATE....", matchdate)
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.result", "Confirmed")))
		pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}})
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", userid)))
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
		overtimeReturn := []OvertimeResultDetail{}
		tk.Println("Data detail data overtime ... ")
		for _, each := range overtimedata {
			tk.Println(each.MembersOvertime.Name)
			overtimeReturn = append(overtimeReturn, OvertimeResultDetail{
				Id:            each.MembersOvertime.UserId,
				Name:          each.MembersOvertime.Name,
				Date:          each.DayList.Date,
				Actualhours:   0,
				Expectedhours: 0,
				Reason:        each.Reason,
			})
		}
		return overtimeReturn
	} else if p.Type == "overtimeappr" {
		var matchdate string
		if p.Yearly {
			matchdate = strconv.Itoa(p.Year)
		} else {
			matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
		}
		pipe := []tk.M{}
		lookup := tk.M{"$lookup": tk.M{"from": "EmployeeOvertime", "localField": "_id", "foreignField": "idovertime", "as": "employeeovertime"}}
		pipe = append(pipe, lookup)
		unwind1 := tk.M{"$unwind": tk.M{"path": "$employeeovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind1)
		unwind2 := tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind2)
		project := tk.M{
			"$project": tk.M{
				"_id":              1,
				"daylist":          "$daylist",
				"reason":           "$reason",
				"membersovertime":  "$membersovertime",
				"employeeovertime": "$employeeovertime",
				"sign": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$membersovertime.userid", "$employeeovertime.userid"}},
							},
						}, 1, 0,
					},
				},
			},
		}
		pipe = append(pipe, project)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("sign", 1)))
		unwind4 := tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind4)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("employeeovertime.userid", p.Id)))
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
		csr, e := c.Ctx.Connection.NewQuery().Select().From("NewOvertime").Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		datas := []OvertimeDetail{}
		e = csr.Fetch(&datas, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		returndata := []OvertimeResultDetail{}
		for _, each := range datas {
			if each.MembersOvertime.Result == "Confirmed" {
				returndata = append(returndata, OvertimeResultDetail{
					Id:            each.Id,
					Name:          each.MembersOvertime.Name,
					Reason:        each.Reason,
					Actualhours:   each.EmployeeOvertime.TrackHour,
					Expectedhours: each.EmployeeOvertime.Hours,
					Date:          each.DayList.Date,
				})
			}
		}
		return c.SetResultInfo(false, "Success", returndata)
	} else if p.Type == "overtimedec" {
		var matchdate string
		if p.Yearly {
			matchdate = strconv.Itoa(p.Year)
		} else {
			matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
		}
		pipe := []tk.M{}
		unwind1 := tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind1)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", p.Id)))
		unwind2 := tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind2)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.result", "Declined")))
		project := tk.M{
			"$project": tk.M{
				"_id":             1,
				"daylist":         "$daylist",
				"reason":          "$reason",
				"membersovertime": "$membersovertime",
			},
		}
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
		pipe = append(pipe, project)
		csr, e := c.Ctx.Connection.NewQuery().Select().From("NewOvertime").Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		datas := []OvertimeDetail{}
		e = csr.Fetch(&datas, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		returndata := []OvertimeResultDetail{}
		for _, each := range datas {
			returndata = append(returndata, OvertimeResultDetail{
				Id:            each.Id,
				Name:          each.MembersOvertime.Name,
				Reason:        each.Reason,
				Expectedhours: 0,
				Actualhours:   0,
				Date:          each.DayList.Date,
			})
		}
		return c.SetResultInfo(false, "Success", returndata)
	} else if p.Type == "leavedetail" {
		if len(p.ArrEmpid) > 0 {
			leaveDetails := []struct {
				EmpId       string
				Name        string
				Project     []string
				RequestType string
				Date        string
				LeaveReason string
			}{}
			arrtk := []RequestLeaveByDateModel{}
			for _, dataArr := range p.ArrEmpid {

				DpipeELeave := []tk.M{}
				matchELeaveDetails := tk.M{}
				tk.Println("-------------- p,day ", p.Day)

				matchELeaveDetails.Set("yearval", p.Year).Set("stsbymanager", "Approved").Set("isdelete", false)
				if dataArr != "" {
					matchELeaveDetails.Set("empid", dataArr)
				}
				// matchELeaveDetails.Set("isemergency", false)
				if p.Yearly == false && p.Day == 32 {
					matchELeaveDetails.Set("monthval", p.Month)
				} else if p.Yearly == false && p.Day != 32 {
					matchELeaveDetails.Set("monthval", p.Month)
					matchELeaveDetails.Set("dayval", p.Day)
				}
				DpipeELeave = append(DpipeELeave, tk.M{}.Set("$match", matchELeaveDetails))
				csrLeaved, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", DpipeELeave).Cursor(nil)
				defer csrLeaved.Close()
				if e != nil {
					return c.SetResultInfo(true, e.Error(), nil)
				}
				dELeave := []RequestLeaveByDateModel{}
				if e = csrLeaved.Fetch(&dELeave, 0, false); e != nil {
					return c.SetResultInfo(true, e.Error(), nil)
				}
				tk.Println("--------- data detail leave ", tk.JsonString(dELeave))
				for _, dec := range dELeave {
					typstr := ""
					if dec.IsEmergency {
						typstr = "Leave"
					} else {
						typstr = "Emergency Leave"
					}
					dec.Name = strings.ToUpper(dec.Name)
					arrtk = append(arrtk, dec)
					leaveDetails = append(leaveDetails, Leave{
						EmpId:       dec.EmpId,
						Name:        dec.Name,
						Project:     dec.Project,
						RequestType: typstr,
						Date:        dec.DateLeave,
						LeaveReason: dec.Reason,
					})
				}

			}

			return arrtk
		}
	} else if p.Type == "remotedetail" {
		pipe := []tk.M{}
		match := tk.M{}
		match.Set("projects.isleadersend", tk.M{}.Set("$eq", true)).Set("projects.ismanagersend", tk.M{}.Set("$eq", true)).Set("projects.isapprovalmanager", tk.M{}.Set("$eq", true)).Set("isdelete", false).Set("isexpired", false)
		if p.Yearly == true {
			match.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))
		} else if p.Date == "32" {
			match.Set("dateleave", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+"-"+fmt.Sprintf("%02d", p.Month)+".*"))
		} else {
			match.Set("dateleave", p.Date)
		}
		pipe = append(pipe, tk.M{}.Set("$match", match))
		csrRemote, e := c.Ctx.Connection.NewQuery().Select().From("remote").Command("pipe", pipe).Cursor(nil)
		defer csrRemote.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dRemote := []RemoteModel{}
		remoteData := []struct {
			UserID      string
			RequestType string
			Date        string
			LeaveReason string
			Project     string
		}{}
		if e = csrRemote.Fetch(&dRemote, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dRemote {
			if dec.Type == "2" {
				remoteData = append(remoteData, Fullremote{
					UserID:      dec.UserId,
					RequestType: "fullremote",
					Date:        dec.DateLeave,
					LeaveReason: dec.Reason,
					Project:     dec.Projects[0].ProjectName,
				})
			} else if dec.Type == "1" {
				remoteData = append(remoteData, Fullremote{
					UserID:      dec.UserId,
					RequestType: "conditionalremote",
					Date:        dec.DateLeave,
					LeaveReason: dec.Reason,
					Project:     dec.Projects[0].ProjectName,
				})
			}
			remoteData = append(remoteData, Fullremote{
				UserID:      dec.UserId,
				RequestType: "remote",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
				Project:     dec.Projects[0].ProjectName,
			})
		}
		return remoteData
	}

	return c.SetResultInfo(true, "Success", nil)

}

func (c *ReportController) TestPayload(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Type   string
		Empid  string
		Name   string
		Yearly bool
		Month  int
		Year   int
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	return p
}

func (c *ReportController) TestPayloadData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Type   string
		Empid  string
		Name   string
		Yearly bool
		Month  int
		Year   int
	}{}
	if e := k.GetPayload(&p); e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	//get userid from empid
	pipem := []tk.M{}
	matchm := tk.M{}
	matchm.Set("empid", p.Empid)
	pipem = append(pipem, tk.M{}.Set("$match", matchm))
	csrm, e := c.Ctx.Connection.NewQuery().Select().From("SysUsers").Command("pipe", pipem).Cursor(nil)
	defer csrm.Close()
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	dUser := []SysUserModel{}
	if e = csrm.Fetch(&dUser, 0, false); e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	//testLog

	return dUser
}

func (c *ReportController) TestPayloadLeave(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Type   string
		Empid  string
		Name   string
		Yearly bool
		Month  int
		Year   int
	}{}
	if e := k.GetPayload(&p); e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	//get userid from empid
	pipem := []tk.M{}
	matchm := tk.M{}
	matchm.Set("empid", p.Empid)
	pipem = append(pipem, tk.M{}.Set("$match", matchm))
	csrm, e := c.Ctx.Connection.NewQuery().Select().From("SysUsers").Command("pipe", pipem).Cursor(nil)
	defer csrm.Close()
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	dUser := []SysUserModel{}
	if e = csrm.Fetch(&dUser, 0, false); e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	if p.Type == "leave" {
		pipeLeave := []tk.M{}
		matchLeave := tk.M{}
		matchLeave.Set("empid", p.Empid).Set("yearval", p.Year).Set("stsbymanager", "Approved").Set("isemergency", false)
		if p.Yearly == false {
			matchLeave.Set("monthval", p.Month)
		}
		pipeLeave = append(pipeLeave, tk.M{}.Set("$match", matchLeave))
		csrLeave, e := c.Ctx.Connection.NewQuery().Select().From("requestLeaveByDate").Command("pipe", pipeLeave).Cursor(nil)
		defer csrLeave.Close()
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		dLeave := []RequestLeaveByDateModel{}
		leaveData := []struct {
			EmpId       string
			Name        string
			Project     []string
			RequestType string
			Date        string
			LeaveReason string
		}{}
		if e = csrLeave.Fetch(&dLeave, 0, false); e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		for _, dec := range dLeave {
			leaveData = append(leaveData, Leave{
				EmpId:       dec.EmpId,
				Name:        dec.Name,
				Project:     dec.Project,
				RequestType: "Leave",
				Date:        dec.DateLeave,
				LeaveReason: dec.Reason,
			})
		}
		return leaveData

	}
	return (tk.M{}).Set("rc", "00").Set("rd", "Missing")
}

// Overtime ...
func (c *ReportController) Overtime(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadOvertime{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	data := []Overtimereport{}
	var matchdate string
	if p.Day == "32" {
		if p.Yearly {
			matchdate = strconv.Itoa(p.Year)
		} else {
			matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
		}
	} else {
		matchdate = p.Day
	}
	tk.Println("Matchdate overtime ----", matchdate)
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{"$lookup": tk.M{"from": "NewOvertime", "localField": "_id", "foreignField": "membersovertime.userid", "as": "newovertime"}})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$newovertime", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$newovertime.membersovertime", "preserveNullAndEmptyArrays": true}})
	project := tk.M{
		"$project": tk.M{"_id": 1,
			"empid":       1,
			"fullname":    1,
			"newovertime": 1,
			"project":     "$newovertime.project",
			"sign": tk.M{
				"$cond": []interface{}{
					tk.M{
						"$and": []tk.M{
							tk.M{"$eq": []interface{}{"$newovertime.membersovertime.userid", "$_id"}},
						},
					}, 1, 0,
				},
			},
		},
	}
	pipe = append(pipe, project)
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("sign", 1)))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.isexpired", false)))
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$newovertime.daylist", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.membersovertime.result", tk.M{}.Set("$in", []string{"Confirmed", "Declined"}))))
	if len(p.Name) > 0 {
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.membersovertime.name", tk.M{}.Set("$in", p.Name))))
	}
	if len(p.Project) != 0 {
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.project", p.Project)))
	}
	if len(p.Location) != 0 && p.Location != "Global" {
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("newovertime.location", p.Location)))
	}
	csr, e := c.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	e = csr.Fetch(&data, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	result := map[string]*Overtimeresult{}
	for _, each := range data {

		if result[each.NewOvertime.MembersOvertime.IdEmp] == nil {
			result[each.NewOvertime.MembersOvertime.IdEmp] = new(Overtimeresult)
			result[each.NewOvertime.MembersOvertime.IdEmp].Id = each.Empid
			result[each.NewOvertime.MembersOvertime.IdEmp].Name = strings.ToUpper(each.Fullname)
			result[each.NewOvertime.MembersOvertime.IdEmp].EmpId = each.Empid
			result[each.NewOvertime.MembersOvertime.IdEmp].TotalApprove = 0
			result[each.NewOvertime.MembersOvertime.IdEmp].TotalDecline = 0
		}
		if each.NewOvertime.MembersOvertime.Result == "Confirmed" {
			result[each.NewOvertime.MembersOvertime.IdEmp].TotalApprove++
		} else if each.NewOvertime.MembersOvertime.Result == "Declined" {
			result[each.NewOvertime.MembersOvertime.IdEmp].TotalDecline++
		}
	}
	returndata := []*Overtimeresult{}
	type Gadata struct {
		Gendata []*Overtimeresult
		Alldata []Overtimereport
	}

	//aget := new(gadata)

	for _, each := range result {
		tk.Println(each.Name)
		returndata = append(returndata, each)
	}
	sdata := Gadata{Gendata: returndata, Alldata: data}

	return c.SetResultInfo(false, "Success", sdata)
}

// Overtimechart ...
func (c *ReportController) Overtimechart(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadOvertime{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := []*ReportChart{}
	datas = append(datas, &ReportChart{
		Name: "Overtime Approve",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	datas = append(datas, &ReportChart{
		Name: "Overtime Decline",
		Data: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("isexpired", false)))
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+strconv.Itoa(p.Year)+".*"))))
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
	for _, each := range overtimedata {
		for index := 0; index < 12; index++ {
			if bil, _ := strconv.Atoi(each.DayList.Date[5:7]); bil == (index + 1) {
				if each.MembersOvertime.Result == "Confirmed" {
					datas[0].Data[index] = datas[0].Data[index] + 1
				} else if each.MembersOvertime.Result == "Declined" {
					datas[1].Data[index] = datas[1].Data[index] + 1
				}
			}
		}
	}
	return c.SetResultInfo(false, "Success", datas)
}

// OvertimeDetail ...
func (c *ReportController) OvertimeDetail(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := PayloadOvertimeDetail{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	datas := []OvertimeDetail{}
	var matchdate string
	if p.Yearly {
		matchdate = strconv.Itoa(p.Year)
	} else {
		matchdate = strconv.Itoa(p.Year) + "-" + fmt.Sprintf("%02d", p.Month)
	}
	if p.Type == "Confirmed" {
		pipe := []tk.M{}
		lookup := tk.M{
			"$lookup": tk.M{
				"from":         "EmployeeOvertime",
				"localField":   "_id",
				"foreignField": "idovertime",
				"as":           "employeeovertime",
			},
		}
		pipe = append(pipe, lookup)
		unwind1 := tk.M{"$unwind": tk.M{"path": "$employeeovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind1)
		unwind2 := tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind2)
		project := tk.M{
			"$project": tk.M{
				"_id":              1,
				"daylist":          "$daylist",
				"reason":           "$reason",
				"membersovertime":  "$membersovertime",
				"employeeovertime": "$employeeovertime",
				"sign": tk.M{
					"$cond": []interface{}{
						tk.M{
							"$and": []tk.M{
								tk.M{"$eq": []interface{}{"$membersovertime.userid", "$employeeovertime.userid"}},
							},
						}, 1, 0,
					},
				},
			},
		}
		pipe = append(pipe, project)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("sign", 1)))
		unwind4 := tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind4)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("employeeovertime.userid", p.Id)))
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))

		csr, e := c.Ctx.Connection.NewQuery().Select().From("NewOvertime").Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		e = csr.Fetch(&datas, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		returndata := []OvertimeResultDetail{}
		for _, each := range datas {
			if each.MembersOvertime.Result == "Confirmed" {
				returndata = append(returndata, OvertimeResultDetail{
					Id:            each.Id,
					Name:          strings.ToUpper(each.MembersOvertime.Name),
					Reason:        each.Reason,
					Actualhours:   each.EmployeeOvertime.TrackHour,
					Expectedhours: each.EmployeeOvertime.Hours,
					Date:          each.DayList.Date,
				})
			}
		}
		return c.SetResultInfo(false, "Success", returndata)

	} else if p.Type == "Declined" {
		pipe := []tk.M{}
		unwind1 := tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind1)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.userid", p.Id)))
		unwind2 := tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": false}}
		pipe = append(pipe, unwind2)
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("membersovertime.result", "Declined")))
		project := tk.M{
			"$project": tk.M{
				"_id":             1,
				"daylist":         "$daylist",
				"reason":          "$reason",
				"membersovertime": "$membersovertime",
			},
		}
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+matchdate+".*"))))
		pipe = append(pipe, project)
		csr, e := c.Ctx.Connection.NewQuery().Select().From("NewOvertime").Command("pipe", pipe).Cursor(nil)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		defer csr.Close()
		e = csr.Fetch(&datas, 0, false)
		if e != nil {
			return c.SetResultInfo(true, e.Error(), nil)
		}
		returndata := []OvertimeResultDetail{}
		for _, each := range datas {
			returndata = append(returndata, OvertimeResultDetail{
				Id:            each.Id,
				Name:          strings.ToUpper(each.MembersOvertime.Name),
				Reason:        each.Reason,
				Expectedhours: 0,
				Actualhours:   0,
				Date:          each.DayList.Date,
			})
		}

		return c.SetResultInfo(false, "Success", returndata)

	}

	return c.SetResultInfo(false, "Success", nil)

}

// Leave ...
type Leave struct {
	EmpId       string
	Name        string
	Project     []string
	RequestType string
	Date        string
	LeaveReason string
}

// Decline ...
type Decline struct {
	RequestType   string
	Date          string
	LeaveReason   string
	DeclineReason string
}

type Fullremote struct {
	UserID      string
	RequestType string
	Date        string
	LeaveReason string
	Project     string
}

func (c *ReportController) GetRemoteByDay(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := struct {
		Param string
	}{}

	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	pipe := []tk.M{}
	results := tk.M{}
	pipe = append(pipe, tk.M{
		"$lookup": tk.M{
			"from":         "remote",
			"localField":   "_id",
			"foreignField": "userid",
			"as":           "detailRemote",
		},
	})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote", "preserveNullAndEmptyArrays": false}})
	pipe = append(pipe, tk.M{"$project": tk.M{
		"_id":          1,
		"empid":        1,
		"fullname":     1,
		"detailRemote": 1,
		"isreqchange":  "$detailRemote.isrequestchange",
		"dtleave":      "$detailRemote.dateleave",
	},
	})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailRemote.projects", "preserveNullAndEmptyArrays": false}})
	pipe = append(pipe, tk.M{"$project": tk.M{
		"_id":             1,
		"empid":           1,
		"fullname":        1,
		"detailRemote":    1,
		"isreqchange":     1,
		"dtleave":         1,
		"approvalleader":  "$detailRemote.projects.isapprovalleader",
		"approvalmanager": "$detailRemote.projects.isapprovalmanager",
	},
	})
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					//{"approvalleader": tk.M{"$eq": true}},
					{"approvalmanager": tk.M{"$eq": true}},
					//{"isreqchange": tk.M{"$eq": false}},
					{"dtleave": tk.M{"$eq": p.Param}},
					{"detailRemote.isdelete": tk.M{"$eq": false}},
					{"detailRemote.isexpired": tk.M{"$eq": false}},
				},
			},
		},
	)

	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("SysUsers").
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
	results.Set("data", data)
	return c.SetResultInfo(false, "Success", results)
}

func (c *ReportController) GetLeaveByDay(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	p := struct {
		Param string
	}{}

	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	pipe := []tk.M{}
	results := tk.M{}
	pipe = append(pipe, tk.M{
		"$lookup": tk.M{
			"from":         "requestLeaveByDate",
			"localField":   "empid",
			"foreignField": "empid",
			"as":           "detailLeave",
		},
	})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$detailLeave", "preserveNullAndEmptyArrays": false}})
	pipe = append(pipe, tk.M{"$project": tk.M{
		"_id":          1,
		"empid":        1,
		"fullname":     1,
		"dtleave":      "$detailLeave.dateleave",
		"isemergency":  "$detailLeave.isemergency",
		"isdelete":     "$detailLeave.isdelete",
		"stsbymanager": "$detailLeave.stsbymanager",
		"detailLeave":  1,
	},
	})
	pipe = append(pipe,
		tk.M{
			"$match": tk.M{
				"$and": []tk.M{
					{"_id": tk.M{"$ne": ""}},
					{"stsbymanager": tk.M{"$eq": "Approved"}},
					{"isdelete": tk.M{"$eq": false}},
					{"dtleave": tk.M{"$eq": p.Param}},
				},
			},
		},
	)
	// db.SysUsers.aggregate([
	// 	{
	// 		 "$lookup": {
	// 			 "from":         "requestLeaveByDate",
	// 			 "localField":   "empid",
	// 			 "foreignField": "empid",
	// 			 "as":           "detailLeave",
	// 		 }
	// 	},
	// 	{"$unwind": {"path": "$detailLeave", "preserveNullAndEmptyArrays": false}},
	// 	   {
	// 	"$project": {
	// 		"_id" : 1,
	// 		"empid": 1,
	// 		"fullname" : 1,
	// 		"dtleave": "$detailLeave.dateleave",
	// 		"isemergency": "$detailLeave.isemergency",
	// 		"isdelete": "$detailLeave.isdelete",
	// 		"stsbymanager": "$detailLeave.stsbymanager",
	// 		"detailLeave": 1
	// 	}
	// 	},
	// 	 ])
	csr, err := c.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("SysUsers").
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
	results.Set("data", data)
	return c.SetResultInfo(false, "Success", results)
}

// DetailOvertimeExport ...
func (c *ReportController) DetailOvertimeExport(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Param string
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$membersovertime", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{"$unwind": tk.M{"path": "$daylist", "preserveNullAndEmptyArrays": true}})
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("$or", []tk.M{tk.M{}.Set("membersovertime.result", "Confirmed"), tk.M{}.Set("membersovertime.result", "Declined")})))
	if p.Param != "" {
		pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("daylist.date", tk.M{}.Set("$regex", ".*"+p.Param+".*"))))
	}
	csr, e := c.Ctx.Connection.NewQuery().Select().From(NewOvertimeModel().TableName()).Command("pipe", pipe).Cursor(nil)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	data := []CNewOvertimeModel{}
	e = csr.Fetch(&data, 0, false)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}
	returndata := []OvertimeResultExport{}
	for _, each := range data {
		returndata = append(returndata, OvertimeResultExport{
			Name:    strings.ToUpper(each.MembersOvertime.Name),
			EmpId:   each.MembersOvertime.IdEmp,
			Date:    each.DayList.Date,
			Project: each.Project,
			Reason:  each.Reason,
		})
	}
	return c.SetResultInfo(false, "Success", returndata)
}
