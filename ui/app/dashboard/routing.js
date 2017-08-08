'use strict';

angular.module('kubeStatus')
    .config(function($stateProvider, KUBE_STATUS_TEMPLATE_URI_ROOT) {
        $stateProvider
            .state('clusters', {
                url: '/',
                parent: 'layout',
                views: {
                    'content@': {
                        templateUrl: KUBE_STATUS_TEMPLATE_URI_ROOT+'dashboard/views/clusters/list.html',
                        controller: 'ClustersController'
                    }
                }
            })
            .state('cluster-status', {
                url: '/cluster/:clusterIdentifier/status',
                parent: 'layout',
                views: {
                    'content@': {
                        templateUrl: KUBE_STATUS_TEMPLATE_URI_ROOT+'dashboard/views/status/layout.html',
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
                        templateUrl: KUBE_STATUS_TEMPLATE_URI_ROOT+'dashboard/views/status/full.html',
                        controller: 'ClusterStatusController'
                    }
                }
            })
        ;
    });
