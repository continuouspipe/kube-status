# Kubernetes Status API

This application allows to get the status of a Kubernetes cluster through a simple API endpoint.

## Usage (with the binary)

You need the following environment variables:

| Name | Required | Description | Example |
| `CLUSTER_LIST` | If using the default in-memory cluster provider | An inline JSON description of your clusters | ``[{"identifier":"my-cluster","address":"https://1.2.3.4","username":"-","password":"-"}]' |
| `KUBE_STATUS_LISTEN_ADDRESS` | Yes | The to expose the API to | `http://127.0.0.1:8080` |
| `GOOGLE_CLOUD_PROJECT_ID` | If using the Google Cloud Datastore history backend | The project used to store the history to | `my-project-1234` |

```
./main
```

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
