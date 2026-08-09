package skin

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/repo"
)

const (
	logModule  = "skins"
	EntryPoint = "index.css"
)

// SkinProvider defines the public skin service API exposed to the rest of the
// application.
type SkinProvider interface {
	SetSkin(string, bool) error
	CreateSkin(name string, base string) error

	// Editor navigation.
	SelectElement(string) error
	SelectFile(string) error
	SelectRule(string, string) error
	ClearElement() error

	// Editor mutation.
	UpdateActiveRule(dto.CSSRuleEditor) error
	UpdateFileContents(string, string) error

	// Working copy.
	CreateCSSFile(string) error
	CreateCSSRule(dto.CSSRuleEditor) error
	DeleteCSSRule(string) error

	// Persistence.
	SaveWorkingCopy() error
	ReloadEditor() error

	EmitSkinModel() error

	SelectedSkin() string
	GetAvailableSkins() []string

	SetPreviewElements([]dto.SkinPreviewElement) error
}

// DirectoryWatcher watches a directory for changes.
type DirectoryWatcher interface {
	Start(string, func())
	ChangeRoot(string)
}

// Service manages installed skins, the embedded HTTP server, filesystem
// watching, editor state, and persistence of the selected skin.
type Service struct {
	m sync.RWMutex

	initOnce sync.Once
	initErr  error

	skinDir string

	server   *http.Server
	listener net.Listener

	serving atomic.Bool

	address string
	port    int

	selectedSkin string

	configService *config.Service
	repoService   *repo.Service

	watcher DirectoryWatcher

	skinUpdatedCh chan string
	skinModelCh   chan SkinModel

	editor *EditorState
}

// NewService creates a skin service and its notification channels.
func NewService(
	skinDir string,
	configService *config.Service,
	repoService *repo.Service,
	watcher DirectoryWatcher,
) (*Service, chan string, chan SkinModel) {
	updateCh := make(chan string, 1)
	modelCh := make(chan SkinModel, 1)

	return &Service{
		skinDir: skinDir,

		configService: configService,
		repoService:   repoService,

		watcher: watcher,

		skinUpdatedCh: updateCh,
		skinModelCh:   modelCh,

		editor: NewEditorState(),
	}, updateCh, modelCh
}
