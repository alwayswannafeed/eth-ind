package data

type Storage interface {
	Transfers() TransfersQ
	State() StateQ

	Transaction(fn func(Storage) error) error
}