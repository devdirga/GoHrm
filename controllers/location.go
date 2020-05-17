package controllers

import (
	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"

	"github.com/creativelab/knot/knot.v1"
)

type LocationController struct {
	*BaseController
}

func (c *LocationController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	dataLocation := make([]LocationModel, 0)
	crsLocation, errLocation := c.Ctx.Find(NewLocationModel(), nil)

	if crsLocation != nil {
		defer crsLocation.Close()
	} else if crsLocation == nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}
	defer crsLocation.Close()
	if errLocation != nil {
		return c.SetResultInfo(true, errLocation.Error(), nil)
	}

	errLocation = crsLocation.Fetch(&dataLocation, 0, false)
	if errLocation != nil {
		return c.SetResultInfo(true, errLocation.Error(), nil)
	}

	return dataLocation
}

func (c *LocationController) GetLocationParam(k *knot.WebContext, Location string) *LocationModel {
	k.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	query := tk.M{}
	dbFilter = append(dbFilter, db.Eq("Location", Location))
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	dataLocation := new(LocationModel)
	crsLocation, errLocation := c.Ctx.Find(NewLocationModel(), query)

	if crsLocation != nil {
		defer crsLocation.Close()
	} else if crsLocation == nil {
		return dataLocation
	}
	defer crsLocation.Close()
	if errLocation != nil {
		return dataLocation
	}

	errLocation = crsLocation.Fetch(&dataLocation, 1, false)
	if errLocation != nil {
		return dataLocation
	}

	return dataLocation
}

func (c *LocationController) LocationUserid(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	userid := k.Session("userid").(string)
	level := k.Session("jobrolelevel").(int)
	locat := k.Session("location").(string)

	loc := []string{}

	if level != 5 {
		dash := DashboardController(*c)
		data := dash.GetDataSessionUser(k, userid)
		loc = append(loc, data[0].Location)
		return loc
	} else if locat != "Global" {
		dash := DashboardController(*c)
		data := dash.GetDataSessionUser(k, userid)
		loc = append(loc, data[0].Location)
		return loc
	}

	dataLocation := make([]LocationModel, 0)
	crsLocation, errLocation := c.Ctx.Find(NewLocationModel(), nil)

	if crsLocation != nil {
		defer crsLocation.Close()
	} else if crsLocation == nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}
	defer crsLocation.Close()
	if errLocation != nil {
		return c.SetResultInfo(true, errLocation.Error(), nil)
	}

	errLocation = crsLocation.Fetch(&dataLocation, 0, false)
	if errLocation != nil {
		return c.SetResultInfo(true, errLocation.Error(), nil)
	}

	for _, dt := range dataLocation {
		loc = append(loc, dt.Location)
	}

	return loc

}
