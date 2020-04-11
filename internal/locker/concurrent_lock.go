package locker

import (
	"fmt"
	"oglike_server/pkg/logger"
	"sync"

	"github.com/spf13/viper"
)

// ConcurrentLocker :
// Used to provide a concurrent lock mechanism allowing to
// share the access to the resources of a table and allow
// multiple users to wait on a shared resources while still
// providing individual locks.
// Typically this lock is designed to be used when updating
// the upgrade actions defined by a user. In this case, the
// user will have a set of actions that might or might not
// be completed upon processing a request. In order to have
// up-to-date information, a single process should perform
// the update of the upgrade action to determine whether
// some of them have finished or not. Any attempt to fetch
// the associated data should be blocked during this time
// in order not to mess with the update.
// This suggests using a mutex to prevent concurrent uses
// of the resources. But we don't want to lock the entire
// table for the update of a single element: most of the
// time indeed we will update only the data linked to a
// single player and we would like to be able to serve
// other clients as usual.
// We could create a locker per row in the DB but it's
// probably not the ideal solution. So we figured it is
// possible to define a certain number of locks (precise
// number can be configured through the configuration)
// and associate each lock with a particular row. If some
// request needs to verify whether some update actions
// for a specific row have completed, it first needs to
// acquire a lock among the pool, associate it to the
// element that needs to be updated (so that other users
// that need to access it will know that someone has
// already accesses the data) and then perform the update.
// As soon as the update is performed the lock is released
// and other processes can lock it again to perform the
// updates and make sure that nothing is left to do.
//
// The `locker` is the top level mutex that allows to use
// this object concurrently without losing thread safety.
// It is used by most facets to ensure that the data is not
// accessed concurrently.
//
// The `locks` defines a slice of locks that can be used
// to protect the concurrent access to a particular res.
// There are only a finite number of them and once all of
// them are used a call to the `Acquire` method becomes
// blocking.
// A lock can be used to protect the concurrent access
// to a resource and allow other element to wait on the
// resource to make sure that only a single client is
// using the resource at any given moment.
//
// The `availableLocks` is used internally to determine
// which of the locks are available and which one are
// already distributed to clients. This is used by the
// `Acquire` method to determine if it is possible to
// assign a lock to protect a new resource or not. The
// user should release a lock by the `Relase` method
// otherwise we would use up the locks quite rapidly.
//
// The `registered` defines some name for locks that
// already have been assigned to clients. It allows
// to help the `Acquire` method determining whether a
// client should be served a new lock or an existing
// one. Entries in this map are erased with each call
// to the `Release` method.
//
// The `cout` allows to notify errors and information to
// the user about the process going on internally within
// this element.
type ConcurrentLocker struct {
	locker         sync.Mutex
	locks          []*Lock
	availableLocks chan int
	registered     map[string]int
	cout           logger.Logger
}

// Lock :
// Allows to protect the access to a single resource
// by providing a way for concurrent clients to wait
// on a single resource.
//
// The `id` defines the index of this lock in the
// internal channel of the `ConcurrentLocker`. It
// allows to release correctly this lock when it's
// not needed anymore. It's value is negative in
// case the lock is not in use at the moment.
//
// The `res` defines the resource currently assigned
// to this locker. It is used as a mean to easily set
// the resource as deleted when the lock is released
// by its last user.
//
// The `use` defines how many concurrent users are
// currently relying on this lock. It is used as a
// way to determine whether one can release this lock
// from the `ConcurrentLocker` (and thus make it
// available again to other resources).
//
// The `waiter` is used by the `Wait` method to make
// sure that a single client is using the resource
// secured by this lock at any time. It will contain
// a single element which will be concurrently used
// and acquired in order to lock this object.
type Lock struct {
	id     int
	res    string
	use    int
	waiter chan struct{}
}

// configuration :
// Used internally to regroup all the variables that
// can be used to customize the number of locks that
// can be served in parallel by an instance of the
// `Concurrent` object.
//
// The `LocksCount` defines the number of locks that
// can be distributed amongst clients before a call
// to the `Acquire` method becomes blocking. This
// number can be made quite large in order to allow
// more clients to perform update operations on some
// entries in a table.
// The default value is `10`.
type configuration struct {
	LockCount int
}

// parseConfiguration :
// Used to parse the configuration file and environment
// variables provided when executing this server to get
// the values of the `Concurrent` properties.
//
// Returns the parsed configuration where all non-set
// properties have their default values.
func parseConfiguration() configuration {
	// Create the default configuration.
	config := configuration{
		LockCount: 10,
	}

	// Parse custom properties.
	if viper.IsSet("Concurrent.LockCount") {
		config.LockCount = viper.GetInt("Concurrent.LockCount")
	}

	return config
}

// NewConcurrentLocker :
// Perform the creation of a new `ConcurrentLocker` with
// configuration values retrieved from the environment
// variables and conf file provided to the server.
//
// The `log` will be assigned as the internal logging mean
// for this locker.
//
// Returns the created concurrent locker.
func NewConcurrentLocker(log logger.Logger) *ConcurrentLocker {

	// Parse the config.
	config := parseConfiguration()

	// Create the lockers.
	allLocks := make([]*Lock, config.LockCount)
	ids := make(chan int, 0)

	for id := range allLocks {
		// Create the lock.
		allLocks[id] = &Lock{
			id:     -1,
			res:    "",
			use:    0,
			waiter: make(chan struct{}, 1),
		}
		allLocks[id].waiter <- struct{}{}

		// Register this index as free.
		ids <- id
	}

	// Build the locker.
	cl := ConcurrentLocker{
		locker:         sync.Mutex{},
		locks:          allLocks,
		availableLocks: ids,
		registered:     make(map[string]int),
		cout:           log,
	}

	return &cl
}

// Acquire :
// Used to try to acquire a locker for the specified resource.
// This method will query the internal lockers and see whether
// one instance is free. If this is the case the locker will
// be registered for this resource and returned. Other clients
// coming after this call with a similar resource will receive
// a copy of the locker.
// In case a locker already exists for the resource it will be
// returned.
// In case no more lockers are available (i.e. they are all
// registered for other resources), this call will block until
// a locker is released.
//
// The `resource` defines the name of the resource for which
// a locker should be acquired.
//
// Returns the locker acquired for this resource.
func (cl *ConcurrentLocker) Acquire(resource string) *Lock {
	// Acquire the top level lock and make sure that we release
	// it whatever happens.
	var l *Lock

	// Check whether a lock already exists for this resource: if
	// this is the case we will increase its use count by one and
	// return it.
	func() {
		cl.locker.Lock()
		defer cl.locker.Unlock()

		// Check whether a lock already exists for this resource.
		id, ok := cl.registered[resource]
		if ok {
			// Return this lock and increase the use count.
			l = cl.locks[id]
			l.use++

			cl.cout.Trace(logger.Debug, fmt.Sprintf("Adding user to resource \"%s\" (id: %d, usage: %d, available: %d)", l.res, l.id, l.use, len(cl.availableLocks)))
		}
	}()

	if l != nil {
		return l
	}

	// No lock already exists for this locker, we need to create
	// a new one. To do so we will wait on the internal channel
	// containing available locks. This call will block until we
	// can access a locker. It will either return immediately if
	// some locks are still available and block until one current
	// user release one of the locks otherwise.
	id := <-cl.availableLocks

	// At this point we can register the lock for the specified
	// resource. We need to acquire the internal lock again.
	func() {
		cl.locker.Lock()
		defer cl.locker.Unlock()

		// Configure the lock to indicate that it is serving the
		// resource in input.
		cl.registered[resource] = id

		l = cl.locks[id]
		l.id = id
		l.res = resource
		l.use++

		cl.cout.Trace(logger.Debug, fmt.Sprintf("Creating locker on \"%s\" (id: %d, available: %d)", l.res, l.id, len(cl.availableLocks)))
	}()

	// We can return the lock we obtained.
	return l
}

// Release :
// Used to perform the release of the lock provided in input
// and handle the necessary verifications to see whether it
// can be put back in the list of available locks. This can
// only happen if no other user is using this lock.
//
// The `lock` defines the locker to release. If this value
// is `nil` nothing happens.
func (cl *ConcurrentLocker) Release(lock *Lock) {
	// Check consistency.
	if lock == nil {
		return
	}

	// Acquire the top level lock and make sure that we will
	// release it no matter what.
	cl.locker.Lock()
	defer cl.locker.Unlock()

	// Decrease the usage count for this locker.
	lock.use--

	// If some clients are still using it, do not put it back
	// in the list of available lockes.
	if lock.use > 0 {
		return
	}

	// Nobody is using this lock anymore, we can release it
	// and put it back in the pool of available locks. We
	// will also remove the reference to the resources in the
	// `registered` table so that if someone needs to lock it
	// again it will trigger the fetching of a new lock.
	delete(cl.registered, lock.res)
	cl.availableLocks <- lock.id

	lock.id = -1
	lock.res = ""

	cl.cout.Trace(logger.Debug, fmt.Sprintf("Releasing locker on \"%s\" at index %d (available: %d)", lock.res, lock.id, len(cl.availableLocks)))
}

// Lock :
// Used to wait to obtain the lock so as to make sure that
// the client process is the only one able to access to the
// resource secured by this object.
// The wait operation will block until the current user has
// released the resource by a call to `Release` on this
// item.
func (l *Lock) Lock() {
	// Wait on the channel attached to this lock until we can
	// fetch a value.
	<-l.waiter
}

// Release :
// Used to release this locker object so that other clients
// can access to the resource protected by it. This operation
// succeeds if no other `Release` call has been made since
// the last call to `Lock`.
//
// Returns an error in case the lock cannot be released (in
// case the `Release` method has already been called for
// example). Failure also occur if the `Release` method is
// called while no `Lock` call has been made before.
func (l *Lock) Release() error {
	// Check whether we already released the lock.
	if len(l.waiter) > 0 {
		return fmt.Errorf("Cannot release locker on resource, seems already released")
	}

	l.waiter <- struct{}{}

	return nil
}
