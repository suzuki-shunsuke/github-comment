package domain

type ComplementEntry interface {
	Entry() (string, error)
	Type() string
}

type ComplementWithNameEntry interface {
	Entry() (string, error)
	Type() string
}
