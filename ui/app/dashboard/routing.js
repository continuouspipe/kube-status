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
        ;
    });
