angular.module('kubeStatus')
    .service('ClusterRepository', function($resource, $q, API_URL) {
        var resource = $resource(API_URL+'/clusters');

        this.findAll = function() {
            return resource.query().$promise;
        };

        this.find = function(identifier) {
            return this.findAll().then(function(clusters) {
                for (var i = 0; i < clusters.length; i++) {
                    if (clusters[i].identifier == identifier) {
                        return clusters[i];
                    }
                }

                return $q.reject(new Error('Cluster not found'));
            });
        }
    })
;
