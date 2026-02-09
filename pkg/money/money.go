package money

import "fmt"

type Amount int64

func FromInt(value int64) Amount {
	return Amount(value)
}

func FromMajor(major float64) Amount {
	return Amount(major * 100)
}

func (a Amount) ToMajor() float64 {
	return float64(a) / 100
}

func (a Amount) String() string {
	return fmt.Sprintf("%.2f", a.ToMajor())
}

func (a Amount) Add(b Amount) Amount {
	return a + b
}

func (a Amount) Sub(b Amount) Amount {
	return a - b
}

func (a Amount) IsPositive() bool {
	return a > 0
}

func (a Amount) IsNegative() bool {
	return a < 0
}

func (a Amount) GreaterThanOrEqual(b Amount) bool {
	return a >= b
}
