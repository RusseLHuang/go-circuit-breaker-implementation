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
		RefreshInterval: time.Duration(10 * time.Second),
	}

	cb.SetCommand("mycommand", command)

	for i := 0; i < 5; i++ {
		res := cb.Go("mycommand", func(nums ...*int) error {
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
		RefreshInterval: time.Duration(10 * time.Second),
	}

	funcCalled := 0

	cb.SetCommand("mycommand", command)

	for i := 0; i < 10; i++ {
		res := cb.Go("mycommand", func(num ...*int) error {
			funcCalled++
			return errors.New("Test Failed")
		})

		if i > 5 && res != false {
			t.Fail()
		}
	}

	if funcCalled != 5 {
		t.Fail()
	}

}

func TestSuccessOverRefreshInterval(t *testing.T) {

	cb := GetInstance()
	refreshInterval := time.Duration(3 * time.Second)
	command := NewCommand(
		1,
		time.Duration(60*time.Second),
		refreshInterval,
	)

	cb.SetCommand("mycommand", command)

	res1 := cb.Go("mycommand", func(num ...*int) error {
		return errors.New("Test Failed")
	})

	if res1 != true {
		t.Fatalf("Should failed since request handler return error")
	}

	res2 := cb.Go("mycommand", func(num ...*int) error {
		return errors.New("Test Failed")
	})

	if res2 != false {
		t.Fatalf("Should failed after failed threshold has been reached")
	}

	time.Sleep(time.Duration(4 * time.Second))

	normalFnCall := 0
	res3 := cb.Go("mycommand", func(num ...*int) error {
		normalFnCall++
		return nil
	})

	if res3 != true && normalFnCall != 1 {
		t.Fatalf("Should have been refresh")
	}

}
