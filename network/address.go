package network

type Address struct {
	s string
}

func NewAddress(s string) Address {
	return Address{s}
}

func (a Address) String() string {
	return a.s
}

func (a Address) IsZero() bool {
	return a == Address{}
}
