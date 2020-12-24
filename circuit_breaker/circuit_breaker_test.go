package circuit_breaker

import (
	"errors"
	"testing"
	"time"
)

func TestSuccessUnderFailedThreshold(t *testing.T) {

	cb := GetInstance()
	command := Command{
		FailedThreshold: 5,
		TimeWithin:      time.Duration(60 * time.Second),
	}

	cb.SetCommand("mycommand", command)

	for i := 0; i < 5; i++ {
		res := cb.Go("mycommand", func() error {
			return errors.New("Test Failed")
		})

		if res != true {
			t.Fail()
		}
	}

}

func TestFailedOverFailedThreshold(t *testing.T) {

	cb := GetInstance()
	command := Command{
		FailedThreshold: 5,
		TimeWithin:      time.Duration(60 * time.Second),
	}

	cb.SetCommand("mycommand", command)

	for i := 0; i < 10; i++ {
		res := cb.Go("mycommand", func() error {
			return errors.New("Test Failed")
		})

		if i > 5 && res != false {
			t.Fail()
		}
	}

}
