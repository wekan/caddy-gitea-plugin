#!/usr/bin/env python3
"""Copy the parts of Gitea this plugin needs into gitea/, as part of this Go module.

  copy-gitea.py <gitea-source> <plugin-dir>

<gitea-source> is a patched Gitea checkout whose frontend is built (public/assets) -
releases/sync-gitea.sh makes one. gitea/ in <plugin-dir> is replaced whole, so a
package Gitea no longer needs disappears with it.

WHAT IS NEEDED is computed, not listed: every Gitea package that the plugin's own
Go files (the root package) import, and everything those import in turn - `go list
-deps`, for every operating system Weddy is built for - and nothing else: no tests,
no cmd/, no frontend sources, docs, tools or test fixtures.

WHAT CHANGES on the way, so it builds with a plain `go build` / xcaddy, with CGO off
and no build tags:

  - imports "gitea.dev/..." become "github.com/wekan/caddy-gitea-plugin/gitea/...";
  - Gitea's assets are always embedded: in the packages that embed bindata.dat
    (public, templates, options, migration schemas), the asset files themselves are
    copied to assets/ and embedded with //go:embed - text that git stores and diffs
    well, where bindata.dat is a blob of gzipped files - and their !bindata
    variants are left out;
  - files only for cgo builds (sqlite_mattn, pam) or tests are left out.
"""

import json
import os
import re
import shutil
import subprocess
import sys

MODULE = "github.com/wekan/caddy-gitea-plugin"
TARGET = MODULE + "/gitea"
GITEA = "gitea.dev"
GOOS = ["linux", "windows", "darwin", "freebsd", "openbsd", "netbsd"]
CGO_OR_TEST_TAGS = {"sqlite_mattn", "pam", "test_avatar_identicon"}
BUILD = re.compile(r"^//go:build (.+)$", re.M)


def run(cmd, cwd, env=None):
    r = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, env=dict(os.environ, **(env or {})))
    if r.returncode:
        raise SystemExit(f"{' '.join(cmd)}: {r.stderr.strip()[-2000:]}")
    return r.stdout


def entries(plugin):
    """The Gitea packages the plugin's root package imports."""
    out = set()
    for n in os.listdir(plugin):
        if n.endswith(".go") and not n.endswith("_test.go"):
            for m in re.finditer(r'"' + re.escape(TARGET) + r'(/[^"]*)?"', open(os.path.join(plugin, n)).read()):
                out.add(GITEA + (m.group(1) or ""))
    if not out:
        raise SystemExit("the plugin's Go files import no Gitea package")
    return sorted(out)


def packages(src, roots):
    """{import path: dir} of every Gitea package needed, on every OS."""
    found = {}
    for goos in GOOS:
        text = run(["go", "list", "-deps", "-tags", "bindata", "-e", "-json=ImportPath,Dir,Module,Error"] + roots, src,
                   {"CGO_ENABLED": "0", "GOOS": goos, "GOFLAGS": "-mod=mod"})
        for doc in re.split(r"(?<=\n})\n(?={)", text.strip()):
            p = json.loads(doc)
            if (p.get("Module") or {}).get("Path") != GITEA:
                continue
            if p.get("Error"):
                raise SystemExit(f"{goos}: {p['ImportPath']}: {p['Error'].get('Err')}")
            found[p["ImportPath"]] = p["Dir"]
    return found


def constraint(text):
    m = BUILD.search(text.split("\npackage ", 1)[0])
    return m.group(1).strip() if m else ""


def keep(name, text):
    if name.endswith("_test.go"):
        return False
    c = constraint(text)
    if c == "!bindata":
        return False
    tags = set(re.findall(r"(?<![!\w])(\w+)", c))
    return not (tags & CGO_OR_TEST_TAGS)


def rewrite(text, copied):
    """Imports of the COPIED packages point into gitea/. Other modules under gitea.dev
    (gitea.dev/sdk, gitea.dev/actionslib) are dependencies with versions of their own,
    and stay as they are."""
    def one(m):
        path = m.group(1)
        return '"' + TARGET + path[len(GITEA):] + '"' if path in copied else m.group(0)
    return re.sub(r'"(' + re.escape(GITEA) + r'(?:/[^"\s]*)?)"', one, text)


def embed_assets(text, pkg_src, gitea_src, out_dir):
    """A bindata file -> the same variable from //go:embed of the asset files."""
    gen = re.search(r"^//go:generate go run \S*generate-bindata\.go (\S+) bindata\.dat$", text, re.M)
    if not gen:
        raise SystemExit(f"{pkg_src}: embeds bindata.dat but names no source directory")
    assets_src = os.path.normpath(os.path.join(pkg_src, gen.group(1)))
    shutil.copytree(assets_src, os.path.join(out_dir, "assets"))
    text = BUILD.sub("", text, count=1)
    text = re.sub(r"^//go:generate .*\n", "", text, flags=re.M)
    text = text.replace("//go:embed bindata.dat\nvar bindata []byte",
                        "// Copied from " + os.path.relpath(assets_src, gitea_src) + " by releases/copy-gitea.py.\n"
                        "//\n//go:embed all:assets\nvar bindata embed.FS")
    text = text.replace("assetfs.NewEmbeddedFS(bindata)", "embeddedAssets()")
    text = text.replace('_ "embed"', '"embed"' + ('' if '"io/fs"' in text else '\n\t"io/fs"'))
    text += ('\n// embeddedAssets is assets/ as the root of a file system.\n'
             'func embeddedAssets() fs.FS {\n\tsub, err := fs.Sub(bindata, "assets")\n'
             '\tif err != nil {\n\t\tpanic(err)\n\t}\n\treturn sub\n}\n')
    if "assetfs." not in text:  # it was used only for NewEmbeddedFS (migration schemas)
        text = re.sub(r'^\t"[^"]*/modules/assetfs"\n', "", text, flags=re.M)
    if "embeddedAssets()" not in text.split("func embeddedAssets")[0]:
        raise SystemExit(f"{pkg_src}: the bindata file has an unexpected shape; update copy-gitea.py")
    return text


def main(argv):
    if len(argv) != 3:
        raise SystemExit(__doc__)
    src, plugin = os.path.abspath(argv[1]), os.path.abspath(argv[2])
    if not os.path.isdir(os.path.join(src, "public", "assets")):
        raise SystemExit(f"{src}/public/assets is missing: build Gitea's frontend first (make deps-frontend frontend)")
    roots = entries(plugin)
    pkgs = packages(src, roots)
    out = os.path.join(plugin, "gitea")
    shutil.rmtree(out, ignore_errors=True)
    files = assets = 0
    for path, d in sorted(pkgs.items()):
        rel = os.path.relpath(d, src)
        dest = os.path.join(out, rel)
        os.makedirs(dest, exist_ok=True)
        for n in sorted(os.listdir(d)):
            if not n.endswith(".go"):
                continue
            text = open(os.path.join(d, n), encoding="utf-8").read()
            if not keep(n, text):
                continue
            if constraint(text) == "bindata":
                if "//go:embed bindata.dat" in text:
                    text = embed_assets(text, d, src, dest)
                    assets += 1
                else:
                    text = BUILD.sub("", text, count=1)
            open(os.path.join(dest, n), "w", encoding="utf-8").write(rewrite(text, pkgs))
            files += 1
        # Other files the package embeds (templates of its own, schemas, ...).
        # Without the bindata tag: with it, bindata.dat - never generated here - is missing.
        for goos in ("linux",):
            info = json.loads(run(["go", "list", "-e", "-json=EmbedFiles", "."], d,
                                  {"CGO_ENABLED": "0", "GOOS": goos, "GOFLAGS": "-mod=mod"}))
            for f in info.get("EmbedFiles") or []:
                if f == "bindata.dat":
                    continue
                os.makedirs(os.path.dirname(os.path.join(dest, f)), exist_ok=True)
                shutil.copy2(os.path.join(d, f), os.path.join(dest, f))
    shutil.copy2(os.path.join(src, "LICENSE"), os.path.join(out, "LICENSE"))
    run(["gofmt", "-w", out], plugin)
    print(f"gitea/: {len(pkgs)} packages, {files} Go files, {assets} embedded asset trees, from {len(roots)} entry packages.")


if __name__ == "__main__":
    main(sys.argv)
