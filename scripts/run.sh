#!/bin/sh
export IMPORTER_PATH=${PWD}
export LD_LIBRARY_PATH=${IMPORTER_PATH}/lib:${LD_LIBRARY_PATH}

export GO111MODULE=on

# Read environment variables from file and export them.
file_env() {
	while read -r line || [[ -n $line ]]; do
		export $line
	done < "$1"
}

FILE="/opt/container_vars/environment_vars"

# If the input `$FILE` exists we need to export environment variables.
if [ -f $FILE ]; then
	file_env $FILE
fi

./bin/oglike_server -config $1
