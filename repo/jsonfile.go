package repo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/logger"
)

// ErrUserCancelledSave is an error that informs the calling system that the user cancelled a file open/load dialog.
//
// Wails generates exported bound methods to typescript functions that return a promise, if a not nil error is returned
// as the second return, Wails will reject the promise instead of fulfilling it. So this isn't necessarily an error
// that needs to be handled, but it is a convenient way to communicate to the frontend to catch()
// a promise instead of fulfilling it so that it doesn't try to do anything with an empty data structure.
var ErrUserCancelledSave = errors.New("user cancelled file open operation")

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI for testing.
type RuntimeProvider interface {
	Startup(ctx context.Context)
	SaveFileDialog(runtime.SaveDialogOptions) (string, error)
	OpenFileDialog(runtime.OpenDialogOptions) (string, error)
	MessageDialog(runtime.MessageDialogOptions) (string, error)
	Quit()
}

// FileProvider wraps os hooks and file operations to allow DI for testing.
type FileProvider interface {
	WriteFile(string, []byte, os.FileMode) error
	ReadFile(string) ([]byte, error)
	MkdirAll(string, os.FileMode) error
	UserHomeDir() (string, error)
}

// JsonFile represents a SplitFile as a JSON file
//
// JsonFile provide utilities to work with the OS filesystem using the Wails runtimeProvider, and store information like
// the current filename and lastUsedDirectory for UX purposes.
type JsonFile struct {
	runtimeProvider   RuntimeProvider
	fileProvider      FileProvider
	fileName          string
	lastUsedDirectory string
}

// NewJsonFile creates a JsonFile with the provided RuntimeProvider and FileProvider
//
// In production code this will always be platform.WailsRuntime and platform.FileRuntime
func NewJsonFile(runtime RuntimeProvider, fileProvider FileProvider) *JsonFile {
	return &JsonFile{
		runtimeProvider: runtime,
		fileProvider:    fileProvider,
	}
}

func (j *JsonFile) ClearCachedFileName() {
	logger.Debug("clearing last used filename")
	j.fileName = ""
}

func (j *JsonFile) SaveAs(payload []byte, defaultFileName string) error {
	return j.SaveSplitFile(payload, defaultFileName)
}

// SaveSplitFile takes a payload marshalled as bytes and saves it to disk
func (j *JsonFile) SaveSplitFile(payload []byte, identifier string) error {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		logger.Error("save failed: " + err.Error())
		return err
	}

	if j.fileName == "" {
		filename, err := j.runtimeProvider.SaveFileDialog(runtime.SaveDialogOptions{
			Title:            "Save OpenSplit File",
			DefaultDirectory: defaultDirectory,
			DefaultFilename:  identifier,
			Filters: []runtime.FileFilter{{
				DisplayName: "OpenSplit Files",
				Pattern:     "*.osf",
			}},
		})

		if err != nil {
			logger.Error(fmt.Sprintf("failed to get path from save file dialog: %s", err.Error()))
			return err
		}

		if filename == "" {
			logger.Debug("user cancelled save")
			return ErrUserCancelledSave
		}

		j.fileName = filename
	}

	if !strings.HasSuffix(strings.ToLower(j.fileName), ".osf") {
		j.fileName += ".osf"
	}

	j.lastUsedDirectory = filepath.Dir(j.fileName)
	err = j.fileProvider.WriteFile(j.fileName, payload, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save split file: %s", err.Error()))
		return err
	}
	return err
}

// LoadSplitFile reads a JSON (*.osf) file from the path returned from the open file dialog
// and unserializes it into a SplitFilePayload
func (j *JsonFile) LoadSplitFile() ([]byte, error) {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		return nil, err
	}

	filename, err := j.runtimeProvider.OpenFileDialog(runtime.OpenDialogOptions{
		Title:            "load OpenSplit File",
		DefaultDirectory: defaultDirectory,
		Filters: []runtime.FileFilter{{
			DisplayName: "OpenSplit Files",
			Pattern:     "*.osf",
		}},
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to get path from open file dialog: %s", err.Error()))
		return nil, err
	}

	if filename == "" {
		logger.Debug("user cancelled load")
		return nil, ErrUserCancelledSave
	}

	j.fileName = filename

	data, err := j.fileProvider.ReadFile(filename)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load split file: %s", err.Error()))
		return nil, err
	}

	return data, nil
}

func (j *JsonFile) SaveConfig(configServicePayload []byte) error {
	defaultDirectoryBase, err := j.fileProvider.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get user home directory: %s", err.Error()))
		return err
	}

	configDirectory := path.Join(defaultDirectoryBase, "OpenSplit")
	err = j.fileProvider.MkdirAll(configDirectory, 0755)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create OpenSplit user data folder: %s", err.Error()))
		return err
	}

	err = j.fileProvider.WriteFile(path.Join(configDirectory, "os-config.json"), configServicePayload, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save split file: %s", err.Error()))
		return err
	}
	return err
}

func (j *JsonFile) LoadConfig() ([]byte, error) {
	defaultDirectoryBase, err := j.fileProvider.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get user home directory: %s", err.Error()))
		return nil, err
	}

	configDirectory := path.Join(defaultDirectoryBase, "OpenSplit")
	err = j.fileProvider.MkdirAll(configDirectory, 0755)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create OpenSplit user data folder: %s", err.Error()))
		return nil, err
	}

	data, err := j.fileProvider.ReadFile(filepath.Join(configDirectory, "os-config.json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrConfigMissing
		}
		logger.Error(fmt.Sprintf("failed to load OpenSplit config: %s", err.Error()))
		return nil, err
	}
	return data, err
}

func (j *JsonFile) getDefaultDirectory() (string, error) {
	var defaultDirectory string
	if j.lastUsedDirectory != "" {
		defaultDirectory = j.lastUsedDirectory
	} else {
		defaultDirectoryBase, err := j.fileProvider.UserHomeDir()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get user home directory: %s", err.Error()))
			return "", err
		}
		defaultDirectory = path.Join(defaultDirectoryBase, "OpenSplit", "Split Files")
		err = j.fileProvider.MkdirAll(defaultDirectory, 0755)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create OpenSplit user data folder: %s", err.Error()))
			return "", err
		}
	}

	return defaultDirectory, nil
}
