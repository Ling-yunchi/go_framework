package session

import "weorm/log"

func (s *Session) Begin() (err error) {
	log.Info("Transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("Transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
		return
	}
	return
}

func (s *Session) Rollback() (err error) {
	log.Info("Transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
		return
	}
	return
}
