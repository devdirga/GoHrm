package webext

import (
	"bufio"
	"bytes"
	. "creativelab/ecleave-dev/controllers"
	"creativelab/ecleave-dev/helper"
	"creativelab/ecleave-dev/repositories"
	"creativelab/ecleave-dev/services"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"html/template"

	. "creativelab/ecleave-dev/models"

	gomail "gopkg.in/gomail.v2"

	"github.com/creativelab/dbox"
	_ "github.com/creativelab/dbox/dbc/mongo"
	knot "github.com/creativelab/knot/knot.v1"
	"github.com/creativelab/orm"
	tk "github.com/creativelab/toolkit"
)

var (
	wd = func() string {
		d, _ := os.Getwd()
		return d + "/"
	}()
)

func init() {
	conn, err := PrepareConnection()
	if err != nil {
		log.Println(err)
	}
	uploadPath := PrepareUploadPath()
	pdfPath := PreparePDFPath()
	logoFile := PrepareLogoFile()
	ctx := orm.New(conn)
	baseCtrl := new(BaseController)
	repositories.Ctx = ctx
	baseCtrl.Ctx = ctx
	baseCtrl.UploadPath = uploadPath
	baseCtrl.PdfPath = pdfPath
	baseCtrl.LogoFile = logoFile
	baseCtrl.DocPath = PrepareDocPath()

	app := knot.NewApp("ecleave")
	app.ViewsPath = wd + "views/"
	app.Register(&LoginController{baseCtrl})
	app.Register(&LogoutController{baseCtrl})
	app.Register(&DashboardController{baseCtrl})
	app.Register(&ProjectController{baseCtrl})
	app.Register(&DatamasterController{baseCtrl})
	app.Register(&UserProfileController{baseCtrl})
	app.Register(&RequestLeaveController{baseCtrl})
	app.Register(&HRDAdminController{baseCtrl})
	app.Register(&StatusRequestController{baseCtrl})
	app.Register(&DeskAprovalController{baseCtrl})
	app.Register(&NotificationController{baseCtrl})
	app.Register(&DesignationController{baseCtrl})
	app.Register(&DepartementController{baseCtrl})
	app.Register(&TypeLeavesController{baseCtrl})
	app.Register(&DoctorCertificateController{baseCtrl})
	// app.Register(&WorkBackgroundController{baseCtrl})

	app.Register(&MenuSettingController{baseCtrl})
	app.Register(&NationalHolidayController{baseCtrl})
	app.Register(&SysRolesController{baseCtrl})
	app.Register(&UserSettingController{baseCtrl})
	app.Register(&MailController{baseCtrl})
	app.Register(&RemoteController{baseCtrl})
	app.Register(&LocationController{baseCtrl})
	app.Register(&RegisterUserController{baseCtrl})
	app.Register(&HistoryForAdminController{baseCtrl})
	app.Register(&ProjectRuleController{baseCtrl})
	app.Register(&LeaveTypeController{baseCtrl})
	app.Register(&BatchController{baseCtrl})
	app.Register(&LogLeaveRemoteController{baseCtrl})
	app.Register(&ReportController{baseCtrl})
	app.Register(&OvertimeController{baseCtrl})

	app.Static("static", wd+"assets")
	app.LayoutTemplate = "_layout.html"
	knot.RegisterApp(app)
	log.Println("___INIT FINISH_____")

	// go SetNameOnHistory(baseCtrl)
	// go DoEvery10Hours(baseCtrl)

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

func PrepareUploadPath() string {
	config := ReadConfig()
	return config["uploadPath"]
}

func PrepareDocPath() string {
	config := ReadConfig()
	return config["docPath"]
}

func PreparePDFPath() string {
	config := ReadConfig()
	return config["pdfPath"]
}

func PrepareLogoFile() string {
	config := ReadConfig()
	return config["logoFile"]
}

func ReadConfig() map[string]string {
	ret := make(map[string]string)
	file, err := os.Open(wd + "conf/app.conf")
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

type Id struct {
	idop []tk.M
}

type dataQuery struct {
	name   string
	reason string
	date   []string
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

	// tk.Println("-------------- location masuk")
	// dbFilter = append(dbFilter, dbox.Eq("location", location))
	dbFilter = append(dbFilter, dbox.Eq("expremining", remaining))

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

	// tk.Println("--------------- res ", data)

	return data
}

func Pluck(c *BaseController, location string, timezone string) interface{} {
	loc, err := helper.TimeLocation(timezone)
	if err != nil {
		return err.Error()
	}
	var dbFilter []*dbox.Filter
	query := tk.M{}
	queryRem := tk.M{}
	data := make([]*RequestLeaveModel, 0)
	dataRemote := make([]*RemoteModel, 0)

	// tk.Println("-------------- location", location)
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

	err = crRem.Fetch(&dataRemote, 0, false)
	if err != nil {
		return nil
	}

	LeaveExpRemaining := LeaveExpRemaining(c, loc)
	RemoteExpRemaining := RemoteExpRemaining(c, loc)
	// tk.Println("----------- datarem ", RemoteExpRemaining)
	ld := 0
	if len(data) > 0 {
		for _, dt := range data {
			// tk.Println("-------------- data", dt)
			if dt.ResultRequest == "Pending" {
				for _, lm := range dt.StatusProjectLeader {
					if lm.StatusRequest == "Pending" {
						ld = ld + 1
					}
				}

				if ld > 0 {
					dt.ResultRequest = "Expired"

					err = c.Ctx.Save(dt)
					if err != nil {
						return nil
					}

					emptyremote := new(RemoteModel)
					service := services.LogService{
						dt,
						emptyremote,
						"leave",
					}
					desc := "Request Expired by Leader"
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
					user := []string{dt.Email}
					SendMailUser(c, user, dt)
				}

			}
		}
	} else if len(LeaveExpRemaining) > 0 {
		user := []string{}
		for _, rim := range LeaveExpRemaining {
			// tk.Println("-------------- remain ", rim)
			for _, lm := range rim.StatusProjectLeader {
				if lm.StatusRequest == "Pending" {
					ld = ld + 1
					user = append(user, lm.Email)
				}
			}
			if len(rim.StatusProjectLeader) == ld {
				if rim.ResultRequest == "Pending" {

					SendLeaveReminder(c, user, rim)
				}
			}

		}
	}

	i := 0
	userRem := []string{}
	if len(dataRemote) > 0 {
		for _, rm := range dataRemote {
			// tk.Println("------------ dataRemote ", rm)
			if rm.Projects[0].IsApprovalLeader == false {
				userRem = append(userRem, rm.Email)
				rm.IsDelete = true
				rm.IsExpired = true
				err = c.Ctx.Save(rm)
				if err != nil {
					return nil
				}
				// for _, pj := range rm.Projects {
				// 	userRem = append(userRem, pj.ProjectLeader.Email)
				// }

				i = i + 1

				emptyleave := new(RequestLeaveModel)
				service := services.LogService{
					emptyleave,
					dataRemote[0],
					"remote",
				}
				desc := "Request Expired by Leader"
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
			}
		}

		pipe := []tk.M{}
		// pipe = append(pipe, tk.M{"$unwind": "$projects"})
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
				tk.Println("------ userRem", userRem)
				SendMailRemote(c, userRem, name, reason, dt)
			}

		}

	} else if len(RemoteExpRemaining) > 0 {
		i := 0
		for _, rm := range RemoteExpRemaining {
			// tk.Println("------------ RemoteExpRemaining ", rm)
			if rm.Projects[0].IsApprovalLeader == false {
				i = i + 1
			}

			for _, pj := range rm.Projects {
				userRem = append(userRem, pj.ProjectLeader.Email)
			}

			if len(rm.Projects) == i {
				pipe := []tk.M{}
				// pipe = append(pipe, tk.M{"$unwind": "$projects"})
				pipe = append(pipe, tk.M{"$match": tk.M{"expiredon": loc}})
				pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalleader": false}})

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
					// tk.Println("---------------- datas2 ", datas)
					return err
				}

				// tk.Println("---------------- datas3 ", datas)
				dt := []string{}
				if i > 0 {
					for _, lm := range datas {
						name := lm.Get("name").([]interface{})[0].(string)
						reason := lm.Get("reason").([]interface{})[0].(string)
						for _, str := range lm.Get("date").([]interface{}) {
							dt = append(dt, str.(string))
						}
						RemoteReminder(c, userRem, name, reason, dt)
					}

				}
			}

		}
	}
	return "success"
}

func DoEvery10Hours(c *BaseController) interface{} {
	// k.Config.OutputType = knot.OutputJson

	// pollInterval := 100
	loc, err := GetLocation(c)

	if err != nil {
		return err.Error()
	}

	timerCh := time.Tick(1 * time.Minute)

	for range timerCh {
		for _, lc := range loc {
			// fmt.Println("------------ masuk ", lc)
			Pluck(c, lc.Location, lc.TimeZone)
		}

	}

	return "nanana"
}

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

type dataR struct {
	NameRemote string
	Reason     string
	Date       []string
}

func SendMailRemote(d *BaseController, to []string, name string, reason string, date []string) interface{} {
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

func RemoteReminder(d *BaseController, to []string, name string, reason string, date []string) interface{} {
	addd := dataR{}
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	tk.Println("------ masuk", to)

	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)

	mailsubj := tk.Sprintf("%v", "Remote request Epired Reminder")
	m := gomail.NewMessage()

	addd.NameRemote = name
	addd.Reason = reason
	addd.Date = date

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileRemote("remoteexpreminder.html", addd)

	if er != nil {
		// tk.Println("------ masuk1", er.Error())
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		// tk.Println("------ masuk1", err.Error())
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
	ret := ResultInfo{}
	conf, emailAddress := EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)
	// urlConf := helper.ReadConfig()
	// hrd := urlConf.GetString("HrdMail")
	// to = append(to, hrd)

	mailsubj := tk.Sprintf("%v", "Leave Request Expired")
	m := gomail.NewMessage()

	if !leave.IsEmergency {
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
	file, err := os.Open(wd + "conf/app.conf")
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
	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
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
