from __future__ import annotations

import json
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent / "downloaded_prompt_data"
DATA_PATH = BASE_DIR / "all_prompts_merged.json"
OUTPUT_PATH = BASE_DIR / "all_prompts_prompt_editor.html"


HTML_TEMPLATE = """<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1" />
  <title>SD 提示词编辑器 · 合并词库版</title>
  <style>
    :root {
      --bg: #f6f7fb;
      --bg-soft: rgba(255, 255, 255, 0.8);
      --panel: rgba(255, 255, 255, 0.92);
      --panel-strong: #ffffff;
      --line: rgba(59, 77, 128, 0.14);
      --line-strong: rgba(59, 77, 128, 0.24);
      --text: #20263a;
      --text-soft: #65708d;
      --text-faint: #8b95af;
      --brand: #4f46e5;
      --brand-strong: #4338ca;
      --brand-soft: rgba(79, 70, 229, 0.1);
      --accent: #0f9d8a;
      --danger: #e11d48;
      --danger-soft: rgba(225, 29, 72, 0.09);
      --warn: #d97706;
      --warn-soft: rgba(217, 119, 6, 0.1);
      --shadow: 0 22px 54px rgba(43, 57, 99, 0.12);
      --radius-xl: 24px;
      --radius-lg: 18px;
      --radius-md: 14px;
      --header-h: 76px;
    }

    * {
      box-sizing: border-box;
    }

    html,
    body {
      margin: 0;
      min-height: 100%;
      color: var(--text);
      font-family: "Segoe UI Variable Display", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
      background:
        radial-gradient(circle at top left, rgba(99, 102, 241, 0.1), transparent 28%),
        radial-gradient(circle at top right, rgba(15, 157, 138, 0.09), transparent 24%),
        linear-gradient(180deg, #fafbff 0%, #f4f6fb 100%);
    }

    body {
      min-height: 100vh;
    }

    button,
    input,
    textarea {
      font: inherit;
    }

    button {
      cursor: pointer;
    }

    .app {
      min-height: 100vh;
      display: grid;
      grid-template-rows: var(--header-h) 1fr;
    }

    .topbar {
      position: sticky;
      top: 0;
      z-index: 30;
      display: flex;
      align-items: center;
      gap: 14px;
      padding: 14px 18px;
      backdrop-filter: blur(16px);
      background: rgba(250, 251, 255, 0.85);
      border-bottom: 1px solid var(--line);
      box-shadow: 0 10px 26px rgba(43, 57, 99, 0.06);
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 14px;
      min-width: 0;
    }

    .brand-badge {
      width: 42px;
      height: 42px;
      border-radius: 14px;
      display: grid;
      place-items: center;
      color: #fff;
      font-weight: 800;
      background: linear-gradient(135deg, #6366f1, #4338ca);
      box-shadow: 0 12px 24px rgba(79, 70, 229, 0.24);
      flex-shrink: 0;
    }

    .brand-title {
      font-size: 18px;
      font-weight: 800;
      letter-spacing: 0.02em;
      line-height: 1.1;
    }

    .brand-subtitle {
      margin-top: 4px;
      color: var(--text-soft);
      font-size: 12px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .search-wrap {
      flex: 1;
      min-width: 180px;
      position: relative;
    }

    .search-input {
      width: 100%;
      height: 46px;
      border: 1px solid rgba(67, 56, 202, 0.14);
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.84);
      color: var(--text);
      outline: none;
      padding: 0 46px 0 42px;
      transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
    }

    .search-input:focus {
      background: #fff;
      border-color: rgba(79, 70, 229, 0.38);
      box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.08);
    }

    .search-icon,
    .clear-search {
      position: absolute;
      top: 50%;
      transform: translateY(-50%);
      color: var(--text-faint);
    }

    .search-icon {
      left: 14px;
      pointer-events: none;
    }

    .clear-search {
      right: 10px;
      width: 28px;
      height: 28px;
      border-radius: 999px;
      border: 0;
      background: transparent;
    }

    .clear-search:hover {
      color: var(--brand-strong);
      background: rgba(79, 70, 229, 0.08);
    }

    .top-stats {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-wrap: wrap;
      justify-content: flex-end;
    }

    .stat-pill {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 10px 14px;
      border-radius: 14px;
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.78);
      color: var(--text-soft);
      font-size: 13px;
      white-space: nowrap;
    }

    .stat-pill strong {
      color: var(--text);
    }

    .layout {
      display: grid;
      grid-template-columns: 310px minmax(0, 1fr) 390px;
      gap: 18px;
      padding: 18px;
      min-height: 0;
    }

    .panel {
      min-height: 0;
      overflow: hidden;
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: var(--radius-xl);
      box-shadow: var(--shadow);
    }

    .panel-head {
      padding: 18px 18px 14px;
      border-bottom: 1px solid var(--line);
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(255, 255, 255, 0.8));
    }

    .panel-title {
      margin: 0;
      font-size: 15px;
      font-weight: 800;
    }

    .panel-subtitle {
      margin-top: 6px;
      font-size: 12px;
      color: var(--text-soft);
      line-height: 1.5;
    }

    .sidebar,
    .editor {
      display: flex;
      flex-direction: column;
    }

    .sidebar-body,
    .editor-body {
      padding: 12px;
      overflow: auto;
    }

    .source-tabs {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      margin-bottom: 12px;
    }

    .source-tab,
    .scope-tab {
      border: 0;
      border-radius: 12px;
      padding: 8px 11px;
      background: rgba(79, 70, 229, 0.08);
      color: var(--text-soft);
      font-size: 12px;
      font-weight: 700;
      transition: background 0.16s ease, color 0.16s ease, transform 0.16s ease, box-shadow 0.16s ease;
    }

    .source-tab:hover,
    .scope-tab:hover {
      transform: translateY(-1px);
    }

    .source-tab.active,
    .scope-tab.active {
      background: linear-gradient(135deg, #6366f1, #4338ca);
      color: #fff;
      box-shadow: 0 10px 20px rgba(79, 70, 229, 0.18);
    }

    .tree {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .tree-group {
      border: 1px solid rgba(67, 56, 202, 0.1);
      border-radius: 18px;
      background: rgba(255, 255, 255, 0.7);
      overflow: hidden;
    }

    .tree-category {
      width: 100%;
      border: 0;
      background: transparent;
      padding: 13px 14px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      text-align: left;
    }

    .tree-category:hover {
      background: rgba(79, 70, 229, 0.05);
    }

    .tree-category.active {
      color: #fff;
      background: linear-gradient(135deg, #6366f1, #4338ca);
    }

    .tree-main {
      min-width: 0;
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .caret {
      font-size: 12px;
      opacity: 0.8;
    }

    .category-name {
      font-size: 14px;
      font-weight: 700;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    .count-badge {
      flex-shrink: 0;
      padding: 4px 8px;
      border-radius: 999px;
      background: rgba(255, 255, 255, 0.12);
      font-size: 12px;
    }

    .tree-sublist {
      display: flex;
      flex-direction: column;
      gap: 4px;
      padding: 8px;
      border-top: 1px solid rgba(67, 56, 202, 0.08);
      background: rgba(248, 250, 255, 0.8);
    }

    .tree-subitem {
      width: 100%;
      border: 0;
      background: transparent;
      border-radius: 12px;
      padding: 10px 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      text-align: left;
      color: var(--text-soft);
    }

    .tree-subitem:hover {
      background: rgba(79, 70, 229, 0.05);
    }

    .tree-subitem.active {
      color: var(--brand-strong);
      background: rgba(79, 70, 229, 0.09);
    }

    .content {
      display: flex;
      flex-direction: column;
      min-height: 0;
    }

    .content-head {
      padding: 18px 18px 14px;
      border-bottom: 1px solid var(--line);
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.82));
    }

    .toolbar {
      margin-top: 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      flex-wrap: wrap;
    }

    .scope-tabs {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .content-meta {
      margin-top: 10px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      flex-wrap: wrap;
      color: var(--text-soft);
      font-size: 12px;
    }

    .content-body {
      min-height: 0;
      overflow: auto;
      padding: 16px;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .item {
      display: grid;
      grid-template-columns: 92px minmax(0, 1fr);
      gap: 14px;
      padding: 12px;
      border-radius: 20px;
      border: 1px solid var(--line);
      background: var(--panel-strong);
      box-shadow: 0 10px 22px rgba(43, 57, 99, 0.07);
      transition: border-color 0.16s ease, box-shadow 0.16s ease, transform 0.16s ease;
    }

    .item:hover {
      transform: translateY(-1px);
      border-color: rgba(79, 70, 229, 0.24);
      box-shadow: 0 16px 28px rgba(43, 57, 99, 0.1);
    }

    .item.active {
      border-color: rgba(79, 70, 229, 0.34);
      box-shadow: 0 16px 32px rgba(79, 70, 229, 0.14);
    }

    .thumb {
      height: 92px;
      border-radius: 16px;
      overflow: hidden;
      display: grid;
      place-items: center;
      color: var(--text-faint);
      background:
        radial-gradient(circle at top left, rgba(99, 102, 241, 0.14), transparent 40%),
        linear-gradient(135deg, #eef2ff, #f8fbff);
      font-size: 12px;
    }

    .thumb img {
      width: 100%;
      height: 100%;
      display: block;
      object-fit: cover;
    }

    .item-main {
      min-width: 0;
    }

    .item-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 7px;
    }

    .tag {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 9px;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 700;
      color: var(--text-soft);
      background: rgba(79, 70, 229, 0.08);
    }

    .tag.scope-negative {
      color: var(--danger);
      background: var(--danger-soft);
    }

    .tag.scope-r18 {
      color: var(--warn);
      background: var(--warn-soft);
    }

    .item-zh {
      margin-top: 10px;
      font-size: 17px;
      font-weight: 800;
      line-height: 1.25;
      word-break: break-word;
    }

    .item-en {
      margin-top: 6px;
      font-size: 13px;
      color: var(--text-soft);
      line-height: 1.55;
      word-break: break-word;
    }

    .item-actions {
      margin-top: 12px;
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .btn {
      border: 0;
      border-radius: 12px;
      padding: 8px 11px;
      font-size: 12px;
      font-weight: 700;
      transition: transform 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
    }

    .btn:hover {
      transform: translateY(-1px);
    }

    .btn-primary {
      color: #fff;
      background: linear-gradient(135deg, #6366f1, #4338ca);
      box-shadow: 0 10px 20px rgba(79, 70, 229, 0.18);
    }

    .btn-ghost {
      color: var(--text-soft);
      background: rgba(79, 70, 229, 0.08);
    }

    .btn-danger {
      color: var(--danger);
      background: var(--danger-soft);
    }

    .empty {
      padding: 34px 18px;
      border-radius: 18px;
      border: 1px dashed rgba(67, 56, 202, 0.18);
      text-align: center;
      color: var(--text-soft);
      background: rgba(255, 255, 255, 0.52);
    }

    .detail-card,
    .editor-section {
      border: 1px solid var(--line);
      border-radius: 20px;
      background: rgba(255, 255, 255, 0.82);
      box-shadow: 0 10px 26px rgba(43, 57, 99, 0.07);
    }

    .detail-media {
      height: 220px;
      border-bottom: 1px solid var(--line);
      overflow: hidden;
      display: grid;
      place-items: center;
      color: var(--text-faint);
      background:
        radial-gradient(circle at top left, rgba(99, 102, 241, 0.14), transparent 42%),
        linear-gradient(135deg, #eef2ff, #f8fbff);
    }

    .detail-media img {
      width: 100%;
      height: 100%;
      display: block;
      object-fit: cover;
    }

    .detail-body {
      padding: 16px;
    }

    .detail-title {
      font-size: 20px;
      font-weight: 800;
      line-height: 1.25;
      word-break: break-word;
    }

    .detail-en {
      margin-top: 8px;
      color: var(--text-soft);
      font-size: 13px;
      line-height: 1.55;
      word-break: break-word;
    }

    .detail-tags {
      margin-top: 12px;
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .editor-section {
      margin-top: 14px;
      padding: 14px;
    }

    .section-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      flex-wrap: wrap;
    }

    .section-title-sm {
      font-size: 14px;
      font-weight: 800;
    }

    .section-tools {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .chips {
      margin-top: 12px;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .chip {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 12px;
      border-radius: 16px;
      border: 1px solid rgba(67, 56, 202, 0.12);
      background: #fff;
    }

    .chip.neg {
      border-color: rgba(225, 29, 72, 0.12);
      background: rgba(255, 245, 247, 0.96);
    }

    .chip-main {
      flex: 1;
      min-width: 0;
    }

    .chip-zh {
      display: block;
      font-size: 14px;
      font-weight: 700;
      line-height: 1.3;
      word-break: break-word;
    }

    .chip-en {
      display: block;
      margin-top: 4px;
      color: var(--text-soft);
      font-size: 12px;
      word-break: break-word;
    }

    .weight-box {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 5px;
      border-radius: 999px;
      background: rgba(79, 70, 229, 0.08);
      flex-shrink: 0;
    }

    .weight-btn {
      width: 26px;
      height: 26px;
      border: 0;
      border-radius: 999px;
      background: rgba(255, 255, 255, 0.9);
      color: var(--brand-strong);
      font-weight: 800;
    }

    .weight-value {
      min-width: 34px;
      text-align: center;
      font-size: 12px;
      font-weight: 800;
    }

    .chip-remove {
      width: 30px;
      height: 30px;
      border: 0;
      border-radius: 999px;
      background: rgba(225, 29, 72, 0.08);
      color: var(--danger);
      flex-shrink: 0;
    }

    .output {
      width: 100%;
      min-height: 126px;
      margin-top: 12px;
      resize: vertical;
      border: 1px solid rgba(67, 56, 202, 0.14);
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.95);
      color: var(--text);
      outline: none;
      padding: 12px 14px;
      line-height: 1.6;
    }

    .output:focus {
      border-color: rgba(79, 70, 229, 0.38);
      box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.08);
    }

    .helper {
      margin-top: 10px;
      color: var(--text-soft);
      font-size: 12px;
      line-height: 1.5;
    }

    .toast {
      position: fixed;
      left: 50%;
      bottom: 24px;
      transform: translate(-50%, 14px);
      padding: 12px 16px;
      border-radius: 14px;
      color: #fff;
      background: rgba(32, 38, 58, 0.92);
      box-shadow: 0 14px 26px rgba(0, 0, 0, 0.16);
      opacity: 0;
      pointer-events: none;
      transition: opacity 0.18s ease, transform 0.18s ease;
      z-index: 60;
    }

    .toast.show {
      opacity: 1;
      transform: translate(-50%, 0);
    }

    @media (max-width: 1500px) {
      .layout {
        grid-template-columns: 290px minmax(0, 1fr) 360px;
      }
    }

    @media (max-width: 1220px) {
      .layout {
        grid-template-columns: 280px minmax(0, 1fr);
      }

      .editor {
        grid-column: 1 / -1;
      }
    }

    @media (max-width: 900px) {
      .app {
        grid-template-rows: auto 1fr;
      }

      .topbar {
        flex-direction: column;
        align-items: stretch;
      }

      .layout {
        grid-template-columns: 1fr;
      }

      .item {
        grid-template-columns: 1fr;
      }

      .thumb {
        height: 180px;
      }
    }
  </style>
</head>
<body>
  <div class="app">
    <header class="topbar">
      <div class="brand">
        <div class="brand-badge">SD</div>
        <div>
          <div class="brand-title">SD 提示词编辑器 · 合并词库版</div>
          <div class="brand-subtitle">按你本地那种编辑器类型重做：左侧分类树，中间词条列表，右侧正向 / 反向提示词编辑区</div>
        </div>
      </div>

      <div class="search-wrap">
        <span class="search-icon">⌕</span>
        <input id="searchInput" class="search-input" type="text" placeholder="搜索中文、英文、来源、分类..." />
        <button id="clearSearchBtn" class="clear-search" title="清空搜索">×</button>
      </div>

      <div class="top-stats">
        <div class="stat-pill">来源 <strong id="sourceCount">0</strong></div>
        <div class="stat-pill">分类 <strong id="categoryCount">0</strong></div>
        <div class="stat-pill">词条 <strong id="rowCount">0</strong></div>
        <div class="stat-pill">当前结果 <strong id="visibleCount">0</strong></div>
      </div>
    </header>

    <main class="layout">
      <aside class="panel sidebar">
        <div class="panel-head">
          <h2 class="panel-title">来源与分类</h2>
          <div class="panel-subtitle">先选来源，再从分类树里浏览子类。没有子类的词库会自动归到“未分组”。</div>
        </div>
        <div class="sidebar-body">
          <div id="sourceTabs" class="source-tabs"></div>
          <div id="categoryTree" class="tree"></div>
        </div>
      </aside>

      <section class="panel content">
        <div class="content-head">
          <h2 id="contentTitle" class="panel-title">全部词条</h2>
          <div id="contentSubtitle" class="panel-subtitle">正在加载合并词库...</div>

          <div class="toolbar">
            <div id="scopeTabs" class="scope-tabs"></div>
          </div>

          <div class="content-meta">
            <span id="resultSummary">-</span>
            <span>点词条可看详情，也可以直接加入右侧编辑器。</span>
          </div>
        </div>

        <div class="content-body">
          <div id="itemList" class="list"></div>
          <div id="emptyState" class="empty" hidden>没有匹配结果，试试切换来源、分类或搜索关键词。</div>
        </div>
      </section>

      <aside class="panel editor">
        <div class="panel-head">
          <h2 class="panel-title">详情与编辑器</h2>
          <div class="panel-subtitle">保留你本地 HTML 那种“浏览词库 + 即时编辑 prompt”的使用方式，并加入多来源合并筛选。</div>
        </div>
        <div class="editor-body">
          <div class="detail-card">
            <div id="detailMedia" class="detail-media">选择一个词条查看详情</div>
            <div class="detail-body">
              <div id="detailTitle" class="detail-title">未选择词条</div>
              <div id="detailEn" class="detail-en">点击中间词条后，这里会显示中英文、来源、分类、作用域和快捷操作。</div>
              <div id="detailTags" class="detail-tags"></div>
              <div class="item-actions">
                <button id="detailAddPositive" class="btn btn-primary">加入正向</button>
                <button id="detailAddNegative" class="btn btn-danger">加入反向</button>
                <button id="detailCopy" class="btn btn-ghost">复制英文</button>
              </div>
            </div>
          </div>

          <section class="editor-section">
            <div class="section-row">
              <div class="section-title-sm">正向提示词</div>
              <div class="section-tools">
                <button id="copyPositiveBtn" class="btn btn-ghost">复制</button>
                <button id="clearPositiveBtn" class="btn btn-ghost">清空</button>
              </div>
            </div>
            <div id="positiveChips" class="chips"></div>
            <textarea id="positiveOutput" class="output" spellcheck="false" placeholder="这里会自动拼接正向 prompt。"></textarea>
          </section>

          <section class="editor-section">
            <div class="section-row">
              <div class="section-title-sm">反向提示词</div>
              <div class="section-tools">
                <button id="importNegativeDefaultsBtn" class="btn btn-ghost">导入默认反向</button>
                <button id="copyNegativeBtn" class="btn btn-ghost">复制</button>
                <button id="clearNegativeBtn" class="btn btn-ghost">清空</button>
              </div>
            </div>
            <div id="negativeChips" class="chips"></div>
            <textarea id="negativeOutput" class="output" spellcheck="false" placeholder="这里会自动拼接 negative prompt。"></textarea>
            <div class="helper">“导入默认反向”会从合并词库里读取 `scope = negative_default` 的 30 条默认反向词。</div>
          </section>
        </div>
      </aside>
    </main>
  </div>

  <div id="toast" class="toast"></div>
  <script id="merged-data" type="application/json">__DATA_JSON__</script>
  <script>
    const rawRows = JSON.parse(document.getElementById("merged-data").textContent);
    const rows = rawRows.map((row, index) => ({
      id: String(index + 1),
      source: row.source || "unknown",
      category: row.category || "未分类",
      subcategory: row.subcategory || "未分组",
      scope: row.scope || "default",
      textEn: String(row.text_en || "").trim(),
      textZh: String(row.text_zh || "").trim(),
      preview: row.preview || "",
      extraId: row.extra_id ?? "",
    }));

    const STORAGE_KEY = "merged-prompt-editor-state-v1";
    const scopeOrder = ["all", "default", "normal", "r18", "negative_default"];
    const scopeLabelMap = {
      all: "全部",
      default: "默认",
      normal: "普通",
      r18: "R18",
      negative_default: "默认反向",
    };

    const sourceEntries = Array.from(
      rows.reduce((map, row) => {
        const current = map.get(row.source) || 0;
        map.set(row.source, current + 1);
        return map;
      }, new Map()).entries()
    )
      .sort((a, b) => {
        if (b[1] !== a[1]) return b[1] - a[1];
        return a[0].localeCompare(b[0], "zh-CN");
      })
      .map(([name, count]) => ({ name, count }));

    const defaultNegativeRows = rows.filter((row) => row.scope === "negative_default");

    const state = {
      source: "all",
      category: "all",
      subcategory: "all",
      expandedCategory: null,
      scope: "all",
      search: "",
      activeId: null,
      positive: [],
      negative: [],
      positiveText: "",
      negativeText: "",
    };

    function loadState() {
      try {
        const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
        if (Array.isArray(saved.positive)) state.positive = saved.positive;
        if (Array.isArray(saved.negative)) state.negative = saved.negative;
        if (typeof saved.positiveText === "string") state.positiveText = saved.positiveText;
        if (typeof saved.negativeText === "string") state.negativeText = saved.negativeText;
      } catch (error) {
        // ignore bad cache
      }
    }

    function persistState() {
      const payload = {
        positive: state.positive,
        negative: state.negative,
        positiveText: state.positiveText,
        negativeText: state.negativeText,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
    }

    loadState();

    const els = {
      sourceCount: document.getElementById("sourceCount"),
      categoryCount: document.getElementById("categoryCount"),
      rowCount: document.getElementById("rowCount"),
      visibleCount: document.getElementById("visibleCount"),
      searchInput: document.getElementById("searchInput"),
      clearSearchBtn: document.getElementById("clearSearchBtn"),
      sourceTabs: document.getElementById("sourceTabs"),
      categoryTree: document.getElementById("categoryTree"),
      scopeTabs: document.getElementById("scopeTabs"),
      contentTitle: document.getElementById("contentTitle"),
      contentSubtitle: document.getElementById("contentSubtitle"),
      resultSummary: document.getElementById("resultSummary"),
      itemList: document.getElementById("itemList"),
      emptyState: document.getElementById("emptyState"),
      detailMedia: document.getElementById("detailMedia"),
      detailTitle: document.getElementById("detailTitle"),
      detailEn: document.getElementById("detailEn"),
      detailTags: document.getElementById("detailTags"),
      positiveChips: document.getElementById("positiveChips"),
      negativeChips: document.getElementById("negativeChips"),
      positiveOutput: document.getElementById("positiveOutput"),
      negativeOutput: document.getElementById("negativeOutput"),
      toast: document.getElementById("toast"),
    };

    els.positiveOutput.value = state.positiveText;
    els.negativeOutput.value = state.negativeText;

    function escapeHtml(value) {
      return String(value)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
    }

    function normalizeWeight(value) {
      return Math.max(0.1, Math.min(2, Math.round(value * 10) / 10));
    }

    function withWeight(text, weight) {
      const normalized = normalizeWeight(weight);
      if (normalized === 1) return text;
      return `(${text}:${normalized.toFixed(1)})`;
    }

    function formatSelectionPrompt(list) {
      return list
        .map((item) => withWeight(item.textEn, item.weight))
        .join(", ");
    }

    function mergePromptText(generated, manual) {
      const a = generated.trim();
      const b = manual.trim();
      if (a && b) return `${a}, ${b}`;
      return a || b;
    }

    function showToast(message) {
      els.toast.textContent = message;
      els.toast.classList.add("show");
      window.clearTimeout(showToast.timer);
      showToast.timer = window.setTimeout(() => {
        els.toast.classList.remove("show");
      }, 1800);
    }

    async function copyText(text, message) {
      if (!text.trim()) {
        showToast("当前内容还是空的。");
        return;
      }
      try {
        await navigator.clipboard.writeText(text);
        showToast(message);
      } catch (error) {
        showToast("复制失败，请手动复制。");
      }
    }

    function getSourceFilteredRows() {
      return rows.filter((row) => state.source === "all" || row.source === state.source);
    }

    function buildCategoryTree() {
      const sourceRows = getSourceFilteredRows();
      const categoryMap = new Map();

      for (const row of sourceRows) {
        if (!categoryMap.has(row.category)) {
          categoryMap.set(row.category, {
            count: 0,
            subMap: new Map(),
          });
        }
        const categoryEntry = categoryMap.get(row.category);
        categoryEntry.count += 1;
        categoryEntry.subMap.set(row.subcategory, (categoryEntry.subMap.get(row.subcategory) || 0) + 1);
      }

      return Array.from(categoryMap.entries())
        .sort((a, b) => {
          if (b[1].count !== a[1].count) return b[1].count - a[1].count;
          return a[0].localeCompare(b[0], "zh-CN");
        })
        .map(([name, value]) => ({
          name,
          count: value.count,
          subs: Array.from(value.subMap.entries())
            .sort((a, b) => {
              if (b[1] !== a[1]) return b[1] - a[1];
              return a[0].localeCompare(b[0], "zh-CN");
            })
            .map(([subName, subCount]) => ({ name: subName, count: subCount })),
        }));
    }

    function getFilteredRows() {
      const keyword = state.search.trim().toLowerCase();
      return rows.filter((row) => {
        if (state.source !== "all" && row.source !== state.source) return false;
        if (state.category !== "all" && row.category !== state.category) return false;
        if (state.subcategory !== "all" && row.subcategory !== state.subcategory) return false;
        if (state.scope !== "all" && row.scope !== state.scope) return false;
        if (!keyword) return true;
        return [
          row.source,
          row.category,
          row.subcategory,
          row.scope,
          row.textZh,
          row.textEn,
        ].some((value) => value.toLowerCase().includes(keyword));
      });
    }

    function getActiveRow() {
      return rows.find((row) => row.id === state.activeId) || null;
    }

    function upsertSelection(kind, row) {
      if (!row || !row.textEn) {
        showToast("这个词条没有可用英文提示词。");
        return;
      }
      const target = state[kind];
      if (target.some((item) => item.id === row.id)) {
        showToast("这个词条已经在编辑器里了。");
        return;
      }
      target.push({
        id: row.id,
        textEn: row.textEn,
        textZh: row.textZh,
        source: row.source,
        category: row.category,
        subcategory: row.subcategory,
        scope: row.scope,
        weight: 1,
      });
      renderSelections();
      showToast(kind === "positive" ? "已加入正向提示词" : "已加入反向提示词");
    }

    function removeSelection(kind, id) {
      state[kind] = state[kind].filter((item) => item.id !== id);
      renderSelections();
    }

    function updateSelectionWeight(kind, id, delta) {
      state[kind] = state[kind].map((item) => {
        if (item.id !== id) return item;
        return {
          ...item,
          weight: normalizeWeight(item.weight + delta),
        };
      });
      renderSelections();
    }

    function importNegativeDefaults() {
      let added = 0;
      for (const row of defaultNegativeRows) {
        if (!state.negative.some((item) => item.id === row.id) && row.textEn) {
          state.negative.push({
            id: row.id,
            textEn: row.textEn,
            textZh: row.textZh,
            source: row.source,
            category: row.category,
            subcategory: row.subcategory,
            scope: row.scope,
            weight: 1,
          });
          added += 1;
        }
      }
      renderSelections();
      showToast(added ? `已导入 ${added} 条默认反向词` : "默认反向词已经全部在右侧了。");
    }

    function renderSourceTabs() {
      const tabs = [{ name: "all", label: "全部来源", count: rows.length }].concat(
        sourceEntries.map((entry) => ({
          name: entry.name,
          label: entry.name,
          count: entry.count,
        }))
      );

      els.sourceTabs.innerHTML = tabs.map((tab) => `
        <button class="source-tab ${tab.name === state.source ? "active" : ""}" data-source="${escapeHtml(tab.name)}">
          ${escapeHtml(tab.label)} · ${tab.count}
        </button>
      `).join("");

      els.sourceTabs.querySelectorAll(".source-tab").forEach((button) => {
        button.addEventListener("click", () => {
          state.source = button.dataset.source;
          state.category = "all";
          state.subcategory = "all";
          state.expandedCategory = null;
          render();
        });
      });
    }

    function renderCategoryTree() {
      const tree = buildCategoryTree();
      els.categoryCount.textContent = String(tree.length);

      if (state.category !== "all" && !tree.some((entry) => entry.name === state.category)) {
        state.category = "all";
        state.subcategory = "all";
      }

      els.categoryTree.innerHTML = `
        <div class="tree-group">
          <button class="tree-category ${state.category === "all" ? "active" : ""}" data-category="all">
            <span class="tree-main">
              <span class="caret">•</span>
              <span class="category-name">全部分类</span>
            </span>
            <span class="count-badge">${getSourceFilteredRows().length}</span>
          </button>
        </div>
      ` + tree.map((entry) => {
        const expanded = state.expandedCategory === entry.name || state.category === entry.name;
        const isActiveCategory = state.category === entry.name && state.subcategory === "all";
        return `
          <div class="tree-group">
            <button class="tree-category ${isActiveCategory ? "active" : ""}" data-category="${escapeHtml(entry.name)}">
              <span class="tree-main">
                <span class="caret">${expanded ? "▾" : "▸"}</span>
                <span class="category-name">${escapeHtml(entry.name)}</span>
              </span>
              <span class="count-badge">${entry.count}</span>
            </button>
            ${expanded ? `
              <div class="tree-sublist">
                ${entry.subs.map((sub) => `
                  <button class="tree-subitem ${state.category === entry.name && state.subcategory === sub.name ? "active" : ""}" data-category="${escapeHtml(entry.name)}" data-subcategory="${escapeHtml(sub.name)}">
                    <span>${escapeHtml(sub.name)}</span>
                    <span class="count-badge">${sub.count}</span>
                  </button>
                `).join("")}
              </div>
            ` : ""}
          </div>
        `;
      }).join("");

      els.categoryTree.querySelectorAll(".tree-category").forEach((button) => {
        button.addEventListener("click", () => {
          const category = button.dataset.category;
          if (category === "all") {
            state.category = "all";
            state.subcategory = "all";
            state.expandedCategory = null;
            render();
            return;
          }
          state.category = category;
          state.subcategory = "all";
          state.expandedCategory = state.expandedCategory === category ? null : category;
          render();
        });
      });

      els.categoryTree.querySelectorAll(".tree-subitem").forEach((button) => {
        button.addEventListener("click", () => {
          state.category = button.dataset.category;
          state.subcategory = button.dataset.subcategory;
          state.expandedCategory = button.dataset.category;
          render();
        });
      });
    }

    function renderScopeTabs(filteredRows) {
      const countMap = new Map();
      countMap.set("all", filteredRows.length);
      for (const row of filteredRows) {
        countMap.set(row.scope, (countMap.get(row.scope) || 0) + 1);
      }

      const tabs = scopeOrder
        .filter((scope) => scope === "all" || countMap.has(scope))
        .map((scope) => ({
          scope,
          label: scopeLabelMap[scope] || scope,
          count: countMap.get(scope) || 0,
        }));

      if (!tabs.some((tab) => tab.scope === state.scope)) {
        state.scope = "all";
      }

      els.scopeTabs.innerHTML = tabs.map((tab) => `
        <button class="scope-tab ${tab.scope === state.scope ? "active" : ""}" data-scope="${escapeHtml(tab.scope)}">
          ${escapeHtml(tab.label)} · ${tab.count}
        </button>
      `).join("");

      els.scopeTabs.querySelectorAll(".scope-tab").forEach((button) => {
        button.addEventListener("click", () => {
          state.scope = button.dataset.scope;
          render();
        });
      });
    }

    function scopeClass(scope) {
      if (scope === "negative_default") return "scope-negative";
      if (scope === "r18") return "scope-r18";
      return "";
    }

    function renderList(filtered) {
      if (!filtered.length) {
        els.itemList.innerHTML = "";
        els.emptyState.hidden = false;
        return;
      }

      els.emptyState.hidden = true;
      els.itemList.innerHTML = filtered.map((row) => `
        <article class="item ${row.id === state.activeId ? "active" : ""}" data-id="${escapeHtml(row.id)}">
          <div class="thumb">
            ${row.preview
              ? `<img src="${escapeHtml(row.preview)}" alt="${escapeHtml(row.textZh || row.textEn)}" loading="lazy" />`
              : `<span>No Preview</span>`}
          </div>
          <div class="item-main">
            <div class="item-tags">
              <span class="tag">${escapeHtml(row.source)}</span>
              <span class="tag">${escapeHtml(row.category)}</span>
              <span class="tag">${escapeHtml(row.subcategory)}</span>
              <span class="tag ${scopeClass(row.scope)}">${escapeHtml(scopeLabelMap[row.scope] || row.scope)}</span>
            </div>
            <div class="item-zh">${escapeHtml(row.textZh || "未命名词条")}</div>
            <div class="item-en">${escapeHtml(row.textEn)}</div>
            <div class="item-actions">
              <button class="btn btn-primary" data-action="positive">加入正向</button>
              <button class="btn btn-danger" data-action="negative">加入反向</button>
              <button class="btn btn-ghost" data-action="copy">复制英文</button>
            </div>
          </div>
        </article>
      `).join("");

      els.itemList.querySelectorAll(".item").forEach((node) => {
        const row = filtered.find((entry) => entry.id === node.dataset.id);
        node.addEventListener("click", () => {
          state.activeId = row.id;
          renderList(filtered);
          renderDetail();
        });

        node.querySelectorAll("[data-action]").forEach((button) => {
          button.addEventListener("click", (event) => {
            event.stopPropagation();
            if (button.dataset.action === "positive") upsertSelection("positive", row);
            if (button.dataset.action === "negative") upsertSelection("negative", row);
            if (button.dataset.action === "copy") copyText(row.textEn, "已复制英文提示词");
          });
        });
      });
    }

    function renderDetail() {
      const row = getActiveRow();
      if (!row) {
        els.detailMedia.innerHTML = "选择一个词条查看详情";
        els.detailTitle.textContent = "未选择词条";
        els.detailEn.textContent = "点击中间词条后，这里会显示中英文、来源、分类、作用域和快捷操作。";
        els.detailTags.innerHTML = "";
        return;
      }

      els.detailMedia.innerHTML = row.preview
        ? `<img src="${escapeHtml(row.preview)}" alt="${escapeHtml(row.textZh || row.textEn)}" loading="lazy" />`
        : "No Preview";
      els.detailTitle.textContent = row.textZh || "未命名词条";
      els.detailEn.textContent = row.textEn;
      els.detailTags.innerHTML = `
        <span class="tag">${escapeHtml(row.source)}</span>
        <span class="tag">${escapeHtml(row.category)}</span>
        <span class="tag">${escapeHtml(row.subcategory)}</span>
        <span class="tag ${scopeClass(row.scope)}">${escapeHtml(scopeLabelMap[row.scope] || row.scope)}</span>
        ${row.extraId !== "" ? `<span class="tag">ID ${escapeHtml(row.extraId)}</span>` : ""}
      `;
    }

    function renderChipList(container, list, kind) {
      if (!list.length) {
        container.innerHTML = `<div class="empty">还没有添加词条，可以从中间列表直接加入。</div>`;
        return;
      }

      container.innerHTML = list.map((item) => `
        <div class="chip ${kind === "negative" ? "neg" : ""}" data-id="${escapeHtml(item.id)}">
          <div class="chip-main">
            <span class="chip-zh">${escapeHtml(item.textZh || "未命名词条")}</span>
            <span class="chip-en">${escapeHtml(item.textEn)}</span>
          </div>
          <div class="weight-box">
            <button class="weight-btn" data-weight="-0.1">-</button>
            <span class="weight-value">${normalizeWeight(item.weight).toFixed(1)}</span>
            <button class="weight-btn" data-weight="0.1">+</button>
          </div>
          <button class="chip-remove" data-remove="1">×</button>
        </div>
      `).join("");

      container.querySelectorAll(".chip").forEach((chip) => {
        const id = chip.dataset.id;
        chip.querySelectorAll("[data-weight]").forEach((button) => {
          button.addEventListener("click", () => {
            updateSelectionWeight(kind, id, Number(button.dataset.weight));
          });
        });
        chip.querySelector("[data-remove]").addEventListener("click", () => {
          removeSelection(kind, id);
        });
      });
    }

    function renderSelections() {
      renderChipList(els.positiveChips, state.positive, "positive");
      renderChipList(els.negativeChips, state.negative, "negative");

      const positiveGenerated = formatSelectionPrompt(state.positive);
      const negativeGenerated = formatSelectionPrompt(state.negative);
      els.positiveOutput.value = mergePromptText(positiveGenerated, state.positiveText);
      els.negativeOutput.value = mergePromptText(negativeGenerated, state.negativeText);
      persistState();
    }

    function renderSummary(filtered, sourceRows) {
      const tree = buildCategoryTree();
      const previewCount = filtered.filter((row) => row.preview).length;
      els.sourceCount.textContent = String(sourceEntries.length);
      els.rowCount.textContent = String(rows.length);
      els.visibleCount.textContent = String(filtered.length);

      if (state.category === "all") {
        els.contentTitle.textContent = state.source === "all" ? "全部词条" : `${state.source} · 全部分类`;
      } else if (state.subcategory === "all") {
        els.contentTitle.textContent = `${state.category}`;
      } else {
        els.contentTitle.textContent = `${state.category} / ${state.subcategory}`;
      }

      if (state.search.trim()) {
        els.contentSubtitle.textContent = `搜索中：${state.search}`;
      } else if (state.source === "all") {
        els.contentSubtitle.textContent = "当前浏览全部来源的合并词库。";
      } else {
        els.contentSubtitle.textContent = `当前来源：${state.source}`;
      }

      els.resultSummary.textContent = `当前结果 ${filtered.length} 条，带预览图 ${previewCount} 条。来源筛选后共有 ${sourceRows.length} 条、${tree.length} 个分类。`;
    }

    function render() {
      const sourceRows = getSourceFilteredRows();
      renderSourceTabs();
      renderCategoryTree();
      renderScopeTabs(sourceRows);

      const filtered = getFilteredRows();
      if (!filtered.some((row) => row.id === state.activeId)) {
        state.activeId = filtered[0]?.id || null;
      }

      renderSummary(filtered, sourceRows);
      renderList(filtered);
      renderDetail();
      renderSelections();
    }

    els.searchInput.addEventListener("input", () => {
      state.search = els.searchInput.value;
      render();
    });

    els.clearSearchBtn.addEventListener("click", () => {
      els.searchInput.value = "";
      state.search = "";
      render();
    });

    document.getElementById("detailAddPositive").addEventListener("click", () => {
      upsertSelection("positive", getActiveRow());
    });

    document.getElementById("detailAddNegative").addEventListener("click", () => {
      upsertSelection("negative", getActiveRow());
    });

    document.getElementById("detailCopy").addEventListener("click", () => {
      const row = getActiveRow();
      if (row) copyText(row.textEn, "已复制英文提示词");
    });

    document.getElementById("copyPositiveBtn").addEventListener("click", () => {
      copyText(els.positiveOutput.value, "已复制正向提示词");
    });

    document.getElementById("copyNegativeBtn").addEventListener("click", () => {
      copyText(els.negativeOutput.value, "已复制反向提示词");
    });

    document.getElementById("clearPositiveBtn").addEventListener("click", () => {
      state.positive = [];
      renderSelections();
      showToast("已清空正向提示词。");
    });

    document.getElementById("clearNegativeBtn").addEventListener("click", () => {
      state.negative = [];
      renderSelections();
      showToast("已清空反向提示词。");
    });

    document.getElementById("importNegativeDefaultsBtn").addEventListener("click", () => {
      importNegativeDefaults();
    });

    els.positiveOutput.addEventListener("input", () => {
      const generated = formatSelectionPrompt(state.positive);
      const value = els.positiveOutput.value.trim();
      if (!generated) {
        state.positiveText = value;
      } else if (!value.startsWith(generated)) {
        state.positiveText = value;
      } else {
        state.positiveText = value.slice(generated.length).replace(/^,\s*/, "");
      }
      persistState();
    });

    els.negativeOutput.addEventListener("input", () => {
      const generated = formatSelectionPrompt(state.negative);
      const value = els.negativeOutput.value.trim();
      if (!generated) {
        state.negativeText = value;
      } else if (!value.startsWith(generated)) {
        state.negativeText = value;
      } else {
        state.negativeText = value.slice(generated.length).replace(/^,\s*/, "");
      }
      persistState();
    });

    render();
  </script>
</body>
</html>
"""


def build_html(rows: list[dict]) -> str:
    data_json = json.dumps(rows, ensure_ascii=False).replace("</script>", "<\\/script>")
    return HTML_TEMPLATE.replace("__DATA_JSON__", data_json)


def main() -> None:
    rows = json.loads(DATA_PATH.read_text(encoding="utf-8"))
    html = build_html(rows)
    OUTPUT_PATH.write_text(html, encoding="utf-8")
    print(OUTPUT_PATH)


if __name__ == "__main__":
    main()
