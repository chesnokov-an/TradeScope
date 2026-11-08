package tradescope

import (
	"math"
	"testing"
)

func TestComputeBasicMetrics(t *testing.T) {
	equity := []float64{100, 110, 99, 120}
	trades := []Trade{
		{Profit: 5, ReturnPct: 5},
		{Profit: -3, ReturnPct: -3},
		{Profit: 7, ReturnPct: 7},
	}
	result := Compute(100, equity, trades, 252)
	if !almostEqual(result.TotalReturnPct, 20, 1e-9) {
		t.Fatalf("доходность %v, ожидалось 20", result.TotalReturnPct)
	}
	if !almostEqual(result.MaxDrawdownPct, 10, 1e-9) {
		t.Fatalf("просадка %v, ожидалось 10", result.MaxDrawdownPct)
	}
	if result.NumberOfTrades != 3 || result.WinningTrades != 2 || result.LosingTrades != 1 {
		t.Fatalf("неверная статистика сделок: %+v", result)
	}
	if !almostEqual(result.WinRatePct, 200.0/3, 1e-9) || !almostEqual(result.AvgTradeReturnPct, 3, 1e-9) {
		t.Fatalf("неверные производные метрики: %+v", result)
	}
}

func TestDrawdownAndSharpe(t *testing.T) {
	if got := MaxDrawdownPct(100, []float64{90, 95, 100}); !almostEqual(got, 10, 1e-9) {
		t.Fatalf("просадка %v, ожидалось 10", got)
	}
	series := DrawdownSeries(100, []float64{100, 120, 90})
	for i, want := range []float64{0, 0, 25} {
		if !almostEqual(series[i], want, 1e-9) {
			t.Fatalf("drawdown[%d] = %v, ожидалось %v", i, series[i], want)
		}
	}
	equity := []float64{100, 110, 132, 171.6}
	expected := 0.2 / 0.1 * math.Sqrt(252)
	if got := SharpeRatio(equity, 252); !almostEqual(got, expected, 1e-6) {
		t.Fatalf("Шарп %v, ожидалось %v", got, expected)
	}
	if got := SharpeRatio([]float64{100, 100, 100, 100}, 252); got != 0 {
		t.Fatalf("при нулевой волатильности Шарп должен быть 0, получено %v", got)
	}
}

func TestComputeWithoutTrades(t *testing.T) {
	result := Compute(1000, []float64{1000, 1000}, nil, 252)
	if result.NumberOfTrades != 0 || result.WinRatePct != 0 || result.MaxDrawdownPct != 0 {
		t.Fatalf("неверные метрики без сделок: %+v", result)
	}
}

func TestCompare(t *testing.T) {
	results := []Result{
		{StrategyName: "SMA", Metrics: Metrics{TotalReturnPct: 10, MaxDrawdownPct: 20, NumberOfTrades: 3}},
		{StrategyName: "RSI", Metrics: Metrics{TotalReturnPct: 5, MaxDrawdownPct: 10, NumberOfTrades: 1}},
		{StrategyName: "MACD", Metrics: Metrics{TotalReturnPct: 30, MaxDrawdownPct: 15, NumberOfTrades: 8}},
	}
	summary := Compare(results)
	if summary.BestReturn != "MACD" || summary.LowestDrawdown != "RSI" || summary.MostTrades != "MACD" {
		t.Fatalf("неверные лидеры: %+v", summary)
	}
	if !almostEqual(summary.ReturnSpreadPct, 25, 1e-9) {
		t.Fatalf("разброс доходности %v, ожидалось 25", summary.ReturnSpreadPct)
	}
	if len(summary.Conclusions) == 0 {
		t.Fatal("ожидались автоматические выводы")
	}
}
