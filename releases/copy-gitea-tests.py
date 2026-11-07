#!/usr/bin/env python3
"""Copy Gitea's unit tests for the packages in gitea/ into a COPY of this plugin.

  copy-gitea-tests.py <gitea-source> <plugin-copy>

<gitea-source> is the patched Gitea checkout gitea/ was copied from - the
"Resulting Gitea commit" of gitea/SOURCE.md, which releases/sync-gitea.sh makes.
<plugin-copy> is a copy of this repository - never this repository itself: gitea/
here stays free of tests, so `go get` of the plugin does not download them. Weddy's
tests/gitea/run.sh makes the copy, runs this, and runs `go test` there.

WHAT IS COPIED, the same way releases/copy-gitea.py copies gitea/ (imports
"gitea.dev/..." rewritten into this module, bindata files embedding assets/):

  - every *_test.go file of every package that is in gitea/;
  - the packages those tests import that gitea/ does not have - models/unittest,
    modules/test, tests, ... - computed with `go list -deps -test`, without their
    own tests;
  - the data the tests read: each package's non-Go files and directories (testdata,
    fixtures, ...), and from the Gitea root what models/unittest loads -
    models/fixtures and tests/gitea-repositories-meta, tests/gitea-lfs-meta,
    tests/testdata - and tests/sqlite.ini.tmpl for modelmigration's tests;
  - go.mod requirements for test-only modules, at the versions Gitea's go.mod has.
"""

import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("copy_gitea", os.path.join(HERE, "copy-gitea.py"))
cg = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cg)

# Read by models/unittest and modules/setting from the Gitea root, which in the copy
# is gitea/ (detectGiteaTestRoot: three directories above modules/setting/testenv.go).
# modelmigration's tests also write their app.ini from tests/sqlite.ini.tmpl into a
# temporary directory under tests/integration (setting.PrepareIntegrationTestConfig).
ROOT_DATA = ["models/fixtures", "tests/gitea-repositories-meta", "tests/gitea-lfs-meta", "tests/testdata",
             "tests/sqlite.ini.tmpl"]
ROOT_DIRS = ["tests/integration"]


def go_dirs(top):
    """Directories under top that hold Go files (outside testdata)."""
    out = set()
    for d, subdirs, files in os.walk(top):
        subdirs[:] = [s for s in subdirs if s != "testdata" and not s.startswith((".", "_"))]
        if any(f.endswith(".go") for f in files):
            out.add(os.path.relpath(d, top))
    return out


def has_go(d):
    for _, subdirs, files in os.walk(d):
        subdirs[:] = [s for s in subdirs if s != "testdata"]
        if any(f.endswith(".go") for f in files):
            return True
    return False


def copy_data(src_dir, dest_dir):
    """The non-Go files of a package directory, and its subdirectories that are not
    Go packages (testdata, fixtures, ...). Nothing already in dest is overwritten:
    assets/ there is the plugin's own."""
    n = 0
    for name in sorted(os.listdir(src_dir)):
        s, d = os.path.join(src_dir, name), os.path.join(dest_dir, name)
        if os.path.exists(d) or name in ("node_modules",):
            continue
        if os.path.isdir(s):
            if name == "testdata" or not has_go(s):
                shutil.copytree(s, d, symlinks=True)
                n += 1
        elif not name.endswith(".go"):
            shutil.copy2(s, d, follow_symlinks=False)
            n += 1
    return n


def test_deps(src, tested):
    """{import path: dir} of the Gitea packages the tests of `tested` need."""
    text = cg.run(["go", "list", "-deps", "-test", "-tags", "bindata", "-e",
                   "-json=ImportPath,Dir,Module,ForTest"] + tested, src,
                  {"CGO_ENABLED": "0", "GOOS": "linux", "GOFLAGS": "-mod=mod"})
    found = {}
    for doc in re.split(r"(?<=\n})\n(?={)", text.strip()):
        p = json.loads(doc)
        if (p.get("Module") or {}).get("Path") != cg.GITEA or not p.get("Dir"):
            continue
        path = p["ImportPath"].split(" ", 1)[0]
        if path.endswith(".test"):
            continue
        found[path] = p["Dir"]
    return found


def copy_package(path, d, src, dest, copied, tests):
    """Copy one package's Go files: its code unless it is already there, and its tests if asked."""
    os.makedirs(dest, exist_ok=True)
    files = 0
    for n in sorted(os.listdir(d)):
        if not n.endswith(".go"):
            continue
        is_test = n.endswith("_test.go")
        if is_test and not tests:
            continue
        if not is_test and tests:
            continue  # gitea/ already has the code of a tested package
        text = open(os.path.join(d, n), encoding="utf-8").read()
        if is_test:
            # Tests only for cgo builds (sqlite_mattn, pam) would not build here.
            tags = set(re.findall(r"(?<![!\w])(\w+)", cg.constraint(text)))
            if tags & (cg.CGO_OR_TEST_TAGS - {"test_avatar_identicon"}):
                continue
        else:
            if not cg.keep(n, text):
                continue
            if cg.constraint(text) == "bindata":
                if "//go:embed bindata.dat" in text:
                    text = cg.embed_assets(text, d, src, dest)
                else:
                    text = cg.BUILD.sub("", text, count=1)
        open(os.path.join(dest, n), "w", encoding="utf-8").write(cg.rewrite(text, copied))
        files += 1
    return files


def add_requirements(src, plugin):
    """Modules the tests need that the plugin's go.mod does not name, at Gitea's versions."""
    have = set(cg.run(["go", "list", "-m", "-mod=mod", "-f", "{{.Path}}", "all"], plugin,
                      {"GOFLAGS": "-mod=mod"}).split())
    gitea = cg.run(["go", "list", "-m", "-f", "{{if not .Main}}{{.Path}}@{{.Version}}{{end}}", "all"], src)
    reqs = [f"-require={r}" for r in gitea.split() if "@" in r and r.split("@")[0] not in have and r.split("@")[1]]
    if reqs:
        cg.run(["go", "mod", "edit"] + reqs, plugin)
    cg.run(["go", "mod", "tidy"], plugin, {"GOFLAGS": "-mod=mod"})
    return len(reqs)


def main(argv):
    if len(argv) != 3:
        raise SystemExit(__doc__)
    src, plugin = os.path.abspath(argv[1]), os.path.abspath(argv[2])
    if os.path.realpath(plugin) == os.path.realpath(os.path.dirname(HERE)):
        raise SystemExit("refusing to copy tests into this repository: give a copy of it")
    out = os.path.join(plugin, "gitea")
    if not os.path.isdir(out):
        raise SystemExit(f"{out} is missing: <plugin-copy> must be a copy of caddy-gitea-plugin")
    if not os.path.isfile(os.path.join(src, "go.mod")):
        raise SystemExit(f"{src} is not a Gitea source tree")

    have = sorted(r for r in go_dirs(out) if os.path.isdir(os.path.join(src, r)))
    tested = [cg.GITEA + ("" if r == "." else "/" + r) for r in have]
    deps = test_deps(src, tested)
    copied = set(deps) | set(tested)

    extra = files = tests = data = 0
    for path, d in sorted(deps.items()):
        rel = os.path.relpath(d, src)
        if rel in have:
            continue
        files += copy_package(path, d, src, os.path.join(out, rel), copied, tests=False)
        data += copy_data(d, os.path.join(out, rel))
        extra += 1
    for rel in have:
        d = os.path.join(src, rel)
        tests += copy_package(cg.GITEA + "/" + rel, d, src, os.path.join(out, rel), copied, tests=True)
        data += copy_data(d, os.path.join(out, rel))
    for rel in ROOT_DATA:
        s, d = os.path.join(src, rel), os.path.join(out, rel)
        if os.path.exists(d) or not os.path.exists(s):
            continue
        if os.path.isdir(s):
            shutil.copytree(s, d, symlinks=True)
        else:
            os.makedirs(os.path.dirname(d), exist_ok=True)
            shutil.copy2(s, d)
        data += 1
    for rel in ROOT_DIRS:
        os.makedirs(os.path.join(out, rel), exist_ok=True)
    cg.run(["gofmt", "-w", out], plugin)
    reqs = add_requirements(src, plugin)
    print(f"gitea/ tests: {tests} test files in {len(have)} packages, {extra} test-only packages "
          f"({files} Go files), {data} data files/directories, {reqs} added module requirements.")


if __name__ == "__main__":
    main(sys.argv)
