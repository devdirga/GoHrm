package services

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
)

type ProjectService struct{}

func (s *ProjectService) GetByID(id string) (ProjectModel, error) {
	filter := tk.M{}
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("_id", id))
	filter.Set("where", dbox.And(dboxFilter...))

	row := ProjectModel{}
	projects, err := new(repositories.ProjectOrmRepo).GetByParam(filter)
	if err != nil {
		return row, err
	}

	if len(projects) > 0 {
		row = projects[0]
	}

	return row, nil
}

func (s *ProjectService) Delete(id string) error {
	project, err := s.GetByID(id)
	if err != nil {
		return err
	}

	err = new(repositories.ProjectOrmRepo).Delete(&project)
	return err
}
