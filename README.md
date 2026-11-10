# gdirect IDM

A native [MyGo](https://mygo.egoist.dev/docs/ui) Google Drive **public-file downloader** inspired by Internet Download Manager. Built with Go-only native UI (no WebView, no HTML, no JavaScript).

**Workflow:** Copy a Google Drive public file URL → native download dialog detects original filename → rename/select destination → download via bundled [gdirect](https://github.com/mizorewww/gdirect) → pause/resume/remove/reveal from the task list.

## Features (v0.3 portable preview)
- Google Drive file URL allowlist (no other providers).
- Native MyGo toolbar, sortable/scrollable table, context menus, save dialog and Windows/macOS tray.
- Visible file progress, transferred bytes, speed, and ETA when a public file exposes its total size. When total size cannot be determined, progress remains indeterminate rather than pretending to know.
- Direct-only Google endpoint selection by gdirect and SQLite IP caching; no Google login/OAuth for public files.
- Windows helper processes do not open a console window.
- Geometric g application and tray icon.
- User chooses whether removing a task also deletes the local file. Running downloads are stopped before deletion.
- Windows x64, macOS Apple Silicon and macOS Intel preview builds.

### Limitations
Google's quota/download restrictions and network reachability still apply. Folder links and private files are unsupported. No proxy is configured by this program, but the system routing policy can affect connections. Pausing is process cancellation with aria2 control files preserved for resumption, not an in-process aria2 pause. A transfer's bytes are sampled from the actual growing output file (upstream uses `file-allocation=none`).

## Portable, no installation

- **Windows:** Download `gdirect-idm-portable-windows-x64.zip`, extract the entire `gdirect IDM` folder, and double-click `gdirect IDM.exe`. Keep the bundled `resources` next to the executable. No MSI, NSIS installer or uninstall entry.
- **macOS (Apple Silicon / Intel):** Download the matching `gdirect-idm-portable-macos-*.zip`, extract, and open `gdirect IDM.app`. No DMG or drag-to-Applications step required. Unsigned/unnotarized testing builds can be blocked by Gatekeeper; approve from Privacy & Security or use right-click → Open when permitted.
- **Tactile buttons:** Native MyGo Buttons have neutral material tokens for primary, secondary, tertiary, ghost and text. Hover/press/focus/disabled remain native and accessible.
- Portable does not mean zero preferences: the runtime may still use the user data directory for state and the gdirect IP cache, but no installer is registered.

## Build
Requires Go 1.27.3 or later. No Node, Bun, WebView or frontend dependencies.

```shell
go mod tidy
go run ./cmd/genicon
go test ./...
go tool mygo build -skip-installer -skip-dmg -platform windows/amd64
python scripts/package_portable.py --platform windows-amd64
```

The GitHub Actions workflow retrieves **gdirect v0.13.2**, verifies SHA-256 against upstream's signed release checksum list (checksum integrity only; not a cryptographic signature of its author), packages the verified helper, and produces three platform-specific **portable ZIP archives only**.

## License
Bundled gdirect uses **GPL-2.0-or-later**. Redistribution must include applicable license notices and corresponding source obligations; see THIRD_PARTY.md.
