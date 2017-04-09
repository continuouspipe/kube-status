angular.module('kubeStatus')
    .service('ClusterRepository', function($resource) {
        var resource = $resource('http://localhost:8080/clusters');

        this.findAll = function() {
            return resource.query().$promise;
        };
    })
;
