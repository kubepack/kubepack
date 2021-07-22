package loader

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
	"helm.sh/helm/v3/pkg/chart"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// ByteLoader loads a chart from a bytes.Reader
type ByteLoader bytes.Reader

// Load loads a chart
func (l *ByteLoader) Load() (*chart.Chart, error) {
	reader := (*bytes.Reader)(l)

	err := ensureArchiveReader("", reader)
	if err != nil {
		return nil, err
	}

	c, err := LoadArchive(reader)
	if err != nil {
		if err == gzip.ErrHeader {
			return nil, fmt.Errorf("response does not appear to be a valid chart file (details: %s)", err)
		}
	}
	return c, err
}

// ensureArchiveReader's job is to return an informative error if the file does not appear to be a gzipped archive.
//
// Sometimes users will provide a values.yaml for an argument where a chart is expected. One common occurrence
// of this is invoking `helm template values.yaml mychart` which would otherwise produce a confusing error
// if we didn't check for this.
func ensureArchiveReader(name string, raw io.ReadSeeker) (err error) {
	defer func() {
		_, e2 := raw.Seek(0, io.SeekStart) // reset read offset to allow archive loading to proceed.
		err = utilerrors.NewAggregate([]error{err, e2})
	}()

	_, err = raw.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	mime, err := mimetype.DetectReader(raw)
	if err != nil && err != io.EOF {
		err = fmt.Errorf("file '%s' cannot be read: %s", name, err)
		return
	}
	if !mime.Is("application/x-gzip") {
		err = fmt.Errorf("archive does not appear to be a gzipped archive; got '%s'", mime.String())
		return
	}
	return nil
}
