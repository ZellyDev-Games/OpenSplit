package session

import (
	"OpenSplit/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type RuntimeProvider interface {
	SaveFileDialog(context.Context, runtime.SaveDialogOptions) (string, error)
	OpenFileDialog(context.Context, runtime.OpenDialogOptions) (string, error)
}

type FileProvider interface {
	WriteFile(string, []byte, os.FileMode) error
	ReadFile(string) ([]byte, error)
	MkdirAll(string, os.FileMode) error
	UserHomeDir() (string, error)
}

type JsonFile struct {
	ctx               context.Context
	runtime           RuntimeProvider
	fileProvider      FileProvider
	fileName          string
	lastUsedDirectory string
}

func NewJsonFile(runtime RuntimeProvider, fileProvider FileProvider) *JsonFile {
	return &JsonFile{
		runtime:      runtime,
		fileProvider: fileProvider,
	}
}

func (j *JsonFile) Startup(ctx context.Context) {
	j.ctx = ctx
}

func (j *JsonFile) Save(splitFilePayload SplitFilePayload) error {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		logger.Error("save failed: " + err.Error())
		return err
	}

	defaultFileName := j.getDefaultFileName(splitFilePayload)
	filename, err := j.runtime.SaveFileDialog(j.ctx, runtime.SaveDialogOptions{
		Title:            "save OpenSplit File",
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
		return nil
	}

	j.fileName = filename
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

func (j *JsonFile) Load() (SplitFilePayload, error) {
	var splitFilePayload SplitFilePayload
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		return SplitFilePayload{}, err
	}

	filename, err := j.runtime.OpenFileDialog(j.ctx, runtime.OpenDialogOptions{
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
		logger.Debug("user cancelled save")
		return SplitFilePayload{}, nil
	}

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
