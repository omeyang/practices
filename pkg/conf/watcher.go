package config

import "github.com/fsnotify/fsnotify"

// WatcherInterface 定义了 fsnotify.Watcher 需要模拟的方法
type WatcherInterface interface {
	Add(name string) error
	Remove(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type FsNotifyWatcher struct {
	watcher WatcherInterface
}

func NewFsNotifyWatcher(w WatcherInterface) *FsNotifyWatcher {
	return &FsNotifyWatcher{watcher: w}
}

func (f *FsNotifyWatcher) Add(path string) error {
	return f.watcher.Add(path)
}

func (f *FsNotifyWatcher) Remove(path string) error {
	return f.watcher.Remove(path)
}

func (f *FsNotifyWatcher) Close() error {
	return f.watcher.Close()
}

func (f *FsNotifyWatcher) Events() <-chan fsnotify.Event {
	return f.watcher.Events()
}

func (f *FsNotifyWatcher) Errors() <-chan error {
	return f.watcher.Errors()
}
