package controllers

import (
	"github.com/creativelab/knot/knot.v1"
)

type LogoutController struct {
	*BaseController
}

func (c *LogoutController) Do(k *knot.WebContext) interface{} {
	k.SetSession("userid", nil)
	k.SetSession("username", nil)
	k.SetSession("usermodel", nil)
	c.Redirect(k, "login", "default")
	return ""
}
