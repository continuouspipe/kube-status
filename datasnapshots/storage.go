package datasnapshots

import (
	"cloud.google.com/go/storage"
	"fmt"
	"github.com/continuouspipe/kube-status/errors"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"os"
)

var envBucketName, _ = os.LookupEnv("GOOGLE_CLOUD_PLATFORM_BUCKET_NAME")

//BucketObjectWriter writes []bytes, if an error occurs it returns a list of errors
type BucketObjectWriter interface {
	Write([]byte) errors.ErrorListProvider
}

//KubeStatusBucket allows to handle the kubernates status information stored on the google bucket
type KubeStatusBucket struct{}

//NewKubeStatusBucket ctor for KubeStatusBucket
func NewKubeStatusBucket() *KubeStatusBucket {
	return &KubeStatusBucket{}
}

//Write writes []bytes into the kubernetes status google bucket specified via environment variable
func (k KubeStatusBucket) Write(content []byte) errors.ErrorListProvider {
	el := errors.NewErrorList()

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		el.AddErrorf("Failed to create client")
		el.Add(err)
		glog.Error(el)
		glog.Flush()
		return el
	}

	bucket := client.Bucket(envBucketName)

	/*buckets := client.Buckets(ctx, envProjectId)
	buckets.Prefix = envBucketName
	bucketAttr, err := buckets.Next()
	if err != nil {
		el.AddErrorf("We could not find the bucket")
		el.Add(err)
		glog.Error(el)
		glog.Flush()
	}*/

	obj := bucket.Object("data")

	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, string(content)); err != nil {
		el.AddErrorf("Failed to write to object")
		el.Add(err)
		glog.Error(el)
		glog.Flush()
		return el
	}
	if err := w.Close(); err != nil {
		el.AddErrorf("Failed to close writer")
		el.Add(err)
		glog.Error(el)
		glog.Flush()
		return el
	}
	return nil
}
