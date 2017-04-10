#/bin/bash

set -xe

./prepare-ui.sh

exec /usr/bin/kube-status -logtostderr -v 5
