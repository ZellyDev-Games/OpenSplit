# Contributing to OpenSplit

First off — thank you for your interest in contributing! 🎉  
OpenSplit is an open-source, cross-platform speedrun split timer built with **Go + React/TypeScript** via **Wails**. This guide explains how to get set up, how we work, and how to land great PRs.

---

## Table of Contents
- Ways to Contribute
- Prerequisites
- Quick Start (Local Dev)
- Build, Test, and Lint
- Skins & Theming
- Git & PR Workflow
- Code Style & Conventions
- OS-Specific Notes
- Reporting Bugs & Security Issues
- License

---

## Ways to Contribute

- **Issues:** bug reports, feature requests, UX feedback (use the provided templates).
- **Code:** fixes, features, tests, refactors.
- **Docs:** README improvements, developer notes, in-app help.
- **Skins:** new themes (CSS) and assets.
- **Testing:** try nightlies, report regressions, share platform-specific findings.

---

## Development Prerequisites

> This is a very high level introduction to getting started with development.  For a more indepth look at the application [check the docs](./docs/getting-started.md)

- **Go** ≥ 1.22 - [Installation](https://go.dev/doc/install)
- **Node.js** ≥ 20 and **npm** - [Installation](https://nodejs.org/en/download)
- **Wails v2 CLI** — install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Task** - install with: `go install github.com/go-task/task/v3/cmd/task@latest`
- **Git** (on Windows, use **Git Bash** or **PowerShell (pwsh)** for scripts)

### Optional but highly recommended

- **golang CI** [Installation](https://golangci-lint.run/docs/welcome/install/) - install with GitBash if you're on Windows.  
It's a lot easier to deal with its complaints locally than looking at CI logs
---


## Quick Start (Local Dev)
`task clean` (Only needed once)

`task dev`

The app should launch. Edit files in `frontend/` or Go packages and it will rebuild/reload.

---

## Build, Test, and Lint

**Production build**
- `task build`
- Outputs appear in `build/bin/`

**Tests**
- Run all Go tests: `task test`

**Lint & format**
- Will run go vet, and frontend lint: `task lint`
> Note: go vet will return an error in the windows hotkey provider package. This is normal.

> CI runs tests (and optionally lint) on PRs. Keep your branch green for a fast merge.

---

## Skins & Theming

Skins are **plain CSS**. A typical skin folder contains:
- `tokens.css` — CSS variables (colors, fonts, radii, spacing)
- `components.css` — component styles that consume those tokens
- `images/` — optional backgrounds/icons
- `fonts/` — optional `@font-face` sources

Guidelines:
- Define tokens in `:root`; components consume them.
- Use relative URLs (e.g., `images/bg.png`) so assets travel with the skin.
- Using `@layer` to separate tokens vs. components is encouraged.

Include a screenshot/GIF when submitting a skin PR. 🎨

---

## Git & PR Workflow

1. **Branch** from `main`: e.g., `feat/split-editor-drag` or `fix/win-hotkeys-extended`
2. **Conventional Commits** (small, focused commits):  
   `feat: add Speedrun.com search`  
   `fix(windows): handle extended keys in hook`  
   `chore(ci): add nightly workflow`
3. **Run tests locally**: `task test`
4. **Run lint locally** `task lint`
4. **Open a Pull Request**:
  - Fill out the PR template
  - Add screenshots/GIFs for UI changes
  - Link issues with `Fixes #123` when applicable
5. **Address review** feedback; we squash or rebase as needed.

For larger features, open an issue to discuss approach before coding.

---

## Code Style & Conventions

**General**
- Follow `.editorconfig` for line endings/indentation.
- Keep functions small; comment intent where it isn’t obvious.

**Go**
- Don’t store contexts long-term; pass them down call chains.
- Use **build tags** for OS-specific code (e.g., `*_windows.go`).
- Prefer small interfaces for runtime adapters (e.g., `RuntimeProvider`) and inject them for testability.
- Unit-test behavioral logic; treat Wails adapters and OS hooks as integration areas.

**TypeScript/React**
- Prefer strict TypeScript.
- Keep components small; extract logic to hooks.
- Use CSS variables from skin tokens for theming.

**Commits**
- Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`) in the imperative mood.

---

## OS-Specific Notes

- **Windows**: Global hotkeys are implemented first. When casting OS pointers (e.g., from Win32 callbacks), convert and **immediately copy** into Go values; don’t retain foreign pointers.
- **macOS/Linux**: Global hotkeys are planned; APIs differ.
- **Wails builds**: It’s safe to delete `build/bin/` or use `wails build -clean`. Do **not** delete the entire `build/` folder unless the resource files (icons/manifests) are tracked and restored.

---

## Reporting Bugs & Security Issues

- **Bugs/requests**: open an issue using the templates with repro steps, logs, and OS details.
- **Security**: report privately via GitHub Security Advisories or the contact listed in `SECURITY.md` (avoid public issues for vulnerabilities).

---

## License

By contributing, you agree that your contributions are licensed under the **MIT License** (see `LICENSE`). There is no CLA at this time; if that changes, we’ll document it here.

---

## Thank You

Your time and ideas make OpenSplit better for everyone. If you get stuck or want guidance on where to start, open a Discussion or issue.
