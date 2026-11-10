#!/usr/bin/env python3
"""Package *portable* MyGo binaries; never include installers or DMGs.

Windows: zip the EXE and the resources beside it, under one folder.
macOS: zip the complete .app bundle using ditto to preserve bundle metadata,
executable bits, symlinks, and signing resources.
"""
from __future__ import annotations
import argparse
import os
from pathlib import Path
import subprocess
import sys
import zipfile

TARGETS = {
    "windows-amd64": ("windows-amd64", "windows", "gdirect-idm-portable-windows-x64.zip"),
    "darwin-arm64": ("darwin-arm64", "macos", "gdirect-idm-portable-macos-arm64.zip"),
    "darwin-amd64": ("darwin-amd64", "macos", "gdirect-idm-portable-macos-intel.zip"),
}

def ensure_helper(root: Path, name: str) -> None:
    matches = [p for p in root.rglob(name) if p.is_file()]
    if not matches:
        raise SystemExit(f"Bundled gdirect helper missing in {root}. Refusing portable release.")

def win_package(source: Path, destination: Path) -> None:
    application = source / "gdirect IDM.exe"
    if not application.is_file():
        raise SystemExit(f"Portable EXE not found: {application}")
    ensure_helper(source, "gdirect.exe")
    included = 0
    with zipfile.ZipFile(destination, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=6) as out:
        for path in sorted(source.rglob("*")):
            if not path.is_file():
                continue
            if path.name.lower().endswith((".dmg", ".zip")) or path.name.lower().endswith("-setup.exe"):
                continue
            relative = path.relative_to(source)
            out.write(path, (Path("gdirect IDM") / relative).as_posix())
            included += 1
    if included < 2:
        destination.unlink(missing_ok=True)
        raise SystemExit("Portable Windows archive is missing application resources")

def mac_package(source: Path, destination: Path) -> None:
    bundles = list(source.glob("*.app"))
    if len(bundles) != 1:
        raise SystemExit(f"Expected exactly one .app in {source}, got {bundles}")
    application = bundles[0]
    ensure_helper(application, "gdirect")
    if sys.platform != "darwin":
        raise SystemExit("Preserving macOS bundles requires a macOS runner and ditto")
    subprocess.run(
        ["ditto", "-c", "-k", "--sequesterRsrc", "--keepParent", str(application), str(destination)],
        check=True,
    )

def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--platform", choices=TARGETS, required=True)
    args = p.parse_args()
    build_name, kind, filename = TARGETS[args.platform]
    source = Path("build") / build_name
    if not source.is_dir():
        raise SystemExit(f"Build output missing: {source}")
    target = Path("portable")
    target.mkdir(exist_ok=True)
    archive = target / filename
    if archive.exists():
        archive.unlink()
    if kind == "windows":
        win_package(source, archive)
    else:
        mac_package(source, archive)
    if archive.stat().st_size < 1_000_000:
        raise SystemExit("Portable archive suspiciously small")
    print(f"Portable archive: {archive}, bytes={archive.stat().st_size}")

if __name__ == "__main__":
    main()
