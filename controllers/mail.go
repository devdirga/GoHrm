package controllers

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"creativelab/ecleave-dev/helper"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/services"
	"html/template"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	gomail "gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

type MailController struct {
	*BaseController
}

type MailConfig struct {
	dialer  tk.M
	address string
}

func MailReadConfig() map[string]string {
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
func (d *MailController) EmailConfiguration() (*gomail.Dialer, string) {
	// r.Config.OutputType = knot.OutputJson
	config := MailReadConfig()
	conf := gomail.NewPlainDialer(config["Host"], 587, config["MailAddressName"], config["MailAddressPassword"])
	emailAddress := config["MailAddressName"]

	return conf, emailAddress
}

func (d *MailController) File(filename string, data interface{}) (string, error) {
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

func (d *MailController) getDataLeave(Id string) *RequestLeaveModel {
	dataLeave := make([]*RequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("_id", Id))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := d.Ctx.Find(NewRequestLeave(), query)
	if err != nil {
		fmt.Println(err.Error)
	}
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		fmt.Println(err.Error)
	}

	return dataLeave[0]
}

func (d *MailController) SendMailUser(r *knot.WebContext, ID string, to []string) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	dataRequest := d.getDataLeave(ID)

	dataRequestByDate := d.getDataLeaveByDate(ID)
	noPaidLeave := 0
	paidLeave := 0
	for _, requestbydate := range dataRequestByDate {
		if requestbydate.IsPaidLeave {
			paidLeave++
		} else {
			noPaidLeave++
		}
	}

	// if len(dataRequest.ProjectManagerList) > 0 {
	// 	for _, pm := range dataRequest.ProjectManagerList {
	// 		to = append(to, pm.Email)
	// 	}
	// }
	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	to = append(to, hrd)

	mailsubj := tk.Sprintf("%v", "User Project Leave Request Attachment")
	m := gomail.NewMessage()

	if !dataRequest.IsEmergency {
		addd := map[string]string{"namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "result": dataRequest.ResultRequest, "managerReason": dataRequest.StatusManagerProject.Reason, "DateCreate": dataRequest.DateCreateLeave, "noPaidLeave": strconv.Itoa(noPaidLeave), "paidLeave": strconv.Itoa(paidLeave)}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := d.File("requestmail.html", addd)

		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		d.DelayProcess(5)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	} else {
		addd := map[string]string{"namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "result": dataRequest.ResultRequest, "managerReason": dataRequest.StatusManagerProject.Reason, "DateCreate": dataRequest.DateCreateLeave, "noPaidLeave": strconv.Itoa(noPaidLeave), "paidLeave": strconv.Itoa(paidLeave)}

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", mailsubj)

		bd, er := d.File("emergencyRequestEmail.html", addd)
		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)

		d.DelayProcess(5)

		if err := conf.DialAndSend(m); err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		} else {
			ret.Data = "Successfully Send Mails"
		}
		m.Reset()
	}

	return ret
}

func (d *MailController) SendMailUserCancelByDate(r *knot.WebContext, to []string, dateleave string, reasonCancel string, by string, ontype string, manager string, leader string, username string) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", "User "+ontype+" cancel confirmation")
	m := gomail.NewMessage()

	tk.Println("-------- reasonCancel ", reasonCancel)

	addd := map[string]string{"dateLeave": dateleave, "by": by, "reason": reasonCancel, "type": ontype, "nameUser": username, "leader": leader, "manager": manager}

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := d.File("CancelByDate.html", addd)

	if er != nil {
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", bd)

	d.DelayProcess(5)

	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ret
}

func (d *MailController) DelayProcess(n time.Duration) {
	time.Sleep(n * time.Second)
}

func (d *MailController) SendMailManagerDetails(r *knot.WebContext, name string, reason string, req *RequestLeaveModel) interface{} {
	addd := dataR{}
	detail := dataDetails{}
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)
	to := []string{}
	// urlConf := helper.ReadConfig()
	// sUrl := urlConf.GetString("BaseUrlEmail")
	for _, det := range req.DetailsLeave {
		detail.Date = det.DateLeave
		detail.Name = det.LeaderName
		if det.IsApproved {
			detail.Result = "Approved"
		} else {
			detail.Result = "Declined"
		}
		detail.Reason = det.Reason

		addd.Details = append(addd.Details, detail)

	}

	addd.NameRemote = name
	addd.Reason = reason

	for _, usr := range req.BranchManager {
		to = append(to, usr.Email)

	}
	mailsubj := tk.Sprintf("%v", "Result leave by Leader")
	m := gomail.NewMessage()

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileLeaveTemplate("leavedetailleader.html", addd)

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

// func (d *MailController) SendELAttacmentAdmin(r *knot.WebContext, to []string, name string, reason string, date []string) interface{} {
// 	r.Config.OutputType = knot.OutputJson
// 	addd := containEmaiR{}

// 	ret := ResultInfo{}
// 	conf, emailAddress := d.EmailConfiguration()
// 	// dataRequest := d.getDataLeave(ID)

// 	mailsubj := tk.Sprintf("%v", "Remote Request Expired")
// 	m := gomail.NewMessage()

// 	addd.NameRemote = name
// 	addd.Reason = reason
// 	addd.Date = date

// 	m.SetHeader("From", emailAddress)
// 	m.SetHeader("To", to...)
// 	m.SetHeader("Subject", mailsubj)

// 	bd, er := FileForEmail("remoteexpired.html", addd)

// 	if er != nil {
// 		return d.SetResultInfo(true, er.Error(), nil)
// 	}
// 	m.SetBody("text/html", string(bd))

// 	if err := conf.DialAndSend(m); err != nil {
// 		return d.SetResultInfo(true, err.Error(), nil)
// 	} else {
// 		ret.Data = "Successfully Send Mails"
// 	}
// 	m.Reset()

// 	return ""
// }

// func FileForEmail(filename string, data containEmaiR) ([]byte, error) {
// 	fmt.Println("------ masuk file")
// 	t, err := os.Getwd()
// 	body := []byte{}
// 	if err != nil {
// 		return body, err
// 	}
// 	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
// 	if err != nil {
// 		return body, err
// 	}
// 	// fmt.Println("------ masuk data ", data.Date)
// 	buffer := new(bytes.Buffer)
// 	if err = templ.Execute(buffer, data); err != nil {
// 		return body, err
// 	}
// 	body = buffer.Bytes()

// 	return body, nil
// }

// type containEmaiR struct {
// 	Name   string
// 	Reason string
// 	Date   string
// 	Type   string
// }

func (d *MailController) SendMailManager(r *knot.WebContext, userid string, ID string, level int) interface{} {
	// tk.Println("---------------- masuk sendMailManager", ID)
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	urlConf := helper.ReadConfig()

	dataRequest := d.getDataLeave(ID)
	dataRequestByDate := d.getDataLeaveByDate(ID)
	noPaidLeave := 0
	paidLeave := 0
	for _, requestbydate := range dataRequestByDate {
		if requestbydate.IsPaidLeave {
			paidLeave++
		} else {
			noPaidLeave++
		}
	}

	dataResponseLeader := dataRequest.DetailsLeave
	// level := r.Session("jobrolelevel").(int)
	dt := dataRequest.StatusManagerProject
	sUrl := urlConf.GetString("BaseUrlEmail")
	dash := DashboardController(*d)

	for _, dtl := range dataRequest.BranchManager {
		uriApprove := sUrl + "/mail/responsemanager"
		var dec = new(ParameterURLManagerModel)
		dec.UserId = userid
		dec.IdRequest = string(dataRequest.Id)
		dec.UserIdManager = dtl.UserId
		dec.IdManager = dt.IdEmp
		dec.ApproveManager = "yes"
		paramApp, _ := json.Marshal(dec)

		urlApprove, _ := http.NewRequest("GET", uriApprove, nil)
		qApprove := urlApprove.URL.Query()
		qApprove.Add("param", GCMEncrypter(string(paramApp)))
		urlApprove.URL.RawQuery = qApprove.Encode()

		uriDecline := sUrl + "/mail/responsemanagerdecline"

		dec.ApproveManager = "no"
		paramDec, _ := json.Marshal(dec)

		// fmt.Println("------------ leader masuk send manager", dec.UserId)

		urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
		qDecline := urlDecline.URL.Query()
		qDecline.Add("param", GCMEncrypter(string(paramDec)))
		urlDecline.URL.RawQuery = qDecline.Encode()

		conf, emailAddress := d.EmailConfiguration()

		m := gomail.NewMessage()

		// fmt.Println("----------- dt.Email", dt.Email)

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", dtl.Email)

		addd := map[string]string{"to": dtl.Name, "namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": dataRequest.DateCreateLeave, "noPaidLeave": strconv.Itoa(noPaidLeave), "paidLeave": strconv.Itoa(paidLeave)}

		if !dataRequest.IsEmergency {
			mailsubj := tk.Sprintf("%v", "Leave Request from "+dataRequest.Name)
			m.SetHeader("Subject", mailsubj)

			nb := ""
			sttsleader := 0
			if len(dataResponseLeader) > 0 {
				for _, led := range dataResponseLeader {
					if led.IsApproved == false {
						sttsleader = sttsleader + 1
						if nb != "" {
							nb = nb + "Leader Project " + led.LeaderName + " had been Declined leave of " + led.DateLeave + " because : " + led.Reason + " , "
						} else {
							nb = "Leader Project " + led.LeaderName + " had been Declined leave of  " + led.DateLeave + " because : " + led.Reason + " , "
						}
					} else if led.IsApproved == true {
						sttsleader = sttsleader + 1
						if nb != "" {
							nb = nb + "Leader Project " + led.LeaderName + " had been Approved leave of " + led.DateLeave + " because :  " + led.Reason + " , "
						} else {
							nb = "Leader Project " + led.LeaderName + " had been Approved leave of " + led.DateLeave + " because : " + led.Reason + " , "
						}
					}
				}

				if sttsleader == len(dataResponseLeader) {
					_, ok := addd["note"]
					if !ok {
						addd["note"] = nb
					}
					bd, er := d.File("declinedleader.html", addd)
					if er != nil {
						return d.SetResultInfo(true, er.Error(), nil)
					}
					m.SetBody("text/html", bd)

					d.DelayProcess(5)

					if err := conf.DialAndSend(m); err != nil {
						return d.SetResultInfo(true, err.Error(), nil)
					} else {
						ret.Data = "Successfully Send Mails"
					}
					m.Reset()
				}
			} else {
				// level := r.Session("jobrolelevel").(int)
				for i, dt := range dataRequest.StatusProjectLeader {
					if level == 1 || level == 6 {
						dataRequest.StatusProjectLeader[i].StatusRequest = "Approved"
						listDetailLeave := dash.GetLeavebyDateFilterIdTransc(r, dataRequest.Id)
						for _, ireqByDate := range listDetailLeave {
							ireqByDate.StsByLeader = "Approved"
							err := d.Ctx.Save(ireqByDate)
							if err != nil {
								return d.SetResultInfo(true, err.Error(), nil)
							}

						}

						nb = "-"
						_, ok := addd["note"]
						if !ok {
							addd["note"] = nb
						}
						bd, er := d.File("declinedleader.html", addd)
						if er != nil {
							return d.SetResultInfo(true, er.Error(), nil)
						}
						m.SetBody("text/html", bd)

						d.DelayProcess(5)

						if err := conf.DialAndSend(m); err != nil {
							return d.SetResultInfo(true, err.Error(), nil)
						} else {
							ret.Data = "Successfully Send Mails"
						}
						m.Reset()
					} else {
						if dt.StatusRequest == "Declined" {
							sttsleader = sttsleader + 1
							if nb != "" {
								nb = nb + "Leader Project " + dt.Name + " had been Declined leave because : " + dt.Reason + " , "
							} else {
								nb = "Leader Project " + dt.Name + " had been Declined leave because : " + dt.Reason + " , "
							}
						} else if dt.StatusRequest == "Approved" {
							sttsleader = sttsleader + 1
							if nb != "" {
								nb = nb + "Leader Project " + dt.Name + " had been Approved leave because : " + dt.Reason + " , "
							} else {
								nb = "Leader Project " + dt.Name + " had been Approved leave because : " + dt.Reason + " , "
							}
						}

						if sttsleader == len(dataRequest.StatusProjectLeader) {
							_, ok := addd["note"]
							if !ok {
								addd["note"] = nb
							}
							bd, er := d.File("declinedleader.html", addd)
							if er != nil {
								return d.SetResultInfo(true, er.Error(), nil)
							}
							m.SetBody("text/html", bd)

							d.DelayProcess(5)

							if err := conf.DialAndSend(m); err != nil {
								return d.SetResultInfo(true, err.Error(), nil)
							} else {
								ret.Data = "Successfully Send Mails"
							}
							m.Reset()
						}
					}

				}

				if level == 1 || level == 6 {
					err := d.Ctx.Save(dataRequest)
					if err != nil {
						return d.SetResultInfo(true, err.Error(), nil)
					}
				}

			}

		} else {
			mailsubj := tk.Sprintf("%v", "EMERGENCY LEAVE from "+dataRequest.Name)
			m.SetHeader("Subject", mailsubj)
			// _, ok := addd["note"]
			// if !ok {
			// 	addd["note"] = nb
			// }
			bd, er := d.File("emergencyManager.html", addd)
			if er != nil {
				return d.SetResultInfo(true, er.Error(), nil)
			}
			m.SetBody("text/html", bd)

			d.DelayProcess(5)

			if err := conf.DialAndSend(m); err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			} else {
				ret.Data = "Successfully Send Mails"
			}
			m.Reset()
		}

	}

	// }

	return ret
}

func (c *MailController) GetDataHR(k *knot.WebContext) ([]*HRDAdminModel, error) {
	k.Config.OutputType = knot.OutputJson
	dataHR := make([]*HRDAdminModel, 0)
	query := tk.M{}
	crs, err := c.Ctx.Find(NewHRDAdminModel(), query)
	if err != nil {
		return dataHR, err
	}
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}
	// defer crs.Close()
	err = crs.Fetch(&dataHR, 0, false)
	if err != nil {
		return dataHR, err
	}

	return dataHR, nil
}

func (c *MailController) EmergencyLeave(r *knot.WebContext, p *RequestLeaveModel, level int) error {
	// d.SendMailUser(p.Id, )
	user := []string{p.StatusManagerProject.Email}
	for _, ld := range p.StatusProjectLeader {
		user = append(user, ld.Email)
	}
	for _, ba := range p.StatusBusinesAnalyst {
		user = append(user, ba.Email)
	}
	for _, ld := range p.StatusProjectLeader {
		if ld.Email != p.Email {
			user = append(user, ld.Email)
		}
	}
	if p.Location == "Indonesia" {
		dataHR, err := c.GetDataHR(r)
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
	c.RequestLeaveOnDateV2(p, tk.ToString(r.Session("userid")))
	c.SendMailManager(r, tk.ToString(r.Session("userid")), p.Id, level)

	return nil
}

func (d *MailController) SendMailBusinesAnalist(r *knot.WebContext, ID string) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	// fmt.Println("--------- masuk")
	dataRequest := d.getDataLeave(ID)
	dataResponseLeader := dataRequest.StatusProjectLeader

	// fmt.Println("------------ leader", dataResponseLeader)

	for _, dt := range dataRequest.StatusBusinesAnalyst {
		sUrl := r.Server.Address

		uriApprove := sUrl + "/mail/responseba"

		urlApprove, _ := http.NewRequest("GET", uriApprove, nil)
		qApprove := urlApprove.URL.Query()
		qApprove.Add("IdRequest", string(dataRequest.Id))
		qApprove.Add("ApproveBA", "yes")
		// fmt.Println("----------- id leader", dt.IdEmp)
		qApprove.Add("IdBA", dt.IdEmp)
		urlApprove.URL.RawQuery = qApprove.Encode()

		uriDecline := sUrl + "/mail/responsebadecline"

		urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
		qDecline := urlDecline.URL.Query()
		// idDeclinetext := []byte(dataRequest.Id)
		// idDeclineRequest, _ := encrypt(idDeclinetext)
		// qDecline.Add("key", string(idDeclineRequest))
		qDecline.Add("declineBA", "no")
		qDecline.Add("IdRequest", string(dataRequest.Id))
		// fmt.Println("----------- id leader", dt.IdEmp)
		qDecline.Add("IdBA", dt.IdEmp)
		urlDecline.URL.RawQuery = qDecline.Encode()

		// fmt.Println("------------------", addd)
		conf := gomail.NewPlainDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")
		mailsubj := tk.Sprintf("%v", "BA Project Leave Request Attachment")
		m := gomail.NewMessage()

		m.SetHeader("From", "admin.support@creativelab.com")
		m.SetHeader("To", dt.Email)
		m.SetHeader("Subject", mailsubj)

		addd := map[string]string{"to": dt.Name, "namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String()}

		nb := ""
		stts := 0
		for _, ld := range dataResponseLeader {

			if ld.IdEmp == dt.IdEmp {
				return "direcly mail Project Manager"
			}

			if ld.StatusRequest == "Declined" {
				stts = stts + 1

				if nb != "" {
					nb = nb + "Leader of Project " + ld.ProjectName + " had been " + ld.StatusRequest + " request because :" + ld.Reason + ","
				} else {
					nb = "Leader of Project " + ld.ProjectName + " had been " + ld.StatusRequest + " request because :" + ld.Reason + ","
				}

			} else if ld.StatusRequest == "Approved" {
				stts = stts + 1
			}

			// fmt.Println("--------------- stts", stts)

			if stts == len(dataResponseLeader) {
				_, ok := addd["note"]
				if !ok {
					addd["note"] = nb
				}
				bd, er := d.File("declinedleader.html", addd)
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

		}

	}

	return ret
}

func (d *MailController) SendMailLeaveDetails(r *knot.WebContext, to []string, name string, reason string, req *RequestLeaveModel) interface{} {
	addd := dataR{}
	detail := dataDetails{}
	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	to = append(to, hrd)
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	// dataRequest := d.getDataLeave(ID)
	str := ""
	if req.IsEmergency == true {
		str = "emergency leave"
	} else {
		str = "leave"
	}

	mailsubj := tk.Sprintf("%v", "Result "+str+" by details")
	m := gomail.NewMessage()

	addd.NameRemote = name
	addd.Reason = reason
	for _, det := range req.DetailsLeave {
		detail.Date = det.DateLeave
		detail.Name = det.LeaderName
		if det.IsApproved {
			detail.Result = "Approved"
		} else {
			detail.Result = "Declined"
		}
		detail.Reason = det.Reason

		addd.Details = append(addd.Details, detail)

	}

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)

	bd, er := FileLeaveTemplate("leavedetailmanager.html", addd)

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

func (d *MailController) SendMailLeader(r *knot.WebContext, sendReq *RequestLeaveModel, level int) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	confUrl := helper.ReadConfig()
	// sUrl := conf.GetString("BaseUrlEmail")
	md := 1
	userid := r.Session("userid")

	tk.Println("------------ userid ", userid)

	dataRequestByDate := d.getDataLeaveByDate(sendReq.Id)
	noPaidLeave := 0
	paidLeave := 0
	for _, requestbydate := range dataRequestByDate {
		if requestbydate.IsPaidLeave {
			paidLeave++
		} else {
			noPaidLeave++
		}
	}

	if len(sendReq.StatusProjectLeader) == 1 && sendReq.StatusProjectLeader[0].UserId == "" {
		i := 0
		md = i + md
		sendReq.StatusProjectLeader[i].StatusRequest = "Approved"
		dash := DashboardController(*d)
		dash.SetHistoryLeave(r, tk.ToString(userid), sendReq.Id, sendReq.LeaveFrom, sendReq.LeaveTo, "Create request Leave", "Pending", sendReq)

		err := d.Ctx.Save(sendReq)
		if err != nil {
			return err
		}

		tk.Println("--------------------- masuk", md)
		if md == len(sendReq.StatusProjectLeader) {
			d.SendMailManager(r, tk.ToString(userid), sendReq.Id, level)
		}

		desc := "Request Approved by Leader"
		emptyremote := new(RemoteModel)
		service := services.LogService{
			sendReq,
			emptyremote,
			"leave",
		}

		log := tk.M{}
		log.Set("Status", sendReq.ResultRequest)
		log.Set("Desc", desc)
		log.Set("NameLogBy", sendReq.StatusProjectLeader[0].Name)
		log.Set("EmailNameLogBy", sendReq.StatusProjectLeader[0].Email)
		err = service.ApproveDeclineLog(log)
		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}
	} else {
		for i, dt := range sendReq.StatusProjectLeader {
			if sendReq.UserId != dt.UserId && sendReq.UserId != "" {
				sUrl := confUrl.GetString("BaseUrlEmail")

				uri := sUrl + "/mail/responseleader"
				urlApprove, _ := http.NewRequest("GET", uri, nil)
				qApprove := urlApprove.URL.Query()

				// var App *ParameterURLModel
				var dec = new(ParameterURLModel)
				dec.Level = level
				dec.UserId = tk.ToString(userid)
				dec.IdRequest = string(sendReq.Id)
				dec.IdLeader = dt.IdEmp
				dec.ApproveLeader = "yes"
				paramApp, _ := json.Marshal(dec)

				// fmt.Println("------------userid", paramApp)

				qApprove.Add("param", GCMEncrypter(string(paramApp)))

				urlApprove.URL.RawQuery = qApprove.Encode()

				uriDecline := sUrl + "/mail/responseleaderdecline"
				urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
				qDecline := urlDecline.URL.Query()

				dec.ApproveLeader = "no"
				paramDec, _ := json.Marshal(dec)
				qDecline.Add("param", GCMEncrypter(string(paramDec)))

				urlDecline.URL.RawQuery = qDecline.Encode()
				tk.Println(urlApprove.URL.String())

				addd := map[string]string{"to": dt.Name, "namerequest": sendReq.Name, "reason": sendReq.Reason, "leaveFrom": sendReq.LeaveFrom, "leaveTo": sendReq.LeaveTo, "noOfDays": strconv.Itoa(sendReq.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": sendReq.DateCreateLeave, "noPaidLeave": strconv.Itoa(noPaidLeave), "paidLeave": strconv.Itoa(paidLeave)}

				bd, err := d.File("coba.html", addd)

				if err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				}
				// conf := gomail.NewDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")
				// s, err := conf.Dial()
				// if err != nil {
				// 	d.SetResultInfo(true, err.Error(), nil)
				// }
				mailsubj := tk.Sprintf("%v", "Leave Request from "+sendReq.Name)
				// mailmsg := tk.Sprintf("%v", "<button class='btn btn-sm btn-flat success'>OK</button>")
				m := gomail.NewMessage()

				m.SetHeader("From", emailAddress)
				m.SetHeader("To", dt.Email)
				m.SetHeader("Subject", mailsubj)
				m.SetBody("text/html", bd)
				// m.Attach("d:/site-logo.png")

				d.DelayProcess(5)

				if err = conf.DialAndSend(m); err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				} else {
					ret.Data = "Successfully Send Mails"
				}
				m.Reset()
			} else {
				md = i + md
				sendReq.StatusProjectLeader[i].StatusRequest = "Approved"
				dash := DashboardController(*d)
				dash.SetHistoryLeave(r, tk.ToString(userid), sendReq.Id, sendReq.LeaveFrom, sendReq.LeaveTo, "Create request Leave", "Pending", sendReq)

				err := d.Ctx.Save(sendReq)
				if err != nil {
					return err
				}

				tk.Println("--------------------- masuk", md)
				if md == len(sendReq.StatusProjectLeader) {
					d.SendMailManager(r, tk.ToString(userid), sendReq.Id, level)
				}

				desc := "Request Approved by Leader"
				emptyremote := new(RemoteModel)
				service := services.LogService{
					sendReq,
					emptyremote,
					"leave",
				}

				log := tk.M{}
				log.Set("Status", sendReq.ResultRequest)
				log.Set("Desc", desc)
				log.Set("NameLogBy", dt.Name)
				log.Set("EmailNameLogBy", dt.Email)
				err = service.ApproveDeclineLog(log)
				if err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				}
			}
		}
	}

	return ret
}

func (d *MailController) SendMailLeaderEdited(r *knot.WebContext, sendReq *RequestLeaveModel, level int) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	confUrl := helper.ReadConfig()
	// sUrl := conf.GetString("BaseUrlEmail")
	md := 1
	userid := r.Session("userid")

	tk.Println("------------ level---- ", level)

	for i, dt := range sendReq.StatusProjectLeader {
		if sendReq.UserId != dt.UserId {
			sUrl := confUrl.GetString("BaseUrlEmail")

			uri := sUrl + "/mail/responseleader"
			urlApprove, _ := http.NewRequest("GET", uri, nil)
			qApprove := urlApprove.URL.Query()

			// var App *ParameterURLModel
			var dec = new(ParameterURLModel)
			dec.Level = level
			dec.UserId = tk.ToString(userid)
			dec.IdRequest = string(sendReq.Id)
			dec.IdLeader = dt.IdEmp
			dec.ApproveLeader = "yes"
			paramApp, _ := json.Marshal(dec)

			// fmt.Println("------------userid", paramApp)

			qApprove.Add("param", GCMEncrypter(string(paramApp)))

			urlApprove.URL.RawQuery = qApprove.Encode()

			uriDecline := sUrl + "/mail/responseleaderdecline"
			urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
			qDecline := urlDecline.URL.Query()

			dec.ApproveLeader = "no"
			paramDec, _ := json.Marshal(dec)
			qDecline.Add("param", GCMEncrypter(string(paramDec)))

			urlDecline.URL.RawQuery = qDecline.Encode()
			tk.Println(urlApprove.URL.String())

			addd := map[string]string{"to": dt.Name, "namerequest": sendReq.Name, "reason": sendReq.Reason, "leaveFrom": sendReq.LeaveFrom, "leaveTo": sendReq.LeaveTo, "noOfDays": strconv.Itoa(sendReq.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": sendReq.DateCreateLeave}

			bd, err := d.File("editedleave.html", addd)

			if err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			}
			// conf := gomail.NewDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")
			// s, err := conf.Dial()
			// if err != nil {
			// 	d.SetResultInfo(true, err.Error(), nil)
			// }
			mailsubj := tk.Sprintf("%v", "Leave Request from "+sendReq.Name)
			// mailmsg := tk.Sprintf("%v", "<button class='btn btn-sm btn-flat success'>OK</button>")
			m := gomail.NewMessage()

			m.SetHeader("From", emailAddress)
			m.SetHeader("To", dt.Email)
			m.SetHeader("Subject", mailsubj)
			m.SetBody("text/html", bd)
			// m.Attach("d:/site-logo.png")

			d.DelayProcess(5)

			if err = conf.DialAndSend(m); err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			} else {
				ret.Data = "Successfully Send Mails"
			}
			m.Reset()
		} else {
			md = i + md
			sendReq.StatusProjectLeader[i].StatusRequest = "Approved"
			dash := DashboardController(*d)
			dash.SetHistoryLeave(r, tk.ToString(userid), sendReq.Id, sendReq.LeaveFrom, sendReq.LeaveTo, "Create request Leave", "Pending", sendReq)

			err := d.Ctx.Save(sendReq)
			if err != nil {
				return err
			}

			tk.Println("--------------------- masuk", md)
			if md == len(sendReq.StatusProjectLeader) {
				d.SendMailManager(r, tk.ToString(userid), sendReq.Id, level)
			}
		}
	}

	return ret
}

// func (d *MailController) SendMailManagerEdited(r *knot.WebContext, sendReq *RequestLeaveModel, level int) interface{} {
// 	r.Config.OutputType = knot.OutputJson
// 	ret := ResultInfo{}
// 	conf, emailAddress := d.EmailConfiguration()
// 	confUrl := helper.ReadConfig()
// 	// sUrl := conf.GetString("BaseUrlEmail")
// 	// md := 1
// 	userid := r.Session("userid")

// 	tk.Println("------------ level---- ", level)

// 	for _, dt := range sendReq.BranchManager {
// 		// if sendReq.UserId != dt.UserId {
// 		sUrl := confUrl.GetString("BaseUrlEmail")

// 		uri := sUrl + "/mail/responsemanager"
// 		urlApprove, _ := http.NewRequest("GET", uri, nil)
// 		qApprove := urlApprove.URL.Query()

// 		// var App *ParameterURLModel
// 		var dec = new(ParameterURLModel)
// 		dec.Level = level
// 		dec.UserId = tk.ToString(userid)
// 		dec.IdRequest = string(sendReq.Id)
// 		dec.IdLeader = dt.IdEmp
// 		dec.ApproveLeader = "yes"
// 		paramApp, _ := json.Marshal(dec)

// 		// fmt.Println("------------userid", paramApp)

// 		qApprove.Add("param", GCMEncrypter(string(paramApp)))

// 		urlApprove.URL.RawQuery = qApprove.Encode()

// 		uriDecline := sUrl + "/mail/responseleaderdecline"
// 		urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
// 		qDecline := urlDecline.URL.Query()

// 		dec.ApproveLeader = "no"
// 		paramDec, _ := json.Marshal(dec)
// 		qDecline.Add("param", GCMEncrypter(string(paramDec)))

// 		urlDecline.URL.RawQuery = qDecline.Encode()
// 		tk.Println(urlApprove.URL.String())

// 		addd := map[string]string{"to": dt.Name, "namerequest": sendReq.Name, "reason": sendReq.Reason, "leaveFrom": sendReq.LeaveFrom, "leaveTo": sendReq.LeaveTo, "noOfDays": strconv.Itoa(sendReq.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": sendReq.DateCreateLeave}

// 		bd, err := d.File("editedleave.html", addd)

// 		if err != nil {
// 			return d.SetResultInfo(true, err.Error(), nil)
// 		}
// 		// conf := gomail.NewDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")
// 		// s, err := conf.Dial()
// 		// if err != nil {
// 		// 	d.SetResultInfo(true, err.Error(), nil)
// 		// }
// 		mailsubj := tk.Sprintf("%v", "Leave Request from "+sendReq.Name)
// 		// mailmsg := tk.Sprintf("%v", "<button class='btn btn-sm btn-flat success'>OK</button>")
// 		m := gomail.NewMessage()

// 		m.SetHeader("From", emailAddress)
// 		m.SetHeader("To", dt.Email)
// 		m.SetHeader("Subject", mailsubj)
// 		m.SetBody("text/html", bd)
// 		// m.Attach("d:/site-logo.png")

// 		d.DelayProcess(5)

// 		if err = conf.DialAndSend(m); err != nil {
// 			return d.SetResultInfo(true, err.Error(), nil)
// 		} else {
// 			ret.Data = "Successfully Send Mails"
// 		}
// 		m.Reset()

// 	}

// 	return ret
// }

func (d *MailController) SendMailManagerEdited(r *knot.WebContext, userid string, ID string, level int) interface{} {
	// tk.Println("---------------- masuk sendMailManager", ID)
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	urlConf := helper.ReadConfig()

	dataRequest := d.getDataLeave(ID)

	dataResponseLeader := dataRequest.DetailsLeave
	// level := r.Session("jobrolelevel").(int)
	dt := dataRequest.StatusManagerProject
	sUrl := urlConf.GetString("BaseUrlEmail")
	dash := DashboardController(*d)

	for _, dtl := range dataRequest.BranchManager {
		uriApprove := sUrl + "/mail/responsemanager"
		var dec = new(ParameterURLManagerModel)
		dec.UserId = userid
		dec.IdRequest = string(dataRequest.Id)
		dec.UserIdManager = dtl.UserId
		dec.IdManager = dt.IdEmp
		dec.ApproveManager = "yes"
		paramApp, _ := json.Marshal(dec)

		urlApprove, _ := http.NewRequest("GET", uriApprove, nil)
		qApprove := urlApprove.URL.Query()
		qApprove.Add("param", GCMEncrypter(string(paramApp)))
		urlApprove.URL.RawQuery = qApprove.Encode()

		uriDecline := sUrl + "/mail/responsemanagerdecline"

		dec.ApproveManager = "no"
		paramDec, _ := json.Marshal(dec)

		// fmt.Println("------------ leader masuk send manager", dec.UserId)

		urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
		qDecline := urlDecline.URL.Query()
		qDecline.Add("param", GCMEncrypter(string(paramDec)))
		urlDecline.URL.RawQuery = qDecline.Encode()

		conf, emailAddress := d.EmailConfiguration()

		m := gomail.NewMessage()

		// fmt.Println("----------- dt.Email", dt.Email)

		m.SetHeader("From", emailAddress)
		m.SetHeader("To", dtl.Email)

		addd := map[string]string{"to": dtl.Name, "namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": dataRequest.DateCreateLeave}

		// if !dataRequest.IsEmergency {
		mailsubj := tk.Sprintf("%v", "Leave Request from "+dataRequest.Name)
		m.SetHeader("Subject", mailsubj)

		nb := ""
		sttsleader := 0
		if len(dataResponseLeader) > 0 {
			for _, led := range dataResponseLeader {
				if led.IsApproved == false {
					sttsleader = sttsleader + 1
					if nb != "" {
						nb = nb + "Leader Project " + led.LeaderName + " had been Declined leave of " + led.DateLeave + " because : " + led.Reason + " , "
					} else {
						nb = "Leader Project " + led.LeaderName + " had been Declined leave of  " + led.DateLeave + " because : " + led.Reason + " , "
					}
				} else if led.IsApproved == true {
					sttsleader = sttsleader + 1
					if nb != "" {
						nb = nb + "Leader Project " + led.LeaderName + " had been Approved leave of " + led.DateLeave + " because :  " + led.Reason + " , "
					} else {
						nb = "Leader Project " + led.LeaderName + " had been Approved leave of " + led.DateLeave + " because : " + led.Reason + " , "
					}
				}
			}

			if sttsleader == len(dataResponseLeader) {
				_, ok := addd["note"]
				if !ok {
					addd["note"] = nb
				}
				bd, er := d.File("leaveeditmanager.html", addd)
				if er != nil {
					return d.SetResultInfo(true, er.Error(), nil)
				}
				m.SetBody("text/html", bd)

				d.DelayProcess(5)

				if err := conf.DialAndSend(m); err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				} else {
					ret.Data = "Successfully Send Mails"
				}
				m.Reset()
			}
		} else {
			// level := r.Session("jobrolelevel").(int)
			for i, dt := range dataRequest.StatusProjectLeader {
				if level == 1 || level == 6 {
					dataRequest.StatusProjectLeader[i].StatusRequest = "Approved"
					listDetailLeave := dash.GetLeavebyDateFilterIdTransc(r, dataRequest.Id)
					for _, ireqByDate := range listDetailLeave {
						ireqByDate.StsByLeader = "Approved"
						err := d.Ctx.Save(ireqByDate)
						if err != nil {
							return d.SetResultInfo(true, err.Error(), nil)
						}

					}

					nb = "-"
					_, ok := addd["note"]
					if !ok {
						addd["note"] = nb
					}
					bd, er := d.File("leaveeditmanager.html", addd)
					if er != nil {
						return d.SetResultInfo(true, er.Error(), nil)
					}
					m.SetBody("text/html", bd)

					d.DelayProcess(5)

					if err := conf.DialAndSend(m); err != nil {
						return d.SetResultInfo(true, err.Error(), nil)
					} else {
						ret.Data = "Successfully Send Mails"
					}
					m.Reset()
				} else {
					if dt.StatusRequest == "Declined" {
						sttsleader = sttsleader + 1
						if nb != "" {
							nb = nb + "Leader Project " + dt.Name + " had been Declined leave because : " + dt.Reason + " , "
						} else {
							nb = "Leader Project " + dt.Name + " had been Declined leave because : " + dt.Reason + " , "
						}
					} else if dt.StatusRequest == "Approved" {
						sttsleader = sttsleader + 1
						if nb != "" {
							nb = nb + "Leader Project " + dt.Name + " had been Approved leave because : " + dt.Reason + " , "
						} else {
							nb = "Leader Project " + dt.Name + " had been Approved leave because : " + dt.Reason + " , "
						}
					}

					if sttsleader == len(dataRequest.StatusProjectLeader) {
						_, ok := addd["note"]
						if !ok {
							addd["note"] = nb
						}
						bd, er := d.File("leaveeditmanager.html", addd)
						if er != nil {
							return d.SetResultInfo(true, er.Error(), nil)
						}
						m.SetBody("text/html", bd)

						d.DelayProcess(5)

						if err := conf.DialAndSend(m); err != nil {
							return d.SetResultInfo(true, err.Error(), nil)
						} else {
							ret.Data = "Successfully Send Mails"
						}
						m.Reset()
					}
				}

			}

			if level == 1 || level == 6 {
				err := d.Ctx.Save(dataRequest)
				if err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				}
			}

		}

		// } else {
		// 	mailsubj := tk.Sprintf("%v", "EMERGENCY LEAVE from "+dataRequest.Name)
		// 	m.SetHeader("Subject", mailsubj)
		// 	// _, ok := addd["note"]
		// 	// if !ok {
		// 	// 	addd["note"] = nb
		// 	// }
		// 	bd, er := d.File("emergencyManager.html", addd)
		// 	if er != nil {
		// 		return d.SetResultInfo(true, er.Error(), nil)
		// 	}
		// 	m.SetBody("text/html", bd)

		// 	d.DelayProcess(5)

		// 	if err := conf.DialAndSend(m); err != nil {
		// 		return d.SetResultInfo(true, err.Error(), nil)
		// 	} else {
		// 		ret.Data = "Successfully Send Mails"
		// 	}
		// 	m.Reset()
		// }

	}

	// }

	return ret
}

func (d *MailController) SendMailLeaderCancel(r *knot.WebContext, sendReq *RequestLeaveModel) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	confUrl := helper.ReadConfig()
	// sUrl := conf.GetString("BaseUrlEmail")
	md := 1
	userid := r.Session("userid")
	for i, dt := range sendReq.StatusProjectLeader {
		if sendReq.UserId != dt.UserId {
			sUrl := confUrl.GetString("BaseUrlEmail")

			uri := sUrl + "/mail/responseleader"
			urlApprove, _ := http.NewRequest("GET", uri, nil)
			qApprove := urlApprove.URL.Query()

			// var App *ParameterURLModel
			var dec = new(ParameterURLModel)
			dec.UserId = tk.ToString(userid)
			dec.IdRequest = string(sendReq.Id)
			dec.IdLeader = dt.IdEmp
			dec.ApproveLeader = "yes"
			paramApp, _ := json.Marshal(dec)

			// fmt.Println("------------userid", paramApp)

			qApprove.Add("param", GCMEncrypter(string(paramApp)))

			urlApprove.URL.RawQuery = qApprove.Encode()

			uriDecline := sUrl + "/mail/responseleaderdecline"
			urlDecline, _ := http.NewRequest("GET", uriDecline, nil)
			qDecline := urlDecline.URL.Query()

			dec.ApproveLeader = "no"
			paramDec, _ := json.Marshal(dec)
			qDecline.Add("param", GCMEncrypter(string(paramDec)))

			urlDecline.URL.RawQuery = qDecline.Encode()
			tk.Println(urlApprove.URL.String())

			addd := map[string]string{"to": dt.Name, "namerequest": sendReq.Name, "reason": sendReq.Reason, "leaveFrom": sendReq.LeaveFrom, "leaveTo": sendReq.LeaveTo, "noOfDays": strconv.Itoa(sendReq.NoOfDays), "urlApprove": urlApprove.URL.String(), "urlDecline": urlDecline.URL.String(), "DateCreate": sendReq.DateCreateLeave}

			bd, err := d.File("leavedatecancel.html", addd)

			if err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			}
			// conf := gomail.NewDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")
			// s, err := conf.Dial()
			// if err != nil {
			// 	d.SetResultInfo(true, err.Error(), nil)
			// }
			mailsubj := tk.Sprintf("%v", "Leader ECLEAVE Gomail Service")
			// mailmsg := tk.Sprintf("%v", "<button class='btn btn-sm btn-flat success'>OK</button>")
			m := gomail.NewMessage()

			m.SetHeader("From", emailAddress)
			m.SetHeader("To", dt.Email)
			m.SetHeader("Subject", mailsubj)
			m.SetBody("text/html", bd)
			// m.Attach("d:/site-logo.png")

			d.DelayProcess(5)

			if err = conf.DialAndSend(m); err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			} else {
				ret.Data = "Successfully Send Mails"
			}
			m.Reset()
		} else {
			md = i + md
			sendReq.StatusProjectLeader[i].StatusRequest = "Approved"
			dash := DashboardController(*d)
			dash.SetHistoryLeave(r, tk.ToString(userid), sendReq.Id, sendReq.LeaveFrom, sendReq.LeaveTo, "Create request Leave", "Pending", sendReq)

			err := d.Ctx.Save(sendReq)
			if err != nil {
				return err
			}

			level := r.Session("jobrolelevel").(int)

			tk.Println("--------------------- masuk", md)
			if md == len(sendReq.StatusProjectLeader) {
				d.SendMailManager(r, tk.ToString(userid), sendReq.Id, level)
			}
		}
	}

	return ret
}

func (d *MailController) SendMailCancelLeave(r *knot.WebContext, sendReq *RequestLeaveModel, level int, nameSes string) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	// confUrl := helper.ReadConfig()
	// sUrl := conf.GetString("BaseUrlEmail")

	// var App *ParameterURLModel

	addd := map[string]string{"to": sendReq.Name, "namerequest": sendReq.Name, "reason": sendReq.Reason, "leaveFrom": sendReq.LeaveFrom, "leaveTo": sendReq.LeaveTo, "noOfDays": strconv.Itoa(sendReq.NoOfDays), "DateCreate": sendReq.DateCreateLeave}

	if level == 2 {
		addd["leaderManager"] = "Leader"
		addd["declinedReason"] = " "
	} else if level == 1 || level == 6 {
		addd["leaderManager"] = sendReq.StatusManagerProject.Name
		addd["declinedReason"] = sendReq.StatusManagerProject.Reason
	} else {
		addd["leaderManager"] = "Admin"
		addd["declinedReason"] = " "
	}

	bd, err := d.File("leavecancel.html", addd)

	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	mailsubj := tk.Sprintf("%v", "Your leave request has been turn down")
	// mailmsg := tk.Sprintf("%v", "<button class='btn btn-sm btn-flat success'>OK</button>")
	m := gomail.NewMessage()

	m.SetHeader("From", emailAddress)
	m.SetHeader("To", sendReq.Email)
	m.SetHeader("Subject", mailsubj)
	m.SetBody("text/html", bd)
	// m.Attach("d:/site-logo.png")

	d.DelayProcess(5)

	if err = conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()

	return ret
}

var (
	wd = func() string {
		d, _ := os.Getwd()
		return d + "/"
	}()
)

func (d *MailController) ResponseManager(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *MailController) ResetPassword(r *knot.WebContext) interface{} {
	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}
func (d *MailController) ResponseLeader(r *knot.WebContext) interface{} {

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *MailController) ResponseManagerDecline(r *knot.WebContext) interface{} {

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *MailController) ResponseLeaderDecline(r *knot.WebContext) interface{} {

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *MailController) ResponseBADecline(r *knot.WebContext) interface{} {

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (d *MailController) ResponseBA(r *knot.WebContext) interface{} {

	r.Config.NoLog = true
	r.Config.OutputType = knot.OutputTemplate
	r.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func rangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}

func (c *MailController) RequestLeaveOnDate(m *RequestLeaveModel, userid string) interface{} {
	// r.Config.OutputType = knot.OutputJson
	lvDate := new(AprovalRequestLeaveModel)
	start, _ := time.Parse("2006-1-2", m.LeaveFrom)
	end, _ := time.Parse("2006-1-2", m.LeaveTo)
	// fmt.Println(end)
	for d := rangeDate(start, end); ; {
		dt := d()
		if dt.IsZero() {
			break
		}

		// then, _ := time.Parse("2006-1-2", tk.ToString(dt))

		onday := dt.Weekday()

		if onday.String() == "Sunday" {

		} else if onday.String() == "Saturday" {

		} else {
			// tk.Println("-----------------", onday.String())
			// tk.Println("-----------------", dt)
			var dbFilter []*db.Filter

			dbFilter = append(dbFilter, db.Eq("userid", userid))
			dbFilter = append(dbFilter, db.Eq("dateleave", dt.Format("2006-01-02")))

			query := tk.M{}

			if len(dbFilter) > 0 {
				query.Set("where", db.And(dbFilter...))
			}

			data := []*RequestLeaveModel{}

			crsData, errData := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
			if crsData != nil {
				defer crsData.Close()
			} else {
				return c.SetResultInfo(true, "error in query", nil)
			}
			// defer crsData.Close()
			if errData != nil {
				return c.SetResultInfo(true, errData.Error(), nil)
			}
			errData = crsData.Fetch(&data, 0, false)

			// fmt.Println("------------- data now", data)

			if len(data) == 0 {
				lvDate.Id = bson.NewObjectId().Hex()
				lvDate.IdRequest = m.Id
				lvDate.Name = m.Name
				lvDate.EmpId = m.EmpId
				lvDate.Designation = m.Designation
				lvDate.Location = m.Location
				lvDate.Departement = m.Departement
				lvDate.Reason = m.Reason
				lvDate.Email = m.Email
				lvDate.Address = m.Address
				lvDate.Contact = m.Contact
				lvDate.Project = m.Project
				lvDate.YearLeave = m.YearLeave
				lvDate.PublicLeave = m.PublicLeave
				// fmt.Println("--------------- emergency", m.IsEmergency)
				lvDate.IsEmergency = m.IsEmergency
				ondate := dt.Format("2006-01-02")
				// fmt.Println("Year   :", ondate)
				lvDate.DateLeave = ondate
				lvDate.UserId = userid
				err := c.Ctx.Save(lvDate)
				if err != nil {
					return err
				}
			}
		}

	}

	return ""
}
func (c *MailController) RequestLeaveOnDateV2(m *RequestLeaveModel, userid string) interface{} {
	// r.Config.OutputType = knot.OutputJson
	lvDate := new(AprovalRequestLeaveModel)
	// start, _ := time.Parse("2006-1-2", m.LeaveFrom)
	// end, _ := time.Parse("2006-1-2", m.LeaveTo)
	// // fmt.Println(end)
	// for d := rangeDate(start, end); ; {
	// 	dt := d()
	// 	if dt.IsZero() {
	// 		break
	// 	}

	//Get NotCutLeave ======================================================
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "stsbymanager": "Approved"}})
	pipe = append(pipe, tk.M{"$match": tk.M{"userid": userid, "iscutoff": false}})
	csr, err := c.Ctx.Connection.NewQuery().Command("pipe", pipe).From("requestLeaveByDate").Cursor(nil)
	if csr != nil {
		defer csr.Close()
	} else {
	}
	if err != nil {
	}
	dataNotCutLeave := []AprovalRequestLeaveModel{}
	if err := csr.Fetch(&dataNotCutLeave, 0, false); err != nil {
	}

	//Get YearLeave User ======================================================
	query := tk.M{}
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("_id", strings.TrimSpace(userid)))
	users := []SysUserModel{}
	query.Set("where", db.And(dbFilter...))
	crs, _ := c.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	}
	err = crs.Fetch(&users, 0, false)
	if err != nil {
	}
	yearLeave := 0
	if len(users) > 0 {
		yearLeave = users[0].YearLeave
	}
	tempY := yearLeave - len(dataNotCutLeave)
	//========================================================================
	for _, d := range m.LeaveDateList {
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}

		// onday := dt.Weekday()

		// if onday.String() == "Sunday" || onday.String() == "Saturday" {

		// } else {

		var dbFilter []*db.Filter

		dbFilter = append(dbFilter, db.Eq("userid", userid))
		dbFilter = append(dbFilter, db.Eq("isdelete", false))
		dbFilter = append(dbFilter, db.Eq("stsbymanager", "Approved"))
		dbFilter = append(dbFilter, db.Eq("stsbymanager", "Pending"))
		dbFilter = append(dbFilter, db.Eq("dateleave", dt.Format("2006-01-02")))

		query := tk.M{}

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		data := []*AprovalRequestLeaveModel{}

		crsData, errData := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
		if crsData != nil {
			defer crsData.Close()
		} else {
			return c.SetResultInfo(true, "error in query", nil)
		}
		if errData != nil {
			return c.SetResultInfo(true, errData.Error(), nil)
		}
		errData = crsData.Fetch(&data, 0, false)

		if len(data) == 0 {
			lvDate.Id = bson.NewObjectId().Hex()
			lvDate.IdRequest = m.Id
			lvDate.Name = m.Name
			lvDate.EmpId = m.EmpId
			lvDate.Designation = m.Designation
			lvDate.Location = m.Location
			lvDate.Departement = m.Departement
			lvDate.Reason = m.Reason
			lvDate.Email = m.Email
			lvDate.Address = m.Address
			lvDate.Contact = m.Contact
			lvDate.Project = m.Project
			lvDate.YearLeave = m.YearLeave
			lvDate.PublicLeave = m.PublicLeave
			lvDate.IsEmergency = m.IsEmergency
			ondate := dt.Format("2006-01-02")
			lvDate.DateLeave = ondate
			valYear, valMonth, valDay := dt.Date()
			lvDate.DayVal = valDay
			lvDate.MonthVal = int(valMonth)
			lvDate.YearVal = valYear
			lvDate.UserId = userid
			if m.IsSpecials {
				lvDate.IsCutOff = true
			} else {
				lvDate.IsCutOff = false
			}
			lvDate.IsReset = false
			lvDate.StsByLeader = m.StatusProjectLeader[0].StatusRequest
			lvDate.StsByManager = m.StatusManagerProject.StatusRequest
			if m.StatusManagerProject.StatusRequest == "Declined" {
				tk.Println("--------------- masuk sini ", m.StatusManagerProject.StatusRequest)
				lvDate.IsDelete = true
			} else if m.StatusManagerProject.StatusRequest == "Approved" {
				lvDate.IsDelete = false
			} else {
				lvDate.IsDelete = false
			}

			//Add Flag IsPaidLeave ========================
			if !m.IsSpecials {
				if tempY > 0 {
					lvDate.IsPaidLeave = false
				} else {
					lvDate.IsPaidLeave = true
				}
			} else {
				lvDate.IsPaidLeave = false
			}
			tempY = tempY - 1
			//=============================================

			err := c.Ctx.Save(lvDate)
			if err != nil {
				return err
			}
		}
		// }

	}

	return ""
}

func (d *MailController) CheckLeaveByDate(r *knot.WebContext, idrequest string) (interface{}, error) {
	r.Config.OutputType = knot.OutputJson
	dataByDate := []AprovalRequestLeaveModel{}

	// var filter []*db.Filter
	// query := tk.M{}
	pipe := []tk.M{}

	pipe = append(pipe, tk.M{"$match": tk.M{"_id": idrequest, "stsbymanager": tk.M{"$ne": "Pending"}}})
	pipe = append(pipe, tk.M{"$match": tk.M{"stsbymanager": tk.M{"$ne": "Pending"}}})

	crs, err := d.Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeaveByDate").
		Cursor(nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// fmt.Println("--------------", dataByDate)

	err = crs.Fetch(&dataByDate, 0, false)

	if err != nil {
		return nil, err
	}

	if len(dataByDate) > 0 {
		// for _, each := range dataByDate {

		// 	err = d.Ctx.Delete(&each)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }

		return "data already exist", nil
	}

	return nil, nil
}

func (d *MailController) ResponseApproveManager(r *knot.WebContext) interface{} {
	// tk.Println("------------- masuk")
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

	payload := new(ParameterURLManagerModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), payload)

	dataLeave := []*RequestLeaveModel{}
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("_id", payload.IdRequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := d.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	tk.Println("----------- dataleave ", payload.IdRequest)
	resp, sts := d.CheckResponse(r, true, dataLeave[0], payload.IdManager)

	if resp == true {
		return d.SetResultInfo(true, "You already "+sts+" request", nil)
	}

	data := dataLeave[0].StatusManagerProject
	// fmt.Println("----------------- data-request", dataLeave[0].Id)

	check, err := d.CheckLeaveByDate(r, payload.IdRequest)

	// fmt.Println("--------------- payload", check)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("------------- check", check)

	for _, ky := range dataLeave[0].BranchManager {
		if ky.UserId == payload.UserIdManager {
			dataLeave[0].StatusManagerProject.IdEmp = ky.IdEmp
			dataLeave[0].StatusManagerProject.Name = ky.Name
			dataLeave[0].StatusManagerProject.Location = ky.Location
			dataLeave[0].StatusManagerProject.Email = ky.Email
			dataLeave[0].StatusManagerProject.PhoneNumber = ky.PhoneNumber
			dataLeave[0].StatusManagerProject.UserId = ky.UserId
		}
	}

	if check != nil {
		return check
	}
	dash := DashboardController(*d)
	// for i, dt := range dataLeave[0].StatusBusinesAnalyst {
	userdata := dash.GetDataSessionUser(r, dataLeave[0].UserId)[0]
	if data.IdEmp == payload.IdManager {
		if payload.ApproveManager == "yes" {
			tk.Println("------------- masuk1")
			dataLeave[0].StatusManagerProject.StatusRequest = "Approved"
			dataLeave[0].ResultRequest = "Approved"
			dash.SetHistoryLeave(r, payload.UserId, dataLeave[0].Id, dataLeave[0].LeaveFrom, dataLeave[0].LeaveTo, "Your Request has been Approved Manager by "+data.Name, "Approved", dataLeave[0])
			if dataLeave[0].IsSpecials == false {
				tk.Println("------------- masuk3")
				if dataLeave[0].NoOfDays > userdata.YearLeave {
					tk.Println("------------- masuk4")
					//================ take here change
					// return d.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
				} else {
					if dataLeave[0].IsEmergency == true {
						if len(dataLeave[0].LeaveDateList) > 0 {
							userdata.DecYear = userdata.DecYear - 1.0
							userdata.YearLeave = userdata.YearLeave - 1
							// update tmpyear
							// userdata.TmpYear = userdata.TmpYear - dataLeave[0].NoOfDays
							err = d.Ctx.Save(userdata)
						}
					} else {
						// update tmpyear
						userdatatmpyear := dash.GetDataSessionUser(r, dataLeave[0].UserId)[0]
						userdatatmpyear.TmpYear = userdatatmpyear.TmpYear - dataLeave[0].NoOfDays
						err = d.Ctx.Save(userdatatmpyear)
					}

				}

			}

			// d.RequestLeaveOnDate(dataLeave[0], dataLeave[0].UserId)

		} else if payload.ApproveManager == "no" {
			tk.Println("------------- masuk2")
			dataLeave[0].StatusManagerProject.StatusRequest = "Declined"
			dataLeave[0].StatusManagerProject.Reason = p.Reason
			dataLeave[0].ResultRequest = "Declined"
			dash.SetHistoryLeave(r, payload.UserId, dataLeave[0].Id, dataLeave[0].LeaveFrom, dataLeave[0].LeaveTo, "Your Request has been Declined Manager by "+data.Name, "Declined", dataLeave[0])

		}

	}
	// }

	// if dataLeave[0].NoOfDays > userdata.YearLeave {
	// 	// tk.Println("------------------- masuk sini yaaa")
	// 	return d.SetResultInfo(true, "Days of year leave remaining is not enaugh to process", nil)
	// }

	err = d.Ctx.Save(dataLeave[0])
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	user := []string{dataLeave[0].Email}

	if len(dataLeave[0].ProjectManagerList) > 0 {
		for _, mg := range dataLeave[0].ProjectManagerList {
			user = append(user, mg.Email)
		}
	}

	if len(dataLeave[0].StatusBusinesAnalyst) > 0 {
		for _, ba := range dataLeave[0].StatusBusinesAnalyst {
			user = append(user, ba.Email)
		}
	}

	if len(dataLeave[0].StatusProjectLeader) > 0 {
		for _, ld := range dataLeave[0].StatusProjectLeader {
			if ld.Email != "" {
				user = append(user, ld.Email)
			}

		}
	}

	if dataLeave[0].Location == "Indonesia" {
		dataHR, err := d.GetDataHR(r)
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
	tk.Println(dataLeave[0], payload.ApproveManager)
	desc := "Request approved by manager"
	desc2 := "Request approved by PM"
	var stsReq = "Approved"
	if payload.ApproveManager != "yes" {
		stsReq = "Declined"
		desc = "Request declined by manager"
		desc2 = "Request declined by PM"
	}
	// d.SendMailManager(r, payload.UserId, payload.IdRequest)
	listDetailLeave := dash.GetLeavebyDateFilterIdTransc(r, dataLeave[0].Id)
	for i, ireqByDate := range listDetailLeave {
		if dataLeave[0].IsEmergency == true {
			if i == 0 {
				ireqByDate.IsCutOff = true
			}
		}

		ireqByDate.StsByManager = stsReq
		err = d.Ctx.Save(ireqByDate)
		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

	}
	//log
	emptyremote := new(RemoteModel)
	service := services.LogService{
		dataLeave[0],
		emptyremote,
		"leave",
	}
	log := tk.M{}
	log.Set("Status", stsReq)
	log.Set("Desc", desc)
	log.Set("NameLogBy", dataLeave[0].StatusManagerProject.Name)
	log.Set("EmailNameLogBy", dataLeave[0].StatusManagerProject.Email)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	// manager project list
	log2 := log
	log2.Set("Desc", desc2)
	if len(dataLeave[0].ProjectManagerList) > 0 {
		log2.Set("NameLogBy", dataLeave[0].ProjectManagerList[0].Name)
		log2.Set("EmailNameLogBy", dataLeave[0].ProjectManagerList[0].Email)
	}
	err = service.ApproveDeclineLog(log2)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	// d.RequestLeaveOnDateV2(dataLeave[0], payload.UserId)
	d.SendMailUser(r, payload.IdRequest, user)

	notif := NotificationController(*d)

	getnotif := notif.GetDataNotification(r, dataLeave[0].Id)
	getnotif.Notif.ManagerApprove = dataLeave[0].StatusManagerProject.Name
	getnotif.Notif.Status = dataLeave[0].ResultRequest
	getnotif.Notif.StatusApproval = dataLeave[0].ResultRequest
	getnotif.Notif.Description = p.Reason

	notif.InsertNotification(getnotif)
	return dataLeave[0].StatusProjectLeader
}

func (d *MailController) ResponseApproveBA(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	payload := struct {
		IdRequest string
		IdBA      string
		ApproveBA string
		Reason    string
	}{}

	err := r.GetPayload(&payload)
	// fmt.Println("--------------- payload", payload.IdRequest)
	if err != nil {
		return err.Error
	}

	dataLeave := make([]*RequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("_id", payload.IdRequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := d.Ctx.Find(NewRequestLeave(), query)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	// fmt.Println("---------------", dataLeave[0].StatusBusinesAnalyst)

	for i, dt := range dataLeave[0].StatusBusinesAnalyst {
		if dt.IdEmp == payload.IdBA {
			if payload.ApproveBA == "yes" {
				dataLeave[0].StatusBusinesAnalyst[i].StatusRequest = "Approved"
			} else {
				dataLeave[0].StatusBusinesAnalyst[i].StatusRequest = "Declined"
				dataLeave[0].StatusBusinesAnalyst[i].Reason = payload.Reason
			}

		}
	}

	err = d.Ctx.Save(dataLeave[0])
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	// d.SendMailManager(r, payload.IdRequest)

	return dataLeave[0].StatusProjectLeader

}

func (d *MailController) CheckResponse(r *knot.WebContext, isManager bool, data *RequestLeaveModel, idemp string) (bool, string) {
	r.Config.OutputType = knot.OutputJson
	if data.ResultRequest != "Expired" {
		if isManager == true {
			if data.StatusManagerProject.StatusRequest != "Pending" {
				return true, data.StatusManagerProject.StatusRequest
			}
		} else {
			for _, dt := range data.StatusProjectLeader {
				if idemp == dt.IdEmp {
					if dt.StatusRequest != "Pending" {
						return true, dt.StatusRequest
					}
				}
			}
		}
	} else {
		return true, data.ResultRequest
	}

	return false, ""
}

func (d *MailController) ResponseApproveLeader(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	p := struct {
		Param  string
		Reason string
	}{}

	err := r.GetPayload(&p)

	if err != nil {
		return err.Error
	}

	payload := new(ParameterURLModel)
	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), payload)

	dataLeave := make([]*RequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	fmt.Println("----------- idrequest ", payload.IdRequest)

	dbFilter = append(dbFilter, db.Eq("_id", payload.IdRequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := d.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return d.SetResultInfo(true, "error on query", nil)
	}
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeave, 0, false)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	tk.Println("---------------- dataLeave ", dataLeave)

	if dataLeave[0].ResultRequest == "Canceled by User" {
		return d.SetResultInfo(true, "Request has been cancel by user", nil)
	}

	resp, sts := d.CheckResponse(r, false, dataLeave[0], payload.IdLeader)

	if resp == true {
		return d.SetResultInfo(true, "already "+sts+" request", nil)
	}

	dash := DashboardController(*d)
	LeaderName := ""
	EmailLeader := ""

	for i, dt := range dataLeave[0].StatusProjectLeader {
		if dt.IdEmp == payload.IdLeader {
			if payload.ApproveLeader == "yes" {
				dataLeave[0].StatusProjectLeader[i].StatusRequest = "Approved"
				dash.SetHistoryLeave(r, payload.UserId, dataLeave[0].Id, dataLeave[0].LeaveFrom, dataLeave[0].LeaveTo, "Your Request has been Approved Leader by "+dt.Name, "Pending", dataLeave[0])

			} else {
				dataLeave[0].StatusProjectLeader[i].StatusRequest = "Declined"
				dataLeave[0].StatusProjectLeader[i].Reason = p.Reason
				dash.SetHistoryLeave(r, payload.UserId, dataLeave[0].Id, dataLeave[0].LeaveFrom, dataLeave[0].LeaveTo, "Your Request has been Declined Leader by "+dt.Name, "Pending", dataLeave[0])
			}
			LeaderName = dt.Name
			EmailLeader = dt.Email
		}
	}

	err = d.Ctx.Save(dataLeave[0])
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	d.CheckLeaderDeclined(r, dataLeave[0], payload.Level)

	// d.SendMailBusinesAnalist(r, payload.IdRequest)
	desc := "Request approved by leader"
	var stsReq = "Approved"
	if payload.ApproveLeader != "yes" {
		stsReq = "Declined"
		desc = "Request declined by leader"
	}

	listDetailLeave := dash.GetLeavebyDateFilterIdTransc(r, dataLeave[0].Id)
	for _, ireqByDate := range listDetailLeave {
		ireqByDate.StsByLeader = stsReq
		err = d.Ctx.Save(ireqByDate)
		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

	}
	//log
	emptyremote := new(RemoteModel)
	service := services.LogService{
		dataLeave[0],
		emptyremote,
		"leave",
	}
	log := tk.M{}
	log.Set("Status", stsReq)
	log.Set("Desc", desc)
	log.Set("NameLogBy", LeaderName)
	log.Set("EmailNameLogBy", EmailLeader)
	err = service.ApproveDeclineLog(log)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	return dataLeave[0].StatusProjectLeader

}

func (d *MailController) CheckLeaderDeclined(r *knot.WebContext, dataLeave *RequestLeaveModel, level int) interface{} {
	aprv := 0
	dec := 0
	notif := NotificationController(*d)
	getnotif := notif.GetDataNotification(r, dataLeave.Id)

	for _, dl := range dataLeave.StatusProjectLeader {
		getnotif.Notif.ManagerApprove = dl.Name
		if dl.StatusRequest == "Declined" {
			getnotif.Notif.Description = "Declined by leader " + dl.Name
			dec = dec + 1
		} else if dl.StatusRequest == "Approved" {
			getnotif.Notif.Description = "Approved by leader " + dl.Name
			aprv = aprv + 1
		}
	}

	tk.Println("-------- masuk ", aprv+dec == len(dataLeave.StatusProjectLeader))

	if aprv > dec && aprv+dec == len(dataLeave.StatusProjectLeader) {
		d.SendMailManager(r, dataLeave.UserId, dataLeave.Id, level)
	} else if aprv+dec == len(dataLeave.StatusProjectLeader) {

		dataLeave.ResultRequest = "Declined"
		dataLeave.StatusManagerProject.StatusRequest = "Declined"
		dataLeave.StatusManagerProject.Name = "Leader"
		d.SendMailCancelLeave(r, dataLeave, level, "leader")
		err := d.Ctx.Save(dataLeave)
		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

		dash := DashboardController(*d)
		dash.SetHistoryLeave(r, dataLeave.UserId, dataLeave.Id, dataLeave.LeaveFrom, dataLeave.LeaveTo, "Your Request has been Declined Leader", "Declined", dataLeave)
		listDetailLeave := dash.GetLeavebyDateFilterIdTransc(r, dataLeave.Id)
		for _, ireqByDate := range listDetailLeave {
			ireqByDate.StsByManager = "Declined"
			err := d.Ctx.Save(ireqByDate)
			if err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			}

		}

	}

	getnotif.Notif.Status = dataLeave.ResultRequest
	getnotif.Notif.StatusApproval = dataLeave.ResultRequest

	notif.InsertNotification(getnotif)
	return ""

}

func EncodeMessage(message string) string {
	sEnc := base64.URLEncoding.EncodeToString([]byte(message))
	return sEnc
}

func DecodeMessage(msg string) string {
	sDec, _ := base64.URLEncoding.DecodeString(msg)
	return string(sDec)
}

func encrypt(text1 string) string {
	tst := MailReadConfig()
	key := []byte(string(tst["key"]))
	// fmt.Println(key)
	text := []byte(text1)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fmt.Println(err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return string(ciphertext)
}

func GCMEncrypter(text string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := nonce
	ciphertext = append(nonce, aesgcm.Seal(nil, nonce, plaintext, nil)...)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func GCMDecrypter(text string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	ciphertext, _ := base64.StdEncoding.DecodeString(text)

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return string(plaintext)
}

func GCMDecrypter_1(text string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	ciphertext, _ := base64.StdEncoding.DecodeString(text)

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		panic(err.Error())
	}
	tk.Println("-------------- aesgcm ", aesgcm)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	tk.Println("-------------- plaintext ", plaintext)
	if err != nil {
		tk.Println("-------------- err ", err.Error())
		panic(err.Error())
	}

	return string(plaintext)
}
func decrypt(text1 string) string {
	tst := MailReadConfig()
	key := []byte(string(tst["key"]))
	// fmt.Println(key)
	text := []byte(text1)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	if len(text) < aes.BlockSize {
		fmt.Println("too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		fmt.Println(err)
	}
	return string(data)
}

type dataR struct {
	NameRemote string
	Reason     string
	Details    []dataDetails
	Decline    string
	Approve    string
}

type dataDetails struct {
	Date   string
	Name   string
	Result string
	Reason string
}

func FileLeaveTemplate(filename string, data dataR) ([]byte, error) {
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

// EmergencyLeaveAdminRequest ...
func (c *MailController) EmergencyLeaveAdminRequest(r *knot.WebContext, p *RequestLeaveModel, level int, userid string) error {
	user := []string{p.StatusManagerProject.Email}
	for _, ld := range p.StatusProjectLeader {
		user = append(user, ld.Email)
	}
	for _, ba := range p.StatusBusinesAnalyst {
		user = append(user, ba.Email)
	}
	for _, ld := range p.StatusProjectLeader {
		if ld.Email != p.Email {
			user = append(user, ld.Email)
		}
	}
	if p.Location == "Indonesia" {
		dataHR, err := c.GetDataHR(r)
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
	c.SendMailManagerAdminRequest(r, userid, p.Id, level)
	c.RequestLeaveOnDateV3(p, userid)
	// c.RequestLeaveOnDateV2(p, userid)
	return nil
}

func (d *MailController) SendMailManagerAdminRequest(r *knot.WebContext, userid string, ID string, level int) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	dataRequest := d.getDataLeave(ID)
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("projectruleid", "5bebcf58805a8039b022c3e3"))
	query := tk.M{}
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	data := []*SysUserModel{}
	crsData, errData := d.Ctx.Find(NewSysUserModel(), query)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return d.SetResultInfo(true, "error in query", nil)
	}
	if errData != nil {
		return d.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)
	adminMail := data[0].Email
	conf, emailAddress := d.EmailConfiguration()
	m := gomail.NewMessage()
	m.SetHeader("From", emailAddress)
	user := []string{}
	//admin
	user = append(user, adminMail)
	//manager
	for _, mgr := range dataRequest.ProjectManagerList {
		user = append(user, mgr.Email)
	}
	//PC
	for _, dtl := range dataRequest.BranchManager {
		user = append(user, dtl.Email)
	}
	//User
	user = append(user, dataRequest.Email)

	from := "Admin"
	if level == 1 {
		from = "Manager"
	} else if level == 5 {
		from = "Admin"
	} else if level == 6 {
		from = "Project Coordiantor"
	}

	m.SetHeader("To", user...)
	addd := map[string]string{"from": from, "to": adminMail, "namerequest": dataRequest.Name, "reason": dataRequest.Reason, "leaveFrom": dataRequest.LeaveFrom, "leaveTo": dataRequest.LeaveTo, "noOfDays": strconv.Itoa(dataRequest.NoOfDays), "DateCreate": dataRequest.DateCreateLeave}
	mailsubj := tk.Sprintf("%v", "emergency leave from "+dataRequest.Name)
	m.SetHeader("Subject", mailsubj)
	bd, er := d.File("adminRequest.html", addd)
	if er != nil {
		return d.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", bd)
	d.DelayProcess(5)
	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()
	return ret
}

// RequestLeaveOnDateV3 ...

func (c *MailController) RequestLeaveOnDateV3(m *RequestLeaveModel, userid string) interface{} {
	lvDate := new(AprovalRequestLeaveModel)
	tnow := time.Now()
	for _, d := range m.LeaveDateList {
		dt, _ := time.Parse("2006-01-02", d)
		if dt.IsZero() {
			break
		}
		onday := dt.Weekday()
		if onday.String() == "Sunday" || onday.String() == "Saturday" {

		} else {
			var dbFilter []*db.Filter
			dbFilter = append(dbFilter, db.Eq("userid", userid))
			dbFilter = append(dbFilter, db.Eq("isdelete", false))
			dbFilter = append(dbFilter, db.Eq("stsbymanager", "Approved"))
			dbFilter = append(dbFilter, db.Eq("stsbymanager", "Approved"))
			dbFilter = append(dbFilter, db.Eq("dateleave", dt.Format("2006-01-02")))
			query := tk.M{}
			if len(dbFilter) > 0 {
				query.Set("where", db.And(dbFilter...))
			}
			data := []*AprovalRequestLeaveModel{}
			crsData, errData := c.Ctx.Find(NewAprovalRequestLeaveModel(), query)
			if crsData != nil {
				defer crsData.Close()
			} else {
				return c.SetResultInfo(true, "error in query", nil)
			}
			if errData != nil {
				return c.SetResultInfo(true, errData.Error(), nil)
			}
			errData = crsData.Fetch(&data, 0, false)
			if len(data) == 0 {
				lvDate.Id = bson.NewObjectId().Hex()
				lvDate.IdRequest = m.Id
				lvDate.Name = m.Name
				lvDate.EmpId = m.EmpId
				lvDate.Designation = m.Designation
				lvDate.Location = m.Location
				lvDate.Departement = m.Departement
				lvDate.Reason = m.Reason
				lvDate.Email = m.Email
				lvDate.Address = m.Address
				lvDate.Contact = m.Contact
				lvDate.Project = m.Project
				lvDate.YearLeave = m.YearLeave
				lvDate.PublicLeave = m.PublicLeave
				lvDate.IsEmergency = m.IsEmergency
				ondate := dt.Format("2006-01-02")
				lvDate.DateLeave = ondate
				valYear, valMonth, valDay := dt.Date()
				lvDate.DayVal = valDay
				lvDate.MonthVal = int(valMonth)
				lvDate.YearVal = valYear
				lvDate.UserId = userid
				//this function intime in dashboard controller
				if inTime(dt, tnow) {
					lvDate.IsCutOff = true
				} else {
					lvDate.IsCutOff = false
				}
				//lvDate.IsCutOff = true
				lvDate.IsReset = false
				lvDate.StsByLeader = "Approved"  //m.StatusProjectLeader[0].StatusRequest
				lvDate.StsByManager = "Approved" //m.StatusManagerProject.StatusRequest
				if m.StatusManagerProject.StatusRequest == "Declined" {
					lvDate.IsDelete = true
				} else if m.StatusManagerProject.StatusRequest == "Approved" {
					lvDate.IsDelete = false
				} else {
					lvDate.IsDelete = false
				}
				err := c.Ctx.Save(lvDate)
				if err != nil {
					return err
				}
			}
		}
	}
	return c.SetResultInfo(false, "Success", nil)
}

func (d *MailController) getDataLeaveByDate(Id string) []*AprovalRequestLeaveModel {
	dataLeaveByDate := make([]*AprovalRequestLeaveModel, 0)
	query := tk.M{}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("idrequest", Id))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	crs, err := d.Ctx.Find(NewAprovalRequestLeaveModel(), query)
	if err != nil {
		fmt.Println(err.Error)
	}
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	err = crs.Fetch(&dataLeaveByDate, 0, false)
	if err != nil {
		fmt.Println(err.Error)
	}
	return dataLeaveByDate
}

func (d *MailController) SendMailUserOption(r *knot.WebContext, option []*ChangeOptionModel) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	conf, emailAddress := d.EmailConfiguration()
	urlConf := helper.ReadConfig()
	hrd := urlConf.GetString("HrdMail")
	var to []string
	to = append(to, hrd)
	for _, usr := range option {
		if usr.Email != "" {
			to = append(to, usr.Email)
		}

	}
	mailsubj := tk.Sprintf("%v", "option Remote Option")
	m := gomail.NewMessage()
	typeofremotestring := ""
	if option[0].Remote.RemoteActive == false {
		typeofremotestring = "Banned Remote"
	} else if option[0].Remote.RemoteActive == true && option[0].Remote.Monthly == true && option[0].Remote.FullMonth == true {
		typeofremotestring = "full access remote"
	} else if option[0].Remote.RemoteActive == true && option[0].Remote.Monthly == false && option[0].Remote.FullMonth == false {
		typeofremotestring = "Conditional for " + strconv.Itoa(option[0].Remote.ConditionalRemote) + " days"
	} else if option[0].Remote.RemoteActive == true && option[0].Remote.Monthly == true && option[0].Remote.FullMonth == false {
		typeofremotestring = "Monthly and Conditional"
	}
	addd := map[string]string{"name": option[0].Name, "typeofremote": typeofremotestring}
	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", mailsubj)
	if len(option) > 1 {
		bd, er := d.File("usersremotemail.html", addd)
		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)
	} else {
		bd, er := d.File("userremotemail.html", addd)
		if er != nil {
			return d.SetResultInfo(true, er.Error(), nil)
		}
		m.SetBody("text/html", bd)
	}

	d.DelayProcess(5)
	if err := conf.DialAndSend(m); err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.Data = "Successfully Send Mails"
	}
	m.Reset()
	return ret
}

var CancelLeaveData = struct {
	Data   []AprovalRequestLeaveModel
	Name   string
	Email  string
	Reason string
}{}

type URLParamCancelLeave struct {
	// Data        []AprovalRequestLeaveModel
	IdRequest   string
	UserIdAdmin string
	NameAdmin   string
	Result      bool
}

type TmplCancelLeave struct {
	Date       []string
	Name       string
	NameAdmin  string
	Reason     string
	UrlDecline string
	UrlApprove string
	Result     string
}

func (c *MailController) RequestCancelLeave(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := CancelLeaveData
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, "error payload", nil)
	}

	data := p.Data
	for _, dt := range data {
		dt.RequestDelete = true
		err = c.Ctx.Save(&dt)
		if err != nil {
			return c.SetResultInfo(true, "error payload", nil)
		}
	}

	er := c.SendAdminCancelLeave(k, p.Data, p.Reason)

	if er != nil {
		return c.SetResultInfo(true, er.Error(), nil)
	}
	return c.SetResultInfo(false, "data canceled successfully", nil)
}

func (c *MailController) SendAdminCancelLeave(k *knot.WebContext, Data []AprovalRequestLeaveModel, reason string) error {
	tmpl := TmplCancelLeave{}
	urlConf := helper.ReadConfig()
	surl := urlConf.GetString("BaseUrlEmail")
	// hrd := urlConf.GetString("HrdMail")
	m := MailController(*c)
	dataRequest := m.getDataLeave(Data[0].IdRequest)
	dateLeave := []string{}
	for _, dte := range Data {
		dateLeave = append(dateLeave, dte.DateLeave)
	}
	tmpl.Date = dateLeave
	tmpl.Name = dataRequest.Name
	tmpl.Reason = reason

	rule := c.GetAdminProjectRule(k)
	admin1 := c.GetAdmin(k, rule.Id.Hex())

	for _, adm := range admin1 {
		appr := new(URLParamCancelLeave)
		tmpl.NameAdmin = adm.Fullname
		uriApproved := surl + "/mail/admapprovecancelleave"
		urlApprove, _ := http.NewRequest("GET", uriApproved, nil)
		mi := gomail.NewMessage()
		// appr.Data = Data
		appr.IdRequest = Data[0].IdRequest
		appr.Result = true
		appr.NameAdmin = adm.Fullname
		appr.UserIdAdmin = adm.Id
		paramApp, _ := json.Marshal(appr)

		qApprove := urlApprove.URL.Query()
		qApprove.Add("param", GCMEncrypter(string(paramApp)))
		urlApprove.URL.RawQuery = qApprove.Encode()

		appr.Result = false

		mi.SetHeader("To", adm.Email)

		paramDec, _ := json.Marshal(appr)
		uriDeclined := surl + "/mail/admdeclinecancelleave"
		urlDecline, _ := http.NewRequest("GET", uriDeclined, nil)
		qDecline := urlDecline.URL.Query()
		qDecline.Add("param", GCMEncrypter(string(paramDec)))
		urlDecline.URL.RawQuery = qDecline.Encode()

		tmpl.UrlApprove = urlApprove.URL.String()
		tmpl.UrlDecline = urlDecline.URL.String()

		conf, emailAddress := m.EmailConfiguration()

		mailsubj := tk.Sprintf("%v", dataRequest.Name+" Request cancel leave")
		mi.SetHeader("From", emailAddress)

		mi.SetHeader("Subject", mailsubj)

		bd, er := FileCancelLeave("requestcancelleave.html", tmpl)

		if er != nil {
			return er
		}
		mi.SetBody("text/html", string(bd))

		m.DelayProcess(5)

		if err := conf.DialAndSend(mi); err != nil {
			return er
		}
		mi.Reset()

	}

	return nil

}

func (c *MailController) AdmApproveCancelLeave(k *knot.WebContext) interface{} {

	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (c *MailController) AdmDeclineCancelLeave(k *knot.WebContext) interface{} {

	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = "_layoutEmail.html"
	return ""
}

func (c *MailController) GetAdminProjectRule(k *knot.WebContext) ProjectRuleModel {
	k.Config.OutputType = knot.OutputJson

	dataProjRule := make([]ProjectRuleModel, 0)
	query := tk.M{}
	var filter []*db.Filter
	filter = append(filter, db.Eq("Level", 5))
	if len(filter) > 0 {
		query.Set("where", db.And(filter...))
	}
	crsProjRule, errProjRule := c.Ctx.Find(NewProjectRuleModel(), query)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return dataProjRule[0]
	}

	errProjRule = crsProjRule.Fetch(&dataProjRule, 0, false)
	if errProjRule != nil {
		return dataProjRule[0]
	}

	return dataProjRule[0]
}

func (c *MailController) GetAdmin(k *knot.WebContext, projectrule string) []SysUserModel {
	k.Config.OutputType = knot.OutputJson

	user := make([]SysUserModel, 0)
	query := tk.M{}
	var filter []*db.Filter
	filter = append(filter, db.Eq("projectruleid", projectrule))
	if len(filter) > 0 {
		query.Set("where", db.And(filter...))
	}
	crsProjRule, errProjRule := c.Ctx.Find(NewSysUserModel(), query)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return user
	}

	errProjRule = crsProjRule.Fetch(&user, 0, false)
	if errProjRule != nil {
		return user
	}

	return user
}

func FileCancelLeave(filename string, data TmplCancelLeave) ([]byte, error) {
	// fmt.Println("------ masuk file", filename)
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		// fmt.Println("------ masuk error 0.1 ", err.Error())
		return body, err
	}

	templ, err := template.ParseFiles(filepath.Join(t, "views", "template", filename))
	// fmt.Println("------ masuk templ ", templ)
	if err != nil {
		// fmt.Println("------ masuk error 0 ", err.Error())
		return body, err
	}
	// fmt.Println("------ masuk data ", data.LastYearLeave)
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		fmt.Println("------ masuk error 1 ", err.Error())
		return body, err
	}
	fmt.Println("------ masuk error 2 ")
	body = buffer.Bytes()

	return body, nil
}

func (d *MailController) ResponseApproveCancel(r *knot.WebContext) interface{} {
	// tk.Println("------------- masuk")
	r.Config.OutputType = knot.OutputJson
	p := struct {
		Param  string
		Reason string
	}{}

	err := r.GetPayload(&p)
	// fmt.Println("--------------- payload", p.IdRequest)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	dash := DashboardController(*d)

	payload := new(URLParamCancelLeave)
	decript := GCMDecrypter_1(p.Param)

	json.Unmarshal([]byte(decript), payload)
	fmt.Println("---------------  decript ", payload)
	res := ""
	if payload.Result == false {
		res = "Declined"
	} else {
		res = "Approved"
	}
	data := d.GetDatarequestCancel(r, payload.IdRequest)
	fmt.Println("---------------  data ", data)

	if len(data) > 0 {
		for _, dt := range data {
			if dt.RequestDelete == false {
				if dt.IsDelete == true {
					return d.SetResultInfo(true, "Cancel leave already approved", nil)
					if dt.IsCutOff == false {
						user := dash.GetDataSessionUser(r, dt.UserId)
						user[0].YearLeave = user[0].YearLeave + 1
						user[0].DecYear = user[0].DecYear + 1.0
						err = d.Ctx.Save(user[0])
						if err != nil {
							return d.SetResultInfo(true, err.Error(), nil)

						}
					}
				} else {
					return d.SetResultInfo(true, "Cancel leave already declined", nil)
				}

			}

			if payload.Result == false {
				dt.IsDelete = false
			} else {
				dt.IsDelete = true
			}

			dt.RequestDelete = false
			err = d.Ctx.Save(&dt)
			if err != nil {
				return d.SetResultInfo(true, err.Error(), nil)

			}

		}
	} else {
		return d.SetResultInfo(true, "Data not found it might already give approval", nil)
	}

	d.SendMailCancelApproval(r, data, res)

	return d.SetResultInfo(false, "Save successfully", nil)
}

func (d *MailController) GetDatarequestCancel(r *knot.WebContext, idrequest string) []AprovalRequestLeaveModel {
	r.Config.OutputType = knot.OutputJson

	data := make([]AprovalRequestLeaveModel, 0)
	query := tk.M{}
	var filter []*db.Filter
	filter = append(filter, db.Eq("idrequest", idrequest))
	filter = append(filter, db.Eq("requestdelete", true))
	if len(filter) > 0 {
		query.Set("where", db.And(filter...))
	}
	crsProjRule, errProjRule := d.Ctx.Find(NewAprovalRequestLeaveModel(), query)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return data
	}

	errProjRule = crsProjRule.Fetch(&data, 0, false)
	if errProjRule != nil {
		return data
	}

	return data
}

func (d *MailController) SendMailCancelApproval(r *knot.WebContext, Data []AprovalRequestLeaveModel, Result string) interface{} {
	tmpl := TmplCancelLeave{}
	// urlConf := helper.ReadConfig()
	// surl := urlConf.GetString("BaseUrlEmail")
	// hrd := urlConf.GetString("HrdMail")
	m := MailController(*d)
	dataRequest := m.getDataLeave(Data[0].IdRequest)
	dateLeave := []string{}
	for _, dte := range Data {
		dateLeave = append(dateLeave, dte.DateLeave)
	}
	tmpl.Date = dateLeave
	tmpl.Name = dataRequest.Name
	tmpl.Result = Result
	mi := gomail.NewMessage()

	to := []string{dataRequest.Email}
	rule := d.GetAdminProjectRule(r)
	admin1 := d.GetAdmin(r, rule.Id.Hex())
	for _, adm := range admin1 {

		to = append(to, adm.Email)

	}
	mi.SetHeader("To", to...)
	conf, emailAddress := m.EmailConfiguration()

	mailsubj := tk.Sprintf("%v", dataRequest.Name+" Request cancel leave "+Result)
	mi.SetHeader("From", emailAddress)

	mi.SetHeader("Subject", mailsubj)

	bd, er := FileCancelLeave("approvalcancelleave.html", tmpl)

	if er != nil {
		return er
	}

	mi.SetBody("text/html", string(bd))

	m.DelayProcess(5)

	if err := conf.DialAndSend(mi); err != nil {
		return er
	}
	mi.Reset()
	return "Success"
}
