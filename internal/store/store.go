package store

type Store[T any] interface {
	Load() (*T, error)
	Save(*T) error
}
