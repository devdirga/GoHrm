package main

import (
	"bufio"
	"bytes"
	. "creativelab/ecleave-dev/controllers"
	"creativelab/ecleave-dev/helper"
	"creativelab/ecleave-dev/repositories"
	"creativelab/ecleave-dev/services"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"html/template"

	. "creativelab/ecleave-dev/models"

	gomail "gopkg.in/gomail.v2"

	"github.com/creativelab/dbox"
	_ "github.com/creativelab/dbox/dbc/mongo"
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
	repositories.Ctx = ctx

	ExpiredRequest(c)

}

func ExpiredRequest(c *BaseController) interface{} {
	loc, err := GetLocation(c)

	if err != nil {
		return err.Error()
	}

	// timerCh := time.Tick(1 * time.Minute)
	fmt.Println("------------ masuk ", loc)
	// for range timerCh {
	for _, lc := range loc {
		// fmt.Println("------------ masuk ", lc)
		Pluck(c, lc.Location, lc.TimeZone)
	}

	return "success"

	// }
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

var (
	basePath = (func(dir string, err error) string { return dir }(os.Getwd()))
)

func ReadNewConfig() tk.M {
	configPath := filepath.Join(basePath, "..", "..", "..", "ecleave-dev", "conf", "newconf.json")
	res := make(tk.M)

	bts, err := ioutil.ReadFile(configPath)
	if err != nil {
		tk.Println("Error when reading config file.", err.Error())
		os.Exit(0)
	}

	err = tk.Unjson(bts, &res)
	if err != nil {
		tk.Println("Error when reading config file.", err.Error())
		os.Exit(0)
	}

	return res
}

func LeaveExpRemaining(c *BaseController, remaining string) []*RequestLeaveModel {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := make([]*RequestLeaveModel, 0)

	// tk.Println("-------------- location", location)
	// dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("expremaining", remaining))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if err != nil {
		return nil
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil
	}

	return data
}

func RemoteExpRemaining(c *BaseController, remaining string) []*RemoteModel {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := make([]*RemoteModel, 0)

	ifs := make([]*RemoteModel, 0)

	// tk.Println("-------------- location masuk")
	// dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("expremining", remaining))
	dbFilter = append(dbFilter, dbox.Eq("isexpired", false))
	// dbFilter = append(dbFilter, dbox.Eq("projects.ismanagersend", true))
	// dbFilter = append(dbFilter, dbox.Eq("projects.isapprovalmanager", false))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRemoteModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		// tk.Println("----------- err ", err.Error())
		return nil
	}
	// defer crs.Close()
	if err != nil {
		// tk.Println("----------- err ", err.Error())
		return nil
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		// tk.Println("----------- err ", err.Error())
		return nil
	}

	for _, dt := range data {
		if dt.Projects[0].IsLeaderSend == false && dt.Projects[0].IsApprovalLeader == false {
			ifs = append(ifs, dt)
		}
	}

	tk.Println("--------------- res remain ", ifs)
	return ifs
}

func OvertimeExpRemaining(c *BaseController, remaining string) []*OvertimeModel {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := make([]*OvertimeModel, 0)

	// tk.Println("-------------- location masuk")
	// dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("expiredremining", remaining))
	dbFilter = append(dbFilter, dbox.Eq("isexpired", false))
	dbFilter = append(dbFilter, dbox.Eq("resultrequest", "Pending"))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		// tk.Println("----------- err ", err.Error())
		return nil
	}
	// defer crs.Close()
	if err != nil {
		// tk.Println("----------- err ", err.Error())
		return nil
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		// tk.Println("----------- err ", err.Error())
		return nil
	}

	// tk.Println("--------------- res ", data)

	return data
}

func Pluck(c *BaseController, location string, timezone string) interface{} {
	loc, err := helper.TimeLocation(timezone)
	logdate, _ := helper.TimeLocationLog(timezone)
	if err != nil {
		return err.Error()
	}
	var dbFilter []*dbox.Filter
	query := tk.M{}
	queryRem := tk.M{}
	data := make([]*RequestLeaveModel, 0)
	dataRemote := make([]*RemoteModel, 0)

	tk.Println("-------------- location", loc)
	// dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("expiredon", loc))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if err != nil {
		return nil
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil
	}

	dbFilter = append(dbFilter, dbox.Eq("isexpired", false))
	if len(dbFilter) > 0 {
		queryRem.Set("where", dbox.And(dbFilter...))
	}
	crRem, err := c.Ctx.Find(NewRemoteModel(), queryRem)
	if crRem != nil {
		defer crRem.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if err != nil {
		return nil
	}

	tk.Println("-------------------- masu error ")

	err = crRem.Fetch(&dataRemote, 0, false)
	if err != nil {
		return nil
	}
	// tk.Println("-------------------- data remote1", dataRemote)

	ovr, err := GetOvertime(c, loc)
	if err != nil {
		return nil
	}

	tk.Println("------------- overtime", ovr)

	LeaveExpRemaining := LeaveExpRemaining(c, loc)
	RemoteExpRemaining := RemoteExpRemaining(c, loc)
	OvertimeExpRemaining := OvertimeExpRemaining(c, loc)
	// tk.Println("----------- datarem ", RemoteExpRemaining)
	// Initiate service
	service := services.LogActivities{}
	//
	ld := 0
	tk.Println(location, data, loc)
	if len(data) > 0 {
		// Initiate log
		// isleaderPending := false
		logsLeave := NewLogCronModel()
		logsLeave.Typelog = "expired-request-leave"
		logsLeave.Date = logdate
		logsLeaveDetail := []LogCronDetailModel{}
		whosPendingL := ""
		//
		for _, dt := range data {
			tk.Println("-------------- data", dt)
			if dt.ResultRequest == "Pending" {
				for _, lm := range dt.StatusProjectLeader {
					if lm.StatusRequest == "Pending" {
						ld = ld + 1
						// 		isleaderPending = true
						// 		whosPendingL = "Leader"
						// 	}
						// }

						// if ld == 0 {
						// 	if dt.StatusManagerProject.StatusRequest == "Pending" {
						// 		whosPendingL = "Manager"
						// 		ld = ld + 1
					}
				}

				if ld > 0 {
					dt.ResultRequest = "Expired"

					err = c.Ctx.Save(dt)
					if err != nil {
						// add log
						logsLeaveDetail = append(logsLeaveDetail, LogCronDetailModel{Message: "Error expired-request-leave (set ResultRequest expired) " + " - " + dt.Name + " - " + dt.UserId})
						logsLeave.Detail = logsLeaveDetail
						service.SaveLog(*logsLeave)
						//
						return nil
					} else {
						logsLeaveDetail = append(logsLeaveDetail, LogCronDetailModel{Message: "Success expired-request-leave (set ResultRequest expired) " + " - " + dt.Name + " - " + dt.UserId})
					}

					emptyremote := new(RemoteModel)
					service := services.LogService{
						dt,
						emptyremote,
						"leave",
					}
					desc := "Request Expired by " + whosPendingL
					var stsReq = "Expired"
					log := tk.M{}
					log.Set("Status", stsReq)
					log.Set("Desc", desc)
					log.Set("NameLogBy", "")
					log.Set("EmailNameLogBy", "")
					err = service.RequestExpired(log)
					if err != nil {
						return nil
					}
					var dateFilter []*dbox.Filter
					queryDate := tk.M{}
					dataDate := make([]*AprovalRequestLeaveModel, 0)
					dateFilter = append(dateFilter, dbox.Eq("idrequest", dt.Id))

					if len(dateFilter) > 0 {
						queryDate.Set("where", dbox.And(dateFilter...))
					}

					crsDate, err := c.Ctx.Find(NewAprovalRequestLeaveModel(), queryDate)
					if crsDate != nil {
						defer crsDate.Close()
					} else {
						return nil
					}
					// defer crsDate.Close()
					if err != nil {
						return nil
					}

					err = crsDate.Fetch(&dataDate, 0, false)
					if err != nil {
						return nil
					}
					// fmt.Println("------------ len ", len(dataDate))
					for _, date := range dataDate {
						date.IsDelete = true
						err = c.Ctx.Save(date)
					}

					SetHistoryLeave(c, dt.UserId, dt.Id, dt.LeaveFrom, dt.LeaveTo, "Request Expired", "Expired", dt)
					logsLeaveDetail = append(logsLeaveDetail, LogCronDetailModel{Message: "Success expired-request-leave (set history leave) " + " - " + dt.Name + " - " + dt.UserId})
					user := []string{dt.Email}
					SendMailUser(c, user, dt)
					// SendMailUser(c, user, dt, isleaderPending)
					logsLeaveDetail = append(logsLeaveDetail, LogCronDetailModel{Message: "Success expired-request-leave (send mail user) " + " - " + dt.Name + " - " + dt.UserId})

					// add notification remote update to expired
					getnotif := GetDataNotification(c, dt.Id)
					getnotif.Notif.Status = "Expired"
					getnotif.Notif.StatusApproval = "Expired"
					if getnotif.Id == "" {
						getnotif.Id = bson.NewObjectId().Hex()
					}
					err = c.Ctx.Save(&getnotif)
					// add notification remote update to expired
				}

			}
		}
		//add log
		logsLeave.Detail = logsLeaveDetail
		service.SaveLog(*logsLeave)
		//
	} else if len(LeaveExpRemaining) > 0 {

		// Initiate log
		logsLeaveRemaining := NewLogCronModel()
		logsLeaveRemaining.Typelog = "expiredrequest-leave-remaining"
		logsLeaveRemaining.Date = logdate
		logsLeaveRemainingDetail := []LogCronDetailModel{}
		//

		user := []string{}
		for _, rim := range LeaveExpRemaining {
			// tk.Println("-------------- remain ", rim)
			for _, lm := range rim.StatusProjectLeader {
				if lm.StatusRequest == "Pending" {
					ld = ld + 1
					user = append(user, lm.Email)
				}
			}

			// if ld == 0 {
			// 	if rim.StatusManagerProject.StatusRequest == "Pending" {
			// 		ld = ld + 1
			// 		for _, bm := range rim.BranchManager {
			// 			user = append(user, bm.Email)
			// 		}
			// 	}
			// }

			if len(rim.StatusProjectLeader) == ld {
				if rim.ResultRequest == "Pending" {

					SendLeaveReminder(c, user, rim)
					logsLeaveRemainingDetail = append(logsLeaveRemainingDetail, LogCronDetailModel{Message: "Success expired-request-leave (send leave reminder)"})
				}
			}

		}
		// add log
		logsLeaveRemaining.Detail = logsLeaveRemainingDetail
		service.SaveLog(*logsLeaveRemaining)
		//
	}

	i := 0
	dtId := ""
	userRem := []string{}
	if len(dataRemote) > 0 {

		// tk.Println("---------------- masuk remote ", dataRemote)

		// Initial log
		logsRemote := NewLogCronModel()
		logsRemote.Typelog = "expired-request-remote"
		logsRemote.Date = logdate
		logDetailRemote := []LogCronDetailModel{}
		//
		whospendingR := ""
		for _, rm := range dataRemote {

			if rm.Projects[0].IsLeaderSend == false && rm.Projects[0].IsApprovalLeader == false {

				tk.Println("---------------- isleadersend ", rm.Projects[0].IsLeaderSend)
				whospendingR = "Leader"
				userRem = append(userRem, rm.Email)
				rm.IsDelete = true
				rm.IsExpired = true
				err = c.Ctx.Save(rm)
				if err != nil {
					//addLog
					logDetailRemote = append(logDetailRemote, LogCronDetailModel{Message: "Error expired-request-Remote - " + rm.Id + " - " + rm.Name + " - " + rm.Reason, Err: err.Error()})
					//
					return nil
				}
				logDetailRemote = append(logDetailRemote, LogCronDetailModel{Message: "Success expired-request-Remote - " + rm.Id + " - " + rm.Name + " - " + rm.Reason, Err: ""})
				// for _, pj := range rm.Projects {
				// 	userRem = append(userRem, pj.ProjectLeader.Email)
				// }

				i = i + 1
				if dtId != rm.IdOp {
					emptyleave := new(RequestLeaveModel)
					service := services.LogService{
						emptyleave,
						dataRemote[0],
						"remote",
					}
					desc := "Request Expired by " + whospendingR
					var stsReq = "Expired"
					log := tk.M{}
					log.Set("Status", stsReq)
					log.Set("Desc", desc)
					log.Set("NameLogBy", "")
					log.Set("EmailNameLogBy", "")
					err = service.RequestExpired(log)
					if err != nil {
						return nil
					}
					tk.Println("---------------- masuk remote ", rm)

					getnotif := GetDataNotification(c, rm.IdOp)
					getnotif.Notif.Status = "Expired"
					getnotif.Notif.StatusApproval = "Expired"
					if getnotif.Id == "" {
						getnotif.Id = bson.NewObjectId().Hex()
					}
					err = c.Ctx.Save(&getnotif)
				}

				dtId = rm.IdOp

			}

			// else if rm.Projects[0].IsApprovalLeader == true && rm.Projects[0].IsManagerSend == false && rm.Projects[0].IsApprovalManager == false {
			// 	tk.Println("---------------- remote ", rm)
			// 	whospendingR = "Manager"
			// 	userRem = append(userRem, rm.Email)
			// 	rm.IsDelete = true
			// 	rm.IsExpired = true
			// 	err = c.Ctx.Save(rm)
			// 	if err != nil {
			// 		//addLog
			// 		logDetailRemote = append(logDetailRemote, LogCronDetailModel{Message: "Error expired-request-Remote - " + rm.Id + " - " + rm.Name + " - " + rm.Reason, Err: err.Error()})
			// 		//
			// 		return nil
			// 	}
			// 	logDetailRemote = append(logDetailRemote, LogCronDetailModel{Message: "Success expired-request-Remote - " + rm.Id + " - " + rm.Name + " - " + rm.Reason, Err: ""})
			// 	// for _, pj := range rm.Projects {
			// 	// 	userRem = append(userRem, pj.ProjectLeader.Email)
			// 	// }

			// 	i = i + 1

			// 	emptyleave := new(RequestLeaveModel)
			// 	service := services.LogService{
			// 		emptyleave,
			// 		dataRemote[0],
			// 		"remote",
			// 	}
			// 	desc := "Request Expired by Manager"
			// 	var stsReq = "Expired"
			// 	log := tk.M{}
			// 	log.Set("Status", stsReq)
			// 	log.Set("Desc", desc)
			// 	log.Set("NameLogBy", "")
			// 	log.Set("EmailNameLogBy", "")
			// 	err = service.RequestExpired(log)
			// 	if err != nil {
			// 		return nil
			// 	}
			// }

			// add notification remote update to expired

			// add notification remote update to expired
		}

		pipe := []tk.M{}
		pipe = append(pipe, tk.M{"$match": tk.M{"isexpired": true}})

		// tk.Println("---------- loc ", loc)

		pipe = append(pipe, tk.M{

			"$match": tk.M{
				"expiredon": tk.M{"$eq": loc},
			},
		})

		pipe = append(pipe, tk.M{

			"$group": tk.M{
				"_id": tk.M{
					"idop": "$idop",
				},
				"name":   tk.M{"$push": "$name"},
				"reason": tk.M{"$push": "$reason"},
				"date":   tk.M{"$push": "$dateleave"},
			},
		})
		// pipe = append(pipe, tk.M{"$unwind": "$projects"})
		// pipe = append(pipe, tk.M{"projects.isleadersend": tk.M{"$eq": false}})
		// pipe = append(pipe, tk.M{"projects.isapprovalleader": tk.M{"$eq": false}})

		// tk.Println("---------------- pipe ", pipe)
		crsR, err := c.Ctx.Connection.NewQuery().Command("pipe", pipe).From("remote").Cursor(nil)
		if crs != nil {
			defer crsR.Close()
		}

		if err != nil {
			return err
		}

		datas := []tk.M{}

		err = crsR.Fetch(&datas, 0, false)
		if err != nil {
			// tk.Println("---------------- datas2 ", datas)
			return err
		}

		tk.Println("----------------  datas ", datas)

		// tk.Println("----------------  i ", i)
		// dt := []string{}
		if i > 0 {
			// tk.Println("----------------  datas ", datas)
			for _, lm := range datas {
				dt := []string{}
				name := lm.Get("name").([]interface{})[0].(string)
				reason := lm.Get("reason").([]interface{})[0].(string)
				for _, str := range lm.Get("date").([]interface{}) {
					dt = append(dt, str.(string))
				}
				tk.Println("---------------- datas user ", dt)
				tk.Println("------ userRem", userRem)
				SendMailRemote(c, userRem, name, reason, dt, whospendingR)
				logDetailRemote = append(logDetailRemote, LogCronDetailModel{Message: "Success send mail (Remote) - " + name + " - " + reason, Err: ""})
			}

		}

		// add log
		logsRemote.Detail = logDetailRemote
		service.SaveLog(*logsRemote)
		//

	} else if len(RemoteExpRemaining) > 0 {
		tk.Println("------------ RemoteExpRemaining masukkkk ")
		whospendingR := ""
		// Initial log
		logsRemoteRemaining := NewLogCronModel()
		logsRemoteRemaining.Typelog = "expired-request-remote-remaining"
		logsRemoteRemaining.Date = logdate
		logDetailRemoteRemaining := []LogCronDetailModel{}
		//

		i := 0
		for _, rm := range RemoteExpRemaining {
			// tk.Println("------------ RemoteExpRemaining ", rm)
			if rm.Projects[0].IsLeaderSend == false && rm.Projects[0].IsApprovalLeader == false {
				i = i + 1

				for _, pj := range rm.Projects {
					userRem = append(userRem, pj.ProjectLeader.Email)
					whospendingR = "Leader"
				}
			}

			// else if rm.Projects[0].IsApprovalLeader == true && rm.Projects[0].IsApprovalManager == false {
			// 	i = i + 1

			// 	for _, Rbm := range rm.BranchManager {
			// 		userRem = append(userRem, Rbm.Email)
			// 		whospendingR = "Manager"
			// 	}
			// }

			// tk.Println("------------ RemoteExpRemaining masukkkk ", i)

			if len(rm.Projects) == i {
				pipe := []tk.M{}
				// pipe = append(pipe, tk.M{"$unwind": "$projects"})
				pipe = append(pipe, tk.M{"$match": tk.M{"expremining": loc}})
				if rm.Projects[0].IsLeaderSend == false && rm.Projects[0].IsApprovalLeader == false {

					pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalleader": false}})
				}

				// else if rm.Projects[0].IsApprovalManager == false {
				// 	// tk.Println("---------------- masuk manager ")
				// 	// pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalleader": true}})
				// 	pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalmanager": false}})
				// }

				// tk.Println("---------- loc ", loc)

				pipe = append(pipe, tk.M{

					"$match": tk.M{
						"expremining": tk.M{"$eq": loc},
					},
				})

				pipe = append(pipe, tk.M{

					"$group": tk.M{
						"_id": tk.M{
							"idop": "$idop",
						},
						"name":   tk.M{"$push": "$name"},
						"reason": tk.M{"$push": "$reason"},
						"date":   tk.M{"$push": "$dateleave"},
					},
				})

				// tk.Println("---------------- pipe ", pipe)
				crsR, err := c.Ctx.Connection.NewQuery().Command("pipe", pipe).From("remote").Cursor(nil)
				if crs != nil {
					defer crsR.Close()
				}

				if err != nil {
					return err
				}

				datas := []tk.M{}

				err = crsR.Fetch(&datas, 0, false)
				if err != nil {

					return err
				}

				tk.Println("---------------- datas2 ", datas)

				dt := []string{}
				if i > 0 {
					tk.Println("---------------- datas3  i ", i)
					for _, lm := range datas {
						name := lm.Get("name").([]interface{})[0].(string)
						reason := lm.Get("reason").([]interface{})[0].(string)
						for _, str := range lm.Get("date").([]interface{}) {
							dt = append(dt, str.(string))
						}
						RemoteReminder(c, userRem, name, reason, dt, whospendingR)
						logDetailRemoteRemaining = append(logDetailRemoteRemaining, LogCronDetailModel{Message: "Success send mail remaining (Remote) - " + name + " - " + reason, Err: ""})
					}

				}
			}

		}
		// add log
		logsRemoteRemaining.Detail = logDetailRemoteRemaining
		service.SaveLog(*logsRemoteRemaining)
		//
	}

	userOver := []string{}
	if len(ovr) > 0 {
		urlConf := ReadNewConfig()
		hrd := urlConf.GetString("HrdMail")

		// Initial log
		logsOvertime := NewLogCronModel()
		logsOvertime.Typelog = "expired-request-overtime"
		logsOvertime.Date = logdate
		logDetailOvertime := []LogCronDetailModel{}
		//

		for _, ov := range ovr {
			ov.ResultRequest = "Expired"
			ov.IsExpired = true
			userOver = append(userOver, ov.ProjectLeader.Email)
			err := c.Ctx.Save(ov)
			if err != nil {
				//addLog
				logDetailOvertime = append(logDetailOvertime, LogCronDetailModel{Message: "Error expired-request-Overtime - " + ov.Id + " - " + ov.Name + " - " + ov.Reason, Err: err.Error()})
				logsOvertime.Detail = logDetailOvertime
				service.SaveLog(*logsOvertime)
				//
				return nil
			}
			//addLog
			logDetailOvertime = append(logDetailOvertime, LogCronDetailModel{Message: "Success expired-request-Remote - " + ov.Id + " - " + ov.Name + " - " + ov.Reason, Err: ""})
			//
			dt := []string{}
			for _, dat := range ov.DayList {
				dt = append(dt, dat.Date)
			}
			for _, mm := range ov.BranchManagers {
				SendMailExpiredOvertime(c, []string{mm.Email}, mm.Name, dt, ov.DateCreated, ov.Project)
				DelayProcess(5)
			}
			SendMailExpiredOvertime(c, []string{hrd}, ov.ProjectLeader.Name, dt, ov.DateCreated, ov.Project)
			DelayProcess(5)
			SendMailExpiredOvertime(c, []string{ov.ProjectLeader.Email}, ov.ProjectLeader.Name, dt, ov.DateCreated, ov.Project)
			DelayProcess(5)
			for _, dov := range ov.MembersOvertime {
				if dov.UserId != ov.UserId {
					DelayProcess(5)
					SendAssignedMailExpiredOvertime(c, []string{dov.Email}, dov.Name, dt, ov.DateCreated, ov.Project)
				}
			}
			//track log overtime expired
			service := services.LogServiceOvertime{TypeRequest: "overtime", DataOvertime: ov}
			desc := "Request Expired by Manager"
			var stsReq = "Expired"
			log := tk.M{}
			log.Set("Status", stsReq)
			log.Set("Desc", desc)
			log.Set("NameLogBy", "")
			log.Set("EmailNameLogBy", "")
			err = service.RequestExpired(log)
			if err != nil {
				return nil
			}

			getnotif := GetDataNotification(c, ov.Id)
			getnotif.Notif.ManagerApprove = ov.ApprovalManager.Name
			getnotif.Notif.Status = ov.ResultRequest
			getnotif.Notif.StatusApproval = "Expired"
			if getnotif.Id == "" {
				getnotif.Id = bson.NewObjectId().Hex()
			}
			err = c.Ctx.Save(&getnotif)

		}
		// add log
		logsOvertime.Detail = logDetailOvertime
		service.SaveLog(*logsOvertime)
		//
	} else if len(OvertimeExpRemaining) > 0 {
		// Initial log
		logsOvertimeRem := NewLogCronModel()
		logsOvertimeRem.Typelog = "expired-request-overtime-remaining"
		logsOvertimeRem.Date = logdate
		logDetailOvertimeRem := []LogCronDetailModel{}
		//
		for _, ovrRem := range OvertimeExpRemaining {

			date := []string{}
			for _, dat := range ovrRem.DayList {
				date = append(date, dat.Date)
			}
			SendManagerRemainingOvertime(c, []string{ovrRem.ProjectManager.Email}, ovrRem.ProjectManager.Name, ovrRem.ProjectLeader.Name, ovrRem.ProjectManager.UserId, date, ovrRem.DateCreated, ovrRem.Project, *ovrRem)
			//addLog
			logDetailOvertimeRem = append(logDetailOvertimeRem, LogCronDetailModel{Message: "Success expired-request-Overtime-Remaining - " + ovrRem.Id + " - " + ovrRem.Name + " - " + ovrRem.Reason, Err: ""})
			//
			for _, br := range ovrRem.BranchManagers {
				SendManagerRemainingOvertime(c, []string{br.Email}, br.Name, ovrRem.ProjectLeader.Name, br.UserId, date, ovrRem.DateCreated, ovrRem.Project, *ovrRem)
			}
		}

		// add log
		logsOvertimeRem.Detail = logDetailOvertimeRem
		service.SaveLog(*logsOvertimeRem)
		//

	}
	return "success"
}

func DelayProcess(n time.Duration) {
	time.Sleep(n * time.Second)
}

func GetOvertime(c *BaseController, timeexp string) ([]*OvertimeModel, error) {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := make([]*OvertimeModel, 0)
	dbFilter = append(dbFilter, dbox.Eq("expiredon", timeexp))
	dbFilter = append(dbFilter, dbox.Eq("isexpired", false))
	dbFilter = append(dbFilter, dbox.Eq("resultrequest", "Pending"))

	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewOvertimeModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}
	// defer crs.Close()
	if err != nil {
		return nil, err
	}

	tk.Println("-------------------- masu error ")

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// func DoEvery10Hours(c *BaseController) interface{} {
// 	// k.Config.OutputType = knot.OutputJson

// 	// pollInterval := 100
// 	loc, err := GetLocation(c)

// 	if err != nil {
// 		return err.Error()
// 	}

// 	timerCh := time.Tick(1 * time.Minute)

// 	for range timerCh {
// 		for _, lc := range loc {
// 			// fmt.Println("------------ masuk ", lc)
// 			Pluck(c, lc.Location, lc.TimeZone)
// 		}

// 	}

// 	return "nanana"
// }

func GetLocation(c *BaseController) ([]LocationModel, error) {
	data := make([]LocationModel, 0)
	crs, err := c.Ctx.Find(NewLocationModel(), nil)
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

func SetHistoryLeave(c *BaseController, userid string, idreq string, dateFrom string, dateTo string, description string, status string, leave *RequestLeaveModel) bool {
	// c.LoadBase(k)

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

	return true
}

func SetNameOnHistory(c *BaseController) interface{} {

	data := make([]*HistoryLeaveModel, 0)
	user := make([]SysUserModel, 0)
	crs, errdata := c.Ctx.Find(NewHistoryLeaveModel(), nil)

	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil
	}

	// tk.Println("---------- data ", data)

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil
	}

	crsUser, errUser := c.Ctx.Find(NewSysUserModel(), nil)

	if crsUser != nil {
		defer crsUser.Close()
	} else {
		return nil
	}
	// defer crsUser.Close()
	if errUser != nil {
		return nil
	}

	errUser = crsUser.Fetch(&user, 0, false)
	if errUser != nil {
		return nil
	}
	for i, _ := range data {
		for _, usr := range user {
			for m, _ := range data[i].Leavehistory {
				if data[i].UserId == usr.Id {
					data[i].Leavehistory[m].UserId = usr.Id
					data[i].Leavehistory[m].Name = usr.Fullname

					err := c.Ctx.Save(data[i])
					if err != nil {
						return nil
					}
				}
			}
		}
	}

	return ""

}

type dataOver struct {
	NameLeader  string
	NameManager string
	Date        []string
	DateRequest string
	Project     string
	URLApproved string
	URLDeclined string
	URL         string
}

func FileOvertime(filename string, data dataOver) ([]byte, error) {
	fmt.Println("------ masuk file")
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "..", "..", "views", "template", filename))
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

func SendManagerRemainingOvertime(d *BaseController, to []string, managername string, leadername string, useridmanager string, date []string, daterequest string, project string, data OvertimeModel) interface{} {
	addd := dataOver{}
	urlConf := ReadNewConfig()
	hrd := urlConf.GetString("HrdMail")
	to = append(to, hrd)

	// tk.Println("------ masuk", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Overtime Request for "+project+" - "+leadername)
	m := gomail.NewMessage()
	sUrl := urlConf.GetString("BaseUrlEmail")
	uriApprove := sUrl + "/overtime/approvedovertime"
	param := new(ParameterURLUserModel)
	param.Name = data.ProjectManager.Name
	param.IdOvertime = data.Id
	param.Project = data.Project
	param.UserId = useridmanager
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

	urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
	qDecline := urlDecline.URL.Query()
	qDecline.Add("param", GCMEncrypter(string(paramDec)))
	urlDecline.URL.RawQuery = qDecline.Encode()

	addd.NameLeader = leadername
	addd.NameManager = managername
	addd.Date = date
	addd.DateRequest = daterequest
	addd.Project = project
	addd.URLApproved = urlApprove.URL.String()
	addd.URLDeclined = urlDecline.URL.String()
	addd.URL = urlmgrpage.URL.String()

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileOvertime("overtimeremainingexpired.html", addd)

	if er != nil {
		tk.Println("masuk eroro ", er.Error())
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""
}

func SendMailExpiredOvertime(d *BaseController, to []string, name string, date []string, daterequest string, project string) interface{} {
	addd := dataOver{}
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	// tk.Println("------ masuk", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Overtime Request Expired for "+daterequest)
	m := gomail.NewMessage()

	addd.NameLeader = name
	addd.Date = date
	addd.DateRequest = daterequest
	addd.Project = project

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileOvertime("overtimeexpired.html", addd)

	if er != nil {
		tk.Println("masuk eroro ", er.Error())
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""
}

func SendAssignedMailExpiredOvertime(d *BaseController, to []string, name string, date []string, daterequest string, project string) interface{} {
	addd := dataOver{}
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	// tk.Println("------ masuk", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Overtime Info Request Expired for "+daterequest)
	m := gomail.NewMessage()

	addd.NameLeader = name
	addd.Date = date
	addd.DateRequest = daterequest
	addd.Project = project

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileOvertime("overtimeassignedexpired.html", addd)

	if er != nil {
		tk.Println("masuk eroro ", er.Error())
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""
}

type dataR struct {
	NameRemote  string
	Reason      string
	Date        []string
	WhosPending string
}

func SendMailRemote(d *BaseController, to []string, name string, reason string, date []string, whospending string) interface{} {
	addd := dataR{}
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	// tk.Println("------ masuk", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Remote Request Expired")
	m := gomail.NewMessage()

	addd.NameRemote = name
	addd.Reason = reason
	addd.Date = date
	addd.WhosPending = whospending

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileRemote("remoteexpired.html", addd)

	if er != nil {
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""
}

func RemoteReminder(d *BaseController, to []string, name string, reason string, date []string, Whospending string) interface{} {
	addd := dataR{}
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	tk.Println("------ masuk email remainder to nya", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Remote request Expired Reminder")
	m := gomail.NewMessage()

	addd.NameRemote = name
	addd.Reason = reason
	addd.Date = date
	addd.WhosPending = Whospending

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileRemote("remoteexpreminder.html", addd)

	if er != nil {
		tk.Println("------ masuk1 email remainder ", er.Error())
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		tk.Println("------ masuk1 email remainder email remainder", err.Error())
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ""
}

func SendLeaveReminder(d *BaseController, to []string, leave *RequestLeaveModel) interface{} {
	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	mailsubj := tk.Sprintf("%v", "Leave Request Expired Reminder")
	m := gomail.NewMessage()

	if !leave.IsEmergency {
		addd := map[string]string{"namerequest": leave.Name, "reason": leave.Reason, "leaveFrom": leave.LeaveFrom, "leaveTo": leave.LeaveTo, "noOfDays": strconv.Itoa(leave.NoOfDays), "result": leave.ResultRequest, "managerReason": leave.StatusManagerProject.Reason, "DateCreate": leave.DateCreateLeave}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := File("leaveexpiredremaining.html", addd)

		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	} else {
		addd := map[string]string{"namerequest": leave.Name, "reason": leave.Reason, "leaveFrom": leave.LeaveFrom, "leaveTo": leave.LeaveTo, "noOfDays": strconv.Itoa(leave.NoOfDays), "result": leave.ResultRequest, "managerReason": leave.StatusManagerProject.Reason, "DateCreate": leave.DateCreateLeave}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := File("expired.html", addd)
		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	}

	return ret
}
func SendMailUser(d *BaseController, to []string, leave *RequestLeaveModel) interface{} {
	// func SendMailUser(d *BaseController, to []string, leave *RequestLeaveModel, isleaderPending bool) interface{} {
	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)
	// whosPending := ""
	// if !isleaderPending {
	// 	whosPending = "Manager"
	// } else {
	// 	whosPending = "Leader"
	// }
	mailsubj := tk.Sprintf("%v", "Leave Request Expired")
	m := gomail.NewMessage()

	if !leave.IsEmergency {
		// addd := map[string]string{"namerequest": leave.Name, "reason": leave.Reason, "leaveFrom": leave.LeaveFrom, "leaveTo": leave.LeaveTo, "noOfDays": strconv.Itoa(leave.NoOfDays), "result": leave.ResultRequest, "managerReason": leave.StatusManagerProject.Reason, "DateCreate": leave.DateCreateLeave, "WhosPending": whosPending}
		addd := map[string]string{"namerequest": leave.Name, "reason": leave.Reason, "leaveFrom": leave.LeaveFrom, "leaveTo": leave.LeaveTo, "noOfDays": strconv.Itoa(leave.NoOfDays), "result": leave.ResultRequest, "managerReason": leave.StatusManagerProject.Reason, "DateCreate": leave.DateCreateLeave}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := File("expired.html", addd)

		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	} else {
		addd := map[string]string{"namerequest": leave.Name, "reason": leave.Reason, "leaveFrom": leave.LeaveFrom, "leaveTo": leave.LeaveTo, "noOfDays": strconv.Itoa(leave.NoOfDays), "result": leave.ResultRequest, "managerReason": leave.StatusManagerProject.Reason, "DateCreate": leave.DateCreateLeave}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := File("expired.html", addd)
		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	}

	return ret
}

func EmailConfiguration() (*gomail.Dialer, string) {
	// r.Config.OutputType = knot.OutputJson
	config := config()
	conf := gomail.NewPlainDialer(config["Host"], 587, config["MailAddressName"], config["MailAddressPassword"])
	emailAddress := config["MailAddressName"]

	return conf, emailAddress
}

func config() map[string]string {
	ret := make(map[string]string)
	file, err := os.Open(wd + "../../conf/app.conf")
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

func File(filename string, data interface{}) (string, error) {
	t, err := os.Getwd()
	if err != nil {
		return "", err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "..", "..", "views", "template", filename))
	tk.Println("--------- template ", filepath.Join(t, "..", "..", "views", "template", filename))
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		return "", err
	}
	body := buffer.String()
	// uri := url.Values{}

	// fmt.Println("------", body)
	return body, nil
}

func FileRemote(filename string, data dataR) ([]byte, error) {
	fmt.Println("------ masuk file")
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "..", "..", "views", "template", filename))
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

// GetDataNotification ...
func GetDataNotification(c *BaseController, idrequest string) NotificationModel {
	var dbFilter []*dbox.Filter
	query := tk.M{}
	data := []NotificationModel{}
	dbFilter = append(dbFilter, dbox.Eq("idrequest", idrequest))
	if len(dbFilter) > 0 {
		query.Set("where", dbox.And(dbFilter...))
	}
	crs, err := c.Ctx.Find(NewNotificationModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return data[0]
	}
	if err != nil {
		return data[0]
	}
	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data[0]
	}
	return data[0]
}
