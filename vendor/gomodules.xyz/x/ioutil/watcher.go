package ioutil

import (
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"gomodules.xyz/sets"
	"k8s.io/klog/v2"
)

type Watcher struct {
	WatchFiles []string
	WatchDir   string
	Reload     func() error

	reloadCount uint64
}

func (w *Watcher) incReloadCount(filename string) {
	atomic.AddUint64(&w.reloadCount, 1)
	klog.Infof("file %s reloaded: %d", filename, atomic.LoadUint64(&w.reloadCount))
}

func (w *Watcher) Run(stopCh <-chan struct{}) error {
	fileset := sets.NewString(w.WatchFiles...)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		<-stopCh
		defer watcher.Close()
	}()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				filename := filepath.Clean(event.Name)
				if fileset.Len() > 0 && !fileset.Has(filename) {
					continue
				}

				switch event.Op {
				case fsnotify.Create:
					if err = watcher.Add(filename); err != nil {
						klog.Errorln("error:", err)
					}
				case fsnotify.Write:
					if err := w.Reload(); err != nil {
						klog.Errorln("error:", err)
					} else {
						w.incReloadCount(filename)
					}
				case fsnotify.Remove, fsnotify.Rename:
					if err = watcher.Remove(filename); err != nil {
						klog.Errorln("error:", err)
					}
				}
			case err := <-watcher.Errors:
				if err != nil {
					klog.Errorln("error:", err)
				}
			}
		}
	}()

	for _, filename := range w.WatchFiles {
		if err = watcher.Add(filename); err != nil {
			klog.Errorf("error watching file %s. Reason: %s", filename, err)
		}
	}
	if err = watcher.Add(w.WatchDir); err != nil {
		return fmt.Errorf("error watching dir %s. Reason: %s", w.WatchDir, err)
	}

	return nil
}
