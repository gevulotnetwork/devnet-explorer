// Package store abstracts the storage layer and provides a simple interface to work with.
package store

type Store struct{}

func New() *Store {
	return &Store{}
}

func (s *Store) Hello() string {
	return "Hello, World!"
}
