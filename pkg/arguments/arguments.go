package arguments

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// AppMetadata :
// Describes some properties used to identify the current instance of
// the application. This includes data about the machine executing it
// but also information about its behavior (such as the port that is
// exposed for external clients to target the app).
// Some information will be retrieved from the machine itself through
// various means and default values can be provided in the case of a
// local machine (typically in development environment).
//
// Most of these information will be used during the logging process
// to provide some context to messages and distinguish among running
// instances of the application (in case several are available).
//
// The `PublicIPv4` corresponds to the IP address of the machine that
// is executing the server and persists through a restart. It allows
// to easily connect to a specific machine based on the logs, and also
// to identify furthermore the instances of a single application.
// The default value is "localhost".
//
// The `InstanceID` describes an identifier of the current instance
// of the server. Each instance has its own identifier which allows
// to start several instances of a given app on the same machine.
// This value is generated at runtime and is meant to be unique and
// change upon restart of the application on the same machine.
// The default value is automatically generated.
//
// The `Environment` is a string describing the configuration used to
// start this application. A configuration describes a set of values
// that are usually suited to launch the app on a given machine or set
// of machines. A common example includes providing lower log level in
// production environment because we usually assumes that the debugging
// has already been performed in development.
// Typical values include can be `development`, `production`, etc.
// The default value is "unknown".
//
// The `Port` specifies on which port the end points defined by the app
// can be accessed. This is useful especially in dev environment where
// we can run multiple API on the same machine and thus should be able
// to configure the port.
// The default value is 3000.
type AppMetadata struct {
	PublicIPv4  string `json:"public_ipv4"`
	InstanceID  string `json:"instance_id"`
	Environment string `json:"environment"`
	Port        int
}

// Parse :
// Used to parse the app arguments and produce the corresponding data. The
// arguments allows to gather information about the runtime machine that is
// executing the application. It is useful to provide contexts in the error
// messages produced by the application but also general properties of the
// environment into which the application is to be executed.
// These properties can be used to adapt the behavior of the application (for
// example by specifying the port to expose to the outside world, etc.).
//
// The `configFile` is a string describing the optional configuration file
// provided by the runtime of the application. This is usually the name of
// the configuration file without the extension which contains the parameters
// to apply to the varuous aspects of the application.
//
// This function returns the built-in application's properties.
func Parse(configFile string) AppMetadata {
	// Assign the extra path to use to reach the configuration file.
	viper.SetEnvPrefix("ENV")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Put the configuration file in the config structure
	// name of config file (without extension).
	viper.SetConfigName(configFile)

	// Optionally look for config in the working directory and in the common
	// `data/config` directory.
	viper.AddConfigPath(".")
	viper.AddConfigPath("data/config")

	// Find and read the config file.
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("could not parse input configuration \"%s\" (err: %v)", configFile, err))
	}

	// Create the default application properties.
	metadata := AppMetadata{
		"localhost",
		uuid.New().String(),
		"unknown",
		3000,
	}

	// Fetch values from the configuration produced by the runtime.
	if len(configFile) > 0 {
		metadata.Environment = configFile
	}
	if viper.IsSet("App.Port") {
		metadata.Port = viper.GetInt("App.Port")
	}

	// Return the built-in configuration object.
	return metadata
}
