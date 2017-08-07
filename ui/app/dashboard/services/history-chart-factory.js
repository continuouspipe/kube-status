angular.module('kubeStatusDashboard')
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
