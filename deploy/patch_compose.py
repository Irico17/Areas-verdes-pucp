#!/usr/bin/env python3
"""Reescribe las imágenes de api y web en el compose ya renderizado de la EC2.

No toca volúmenes ni monta SQL. Imprime las imágenes anteriores en stdout.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def _section_at(line: str) -> str | None:
    if line.startswith("  ") and not line.startswith("   ") and line.endswith(":"):
        name = line.strip()[:-1]
        if name in {"db", "api", "web"}:
            return name
    return None


def images_of(text: str) -> tuple[str, str]:
    section = None
    api = ""
    web = ""
    for line in text.splitlines():
        found = _section_at(line)
        if found:
            section = found
            continue
        if section in {"api", "web"} and line.startswith("    image:"):
            value = line.split(":", 1)[1].strip().strip("'\"")
            if section == "api" and not api:
                api = value
            elif section == "web" and not web:
                web = value
    return api, web


def patch(text: str, api: str, web: str) -> str:
    section = None
    lines = text.splitlines()
    seen_api = False
    seen_web = False
    for i, line in enumerate(lines):
        found = _section_at(line)
        if found:
            section = found
            continue
        if not line.startswith("    image:"):
            continue
        if section == "api" and not seen_api:
            lines[i] = f"    image: {api}"
            seen_api = True
        elif section == "web" and not seen_web:
            lines[i] = f"    image: {web}"
            seen_web = True
    if not seen_api or not seen_web:
        raise SystemExit("el compose no tiene image: en api y en web")
    return "\n".join(lines) + "\n"


def _selftest() -> None:
    sample = "services:\n  api:\n    image: old-api\n  web:\n    image: old-web\n  db:\n    image: postgis\n"
    previous_api, previous_web = images_of(sample)
    if previous_api != "old-api" or previous_web != "old-web":
        raise SystemExit(f"lectura inesperada: {previous_api} {previous_web}")
    out = patch(sample, "new-api", "new-web")
    if "image: new-api" not in out or "image: new-web" not in out:
        raise SystemExit(out)
    if "image: postgis" not in out:
        raise SystemExit("cambió la imagen de db")
    print("patch_compose ok")


def main() -> None:
    if len(sys.argv) == 2 and sys.argv[1] == "--selftest":
        _selftest()
        return
    parser = argparse.ArgumentParser(description="Actualiza las imágenes api y web de un compose")
    parser.add_argument("compose")
    parser.add_argument("--api", required=True)
    parser.add_argument("--web", required=True)
    args = parser.parse_args()
    path = Path(args.compose)
    text = path.read_text(encoding="utf-8")
    api, web = images_of(text)
    print(f"PREVIOUS_API={api}")
    print(f"PREVIOUS_WEB={web}")
    path.write_text(patch(text, args.api, args.web), encoding="utf-8")


if __name__ == "__main__":
    main()
