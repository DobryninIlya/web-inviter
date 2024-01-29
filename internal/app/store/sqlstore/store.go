package sqlstore

import (
	"database/sql"
)

// Store ...
type Store struct {
	db            *sql.DB
	apiRepository ApiRepositoryInterface
}

// New ...
func New(db *sql.DB) Store {
	return Store{
		db: db,
	}
}

func (s *Store) API() ApiRepositoryInterface {
	if s.apiRepository != nil {
		return s.apiRepository
	}

	s.apiRepository = &ApiRepository{
		store: s,
	}

	return s.apiRepository
}
