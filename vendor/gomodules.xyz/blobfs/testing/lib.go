package testing

import (
	"os"

	"gomodules.xyz/blobfs"
)

/*
NewTestGCS returns a BlobFS for a gcs bucket url gs://<bucket> and credential file
*/
func NewTestGCS(bucketURL, credential string) (blobfs.Interface, error) {
	if v, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		credential = v
	} else {
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credential)
		if err != nil {
			return nil, err
		}
	}

	return blobfs.New(bucketURL), nil
}
