package logger

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// configuration :
// Provides a way to configure the way logs are displayed both in terms of
// level and in terms of the machine executing the logger.
// This logger uses a display to the standard input as a logging strategy
// with some coloring based on the severity of the logs to display. The
// logger is initialized with default name for the application and with a
// local configuration but information are retrieved from the configuration
// file to modify it.
//
// The `AppName` describes a string for the name of the application using
// the logger.
// The default value is "Unknown app".
//
// The `Environment` allows to specify which configuration is used by the
// application executing the logger. Typical values include `production`
// and all other settings such as `development`, etc. but other can be set
// if needed. Usually this string is meant to refer to a dedicated file
// describing the related configuration, which allows to quickly determine
// which environment is used by any application without needing to check
// other more obfuscated parameters.
// The default value is "development".
//
// The `ForceLocal` allows to make sure that the instance ID assigned to
// this logger will be "local" no matter what the value provided by the
// runtime is. This allows to make logs in development environment clearer
// by ignoring the automatically generated name.
// The default value is `false`.
//
// The `Level` is a string representing the minimum level of a log message
// in order for it to be displayed. Basically it allows to filter debug
// message from production environment or to also supress info message so
// that important messages get their deserved visibility in critical envs
// (such as production for example).
// The default value is "info".
//
// The `Buffer` allows to specify the size of the buffer to handle log
// messages. As we might have phases where the application produce lots of
// log messages, the logger does not directly output message to the standard
// output. Instead it stores them in an internal buffer with a predefined
// size which is almost instantaneous. This allows to accumulate messages
// without latency up to a certain amount so that we can absorb burst messages
// production before dumping them to the output channel when there are less
// logs to handle.
// A larger value for this attribute allows for a larger buffer and thus to
// absorb even more logs if needed. Note however that if the logger cannot
// process messages fast enough this buffer is bound to
// The default value is 500.
type configuration struct {
	AppName     string
	Environment string
	ForceLocal  bool
	Level       string
	Buffer      int
}

// traceMessage :
// Describes a message to be enqued by the logger. It contains all the needed
// information to be displayed by the logger such as its severity, name and
// content.
// We can distinguish two kind of messages:
//   - simple messages, which are basically strings describing a content and
//     nothing more.
//   - events which are usually represented using a json string where the event's
//     data is stored in an organized fashion.
// Both message will not be logged exactly the same way: indeed usually simple
// messages are used for debug purposes or to display simple information while
// events are usually the carriers of metrics or relevant information about the
// state of the application.
//
// The `severity` value represents the actual importance of the log message.
//
// The `name` might be nil and represents a key describing the message. An example
// can be given as follows:
//    Name: "Player"
//    Content "Creation: date"
// It might be used to identify similar logs (for example sample duration) even
// though it is much easier to compute statistics from events than from simple
// messages. It might also be empty if the message has no particular title.
// Note that this property is not displayed in the case of simple message but is
// used in the case events.
//
// The `content` represent the content of the message and is dumped as is during
// the logging process. It might be anythng the user want, but common values are
// plain strings for simple messages and json string for events.
//
// The `isEvent` boolean is true if the trace represents an event and false otherwise.
// Note that it does not imply anything about the values of `Name` or `Content`, the
// only difference comes from the logging method. Events will be dumped without pretty
// information (such as the name of the application, the timestamp, etc.) while simple
// messages will be dumped with some context.
// The reason behind that is that we consider that events should be self-explanatory
// while simple messages should not.
type traceMessage struct {
	level   Severity
	name    string
	content string
	isEvent bool
}

// StdLogger :
// Describes the logger structure used to perform logging.
// This logger is forwarding log message received as go structure to the standard
// input and handles a buffer mechanism so that anyone can put a log messag and
// not be blocked while the underlying display system is performing the log.
// This will also come in handy if we ever decide to change the logging to another
// more complex such as uploading the logs somewhere where we might need some time
// to perform the logging in itself and where no modifications would be required
// as it is already off-loaded to a dedicated routine.
// Most of the properties are configurable through a dedicated file, which is parsed
// upon creating a new logger.
//
// The `configuration` allows to retrieve information about the settings and changes
// to apply to input log messages before passing them to the C layer.
//
// The `instanceID` represents the name of the instance of the application running
// the logger. The `instanceID` is updated each time the application restarts which
// allows to effectively detect crashes on a single machine or detect various apps
// running on a single machine.
//
// The `publicIP` represents the public IP of the machine as a string. Note that in
// case no public IP can be determined a "localhost" value is used as default in
// order not to mix this instance with true remote machines.
//
// The `logChannel` is used to receive the trace messages from go modules before
// sending them to the logging device. Its size is determined by the configuration
// file and it allows a lagless enqueuing of messages as long as the buffer is not
// full.
//
// The `endChannel` allows to terminate the active loop which transmit log message
// from the `logChannel` to the logging device.
//
// The `closed` value indicates whether the logger has been terminated or not. One
// can access this value after locking the `locker` attribute to determine whether
// it is safe to post messages in the `logChannel`. It is mostly used to ensure that
// the logger always display the messages posted up until the `Release` method is
// called.
//
// The `locker` allows to protect the `closed` boolean from concurrent accesses.
//
// The `waiter` allows to wait for the proper termination of the logging routine in
// order to allow the display of the last posted log messages.
type StdLogger struct {
	config     configuration
	instanceID string
	publicIP   string
	logChannel chan traceMessage
	endChannel chan bool
	closed     bool
	locker     sync.Mutex
	waiter     sync.WaitGroup
}

// parseConfiguration :
// Used to retrieve the parameters to apply to the logger from the configuration
// file. A default configuration is provided to work in most cases but one can
// modify some settings at runtime.
//
// Returns the arguments parsed from the configuration file.
func parseConfiguration() configuration {
	// Provide a default configuration.
	config := configuration{
		"Unknown app",
		"development",
		false,
		"info",
		500,
	}

	// Parse the description file if any.
	if viper.IsSet("Logger.Name") {
		config.AppName = viper.GetString("Logger.Name")
	}
	if viper.IsSet("Logger.Environment") {
		config.Environment = viper.GetString("Logger.Environment")
	}
	if viper.IsSet("Logger.ForceLocal") {
		config.ForceLocal = viper.GetBool("Logger.ForceLocal")
	}
	if viper.IsSet("Logger.Level") {
		config.Level = viper.GetString("Logger.Level")
	}
	if viper.IsSet("Logger.Buffer") {
		config.Buffer = viper.GetInt("Logger.Buffer")
	}

	// All is well
	return config
}

// NewLogger :
// Used to create a new logger with the specified instance name and public ip.
// The created logger will parse the configuration file provided by the env
// and adapt its configuration right away.
//
// The `instanceID` string might be equal to "local" if no instance ID is
// provided by the server's properties. Otherwise it corresponds to a unique
// identifier of the machine running the logger.
//
// The `publicIP` provides the IP to use to target the machine executing the
// logger. If no such IP is provided (i.e. empty value) the default value is
// set to "localhost" so that we can still provide a consistent behavior by
// assuming that the server is ran locally.
//
// The return value represents the produced logger.
func NewStdLogger(instanceID string, publicIP string) Logger {
	// Retrieve the configuration.
	config := parseConfiguration()

	// Create the logger.
	log := StdLogger{
		config,
		instanceID,
		publicIP,
		make(chan traceMessage, config.Buffer),
		make(chan bool),
		false,
		sync.Mutex{},
		sync.WaitGroup{},
	}

	// Update the public IP and instance ID in case no values are provided.
	if len(log.instanceID) == 0 || config.ForceLocal {
		log.instanceID = "local"
	}
	if len(log.publicIP) == 0 {
		log.publicIP = "localhost"
	}

	// Start logging.
	log.waiter.Add(1)
	go log.performLogging()

	// Return the built-in logger.
	return &log
}

// Release :
// Used to perform the stopping of the active loop meant to handle logging
// to the underlying device. It will block until the method actually does
// return to make sure that the last logs posted will be dumped.
func (log *StdLogger) Release() {
	// Request the termination of the active loop.
	log.endChannel <- false

	// Close the log channel.
	log.locker.Lock()
	log.closed = true
	close(log.logChannel)
	log.locker.Unlock()

	// Wait for the routine termination.
	log.waiter.Wait()
}

// Trace :
// Used to perform the log of the input message with the specified level.
// The log message is not directly transmitted to the logging device but
// instead placed in the internal buffer of trace message so that it can
// be processed by the active logger loop.
// Note that this function does not block the caller if the channel is not
// full. Otherwise the caller will be blocked until a slot is available in
// the internal buffer.
//
// The `level` describes the severity of the message to log.
//
// The `message` describes the content of the message to log.
func (log *StdLogger) Trace(level Severity, message string) {
	// Create a trace object from the input element.
	trace := traceMessage{
		level,
		"",
		message,
		false,
	}

	// Enqueue the trace to the internal channel if it is not closed yet.
	log.locker.Lock()
	defer log.locker.Unlock()
	if !log.closed {
		log.logChannel <- trace
	}
}

// performLogging :
// Used to perform logging. This method is meant to be launched as a go routine
// and will regularly poll the internal trace channel to perform logging.
func (log *StdLogger) performLogging() {
	// Until we request stop, we must continue logging.
	keepConnection := true

	for keepConnection {
		select {
		case keepConnection = <-log.endChannel:
			// The end channel has been activated, terminate the logging process.
			break
		case trace := <-log.logChannel:
			// A new trace is available, log it.
			log.performSingleLog(trace)
		}
	}

	// Iterate over the remaining message of the log channel.
	for trace := range log.logChannel {
		log.performSingleLog(trace)
	}

	// Set the routine as done.
	log.waiter.Done()
}

// performSingleLog :
// Used to perform a single log for the input trace. This method is called from
// the active logging loop and perform the conversion of the input message into
// something that can be displayed by the associated logging device.
// A distinction is done based on the type of trace to log (message or event)
// and the suited wrapper method is used.
//
// The `trace` describes the message or event to log.
func (log *StdLogger) performSingleLog(trace traceMessage) {
	// Format the log to the standard output by providing some information about
	// the message to log and the instance producing it.
	out := FormatWithBrackets(log.config.AppName, Magenta)
	out += " " + FormatWithBrackets(log.instanceID, Magenta)
	out += " " + FormatWithNoBrackets(time.Now().Format("2006-01-02 15:04:05"), Magenta)
	out += " " + trace.level.String()

	out += " " + trace.content

	fmt.Println(out)
}
