package testing

import (
	"io/ioutil"
	"net/url"
	"os"

	"golang.org/x/oauth2/google"
	"gomodules.xyz/blobfs"
)

/*
NewTestGCS returns a BlobFS for a gcs bucket url gs://<bucket> and credential file
*/
func NewTestGCS(bucketURL, credential string) (*blobfs.BlobFS, error) {
	if v, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		credential = v
	} else {
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credential)
		if err != nil {
			return nil, err
		}
	}

	saKey, err := ioutil.ReadFile(credential)
	if err != nil {
		return nil, err
	}

	cfg, err := google.JWTConfigFromJSON(saKey)
	if err != nil {
		return nil, err
	}

	filename := credential + "-private-key"
	err = ioutil.WriteFile(filename, cfg.PrivateKey, 0644)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("access_id", cfg.Email)
	q.Set("private_key_path", filename)
	u.RawQuery = q.Encode()
	return blobfs.New(u.String()), nil
}
