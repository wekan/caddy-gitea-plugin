#!/usr/bin/env python3
"""Make the two source zips of a release of this plugin.

  source-zips.py <source-dir> <out-dir>

  <out-dir>/caddy-gitea-plugin-source-with-history.zip     caddy-gitea-plugin/ with .git
  <out-dir>/caddy-gitea-plugin-source-without-history.zip  caddy-gitea-plugin/ without it
  and a .sha256 of each, in the format of `sha256sum` (checked with `sha256sum -c`).

<source-dir> is a clone of this repository at the release's commit; the top directory
inside both zips is "caddy-gitea-plugin", whatever it is called.

WHY NOT `zip`: Info-ZIP's zip 3.0 aborts with "buffer overflow detected" on a
non-ASCII file name on Arch Linux, and Gitea's own test fixtures have one
(tests/gitea-repositories-meta/user2/utf8.git/refs/heads/Grüßen). Python's zipfile
writes UTF-8 names correctly everywhere.

THE ZIP WITHOUT HISTORY IS REPRODUCIBLE: entries are sorted, and every entry's
time is the time of the repository's HEAD commit rather than when the
files were checked out, so the same source gives a byte-identical zip on every run. The zip
with history holds the same commits, but not the same bytes from one clone to the
next: git's index records each file's timestamps and inode, and a clone's packs
can differ. File modes are kept
(the executable bit of the build scripts matters); a symbolic link is stored as a
link, the way Info-ZIP stores it and unzip restores it. Files git stores already
compressed (.pack, .idx) are stored, not deflated again.
"""

import hashlib
import os
import stat
import subprocess
import sys
import time
import zipfile

WITH = "caddy-gitea-plugin-source-with-history.zip"
WITHOUT = "caddy-gitea-plugin-source-without-history.zip"
TOP = "caddy-gitea-plugin"
STORED = (".pack", ".idx", ".rev", ".zip", ".gz", ".xz", ".png", ".jpg", ".woff2")


def entries(top, history):
    """Relative paths under top, sorted, directories included (so empty ones survive).
    Without history, every .git - of each repository - is left out."""
    out = []
    for d, dirs, files in os.walk(top):
        rel = os.path.relpath(d, top)
        if not history:
            dirs[:] = [n for n in dirs if n != ".git"]
            files = [n for n in files if n != ".git"]
        # A link to a directory is listed with the directories; it is stored as a link.
        for n in list(dirs):
            if os.path.islink(os.path.join(d, n)):
                dirs.remove(n)
                files.append(n)
        if rel != ".":
            out.append(rel + "/")
        out += [os.path.normpath(os.path.join(rel, n)) for n in files]
    return sorted(out)


def write(top, path, history, when):
    tmp = path + ".tmp"
    with zipfile.ZipFile(tmp, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for rel in entries(top, history):
            full = os.path.join(top, rel)
            st = os.lstat(full)
            info = zipfile.ZipInfo(TOP + "/" + rel.replace(os.sep, "/"), date_time=when)
            info.create_system = 3  # Unix: the mode below is meaningful
            info.external_attr = (st.st_mode & 0xFFFF) << 16
            if rel.endswith("/"):
                info.external_attr |= 0x10  # MS-DOS directory flag
                z.writestr(info, b"")
            elif stat.S_ISLNK(st.st_mode):
                info.compress_type = zipfile.ZIP_STORED
                z.writestr(info, os.readlink(full).encode())
            else:
                info.compress_type = zipfile.ZIP_STORED if rel.endswith(STORED) else zipfile.ZIP_DEFLATED
                with open(full, "rb") as src, z.open(info, "w", force_zip64=True) as dst:
                    while chunk := src.read(1 << 20):
                        dst.write(chunk)
    os.replace(tmp, path)
    h = hashlib.sha256()
    with open(path, "rb") as f:
        while chunk := f.read(1 << 20):
            h.update(chunk)
    with open(path + ".sha256", "w") as f:
        f.write(f"{h.hexdigest()}  {os.path.basename(path)}\n")
    print(f"{os.path.basename(path)}: {os.path.getsize(path) // 1048576} MB, sha256 {h.hexdigest()}")


def main(argv):
    if len(argv) != 3:
        raise SystemExit(__doc__)
    top, out = argv[1], argv[2]
    os.makedirs(out, exist_ok=True)
    stamps = [int(subprocess.run(["git", "-C", os.path.join(top, n), "log", "-1", "--format=%ct"],
                                 capture_output=True, text=True, check=True).stdout)
              for n in ["."] if os.path.isdir(os.path.join(top, ".git"))]
    if not stamps:
        raise SystemExit(f"{top}: no git repositories in it")
    stamp = max(stamps)
    # zip stores local time with no zone; UTC keeps it the same on every machine.
    when = time.gmtime(max(stamp, 315532800))[:6]  # zip cannot store a time before 1980
    write(top, os.path.join(out, WITHOUT), False, when)
    write(top, os.path.join(out, WITH), True, when)


if __name__ == "__main__":
    main(sys.argv)
