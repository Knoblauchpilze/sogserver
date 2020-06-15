package background

import (
	"fmt"
	"oglike_server/pkg/logger"
	"sync"
	"time"
)

// Process :
// Defines a process that can be started with a certain
// repeatbility and will spawn a go routine to do so.
// The function to execute is provided as input so that
// it is customizable. The user can also specify whether
// the function should be retried in case of a failure.
//
// The `interval` defines the duration between two calls
// of the function by this process.
//
// The `retryInterval` defines the interval to wait in
// case the `operation` fails. The default value is `1``
// second.
//
// The `operation` defines the function to be executed
// by the process.
//
// The `retry` defines whether the operation should be
// rescheduled immediately in case it fails.
//
// The `log` defines a way for this process to notify
// information and failures to the user.
//
// The `module` defines a string identifying the func
// attached to this process to make logs more relevant.
//
// The `lock` allows to protect concurrent accesses
// to some internal variables.
//
// The `running` defines whether or not the main
// processing loop is running.
//
// The `termination` is a channel used to terminate
// the execution of the main processing loop.
//
// The `waiter` allows to wait for this process to
// complete before returning from the `Stop` func.
type Process struct {
	interval      time.Duration
	retryInterval time.Duration
	operation     OperationFunc
	retry         bool
	log           logger.Logger
	module        string

	lock        sync.Mutex
	running     bool
	termination chan bool
	waiter      sync.WaitGroup
}

// OperationFunc :
// Defines an operation that can be associated to a
// process object. It should take no argument and
// return any error along with a status indicating
// whether it could be executed successfully.
type OperationFunc func() (bool, error)

// ErrAlreadyRunning : Indicates that this process is
// already running and cannot be started again.
var ErrAlreadyRunning = fmt.Errorf("Unable to start already running process")

// ErrInvalidOperation : Indicates that the operation
// associated to this process is not valid.
var ErrInvalidOperation = fmt.Errorf("Invalid operation to start process")

// NewProcess :
// Defines a new process object with the specified
// interval and logger.
//
// The `interval` defines the time interval between
// two consecutive calls to the main process func.
//
// The `log` defines the logger to use to notify
// info and errors.
//
// Returns the built-in object.
func NewProcess(interval time.Duration, log logger.Logger) *Process {
	return &Process{
		interval:      interval,
		retryInterval: 1 * time.Second,
		retry:         false,
		log:           log,

		lock:        sync.Mutex{},
		running:     false,
		termination: make(chan bool, 1),
	}
}

// WithModule :
// Assigns a new string as the module name for this
// process.
//
// The `module` defines the name of the module to
// assign to this object.
//
// Returns this process to allow chain calling.
func (p *Process) WithModule(module string) *Process {
	// Make sure that we're the only process changing
	// this value.
	func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		p.module = module
	}()

	return p
}

// WithRetry :
// Defines that this process should try to schedule
// the operation function if it fails during its
// first execution until it succeeds.
//
// Returns this process to allow chain calling.
func (p *Process) WithRetry() *Process {
	// Make sure that we're the only process changing
	// this value.
	func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		p.retry = true
	}()

	return p
}

// WithRetryInterval :
// Defines a new retry interval for the time to
// wait when the main operation fails to execute.
//
// The `interval` defines the retry interval.
//
// Returns this process to allow chain calling.
func (p *Process) WithRetryInterval(interval time.Duration) *Process {
	// Make sure that we're the only process changing
	// this value.
	func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		p.retryInterval = interval
	}()

	return p
}

// WithOperation :
// Defines the core processing function to execute
// when needed.
//
// The `operation` defines the processing function
// to execute at each interval.
//
// Returns this process to allow chain calling.
func (p *Process) WithOperation(operation OperationFunc) *Process {
	// Make sure that we're the only process changing
	// this value.
	func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		p.operation = operation
	}()

	return p
}

// Stop :
// Used to indicate the termination of the active
// loop for this process. It is used to prevent
// any further execution of the main operation
// callback.
func (p *Process) Stop() {
	// Make sure that this process is started.
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.running {
		return
	}

	// The process is running, stop it.
	p.termination <- true

	// And wait for the process to terminate.
	p.waiter.Wait()
}

// Start :
// Used to start the process associated with
// this object. Note that we will check that
// the operation is valid otherwise an error
// is returned.
//
// Returns any error.
func (p *Process) Start() error {
	// Make sure that the operation to perform
	// is valid.
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.running {
		return ErrAlreadyRunning
	}
	if p.operation == nil {
		return ErrInvalidOperation
	}

	p.running = true
	p.waiter.Add(1)

	go p.activeLoop()

	return nil
}

// activeLoop :
// Main processing loop for this object. It
// will sleep for the required period of time
// and execute the attached operation.
func (p *Process) activeLoop() {
	// Create the timer.
	timer := time.NewTimer(p.interval)

	// Prevent errors.
	defer func() {
		err := recover()
		if err != nil {
			func() {
				p.lock.Lock()
				defer p.lock.Unlock()

				p.log.Trace(logger.Critical, p.module, fmt.Sprintf("Recovered from error in process (err: %v)", err))
			}()
		}

		// The process is not running anymore.
		p.lock.Lock()
		p.running = false
		p.lock.Unlock()

		// Release the wait group.
		p.waiter.Done()
	}()

	connected := true

	// While we're still askec to continue the
	// main operation.
	for connected {
		// Select from either the termination channel
		// or from the timer.
		select {
		case connected = <-p.termination:
			// Termination requested.
			break
		case <-timer.C:
			err := p.execute()
			if err != nil {
				func() {
					p.lock.Lock()
					defer p.lock.Unlock()

					p.log.Trace(logger.Critical, p.module, fmt.Sprintf("Caught error while executing process (err: %v)", err))
				}()
			}
		}

		// Update the connected status.
		if connected {
			func() {
				p.lock.Lock()
				defer p.lock.Unlock()

				connected = p.running
			}()
		}
	}
}

// execute :
// Wrapper function allowing to execute the main
// operation binded to this process. The process
// will be retried as long as it does not succeed
// based on the internal flag.
//
// Returns any error.
func (p *Process) execute() error {
	// Perform the operation until we succeed.
	success := false
	var err error

	for !success {
		func() {
			p.lock.Lock()
			defer p.lock.Unlock()

			p.log.Trace(logger.Verbose, p.module, fmt.Sprintf("Executing process"))

			// Perform the operation.
			success, err = p.operation()

			if err != nil {
				p.log.Trace(logger.Error, p.module, fmt.Sprintf("Caught error while executing process (err: %v)", err))
			}
		}()

		// Override the success in case the operation
		// failed and the retry flag is not set.
		if p.retry && !success {
			// Wait for a certain amount of time.
			var wait time.Duration
			func() {
				p.lock.Lock()
				defer p.lock.Unlock()

				wait = p.retryInterval

				p.log.Trace(logger.Verbose, p.module, fmt.Sprintf("Failed to execute process, retrying in %v", wait))
			}()

			time.Sleep(wait)
		}

		if !p.retry {
			success = true
		}
	}

	return err
}
