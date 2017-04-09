angular.module('kubeStatus')
    .controller('ClusterStatusController', function($scope, $remoteResource, $mdColors, $mdDialog, StatusFetcher, cluster) {
        $scope.cluster = cluster;

        $remoteResource.load('status', StatusFetcher.findByCluster(cluster)).then(function (status) {
            $scope.status = status;

            $scope.podsByNode = {};
            status.nodes.forEach(function(node) {
                $scope.podsByNode[node.name] = byNode(status.pods, node.name);
            });
        });

        $scope.showPodDetails = function(namespace, pod) {
            var scope = $scope.$new();
            scope.pod = pod;
            scope.namespace = namespace;

            $mdDialog.show({
                controller: function($mdDialog, $scope) {
                    $scope.close = function() {
                        $mdDialog.cancel();
                    };
                },
                templateUrl: 'dashboard/views/pod/details.html',
                parent: angular.element(document.body),
                clickOutsideToClose:true,
                scope: scope
            });
        };

        var byNode = function(podsByNamespace, nodeName) {
            var filteredPodsByNode = {};

            for (var namespace in podsByNamespace) {
                podsByNamespace[namespace].forEach(function(pod) {
                    if (pod.nodeName != nodeName) {
                        return;
                    }

                    if (!(namespace in filteredPodsByNode)) {
                        filteredPodsByNode[namespace] = [];
                    }

                    filteredPodsByNode[namespace].push(pod);
                });
            }

            return filteredPodsByNode;
        };

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
