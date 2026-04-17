#!/usr/bin/env python3
"""Build a cleaned prompt library copy for Comfy Manager.

This script reads the merged prompt data from the source prompt directory,
normalizes the records, removes exact duplicates, and writes a UTF-8 runtime
copy into data/prompt-library/.
"""

from __future__ import annotations

import collections
import datetime as dt
import hashlib
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DATA_DIR = ROOT / "data" / "prompt-library"


def find_prompt_source_dir(root: Path) -> Path:
    for child in root.iterdir():
        if child.is_dir() and (child / "downloaded_prompt_data").exists():
            return child / "downloaded_prompt_data"
    raise FileNotFoundError("Could not find downloaded_prompt_data source directory")


def normalize_text(value: object) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    return str(value).strip()


def build() -> None:
    source_dir = find_prompt_source_dir(ROOT)
    source_file = source_dir / "all_prompts_merged.json"
    output_file = DATA_DIR / "all_prompts_merged.cleaned.json"
    manifest_file = DATA_DIR / "manifest.cleaned.json"

    raw_data = json.loads(source_file.read_text(encoding="utf-8"))
    DATA_DIR.mkdir(parents=True, exist_ok=True)

    trimmed_fields: collections.Counter[str] = collections.Counter()
    null_fields_filled: collections.Counter[str] = collections.Counter()
    source_counter: collections.Counter[str] = collections.Counter()
    category_counter: collections.Counter[str] = collections.Counter()
    scope_counter: collections.Counter[str] = collections.Counter()

    cleaned: list[dict[str, str]] = []
    seen: set[tuple[str, ...]] = set()
    duplicates_removed = 0

    for row in raw_data:
        item: dict[str, str] = {}
        for key in [
            "source",
            "category",
            "subcategory",
            "scope",
            "text_en",
            "text_zh",
            "preview",
            "extra_id",
        ]:
            original = row.get(key, "")
            if original is None:
                null_fields_filled[key] += 1
            value = normalize_text(original)
            if isinstance(original, str) and value != original:
                trimmed_fields[key] += 1
            item[key] = value

        signature = (
            item["source"].lower(),
            item["category"].lower(),
            item["subcategory"].lower(),
            item["scope"].lower(),
            item["text_en"].lower(),
            item["text_zh"],
            item["preview"],
            item["extra_id"],
        )
        if signature in seen:
            duplicates_removed += 1
            continue
        seen.add(signature)

        stable_id_source = "|".join(signature)
        item["id"] = hashlib.sha1(stable_id_source.encode("utf-8")).hexdigest()[:16]
        item["search_text"] = " ".join(
            [
                item["source"],
                item["category"],
                item["subcategory"],
                item["scope"],
                item["text_en"],
                item["text_zh"],
            ]
        ).strip().lower()

        cleaned.append(item)
        source_counter[item["source"]] += 1
        category_counter[item["category"]] += 1
        scope_counter[item["scope"]] += 1

    output_file.write_text(
        json.dumps(cleaned, ensure_ascii=False, indent=2),
        encoding="utf-8",
        newline="\n",
    )

    manifest = {
        "created_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "source_file": str(source_file),
        "output_file": str(output_file),
        "encoding": "utf-8",
        "record_count_before": len(raw_data),
        "record_count_after": len(cleaned),
        "duplicates_removed": duplicates_removed,
        "trimmed_fields": dict(trimmed_fields),
        "null_fields_filled": dict(null_fields_filled),
        "sources": dict(source_counter),
        "scopes": dict(scope_counter),
        "top_categories": dict(category_counter.most_common(20)),
        "notes": [
            "Runtime should read the cleaned copy in data/prompt-library.",
            "Do not read the original merged file from the source prompt folder at runtime.",
            "Chinese text is preserved in UTF-8.",
        ],
    }
    manifest_file.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2),
        encoding="utf-8",
        newline="\n",
    )

    print(f"Wrote {output_file}")
    print(f"Wrote {manifest_file}")
    print(f"Before: {len(raw_data)}")
    print(f"After: {len(cleaned)}")
    print(f"Duplicates removed: {duplicates_removed}")


if __name__ == "__main__":
    build()
