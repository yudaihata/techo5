#!/usr/bin/env python3
"""Download the pinned, open-licensed Japanese fallback font into the supplied folder."""
import hashlib
import os
from pathlib import Path
import sys
import urllib.request

COMMIT = 'f8d157532fbfaeda587e826d4cd5b21a49186f7c'
NAME = 'NotoSansCJKjp-Regular.otf'
SHA256 = '68a3fc98800b2a27b371f2fb79991daf3633bd89309d4ffaa6946fd587f375b5'
BASE = f'https://raw.githubusercontent.com/notofonts/noto-cjk/{COMMIT}/Sans/'

def main():
    if len(sys.argv) != 2:
        raise SystemExit('usage: fetch-font.py OUTPUT_DIRECTORY')
    folder = Path(sys.argv[1])
    folder.mkdir(parents=True, exist_ok=True)
    target = folder / NAME
    if target.exists() and hashlib.sha256(target.read_bytes()).hexdigest() == SHA256:
        print(f'Font already verified: {target}')
        return
    data = urllib.request.urlopen(BASE + 'OTF/Japanese/' + NAME, timeout=60).read()
    if hashlib.sha256(data).hexdigest() != SHA256:
        raise SystemExit('font checksum mismatch; existing font was not changed')
    license_data = urllib.request.urlopen(BASE + 'LICENSE', timeout=60).read()
    tmp = folder / (NAME + '.part')
    try:
        tmp.write_bytes(data)
        os.replace(tmp, target)
    finally:
        tmp.unlink(missing_ok=True)
    (folder / 'NotoSansCJK-LICENSE.txt').write_bytes(license_data)
    print(f'Verified font saved: {target}')

if __name__ == '__main__':
    main()
