'use strict';

angular.module('kubeStatus')
    .config(function($stateProvider) {
        $stateProvider
            .state('layout', {
                abstract: true,
                views: {
                    header: {
                        templateUrl: 'layout/views/header.html'
                    }
                }
            })
            .state('error', {
                abstract: true,
                parent: 'layout'
            })
            .state('error.404', {
                url: '/error/404',
                views: {
                    'content@': {
                        templateUrl: 'layout/views/error/404.html',
                        controller: 'ErrorController'
                    }
                }
            })
            .state('error.403', {
                url: '/error/403',
                views: {
                    'content@': {
                        templateUrl: 'layout/views/error/403.html',
                        controller: 'ErrorController'
                    }
                }
            })
            .state('error.500', {
                url: '/error/500',
                views: {
                    'content@': {
                        templateUrl: 'layout/views/error/500.html',
                        controller: 'ErrorController'
                    }
                }
            })
        ;
    });
