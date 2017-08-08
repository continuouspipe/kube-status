'use strict';

angular
    .module('kubeStatus', [
        'config',
        'ngAnimate',
        'ngMessages',
        'ngSanitize',
        'angular-loading-bar',
        'ngResource',
        'ui.router',
        'ngMaterial',
        'yaru22.angular-timeago',
        'googlechart',
        'kubeStatusDashboard'
    ])
    .constant('KUBE_STATUS_TEMPLATE_URI_ROOT', '')
    .config(function ($urlRouterProvider, $locationProvider, $mdThemingProvider) {
        $urlRouterProvider.otherwise('/');
        $locationProvider.html5Mode(false);
    })
    .run(function($rootScope, $http) {
        $http.getError = function (error) {
            var response = error || {};
            var body = response.data || {};
            var message = body.message || body.error;

            if (!message && response.status == 400) {
                // We are seeing a constraint violation list here, let's return the first one
                message = body[0] && body[0].message;
            }

            if (typeof message == 'object') {
                message = message.message;
            }

            return message;
        };
    })
;
