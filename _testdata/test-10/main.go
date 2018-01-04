package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/Masterminds/vcs"
)

func main()  {
	for ; ; {
		path, err := os.Getwd()
		if err != nil {
			break
		}

		repo, err := getRootDir(path)
		if err != nil {
			break
		}

		dirty := repo.IsDirty()

		if !dirty {
			fmt.Println("dirty")
			break
		}
	}
}

func getRootDir(path string) (vcs.Repo, error) {
	var err error
	for ; ; {
		repo, err := vcs.NewRepo("", path)
		if err == nil {
			return repo, err
		}
		if os.Getenv("HOME") == path {
			break
		}
		path = filepath.Dir(path)
	}

	return nil, err
}
