package controllers

import (
	// . "creativelab/ecleave-dev/models"

	// db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	// tk "github.com/creativelab/toolkit"
	// tk "github.com/creativelab/toolkit"
)

type PageAdminController struct {
	*BaseController
}

func (c *PageAdminController) Default(k *knot.WebContext) interface{} {
	access := c.LoadBase(k)
	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	DataAccess := Previlege{}

	for _, o := range access {
		DataAccess.Create = o["Create"].(bool)
		DataAccess.View = o["View"].(bool)
		DataAccess.Delete = o["Delete"].(bool)
		DataAccess.Process = o["Process"].(bool)
		DataAccess.Delete = o["Delete"].(bool)
		DataAccess.Edit = o["Edit"].(bool)
		DataAccess.Menuid = o["Menuid"].(string)
		DataAccess.Menuname = o["Menuname"].(string)
		DataAccess.Approve = o["Approve"].(bool)
		DataAccess.Username = o["Username"].(string)
	}

	return DataAccess
}
