package config

import (
	"testing"

	mocks "github.com/omeyang/practices/mocks/conf"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/mock/gomock"
)

func TestFsNotifyWatcher_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	fsw := NewFsNotifyWatcher(mockWatcher)

	mockWatcher.EXPECT().Add("/path/to/watch").Return(nil)

	err := fsw.Add("/path/to/watch")
	if err != nil {
		t.Errorf("Add returned an error: %v", err)
	}
}

func TestFsNotifyWatcher_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	fsw := NewFsNotifyWatcher(mockWatcher)

	mockWatcher.EXPECT().Remove("/path/to/remove").Return(nil)

	err := fsw.Remove("/path/to/remove")
	if err != nil {
		t.Errorf("Remove returned an error: %v", err)
	}
}

func TestFsNotifyWatcher_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	fsw := NewFsNotifyWatcher(mockWatcher)

	mockWatcher.EXPECT().Close().Return(nil)

	err := fsw.Close()
	if err != nil {
		t.Errorf("Close returned an error: %v", err)
	}
}

func TestFsNotifyWatcher_Events(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	fsw := NewFsNotifyWatcher(mockWatcher)

	eventChan := make(chan fsnotify.Event, 1)
	mockWatcher.EXPECT().Events().Return(eventChan)

	events := fsw.Events()
	if events != eventChan {
		t.Errorf("Events did not return the correct channel")
	}
}

func TestFsNotifyWatcher_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWatcher := mocks.NewMockWatcherInterface(ctrl)
	fsw := NewFsNotifyWatcher(mockWatcher)

	errorChan := make(chan error, 1)
	mockWatcher.EXPECT().Errors().Return(errorChan)

	errors := fsw.Errors()
	if errors != errorChan {
		t.Errorf("Errors did not return the correct channel")
	}
}
