package controllers

import (
	"fmt"
	"time"

	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type WorkBackgroundController struct {
	*BaseController
}

func (m *WorkBackgroundController) Pluck(k *knot.WebContext) interface{} {
	c := new(BaseController)
	fmt.Println("Chime-")
	now := time.Now()
	fmt.Println("----------------- time now", now)
	timein := time.Now().Add(1 * time.Hour)
	fmt.Println("----------- timein ", timein)

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*RequestLeaveModel, 0)
	dbFilter = append(dbFilter, db.Eq("expiredon", now.Format("2006-01-02 15:04")))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}
	tk.Println("--------------", c)
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

	if len(data) > 0 {
		for _, dt := range data {
			if dt.ResultRequest == "Pending" {
				dt.ResultRequest = "Expired"

				err = c.Ctx.Save(dt)
				if err != nil {
					return nil
				}

				dash := DashboardController(*m)

				dash.SetHistoryLeave(k, dt.UserId, dt.Id, dt.LeaveFrom, dt.LeaveTo, "Request Expired", "Expired", dt)
			}
		}
	}
	return "success"
}

func (c *WorkBackgroundController) DoEvery10Hours(k *knot.WebContext) interface{} {
	// k.Config.OutputType = knot.OutputJson

	// pollInterval := 100

	timerCh := time.Tick(1 * time.Minute)

	for range timerCh {
		c.Pluck(k)
	}

	return "nanana"
}

var countryTz = map[string]string{
	"Hungary": "Europe/Budapest",
	"Egypt":   "Africa/Cairo",
	"Utc":     "UTC",
}

func timeIn(name string) time.Time {
	loc, err := time.LoadLocation(countryTz[name])
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}

func (c *WorkBackgroundController) CobaDateTime(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	t := time.Now()
	z, _ := t.Zone()
	fmt.Println("ZONE : ", z, " Time : ", t) // local time

	location, err := time.LoadLocation("EST")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ZONE : ", location, " Time : ", t.In(location)) // EST

	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)
	fmt.Println("ZONE : ", loc, " Time : ", now) // UTC

	loc, _ = time.LoadLocation("Asia/Jakarta")
	now = time.Now().In(loc)
	fmt.Println("ZONE : ", loc, " Time : ", now) // MST
	// utc := timeIn("Utc").Format("15:04")
	// hun := timeIn("Hungary").Format("15:04")
	// eg := timeIn("Egypt").Format("15:04")
	// fmt.Println(utc, hun, eg)
	return ""
}
