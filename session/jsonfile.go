package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"

	"github.com/zellydev-games/opensplit/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// JsonFile represents a SplitFile as a JSON file
//
// JsonFile provide utilities to work with the OS filesystem using the Wails runtime, and store information like
// the current filename and lastUsedDirectory for UX purposes.
type JsonFile struct {
	ctx               context.Context
	runtime           RuntimeProvider
	fileProvider      FileProvider
	fileName          string
	lastUsedDirectory string
}

// UserCancelledSave is a error that informs the calling system that the user cancelled a file open/load dialog.
//
// Wails generates exported bound methods to typescript functions that return a promise, if a not nil error is returned
// as the second return, Wails will reject the promise instead of fulfilling it. So this isn't necessarily an error
// that needs to be handled, but it is a convenient way to communicate to the frontend to catch()
// a promise instead of fulfilling it so that it doesn't try to do anything with an empty data structure.
type UserCancelledSave struct {
	error
}

// NewJsonFile creates a JsonFile with the provided RuntimeProvider and FileProvider
//
// In production code this will always be runtime.WailsRuntime and runtime.FileRuntime
func NewJsonFile(runtime RuntimeProvider, fileProvider FileProvider) *JsonFile {
	return &JsonFile{
		runtime:      runtime,
		fileProvider: fileProvider,
	}
}

// Startup is called either directly by Wails.run OnStartup, or by something else in that chain.
//
// The specific context.Context must be provided by Wails.run OnStartup or opening save/load file dialogs will panic.
func (j *JsonFile) Startup(ctx context.Context, service *Service) error {
	j.ctx = ctx
	service.AddCallback(func(ctx context.Context, payload ServicePayload) {
		if payload.SplitFile == nil {
			logger.Debug("clearing last used filename on split file close")
			j.fileName = ""
		}
	})

	return nil
}

// Save takes a SplitFile payload from the frontend, which modifies the passed in spitFile (or nil if a new file) from
// the Session Service backend.
func (j *JsonFile) Save(splitFilePayload SplitFilePayload) error {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		logger.Error("save failed: " + err.Error())
		return err
	}

	if j.fileName == "" {
		defaultFileName := j.getDefaultFileName(splitFilePayload)
		filename, err := j.runtime.SaveFileDialog(runtime.SaveDialogOptions{
			Title:            "Save OpenSplit File",
			DefaultFilename:  defaultFileName,
			DefaultDirectory: defaultDirectory,
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
			return UserCancelledSave{errors.New("user cancelled save")}
		}

		j.fileName = filename
	}

	j.lastUsedDirectory = filepath.Dir(j.fileName)
	data, err := json.Marshal(splitFilePayload)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal split file payload: %s", err.Error()))
		return err
	}
	err = j.fileProvider.WriteFile(j.fileName, data, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save split file: %s", err.Error()))
		return err
	}
	return err
}

// Load reads a JSON (*.osf) file from the path returned from the open file dialog
// and unserializes it into a SplitFilePayload
func (j *JsonFile) Load() (SplitFilePayload, error) {
	var splitFilePayload SplitFilePayload
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		return SplitFilePayload{}, err
	}

	filename, err := j.runtime.OpenFileDialog(runtime.OpenDialogOptions{
		Title:            "load OpenSplit File",
		DefaultDirectory: defaultDirectory,
		Filters: []runtime.FileFilter{{
			DisplayName: "OpenSplit Files",
			Pattern:     "*.osf",
		}},
	})

	if err != nil {
		logger.Error(fmt.Sprintf("failed to get path from open file dialog: %s", err.Error()))
		return SplitFilePayload{}, err
	}

	if filename == "" {
		logger.Debug("user cancelled load")
		return SplitFilePayload{}, UserCancelledSave{errors.New("user cancelled load")}
	}

	j.fileName = filename

	data, err := j.fileProvider.ReadFile(filename)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load split file: %s", err.Error()))
		return SplitFilePayload{}, err
	}

	err = json.Unmarshal(data, &splitFilePayload)
	if err != nil {
		return SplitFilePayload{}, err
	}

	return splitFilePayload, nil
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

func (j *JsonFile) getDefaultFileName(splitFile SplitFilePayload) string {
	if j.fileName != "" {
		return j.fileName
	} else {
		return fmt.Sprintf("%s-%s.osf", splitFile.GameName, splitFile.GameCategory)
	}
}
