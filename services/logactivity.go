package services

import (
	md "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
)

// LogActivities ...
type LogActivities struct{}

// SaveLog ...
func (l *LogActivities) SaveLog(logs md.LogCronModel) {
	_ = new(repositories.LogCronRepo).Insert(&logs)
}
