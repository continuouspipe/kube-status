#!/bin/bash

set -xe

./prepare-ui.sh

exec /usr/bin/kube-status "$@"
