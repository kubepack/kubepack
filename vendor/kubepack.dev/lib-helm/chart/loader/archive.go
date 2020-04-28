/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package loader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

var drivePathPattern = regexp.MustCompile(`^[a-zA-Z]:/`)

// FileLoader loads a chart from a file
type FileLoader string

// Load loads a chart
func (l FileLoader) Load() (*chart.Chart, error) {
	return LoadFile(string(l))
}

// LoadFile loads from an archive file.
func LoadFile(name string) (*chart.Chart, error) {
	if fi, err := os.Stat(name); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory")
	}

	raw, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	err = ensureArchive(name, raw)
	if err != nil {
		return nil, err
	}

	c, err := LoadArchive(raw)
	if err != nil {
		if err == gzip.ErrHeader {
			return nil, fmt.Errorf("file '%s' does not appear to be a valid chart file (details: %s)", name, err)
		}
	}
	return c, err
}

// ReaderLoader loads a chart from a bytes.Reader
type ReaderLoader bytes.Reader

// Load loads a chart
func (l *ReaderLoader) Load() (*chart.Chart, error) {
	return LoadReader((*bytes.Reader)(l))
}

// LoadReader loads from an archive file.
func LoadReader(reader *bytes.Reader) (*chart.Chart, error) {
	err := ensureArchive("", reader)
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

// ensureArchive's job is to return an informative error if the file does not appear to be a gzipped archive.
//
// Sometimes users will provide a values.yaml for an argument where a chart is expected. One common occurrence
// of this is invoking `helm template values.yaml mychart` which would otherwise produce a confusing error
// if we didn't check for this.
func ensureArchive(name string, raw io.ReadSeeker) (err error) {
	defer func() {
		_, e2 := raw.Seek(0, io.SeekStart) // reset read offset to allow archive loading to proceed.
		err = utilerrors.NewAggregate([]error{err, e2})
	}()

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

// LoadArchiveFiles reads in files out of an archive into memory. This function
// performs important path security checks and should always be used before
// expanding a tarball
func LoadArchiveFiles(in io.Reader) ([]*BufferedFile, error) {
	unzipped, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer unzipped.Close()

	var files []*BufferedFile
	tr := tar.NewReader(unzipped)
	for {
		b := bytes.NewBuffer(nil)
		hd, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hd.FileInfo().IsDir() {
			// Use this instead of hd.Typeflag because we don't have to do any
			// inference chasing.
			continue
		}

		switch hd.Typeflag {
		// We don't want to process these extension header files.
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			continue
		}

		// Archive could contain \ if generated on Windows
		delimiter := "/"
		if strings.ContainsRune(hd.Name, '\\') {
			delimiter = "\\"
		}

		parts := strings.Split(hd.Name, delimiter)
		n := strings.Join(parts[1:], delimiter)

		// Normalize the path to the / delimiter
		n = strings.ReplaceAll(n, delimiter, "/")

		if path.IsAbs(n) {
			return nil, errors.New("chart illegally contains absolute paths")
		}

		n = path.Clean(n)
		if n == "." {
			// In this case, the original path was relative when it should have been absolute.
			return nil, errors.Errorf("chart illegally contains content outside the base directory: %q", hd.Name)
		}
		if strings.HasPrefix(n, "..") {
			return nil, errors.New("chart illegally references parent directory")
		}

		// In some particularly arcane acts of path creativity, it is possible to intermix
		// UNIX and Windows style paths in such a way that you produce a result of the form
		// c:/foo even after all the built-in absolute path checks. So we explicitly check
		// for this condition.
		if drivePathPattern.MatchString(n) {
			return nil, errors.New("chart contains illegally named files")
		}

		if parts[0] == "Chart.yaml" {
			return nil, errors.New("chart yaml not in base directory")
		}

		if _, err := io.Copy(b, tr); err != nil {
			return nil, err
		}

		files = append(files, &BufferedFile{Name: n, Data: b.Bytes()})
		b.Reset()
	}

	if len(files) == 0 {
		return nil, errors.New("no files in chart archive")
	}
	return files, nil
}

// LoadArchive loads from a reader containing a compressed tar archive.
func LoadArchive(in io.Reader) (*chart.Chart, error) {
	files, err := LoadArchiveFiles(in)
	if err != nil {
		return nil, err
	}

	return LoadFiles(files)
}
