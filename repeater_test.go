package repeater

import (
	"errors"
	"testing"
	"time"
)

func TestIncreaseDelay(t *testing.T) {
	tests := []struct {
		name         string
		r            *repeaterImpl
		initial      time.Duration
		expectations []time.Duration
	}{
		{
			name:    "Default",
			r:       &defaultRepeaterImpl,
			initial: 100 * time.Millisecond,
			expectations: []time.Duration{
				200 * time.Millisecond,
				400 * time.Millisecond,
				800 * time.Millisecond,
				1600 * time.Millisecond,
				3200 * time.Millisecond,
				5 * time.Second,
				5 * time.Second,
			},
		},
		{
			name: "Linear",
			r: &repeaterImpl{
				initialRepeatTime: time.Duration(0),
				maxRepeatTime:     time.Hour,
				linearGrowth:      100 * time.Millisecond,
				exponentialGrowth: 1.0,
				maxRetries:        10,
			},
			initial: 100 * time.Millisecond,
			expectations: []time.Duration{
				200 * time.Millisecond,
				300 * time.Millisecond,
				400 * time.Millisecond,
			},
		},
		{
			name: "Linear and exponential",
			r: &repeaterImpl{
				initialRepeatTime: time.Duration(0),
				maxRepeatTime:     time.Hour,
				linearGrowth:      100 * time.Millisecond,
				exponentialGrowth: 2.0,
				maxRetries:        10,
			},
			initial: 100 * time.Millisecond,
			expectations: []time.Duration{
				400 * time.Millisecond,
				1000 * time.Millisecond,
				2200 * time.Millisecond,
			},
		},
	}

	for _, test := range tests {
		val := test.initial
		for _, expected := range test.expectations {
			val = test.r.increaseDelay(val)
			if val != expected {
				t.Errorf("%s: expected %v, saw %v", test.name, expected, val)
			}
		}

	}
}

func TestSucceedsFunction(t *testing.T) {
	repeater := Default()
	err := repeater.Repeat(func() error {
		// Do nothing and succeed
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error; error was %v", err)
	}
}

func TestRetriesOnFail(t *testing.T) {
	repeater := Default()
	loop := 0
	err := repeater.Repeat(func() error {
		if loop != 3 {
			loop += 1
			return errors.New("Not ready to break yet.")
		}
		return nil
	})
	if loop != 3 {
		t.Errorf("Expected loop to be 3; was %d", loop)
	}
	if err != nil {
		t.Errorf("Expected nil error; was %v", err)
	}
}

func TestStopsRetryingWhenThresholdExceeded(t *testing.T) {
	repeater := &repeaterImpl{
		initialRepeatTime: time.Millisecond,
		maxRepeatTime:     time.Millisecond,
		linearGrowth:      time.Duration(0),
		exponentialGrowth: 1.0,
		maxRetries:        5,
	}
	loop := 0
	err := repeater.Repeat(func() error {
		loop += 1
		return errors.New("This fails forever.")
	})

	if loop != 6 {
		t.Errorf("Expected 6 loops, saw %d", loop)
	}
	if err == nil {
		t.Errorf("Expected an error; saw none")
	} else if err.Error() != "This fails forever." {
		t.Errorf("Error message was wrong: saw '%s'", err.Error())
	}
}

// TODO: test that repeater respects backoff logic
