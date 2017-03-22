# Kubernetes Status API

This application allows to get the status of a Kubernetes cluster through a simple API endpoint.

## API usage

### `POST /cluster/full-status`

*Request body*
```
{
    "address": "https://your-cluster-address",
    "username": "username",
    "password": "password"
}
```

*Example response body*
```
{
    "resources": {
        "cpu": {
            "requests": "1300m",
            "limits": "14800m",
        }
        "memory": {
            "requests": "6550Mi",
            "limits": "30368660992",
        },
        "percents": {
            "cpu": {
                "requests": "32%",
                "limits": "370%",
            }
            "memory": {
                "requests": "43%",
                "limits": "192%",
            },
        }
    }
}
```


## Development

1. Install the dependencies
  ```
  make install-dep
  ```

2. Run the API
  ```
  KUBE_STATUS_LISTEN_ADDRESS=http://0.0.0.0:80 go run main.go
  ```
