package command

//go:generate go run tools/gencommands/main.go

// Command bytes are sent to dispatcher.Service.Dispatch.
// The active state decides how each command is handled.
type Command byte

const (
	QUIT Command = iota

	//
	// Split files
	//

	NEW
	LOAD
	EDIT

	CANCEL
	SUBMIT

	CLOSE

	RESET
	SAVE

	//
	// Timer
	//

	SPLIT
	UNDO
	SKIP

	PAUSE

	//
	// Configuration
	//

	TOGGLEGLOBAL
	SETLAYOUT
	TOGGLEWR

	FOCUS

	//
	// Internal
	//

	HELLO

	DONE
	UNDONE

	//
	// Runtime offset
	//

	SET_RUNTIME_OFFSET
	CLEAR_RUNTIME_OFFSET

	//
	// Display control
	//

	COMPARISON_LEFT
	COMPARISON_RIGHT

	//
	// Skin management
	//

	NEW_SKIN
	EDIT_SKIN

	SKIN_SELECT
	SKIN_SET_DEFAULT_LAYOUT

	//
	// Skin editor navigation
	//

	SKIN_FILE
	SKIN_ELEMENT
	SKIN_RULE

	CLEAR_ELEMENT

	//
	// Skin editor working copy mutations
	//

	SKIN_CREATE_FILE
	SKIN_CREATE_RULE

	SKIN_RULE_UPDATE
	SKIN_FILE_UPDATE

	SKIN_RULE_DELETE

	SKIN_PREVIEW_SET

	//
	// Skin editor persistence
	//

	SKIN_SAVE
	SKIN_RELOAD
)
