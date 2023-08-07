package bff

import "context"

type Bff struct {
	caculator *Caculator
}

func NewBff(caculator *Caculator) *Bff {
	return &Bff{caculator: caculator}
}

func (c *Bff) Fibo(ctx context.Context, n int) (int, error) {
	return c.caculator.Fibo(ctx, n)
}
