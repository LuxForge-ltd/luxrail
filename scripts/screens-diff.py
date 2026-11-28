#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Compares two folders of console screenshots (web/e2e/specs/visual/screens.e2e.ts).

    python3 scripts/screens-diff.py BEFORE_DIR AFTER_DIR OUT_DIR

For every PNG in both folders it counts the pixels that changed by more than a
small threshold, prints one line per screen, and writes OUT_DIR/<name> with
before, after and the changed pixels in red side by side for every screen
that differs. Needs Pillow and NumPy. Exits 1 when any screen differs or is
missing from one side, so a run can gate on it.
"""
import sys
from pathlib import Path

import numpy as np
from PIL import Image

THRESHOLD = 24  # largest per-channel change still treated as the same pixel (antialiasing)


def load(path: Path) -> np.ndarray:
    return np.asarray(Image.open(path).convert("RGB"), dtype=np.int16)


def pad(a: np.ndarray, h: int, w: int) -> np.ndarray:
    out = np.zeros((h, w, 3), dtype=np.int16)
    out[: a.shape[0], : a.shape[1]] = a
    return out


def main() -> int:
    if len(sys.argv) != 4:
        print(__doc__)
        return 2
    before, after, out = (Path(p) for p in sys.argv[1:])
    out.mkdir(parents=True, exist_ok=True)
    names = sorted({p.name for p in before.glob("*.png")} | {p.name for p in after.glob("*.png")})
    differ = 0
    for name in names:
        a, b = before / name, after / name
        if not a.exists() or not b.exists():
            print(f"{name}: only in {'after' if not a.exists() else 'before'}")
            differ += 1
            continue
        x, y = load(a), load(b)
        h, w = max(x.shape[0], y.shape[0]), max(x.shape[1], y.shape[1])
        size = "" if x.shape == y.shape else f", size {x.shape[1]}x{x.shape[0]} -> {y.shape[1]}x{y.shape[0]}"
        x, y = pad(x, h, w), pad(y, h, w)
        mask = (np.abs(x - y) > THRESHOLD).any(axis=2)
        changed = int(mask.sum())
        if changed == 0 and not size:
            print(f"{name}: same")
            continue
        differ += 1
        rows, cols = np.nonzero(mask)
        box = f", box x {cols.min()}-{cols.max()} y {rows.min()}-{rows.max()}" if changed else ""
        print(f"{name}: {changed} pixels ({100 * changed / (h * w):.2f}%){box}{size}")
        marked = y.copy()
        marked[mask] = (255, 0, 0)
        sheet = np.concatenate([x, np.full((h, 8, 3), 128, np.int16), y, np.full((h, 8, 3), 128, np.int16), marked], axis=1)
        Image.fromarray(sheet.astype(np.uint8)).save(out / name)
    print(f"{differ} of {len(names)} screens differ")
    return 1 if differ else 0


if __name__ == "__main__":
    sys.exit(main())
