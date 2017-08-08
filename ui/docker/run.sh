#!/bin/sh
set -xe

# Updates the configuration file
echo "
angular.module('config', [])
.constant('KUBE_STATUS_API_URL', '"$KUBE_STATUS_API_URL"')
;" > ./var/static/scripts/config.js
