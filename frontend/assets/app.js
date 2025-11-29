const state = {
  apiBase: localStorage.getItem('pgweb.apiBase') || 'http://localhost:8080',
  schema: '',
  table: '',
};

const els = {
  apiInput: document.getElementById('api-base'),
  toast: document.getElementById('toast'),
  connectMessage: document.getElementById('connect-message'),
  schemasList: document.getElementById('schemas-list'),
  tablesList: document.getElementById('tables-list'),
  viewsList: document.getElementById('views-list'),
  indexesList: document.getElementById('indexes-list'),
  columnsList: document.getElementById('columns-list'),
  dataPreview: document.getElementById('data-preview'),
  queryResult: document.getElementById('query-result'),
  activeSchema: document.getElementById('active-schema'),
  activeTable: document.getElementById('active-table'),
};

els.apiInput.value = state.apiBase;
els.apiInput.addEventListener('change', () => {
  state.apiBase = els.apiInput.value.trim();
  localStorage.setItem('pgweb.apiBase', state.apiBase);
  showToast('API base updated');
});

document.body.addEventListener('htmx:configRequest', (event) => {
  const target = event.target;
  if (!target) return;

  let path = event.detail.path || target.getAttribute('hx-get') || target.getAttribute('hx-post') || '';
  if (!path) return;

  if (path.includes('{schema}')) {
    if (!state.schema) {
      event.preventDefault();
      showToast('Select a schema first');
      return;
    }
    path = path.replaceAll('{schema}', encodeURIComponent(state.schema));
  }
  if (path.includes('{table}')) {
    if (!state.table) {
      event.preventDefault();
      showToast('Select a table first');
      return;
    }
    path = path.replaceAll('{table}', encodeURIComponent(state.table));
  }

  if (!/^https?:/i.test(path)) {
    path = buildApiUrl(path);
  }
  event.detail.path = path;
  event.detail.headers['Accept'] = 'application/json';
});

document.body.addEventListener('htmx:afterRequest', (event) => {
  const target = event.target;
  if (!target) return;
  const key = target.dataset.jsonTarget;
  if (!key) return;

  if (!event.detail.successful) {
    handleError(event);
    return;
  }

  let payload = null;
  try {
    payload = JSON.parse(event.detail.xhr.responseText || '{}');
  } catch (err) {
    handleError(event, 'Failed parsing response JSON');
    return;
  }

  const handler = responseHandlers[key];
  if (handler) {
    handler(payload, target);
  }
});

document.body.addEventListener('htmx:responseError', handleError);
document.body.addEventListener('htmx:sendError', handleError);

const responseHandlers = {
  connect: (payload) => {
    els.connectMessage.textContent = payload.message || 'Connected';
  },
  validate: (payload) => {
    showToast(payload.message || 'Connection healthy');
  },
  schemas: (payload) => {
    renderSchemas(payload.schemas || []);
  },
  tables: (payload) => {
    if (payload.schema) {
      state.schema = payload.schema;
      updateSchemaChip();
    }
    renderTables(payload.tables || []);
  },
  views: (payload) => {
    renderSimpleList(els.viewsList, payload || [], 'view');
  },
  indexes: (payload) => {
    renderIndexes(payload || []);
  },
  columns: (payload) => {
    renderColumns(payload.columns || []);
  },
  'table-data': (payload) => {
    renderTableData(payload.rows || []);
  },
  query: (payload) => {
    if (payload.rows_affected !== undefined) {
      els.queryResult.innerHTML = `<p>${payload.rows_affected} row(s) affected.</p>`;
    } else {
      renderResultTable(els.queryResult, payload.columns || [], payload.rows || []);
    }
  },
};

function renderSchemas(items) {
  els.schemasList.innerHTML = '';
  if (!items.length) {
    els.schemasList.innerHTML = '<p class="muted">No schemas.</p>';
    return;
  }
  items.forEach((schema) => {
    const btn = document.createElement('button');
    btn.textContent = schema;
    btn.className = 'ghost';
    if (schema === state.schema) btn.classList.add('active');
    btn.addEventListener('click', () => selectSchema(schema, btn));
    els.schemasList.appendChild(btn);
  });
}

function renderTables(items) {
  els.tablesList.innerHTML = '';
  if (!items.length) {
    els.tablesList.innerHTML = '<p class="muted">No tables.</p>';
    return;
  }
  items.forEach((table) => {
    const btn = document.createElement('button');
    btn.textContent = table;
    btn.className = 'ghost';
    if (table === state.table) btn.classList.add('active');
    btn.addEventListener('click', () => selectTable(table, btn));
    els.tablesList.appendChild(btn);
  });
}

function renderSimpleList(container, payload, key) {
  container.innerHTML = '';
  const list = Array.isArray(payload) ? payload : payload[key + 's'] || [];
  if (!list.length) {
    container.innerHTML = '<p class="muted">No data.</p>';
    return;
  }
  list.forEach((item) => {
    const text = typeof item === 'string' ? item : item[key];
    const row = document.createElement('div');
    row.textContent = text;
    container.appendChild(row);
  });
}

function renderIndexes(items) {
  els.indexesList.innerHTML = '';
  if (!items.length) {
    els.indexesList.innerHTML = '<p class="muted">No indexes.</p>';
    return;
  }
  items.forEach((idx) => {
    const row = document.createElement('div');
    row.innerHTML = `<strong>${idx.index}</strong> <span class="muted">(${idx.table})</span>`;
    els.indexesList.appendChild(row);
  });
}

function renderColumns(columns) {
  if (!columns.length) {
    els.columnsList.innerHTML = '<p class="muted">No columns.</p>';
    return;
  }
  const rows = columns
    .map(
      (c) =>
        `<tr><td>${c.name}</td><td>${c.type}</td><td>${c.constraints.join(', ') || '—'}</td></tr>`
    )
    .join('');
  els.columnsList.innerHTML = `<table><thead><tr><th>Name</th><th>Type</th><th>Constraints</th></tr></thead><tbody>${rows}</tbody></table>`;
}

function renderTableData(rows) {
  if (!rows.length) {
    els.dataPreview.innerHTML = '<p class="muted">No rows.</p>';
    return;
  }
  const columns = Object.keys(rows[0]);
  renderResultTable(els.dataPreview, columns, rows);
}

function renderResultTable(container, columns, rows) {
  if (!columns.length) {
    container.innerHTML = '<p class="muted">No data.</p>';
    return;
  }
  const head = columns.map((c) => `<th>${c}</th>`).join('');
  const body = rows
    .map((row) => {
      const cells = columns
        .map((c) => `<td>${formatValue(row[c])}</td>`)
        .join('');
      return `<tr>${cells}</tr>`;
    })
    .join('');
  container.innerHTML = `<table><thead><tr>${head}</tr></thead><tbody>${body}</tbody></table>`;
}

function formatValue(value) {
  if (value === null || value === undefined) return '<span class="muted">NULL</span>';
  if (Array.isArray(value)) return value.join(', ');
  return value;
}

function selectSchema(schema, btn) {
  state.schema = schema;
  state.table = '';
  updateSchemaChip();
  updateTableChip();
  highlightSelection(els.schemasList, btn);
  els.tablesList.innerHTML = '<p class="muted">Loading…</p>';
  triggerRefresh('tables');
  triggerRefresh('views');
  triggerRefresh('indexes');
}

function selectTable(table, btn) {
  state.table = table;
  updateTableChip();
  highlightSelection(els.tablesList, btn);
  triggerRefresh('columns');
  triggerRefresh('table-data');
}

function triggerRefresh(action) {
  const map = {
    tables: 'btn-refresh-tables',
    views: 'btn-refresh-views',
    indexes: 'btn-refresh-indexes',
    columns: 'btn-refresh-columns',
    'table-data': 'btn-refresh-data',
  };
  const id = map[action];
  if (!id) return;
  const node = document.getElementById(id);
  if (node) node.click();
}

function highlightSelection(container, activeBtn) {
  container.querySelectorAll('button').forEach((btn) => btn.classList.remove('active'));
  if (activeBtn) activeBtn.classList.add('active');
}

function updateSchemaChip() {
  els.activeSchema.textContent = state.schema ? `Schema: ${state.schema}` : 'No schema selected';
}

function updateTableChip() {
  els.activeTable.textContent = state.table ? `Table: ${state.table}` : 'No table selected';
}

function buildApiUrl(path) {
  if (!path.startsWith('/')) {
    return `${state.apiBase.replace(/\/$/, '')}/${path}`;
  }
  return `${state.apiBase.replace(/\/$/, '')}${path}`;
}

function handleError(event, fallback) {
  const xhr = event.detail?.xhr;
  let message = fallback || xhr?.responseText || 'Request failed';
  try {
    const parsed = JSON.parse(message);
    message = parsed.message || JSON.stringify(parsed);
  } catch (err) {
    // ignore
  }
  showToast(message || 'Request failed');
}

let toastTimer;
function showToast(message) {
  if (!message) return;
  els.toast.textContent = message;
  els.toast.classList.add('show');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    els.toast.classList.remove('show');
  }, 3000);
}

// Pre-load schema list on startup
setTimeout(() => document.getElementById('btn-load-schemas').click(), 500);
