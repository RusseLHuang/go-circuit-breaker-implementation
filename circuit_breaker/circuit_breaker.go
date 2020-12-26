package circuit_breaker

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const Closed = "Closed"
const Open = "Open"
const HalfOpen = "HalfOpen"

type CircuitBreaker struct {
	Command map[string]*Command
}

type Command struct {
	FailedCount      int32
	FailedThreshold  int32
	TimeStart        time.Time
	TimeWithin       time.Duration
	RefreshInterval  time.Duration
	HalfOpenRequest  int32
	RequestHalfOpen  chan bool
	HalfOpenResponse chan bool
	FallbackHandler  func()
	Status           string
	StatusLocker     *sync.Mutex
}

func (c *Command) Refresh() {
	c.TimeStart = time.Now()
	atomic.StoreInt32(&c.FailedCount, 0)
}

func (c *Command) IsTimeToRefresh() bool {
	return time.Now().Unix() > c.TimeStart.Add(c.TimeWithin).Unix()
}

func (c *Command) GetCurrentFailedCount() int32 {
	return atomic.LoadInt32(&c.FailedCount)
}

func (c *CircuitBreaker) SetCommand(
	commandName string,
	command Command,
) {
	c.Command[commandName] = &command
}

func (c *CircuitBreaker) Go(
	commandName string,
	commandHandler func(num ...*int) error,
) bool {
	command, ok := c.Command[commandName]
	command.StatusLocker.Lock()
	defer command.StatusLocker.Unlock()

	if ok != true {
		log.Println("Command not found")
		return false
	}

	if command.IsTimeToRefresh() {
		command.Refresh()
	}

	var status string
	select {
	case <-command.RequestHalfOpen:
		status = HalfOpen
	default:
		status = command.Status
	}

	if status == Open {
		log.Println("Failed Threshold reached")

		if command.FallbackHandler != nil {
			command.FallbackHandler()
		}

		return false
	}

	err := commandHandler()

	if err != nil {
		log.Println("Err nil")
		atomic.AddInt32(&(command.FailedCount), 1)
	}

	if status == HalfOpen {
		if err != nil {
			command.HalfOpenResponse <- false
		} else {
			command.HalfOpenResponse <- true
		}

		return false
	}

	if atomic.LoadInt32(&command.FailedCount) >= command.FailedThreshold {
		log.Println("Circuit Breaker Opened")
		command.Status = Open
		ticker := time.NewTicker(command.RefreshInterval)
		go initHalfOpenHandler(ticker, command)
	}

	return true
}

func initHalfOpenHandler(ticker *time.Ticker, command *Command) {
	for {
		<-ticker.C

		log.Println("Half open ticker tick")

		command.RequestHalfOpen <- true

		isCloseCircuitBreaker := <-command.HalfOpenResponse

		if isCloseCircuitBreaker == true {
			log.Println("Circuit Breaker Closed")

			command.StatusLocker.Lock()
			command.Status = Closed
			atomic.StoreInt32(&command.FailedCount, 0)
			command.StatusLocker.Unlock()
			return
		}
	}
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

func NewCommand(
	failedThrehold int32,
	timeWithin time.Duration,
	refreshInterval time.Duration,
) Command {
	return Command{
		FailedThreshold:  failedThrehold,
		TimeWithin:       timeWithin,
		RefreshInterval:  refreshInterval,
		RequestHalfOpen:  make(chan bool),
		HalfOpenResponse: make(chan bool),
		StatusLocker:     &sync.Mutex{},
	}
}
