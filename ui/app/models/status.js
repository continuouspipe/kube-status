angular.module('kubeStatus')
    .service('StatusFetcher', function($http, API_URL) {
        this.findByCluster = function(cluster) {
            return $http.post(API_URL+'/cluster/status/full', cluster).then(function(response) {
                return response.data;
            });
        };
    })
;
