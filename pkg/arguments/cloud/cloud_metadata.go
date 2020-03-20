package cloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var infos = []string{
	"ami-id",
	"hostname",
	"instance-id",
	"instance-type",
	"public-hostname",
	"public-ipv4",
}

// Metadata :
// Used to describe common properties which can be retrieved from
// the cloud configuration. Most of the information are retrieved
// from a common service provided by aws cloud configuration. This
// oes not extend well to other cloud service providers.
type Metadata struct {
	AmiID          *string `json:"ami-id"`
	Hostname       *string `json:"hostname"`
	InstanceID     *string `json:"instance-id"`
	InstanceType   *string `json:"instance-type"`
	PublicHostname *string `json:"public-hostname"`
	PublicIPv4     *string `json:"public-ipv4"`
}

// InitMetadata :
// Used to connect to the server providing metadata and to retrieve
// all the general information from this server.
// Returns the computed metadata along with any errors.
func InitMetadata() (Metadata, error) {
	// Retrieve information from the server.
	jsonString := []string{"{"}
	for _, info := range infos {
		resp, err := http.Get("169.254.169.254/2009-04-04/meta-data/" + info)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		jsonString = append(jsonString, fmt.Sprintf("%s:%s", info, string(body)))
	}
	jsonString = append(jsonString, "}")

	// Umarshal the retrieved information.
	meta := Metadata{}
	err := json.Unmarshal([]byte(strings.Join(jsonString, "")), &meta)
	if err != nil {
		return meta, fmt.Errorf("Could not unmarshal metadata retrieved from server (err: %v)", err)
	}

	// All is well.
	return meta, nil
}
