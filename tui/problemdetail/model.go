package problemdetail

import "github.com/ARJ2211/cpgrinder/internal/store"

type ProblemDetailModel struct {
	dbStore   *store.Store
	problemID string
	rawMD     string
}

// Initialize the model with the raw markdown
func InitializeModel(
	id string,
	dbStore *store.Store,
) (*ProblemDetailModel, error) {
	problemDetail, err := dbStore.GetProblemByID(id)
	if err != nil {
		return nil, err
	}

	m := ProblemDetailModel{
		dbStore:   dbStore,
		problemID: id,
		rawMD:     problemDetail.StatementMd,
	}

	return &m, nil
}
