package knot

import "github.com/creativelab/toolkit"

type sharedObject struct {
	toolkit.M
}

var instance *sharedObject

func SharedObject() *sharedObject {
	if instance == nil {
		instance = &sharedObject{M: toolkit.M{}}
	}

	return instance
}
