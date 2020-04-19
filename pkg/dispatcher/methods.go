package dispatcher

import (
	"fmt"
	"oglike_server/pkg/logger"
	"strings"
)

// getSupportedMethods :
// Returns the list of `HTTP` verbs that can be used as valid filtering
// methods for such a handler.
func getSupportedMethods() map[string]bool {
	return map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"CONNECT": true,
		"OPTIONS": true,
		"TRACE":   true,
		"PATCH":   true,
	}
}

// filterMethods :
func filterMethods(methods []string, log logger.Logger) map[string]bool {
	filtered := make(map[string]bool, 0)
	supported := getSupportedMethods()

	for _, method := range methods {
		consolidated := strings.ToUpper(method)
		_, ok := supported[consolidated]

		// Filter invalid methods.
		if !ok {
			log.Trace(logger.Error, getModuleName(), fmt.Sprintf("Filtering invalid HTTP method \"%s\"", method))
			continue
		}

		filtered[consolidated] = true
	}

	return filtered
}
