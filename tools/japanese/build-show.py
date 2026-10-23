#!/usr/bin/env python3
"""Build a personal ARM Show rootfs from a signed official release; never install or flash it."""
import argparse
import copy
import datetime
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tarfile

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'tools'))
from techo5lib import Release

def sha(path):
    with open(path, 'rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=ROOT / 'bin' / 'japanese')
    args = parser.parse_args()
    def git(*a):
        return subprocess.check_output(['git', '-C', str(ROOT), *a], text=True).strip()
    if git('status', '--porcelain'):
        raise SystemExit('Commit or stash source changes before building.')
    base = (ROOT / 'tools/japanese/upstream-version.txt').read_text().strip()
    commit = git('rev-parse', 'HEAD')
    version = base + '+ja.' + commit[:7]
    date = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    folder = args.output.resolve()
    folder.mkdir(parents=True, exist_ok=True)
    release = Release('HuskerMinion/techo5', base, str(folder))
    if release.version != base:
        raise SystemExit('Signed release version does not match the requested base.')
    official = Path(release.rootfs('arm'))
    binary = folder / ('techo5-' + version + '-arm')
    pkg = 'github.com/HuskerMinion/techo5/echod/internal/layout'
    flags = f'-s -w -X {pkg}.Version={version} -X {pkg}.GitCommit={commit[:7]} -X {pkg}.BuildDate={date}'
    env = dict(os.environ, GOOS='linux', GOARCH='arm', GOARM='7', CGO_ENABLED='0')
    subprocess.run(['go', 'build', '-trimpath', '-ldflags', flags, '-o', str(binary), './cmd/echod'],
                   cwd=ROOT / 'echod', env=env, check=True)
    output = folder / ('rootfs-' + version + '.tar.gz')
    replacements = {'usr/local/bin/techo5': binary.read_bytes(),
                    'etc/techo5-release': f'techo5 rootfs {version}, Japanese fork, base {base}, commit {commit[:7]}\n'.encode()}
    changed = set()
    with tarfile.open(official, 'r:gz') as source, tarfile.open(output, 'w:gz') as target:
        for original in source:
            member = copy.copy(original)
            key = member.name.lstrip('./')
            if key in replacements:
                if not member.isfile() or key in changed:
                    raise SystemExit('Unexpected official archive structure.')
                data = replacements[key]
                member.size = len(data)
                target.addfile(member, io.BytesIO(data))
                changed.add(key)
            else:
                target.addfile(member, source.extractfile(original) if original.isfile() else None)
    if changed != replacements.keys():
        raise SystemExit('The official rootfs is missing the files to replace.')
    # Verify all original members, their metadata and unchanged contents after repacking.
    with tarfile.open(official, 'r:gz') as source, tarfile.open(output, 'r:gz') as target:
        before, after = source.getmembers(), target.getmembers()
        if len(before) != len(after):
            raise SystemExit('Archive member count changed.')
        for a, b in zip(before, after):
            fields = ('name', 'mode', 'uid', 'gid', 'type', 'linkname', 'mtime', 'uname', 'gname', 'devmajor', 'devminor', 'pax_headers')
            if any(getattr(a, field) != getattr(b, field) for field in fields):
                raise SystemExit('Archive metadata changed: ' + a.name)
            if a.isfile():
                expected = replacements.get(a.name.lstrip('./'))
                if expected is None:
                    expected = source.extractfile(a).read()
                if target.extractfile(b).read() != expected:
                    raise SystemExit('Archive contents mismatch: ' + a.name)
    metadata = dict(base=base, commit=commit, version=version, built_at=date,
                    binary_sha256=sha(binary), rootfs_sha256=sha(output), rootfs_bytes=output.stat().st_size,
                    changed_archive_entries=sorted(changed), members=len(before))
    (folder / 'build.json').write_text(json.dumps(metadata, indent=2) + '\n')
    print(json.dumps(metadata, indent=2))
    print('Verified personal rootfs:', output)

if __name__ == '__main__':
    main()
