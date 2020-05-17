package repositories

import (
	. "creativelab/ecleave-dev/models"
)

type LogCronRepo struct{}

func (r *LogCronRepo) Insert(m *LogCronModel) error {
	err := Ctx.Insert(m)
	return err
}
