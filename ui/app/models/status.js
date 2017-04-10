angular.module('kubeStatus')
    .service('StatusFetcher', function($http, API_URL) {
        this.findByCluster = function(cluster) {
            return $http.get(API_URL+'/clusters/'+cluster.identifier+'/status').then(function(response) {
                return response.data;
            });
        };

        this.findBySnaphost = function(cluster, snapshot) {
            return $http.get(API_URL+'/clusters/'+cluster.identifier+'/history/'+snapshot).then(function(response) {
                return response.data;
            });
        };

        this.historyEntriesByCluster = function(cluster) {
            return $http.get(API_URL+'/clusters/'+cluster.identifier+'/history').then(function(response) {
                return response.data;
            });
        };
    })
;
