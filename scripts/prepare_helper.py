#!/usr/bin/env python3
"""Bundle a pinned, checksum-verified gdirect release into MyGo resources."""
import argparse
import hashlib
import io
import re
import shutil
import tarfile
import urllib.request
import zipfile
from pathlib import Path, PurePosixPath

VERSION = "0.13.2"
TARGETS = {
 "darwin-arm64": ("aarch64-apple-darwin", "tar.gz", "gdirect"),
 "darwin-amd64": ("x86_64-apple-darwin", "tar.gz", "gdirect"),
 "windows-amd64": ("x86_64-pc-windows-gnu", "zip", "gdirect.exe"),
}

def download(url: str) -> bytes:
 request = urllib.request.Request(url, headers={"User-Agent": "gdiretc-idm-build"})
 with urllib.request.urlopen(request, timeout=120) as response:
  return response.read()

def main(platform: str) -> None:
 target, ext, executable = TARGETS[platform]
 name = f"gdirect-{VERSION}-{target}.{ext}"
 prefix = f"https://github.com/mizorewww/gdirect/releases/download/v{VERSION}"
 sums = download(f"{prefix}/SHA256SUMS").decode("utf-8")
 match = re.search(rf"(?m)^([0-9a-fA-F]{{64}})\s+\*?{re.escape(name)}\s*$", sums)
 if not match:
  raise RuntimeError("Upstream checksum not found: " + name)
 archive = download(f"{prefix}/{name}")
 digest = hashlib.sha256(archive).hexdigest()
 if digest.lower() != match.group(1).lower():
  raise RuntimeError("gdirect release checksum mismatch")
 output = Path("resources") / platform / "bin"
 output.mkdir(parents=True, exist_ok=True)
 seen = set()
 root = f"gdirect-{VERSION}-{target}"
 def write_one(name: str, reader) -> None:
  parts = PurePosixPath(name).parts
  if not parts or parts[0] != root or any(p in (".", "..") for p in parts):
   raise RuntimeError("Unsafe archive member: " + name)
  relative = PurePosixPath(*parts[1:])
  if str(relative) in ("", "."):
   return
  dest = output.joinpath(*relative.parts)
  dest.parent.mkdir(parents=True, exist_ok=True)
  with dest.open("wb") as out:
   shutil.copyfileobj(reader, out)
  seen.add(str(relative))
  if relative.name == executable:
   dest.chmod(0o755)
 if ext == "zip":
  with zipfile.ZipFile(io.BytesIO(archive)) as z:
   for item in z.infolist():
    if not item.is_dir():
     with z.open(item) as stream:
      write_one(item.filename, stream)
 else:
  with tarfile.open(fileobj=io.BytesIO(archive), mode="r:gz") as tar:
   for item in tar:
    if item.isfile():
     with tar.extractfile(item) as stream:
      write_one(item.name, stream)
    elif not item.isdir():
     raise RuntimeError("Unexpected entry in release archive")
 if not {executable, "LICENSE"}.issubset(seen):
  raise RuntimeError("Upstream archive missing executable or LICENSE")
 print(f"Verified {name}: {digest}; packaged at {output}")

if __name__ == "__main__":
 parser = argparse.ArgumentParser()
 parser.add_argument("--platform", choices=list(TARGETS), required=True)
 main(parser.parse_args().platform)
