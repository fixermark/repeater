// Package repeater provides a simple repetition wrapper that retries a fallible
// operation with backoff. The operation can be retried a finite amount of times
// or indefinitely.
package repeater

import (
	"math"
	"time"
)

// A repeatable function. If the function returns an error and the retry
// thresholds are not exceeded, it will be tried again.
type Repeatable func() error

// A repeater, encapsulating the logic for retrying a Repeatable operation.
type Repeater interface {
	// Execute Repeatable, and if it returns a non-nil error, repeat
	// execution. If the threshold of repetitions is exceeded, Repeat
	// returns the last error sent.
	Repeat(r Repeatable) error
}

type repeaterImpl struct {
	initialRepeatTime time.Duration
	maxRepeatTime     time.Duration
	linearGrowth      time.Duration
	exponentialGrowth float64
	// If 0, retry indefinitely
	maxRetries int
}

var defaultRepeaterImpl = repeaterImpl{
	initialRepeatTime: time.Duration(100) * time.Millisecond,
	maxRepeatTime:     time.Duration(5) * time.Second,
	linearGrowth:      time.Duration(0),
	exponentialGrowth: 2.0,
	maxRetries:        10,
}

var defaultInfiniteRepeaterImpl = repeaterImpl{
	initialRepeatTime: time.Duration(100) * time.Millisecond,
	maxRepeatTime:     time.Duration(5) * time.Second,
	linearGrowth:      time.Duration(0),
	exponentialGrowth: 2.0,
	maxRetries:        0,
}

// Get a default repeater (100ms initial repeat time, each retry 2x time after
// previous, fail after 10 retries).
func Default() Repeater {
	return &defaultRepeaterImpl
}

// Get a default repeater that retries indefinitely.
func DefaultInfinite() Repeater {
	return &defaultInfiniteRepeaterImpl
}

// Get a repeater with the specified initial retry delay, maximum retry delay,
// and max retries. Delay expansion is calculated as (new delay = (old delay +
// linear) * exponential). If 0 is specified for maxRetries, retry indefinitely.
func NewRepeater(
	initial time.Duration,
	max time.Duration,
	linear time.Duration,
	exponential float64,
	maxRetries int) Repeater {
	return &repeaterImpl{initial, max, linear, exponential, maxRetries}
}

// Get a new repeater with the specified parameters that retries
// indefinitely. See NewRepeater for details.
func NewInfiniteRepeater(
	initial time.Duration,
	max time.Duration,
	linear time.Duration,
	exponential float64) Repeater {
	return NewRepeater(initial, max, linear, exponential, 0)
}

func (r *repeaterImpl) Repeat(do Repeatable) error {
	err := do()
	repetitions := 0
	if err == nil {
		return nil
	}
	repetitions += 1
	time.Sleep(r.initialRepeatTime)
	err = do()
	if err == nil {
		return nil
	}
	if r.maxRetries == 1 {
		return err
	}

	delay := r.initialRepeatTime
	for err != nil {
		delay = r.increaseDelay(delay)
		time.Sleep(delay)
		err = do()
		if err != nil && r.maxRetries != 0 {
			repetitions += 1
			if repetitions >= r.maxRetries {
				return err
			}
		}
	}
	return nil
}

func (r *repeaterImpl) increaseDelay(d time.Duration) time.Duration {
	return time.Duration(math.Min(
		float64(r.maxRepeatTime),
		r.exponentialGrowth*float64(
			d+r.linearGrowth)))
}
