package newemp_leave

import (
	. "creativelab/ecleave-dev/controllers"
	"creativelab/ecleave-dev/helper"
	"time"

	. "creativelab/ecleave-dev/models"

	_ "github.com/creativelab/dbox/dbc/mongo"
	tk "github.com/creativelab/toolkit"
)

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

func GetuserData(c *BaseController) []*SysUserModel {
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

func countLeaveNewEmp(c *BaseController) interface{} {
	user := GetuserData(c)

	for _, usr := range user {
		isMyear, year, month, day := helper.IsMoreAYear(usr.JointDate)
		// tk.Println("------------- masuk sini", year)
		if isMyear == false {

			tm := time.Now()
			dt := tm.Format("2006-01-02")
			if year < 1 {

				var leave float64 = 0.0
				// for i := 0; i < month; i++ {
				if usr.YearLeave == 0 {
					tk.Println("------------ user1 ", usr.Fullname)
					tk.Println("------------ month ", month)
					tk.Println("------------ year ", year)
					tk.Println("------------ day ", day)

					if day > 0 && day < 30 {

						if usr.AddLeave == "" {
							tk.Println("------------ baru1 ")
							ms, _ := time.Parse("2006-01-02", usr.JointDate)
							tk.Println("------------ day ", ms)
							next := ms.AddDate(0, 1, 0)
							tk.Println("------------ add ", next)
							snow := next.Format("2006-01-02")
							usr.AddLeave = snow
							leave = leave + 1.5
						} else {
							if dt == usr.AddLeave {
								ms, _ := time.Parse("2006-01-02", usr.JointDate)
								next := ms.AddDate(0, 1, 0)
								snow := next.Format("2006-01-01")
								usr.AddLeave = snow
								leave = leave + 1.5
							} else {
								leave = float64(usr.YearLeave)
							}

						}

					} else {
						if dt == usr.AddLeave {
							tk.Println("------------ baru2 ")
							leave = float64(usr.YearLeave) + 1.5
							ms, _ := time.Parse("2006-01-02", usr.JointDate)
							next := ms.AddDate(0, 1, 0)
							snow := next.Format("2006-01-02")
							usr.AddLeave = snow
						} else {
							leave = float64(usr.YearLeave)
						}
					}
				} else {
					if dt == usr.AddLeave {
						leave = float64(usr.YearLeave) + 1.5
						ms, _ := time.Parse("2006-01-02", usr.JointDate)
						next := ms.AddDate(0, 1, 0)
						snow := next.Format("2006-01-02")
						usr.AddLeave = snow
					} else {
						leave = float64(usr.YearLeave)
					}

				}

				// }

				onleave := int(leave)
				usr.YearLeave = onleave
				err := c.Ctx.Save(usr)

				if err != nil {
					return err.Error()
				}
			} else if year == 1 {

				t := time.Now()
				mt := int(t.Month())
				if mt != 12 {
					tk.Println("masuk nama ", usr.Fullname)
					tk.Println("masuk hitung sini", mt)
					tk.Println("masuk hitung year", 12-mt)
					if mt < 12 {
						sisaM := 12 - mt
						if sisaM > 0 {
							sisaL := float64(sisaM) * 1.5
							resL := int(sisaL)
							usr.YearLeave = resL
							err := c.Ctx.Save(usr)

							if err != nil {
								return err.Error()
							}
						}
					}
				}

			}

		}
	}

	return ""

}
