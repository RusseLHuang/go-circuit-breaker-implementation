# go-circuit-breaker-implementation

## General Idea
Inspired by Netflix Hystrix library to stop cascading failures from external services or library failures. This repository intended to make a simple implementation of circuit breaker pattern.

## General Concept
Circuit Breaker Pattern avoid unecessary external calls when its know it is gonna fail. Most of the system will perform retry when external services fail. Circuit Breaker allow us to have these benefits.

- Avoid cascading failures
- Save resource consumption (Allow threads not doing unecessary works)
- Better user experience (In some business logic)

## How it works

![Circuit Breaker Pattern](https://martinfowler.com/bliki/images/circuitBreaker/sketch.png)

There are three state:
- Closed: Allow request to go to external services, after a certain failed threshold is reached, it will switch to Open state.
- Open: Request will not go to external services, instead will invoke expected behaviour either throw error immediately or perform another logic.
- Half-Open: After certain time, state will switch to this to allow some of request to external services to check if the external services still failed or not. If it is failed, switch to Open state otherwise to Closed state.

![Circuit Breaker State](https://martinfowler.com/bliki/images/circuitBreaker/state.png)

References:

https://martinfowler.com/bliki/CircuitBreaker.html

