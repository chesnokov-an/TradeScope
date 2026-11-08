// Команда experiment воспроизводит эксперимент из пояснительной записки.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chesnokov-an/TradeScope/internal/tradescope"
)

type output struct {
	GeneratedAt      string                  `json:"generatedAt"`
	Dataset          datasetOutput           `json:"dataset"`
	Config           configOutput            `json:"config"`
	BuyHoldReturnPct float64                 `json:"buyHoldReturnPct"`
	Results          []resultOutput          `json:"results"`
	Comparison       tradescope.Summary      `json:"comparison"`
	Stability        []stabilityPeriodOutput `json:"stability"`
}

type datasetOutput struct {
	Name   string       `json:"name"`
	Source string       `json:"source"`
	Rows   int          `json:"rows"`
	Start  string       `json:"start"`
	End    string       `json:"end"`
	Stats  datasetStats `json:"stats"`
}

type datasetStats struct {
	MinClose              float64 `json:"minClose"`
	MaxClose              float64 `json:"maxClose"`
	MeanClose             float64 `json:"meanClose"`
	MeanDailyReturnPct    float64 `json:"meanDailyReturnPct"`
	DailyStdDevPct        float64 `json:"dailyStdDevPct"`
	AnnualVolatilityPct   float64 `json:"annualVolatilityPct"`
	UpDaysPct             float64 `json:"upDaysPct"`
	MaxDailyGainPct       float64 `json:"maxDailyGainPct"`
	MaxDailyLossPct       float64 `json:"maxDailyLossPct"`
	BuyHoldMaxDrawdownPct float64 `json:"buyHoldMaxDrawdownPct"`
}

type configOutput struct {
	InitialCapital float64 `json:"initialCapital"`
	CommissionPct  float64 `json:"commissionPct"`
	SlippagePct    float64 `json:"slippagePct"`
}

type resultOutput struct {
	Key     string             `json:"key"`
	Name    string             `json:"name"`
	Params  []tradescope.Param `json:"params"`
	Metrics tradescope.Metrics `json:"metrics"`
}

type stabilityPeriodOutput struct {
	Label      string         `json:"label"`
	Start      string         `json:"start"`
	End        string         `json:"end"`
	BuyHoldPct float64        `json:"buyHoldPct"`
	Results    []resultOutput `json:"results"`
}

func main() {
	csvPath := flag.String("csv", "data/samples/aapl_daily.csv", "CSV с котировками")
	outPath := flag.String("out", "report/generated/results.typ", "файл с числами для Typst")
	jsonPath := flag.String("json", "report/generated/results.json", "файл с результатами")
	capital := flag.Float64("capital", 10000, "начальный капитал")
	commission := flag.Float64("commission", 0.1, "комиссия, %")
	slippage := flag.Float64("slippage", 0.05, "проскальзывание, %")
	flag.Parse()

	candles, err := tradescope.LoadCSV(*csvPath)
	if err != nil {
		log.Fatal(err)
	}
	result := buildAll(candles, *capital, *commission, *slippage)

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := writeJSON(*jsonPath, result); err != nil {
		log.Fatal(err)
	}
	if err := writeTypst(*outPath, result); err != nil {
		log.Fatal(err)
	}
	printSummary(result)
}

func buildAll(candles []tradescope.Candle, capital, commission, slippage float64) output {
	cfg := tradescope.Config{InitialCapital: capital, Commission: commission / 100, Slippage: slippage / 100}
	params := tradescope.DefaultParams()
	results := runAll(candles, params, cfg)

	out := output{
		GeneratedAt:      time.Now().Format("2006-01-02 15:04"),
		BuyHoldReturnPct: (candles[len(candles)-1].Close/candles[0].Close - 1) * 100,
		Comparison:       tradescope.Compare(results),
	}
	out.Dataset = datasetOutput{
		Name:   "AAPL — дневные свечи",
		Source: "Yahoo Finance, AAPL, дневные свечи, 02.01.2019 — 31.12.2024, цены с поправкой на дробление акций",
		Rows:   len(candles),
		Start:  candles[0].Date.Format("2006-01-02"),
		End:    candles[len(candles)-1].Date.Format("2006-01-02"),
		Stats:  stats(candles),
	}
	out.Config = configOutput{InitialCapital: capital, CommissionPct: commission, SlippagePct: slippage}
	for _, item := range results {
		out.Results = append(out.Results, toOutput(item))
	}

	split := indexFromDate(candles, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	out.Stability = []stabilityPeriodOutput{
		runPeriod("2019–2022", candles[:split], params, cfg),
		runPeriod("2023–2024", candles[split:], params, cfg),
	}
	return out
}

func runAll(candles []tradescope.Candle, params tradescope.Params, cfg tradescope.Config) []tradescope.Result {
	strategies, err := tradescope.BuildStrategies(candles, params)
	if err != nil {
		log.Fatal(err)
	}
	results := make([]tradescope.Result, 0, len(strategies))
	for _, strategy := range strategies {
		item, err := tradescope.Run(candles, strategy, cfg)
		if err != nil {
			log.Fatalf("%s: %v", strategy.Name, err)
		}
		results = append(results, item)
	}
	return results
}

func runPeriod(label string, candles []tradescope.Candle, params tradescope.Params, cfg tradescope.Config) stabilityPeriodOutput {
	period := stabilityPeriodOutput{
		Label: label, Start: candles[0].Date.Format("2006-01-02"), End: candles[len(candles)-1].Date.Format("2006-01-02"),
		BuyHoldPct: (candles[len(candles)-1].Close/candles[0].Close - 1) * 100,
	}
	for _, item := range runAll(candles, params, cfg) {
		period.Results = append(period.Results, toOutput(item))
	}
	return period
}

func toOutput(result tradescope.Result) resultOutput {
	return resultOutput{Key: result.StrategyKey, Name: result.StrategyName, Params: result.Params, Metrics: result.Metrics}
}

func indexFromDate(candles []tradescope.Candle, date time.Time) int {
	for i, candle := range candles {
		if !candle.Date.Before(date) {
			return i
		}
	}
	return -1
}

func stats(candles []tradescope.Candle) datasetStats {
	result := datasetStats{MinClose: candles[0].Close, MaxClose: candles[0].Close}
	closes := make([]float64, len(candles))
	sum := 0.0
	for i, candle := range candles {
		closes[i] = candle.Close
		sum += candle.Close
		result.MinClose = math.Min(result.MinClose, candle.Close)
		result.MaxClose = math.Max(result.MaxClose, candle.Close)
	}
	result.MeanClose = sum / float64(len(candles))
	result.BuyHoldMaxDrawdownPct = tradescope.MaxDrawdownPct(closes[0], closes)

	returns := make([]float64, 0, len(candles)-1)
	upDays := 0
	result.MaxDailyGainPct = math.Inf(-1)
	result.MaxDailyLossPct = math.Inf(1)
	for i := 1; i < len(candles); i++ {
		change := candles[i].Close/candles[i-1].Close - 1
		returns = append(returns, change)
		if change > 0 {
			upDays++
		}
		result.MaxDailyGainPct = math.Max(result.MaxDailyGainPct, change)
		result.MaxDailyLossPct = math.Min(result.MaxDailyLossPct, change)
	}
	mean := 0.0
	for _, change := range returns {
		mean += change
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, change := range returns {
		variance += (change - mean) * (change - mean)
	}
	variance /= float64(len(returns) - 1)

	result.MeanDailyReturnPct = mean * 100
	result.DailyStdDevPct = math.Sqrt(variance) * 100
	result.AnnualVolatilityPct = math.Sqrt(variance) * math.Sqrt(252) * 100
	result.UpDaysPct = float64(upDays) / float64(len(returns)) * 100
	result.MaxDailyGainPct *= 100
	result.MaxDailyLossPct *= 100
	return result
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeTypst(path string, out output) error {
	var b strings.Builder
	b.WriteString("// Файл создан командой go run ./cmd/experiment.\n")
	b.WriteString("// Содержит фактические результаты эксперимента; вручную не редактируется.\n\n")

	fmt.Fprintf(&b, "#let exp = (\n")
	fmt.Fprintf(&b, "  generatedAt: %q,\n", out.GeneratedAt)
	fmt.Fprintf(&b, "  rows: %d,\n", out.Dataset.Rows)
	fmt.Fprintf(&b, "  start: %q,\n", formatDate(out.Dataset.Start))
	fmt.Fprintf(&b, "  end: %q,\n", formatDate(out.Dataset.End))
	fmt.Fprintf(&b, "  source: %q,\n", out.Dataset.Source)
	fmt.Fprintf(&b, "  capital: %q,\n", formatInt(out.Config.InitialCapital))
	fmt.Fprintf(&b, "  commission: %q,\n", formatFloat(out.Config.CommissionPct, 2))
	fmt.Fprintf(&b, "  slippage: %q,\n", formatFloat(out.Config.SlippagePct, 2))
	fmt.Fprintf(&b, "  buyHold: %q,\n", formatFloat(out.BuyHoldReturnPct, 2))
	fmt.Fprintf(&b, "  minClose: %q,\n", formatFloat(out.Dataset.Stats.MinClose, 2))
	fmt.Fprintf(&b, "  maxClose: %q,\n", formatFloat(out.Dataset.Stats.MaxClose, 2))
	fmt.Fprintf(&b, "  meanClose: %q,\n", formatFloat(out.Dataset.Stats.MeanClose, 2))
	fmt.Fprintf(&b, "  avgDailyReturn: %q,\n", formatFloat(out.Dataset.Stats.MeanDailyReturnPct, 2))
	fmt.Fprintf(&b, "  dailyStdDev: %q,\n", formatFloat(out.Dataset.Stats.DailyStdDevPct, 2))
	fmt.Fprintf(&b, "  annualVolatility: %q,\n", formatFloat(out.Dataset.Stats.AnnualVolatilityPct, 2))
	fmt.Fprintf(&b, "  upDays: %q,\n", formatFloat(out.Dataset.Stats.UpDaysPct, 2))
	fmt.Fprintf(&b, "  maxDailyGain: %q,\n", formatSigned(out.Dataset.Stats.MaxDailyGainPct))
	fmt.Fprintf(&b, "  maxDailyLoss: %q,\n", formatSigned(out.Dataset.Stats.MaxDailyLossPct))
	fmt.Fprintf(&b, "  buyHoldDrawdown: %q,\n", formatFloat(out.Dataset.Stats.BuyHoldMaxDrawdownPct, 2))
	for _, item := range out.Results {
		fmt.Fprintf(&b, "  %s: %s,\n", item.Key, typstResult(item))
	}
	b.WriteString(")\n\n")

	b.WriteString("#let resultsTable = table(\n")
	b.WriteString("  columns: 4,\n  stroke: none,\n  align: (left, right, right, right),\n")
	b.WriteString("  table.header([Показатель], [SMA Crossover], [RSI Mean Reversion], [MACD]),\n  table.hline(),\n")
	rows := []struct {
		label string
		value func(tradescope.Metrics) string
	}{
		{"Итоговая доходность, %", func(m tradescope.Metrics) string { return formatSigned(m.TotalReturnPct) }},
		{"Максимальная просадка, %", func(m tradescope.Metrics) string { return formatFloat(m.MaxDrawdownPct, 2) }},
		{"Количество сделок", func(m tradescope.Metrics) string { return fmt.Sprintf("%d", m.NumberOfTrades) }},
		{"Прибыльные сделки", func(m tradescope.Metrics) string { return fmt.Sprintf("%d", m.WinningTrades) }},
		{"Убыточные сделки", func(m tradescope.Metrics) string { return fmt.Sprintf("%d", m.LosingTrades) }},
		{"Win Rate, %", func(m tradescope.Metrics) string { return formatFloat(m.WinRatePct, 2) }},
		{"Средняя доходность сделки, %", func(m tradescope.Metrics) string { return formatSigned(m.AvgTradeReturnPct) }},
		{"Коэффициент Шарпа", func(m tradescope.Metrics) string { return formatFloat(m.SharpeRatio, 2) }},
		{"Итоговый капитал, у.е.", func(m tradescope.Metrics) string { return formatInt(m.FinalEquity) }},
	}
	for i, row := range rows {
		fmt.Fprintf(&b, "  [%s], [%s], [%s], [%s],\n",
			row.label, row.value(out.Results[0].Metrics), row.value(out.Results[1].Metrics), row.value(out.Results[2].Metrics))
		if i == len(rows)-1 {
			b.WriteString("  table.hline(),\n")
		}
	}
	b.WriteString(")\n\n")

	b.WriteString("#let stabilityTable = table(\n")
	b.WriteString("  columns: 5,\n  stroke: none,\n  align: (left, left, right, right, right),\n")
	b.WriteString("  table.header([Период], [Стратегия], [Доходность, %], [Просадка, %], [Сделок]),\n  table.hline(),\n")
	for _, period := range out.Stability {
		for _, item := range period.Results {
			fmt.Fprintf(&b, "  [%s], [%s], [%s], [%s], [%d],\n",
				period.Label, item.Name, formatSigned(item.Metrics.TotalReturnPct),
				formatFloat(item.Metrics.MaxDrawdownPct, 2), item.Metrics.NumberOfTrades)
		}
	}
	b.WriteString("  table.hline(),\n)\n\n")
	fmt.Fprintf(&b, "#let buyHoldResearch = %q\n", formatSigned(out.Stability[0].BuyHoldPct))
	fmt.Fprintf(&b, "#let buyHoldControl = %q\n", formatSigned(out.Stability[1].BuyHoldPct))
	for i, period := range out.Stability {
		best := period.Results[0]
		for _, item := range period.Results[1:] {
			if item.Metrics.TotalReturnPct > best.Metrics.TotalReturnPct {
				best = item
			}
		}
		fmt.Fprintf(&b, "#let leaderPeriod%d = (name: %q, ret: %q)\n", i+1, best.Name, formatSigned(best.Metrics.TotalReturnPct))
	}
	fmt.Fprintf(&b, "#let leaderReturn = %q\n", out.Comparison.BestReturn)
	fmt.Fprintf(&b, "#let leaderDrawdown = %q\n", out.Comparison.LowestDrawdown)
	fmt.Fprintf(&b, "#let leaderTrades = %q\n", out.Comparison.MostTrades)
	fmt.Fprintf(&b, "#let leaderRiskAdjusted = %q\n", out.Comparison.BestRiskAdjusted)
	fmt.Fprintf(&b, "#let returnSpread = %q\n", formatFloat(out.Comparison.ReturnSpreadPct, 2))

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func typstResult(item resultOutput) string {
	m := item.Metrics
	return fmt.Sprintf(
		`(name: %q, ret: %q, dd: %q, trades: %d, wins: %d, losses: %d, winRate: %q, avg: %q, sharpe: %q, final: %q)`,
		item.Name, formatSigned(m.TotalReturnPct), formatFloat(m.MaxDrawdownPct, 2),
		m.NumberOfTrades, m.WinningTrades, m.LosingTrades, formatFloat(m.WinRatePct, 2),
		formatSigned(m.AvgTradeReturnPct), formatFloat(m.SharpeRatio, 2), formatInt(m.FinalEquity))
}

func printSummary(out output) {
	fmt.Printf("Данные: %s, %d свечей, %s — %s\n", out.Dataset.Name, out.Dataset.Rows, out.Dataset.Start, out.Dataset.End)
	fmt.Printf("Начальный капитал: %.0f, комиссия %.2f %%, проскальзывание %.2f %%\n",
		out.Config.InitialCapital, out.Config.CommissionPct, out.Config.SlippagePct)
	fmt.Printf("%-20s %10s %10s %8s %8s %8s\n", "Стратегия", "Доходн.,%", "Просадка,%", "Сделок", "WinRate", "Шарп")
	for _, item := range out.Results {
		m := item.Metrics
		fmt.Printf("%-20s %10.2f %10.2f %8d %7.2f%% %8.2f\n",
			item.Name, m.TotalReturnPct, m.MaxDrawdownPct, m.NumberOfTrades, m.WinRatePct, m.SharpeRatio)
	}
	fmt.Printf("Покупай и держи: %.2f %%\n", out.BuyHoldReturnPct)
}

func formatDate(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("02.01.2006")
}

func formatFloat(value float64, digits int) string {
	return strings.ReplaceAll(fmt.Sprintf("%.*f", digits, value), ".", ",")
}

func formatSigned(value float64) string {
	if value > 0 {
		return "+" + formatFloat(value, 2)
	}
	formatted := formatFloat(value, 2)
	if strings.HasPrefix(formatted, "-") {
		return "−" + formatted[1:]
	}
	return formatted
}

func formatInt(value float64) string {
	text := fmt.Sprintf("%.0f", value)
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign, text = "-", text[1:]
	}
	var groups []string
	for len(text) > 3 {
		groups = append([]string{text[len(text)-3:]}, groups...)
		text = text[:len(text)-3]
	}
	groups = append([]string{text}, groups...)
	return sign + strings.Join(groups, " ")
}
