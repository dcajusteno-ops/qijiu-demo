from __future__ import annotations

import json
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent / "downloaded_prompt_data"
DATA_PATH = BASE_DIR / "aitag.prompts.json"
OUTPUT_PATH = BASE_DIR / "aitag_prompt_viewer.html"


HTML_TEMPLATE = """<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>AITag 提示词查看器</title>
  <style>
    :root {
      --bg: #f6f2ff;
      --bg-soft: rgba(255, 255, 255, 0.82);
      --panel: rgba(255, 255, 255, 0.9);
      --panel-strong: rgba(255, 255, 255, 0.96);
      --line: rgba(109, 40, 217, 0.12);
      --line-strong: rgba(109, 40, 217, 0.22);
      --text: #2b2142;
      --text-soft: #695a85;
      --text-faint: #9b8db1;
      --violet: #7c3aed;
      --violet-strong: #6d28d9;
      --violet-soft: rgba(124, 58, 237, 0.1);
      --rose: #e11d48;
      --rose-soft: rgba(225, 29, 72, 0.1);
      --shadow: 0 24px 60px rgba(109, 40, 217, 0.12);
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
        radial-gradient(circle at top left, rgba(168, 85, 247, 0.16), transparent 28%),
        radial-gradient(circle at top right, rgba(59, 130, 246, 0.14), transparent 22%),
        linear-gradient(180deg, #fbf9ff 0%, #f5f2ff 100%);
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
      gap: 16px;
      padding: 14px 20px;
      backdrop-filter: blur(16px);
      background: rgba(250, 247, 255, 0.82);
      border-bottom: 1px solid var(--line);
      box-shadow: 0 10px 32px rgba(109, 40, 217, 0.06);
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 14px;
      min-width: 0;
    }

    .brand-badge {
      width: 44px;
      height: 44px;
      border-radius: 15px;
      display: grid;
      place-items: center;
      font-weight: 800;
      color: #fff;
      background: linear-gradient(135deg, #8b5cf6, #6d28d9);
      box-shadow: 0 16px 30px rgba(124, 58, 237, 0.24);
      flex-shrink: 0;
    }

    .brand-title {
      font-size: 18px;
      font-weight: 700;
      line-height: 1.1;
      letter-spacing: 0.02em;
    }

    .brand-subtitle {
      margin-top: 4px;
      font-size: 12px;
      color: var(--text-soft);
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
      padding: 0 44px 0 42px;
      border: 1px solid rgba(109, 40, 217, 0.14);
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.84);
      color: var(--text);
      outline: none;
      transition: border-color 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    }

    .search-input:focus {
      background: #fff;
      border-color: rgba(124, 58, 237, 0.42);
      box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
    }

    .search-icon,
    .search-clear {
      position: absolute;
      top: 50%;
      transform: translateY(-50%);
      color: var(--text-faint);
    }

    .search-icon {
      left: 14px;
      pointer-events: none;
    }

    .search-clear {
      right: 10px;
      width: 28px;
      height: 28px;
      border: 0;
      border-radius: 999px;
      background: transparent;
    }

    .search-clear:hover {
      background: rgba(124, 58, 237, 0.08);
      color: var(--violet-strong);
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
      border: 1px solid var(--line);
      border-radius: 15px;
      background: rgba(255, 255, 255, 0.74);
      font-size: 13px;
      color: var(--text-soft);
      white-space: nowrap;
    }

    .stat-pill strong {
      color: var(--text);
      font-weight: 700;
    }

    .layout {
      display: grid;
      grid-template-columns: 280px minmax(0, 1fr) 370px;
      gap: 18px;
      padding: 18px;
      min-height: 0;
    }

    .panel {
      min-height: 0;
      overflow: hidden;
      border: 1px solid var(--line);
      border-radius: var(--radius-xl);
      background: var(--panel);
      box-shadow: var(--shadow);
    }

    .panel-head {
      padding: 18px 18px 14px;
      border-bottom: 1px solid var(--line);
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.76));
    }

    .panel-title {
      margin: 0;
      font-size: 15px;
      font-weight: 700;
    }

    .panel-subtitle {
      margin-top: 6px;
      font-size: 12px;
      color: var(--text-soft);
    }

    .sidebar,
    .builder {
      display: flex;
      flex-direction: column;
    }

    .sidebar-body,
    .builder-body {
      padding: 12px;
      overflow: auto;
    }

    .sub-list {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .sub-item {
      width: 100%;
      border: 0;
      border-radius: 16px;
      padding: 13px 14px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      text-align: left;
      color: var(--text);
      background: transparent;
      transition: transform 0.16s ease, background 0.16s ease, box-shadow 0.16s ease;
    }

    .sub-item:hover {
      background: rgba(124, 58, 237, 0.06);
      transform: translateX(2px);
    }

    .sub-item.active {
      color: #fff;
      background: linear-gradient(135deg, #8b5cf6, #6d28d9);
      box-shadow: 0 14px 24px rgba(124, 58, 237, 0.22);
    }

    .sub-name {
      font-size: 14px;
      font-weight: 600;
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .sub-count {
      flex-shrink: 0;
      font-size: 12px;
      padding: 4px 8px;
      border-radius: 999px;
      color: inherit;
      background: rgba(255, 255, 255, 0.12);
    }

    .content {
      display: flex;
      flex-direction: column;
      min-height: 0;
    }

    .content-head {
      padding: 18px 18px 14px;
      border-bottom: 1px solid var(--line);
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.78));
    }

    .content-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
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

    .segmented {
      display: inline-flex;
      gap: 8px;
      padding: 5px;
      border-radius: 16px;
      background: rgba(124, 58, 237, 0.08);
    }

    .seg-btn {
      border: 0;
      border-radius: 12px;
      padding: 9px 12px;
      color: var(--text-soft);
      background: transparent;
      transition: background 0.16s ease, color 0.16s ease, box-shadow 0.16s ease;
    }

    .seg-btn.active {
      color: #fff;
      background: linear-gradient(135deg, #8b5cf6, #6d28d9);
      box-shadow: 0 10px 18px rgba(124, 58, 237, 0.18);
    }

    .content-body {
      min-height: 0;
      overflow: auto;
      padding: 16px;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(208px, 1fr));
      gap: 14px;
    }

    .card {
      border: 1px solid var(--line);
      border-radius: 20px;
      overflow: hidden;
      background: var(--panel-strong);
      transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
      box-shadow: 0 10px 28px rgba(124, 58, 237, 0.08);
    }

    .card:hover {
      transform: translateY(-2px);
      border-color: rgba(124, 58, 237, 0.28);
      box-shadow: 0 18px 34px rgba(124, 58, 237, 0.12);
    }

    .card.active {
      border-color: rgba(124, 58, 237, 0.42);
      box-shadow: 0 18px 36px rgba(124, 58, 237, 0.18);
    }

    .card-media {
      position: relative;
      height: 168px;
      overflow: hidden;
      background:
        radial-gradient(circle at top left, rgba(139, 92, 246, 0.22), transparent 46%),
        linear-gradient(135deg, #f2ecff, #faf7ff);
      display: grid;
      place-items: center;
      color: var(--text-faint);
      font-size: 13px;
    }

    .card-media img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    .card-body {
      padding: 14px;
    }

    .card-sub {
      display: inline-flex;
      align-items: center;
      padding: 5px 9px;
      border-radius: 999px;
      background: rgba(124, 58, 237, 0.08);
      color: var(--violet-strong);
      font-size: 12px;
      font-weight: 700;
    }

    .card-cn {
      margin-top: 12px;
      font-size: 16px;
      font-weight: 700;
      line-height: 1.3;
      min-height: 42px;
    }

    .card-en {
      margin-top: 7px;
      font-size: 13px;
      color: var(--text-soft);
      line-height: 1.45;
      word-break: break-word;
      min-height: 38px;
    }

    .card-actions {
      margin-top: 14px;
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .action-btn {
      border: 0;
      border-radius: 12px;
      padding: 8px 11px;
      font-size: 12px;
      font-weight: 700;
      transition: transform 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
    }

    .action-btn:hover {
      transform: translateY(-1px);
    }

    .action-btn.pos {
      color: #fff;
      background: linear-gradient(135deg, #8b5cf6, #6d28d9);
      box-shadow: 0 10px 18px rgba(124, 58, 237, 0.18);
    }

    .action-btn.neg {
      color: var(--rose);
      background: rgba(225, 29, 72, 0.08);
    }

    .action-btn.copy {
      color: var(--text-soft);
      background: rgba(109, 40, 217, 0.08);
    }

    .empty {
      padding: 34px 18px;
      text-align: center;
      border: 1px dashed rgba(109, 40, 217, 0.18);
      border-radius: 18px;
      color: var(--text-soft);
      background: rgba(255, 255, 255, 0.45);
    }

    .detail-card {
      overflow: hidden;
      border: 1px solid var(--line);
      border-radius: 20px;
      background: var(--panel-strong);
      box-shadow: 0 10px 30px rgba(124, 58, 237, 0.08);
    }

    .detail-media {
      height: 230px;
      display: grid;
      place-items: center;
      color: var(--text-faint);
      background:
        radial-gradient(circle at top left, rgba(139, 92, 246, 0.22), transparent 44%),
        linear-gradient(135deg, #f4efff, #fbf9ff);
      overflow: hidden;
    }

    .detail-media img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    .detail-body {
      padding: 16px;
    }

    .detail-title {
      font-size: 20px;
      font-weight: 800;
      line-height: 1.25;
    }

    .detail-en {
      margin-top: 8px;
      font-size: 13px;
      color: var(--text-soft);
      line-height: 1.55;
      word-break: break-word;
    }

    .detail-tags {
      margin-top: 12px;
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .mini-tag {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 6px 10px;
      border-radius: 999px;
      font-size: 12px;
      color: var(--text-soft);
      background: rgba(109, 40, 217, 0.08);
    }

    .builder-section {
      margin-top: 14px;
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 20px;
      background: rgba(255, 255, 255, 0.72);
    }

    .builder-head-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      flex-wrap: wrap;
    }

    .builder-title {
      font-size: 14px;
      font-weight: 700;
    }

    .builder-tools {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .small-btn {
      border: 0;
      border-radius: 11px;
      padding: 8px 10px;
      font-size: 12px;
      color: var(--text-soft);
      background: rgba(109, 40, 217, 0.08);
    }

    .small-btn:hover {
      color: var(--violet-strong);
      background: rgba(109, 40, 217, 0.14);
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
      border: 1px solid rgba(109, 40, 217, 0.12);
      background: #fff;
    }

    .chip.neg {
      border-color: rgba(225, 29, 72, 0.14);
      background: rgba(255, 245, 247, 0.92);
    }

    .chip-label {
      min-width: 0;
      flex: 1;
    }

    .chip-cn {
      display: block;
      font-size: 14px;
      font-weight: 700;
      line-height: 1.3;
    }

    .chip-en {
      display: block;
      margin-top: 4px;
      font-size: 12px;
      color: var(--text-soft);
      word-break: break-word;
    }

    .weight-box {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 5px;
      border-radius: 999px;
      background: rgba(109, 40, 217, 0.08);
      flex-shrink: 0;
    }

    .weight-btn {
      width: 26px;
      height: 26px;
      border: 0;
      border-radius: 999px;
      color: var(--violet-strong);
      background: rgba(255, 255, 255, 0.82);
    }

    .weight-value {
      min-width: 34px;
      text-align: center;
      font-size: 12px;
      font-weight: 700;
      color: var(--text);
    }

    .chip-remove {
      width: 30px;
      height: 30px;
      border: 0;
      border-radius: 999px;
      color: var(--rose);
      background: rgba(225, 29, 72, 0.08);
      flex-shrink: 0;
    }

    .prompt-box {
      margin-top: 12px;
    }

    .prompt-output {
      width: 100%;
      min-height: 122px;
      resize: vertical;
      border: 1px solid rgba(109, 40, 217, 0.14);
      border-radius: 16px;
      padding: 12px 14px;
      background: rgba(255, 255, 255, 0.94);
      color: var(--text);
      line-height: 1.6;
      outline: none;
    }

    .prompt-output:focus {
      border-color: rgba(124, 58, 237, 0.42);
      box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
    }

    .toast {
      position: fixed;
      left: 50%;
      bottom: 24px;
      transform: translate(-50%, 14px);
      padding: 12px 16px;
      border-radius: 14px;
      color: #fff;
      background: rgba(43, 33, 66, 0.92);
      box-shadow: 0 14px 28px rgba(0, 0, 0, 0.18);
      opacity: 0;
      pointer-events: none;
      transition: opacity 0.18s ease, transform 0.18s ease;
      z-index: 60;
    }

    .toast.show {
      opacity: 1;
      transform: translate(-50%, 0);
    }

    @media (max-width: 1480px) {
      .layout {
        grid-template-columns: 250px minmax(0, 1fr) 330px;
      }
    }

    @media (max-width: 1180px) {
      .layout {
        grid-template-columns: 240px minmax(0, 1fr);
      }

      .builder {
        grid-column: 1 / -1;
      }
    }

    @media (max-width: 860px) {
      .app {
        grid-template-rows: auto 1fr;
      }

      .topbar {
        align-items: stretch;
        flex-direction: column;
      }

      .layout {
        grid-template-columns: 1fr;
      }

      .grid {
        grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      }
    }
  </style>
</head>
<body>
  <div class="app">
    <header class="topbar">
      <div class="brand">
        <div class="brand-badge">AI</div>
        <div>
          <div class="brand-title">AITag 提示词查看器</div>
          <div class="brand-subtitle">参考你本地那份编辑器的三栏布局，做成更适合浏览与临时组词的独立页面</div>
        </div>
      </div>

      <div class="search-wrap">
        <span class="search-icon">⌕</span>
        <input id="searchInput" class="search-input" type="text" placeholder="搜索中文、英文或分类..." />
        <button id="clearSearchBtn" class="search-clear" title="清空搜索">×</button>
      </div>

      <div class="top-stats">
        <div class="stat-pill">分类 <strong id="subCount">0</strong></div>
        <div class="stat-pill">总词条 <strong id="itemCount">0</strong></div>
        <div class="stat-pill">当前结果 <strong id="visibleCount">0</strong></div>
      </div>
    </header>

    <main class="layout">
      <aside class="panel sidebar">
        <div class="panel-head">
          <h2 class="panel-title">分类导航</h2>
          <div class="panel-subtitle">左边切分类，中间看卡片，右边临时整理正向和反向提示词</div>
        </div>
        <div class="sidebar-body">
          <div id="subList" class="sub-list"></div>
        </div>
      </aside>

      <section class="panel content">
        <div class="content-head">
          <div class="content-row">
            <div>
              <h2 id="contentTitle" class="panel-title">全部词条</h2>
              <div id="contentSubtitle" class="panel-subtitle">正在加载数据...</div>
            </div>

            <div class="segmented">
              <button class="seg-btn active" data-view="all">全部</button>
              <button class="seg-btn" data-view="with-image">仅看带图</button>
              <button class="seg-btn" data-view="without-image">仅看无图</button>
            </div>
          </div>

          <div class="content-meta">
            <span id="resultSummary">-</span>
            <span>点卡片可查看详情，也可以直接加入右侧的 prompt 组装区</span>
          </div>
        </div>

        <div class="content-body">
          <div id="cardGrid" class="grid"></div>
          <div id="emptyState" class="empty" hidden>没有匹配结果，试试换个关键词或切换分类。</div>
        </div>
      </section>

      <aside class="panel builder">
        <div class="panel-head">
          <h2 class="panel-title">详情与组词</h2>
          <div class="panel-subtitle">保留本地编辑器那种“浏览 + 即时拼接”的感觉，但更偏向查看、筛选和复制。</div>
        </div>

        <div class="builder-body">
          <div class="detail-card">
            <div id="detailMedia" class="detail-media">选择一张卡片查看详情</div>
            <div class="detail-body">
              <div id="detailTitle" class="detail-title">未选择词条</div>
              <div id="detailEn" class="detail-en">点击中间的卡片后，这里会显示中文、英文、分类和快捷操作。</div>
              <div id="detailTags" class="detail-tags"></div>
              <div class="card-actions">
                <button id="detailAddPos" class="action-btn pos">加入正向</button>
                <button id="detailAddNeg" class="action-btn neg">加入反向</button>
                <button id="detailCopy" class="action-btn copy">复制英文</button>
              </div>
            </div>
          </div>

          <section class="builder-section">
            <div class="builder-head-row">
              <div class="builder-title">正向提示词</div>
              <div class="builder-tools">
                <button id="copyPositiveBtn" class="small-btn">复制</button>
                <button id="clearPositiveBtn" class="small-btn">清空</button>
              </div>
            </div>
            <div id="positiveChips" class="chips"></div>
            <div class="prompt-box">
              <textarea id="positiveOutput" class="prompt-output" spellcheck="false" placeholder="加入正向词条后，这里会自动生成英文 prompt。"></textarea>
            </div>
          </section>

          <section class="builder-section">
            <div class="builder-head-row">
              <div class="builder-title">反向提示词</div>
              <div class="builder-tools">
                <button id="copyNegativeBtn" class="small-btn">复制</button>
                <button id="clearNegativeBtn" class="small-btn">清空</button>
              </div>
            </div>
            <div id="negativeChips" class="chips"></div>
            <div class="prompt-box">
              <textarea id="negativeOutput" class="prompt-output" spellcheck="false" placeholder="加入反向词条后，这里会自动生成 negative prompt。"></textarea>
            </div>
          </section>
        </div>
      </aside>
    </main>
  </div>

  <div id="toast" class="toast"></div>
  <script id="aitag-data" type="application/json">__DATA_JSON__</script>
  <script>
    const rawData = JSON.parse(document.getElementById("aitag-data").textContent);
    const items = (rawData.items || []).map((item, index) => ({
      id: String(item.id ?? index),
      sub: item.sub || "未分类",
      cn: item.desc || "",
      en: item.name || "",
      image: item.image || "",
    }));

    const subCounts = rawData.counts_by_sub || {};
    const subEntries = Object.entries(subCounts)
      .sort((a, b) => {
        if (b[1] !== a[1]) return b[1] - a[1];
        return a[0].localeCompare(b[0], "zh-CN");
      })
      .map(([name, count]) => ({ name, count }));

    const state = {
      search: "",
      viewMode: "all",
      activeSub: "全部",
      activeId: null,
      positive: [],
      negative: [],
    };

    const els = {
      subCount: document.getElementById("subCount"),
      itemCount: document.getElementById("itemCount"),
      visibleCount: document.getElementById("visibleCount"),
      subList: document.getElementById("subList"),
      cardGrid: document.getElementById("cardGrid"),
      emptyState: document.getElementById("emptyState"),
      contentTitle: document.getElementById("contentTitle"),
      contentSubtitle: document.getElementById("contentSubtitle"),
      resultSummary: document.getElementById("resultSummary"),
      searchInput: document.getElementById("searchInput"),
      clearSearchBtn: document.getElementById("clearSearchBtn"),
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

    function formatPrompt(list) {
      return list
        .map((item) => {
          const weight = normalizeWeight(item.weight);
          if (weight === 1) return item.en;
          return `(${item.en}:${weight.toFixed(1)})`;
        })
        .join(", ");
    }

    function getFilteredItems() {
      const keyword = state.search.trim().toLowerCase();
      return items.filter((item) => {
        if (state.activeSub !== "全部" && item.sub !== state.activeSub) return false;
        if (state.viewMode === "with-image" && !item.image) return false;
        if (state.viewMode === "without-image" && item.image) return false;
        if (!keyword) return true;
        return [
          item.cn,
          item.en,
          item.sub,
        ].some((value) => value.toLowerCase().includes(keyword));
      });
    }

    function getActiveItem() {
      return items.find((item) => item.id === state.activeId) || null;
    }

    function showToast(message) {
      els.toast.textContent = message;
      els.toast.classList.add("show");
      window.clearTimeout(showToast.timer);
      showToast.timer = window.setTimeout(() => {
        els.toast.classList.remove("show");
      }, 1800);
    }

    async function copyText(text, successMessage) {
      if (!text) return;
      try {
        await navigator.clipboard.writeText(text);
        showToast(successMessage);
      } catch (error) {
        showToast("复制失败，请手动复制。");
      }
    }

    function addSelection(kind, item) {
      if (!item) return;
      const list = state[kind];
      if (list.some((entry) => entry.id === item.id)) {
        showToast("这个词条已经在右侧了。");
        return;
      }
      list.push({ ...item, weight: 1 });
      renderSelections();
      showToast(kind === "positive" ? "已加入正向提示词" : "已加入反向提示词");
    }

    function removeSelection(kind, id) {
      state[kind] = state[kind].filter((item) => item.id !== id);
      renderSelections();
    }

    function updateWeight(kind, id, delta) {
      state[kind] = state[kind].map((item) => {
        if (item.id !== id) return item;
        return { ...item, weight: normalizeWeight(item.weight + delta) };
      });
      renderSelections();
    }

    function renderSidebar() {
      const sidebarItems = [{ name: "全部", count: items.length }, ...subEntries];
      els.subList.innerHTML = sidebarItems.map((item) => `
        <button class="sub-item ${item.name === state.activeSub ? "active" : ""}" data-sub="${escapeHtml(item.name)}">
          <span class="sub-name">${escapeHtml(item.name)}</span>
          <span class="sub-count">${item.count}</span>
        </button>
      `).join("");

      els.subList.querySelectorAll(".sub-item").forEach((button) => {
        button.addEventListener("click", () => {
          state.activeSub = button.dataset.sub;
          render();
        });
      });
    }

    function renderSummary(filtered) {
      const imageCount = filtered.filter((item) => item.image).length;
      els.subCount.textContent = String(subEntries.length);
      els.itemCount.textContent = String(items.length);
      els.visibleCount.textContent = String(filtered.length);
      els.contentTitle.textContent = state.activeSub === "全部" ? "全部词条" : state.activeSub;

      if (state.search.trim()) {
        els.contentSubtitle.textContent = `搜索中：${state.search}`;
      } else if (state.activeSub === "全部") {
        els.contentSubtitle.textContent = "浏览全部 AITag 词条。";
      } else {
        els.contentSubtitle.textContent = `当前分类：${state.activeSub}`;
      }

      els.resultSummary.textContent = `共 ${filtered.length} 条结果，其中 ${imageCount} 条带预览图。`;
    }

    function renderCards(filtered) {
      if (!filtered.length) {
        els.cardGrid.innerHTML = "";
        els.emptyState.hidden = false;
        return;
      }

      els.emptyState.hidden = true;
      els.cardGrid.innerHTML = filtered.map((item) => `
        <article class="card ${item.id === state.activeId ? "active" : ""}" data-id="${escapeHtml(item.id)}">
          <div class="card-media">
            ${item.image
              ? `<img src="${escapeHtml(item.image)}" alt="${escapeHtml(item.cn || item.en)}" loading="lazy" />`
              : `<span>No Preview</span>`}
          </div>
          <div class="card-body">
            <span class="card-sub">${escapeHtml(item.sub)}</span>
            <div class="card-cn">${escapeHtml(item.cn || "未命名")}</div>
            <div class="card-en">${escapeHtml(item.en)}</div>
            <div class="card-actions">
              <button class="action-btn pos" data-action="pos">正向</button>
              <button class="action-btn neg" data-action="neg">反向</button>
              <button class="action-btn copy" data-action="copy">复制</button>
            </div>
          </div>
        </article>
      `).join("");

      els.cardGrid.querySelectorAll(".card").forEach((card) => {
        const item = filtered.find((entry) => entry.id === card.dataset.id);
        card.addEventListener("click", () => {
          state.activeId = item.id;
          renderCards(filtered);
          renderDetail();
        });

        card.querySelectorAll("[data-action]").forEach((button) => {
          button.addEventListener("click", (event) => {
            event.stopPropagation();
            if (button.dataset.action === "pos") addSelection("positive", item);
            if (button.dataset.action === "neg") addSelection("negative", item);
            if (button.dataset.action === "copy") copyText(item.en, "已复制英文提示词");
          });
        });
      });
    }

    function renderDetail() {
      const item = getActiveItem();
      if (!item) {
        els.detailMedia.className = "detail-media";
        els.detailMedia.innerHTML = "选择一张卡片查看详情";
        els.detailTitle.textContent = "未选择词条";
        els.detailEn.textContent = "点击中间的卡片后，这里会显示中文、英文、分类和快捷操作。";
        els.detailTags.innerHTML = "";
        return;
      }

      els.detailMedia.className = "detail-media";
      els.detailMedia.innerHTML = item.image
        ? `<img src="${escapeHtml(item.image)}" alt="${escapeHtml(item.cn || item.en)}" loading="lazy" />`
        : "No Preview";
      els.detailTitle.textContent = item.cn || "未命名";
      els.detailEn.textContent = item.en;
      els.detailTags.innerHTML = `
        <span class="mini-tag">分类 ${escapeHtml(item.sub)}</span>
        <span class="mini-tag">编号 ${escapeHtml(item.id)}</span>
        <span class="mini-tag">${item.image ? "带预览图" : "无预览图"}</span>
      `;
    }

    function renderChipList(container, list, kind) {
      if (!list.length) {
        container.innerHTML = `<div class="empty">还没有添加内容，可以从中间卡片里直接加入。</div>`;
        return;
      }

      container.innerHTML = list.map((item) => `
        <div class="chip ${kind === "negative" ? "neg" : ""}" data-id="${escapeHtml(item.id)}">
          <div class="chip-label">
            <span class="chip-cn">${escapeHtml(item.cn || "未命名")}</span>
            <span class="chip-en">${escapeHtml(item.en)}</span>
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
            updateWeight(kind, id, Number(button.dataset.weight));
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
      els.positiveOutput.value = formatPrompt(state.positive);
      els.negativeOutput.value = formatPrompt(state.negative);
    }

    function renderSegmentButtons() {
      document.querySelectorAll(".seg-btn").forEach((button) => {
        button.classList.toggle("active", button.dataset.view === state.viewMode);
      });
    }

    function render() {
      const filtered = getFilteredItems();
      if (!filtered.some((item) => item.id === state.activeId)) {
        state.activeId = filtered[0]?.id || null;
      }

      renderSidebar();
      renderSummary(filtered);
      renderSegmentButtons();
      renderCards(filtered);
      renderDetail();
      renderSelections();
    }

    document.querySelectorAll(".seg-btn").forEach((button) => {
      button.addEventListener("click", () => {
        state.viewMode = button.dataset.view;
        render();
      });
    });

    els.searchInput.addEventListener("input", () => {
      state.search = els.searchInput.value;
      render();
    });

    els.clearSearchBtn.addEventListener("click", () => {
      els.searchInput.value = "";
      state.search = "";
      render();
    });

    document.getElementById("detailAddPos").addEventListener("click", () => {
      addSelection("positive", getActiveItem());
    });

    document.getElementById("detailAddNeg").addEventListener("click", () => {
      addSelection("negative", getActiveItem());
    });

    document.getElementById("detailCopy").addEventListener("click", () => {
      const item = getActiveItem();
      if (item) copyText(item.en, "已复制英文提示词");
    });

    document.getElementById("copyPositiveBtn").addEventListener("click", () => {
      const value = els.positiveOutput.value.trim();
      if (!value) {
        showToast("正向提示词还是空的。");
        return;
      }
      copyText(value, "已复制正向提示词");
    });

    document.getElementById("copyNegativeBtn").addEventListener("click", () => {
      const value = els.negativeOutput.value.trim();
      if (!value) {
        showToast("反向提示词还是空的。");
        return;
      }
      copyText(value, "已复制反向提示词");
    });

    document.getElementById("clearPositiveBtn").addEventListener("click", () => {
      state.positive = [];
      renderSelections();
    });

    document.getElementById("clearNegativeBtn").addEventListener("click", () => {
      state.negative = [];
      renderSelections();
    });

    render();
  </script>
</body>
</html>
"""


def build_html(data: dict) -> str:
    data_json = json.dumps(data, ensure_ascii=False).replace("</script>", "<\\/script>")
    return HTML_TEMPLATE.replace("__DATA_JSON__", data_json)


def main() -> None:
    data = json.loads(DATA_PATH.read_text(encoding="utf-8"))
    html = build_html(data)
    OUTPUT_PATH.write_text(html, encoding="utf-8")
    print(OUTPUT_PATH)


if __name__ == "__main__":
    main()
