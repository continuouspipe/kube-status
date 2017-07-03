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
                        templateUrl: 'dashboard/views/status/layout.html',
                        controller: 'ClusterStatusLayoutController'
                    }
                },
                resolve: {
                    cluster: function(ClusterRepository, $stateParams) {
                        return ClusterRepository.find($stateParams.clusterIdentifier);
                    }
                }
            })
            .state('cluster-status-view', {
                url: '/{status}',
                parent: 'cluster-status',
                views: {
                    'status': {
                        templateUrl: 'dashboard/views/status/full.html',
                        controller: 'ClusterStatusController'
                    }
                }
            })
        ;
    });
