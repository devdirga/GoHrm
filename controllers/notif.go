package controllers

import (
	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

func (c *NotificationController) InsertNotification(data NotificationModel) bool {
	tk.Println("------------------->>> save ", data.Id)
	tk.Println("------------------->>> save ", data.UserId)
	if data.Id == "" {
		data.Id = bson.NewObjectId().Hex()
	}

	err := c.Ctx.Save(&data)

	if err != nil {
		return true
	}

	return false

}

func (c *NotificationController) GetDataNotification(k *knot.WebContext, idrequest string) NotificationModel {
	k.Config.OutputType = knot.OutputJson
	var dbFilter []*db.Filter
	query := tk.M{}
	data := []NotificationModel{}

	tk.Println("------------------>>>>>", idrequest)

	dbFilter = append(dbFilter, db.Eq("idrequest", idrequest))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewNotificationModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return data[0]
	}
	// defer crs.Close()
	if err != nil {
		return data[0]
	}

	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return data[0]
	}

	tk.Println("-------------- >>> data ", data[0])

	return data[0]

}
