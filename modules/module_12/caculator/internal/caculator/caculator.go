package caculator

import "context"

type Caculator struct {
	fibo *Fibo
}

func NewCaculator(fibo *Fibo) *Caculator {
	return &Caculator{fibo: fibo}
}

func (c *Caculator) Fibo(ctx context.Context, n int) (int, error) {
	return c.fibo.Fibo(ctx, n)
}
