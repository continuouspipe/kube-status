#!/bin/sh
set -xe

# Updates the configuration file
echo "
angular.module('config', [])
.constant('API_URL', '"$API_URL"')
;" > ./var/static/scripts/config.js
