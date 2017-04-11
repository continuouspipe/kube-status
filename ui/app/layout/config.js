'use strict';

angular.module('kubeStatus')
    .config(function(cfpLoadingBarProvider) {
        cfpLoadingBarProvider.includeSpinner = false;
    });
