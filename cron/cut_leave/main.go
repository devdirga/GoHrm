package main

import (
	"bufio"
	. "creativelab/ecleave-dev/controllers"
	"creativelab/ecleave-dev/repositories"
	"creativelab/ecleave-dev/services"
	"log"
	"os"
	"strings"

	"github.com/creativelab/dbox"
	_ "github.com/creativelab/dbox/dbc/mongo"

	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/orm"

	tk "github.com/creativelab/toolkit"
)

var (
	wd = func() string {
		d, _ := os.Getwd()
		return d + "/"
	}()
)

func main() {
	conn, err := PrepareConnection()
	if err != nil {
		log.Println(err)
	}

	ctx := orm.New(conn)

	c := new(BaseController)
	c.Ctx = ctx

	CutLeave(c)
}

func PrepareConnection() (dbox.IConnection, error) {
	config := ReadConfig()
	ci := &dbox.ConnectionInfo{config["host"], config["database"], config["username"], config["password"], nil}
	c, e := dbox.NewConnection("mongo", ci)

	if e != nil {
		return nil, e
	}

	e = c.Connect()
	if e != nil {
		return nil, e
	}

	return c, nil
}

func ReadConfig() map[string]string {
	ret := make(map[string]string)
	file, err := os.Open("../../conf/app.conf")
	if err == nil {
		defer file.Close()

		reader := bufio.NewReader(file)
		for {
			line, _, e := reader.ReadLine()
			if e != nil {
				break
			}

			sval := strings.Split(string(line), "=")
			ret[sval[0]] = sval[1]
		}
	} else {
		log.Println(err.Error())
	}

	return ret
}

func CutLeave(c *BaseController) interface{} {

	loc, err := GetLocation(c)

	if err != nil {
		return err.Error()
	}

	for _, lc := range loc {

		tk.Println("---------- location ", lc.Location)

		logdate, _ := helper.TimeLocationLog(lc.TimeZone)

		day, month, year := GetDayMonth(lc.TimeZone)
		data := GetLeaveDate(c, day, month, year, lc.Location)
		if len(data) > 0 {
			//...add log
			logs := NewLogCronModel()
			logs.Typelog = "cut-leave"
			logs.Date = logdate
			logDetail := []LogCronDetailModel{}
			repositories.Ctx = c.Ctx
			service := services.LogActivities{}
			//...
			for _, det := range data {
				if det.IsCutOff == false {
					// indetail, err := GetLeaveDetails(c, det.IdRequest)
					// if err != nil {
					// 	return err.Error()
					// }

					is := CheckIsSpcialIsReset(c, det.IdRequest, lc.Location, det.IsReset)
					tk.Println("----------- is ", is)
					if is == false {
						det.IsCutOff = true

						err = c.Ctx.Save(&det)

						if err != nil {
							tk.Println(err.Error())
						}
						// cut leave here
						// tk.Println("------ indetail ", indetail)
						usr, err := GetuserData(c, det.UserId)
						if err != nil {
							tk.Println(err.Error())
						}

						tk.Println("---------- usr ", usr)
						if usr.DecYear > 0.0 {
							leave := usr.DecYear - 1.0
							usr.DecYear = leave
							usr.YearLeave = int(leave)
						} else if usr.DecYear < 0.0 {
							usr.DecYear = float64(usr.YearLeave)
							leave := usr.DecYear - 1.0
							usr.DecYear = leave        // cek ini
							usr.YearLeave = int(leave) // cek ini
						}
						// usr.YearLeave = usr.YearLeave - 1
						err = c.Ctx.Save(&usr)
						if err != nil {
							//...add log
							logDetail = append(logDetail, LogCronDetailModel{Date: logdate, Message: "Error cut-leave" + " - " + usr.EmpId + " - " + usr.Fullname})
							//...
							tk.Println(err.Error())
						} else {
							//...add log
							logDetail = append(logDetail, LogCronDetailModel{Date: logdate, Message: "Success cut-leave" + " - " + usr.EmpId + " - " + usr.Fullname})
							//...
						}
					} else {
						tk.Println("empty")
					}
				}

			}
			//...add log
			logs.Detail = logDetail
			service.SaveLog(*logs)
			//...
		}

	}

	return "nananana"
}

func GetuserData(c *BaseController, userid string) (SysUserModel, error) {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := SysUserModel{}
	dbFilter = append(dbFilter, dbox.Eq("_id", userid))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
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

func GetDayMonth(timezone string) (int, int, int) {
	now, err := helper.TimeLocationUnFormat(timezone)
	if err != nil {
		return 0, 0, 0
	}
	return now.Day(), int(now.Month()), int(now.Year())
}

func CheckIsSpcialIsReset(c *BaseController, idrequest string, location string, isreset bool) bool {
	tk.Println("------------   id0", idrequest)
	data, err := GetLeaveDetails(c, idrequest)

	if err != nil {
		tk.Println("------------  masuk sini0")
		return true
	}

	// tk.Println("------------  location ", location)
	// tk.Println("------------  data location ", data.Location)
	if location == data.Location {
		if isreset == true {
			tk.Println("------------  masuk sini1")
			return true
		} else if data.IsSpecials == true {
			tk.Println("------------  masuk sini2")
			return true
		} else {
			return false
		}
	} else {
		tk.Println("------------  masuk sini")
		return true
	}

	return false
}

func GetLeaveDetails(c *BaseController, idrequest string) (RequestLeaveModel, error) {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := RequestLeaveModel{}
	dbFilter = append(dbFilter, dbox.Eq("_id", idrequest))
	dbFilter = append(dbFilter, dbox.Eq("resultrequest", "Approved"))
	// dbFilter = append(dbFilter, dbox.Eq("iscutoff", false))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
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

func GetLeaveDate(c *BaseController, day int, month int, year int, location string) []AprovalRequestLeaveModel {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	tk.Println("------------- data location ", location)

	data := make([]AprovalRequestLeaveModel, 0)
	dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("dayval", day))
	dbFilter = append(dbFilter, dbox.Eq("monthval", month))
	dbFilter = append(dbFilter, dbox.Eq("yearval", year))
	dbFilter = append(dbFilter, dbox.Eq("isdelete", false))
	dbFilter = append(dbFilter, dbox.Eq("isreset", false))

	// dbFilter = append(dbFilter, dbox.Eq("isreset", false))
	dbFilter = append(dbFilter, dbox.Eq("stsbymanager", "Approved"))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return data
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data
	}

	// tk.Println("------------- res data ", data)
	return data
}

func GetLocation(c *BaseController) ([]LocationModel, error) {
	data := make([]LocationModel, 0)
	crs, err := c.Ctx.Find(NewLocationModel(), nil)
	// tk.Println("-------------- ", err.Error())
	if crs != nil {
		defer crs.Close()
	} else {
		return data, nil
	}
	if err != nil {
		return nil, err
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil, err
	}

	return data, nil
}
