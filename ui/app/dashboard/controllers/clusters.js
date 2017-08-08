'use strict';

angular.module('kubeStatusDashboard')
    .controller('ClustersController', function ($scope, $remoteResource, ClusterRepository) {
        $remoteResource.load('clusters', ClusterRepository.findAll()).then(function (clusters) {
            $scope.clusters = clusters;
        });
    });
