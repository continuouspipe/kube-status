angular.module('kubeStatus')
    .controller('ClusterStatusLayoutController', function($scope, $state, $remoteResource, StatusFetcher, HistoryChartFactory, cluster) {
        $scope.cluster = cluster;

        $remoteResource.load('history', StatusFetcher.historyEntriesByCluster(cluster).catch(function() {
            return [];
        }).then(function(history) {
            if (history.length == 0) {
                return $state.go('cluster-status-view', {'status': 'live'});
            }

            $scope.history = history;
            $scope.historyChartDefinition = HistoryChartFactory.fromHistory(history);

            return $state.go('cluster-status-view', {'status': history[history.length -1].UUID});
        }));

        $scope.selectHandler = function(selectedItem) {
            var snapshot = $scope.history[selectedItem.row];

            return $state.go('cluster-status-view', {'status': snapshot.UUID});
        };
    })
    .controller('ClusterStatusController', function($scope, $remoteResource, $mdColors, $mdDialog, $stateParams, StatusFetcher, cluster) {
        var statusFetcher;
        if ($stateParams.status == 'live') {
            statusFetcher = StatusFetcher.findByCluster(cluster);
        } else {
            statusFetcher = StatusFetcher.findBySnaphost(cluster, $stateParams.status);
        }

        $remoteResource.load('status', statusFetcher).then(function (status) {
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

        $scope.node_used_volumes_in_percents = function(node) {
            return (node.volumesInUse / 16) * 100;
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
    .service('HistoryChartFactory', function() {
        this.fromHistory = function(history) {
            return {
                "data": {
                    "cols": [
                        { type: 'string', id: 'Snapshot' },
                        { type: 'date', id: 'Start' },
                        { type: 'date', id: 'End' }
                    ], 
                    "rows": history.map(function(snapshot) {
                        var time = Date.parse(snapshot.EntryTime),
                            left = new Date(time),
                            right = new Date(time + 1 * 60000);

                        return {
                            c: [
                                {v: 'Snapshots'},
                                {v: left},
                                {v: right},
                            ]
                        }
                    })
                },
                "type": "Timeline",
                "displayed": false,
                "options": {
                    timeline: { 
                        colorByRowLabel: true
                    },
                    height: 100,
                    'tooltip' : {
                      trigger: 'none'
                    }
                }
            };
        };
    })
;
