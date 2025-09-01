package persister

import (
	"OpenSplit/logger"
	"OpenSplit/session"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type JsonFile struct {
	ctx               context.Context
	fileName          string
	lastUsedDirectory string
}

func (j *JsonFile) Startup(ctx context.Context) {
	j.ctx = ctx
}

func (j *JsonFile) Save(splitFile *session.SplitFile) error {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		logger.Error("save failed: " + err.Error())
		return err
	}

	splitFilePayload := splitFile.GetPayload()
	defaultFileName := j.getDefaultFileName(splitFilePayload)
	filename, err := runtime.SaveFileDialog(j.ctx, runtime.SaveDialogOptions{
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
		return nil
	}

	j.fileName = filename
	j.lastUsedDirectory = filepath.Dir(j.fileName)
	data, err := json.Marshal(splitFilePayload)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal split file payload: %s", err.Error()))
		return err
	}
	err = os.WriteFile(j.fileName, data, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save split file: %s", err.Error()))
	}

	return err
}

func (j *JsonFile) Load() (*session.SplitFile, error) {
	defaultDirectory, err := j.getDefaultDirectory()
	if err != nil {
		return nil, err
	}

	filename, err := runtime.OpenFileDialog(j.ctx, runtime.OpenDialogOptions{
		Title:            "Load OpenSplit File",
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
		logger.Debug("user cancelled save")
		return nil, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load split file: %s", err.Error()))
		return nil, err
	}

	var splitFilePayload session.SplitFilePayload
	err = json.Unmarshal(data, &splitFilePayload)
	if err != nil {
		return nil, err
	}

	return session.NewFromPayload(splitFilePayload), nil
}

func (j *JsonFile) getDefaultDirectory() (string, error) {
	var defaultDirectory string
	if j.lastUsedDirectory != "" {
		defaultDirectory = j.lastUsedDirectory
	} else {
		defaultDirectoryBase, err := os.UserHomeDir()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get user home directory: %s", err.Error()))
			return "", err
		}
		defaultDirectory = path.Join(defaultDirectoryBase, "OpenSplit", "Split Files")
		err = os.MkdirAll(defaultDirectory, 0755)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to create OpenSplit user data folder: %s", err.Error()))
			return "", err
		}
	}

	return defaultDirectory, nil
}

func (j *JsonFile) getDefaultFileName(splitFile session.SplitFilePayload) string {
	if j.fileName != "" {
		return j.fileName
	} else {
		return fmt.Sprintf("%s-%s.osf", splitFile.GameName, splitFile.GameCategory)
	}
}
