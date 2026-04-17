from __future__ import annotations

import csv
import json
import re
import shutil
import subprocess
import tempfile
from collections import Counter
from dataclasses import dataclass
from datetime import datetime, timezone
from html import unescape
from html.parser import HTMLParser
from pathlib import Path
from typing import Any

import requests


BASE_DIR = Path(__file__).resolve().parent
OUT_DIR = BASE_DIR / "downloaded_prompt_data"
RAW_DIR = OUT_DIR / "raw"

GWLIANG_URL = "https://prompt.gwliang.com/"
NEWZONE_URL = "https://prompt.newzone.top/app/zh"
QPIPI_URL = "https://prompt.qpipi.com/"
AITAG_URL = "https://aitag.top/"
AITAG_API_BASE = "https://api.aitag.top/"
LOCAL_SD_PROMPT_EDITOR_PATH = Path(r"F:\0-代码\html\AI画风\sd提示词编辑器.html")

UA = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0 Safari/537.36"
)
HEADERS = {
    "User-Agent": UA,
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
}


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def ensure_dirs() -> None:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    RAW_DIR.mkdir(parents=True, exist_ok=True)


def save_json(path: Path, data: Any) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")


def save_text(path: Path, text: str) -> None:
    path.write_text(text, encoding="utf-8")


def get_session() -> requests.Session:
    session = requests.Session()
    session.headers.update(HEADERS)
    return session


def fetch_text(
    session: requests.Session,
    url: str,
    dest: Path,
    *,
    allow_curl_fallback: bool = False,
    force_encoding: str | None = None,
) -> str:
    try:
        response = session.get(url, timeout=60)
        response.raise_for_status()
        if force_encoding:
            text = response.content.decode(force_encoding, errors="replace")
        else:
            text = response.text
    except Exception:
        if not allow_curl_fallback:
            raise
        result = subprocess.run(
            [
                "curl.exe",
                "-L",
                url,
                "-A",
                UA,
                "-H",
                "Accept-Language: zh-CN,zh;q=0.9,en;q=0.8",
            ],
            capture_output=True,
            check=True,
            text=True,
            encoding="utf-8",
            errors="ignore",
        )
        text = result.stdout
    save_text(dest, text)
    return text


def fetch_bytes(session: requests.Session, url: str, dest: Path) -> bytes:
    response = session.get(url, timeout=60)
    response.raise_for_status()
    content = response.content
    dest.write_bytes(content)
    return content


def parse_gwliang_js_url(html: str) -> str:
    match = re.search(r'<script type="module" crossorigin src="([^"]+\.js)">', html)
    if not match:
        raise RuntimeError("Unable to locate gwliang JS asset URL")
    src = match.group(1)
    if src.startswith("http://") or src.startswith("https://"):
        return src
    return GWLIANG_URL.rstrip("/") + src


def extract_gwliang(session: requests.Session) -> dict[str, Any]:
    html = fetch_text(session, GWLIANG_URL, RAW_DIR / "gwliang.html")
    js_url = parse_gwliang_js_url(html)
    js = fetch_bytes(session, js_url, RAW_DIR / "gwliang.js").decode("utf-8", errors="ignore")

    start = js.find("KP={default:[")
    end = js.find("function bc", start)
    if start == -1 or end == -1:
        raise RuntimeError("Unable to locate gwliang data chunk")

    chunk = js[start:end]
    (RAW_DIR / "gwliang.data_chunk.js").write_text(chunk, encoding="utf-8")

    node_code = f"""
const fs = require("fs");
const chunk = fs.readFileSync({json.dumps(str(RAW_DIR / "gwliang.data_chunk.js"))}, "utf8");
const script = "var " + chunk + "\\n" + "process.stdout.write(JSON.stringify({{ pc, gc, fi }}));";
eval(script);
"""
    result = subprocess.run(
        ["node", "-e", node_code],
        capture_output=True,
        check=True,
        text=True,
        encoding="utf-8",
        errors="ignore",
    )
    payload = json.loads(result.stdout)

    flat_items: list[dict[str, Any]] = []
    pc = payload["pc"]
    for top_category, category_body in pc.items():
        for scope in ("normal", "r18"):
            if scope not in category_body:
                continue
            for subgroup, items in category_body[scope].items():
                for idx, item in enumerate(items):
                    flat_items.append(
                        {
                            "source": "gwliang",
                            "top_category": top_category,
                            "scope": scope,
                            "subgroup": subgroup,
                            "index": idx,
                            "en": item.get("en", ""),
                            "zh": item.get("zh", ""),
                        }
                    )

    data = {
        "source": GWLIANG_URL,
        "fetched_at": utc_now_iso(),
        "js_url": js_url,
        "top_categories": list(pc.keys()),
        "categories": pc,
        "all_normal": payload["gc"],
        "all_with_r18": payload["fi"],
        "flat_items": flat_items,
        "summary": {
            "top_category_count": len(pc),
            "flat_item_count": len(flat_items),
            "all_normal_count": len(payload["gc"]),
            "all_with_r18_count": len(payload["fi"]),
        },
    }
    save_json(OUT_DIR / "gwliang.prompts.json", data)
    return data


def extract_newzone(session: requests.Session) -> dict[str, Any]:
    # The site omits a charset header, so requests defaults to ISO-8859-1.
    # We force UTF-8 here to preserve Chinese content before extraction.
    html = fetch_text(session, NEWZONE_URL, RAW_DIR / "newzone.html", force_encoding="utf-8")
    needle = '\\"tagsData\\":['
    start = html.find(needle)
    if start == -1:
        raise RuntimeError("Unable to locate newzone tagsData")

    i = start + len(needle) - 1
    depth = 0
    in_str = False
    escape = False
    end = None
    while i < len(html):
        ch = html[i]
        if escape:
            escape = False
        elif ch == "\\":
            escape = True
        elif ch == '"':
            in_str = not in_str
        elif not in_str:
            if ch == "[":
                depth += 1
            elif ch == "]":
                depth -= 1
                if depth == 0:
                    end = i + 1
                    break
        i += 1

    if end is None:
        raise RuntimeError("Unable to match newzone tagsData brackets")

    raw = html[start + len('\\"tagsData\\":') : end]
    # tagsData is embedded inside an escaped JSON string. Most quotes are escaped
    # once, while some values contain quotes that are escaped twice. We reduce the
    # double-escaped quotes first, then unescape the normal JSON quotes.
    normalized = raw.replace('\\\\\\"', '\\\\"').replace('\\"', '"')
    tags = json.loads(normalized)

    grouped_counts = Counter((item.get("object", ""), item.get("attribute", "")) for item in tags)
    grouped = [
        {
            "object": object_name,
            "attribute": attribute_name,
            "count": count,
        }
        for (object_name, attribute_name), count in sorted(grouped_counts.items())
    ]

    data = {
        "source": NEWZONE_URL,
        "fetched_at": utc_now_iso(),
        "item_count": len(tags),
        "tags": tags,
        "grouped_counts": grouped,
    }
    save_json(OUT_DIR / "newzone.prompts.json", data)
    return data


class QpipiParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.in_tab_title_ul = False
        self.capture_li = False
        self.tab_titles: list[str] = []
        self.current_li: list[str] = []

        self.tab_index = -1
        self.tab_depth: int | None = None
        self.div_depth = 0

        self.capture_p = False
        self.current_p: list[str] = []
        self.current_section: str | None = None

        self.in_tagbutton = False
        self.tagbutton_depth: int | None = None
        self.capture_eng = False
        self.capture_zh = False
        self.current_eng: list[str] = []
        self.current_zh: list[str] = []
        self.items: list[dict[str, Any]] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attrs_dict = dict(attrs)
        if tag == "div":
            self.div_depth += 1
            cls = attrs_dict.get("class", "")
            if cls in ("layui-tab-item", "layui-tab-item layui-show"):
                self.tab_index += 1
                self.tab_depth = self.div_depth
                self.current_section = None
            elif cls == "tagbutton":
                self.in_tagbutton = True
                self.tagbutton_depth = self.div_depth
                self.current_eng = []
                self.current_zh = []
        elif tag == "ul" and attrs_dict.get("class") == "layui-tab-title":
            self.in_tab_title_ul = True
        elif tag == "li" and self.in_tab_title_ul:
            self.capture_li = True
            self.current_li = []
        elif tag == "p" and self.tab_index >= 0 and self.tab_depth is not None and self.div_depth >= self.tab_depth:
            self.capture_p = True
            self.current_p = []
        elif tag == "span":
            cls = attrs_dict.get("class", "")
            if self.in_tagbutton and cls == "english":
                self.capture_eng = True
            elif self.in_tagbutton and cls == "chinese":
                self.capture_zh = True

    def handle_endtag(self, tag: str) -> None:
        if tag == "li" and self.capture_li:
            self.tab_titles.append(unescape("".join(self.current_li).strip()))
            self.capture_li = False
        elif tag == "ul" and self.in_tab_title_ul:
            self.in_tab_title_ul = False
        elif tag == "p" and self.capture_p:
            text = unescape("".join(self.current_p).strip())
            if text:
                self.current_section = text
            self.capture_p = False
        elif tag == "span":
            self.capture_eng = False
            self.capture_zh = False
        elif tag == "div":
            if self.in_tagbutton and self.div_depth == self.tagbutton_depth:
                english = unescape("".join(self.current_eng).strip())
                chinese = unescape("".join(self.current_zh).strip())
                if english or chinese:
                    self.items.append(
                        {
                            "tab_index": self.tab_index,
                            "tab": self.tab_titles[self.tab_index] if 0 <= self.tab_index < len(self.tab_titles) else None,
                            "section": self.current_section,
                            "english": english,
                            "chinese": chinese,
                        }
                    )
                self.in_tagbutton = False
                self.tagbutton_depth = None
            if self.tab_depth is not None and self.div_depth == self.tab_depth:
                self.tab_depth = None
                self.current_section = None
            self.div_depth -= 1

    def handle_data(self, data: str) -> None:
        if self.capture_li:
            self.current_li.append(data)
        if self.capture_p:
            self.current_p.append(data)
        if self.capture_eng:
            self.current_eng.append(data)
        if self.capture_zh:
            self.current_zh.append(data)


def extract_qpipi(session: requests.Session) -> dict[str, Any]:
    html = fetch_text(session, QPIPI_URL, RAW_DIR / "qpipi.html", allow_curl_fallback=True)
    parser = QpipiParser()
    parser.feed(html)

    pos_match = re.search(r'<textarea id="tagarea" class="layui-textarea">(.*?)</textarea>', html, re.S)
    neg_match = re.search(r'<textarea id="tagarea2" class="layui-textarea"[^>]*>(.*?)</textarea>', html, re.S)
    default_positive = unescape(pos_match.group(1).strip()) if pos_match else ""
    default_negative = unescape(neg_match.group(1).strip()) if neg_match else ""

    counts_by_tab = Counter(item["tab"] for item in parser.items)
    counts_by_section = Counter((item["tab"], item["section"]) for item in parser.items)

    data = {
        "source": QPIPI_URL,
        "fetched_at": utc_now_iso(),
        "default_positive": default_positive,
        "default_negative": default_negative,
        "tab_titles": parser.tab_titles,
        "item_count": len(parser.items),
        "items": parser.items,
        "counts_by_tab": dict(counts_by_tab),
        "counts_by_tab_and_section": [
            {
                "tab": tab,
                "section": section,
                "count": count,
            }
            for (tab, section), count in sorted(
                counts_by_section.items(),
                key=lambda item: ((item[0][0] or ""), (item[0][1] or "")),
            )
        ],
    }
    save_json(OUT_DIR / "qpipi.prompts.json", data)
    return data


def extract_aitag(session: requests.Session) -> dict[str, Any]:
    root_html = fetch_text(session, AITAG_URL, RAW_DIR / "aitag.html")
    subs_resp = session.get(AITAG_API_BASE + "tagv2/get_subs", timeout=60)
    subs_resp.raise_for_status()
    subs = subs_resp.json()["result"]
    save_json(RAW_DIR / "aitag.subs.json", subs)

    all_items: list[dict[str, Any]] = []
    pages_summary: list[dict[str, Any]] = []
    seen_ids: set[int] = set()

    for sub in subs:
        page = 1
        while True:
            resp = session.post(
                AITAG_API_BASE + "tagv2",
                json={"method": "get_tags_from_sub", "sub": sub, "page": page},
                timeout=60,
            )
            resp.raise_for_status()
            payload = resp.json()
            result = payload.get("result", [])
            page_data = payload.get("page_data", {})
            pages_summary.append(
                {
                    "sub": sub,
                    "page": page,
                    "count": len(result),
                    "page_data": page_data,
                }
            )
            for item in result:
                item_id = item.get("id")
                if isinstance(item_id, int) and item_id in seen_ids:
                    continue
                if isinstance(item_id, int):
                    seen_ids.add(item_id)
                all_items.append(item)
            if not page_data.get("has_next"):
                break
            page += 1

    counts_by_sub = Counter(item.get("sub", "") for item in all_items)
    data = {
        "source": AITAG_URL,
        "api_base": AITAG_API_BASE,
        "fetched_at": utc_now_iso(),
        "root_html_saved": bool(root_html),
        "subcategories": subs,
        "item_count": len(all_items),
        "items": all_items,
        "counts_by_sub": dict(counts_by_sub),
        "pages_summary": pages_summary,
    }
    save_json(OUT_DIR / "aitag.prompts.json", data)
    return data


def extract_local_sd_prompt_editor() -> dict[str, Any] | None:
    if not LOCAL_SD_PROMPT_EDITOR_PATH.exists():
        return None

    raw_copy_path = RAW_DIR / "local_sd_prompt_editor.html"
    shutil.copy2(LOCAL_SD_PROMPT_EDITOR_PATH, raw_copy_path)

    with tempfile.NamedTemporaryFile(delete=False, suffix=".html") as tmp:
        temp_html_path = Path(tmp.name)
    shutil.copy2(LOCAL_SD_PROMPT_EDITOR_PATH, temp_html_path)

    try:
        node_code = f"""
const fs = require("fs");
const text = fs.readFileSync({json.dumps(str(temp_html_path))}, "utf8");

function extractObject(source, marker) {{
  const start = source.indexOf(marker);
  if (start === -1) throw new Error("Marker not found: " + marker);
  let i = start + marker.length;
  let depth = 0;
  let inStr = false;
  let quote = "";
  let esc = false;
  let end = -1;
  for (; i < source.length; i++) {{
    const ch = source[i];
    if (inStr) {{
      if (esc) esc = false;
      else if (ch === "\\\\") esc = true;
      else if (ch === quote) inStr = false;
    }} else {{
      if (ch === '"' || ch === "'" || ch === "`") {{
        inStr = true;
        quote = ch;
      }} else if (ch === "{{") {{
        depth++;
      }} else if (ch === "}}") {{
        depth--;
        if (depth === 0) {{
          end = i + 1;
          break;
        }}
      }}
    }}
  }}
  if (end === -1) throw new Error("Object end not found for marker: " + marker);
  return source.slice(start + marker.length, end);
}}

function extractArray(source, marker, fromIndex) {{
  const start = source.indexOf(marker, fromIndex || 0);
  if (start === -1) throw new Error("Marker not found: " + marker);
  let i = start + marker.length;
  let depth = 0;
  let inStr = false;
  let quote = "";
  let esc = false;
  let end = -1;
  for (; i < source.length; i++) {{
    const ch = source[i];
    if (inStr) {{
      if (esc) esc = false;
      else if (ch === "\\\\") esc = true;
      else if (ch === quote) inStr = false;
    }} else {{
      if (ch === '"' || ch === "'" || ch === "`") {{
        inStr = true;
        quote = ch;
      }} else if (ch === "[") {{
        depth++;
      }} else if (ch === "]") {{
        depth--;
        if (depth === 0) {{
          end = i + 1;
          break;
        }}
      }}
    }}
  }}
  if (end === -1) throw new Error("Array end not found for marker: " + marker);
  return source.slice(start + marker.length, end);
}}

const raRaw = extractObject(text, "const Ra=");
const fzRaw = extractArray(text, "Fz=", text.indexOf(raRaw));
const Ra = eval("(" + raRaw + ")");
const Fz = eval("(" + fzRaw + ")");
process.stdout.write(JSON.stringify({{ Ra, Fz }}));
"""
        result = subprocess.run(
            ["node", "-e", node_code],
            capture_output=True,
            check=True,
            text=True,
            encoding="utf-8",
            errors="ignore",
        )
        payload = json.loads(result.stdout)
    finally:
        try:
            temp_html_path.unlink(missing_ok=True)
        except Exception:
            pass
    categories = payload["Ra"]
    negative_defaults = payload["Fz"]

    flat_items: list[dict[str, Any]] = []
    for category_name, subcategories in categories.items():
        for subcategory_name, pairs in subcategories.items():
            for index, pair in enumerate(pairs):
                en = pair[0] if len(pair) > 0 else ""
                zh = pair[1] if len(pair) > 1 else ""
                flat_items.append(
                    {
                        "source": "local_sd_prompt_editor",
                        "category": category_name,
                        "subcategory": subcategory_name,
                        "index": index,
                        "en": en,
                        "zh": zh,
                    }
                )

    data = {
        "source_file": str(LOCAL_SD_PROMPT_EDITOR_PATH),
        "fetched_at": utc_now_iso(),
        "category_count": len(categories),
        "subcategory_count": sum(len(subcategories) for subcategories in categories.values()),
        "item_count": len(flat_items),
        "categories": categories,
        "flat_items": flat_items,
        "negative_defaults": negative_defaults,
        "negative_default_count": len(negative_defaults),
    }
    save_json(OUT_DIR / "local_sd_prompt_editor.prompts.json", data)
    return data


def build_merged_rows(
    gwliang: dict[str, Any],
    newzone: dict[str, Any],
    qpipi: dict[str, Any],
    aitag: dict[str, Any],
    local_sd_prompt_editor: dict[str, Any] | None = None,
) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []

    for item in gwliang["flat_items"]:
        rows.append(
            {
                "source": "gwliang",
                "category": item["top_category"],
                "subcategory": item["subgroup"],
                "scope": item["scope"],
                "text_en": item["en"],
                "text_zh": item["zh"],
                "preview": "",
                "extra_id": "",
            }
        )

    for item in newzone["tags"]:
        rows.append(
            {
                "source": "newzone",
                "category": item.get("object", ""),
                "subcategory": item.get("attribute", ""),
                "scope": "default",
                "text_en": item.get("displayName", ""),
                "text_zh": item.get("langName", ""),
                "preview": item.get("preview", ""),
                "extra_id": "",
            }
        )

    for item in qpipi["items"]:
        rows.append(
            {
                "source": "qpipi",
                "category": item.get("tab", ""),
                "subcategory": item.get("section", ""),
                "scope": "r18" if item.get("tab") == "R18" else "default",
                "text_en": item.get("english", ""),
                "text_zh": item.get("chinese", ""),
                "preview": "",
                "extra_id": "",
            }
        )

    for item in aitag["items"]:
        rows.append(
            {
                "source": "aitag",
                "category": item.get("sub", ""),
                "subcategory": "",
                "scope": "default",
                "text_en": item.get("name", ""),
                "text_zh": item.get("desc", ""),
                "preview": item.get("image", ""),
                "extra_id": item.get("id", ""),
            }
        )

    if local_sd_prompt_editor is not None:
        for item in local_sd_prompt_editor["flat_items"]:
            rows.append(
                {
                    "source": "local_sd_prompt_editor",
                    "category": item.get("category", ""),
                    "subcategory": item.get("subcategory", ""),
                    "scope": "default",
                    "text_en": item.get("en", ""),
                    "text_zh": item.get("zh", ""),
                    "preview": "",
                    "extra_id": "",
                }
            )

        for index, text in enumerate(local_sd_prompt_editor["negative_defaults"]):
            rows.append(
                {
                    "source": "local_sd_prompt_editor",
                    "category": "反向提示词",
                    "subcategory": "常用反向提示词",
                    "scope": "negative_default",
                    "text_en": text,
                    "text_zh": text,
                    "preview": "",
                    "extra_id": f"negative_default_{index}",
                }
            )

    return rows


def save_merged(rows: list[dict[str, Any]]) -> None:
    save_json(OUT_DIR / "all_prompts_merged.json", rows)

    csv_path = OUT_DIR / "all_prompts_merged.csv"
    with csv_path.open("w", newline="", encoding="utf-8-sig") as f:
        writer = csv.DictWriter(
            f,
            fieldnames=["source", "category", "subcategory", "scope", "text_en", "text_zh", "preview", "extra_id"],
        )
        writer.writeheader()
        writer.writerows(rows)


@dataclass
class RunSummary:
    gwliang_count: int
    newzone_count: int
    qpipi_count: int
    aitag_count: int
    local_sd_prompt_editor_count: int
    merged_count: int


def main() -> None:
    ensure_dirs()
    session = get_session()

    gwliang = extract_gwliang(session)
    newzone = extract_newzone(session)
    qpipi = extract_qpipi(session)
    aitag = extract_aitag(session)
    local_sd_prompt_editor = extract_local_sd_prompt_editor()

    merged_rows = build_merged_rows(gwliang, newzone, qpipi, aitag, local_sd_prompt_editor)
    save_merged(merged_rows)

    summary = RunSummary(
        gwliang_count=gwliang["summary"]["flat_item_count"],
        newzone_count=newzone["item_count"],
        qpipi_count=qpipi["item_count"],
        aitag_count=aitag["item_count"],
        local_sd_prompt_editor_count=(
            0
            if local_sd_prompt_editor is None
            else local_sd_prompt_editor["item_count"] + local_sd_prompt_editor["negative_default_count"]
        ),
        merged_count=len(merged_rows),
    )

    manifest = {
        "created_at": utc_now_iso(),
        "output_dir": str(OUT_DIR),
        "sources": {
            "gwliang": {
                "url": GWLIANG_URL,
                "output_file": str(OUT_DIR / "gwliang.prompts.json"),
                "count": summary.gwliang_count,
            },
            "newzone": {
                "url": NEWZONE_URL,
                "output_file": str(OUT_DIR / "newzone.prompts.json"),
                "count": summary.newzone_count,
            },
            "qpipi": {
                "url": QPIPI_URL,
                "output_file": str(OUT_DIR / "qpipi.prompts.json"),
                "count": summary.qpipi_count,
            },
            "aitag": {
                "url": AITAG_URL,
                "output_file": str(OUT_DIR / "aitag.prompts.json"),
                "count": summary.aitag_count,
            },
        },
        "merged": {
            "json_file": str(OUT_DIR / "all_prompts_merged.json"),
            "csv_file": str(OUT_DIR / "all_prompts_merged.csv"),
            "count": summary.merged_count,
        },
    }
    if local_sd_prompt_editor is not None:
        manifest["sources"]["local_sd_prompt_editor"] = {
            "source_file": str(LOCAL_SD_PROMPT_EDITOR_PATH),
            "output_file": str(OUT_DIR / "local_sd_prompt_editor.prompts.json"),
            "count": local_sd_prompt_editor["item_count"],
            "negative_default_count": local_sd_prompt_editor["negative_default_count"],
        }
    save_json(OUT_DIR / "manifest.json", manifest)

    print(json.dumps(manifest, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
