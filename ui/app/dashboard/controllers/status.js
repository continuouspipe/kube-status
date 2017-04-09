angular.module('kubeStatus')
    .controller('ClusterStatusController', function($scope, $remoteResource, $mdColors, StatusFetcher, cluster) {
        $scope.cluster = cluster;

        $remoteResource.load('status', StatusFetcher.findByCluster(cluster)).then(function (status) {
            $scope.status = status;
        });

        $scope.colour_from_percents = function(percents) {
            return $mdColors.getThemeColor(raw_colour_from_percents(percents));
        };

        $scope.colour_from_node_status = function(status) {
            return $mdColors.getThemeColor(raw_colour_from_node_status(status));
        };

        $scope.colour_from_pod = function(pod) {
            return $mdColors.getThemeColor(colour_from_pod(pod));
        };

        var colour_from_pod = function(pod) {
            if (pod.status == 'Running') {
                if (pod.isReady) {
                    return 'green';
                }

                return 'blue';
            } else if (pod.status == 'Pending') {
                return 'orange';
            }

            return 'red';
        };

        var raw_colour_from_node_status = function(status) {
            if ('Ready' == status) {
                return 'green';
            } 

            return 'red';
        };

        var raw_colour_from_percents = function(percents) {
            if (percents > 90) { return 'red'; }
            else if (percents > 80) { return 'orange'; }
            else if (percents > 40) { return 'blue'; }
            return 'blue';
        }
    })
;
