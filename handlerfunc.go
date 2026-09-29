package aio

import "context"

type HandlerFunc[In, Out any] func(context.Context, In) (Out, error)
