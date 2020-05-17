package controllers

import (
	knot "github.com/creativelab/knot/knot.v1"
)

type AdminReportController struct {
	*BaseController
}

func (c *AdminReportController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputTemplate

	DataAccess := c.SetViewData(k, nil)

	return DataAccess
}
