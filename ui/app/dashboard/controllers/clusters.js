'use strict';

angular.module('kubeStatus')
    .controller('ClustersController', function ($scope, $remoteResource, ClusterRepository) {
        $remoteResource.load('clusters', ClusterRepository.findAll()).then(function (clusters) {
            $scope.clusters = clusters;
        });
    });
