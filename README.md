# OpenSplit
<div>
    <img height="175" style="margin: auto" src="docs/images/screenshot.png"  alt=""/>
</div>

<hr />

> **Free & open-source speedrun split timer with an emphasis on customization.**

<p>
  <a href="#downloads">Nightly builds</a> •
  <a href="#quickstart">Quickstart</a> •
  <a href="#skins">Skins</a> •
  <a href="#development">Development</a> •
  <a href="#contributing">Contributing</a>
</p>

---

## Highlights
- 🕒 **Fast, readable timer** built for speedrunning.
- 🎨 **Fully skinnable UI** — drop CSS-based skins (tokens + components + images) into a folder and switch at runtime.
- 🧩 **Split editor & segment management** (in progress) for painless editing.
- 🎮 **Global hotkeys** (Windows first; cross-platform planned).
- 🔎 **Speedrun.com integration** (planned) to search games, categories, and fetch art.
- 🧰 **Modern stack**: Go + React/TypeScript via Wails.

> Status: early development/alpha. Expect rapid change and frequent nightlies.

---

## Downloads

**Nightly builds** (updated on each merge to `main`):

- Windows (x64): `https://github.com/ZellyDev-Games/OpenSplit/releases/download/nightly/opensplit-windows-amd64.zip`
- macOS (Intel/Apple Silicon): `https://github.com/ZellyDev-Games/OpenSplit/releases/download/nightly/opensplit-darwin-<arch>.zip`
- Linux (x64/ARM64): `https://github.com/ZellyDev-Games/OpenSplit/releases/download/nightly/opensplit-linux-<arch>.zip`

> If these links 404, nightlies might not be enabled yet. See **Building from source** below.

---

## Quickstart

### Run the app
1. Download a nightly for your OS (or build from source).
2. Unzip and run the binary. On macOS, you may need to right-click → Open the first time.
3. From the app, load a sample or create a new split file.

### Create your first splits (basic flow)
- Create a new split file (Game + Category).
- Add segments, then start a run.
- Press your **Split** hotkey (default: `Space`) at each segment end.

> Hotkeys are configurable in the app settings (Windows supported first). Global hotkeys on macOS/Linux are planned.

---

## Features (current & roadmap)
- **Timer**: HH:MM:SS.cc display with centiseconds; formatting adapts to hours/minutes.
- **Split editor**: add/rename/reorder segments; total attempts; planned import/export.
- **Hotkeys**: split, reset, pause; Windows global hooks implemented; cross-platform planned.
- **Data**: simple JSON split files (`.osf`) for portability; future migration tooling planned.
- **Skins**: theme tokens + component styles + images; per-skin folder with live switching.
- **Integrations**: Speedrun.com lookup for game/category art (upcoming).

---

## Architecture

OpenSplit uses **[Wails](https://wails.io/)** to bundle a Go backend and a React/TypeScript frontend.
