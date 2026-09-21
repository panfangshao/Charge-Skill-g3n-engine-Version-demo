"""Diff a frame of this port against the same frame of the Godot build.

    python tools/compare.py godot.png rbfx.png

Prints the mean per-channel error over the whole frame and over the regions
that matter, plus the exact colours at a few probe points.
"""

from __future__ import annotations

import sys

from PIL import Image

# (name, left, top, right, bottom) in 1280x720 design pixels.
REGIONS = [
    ("whole frame", 0, 0, 1280, 720),
    ("arena (3D)", 280, 200, 1000, 460),
    ("bottom panel", 24, 598, 1256, 696),
    ("title / instructions", 24, 20, 900, 110),
]

PROBES = [
    ("background", 40, 300),
    ("ground", 640, 400),
    ("player capsule", 499, 300),
    ("target capsule", 753, 300),
    ("panel fill", 600, 640),
    ("button fill", 1120, 648),
]


def mean_error(a: Image.Image, b: Image.Image, box) -> float:
    ca = a.crop(box).tobytes()
    cb = b.crop(box).tobytes()
    return sum(abs(x - y) for x, y in zip(ca, cb)) / len(ca)


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(__doc__)
        return 2

    reference = Image.open(argv[0]).convert("RGB")
    port = Image.open(argv[1]).convert("RGB")
    if reference.size != port.size:
        print(f"size mismatch: {reference.size} vs {port.size}")
        return 1

    print(f"{'region':<24} {'mean per-channel error (0-255)':>30}")
    for name, *box in REGIONS:
        print(f"{name:<24} {mean_error(reference, port, tuple(box)):>30.2f}")

    print()
    print(f"{'probe':<20} {'godot':>16} {'rbfx':>16}")
    for name, x, y in PROBES:
        print(f"{name:<20} {str(reference.getpixel((x, y))):>16} {str(port.getpixel((x, y))):>16}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
