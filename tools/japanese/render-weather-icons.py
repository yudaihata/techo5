#!/usr/bin/env python3
"""Render vendored Meteocons Flat SVGs. Dev dependencies: CairoSVG2.8.2, Pillow12.0.0, Cairo."""
import hashlib,io,json
from pathlib import Path
import cairosvg
from PIL import Image
folder=Path(__file__).resolve().parents[2]/'echod/internal/feature/display/assets/meteocons-flat'
meta=json.loads((folder/'source.json').read_text())
for name,expected in meta['svg_sha256'].items():
    svg=(folder/'svg'/name).read_bytes()
    if hashlib.sha256(svg).hexdigest()!=expected:raise SystemExit('SVG hash mismatch: '+name)
    png=cairosvg.svg2png(bytestring=svg,output_width=512,output_height=512)
    img=Image.open(io.BytesIO(png)).convert('RGBA')
    bounds=img.getchannel('A').getbbox()
    if not bounds:raise SystemExit('Empty icon: '+name)
    img=img.crop(bounds)
    img.thumbnail((88,88),Image.Resampling.LANCZOS)
    out=Image.new('RGBA',(96,96))
    out.alpha_composite(img,((96-img.width)//2,(96-img.height)//2))
    out.save(folder/(Path(name).stem+'.png'),optimize=True)
print('Rendered',len(meta['svg_sha256']),'96px transparent icons')
