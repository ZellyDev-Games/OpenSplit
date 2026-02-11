package dispatcher

import (
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

const logModule = "dispatcher"

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI for testing.
type RuntimeProvider interface {
	OpenFileDialog(runtime.OpenDialogOptions) (string, error)
}

type FolderProvider interface {
	OpenSplitFileDir()
	OpenSkinsDir()
}

type RepoProvider interface {
	LoadSplitFile() (dto.SplitFile, error)
	SaveSplitFile(dto.SplitFile) error
}

// DispatchReply is sent in response to Dispatch
//
// Code greater than zero indicates an error situation
type DispatchReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DispatchReceiver interface {
	ReceiveDispatch(command.Command, *string) (DispatchReply, error)
}

type Service struct {
	mu             sync.Mutex
	receiver       DispatchReceiver
	runtime        RuntimeProvider
	folderProvider FolderProvider
	repo           RepoProvider
}

func NewService(receiver DispatchReceiver,
	runtime RuntimeProvider,
	folderProvider FolderProvider,
	repo RepoProvider,
) *Service {
	return &Service{
		runtime:        runtime,
		receiver:       receiver,
		folderProvider: folderProvider,
		repo:           repo,
	}
}

func (s *Service) Dispatch(cmd command.Command, payload *string) (DispatchReply, error) {
	logger.Debugf(logModule, "dispatching cmd: %v", cmd)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver.ReceiveDispatch(cmd, payload)
}

func (s *Service) OpenSplitFileFolder() {
	s.folderProvider.OpenSplitFileDir()
}

func (s *Service) OpenSkinsFolder() {
	s.folderProvider.OpenSkinsDir()
}

func (s *Service) ExportSplitFile(platform string) error {
	sf, err := s.repo.LoadSplitFile()
	if err != nil {
		return err
	}

	_, err = adapters.CleanSplitFile(sf)
	if err != nil {
		return err
	}

	//fileName, err := s.runtime.OpenFileDialog(runtime.OpenDialogOptions{
	//	DefaultFilename: fmt.Sprintf("%s-%s-%s.osf", file.GameName, file.GameCategory, platform),
	//	Title:           "Save Exported File",
	//	Filters: []runtime.FileFilter{{
	//		DisplayName: "OpenSplit File",
	//		Pattern:     "*.osf",
	//	}},
	//	CanCreateDirectories: true,
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}
