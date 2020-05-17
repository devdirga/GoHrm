package controllers

import (
	//"time"

	"github.com/creativelab/knot/knot.v1"
)

type CobaController struct {
	*BaseController
}

func (c *CobaController) Default(k *knot.WebContext) interface{} {
	// fmt.Println("masuk")
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

	// if access != nil {
	// 	tk.Println("sdsfsdf")
	// 	e := tk.Serde(access, DataAccess, "json")
	// 	if e != nil {
	// 		tk.Println(e.Error(), "<<")
	// 	}
	// }

	return DataAccess
}
