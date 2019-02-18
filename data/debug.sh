#!/bin/sh

CURR_DIR=$(dirname $0)

gdb --args ./bin/sogserver $(cat data/config/local)
