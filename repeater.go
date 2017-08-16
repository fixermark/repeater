// Simple repetition wrapper: runs an operation indefinitely, optionally with backoff.

package repeater

import (
	"math"
	"time"
)

type Repeatable func() error

type Repeater interface {
	// Execute Repeatable, and if it returns a non-nil error, repeat
	// execution. If the threshold of repetitions is exceeded, Repeat
	// returns the last error sent.
	Repeat(r Repeatable) error
}

// TODO(mtomczak): definitely tests in this one. ;) Also, constructors for a
// basic exponential (linear component, exponential component) and a const
// default exponential (Default)
//
// Lets use time.Durations for the input values, since time.Sleep uses Duration
// as input

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

func Default() Repeater {
	return &defaultRepeaterImpl
}

func DefaultInfinite() Repeater {
	return &defaultInfiniteRepeaterImpl
}

func NewInfiniteRepeater(
	initial time.Duration,
	max time.Duration,
	linear time.Duration,
	exponential float64) Repeater {
	return NewRepeater(initial, max, linear, exponential, 0)
}

func NewRepeater(
	initial time.Duration,
	max time.Duration,
	linear time.Duration,
	exponential float64,
	maxRetries int) Repeater {
	return &repeaterImpl{initial, max, linear, exponential, maxRetries}
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
