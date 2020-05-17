package main

import (
	"bufio"
	"bytes"
	. "creativelab/ecleave-dev/controllers"
	"creativelab/ecleave-dev/helper"
	"creativelab/ecleave-dev/repositories"
	"creativelab/ecleave-dev/services"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/dbox"
	_ "github.com/creativelab/dbox/dbc/mongo"
	gomail "gopkg.in/gomail.v2"

	"github.com/creativelab/orm"
	tk "github.com/creativelab/toolkit"
)

var (
	wd = func() string {
		d, _ := os.Getwd()
		return d + "/"
	}()
	basePath = (func(dir string, err error) string { return dir }(os.Getwd()))
)

func main() {
	conn, err := PrepareConnection()
	if err != nil {
		log.Println(err)
	}

	ctx := orm.New(conn)

	c := new(BaseController)
	c.Ctx = ctx

	// messages := make(chan int)
	Count_Leave(c)
	// Use this channel to follow the execution status
	// of our goroutines :D
	// done := make(chan bool)

	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	Count_Leave(c)
	// 	messages <- 1
	// 	done <- true
	// }()
	// go func() {
	// 	time.Sleep(time.Second * 2)
	// 	messages <- 2
	// 	// CheckOneYear(c)
	// 	done <- true
	// }()
	// go func() {
	// 	time.Sleep(time.Second * 1)
	// 	// Count_Leave(c)
	// 	countLeaveNewEmp(c)
	// 	messages <- 3
	// 	done <- true
	// }()
	// go func() {
	// 	for i := range messages {
	// 		fmt.Println(i)
	// 	}
	// }()
	// for i := 0; i < 3; i++ {
	// 	<-done
	// }

}

func check(u string, checked chan<- bool) {
	time.Sleep(4 * time.Second)
	checked <- true
}

func IsReachable(c *BaseController, urls []string) bool {

	ch := make(chan bool, 1)
	for _, url := range urls {
		go func(u string) {
			checked := make(chan bool)
			go check(u, checked)
			select {
			case ret := <-checked:
				// Count_Leave(c)
				ch <- ret
			case <-time.After(1 * time.Second):
				countLeaveNewEmp(c)
				ch <- false
			}
		}(url)
	}
	return <-ch
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

func Count_Leave(c *BaseController) {

	t := time.Now()

	if int(t.Month()) == 1 {
		if t.Day() == 1 {
			ResetJanuary(c)
		}

	} else if int(t.Month()) == 7 {
		if t.Day() == 1 {
			ResetJuly(c)
		}

	}
	countLeaveNewEmp(c)
	CheckOneYear(c)

	// return "Success"
}

func ResetJanuary(c *BaseController) interface{} {
	user, err := GetuserData(c)
	if err != nil {
		return err.Error()
	}

	urlConf := ReadNewConfig()
	hrd := urlConf.GetString("HrdMail")

	if len(user) > 0 {

		//...add log
		logs := NewLogCronModel()
		logs.Typelog = "count leave reset january"
		logs.Date = time.Now().Format("2006-01-02 15:04:05")
		logDetail := []LogCronDetailModel{}
		repositories.Ctx = c.Ctx
		service := services.LogActivities{}
		//...
		dataparam_name := ""
		dataparam_lastYearLeave := 0
		dataparam_periodReset := ""
		dataparam_resetYearLeave := 0
		for _, usr := range user {
			_, year, _, _ := helper.IsMoreAYear(usr.JointDate)

			to := []string{usr.Email, hrd}
			// var dataparam DataReset

			dataparam_name = usr.Fullname
			dataparam_lastYearLeave = usr.YearLeave
			dataparam_periodReset = "Reset January"

			// tk.Println("-------------------- year nya ", usr.Fullname+"  "+it)
			if year == 1 {

				usr.DecYear = 18.0
				usr.YearLeave = 18

				dataparam_resetYearLeave = usr.YearLeave
				_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
				if er != nil {
					tk.Println(er.Error())
				}

			} else {
				if year >= 2 {
					tk.Println("-------------------- masuk 2")
					if usr.YearLeave >= 1 {
						// tk.Println("-------------- sikkpy")
						// tk.Println("-------------- name ", usr.Fullname)
						// tk.Println("-------------- year ", usr.YearLeave)
						if usr.DecYear > 0.0 {

							leave := usr.DecYear + 18.0
							usr.DecYear = leave
							usr.YearLeave = int(leave)

							dataparam_resetYearLeave = int(leave)
							_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
							if er != nil {
								tk.Println(er.Error())
							}
						} else {
							leave := float64(usr.YearLeave) + 18.0
							usr.DecYear = leave
							usr.YearLeave = int(leave)

							dataparam_resetYearLeave = int(leave)
							_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
							if er != nil {
								tk.Println(er.Error())
							}
						}
					} else if usr.YearLeave < 1 {
						tk.Println("-------------------- tahun yang kurang 2")
						usr.YearLeave = usr.YearLeave + 0
						usr.DecYear = usr.DecYear

					}
				}

			}

			err = c.Ctx.Save(usr)
			if err != nil {
				//...add log
				logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted Janury" + " - " + usr.EmpId + " - " + usr.Fullname, Empid: usr.EmpId, Err: err.Error()})
				logs.Detail = logDetail
				service.SaveLog(*logs)
				//...
				return err.Error()
			}
			logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted January" + " - " + usr.EmpId + " - " + usr.Fullname, Empid: usr.EmpId, Err: ""})
		}
		//...add log
		logs.Detail = logDetail
		service.SaveLog(*logs)
		//...

	}

	return "success"
}

func ResetJuly(c *BaseController) interface{} {
	user, err := GetuserData(c)
	if err != nil {
		return err.Error()
	}

	urlConf := ReadNewConfig()
	hrd := urlConf.GetString("HrdMail")

	if len(user) > 0 {

		//...add log
		logs := NewLogCronModel()
		logs.Typelog = "count leave reset july"
		logs.Date = time.Now().Format("2006-01-02 15:04:05")
		logDetail := []LogCronDetailModel{}
		repositories.Ctx = c.Ctx
		service := services.LogActivities{}
		//...

		dataparam_name := ""
		dataparam_lastYearLeave := 0
		dataparam_periodReset := ""
		dataparam_resetYearLeave := 0

		for _, usr := range user {
			to := []string{usr.Email, hrd}
			// var dataparam DataReset

			dataparam_name = usr.Fullname
			dataparam_lastYearLeave = usr.YearLeave
			dataparam_periodReset = "Reset July"

			_, year, _, _ := helper.IsMoreAYear(usr.JointDate)
			tk.Println("------------------ ", usr.Fullname+" "+strconv.Itoa(year))
			if year >= 1 {
				if usr.YearLeave >= 18 {
					tk.Println("------------ 18 ")
					usr.YearLeave = 18
					usr.DecYear = 18.0

					dataparam_resetYearLeave = 18
					_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
					if er != nil {
						tk.Println(er.Error())
					}

				} else if usr.YearLeave < 18 {
					usr.YearLeave = usr.YearLeave + 0
					usr.DecYear = usr.DecYear

				}
			}

			err = c.Ctx.Save(usr)
			if err != nil {
				//...add log
				logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted july" + " - " + usr.EmpId + " - " + usr.Fullname, Empid: usr.EmpId, Err: err.Error()})
				logs.Detail = logDetail
				service.SaveLog(*logs)
				//...
				return err.Error()
			}
			logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted july" + " - " + usr.EmpId + " - " + usr.Fullname, Empid: usr.EmpId, Err: ""})
		}
		//...add log
		logs.Detail = logDetail
		service.SaveLog(*logs)
		//...
	}
	return "success"
}

func GetuserData(c *BaseController) ([]*SysUserModel, error) {
	// var dbFilter []*dbox.Filter
	// query := tk.M{}
	data := make([]*SysUserModel, 0)
	// dbFilter = append(dbFilter, dbox.Eq("_id", userid))

	crs, err := c.Ctx.Find(NewSysUserModel(), nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return data, nil
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data, err
	}
	return data, nil

}
func GetuserData2(c *BaseController) []*SysUserModel {
	// var dbFilter []*dbox.Filter
	// query := tk.M{}
	data := make([]*SysUserModel, 0)
	// dbFilter = append(dbFilter, dbox.Eq("empid", empid))

	// if len(dbFilter) > 0 {
	// 	query.Set("where", dbox.And(dbFilter...))
	// }

	crs, err := c.Ctx.Find(NewSysUserModel(), nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return data
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data
	}
	return data

}

// func ImportJointDate(c *BaseController) interface{} {
// 	t, err := os.Getwd()
// 	if err != nil {
// 		return err.Error()
// 	}

// 	pathFile := filepath.Join(t, "..", "..", "assets", "doc", "jointdate.xlsx")
// 	tk.Println("------------------ 1")
// 	xlFile, err := xlsx.OpenFile(pathFile)
// 	if err != nil {
// 		tk.Println("------------------ 3", err.Error())
// 		return err.Error()
// 	}
// 	tk.Println("------------------ 2")

// 	for _, sheet := range xlFile.Sheets {
// 		for a, row := range sheet.Rows {
// 			if a == 0 {
// 				continue
// 			}

// 			if len(row.Cells) == 0 {
// 				continue
// 			}

// 			var ID, joint string

// 			if len(row.Cells) > 0 {
// 				ID = row.Cells[0].String()
// 			}
// 			if len(row.Cells) > 1 {
// 				tm, err := row.Cells[1].GetTime(false)
// 				if err != nil {
// 					tk.Println("------------------ 4", err.Error())
// 					return err
// 				}

// 				joint = tm.Format("2006-01-02")

// 			}

// 			fmt.Printf("%s\n", joint)
// 			usr := GetuserData(c, ID)
// 			tk.Println("------------------ 518", ID)

// 			if len(usr) > 0 {
// 				usr[0].JointDate = joint
// 				err = c.Ctx.Save(usr[0])
// 				if err != nil {
// 					return err
// 				}
// 			}

// 		}
// 	}
// 	return "success"
// }

func countLeaveNewEmp(c *BaseController) {
	user := GetuserData2(c)

	// urlConf := ReadNewConfig()
	// tk.Println("-------------------- path ", urlConf)
	// hrd := urlConf.GetString("HrdMail")

	if len(user) > 0 {
		//...add log
		logs := NewLogCronModel()
		logs.Typelog = "count leave New Employee"
		logs.Date = getTimeIndo(c)
		logDetail := []LogCronDetailModel{}
		repositories.Ctx = c.Ctx
		service := services.LogActivities{}
		//...
		tk.Println("len user ------", len(user))
		tm := time.Now()
		tmNext := tm.AddDate(0, 1, 0)
		dt := tm.Format("2006-01-02")

		// dataparam_name := ""
		// dataparam_lastYearLeave := 0
		// dataparam_periodReset := ""
		// dataparam_resetYearLeave := 0

		for _, usr := range user {

			isMyear, year, month, day := helper.IsMoreAYear(usr.JointDate)
			tk.Println("name user ------", usr.Fullname+" "+strconv.Itoa(year)+" "+strconv.Itoa(month))
			if isMyear == false {

				if year < 1 {
					var leave float64 = 0.0
					// for i := 0; i < month; i++ {
					if usr.YearLeave == 0 {
						tk.Println("------------- masuk sini", year)
						tk.Println("------------ user1 ", usr.Fullname)
						tk.Println("------------ month ", month)
						tk.Println("------------ year ", year)
						tk.Println("------------ day1 ", day)

						if day <= 30 && day > 0 {

							if usr.AddLeave == "" {
								tk.Println("------------ baru144 ")
								ms, _ := time.Parse("2006-01-02", usr.JointDate)
								// tk.Println("------------ day ", ms)
								next := ms.AddDate(0, 1, 0)
								if month == 0 {
									month = 1
								}
								yearLeave := 1.5 * float64(month)
								if next.Before(tm) {
									if tm.Day() < next.Day() {
										tk.Println("Masuk Test Tyo ----")
										next = time.Date(next.Year(), tm.Month(), next.Day(), 0, 0, 0, 0, tm.Location())
									} else {
										next = time.Date(next.Year(), tmNext.Month(), next.Day(), 0, 0, 0, 0, tm.Location())
										yearLeave = yearLeave + 1.5
									}
								}
								// tk.Println("------------ add ", next)
								snow := next.Format("2006-01-02")
								usr.AddLeave = snow
								leave = leave + yearLeave
								usr.DecYear = leave
								logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted new employee" + " - " + usr.EmpId + " - " + usr.Fullname})
							} else {
								if dt == usr.AddLeave {
									tk.Println("------------ kurang setaun ")
									ms, _ := time.Parse("2006-01-02", usr.JointDate)
									next := ms.AddDate(0, 1, 0)
									snow := next.Format("2006-01-01")
									usr.AddLeave = snow
									leave = usr.DecYear + 1.5
									usr.DecYear = leave
									logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted new employee" + " - " + usr.EmpId + " - " + usr.Fullname})
								} else {
									tk.Println("------------ kurang setaun ")
									leave = float64(usr.YearLeave)
								}

							}

						} else {

							if dt == usr.AddLeave {
								tk.Println("masuk count kurang setaunss")
								// tk.Println("------------ baru2 ")
								leave = usr.DecYear + 1.5
								ms, _ := time.Parse("2006-01-02", usr.AddLeave)
								next := ms.AddDate(0, 1, 0)
								snow := next.Format("2006-01-02")
								usr.AddLeave = snow
								usr.DecYear = leave
							} else {
								if usr.AddLeave == "" {
									ms, _ := time.Parse("2006-01-02", usr.JointDate)
									if ms.Year() != tm.Year() {
										tk.Println("------------ baru1 ")
										next := tm.AddDate(0, 1, 0)
										snow := next.Format("2006-01-02")
										usr.AddLeave = snow
										leave = float64(month) * 1.5
										tk.Println("------------ baru1 ", leave)
										usr.DecYear = leave
									} else {
										tk.Println("------------ day00000 ")
										next := tm.AddDate(0, 1, 0)
										// tk.Println("------------ add ", next)
										snow := next.Format("2006-01-02")
										usr.AddLeave = snow
										leave = usr.DecYear + 1.5
										usr.DecYear = leave
									}

								} else {
									tk.Println("masuk count kurang setaun nanana")
									leave = float64(usr.YearLeave)
								}

							}
						}
					} else {
						if dt == usr.AddLeave {
							tk.Println("masuk count kurang setaun nananammm")
							leave = float64(usr.DecYear) + 1.5
							ms, _ := time.Parse("2006-01-02", usr.AddLeave)
							next := ms.AddDate(0, 1, 0)
							snow := next.Format("2006-01-02")
							usr.AddLeave = snow
							usr.DecYear = leave
						} else {
							if usr.AddLeave == "" {
								tk.Println("------------ baru133 ")
								ms, _ := time.Parse("2006-01-02", usr.JointDate)
								// tk.Println("------------ day ", ms)
								next := ms.AddDate(0, 1, 0)
								yearLeave := 1.5 * float64(month)
								if next.Before(tm) {
									if tm.Day() < next.Day() {
										next = time.Date(next.Year(), tm.Month(), next.Day(), 0, 0, 0, 0, tm.Location())
									} else {
										next = time.Date(next.Year(), tmNext.Month(), next.Day(), 0, 0, 0, 0, tm.Location())
										yearLeave = yearLeave + 1.5
									}
								}
								// tk.Println("------------ add ", next)
								snow := next.Format("2006-01-02")
								usr.AddLeave = snow
								usr.DecYear = yearLeave
								leave = yearLeave
							} else {
								leave = float64(usr.YearLeave)
							}

						}

					}

					// }
					onleave := int(leave)
					usr.YearLeave = onleave
					err := c.Ctx.Save(usr)

					if err != nil {
						//...add log
						logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted < 1 new employee" + " - " + usr.EmpId + " - " + usr.Fullname})
						//...
						tk.Println(err.Error())
					}
				} //else if year == 1 && month == 0 {
				// 	logs.Typelog = "count leave New Employee 1 year"
				// 	tk.Println("---------------------- 1 tahun ", month)
				// 	tk.Println("name year for 1 year------", usr.Fullname)
				// 	to := []string{usr.Email, hrd}
				// 	dataparam_name = usr.Fullname
				// 	dataparam_lastYearLeave = usr.YearLeave
				// 	dataparam_periodReset = "Reset new employee for 1 year"

				// 	t := time.Now()
				// 	now := t.Format("2006-01-02")
				// 	// tk.Println("now ------", now)
				// 	mt := int(t.Month())
				// 	joint, _ := time.Parse("2006-01-02", usr.JointDate)
				// 	// tk.Println("joint ------", joint)
				// 	ayear := joint.AddDate(0, 12, 0)
				// 	yearForm := ayear.Format("2006-01-02")
				// 	tk.Println("now ------", now)
				// 	tk.Println("yearForm ------", yearForm)
				// 	if now == yearForm {
				// 		tk.Println("yearForm ------", mt)
				// 		if mt <= 12 {
				// 			// tk.Println("masuk nama ", usr.Fullname)
				// 			// tk.Println("masuk hitung sini", mt)
				// 			tk.Println("masuk hitung year", 12-mt)
				// 			if mt <= 12 {
				// 				sisaM := 12 - mt + 1
				// 				if sisaM > 0 {
				// 					sisaL := float64(sisaM) * 1.5
				// 					resL := int(sisaL)
				// 					usr.YearLeave = resL
				// 					usr.DecYear = sisaL
				// 					next := t.AddDate(0, 1, 0)
				// 					tk.Println("------------ add ", next)
				// 					snow := next.Format("2006-01-02")
				// 					usr.AddLeave = snow
				// 					err := c.Ctx.Save(usr)

				// 					dataparam_resetYearLeave = usr.YearLeave
				// 					_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
				// 					if er != nil {
				// 						tk.Println(er.Error())
				// 					}

				// 					if err != nil {
				// 						//...add log
				// 						logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
				// 						//...
				// 						tk.Println(err.Error())
				// 					} else {
				// 						logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted = 1 year" + " - " + usr.EmpId + " - " + usr.Fullname})
				// 					}
				// 				}
				// 			}
				// 		}

				// 	}

				// 	// mt := int(t.Month())
				// 	// if mt != 12 {
				// 	// 	tk.Println("masuk nama ", usr.Fullname)
				// 	// 	tk.Println("masuk hitung sini", mt)
				// 	// 	tk.Println("masuk hitung year", 12-mt)
				// 	// 	if mt < 12 {
				// 	// 		sisaM := 12 - mt
				// 	// 		if sisaM > 0 {
				// 	// 			sisaL := float64(sisaM) * 1.5
				// 	// 			resL := int(sisaL)
				// 	// 			usr.YearLeave = resL
				// 	// 			err := c.Ctx.Save(usr)

				// 	// 			if err != nil {
				// 	// 				//...add log
				// 	// 				logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
				// 	// 				//...
				// 	// 				tk.Println(err.Error())
				// 	// 			} else {
				// 	// 				logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
				// 	// 			}
				// 	// 		}
				// 	// 	}
				// 	// }

				// }

			}
		}

		//...add log
		logs.Detail = logDetail
		service.SaveLog(*logs)
		//...

	}

	// return "success"

}

func CheckOneYear(c *BaseController) {
	user := GetuserData2(c)

	urlConf := ReadNewConfig()
	// tk.Println("-------------------- path ", urlConf)
	hrd := urlConf.GetString("HrdMail")

	if len(user) > 0 {
		//...add log
		logs := NewLogCronModel()
		logs.Typelog = "count leave 1 year"
		logs.Date = getTimeIndo(c)
		logDetail := []LogCronDetailModel{}
		repositories.Ctx = c.Ctx
		service := services.LogActivities{}
		//...
		tk.Println("len user ------", len(user))
		// tm := time.Now()
		// tmNext := tm.AddDate(0, 1, 0)
		// dt := tm.Format("2006-01-02")

		dataparam_name := ""
		dataparam_lastYearLeave := 0
		dataparam_periodReset := ""
		dataparam_resetYearLeave := 0

		for _, usr := range user {

			isMyear, year, month, _ := helper.IsMoreAYear(usr.JointDate)
			tk.Println("name user ------", usr.Fullname+" "+strconv.Itoa(year)+" "+strconv.Itoa(month))
			start, _ := time.Parse("2006-01-02", usr.JointDate)
			now := time.Now()

			if isMyear == false && start.Day() == now.Day() {

				if year == 1 && month == 0 {
					// if month == 0 {

					tk.Println("---------------------- 1 tahun ", month)
					tk.Println("name year for 1 year------", usr.Fullname)
					to := []string{usr.Email, hrd}
					dataparam_name = usr.Fullname
					dataparam_lastYearLeave = usr.YearLeave
					dataparam_periodReset = "Reset new employee for 1 year"

					t := time.Now()
					now := t.Format("2006-01-02")
					// tk.Println("now ------", now)
					mt := int(t.Month())
					joint, _ := time.Parse("2006-01-02", usr.JointDate)
					// tk.Println("joint ------", joint)
					ayear := joint.AddDate(0, 12, 0)
					yearForm := ayear.Format("2006-01-02")
					tk.Println("now ------", now)
					tk.Println("yearForm ------", yearForm)
					// if now == yearForm {
					tk.Println("yearForm ------", mt)
					// if mt <= 12 {
					// tk.Println("masuk nama ", usr.Fullname)
					// tk.Println("masuk hitung sini", mt)
					// tk.Println("masuk hitung year", 12-mt)
					// if mt <= 12 {
					sisaM := 12 - mt + 1
					if sisaM > 0 {
						sisaL := float64(sisaM) * 1.5
						resL := int(sisaL)
						usr.YearLeave = resL
						usr.DecYear = sisaL
						next := t.AddDate(0, 1, 0)
						tk.Println("------------ add ", next)
						snow := next.Format("2006-01-02")
						usr.AddLeave = snow
						err := c.Ctx.Save(usr)

						dataparam_resetYearLeave = usr.YearLeave
						_, er := SendEmailReset(c, dataparam_name, dataparam_lastYearLeave, dataparam_resetYearLeave, dataparam_periodReset, to)
						if er != nil {
							tk.Println(er.Error())
						}

						if err != nil {
							//...add log
							logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
							//...
							tk.Println(err.Error())
						} else {
							logDetail = append(logDetail, LogCronDetailModel{Message: "Success count leave 1 year" + " - " + usr.EmpId + " - " + usr.Fullname})
						}
					}
					// }
					// }

					// }

					// mt := int(t.Month())
					// if mt != 12 {
					// 	tk.Println("masuk nama ", usr.Fullname)
					// 	tk.Println("masuk hitung sini", mt)
					// 	tk.Println("masuk hitung year", 12-mt)
					// 	if mt < 12 {
					// 		sisaM := 12 - mt
					// 		if sisaM > 0 {
					// 			sisaL := float64(sisaM) * 1.5
					// 			resL := int(sisaL)
					// 			usr.YearLeave = resL
					// 			err := c.Ctx.Save(usr)

					// 			if err != nil {
					// 				//...add log
					// 				logDetail = append(logDetail, LogCronDetailModel{Message: "Error Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
					// 				//...
					// 				tk.Println(err.Error())
					// 			} else {
					// 				logDetail = append(logDetail, LogCronDetailModel{Message: "Success Counted = 1" + " - " + usr.EmpId + " - " + usr.Fullname})
					// 			}
					// 		}
					// 	}
					// }

					// }
				}

			}
		}

		//...add log
		logs.Detail = logDetail
		service.SaveLog(*logs)
		//...

	}

	// return "success"

}

func getTimeIndo(c *BaseController) string {
	loc, _ := GetLocation(c)
	var timeindo string
	for _, lc := range loc {
		if lc.Location == "Indonesia" {
			timeindo, _ = helper.TimeLocationLog(lc.TimeZone)
		}
	}
	return timeindo
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

type DataReset struct {
	Name           string
	LastYearLeave  int
	ResetYearLeave int
	PeriodReset    string
}

func SendEmailReset(c *BaseController, name string, lastYearLeave int, resetYearleave int, periodreset string, to []string) (bool, error) {
	datareset := DataReset{}
	datareset.Name = name
	datareset.LastYearLeave = lastYearLeave
	// tk.Println("---------------------- lastyear ", lastYearLeave)
	datareset.ResetYearLeave = resetYearleave
	datareset.PeriodReset = periodreset
	// urlConf := config()
	conf, emailAddress := EmailConfiguration()

	mailsubj := tk.Sprintf("%v", "Leave Reset Count for "+datareset.PeriodReset)
	m := gomail.NewMessage()
	m.SetHeader("From", emailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("subject", mailsubj)

	bd, er := FileResetLeave("resetleaveemailtemplate.html", datareset)
	// tk.Println("----------------- masuk bd ", bd)
	if er != nil {
		tk.Println("masuk eroro ", er.Error())
		return false, er
	}
	m.SetBody("text/html", string(bd))

	if err := conf.DialAndSend(m); err != nil {
		tk.Println("----------------- error dial ", err.Error())
		return false, err
	}
	m.Reset()
	return true, nil
}

func EmailConfiguration() (*gomail.Dialer, string) {
	// r.Config.OutputType = knot.OutputJson
	config := config()
	conf := gomail.NewPlainDialer(config["Host"], 587, config["MailAddressName"], config["MailAddressPassword"])
	emailAddress := config["MailAddressName"]

	return conf, emailAddress
}

func FileResetLeave(filename string, data DataReset) ([]byte, error) {
	fmt.Println("------ masuk file", filename)
	t, err := os.Getwd()
	body := []byte{}
	if err != nil {
		fmt.Println("------ masuk error 0.1 ", err.Error())
		return body, err
	}
	templ, err := template.ParseFiles(filepath.Join(t, "..", "..", "views", "template", filename))
	fmt.Println("------ masuk templ ", templ)
	if err != nil {
		fmt.Println("------ masuk error 0 ", err.Error())
		return body, err
	}
	fmt.Println("------ masuk data ", data.LastYearLeave)
	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, data); err != nil {
		fmt.Println("------ masuk error 1 ", err.Error())
		return body, err
	}
	fmt.Println("------ masuk error 2 ")
	body = buffer.Bytes()

	return body, nil
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
