// Файл создан командой go run ./cmd/experiment.
// Содержит фактические результаты эксперимента; вручную не редактируется.

#let exp = (
  generatedAt: "2026-11-09 01:47",
  rows: 1510,
  start: "02.01.2019",
  end: "31.12.2024",
  source: "Yahoo Finance, AAPL, дневные свечи, 02.01.2019 — 31.12.2024, цены с поправкой на дробление акций",
  capital: "10 000",
  commission: "0,10",
  slippage: "0,05",
  buyHold: "534,30",
  minClose: "35,55",
  maxClose: "259,02",
  meanClose: "137,08",
  avgDailyReturn: "0,14",
  dailyStdDev: "1,94",
  annualVolatility: "30,85",
  upDays: "54,01",
  maxDailyGain: "+11,98",
  maxDailyLoss: "−12,86",
  buyHoldDrawdown: "31,43",
  sma: (name: "SMA Crossover", ret: "+193,34", dd: "28,21", trades: 16, wins: 9, losses: 7, winRate: "56,25", avg: "+8,27", sharpe: "0,95", final: "29 334"),
  rsi: (name: "RSI Mean Reversion", ret: "+65,33", dd: "26,22", trades: 6, wins: 6, losses: 0, winRate: "100,00", avg: "+9,14", sharpe: "0,53", final: "16 533"),
  macd: (name: "MACD", ret: "+228,02", dd: "18,91", trades: 54, wins: 28, losses: 26, winRate: "51,85", avg: "+2,46", sharpe: "1,17", final: "32 802"),
)

#let resultsTable = table(
  columns: 4,
  stroke: none,
  align: (left, right, right, right),
  table.header([Показатель], [SMA Crossover], [RSI Mean Reversion], [MACD]),
  table.hline(),
  [Итоговая доходность, %], [+193,34], [+65,33], [+228,02],
  [Максимальная просадка, %], [28,21], [26,22], [18,91],
  [Количество сделок], [16], [6], [54],
  [Прибыльные сделки], [9], [6], [28],
  [Убыточные сделки], [7], [0], [26],
  [Win Rate, %], [56,25], [100,00], [51,85],
  [Средняя доходность сделки, %], [+8,27], [+9,14], [+2,46],
  [Коэффициент Шарпа], [0,95], [0,53], [1,17],
  [Итоговый капитал, у.е.], [29 334], [16 533], [32 802],
  table.hline(),
)

#let stabilityTable = table(
  columns: 5,
  stroke: none,
  align: (left, left, right, right, right),
  table.header([Период], [Стратегия], [Доходность, %], [Просадка, %], [Сделок]),
  table.hline(),
  [2019–2022], [SMA Crossover], [+104,26], [24,51], [11],
  [2019–2022], [RSI Mean Reversion], [+34,02], [26,22], [4],
  [2019–2022], [MACD], [+141,42], [18,91], [30],
  [2023–2024], [SMA Crossover], [+24,07], [11,75], [4],
  [2023–2024], [RSI Mean Reversion], [+8,28], [15,46], [2],
  [2023–2024], [MACD], [+21,93], [11,91], [21],
  table.hline(),
)

#let buyHoldResearch = "+229,10"
#let buyHoldControl = "+100,22"
#let leaderPeriod1 = (name: "MACD", ret: "+141,42")
#let leaderPeriod2 = (name: "SMA Crossover", ret: "+24,07")
#let leaderReturn = "MACD"
#let leaderDrawdown = "MACD"
#let leaderTrades = "MACD"
#let leaderRiskAdjusted = "MACD"
#let returnSpread = "162,69"
