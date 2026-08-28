package ioutil

import (
	"bytes"
	"embed"
	iofs "io/fs"
	"os"
	"path/filepath"
)

const (
	TriggerFile = "trigger"
)

type Reloader struct {
	dir string
	fs  embed.FS

	trigger []byte
	loaded  bool
	loadFn  func(fsys iofs.FS)
}

func NewReloader(dir string, fs embed.FS, loadFn func(fsys iofs.FS)) *Reloader {
	return &Reloader{
		dir:     dir,
		fs:      fs,
		trigger: nil,
		loaded:  false,
		loadFn:  loadFn,
	}
}

func (r *Reloader) FS() iofs.FS {
	// Fall back to the embedded FS unless the override directory exists.
	// Guard against every stat error (not just IsNotExist) so a nil FileInfo
	// is never dereferenced.
	if fi, err := os.Stat(r.dir); err != nil || !fi.IsDir() {
		return r.fs
	}
	// Only adopt the override directory once it has been fully populated,
	// signaled by the presence of the trigger file. Without this check an
	// empty or partially-written directory (e.g. a stale leftover) would
	// silently shadow the embedded FS and load nothing.
	if _, err := os.Stat(filepath.Join(r.dir, TriggerFile)); err != nil {
		return r.fs
	}
	return os.DirFS(r.dir)
}

func (r *Reloader) needsReload(fsys iofs.FS) bool {
	if data, err := iofs.ReadFile(fsys, TriggerFile); err == nil {
		yes := bytes.Compare(r.trigger, data) != 0
		r.trigger = data
		return yes || !r.loaded // ensure loads at least first time
	}
	return !r.loaded
}

func (r *Reloader) ReloadIfTriggered() {
	fsys := r.FS()
	if r.needsReload(fsys) {
		r.loadFn(fsys)
		r.loaded = true
	}
}
