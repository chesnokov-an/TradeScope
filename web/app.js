const $ = (id) => document.getElementById(id);
const COLORS = { sma: "#4fa3ff", rsi: "#ef8f6b", macd: "#b18cff", price: "#e8b44c", muted: "#8fa1b8" };
const state = { data: null };

let csvText = "";

function shortName(name) {
  if (name.startsWith("SMA")) return "SMA";
  if (name.startsWith("RSI")) return "RSI";
  return name;
}

function collect() {
  const payload = {
    capital: Number($("capital").value),
    commission: Number($("commission").value),
    slippage: Number($("slippage").value),
    params: {
      smaShort: Number($("smaShort").value),
      smaLong: Number($("smaLong").value),
      rsiPeriod: Number($("rsiPeriod").value),
      rsiOversold: Number($("rsiLow").value),
      rsiOverbought: Number($("rsiHigh").value),
      macdFast: Number($("macdFast").value),
      macdSlow: Number($("macdSlow").value),
      macdSignal: Number($("macdSignal").value),
    },
  };
  if (csvText) payload.csv = csvText;
  return payload;
}

async function run() {
  const status = $("status");
  const button = $("run");
  button.disabled = true;
  status.className = "";
  status.textContent = "Расчёт…";
  try {
    const response = await fetch("/api/backtest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(collect()),
    });
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || "Ошибка расчёта");
    state.data = data;
    render(data);
    status.textContent = "Готово";
  } catch (error) {
    status.className = "error";
    status.textContent = error.message;
  } finally {
    button.disabled = false;
  }
}

function render(data) {
  $("dataset").textContent =
    `${data.dataset.name} · ${data.dataset.rows} свечей · ` +
    `${formatDate(data.dataset.start)} — ${formatDate(data.dataset.end)} · ${data.dataset.source}`;

  const results = data.results;
  const best = results.reduce((a, b) => (b.metrics.totalReturnPct > a.metrics.totalReturnPct ? b : a));
  const safest = results.reduce((a, b) => (b.metrics.maxDrawdownPct < a.metrics.maxDrawdownPct ? b : a));
  const trades = results.reduce((sum, item) => sum + item.metrics.numberOfTrades, 0);

  $("summary").innerHTML = [
    card("Лучшая доходность", `${shortName(best.name)}: ${pct(best.metrics.totalReturnPct)}`),
    card("Минимальная просадка", `${shortName(safest.name)}: ${plainPct(safest.metrics.maxDrawdownPct)}`),
    card("Итоговый капитал", money(best.metrics.finalEquity)),
    card("Всего сделок", String(trades)),
  ].join("");
  $("summary").classList.remove("hidden");

  $("resultsBody").innerHTML = results.map((item) => {
    const m = item.metrics;
    return `<tr>
      <td>${item.name}</td>
      <td class="${m.totalReturnPct >= 0 ? "pos" : "neg"}">${pct(m.totalReturnPct)}</td>
      <td>${plainPct(m.maxDrawdownPct)}</td>
      <td>${m.numberOfTrades}</td>
      <td>${m.winningTrades}</td>
      <td>${plainPct(m.winRatePct)}</td>
      <td>${m.sharpeRatio.toFixed(2).replace(".", ",")}</td>
      <td>${money(m.finalEquity)}</td>
    </tr>`;
  }).join("");

  $("conclusions").innerHTML = data.conclusions.map((text) => `<li>${text}</li>`).join("");
  $("results").classList.remove("hidden");
  $("charts").classList.remove("hidden");
  drawCharts(data);
}

function card(label, value) {
  return `<div class="card"><span>${label}</span><strong>${value}</strong></div>`;
}

function pct(value) {
  return `${value >= 0 ? "+" : "−"}${Math.abs(value).toFixed(2).replace(".", ",")} %`;
}

function plainPct(value) {
  return `${value.toFixed(2).replace(".", ",")} %`;
}

function money(value) {
  return new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 0 }).format(value) + " у.е.";
}

function formatDate(value) {
  const [year, month, day] = value.split("-");
  return `${day}.${month}.${year}`;
}

function drawCharts(data) {
  const capital = data.config.initialCapital;
  const priceDates = data.price.map((p) => p.date);

  lineChart($("priceChart"), [
    { label: "Close", color: COLORS.price, values: data.price },
  ], { format: (v) => v.toFixed(0) });

  lineChart($("equityChart"), [
    ...data.results.map((item) => ({
      label: shortName(item.name), color: COLORS[item.key], values: item.equity,
    })),
    { label: "Старт", color: COLORS.muted, dashed: true,
      values: priceDates.map((date) => ({ date, value: capital })) },
  ], { format: (v) => Math.round(v).toLocaleString("ru-RU") });

  lineChart($("drawdownChart"), data.results.map((item) => ({
    label: shortName(item.name),
    color: COLORS[item.key],
    values: item.drawdown.map((p) => ({ date: p.date, value: -p.value })),
  })), { format: (v) => `${Math.abs(v).toFixed(0)} %`, zero: true });

  barChart($("returnsChart"), data.results.map((item) => ({
    label: shortName(item.name), value: item.metrics.totalReturnPct, color: COLORS[item.key],
  })));

  scatterChart($("riskChart"), data.results.map((item) => ({
    label: shortName(item.name),
    x: item.metrics.maxDrawdownPct,
    y: item.metrics.totalReturnPct,
    color: COLORS[item.key],
  })));
}

function setup(canvas) {
  const ratio = window.devicePixelRatio || 1;
  const width = Math.max(canvas.clientWidth, 320);
  const height = canvas.clientHeight || 220;
  canvas.width = Math.round(width * ratio);
  canvas.height = Math.round(height * ratio);
  const ctx = canvas.getContext("2d");
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
  ctx.clearRect(0, 0, width, height);
  ctx.font = "12px system-ui, sans-serif";
  return { ctx, width, height };
}

function lineChart(canvas, series, options = {}) {
  const { ctx, width, height } = setup(canvas);
  const pad = { left: 58, right: 14, top: 32, bottom: 24 };
  const plotW = width - pad.left - pad.right;
  const plotH = height - pad.top - pad.bottom;

  let min = Infinity;
  let max = -Infinity;
  for (const line of series) {
    for (const point of line.values) {
      if (!Number.isFinite(point.value)) continue;
      min = Math.min(min, point.value);
      max = Math.max(max, point.value);
    }
  }
  if (options.zero) { min = Math.min(min, 0); max = Math.max(max, 0); }
  if (!Number.isFinite(min) || min === max) return;
  const span = max - min;
  min -= span * 0.05;
  max += span * 0.05;

  const x = (i) => pad.left + (i / Math.max(series[0].values.length - 1, 1)) * plotW;
  const y = (value) => pad.top + (1 - (value - min) / (max - min)) * plotH;

  ctx.strokeStyle = "#2b394b";
  ctx.fillStyle = COLORS.muted;
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const value = min + ((max - min) * i) / 4;
    const yy = y(value);
    ctx.beginPath();
    ctx.moveTo(pad.left, yy);
    ctx.lineTo(width - pad.right, yy);
    ctx.stroke();
    ctx.textAlign = "right";
    ctx.textBaseline = "middle";
    ctx.fillText(options.format ? options.format(value) : value.toFixed(0), pad.left - 8, yy);
  }

  const dates = series[0].values;
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  for (const index of [0, Math.floor((dates.length - 1) / 2), dates.length - 1]) {
    ctx.fillText(dates[index].date.slice(0, 7), x(index), height - pad.bottom + 8);
  }

  ctx.save();
  ctx.beginPath();
  ctx.rect(pad.left, pad.top, plotW, plotH);
  ctx.clip();
  for (const line of series) {
    ctx.strokeStyle = line.color;
    ctx.lineWidth = line.dashed ? 1.2 : 1.8;
    ctx.setLineDash(line.dashed ? [4, 4] : []);
    ctx.beginPath();
    let started = false;
    line.values.forEach((point, i) => {
      if (!Number.isFinite(point.value)) { started = false; return; }
      if (!started) { ctx.moveTo(x(i), y(point.value)); started = true; }
      else ctx.lineTo(x(i), y(point.value));
    });
    ctx.stroke();
  }
  ctx.restore();
  ctx.setLineDash([]);

  let legendX = pad.left;
  for (const line of series) {
    ctx.strokeStyle = line.color;
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(legendX, 15);
    ctx.lineTo(legendX + 14, 15);
    ctx.stroke();
    ctx.fillStyle = COLORS.muted;
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText(line.label, legendX + 19, 15);
    legendX += ctx.measureText(line.label).width + 38;
  }
}

function barChart(canvas, items) {
  const { ctx, width, height } = setup(canvas);
  const pad = { left: 58, right: 14, top: 26, bottom: 32 };
  const plotW = width - pad.left - pad.right;
  const plotH = height - pad.top - pad.bottom;
  const max = Math.max(...items.map((item) => item.value), 0) * 1.15 || 1;

  ctx.strokeStyle = "#2b394b";
  ctx.fillStyle = COLORS.muted;
  ctx.textAlign = "right";
  ctx.textBaseline = "middle";
  for (let i = 0; i <= 4; i++) {
    const value = (max * i) / 4;
    const yy = height - 24 - (value / max) * plotH;
    ctx.beginPath();
    ctx.moveTo(58, yy);
    ctx.lineTo(width - 14, yy);
    ctx.stroke();
    ctx.fillText(`${Math.round(value)} %`, 50, yy);
  }

  const slot = plotW / items.length;
  const barWidth = Math.min(slot * 0.45, 80);
  items.forEach((item, i) => {
    const cx = pad.left + slot * (i + 0.5);
    const top = height - 24 - (item.value / max) * plotH;
    ctx.fillStyle = item.color;
    ctx.fillRect(cx - barWidth / 2, top, barWidth, height - 24 - top);
    ctx.fillStyle = COLORS.muted;
    ctx.textAlign = "center";
    ctx.textBaseline = "top";
    ctx.fillText(item.label, cx, height - 16);
    ctx.fillStyle = item.color;
    ctx.textBaseline = "bottom";
    ctx.fillText(pct(item.value), cx, top - 6);
  });
}

function scatterChart(canvas, points) {
  const { ctx, width, height } = setup(canvas);
  const pad = { left: 58, right: 20, top: 28, bottom: 36 };
  const plotW = width - pad.left - pad.right;
  const plotH = height - pad.top - pad.bottom;
  const maxX = Math.max(...points.map((p) => p.x)) * 1.2 || 1;
  let minY = Math.min(...points.map((p) => p.y), 0);
  let maxY = Math.max(...points.map((p) => p.y), 0);
  const spanY = maxY - minY || 1;
  minY -= spanY * 0.1;
  maxY += spanY * 0.1;

  const x = (value) => pad.left + (value / maxX) * plotW;
  const y = (value) => pad.top + (1 - (value - minY) / (maxY - minY)) * plotH;

  ctx.strokeStyle = "#2b394b";
  ctx.fillStyle = COLORS.muted;
  ctx.textAlign = "right";
  ctx.textBaseline = "middle";
  for (let i = 0; i <= 4; i++) {
    const value = minY + ((maxY - minY) * i) / 4;
    const yy = y(value);
    ctx.beginPath();
    ctx.moveTo(pad.left, yy);
    ctx.lineTo(width - pad.right, yy);
    ctx.stroke();
    ctx.fillText(`${Math.round(value)} %`, pad.left - 8, yy);
  }
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  for (let i = 0; i <= 4; i++) {
    const value = (maxX * i) / 4;
    ctx.fillText(`${Math.round(value)} %`, x(value), height - pad.bottom + 8);
  }
  ctx.fillText("просадка", pad.left + plotW / 2, height - 14);

  for (const point of points) {
    ctx.fillStyle = point.color;
    ctx.beginPath();
    ctx.arc(x(point.x), y(point.y), 6, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = COLORS.muted;
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText(point.label, x(point.x) + 10, y(point.y));
  }
}

$("run").addEventListener("click", run);
$("reset").addEventListener("click", () => {
  csvText = "";
  $("file").value = "";
  run();
});
$("file").addEventListener("change", async (event) => {
  const file = event.target.files[0];
  if (!file) return;
  csvText = await file.text();
  $("status").textContent = `Загружен файл ${file.name}`;
  run();
});

let resizeTimer = 0;
window.addEventListener("resize", () => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => {
    if (state.data) drawCharts(state.data);
  }, 150);
});

run();
