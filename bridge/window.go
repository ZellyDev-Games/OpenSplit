package bridge

func WindowForView(view View) WindowConfig {
	switch view {
	case AppViewRunning:
		return WindowConfig{
			Width:     800,
			Height:    600,
			Resizable: true,
		}

	case AppViewWelcome:
		return WindowConfig{
			Width:     320,
			Height:    580,
			Resizable: false,
		}

	case AppViewNewSplitFile:
		return WindowConfig{
			Width:     1000,
			Height:    900,
			Resizable: false,
		}

	case AppViewEditSplitFile:
		return WindowConfig{
			Width:     1000,
			Height:    1100,
			Resizable: false,
		}

	case AppViewSettings:
		return WindowConfig{
			Width:     1000,
			Height:    900,
			Resizable: false,
		}

	case AppViewSkinEditor:
		return WindowConfig{
			Width:     1700,
			Height:    900,
			Resizable: false,
		}

	case AppViewNewSkin:
		return WindowConfig{
			Width:     300,
			Height:    350,
			Resizable: false,
		}

	case AppViewEditSkin:
		return WindowConfig{
			Width:     300,
			Height:    350,
			Resizable: false,
		}
	}

	return WindowConfig{
		Width:     800,
		Height:    600,
		Resizable: false,
	}
}
