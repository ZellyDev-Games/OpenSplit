package platform

import "os"

type FileRuntime struct{}

func NewFileRuntime() *FileRuntime {
	return &FileRuntime{}
}

func (f *FileRuntime) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (f *FileRuntime) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f *FileRuntime) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *FileRuntime) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
