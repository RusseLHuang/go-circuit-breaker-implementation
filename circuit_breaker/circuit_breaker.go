package circuit_breaker

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type CircuitBreaker struct {
	Command map[string]*Command
}

type Command struct {
	Request         int
	FailedCount     int32
	FailedThreshold int32
	TimeStart       time.Time
	TimeWithin      time.Duration
	FallbackHandler func()
}

func (c *CircuitBreaker) SetCommand(
	commandName string,
	command Command,
) {

	c.Command[commandName] = &command
}

func (c *CircuitBreaker) Go(
	commandName string,
	externalFn func() error,
) bool {
	command, ok := c.Command[commandName]

	if ok != true {
		log.Println("Command not found")
		return false
	}

	err := externalFn()

	if err != nil {
		log.Println("Err nil")
		atomic.AddInt32(&(c.Command[commandName].FailedCount), 1)
	}

	if atomic.LoadInt32(&command.FailedCount) > command.FailedThreshold {
		log.Println("Failed Threshold reached")

		if command.FallbackHandler != nil {
			command.FallbackHandler()
		}

		return false
	}

	return true
}

var circuitBreaker *CircuitBreaker
var once sync.Once

func GetInstance() *CircuitBreaker {
	once.Do(func() {
		circuitBreaker = &CircuitBreaker{
			Command: make(map[string]*Command),
		}
	})
	return circuitBreaker
}
