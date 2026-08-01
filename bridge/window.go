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

	case AppViewNewSplitFile, AppViewEditSplitFile, AppViewSettings:
		return WindowConfig{
			Width:     1000,
			Height:    900,
			Resizable: false,
		}

	case AppViewSkinEditor:
		return WindowConfig{
			Width:     1200,
			Height:    800,
			Resizable: false,
		}

	case AppViewNewSkin, AppViewEditSkin:
		return WindowConfig{
			Width:     300,
			Height:    200,
			Resizable: false,
		}
	}

	return WindowConfig{
		Width:     800,
		Height:    600,
		Resizable: false,
	}
}
