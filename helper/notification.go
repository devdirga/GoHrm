package helper

import (
	. "creativelab/ecleave-dev/models"

	"gopkg.in/mgo.v2/bson"
)

func InsertNotification(data NotificationModel) bool {
	if data.Id == "" {
		data.Id = bson.NewObjectId().Hex()
	}

	return false

}
