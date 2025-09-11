# Getting Started, A Developer's Guide

## Philosophy
OpenSplit was created to provide a cross-platform speedrun split timer with a high degree of visual customization via CSS skins.

## Architecture
Opensplit uses [Wails](https://wails.io) which provides a native application that uses Go and [React with Typescript](https://react.dev/).
To effectively contribute you'll need some familiarity with both Go and React.

## Prerequisites
#### Follow the installation steps at each link
* [**Go >= 1.20**](https://go.dev/doc/install)
  * Ensure that the bin folder modules are installed to is on your system path, you'll need the Wails CLI to be findable by your system
  * Windows after using installer: `setx PATH "$($env:PATH);$env:USERPROFILE\go\bin"`
  * Linux: `echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc && source ~/.bashrc`
  * macOS (zsh): `echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc && source ~/.zshrc`
* [**Node >= 22.19**](https://nodejs.org/en/download/)
  * [**NVM**](https://github.com/nvm-sh/nvm) or [**NVM for Windows**](https://github.com/coreybutler/nvm-windows) is highly recommended
  * After NVM install:
    * Windows:
        ```
        nvm install lts
        nvm use lts
      ```
    * Linux/macOS:
        ```
        nvm install --lts
        nvm use --lts 
        ```
* [**Wails CLI**](https://wails.io/docs/gettingstarted/installation)
  * install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
  * Verify installed and Go PATH bin is correctly on path: `wails version`
* [**Task**](https://github.com/go-task/task)
  * install with `go install github.com/go-task/task/v3/cmd/task@latest`
  * Makefile alternative that runs tasks to make your life a bit easier than shell commands via `Taskfile.yml`
  * Verify: `task --version`
* [**Git**](https://git-scm.com/downloads)
* [**Golang-CI**](https://golangci-lint.run/docs/welcome/install/)
  * Install with GitBash if you're on Windows
  * Technically optional, but CI runs this and will block PRs until it passes
  * Verify: `golangci-lint --version`

## Development Server
* From the checkout root run: `task dev` 
  * Compiles Go backend, generates frontend bindings, and installs all frontend dependencies
  * This can take some time the first time you run it
  * Hot-Reload
    * Changes to frontend (React) are hot-reloaded and instantly viewable in the application
    * Changes to backend (Go) will cause a recompilation and reload of the application

## Code Contribution
* Use relevant prefixes in your branch naming
  * New features or enhancements: `feat/feature-description`
  * Bugs or other problems: `fix/problem-to-fix`
  * Non-bug project upkeep: `chore/what-needs-doing`
  * Documentation updates: `docs/new-docs-page`
* Add unit test coverage against mock interface implementations for new Go features, 
no coverage needed for Wails internals or concrete implementations (e.g. OS filesystem hooks)
* Rebase main before creating PRs
* `task fmt` and `task lint` before creating PRs
* PRs must pass lint and unit tests before merge.
* All merges to main must be squash commits
