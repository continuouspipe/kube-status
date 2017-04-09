'use strict';

angular.module('kubeStatus')
    .config(function($stateProvider) {
        $stateProvider
            .state('clusters', {
                url: '/',
                parent: 'layout',
                views: {
                    'content@': {
                        templateUrl: 'dashboard/views/clusters/list.html',
                        controller: 'ClustersController'
                    }
                }
            })
            .state('cluster-status', {
                url: '/cluster/:clusterIdentifier/status',
                parent: 'layout',
                views: {
                    'content@': {
                        templateUrl: 'dashboard/views/clusters/status.html',
                        controller: 'ClusterStatusController'
                    }
                },
                resolve: {
                    cluster: function(ClusterRepository, $stateParams) {
                        return ClusterRepository.find($stateParams.clusterIdentifier);
                    }
                }
            })
        ;
    });
