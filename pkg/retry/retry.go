package retry

import (
	"Alice088/essentia/pkg/time"
	"context"

	"github.com/sethvargo/go-retry"
)

type ExponentialOpts struct {
	Seconds int
	Tries   uint64
	Fn      func(ctx context.Context) error
}

func Exponential(ctx context.Context, opts ExponentialOpts) error {
	return retry.Do(ctx,
		retry.WithMaxRetries(opts.Tries,
			retry.NewExponential(time.Seconds(opts.Seconds)),
		),
		opts.Fn,
	)
}
