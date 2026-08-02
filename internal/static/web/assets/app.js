/* =========================================================================
   BLX — aplicação front-end
   Vanilla JS SPA: escolha a ferramenta → responda as perguntas → comandos.
   ========================================================================= */

"use strict";

const app = document.getElementById("app");

const state = {
  tools: [],
  categories: [],
  catName: {},
  loaded: false,
  view: "home", // home | wizard | result
  activeCat: "all",
  query: "",
  tool: null,
  answers: {},
  result: null,
  generating: false,
};

/* ------------------------------------------------------------------ API */

async function api(path, opts) {
  const res = await fetch(path, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw Object.assign(new Error(data.error || "Falha na requisição"), { data, status: res.status });
  }
  return data;
}

/* ---------------------------------------------------------------- Toast */

function showToast(message, type = "ok") {
  let wrap = document.querySelector(".toast-wrap");
  if (!wrap) {
    wrap = document.createElement("div");
    wrap.className = "toast-wrap";
    document.body.appendChild(wrap);
  }
  const t = document.createElement("div");
  t.className = "toast " + (type === "error" ? "toast--error" : type === "info" ? "toast--info" : "");
  t.textContent = message;
  wrap.appendChild(t);
  setTimeout(() => {
    t.style.opacity = "0";
    t.style.transition = "opacity 0.3s";
    setTimeout(() => t.remove(), 320);
  }, 2600);
}

/* -------------------------------------------------------------- Helpers */

function esc(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

const RISK_LABEL = { low: "Baixo", medium: "Médio", high: "Alto" };

function riskBadge(risk) {
  return `<span class="badge badge--risk-${risk}">Risco ${RISK_LABEL[risk] || risk}</span>`;
}

function highlight(code, lang) {
  let h = esc(code);
  if (lang === "resource") {
    h = h.replace(/^(use|set|run|exploit|handler|get|search|check)\b/gm, '<span class="tok-k">$1</span>');
    h = h.replace(/^([A-Z][A-Z0-9_-]+)/gm, '<span class="tok-f">$1</span>');
    return h;
  }
  h = h.replace(/('(?:[^'\\]|\\.)*')/g, '<span class="tok-s">$1</span>');
  h = h.replace(/(^|\s)(--?[a-zA-Z0-9][a-zA-Z0-9_.-]*)/g, '$1<span class="tok-f">$2</span>');
  h = h.replace(/(#[^\n]*)/g, '<span class="tok-c">$1</span>');
  return h;
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    document.execCommand("copy");
    ta.remove();
  }
}

function slugify(s) {
  return String(s)
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "") || "script";
}

function downloadScript(result) {
  const lines = [
    "#!/usr/bin/env bash",
    "set -euo pipefail",
    `# Gerado por BLX — ${result.toolName}`,
    "# Apenas execute em ambientes com autorização formal.",
    "",
  ];
  result.commands.forEach((c, i) => {
    lines.push(`# --- ${c.title} ---`);
    if (c.language === "resource") {
      const file = slugify(c.title) + ".rc";
      lines.push(`cat > ${file} <<'RCEOF'`);
      lines.push(c.code);
      lines.push("RCEOF");
      lines.push(`echo ">> salvo ${file} — rode: msfconsole -q -r ${file}"`);
    } else {
      lines.push(c.code);
    }
    lines.push("");
  });
  const blob = new Blob([lines.join("\n")], { type: "application/x-sh" });
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = `blx-${slugify(result.toolName)}.sh`;
  a.click();
  URL.revokeObjectURL(a.href);
  showToast("Script baixado com sucesso");
}

/* ------------------------------------------------------------- Rendering */

function render() {
  if (state.view === "home") renderHome();
  else if (state.view === "wizard") renderWizard();
  else renderResult();
  window.scrollTo({ top: 0, behavior: "instant" });
}

function renderHome() {
  const cat = state.activeCat;
  const q = state.query.trim().toLowerCase();
  const tools = state.tools.filter((t) => {
    const okCat = cat === "all" || t.category === cat;
    const hay = (t.name + " " + t.description + " " + t.tags.join(" ")).toLowerCase();
    return okCat && (!q || hay.includes(q));
  });

  const chips =
    `<button class="chip ${cat === "all" ? "active" : ""}" data-cat="all">Todas</button>` +
    state.categories
      .map((c) => `<button class="chip ${cat === c.id ? "active" : ""}" data-cat="${c.id}">${c.name}</button>`)
      .join("");

  const grid = tools.length
    ? tools
        .map((t, i) => {
          const catName = state.catName[t.category] || t.category;
          return `<div class="card" data-tool="${t.id}" style="animation-delay:${Math.min(i * 40, 400)}ms">
            <div class="card__top">
              <div class="card__icon">${t.icon}</div>
              ${riskBadge(t.risk)}
            </div>
            <div class="card__name">${esc(t.name)}</div>
            <div class="card__desc">${esc(t.description)}</div>
            <div class="card__tags">${t.tags.map((tag) => `<span class="tag">${esc(tag)}</span>`).join("")}</div>
            <div class="card__foot">
              <span class="badge badge--cat">${esc(catName)}</span>
              <span class="card__cta">configurar →</span>
            </div>
          </div>`;
        })
        .join("")
    : `<div class="empty">// nenhuma ferramenta encontrada</div>`;

  app.innerHTML = `
    ${topbar()}
    <main class="container">
      <section class="hero">
        <div class="hero__eyebrow"><span class="status__dot" style="animation:none"></span> terminal de pentest</div>
        <h1 class="hero__title">Escolha a ferramenta.<br /><span class="grad">Responda. Receba o comando.</span></h1>
        <p class="hero__sub">
          BLX monta os comandos de pentest para você: selecione a ferramenta,
          informe como e com o que será usada e copie o código pronto para executar.
        </p>
        <div class="hero__stats">
          <span class="stat"><b>${state.tools.length}</b> ferramentas</span>
          <span class="stat"><b>${state.categories.length}</b> categorias</span>
          <span class="stat"><b>1</b> comando pronto</span>
        </div>
      </section>

      <section class="controls">
        <div class="search">
          <span class="search__icon">⌕</span>
          <input class="search__input" id="search" type="text" placeholder="buscar por nome, tag ou descrição…" value="${esc(state.query)}" />
        </div>
        <div class="chips">${chips}</div>
      </section>

      <section class="grid">${grid}</section>
    </main>
    ${footer()}
  `;

  const search = document.getElementById("search");
  search.addEventListener("input", () => {
    const pos = search.selectionStart;
    state.query = search.value;
    renderHome();
    const s = document.getElementById("search");
    s.focus();
    s.setSelectionRange(pos, pos);
  });

  app.querySelectorAll("[data-cat]").forEach((el) =>
    el.addEventListener("click", () => {
      state.activeCat = el.dataset.cat;
      renderHome();
    })
  );

  app.querySelectorAll("[data-tool]").forEach((el) =>
    el.addEventListener("click", () => openWizard(state.tools.find((t) => t.id === el.dataset.tool)))
  );
}

function openWizard(tool) {
  if (!tool) return;
  state.tool = tool;
  state.result = null;
  state.answers = {};
  tool.questions.forEach((qq) => {
    if (qq.default !== undefined && qq.default !== "") state.answers[qq.id] = qq.default;
  });
  state.view = "wizard";
  render();
}

function wizardProgress() {
  const questions = state.tool.questions;
  if (!questions.length) return 100;
  const filled = questions.filter((q) => state.answers[q.id]).length;
  return Math.round((filled / questions.length) * 100);
}

function renderQuestion(q) {
  const val = state.answers[q.id] ?? "";
  const err = q.error;
  let control = "";

  if (q.type === "select") {
    control = `<div class="pillgroup" data-q="${q.id}">${q.options
      .map(
        (o) =>
          `<button type="button" class="pill ${o.value === val ? "active" : ""}" data-val="${esc(o.value)}">${esc(o.label)}</button>`
      )
      .join("")}</div>`;
  } else if (q.type === "multi") {
    const selected = val ? val.split(",") : [];
    control = `<div class="checklist" data-q="${q.id}">${q.options
      .map((o) => {
        const on = selected.includes(o.value);
        return `<label class="check ${on ? "checked" : ""}">
          <span class="check__box">✓</span>
          <span data-val="${esc(o.value)}">${esc(o.label)}</span>
        </label>`;
      })
      .join("")}</div>`;
  } else if (q.type === "boolean") {
    const on = val === "true" || val === "1" || val === "on";
    control = `<label class="toggle">
      <div class="toggle__row">
        <input type="checkbox" data-q="${q.id}" ${on ? "checked" : ""} />
        <span class="toggle__track"></span>
        <span class="toggle__state">${on ? "ativado" : "desativado"}</span>
      </div>
    </label>`;
  } else {
    const type = q.type === "number" ? "number" : "text";
    control = `<input class="input" data-q="${q.id}" type="${type}" placeholder="${esc(q.placeholder || "")}" value="${esc(val)}" />`;
  }

  return `
    <div class="question ${err ? "has-error" : ""}" id="q-${q.id}">
      <div class="question__label">${esc(q.label)} ${q.required ? '<span class="question__req">*</span>' : ""}</div>
      ${q.help ? `<div class="question__help">${esc(q.help)}</div>` : ""}
      ${control}
      <div class="question__error">${err ? err : ""}</div>
    </div>
  `;
}

function renderWizard() {
  const t = state.tool;
  const q = t.questions;

  const qs = q.map((qq) => renderQuestion(qq)).join("");

  app.innerHTML = `
    ${topbar()}
    <main class="container">
      <div class="wizard__head">
        <div class="wizard__icon">${t.icon}</div>
        <div>
          <h1 class="wizard__title">${esc(t.name)}</h1>
          <div class="wizard__meta">
            <span class="badge badge--cat">${esc(state.catName[t.category] || t.category)}</span>
            ${riskBadge(t.risk)}
            <span>instalar: <code>${esc(t.install)}</code></span>
            ${t.docs ? `<a href="${esc(t.docs)}" target="_blank" rel="noopener">documentação ↗</a>` : ""}
          </div>
        </div>
      </div>

      <div class="stepper">
        <div class="step done"><span class="step__num">1</span><span class="step__label">Ferramenta</span></div>
        <div class="step active"><span class="step__num">2</span><span class="step__label">Configuração</span></div>
        <div class="step"><span class="step__num">3</span><span class="step__label">Comandos</span></div>
      </div>
      <div class="progress"><div class="progress__fill" style="width:${wizardProgress()}%"></div></div>

      <form id="wizard-form" novalidate>
        ${qs}
        <div class="btn-row">
          <button type="button" class="btn btn--ghost" data-action="back-home">← escolher outra</button>
          <button type="submit" class="btn btn--primary" ${state.generating ? "disabled" : ""}>
            ${state.generating ? "gerando…" : "⚡ gerar comandos"}
          </button>
        </div>
      </form>
    </main>
    ${footer()}
  `;

  attachQuestionEvents();
  renderProgress();

  app.querySelector('[data-action="back-home"]').addEventListener("click", () => {
    state.view = "home";
    render();
  });

  const form = document.getElementById("wizard-form");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (state.generating) return;
    state.generating = true;
    renderWizard();
    await generate();
    state.generating = false;
    if (state.view === "result") {
      renderResult();
    } else {
      renderWizard();
      scrollToFirstError();
    }
  });
}

function attachQuestionEvents() {
  // text / number
  app.querySelectorAll("input.input").forEach((input) => {
    input.addEventListener("input", () => {
      state.answers[input.dataset.q] = input.value;
      renderProgress();
    });
  });

  // select pills
  app.querySelectorAll(".pillgroup").forEach((group) => {
    group.querySelectorAll(".pill").forEach((pill) => {
      pill.addEventListener("click", () => {
        const qid = group.dataset.q;
        state.answers[qid] = pill.dataset.val;
        group.querySelectorAll(".pill").forEach((p) => p.classList.toggle("active", p === pill));
        clearError(qid);
        renderProgress();
      });
    });
  });

  // multi checklist
  app.querySelectorAll(".checklist").forEach((list) => {
    list.querySelectorAll(".check").forEach((check) => {
      check.addEventListener("click", () => {
        const qid = list.dataset.q;
        const val = check.querySelector("[data-val]").dataset.val;
        let sel = state.answers[qid] ? state.answers[qid].split(",") : [];
        sel = sel.includes(val) ? sel.filter((v) => v !== val) : sel.concat(val);
        state.answers[qid] = sel.join(",");
        check.classList.toggle("checked", sel.includes(val));
        clearError(qid);
      });
    });
  });

  // boolean toggle
  app.querySelectorAll(".toggle input").forEach((input) => {
    input.addEventListener("change", () => {
      state.answers[input.dataset.q] = input.checked ? "true" : "false";
      const row = input.closest(".toggle__row");
      row.querySelector(".toggle__state").textContent = input.checked ? "ativado" : "desativado";
      clearError(input.dataset.q);
    });
  });
}

function clearError(qid) {
  const q = state.tool.questions.find((x) => x.id === qid);
  if (q) delete q.error;
  const el = document.getElementById("q-" + qid);
  if (el) {
    el.classList.remove("has-error");
    const err = el.querySelector(".question__error");
    if (err) err.textContent = "";
  }
}

function renderProgress() {
  const fill = app.querySelector(".progress__fill");
  if (fill) fill.style.width = wizardProgress() + "%";
}

async function generate() {
  // client-side required check (errors render on the re-rendered wizard)
  let hasClientError = false;
  state.tool.questions.forEach((q) => {
    const val = state.answers[q.id] ?? "";
    delete q.error;
    if (q.required && !String(val).trim()) {
      q.error = "campo obrigatório — preencha este valor";
      hasClientError = true;
    }
  });
  if (hasClientError) {
    showToast("Preencha os campos obrigatórios", "error");
    return;
  }

  try {
    const result = await api("/api/generate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ toolId: state.tool.id, answers: state.answers }),
    });
    state.result = result;
    state.view = "result";
  } catch (err) {
    if (err.status === 422) {
      const q = state.tool.questions.find(
        (x) =>
          x.id === err.data.question ||
          x.label === err.data.question ||
          (err.data.question && x.label.startsWith(err.data.question))
      );
      if (q) q.error = err.data.reason;
      showToast("Verifique os dados informados", "error");
    } else {
      showToast(err.message || "Erro ao gerar comandos", "error");
    }
  }
}

function scrollToFirstError() {
  const el = app.querySelector(".question.has-error");
  if (el) el.scrollIntoView({ behavior: "smooth", block: "center" });
}

function renderResult() {
  const r = state.result;
  const t = state.tool;

  const blocks = r.commands
    .map(
      (c) => `
    <div class="cmdblock">
      <div class="cmdblock__bar">
        <div class="dots"><span></span><span></span><span></span></div>
        <span class="cmdblock__title">${esc(c.title)}</span>
        ${c.language && c.language !== "shell" ? `<span class="lang">${esc(c.language)}</span>` : ""}
        <button class="copy-btn" data-copy="${esc(encodeURIComponent(c.code))}">⧉ copiar</button>
      </div>
      <pre class="cmdblock__code">${highlight(c.code, c.language)}</pre>
      ${c.hint ? `<div class="cmdblock__hint">💡 ${esc(c.hint)}</div>` : ""}
    </div>`
    )
    .join("");

  const warnings = (r.warnings || []).map((w) => `<li>${esc(w)}</li>`).join("");
  const notes = (r.notes || []).map((n) => `<li>${esc(n)}</li>`).join("");

  app.innerHTML = `
    ${topbar()}
    <main class="container">
      <div class="result__head">
        <div>
          <h1 class="result__title">${t.icon} ${esc(r.toolName)} <small>// comandos gerados</small></h1>
          <div class="risk-meter" style="margin-top:10px">
            <span>risco</span>
            <span class="risk-meter__bar"><span class="risk-meter__fill ${r.risk}"></span></span>
            <span>${RISK_LABEL[r.risk] || r.risk}</span>
          </div>
        </div>
      </div>

      <div class="banner banner--risk">⚠️ Estas são operações <b>&nbsp;invasivas&nbsp;</b>. Utilize exclusivamente em
        ambientes e ativos para os quais você possui autorização escrita. A responsabilidade pelo uso é de quem executa.</div>

      ${blocks}

      ${warnings ? `<div class="panel panel--warn"><h3>⚠️ atenção</h3><ul>${warnings}</ul></div>` : ""}
      ${notes ? `<div class="panel"><h3>ℹ️ notas</h3><ul>${notes}</ul></div>` : ""}

      <div class="btn-row">
        <button class="btn btn--primary" data-action="copy-all">⧉ copiar todos</button>
        <button class="btn btn--ghost" data-action="download">⬇ baixar script .sh</button>
        <button class="btn btn--ghost" data-action="edit">✎ editar respostas</button>
        <button class="btn btn--danger" data-action="new">nova configuração</button>
      </div>
    </main>
    ${footer()}
  `;

  app.querySelectorAll("[data-copy]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      await copyText(decodeURIComponent(btn.dataset.copy));
      btn.classList.add("copied");
      btn.textContent = "✓ copiado";
      setTimeout(() => {
        btn.classList.remove("copied");
        btn.textContent = "⧉ copiar";
      }, 1800);
    })
  );

  app.querySelector('[data-action="copy-all"]').addEventListener("click", async () => {
    const text = r.commands.map((c) => `# ${c.title}\n${c.code}`).join("\n\n");
    await copyText(text);
    showToast("Todos os comandos copiados");
  });

  app.querySelector('[data-action="download"]').addEventListener("click", () => downloadScript(r));
  app.querySelector('[data-action="edit"]').addEventListener("click", () => {
    state.view = "wizard";
    render();
  });
  app.querySelector('[data-action="new"]').addEventListener("click", () => {
    state.view = "home";
    state.tool = null;
    state.result = null;
    render();
  });
}

/* ------------------------------------------------------------------ Frame */

function topbar() {
  return `
    <header class="topbar">
      <div class="topbar__inner">
        <button class="logo" data-action="home">
          <span class="logo__mark">🛡</span>
          <span class="logo__name">BLX</span>
        </button>
        <div class="status"><span class="status__dot"></span> API online</div>
      </div>
    </header>`;
}

function footer() {
  return `
    <footer class="footer">
      <div><b>BLX</b> — gerador de comandos de pentest para profissionais autorizados.</div>
      <div>Teste de intrusão sem autorização é crime (Lei 12.737/2012). Use com responsabilidade.</div>
    </footer>`;
}

/* --------------------------------------------------------------- Bootstrap */

async function init() {
  // Logo navigation via event delegation (persists across re-renders).
  document.addEventListener("click", (e) => {
    if (e.target.closest('[data-action="home"]')) {
      state.view = "home";
      render();
    }
  });

  try {
    const data = await api("/api/tools");
    state.tools = data.tools;
    state.categories = data.categories;
    data.categories.forEach((c) => (state.catName[c.id] = c.name));
    state.loaded = true;
  } catch (err) {
    showToast("Não foi possível carregar o catálogo: " + err.message, "error");
  }

  if (!state.loaded) {
    app.innerHTML = `
      ${topbar()}
      <main class="container">
        <div class="empty">// falha ao carregar o catálogo<br /><button class="btn btn--primary" style="margin-top:16px" id="retry">tentar novamente</button></div>
      </main>`;
    document.getElementById("retry").addEventListener("click", init);
    return;
  }

  render();
}

init();
