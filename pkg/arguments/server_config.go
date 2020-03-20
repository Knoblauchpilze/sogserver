package arguments

import (
	"fmt"
	"strings"

	uuid "github.com/google/uuid"
	"github.com/mycsHQ/webutils/arguments/cloud"
	"github.com/spf13/viper"
)

// ServerMetadata :
// Describes some properties used to identify the current instance of the server.
// This includes data about the machine executing the server but also information
// about the behavior of the server (such as the port on which end point can be
// targeted).
// Some information will be retrieved from the machine itself through the aws sdk
// if any can be found on the machine, and default values can be provided in the
// case of a local machine (typically in dev environment).
//
// Most of these information will be used during the logging process to provide
// some context to messages and be able to distinguish between instances when
// on production environment.
//
// The `AmiID` represents a string which identify the image of the system deployed on
// the machine executing the server. Typically the image can be any linux system with
// some additional drivers and tools (cuda, opencl, etc.). This identifier persists
// through a machine restart and allows to identify the version of the system
// executing the server which is useful to detect problems of compatibility
// between tools versions.
//
// The `PublicIPv4` corresponds to the IP address of the machine executing the server
// and persists through a restart. It allows to easily connect to a specific
// machine based on the logs, and also to identify furthermore the instances of a
// single application.
// If no such address is provided on the machine, the default "localhost" will most
// likely be used.
//
// The `PublicHostname` describes a host name which can be use to target the machine
// outside from the aws network. Typical values include the type of client (ec2,
// etc.), the IP address of the machine, the aws region of the machine, etc.
//
// The `InstanceType` describes the type of machine onto which the server is running.
// Typical amazon values include `t2.micro`, `p3.8xlarge`, etc.
// For other providers the aim is to provide a meaningful way to identify the type
// of machines used to run the server so that metrics are more relevant combined with
// information relative to the machine.
//
// The `InstanceID` describes an identifier of the current instance of the server.
// Each instance has its own identifier which allows to start several instances of
// a given application on the same machine.
//
// The `MachineID` describes the physical machine onto which the server is deployed.
// This ID persists through a restart of the application and can be used to identify
// the number of machines dedicated to a particular service.
//
// The `Environment` is a string describing the environment configuration used to
// start this server. Typical values include _local_, _master_, _production_, etc. It
// allows to quickly determine which set of parameters is used, as the environment
// usually describe different machines and different objectives. For example we
// usually have more devices and cache on a _production_ environment than in a _local_
// one.
// This helps diagnose errors at startup for example when a wrong number of devices is
// defined, etc.
//
// The `AppPort` specifies on which port the end point defined by the server can be
// accessed. This is useful especially in dev environment where we can run multiple
// API on the same machine and thus should be able to configure the port.
// The default value is 3000.
type ServerMetadata struct {
	AmiID          *string `json:"ami_id"`
	PublicIpv4     string  `json:"public_ipv4"`
	PublicHostname *string `json:"public_hostname"`
	InstanceType   *string `json:"instance_type"`
	InstanceID     string  `json:"instance_id"`
	MachineID      string  `json:"machine_id"`
	Environment    string  `json:"environment"`
	AppPort        int
}

// ParseConfig :
// Used to parse the server arguments and produce the corresponding structure. Server
// arguments allows to gather information about the runtime machine executing the application
// which is cool to provide contexts in error message and logs in general but also general
// properties of the environment into which the application will be executed.
// These properties can be used to adapt the behavior of the server (for example by specifying
// the port to use for end points, etc.).
//
// The `configFile` is a string describing the optional configuration file provided by the
// runtime of the application. This is usually the name of the configuration file without
// the extension which contains the parameters to apply to the varuous aspects of the
// application.
//
// This function returns the built-in server's properties.
func ParseConfig(configFile string) ServerMetadata {
	// Assign the extra path to use to reach the configuration file.
	viper.SetEnvPrefix("ENV")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Put the configuration file in the config structure
	// name of config file (without extension)
	viper.SetConfigName(configFile)

	// Optionally look for config in the working directory
	viper.AddConfigPath(".")
	viper.AddConfigPath("data/config")

	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Could not parse input configuration \"%s\" (err: %v)", configFile, err))
	}

	// Retrieve server metadata
	cloudMetadata, err := cloud.InitMetadata()
	if err != nil {
		panic(fmt.Errorf("Could not retrieve cloud metadata (err: %v)", err))
	}

	// Assign the public ip address.
	publicIPv4 := viper.GetString("PUBLIC_IPV4")
	if cloudMetadata.PublicIPv4 != nil {
		publicIPv4 = *cloudMetadata.PublicIPv4
	}

	// Assign the machine id.
	machineID := viper.GetString("MACHINE_ID")
	switch {
	case cloudMetadata.InstanceID != nil:
		machineID = *cloudMetadata.InstanceID
	case cloudMetadata.PublicIPv4 != nil:
		machineID = *cloudMetadata.PublicIPv4
	case machineID == "":
		machineID = "localhost"
	}

	// Create the server's configuration.
	serverMetadata := ServerMetadata{
		cloudMetadata.AmiID,
		publicIPv4,
		cloudMetadata.PublicHostname,
		cloudMetadata.InstanceType,
		uuid.New().String(),
		machineID,
		configFile,
		3000,
	}

	// Try to retrieve the configuration properties for the application.
	if viper.IsSet("ServerPort") {
		serverMetadata.AppPort = viper.GetInt("ServerPort")
	}

	// All is well.
	return serverMetadata
}
