package circuit_breaker

import (
	"errors"
	"testing"
	"time"
)

func TestSuccessUnderFailedThreshold(t *testing.T) {

	cb := GetInstance()
	command := NewCommand(
		5,
		time.Duration(60*time.Second),
		time.Duration(10*time.Second),
	)

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
	command := NewCommand(
		5,
		time.Duration(60*time.Second),
		time.Duration(10*time.Second),
	)

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

func TestCommandHandlerInvokedAfterCircuitBreakerClosed(t *testing.T) {

	cb := GetInstance()
	command := NewCommand(
		2,
		time.Duration(3*time.Second),
		time.Duration(3*time.Second),
	)

	cb.SetCommand("mycommand", command)

	_ = cb.Go("mycommand", func(num ...*int) error {
		return errors.New("Test Failed")
	})

	_ = cb.Go("mycommand", func(num ...*int) error {
		return errors.New("Test Failed")
	})

	failedCount := cb.Command["mycommand"].GetCurrentFailedCount()
	if failedCount != 2 {
		t.Fatalf("Failed count should be two when previous two request is failed")
	}

	time.Sleep(time.Duration(4 * time.Second))

	_ = cb.Go("mycommand", func(num ...*int) error {
		return nil
	})

	time.Sleep(time.Duration(1 * time.Second))

	functionCalled := 0
	failedCountAfterRefresh := cb.Command["mycommand"].GetCurrentFailedCount()
	_ = cb.Go("mycommand", func(num ...*int) error {
		functionCalled++
		return nil
	})

	if functionCalled != 1 {
		t.Fatalf("Command Handler should be called after circuit breaker is closed")
	}
	if failedCountAfterRefresh != 0 {
		t.Fatalf("Command Failed Count should refresh after circuit breaker is closed")
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

func TestHalfOpenConcurrency(t *testing.T) {

	cb := GetInstance()
	refreshInterval := time.Duration(3 * time.Second)
	command := NewCommand(
		1,
		time.Duration(60*time.Second),
		refreshInterval,
	)

	cb.SetCommand("mycommand", command)

	res1 := make(chan bool)
	res2 := make(chan bool)

	_ = cb.Go("mycommand", func(num ...*int) error {
		return errors.New("Test Failed")
	})

	functionCall := 0

	time.Sleep(time.Duration(4 * time.Second))

	go func() {
		commandResponse := cb.Go("mycommand", func(num ...*int) error {
			functionCall++
			return nil
		})

		res1 <- commandResponse
	}()

	go func() {
		commandResponse := cb.Go("mycommand", func(num ...*int) error {
			functionCall++
			return nil
		})

		res2 <- commandResponse
	}()

	<-res1
	<-res2

	if functionCall > 1 {
		t.Fatalf("Should only one request allowed to check external function is good or not")
	}

}
