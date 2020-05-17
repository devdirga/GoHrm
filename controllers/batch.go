package controllers

import (
	"creativelab/ecleave-dev/batch"

	knot "github.com/creativelab/knot/knot.v1"
)

type BatchController struct {
	*BaseController
}

func (c *BatchController) Remote(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	return new(batch.RemoteBatch).FixingDataRemote()
}

func (c *BatchController) Leave(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	return new(batch.LeaveBatch).FixingDataLeave()
}
