package controllers

import (
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

// . "creativelab/ecleave-dev/models"

type LogForAdminController struct {
	*BaseController
}

func (c *LogForAdminController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	viewData := tk.M{}
	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel").(int))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
	}

	DataAccess := c.SetViewData(k, viewData)
	// k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = []string{
		"_modal.html",
		"_loader.html",
	}
	// DataAccess := Previlege{}

	// for _, o := range access {
	// 	DataAccess.Create = o["Create"].(bool)
	// 	DataAccess.View = o["View"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Process = o["Process"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Edit = o["Edit"].(bool)
	// 	DataAccess.Menuid = o["Menuid"].(string)
	// 	DataAccess.Menuname = o["Menuname"].(string)
	// 	DataAccess.Approve = o["Approve"].(bool)
	// 	DataAccess.Username = o["Username"].(string)
	// }

	return DataAccess

}
