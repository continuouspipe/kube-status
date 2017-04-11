'use strict';

angular.module('kubeStatus')
    .controller('ErrorController', function($scope, $errorContext, $http) {
    	$scope.error = $errorContext.get();
        $scope.message = $http.getError($scope.error);
    });
