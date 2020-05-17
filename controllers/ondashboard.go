package controllers

import (
	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type OnDashboardController struct {
	*BaseController
}

func (c *OnDashboardController) GetDataUser(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	// fmt.Println("----------------masuk")
	p := k.Session("userid")
	data := make([]*SysUserModel, 0)
	var dbFilter []*db.Filter
	if p != nil {

		dbFilter = append(dbFilter, db.Eq("_id", p))

		query := tk.M{}

		if len(dbFilter) > 0 {
			query.Set("where", db.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewSysUserModel(), query)
		defer crs.Close()
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
	}

	return data
}
