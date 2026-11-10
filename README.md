# gdirect IDM

A native [MyGo](https://mygo.egoist.dev/docs/ui) Google Drive **public-file downloader** inspired by Internet Download Manager. Built with Go-only native UI (no WebView, no HTML, no JavaScript).

**Workflow:** Copy a Google Drive public file URL → native download dialog detects original filename → rename/select destination → download via bundled [gdirect](https://github.com/mizorewww/gdirect) → pause/resume/remove/reveal from the task list.

## Features (v0.4 preview)
- Google Drive file URL allowlist (no other providers). Clipboard recognition is always on while running, with no redundant checkbox.
- Native MyGo toolbar, sortable/scrollable table, context menus, save dialog and Windows/macOS tray.
- Compact progress track fixed to a consistent width, using **completed aria2 pieces** rather than logical output-file size (parallel range writes make file size misleading). The bar and percentage use the same normalized value; progress advances at aria2 checkpoint intervals. Download speed is read from gdirect's real libaria2 stderr updates. Unknown totals remain indeterminate.
- Filename suggestions repair Google Drive copy names such as `movie.mp4 的副本` → `movie 的副本.mp4` and use a trusted public download MIME type as an extension hint when available.
- Direct-only Google endpoint selection by gdirect and SQLite IP caching; no Google login/OAuth for public files.
- Windows helper processes do not open a console window.
- Geometric g application and tray icon.
- Right-click or toolbar **Restart Download** removes the existing file and aria2 resume data after stopping the helper, then starts from zero; **Resume** preserves partial data.
- A native Google IP selection dialog offers Auto (recommended: cached fastest route and fallback) plus scanned IPs with observed latency. Manual pinning uses gdirect's `--ip`/`--sni` and disables automatic failover. User can apply a chosen route to an active task with stop/resume; reconnect success still depends on network.
- Windows x64, macOS Apple Silicon and macOS Intel preview builds.

### Limitations
Google's quota/download restrictions and network reachability still apply. Folder links and private files are unsupported. No proxy is configured by this program, but the system routing policy can affect connections. Pausing is process cancellation with aria2 control files preserved for resumption, not an in-process aria2 pause. Progress is conservative between aria2 control-file checkpoints (normally about every five seconds); partial in-flight pieces are not counted until committed. Direct reachability, manual IP switching, and resume require device validation.

## Windows portable ZIP / macOS drag-to-Applications DMG

- **Windows:** Download `gdirect-idm-portable-windows-x64.zip`, extract the entire `gdirect IDM` folder, and double-click `gdirect IDM.exe`. Keep the bundled `resources` next to the executable. No MSI, NSIS installer or uninstall entry.
- **macOS (Apple Silicon / Intel):** Download the matching `gdirect-idm-macos-*.dmg`, double-click to mount, drag `gdirect IDM.app` onto the `Applications` shortcut, eject the disk image, and launch from Applications. Removal is by dragging the app to Trash; preferences/cache may remain in the user's data directory. Unsigned/unnotarized preview builds may be blocked by Gatekeeper; use the system's Privacy & Security approval when permitted.
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

The GitHub Actions workflow retrieves **gdirect v0.13.2**, verifies SHA-256 against upstream's signed release checksum list (checksum integrity only; not a cryptographic signature of its author), packages the verified helper, and produces a **portable Windows ZIP** and **native macOS drag-to-Applications DMGs** (Apple Silicon and Intel).

## License
Bundled gdirect uses **GPL-2.0-or-later**. Redistribution must include applicable license notices and corresponding source obligations; see THIRD_PARTY.md.
