# gdiretc-idm

IDM-style lightweight Windows/macOS downloader for publicly shared Google Drive files.

**Goal:** Copy Google Drive link → choose save path and filename → download directly with a bundled [gdirect](https://github.com/mizorewww/gdirect) helper. No Google login or OAuth configuration for public files.

- MyGo desktop UI (Go + TypeScript).
- Strict Drive URL validation and clipboard monitoring.
- Native Save dialog, download list, stop/resume actions.
- gdirect v0.13.2 packaged as an internal helper, not installed separately.
- CI preview builds for Windows x64 and macOS Intel/Apple Silicon.

**Status:** development. Windows/macOS installers and real public-file network transfers are not validated yet. [Build workflow](https://github.com/lesnolie/gdiretc-idm/actions).

Private or download-restricted Drive files are not supported. Direct reachability depends on the user's network; Google quotas still apply. Bundled gdirect is GPL-2.0-or-later; redistributors must meet applicable license obligations.
