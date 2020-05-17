package batch

import (
	"creativelab/ecleave-dev/repositories"
	"time"
)

type RemoteBatch struct{}

func (r *RemoteBatch) FixingDataRemote() error {
	repoOrm := new(repositories.RemoteOrmRepo)
	remotes, err := repoOrm.GetAll()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		remote.CreatedAt = time.Now()
		remote.UpdatedAt = time.Now()

		err := repoOrm.Save(&remote)
		if err != nil {
			return err
		}
	}

	return nil
}
