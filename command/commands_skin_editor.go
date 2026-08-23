// command/commands_skin_editor.go
package command

const (
	//
	// Skin editor navigation
	//

	SKIN_FILE    Command = 28
	SKIN_ELEMENT Command = 29
	SKIN_RULE    Command = 30

	CLEAR_ELEMENT Command = 31

	//
	// Skin editor working copy mutations
	//

	SKIN_CREATE_FILE Command = 32
	SKIN_CREATE_RULE Command = 33

	SKIN_RULE_UPDATE Command = 34
	SKIN_FILE_UPDATE Command = 35

	SKIN_RULE_DELETE Command = 36

	SKIN_PREVIEW_SET Command = 37

	SKIN_SAVE   Command = 38
	SKIN_RELOAD Command = 39
)
