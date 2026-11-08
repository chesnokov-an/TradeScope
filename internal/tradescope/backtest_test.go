package tradescope

import (
	"math"
	"testing"
	"time"
)

func testCandles(opens, closes []float64) []Candle {
	candles := make([]Candle, len(opens))
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range opens {
		candles[i] = Candle{
			Date: start.AddDate(0, 0, i), Open: opens[i],
			High:  math.Max(opens[i], closes[i]) + 1,
			Low:   math.Min(opens[i], closes[i]) - 1,
			Close: closes[i], Volume: 1000,
		}
	}
	return candles
}

func fixed(signals ...Signal) Strategy {
	return Strategy{Key: "fixed", Name: "Fixed", Signals: signals}
}

func TestRunBuyAndSellNextOpen(t *testing.T) {
	candles := testCandles(
		[]float64{10, 10, 20, 15},
		[]float64{11, 12, 21, 16},
	)
	result, err := Run(candles, fixed(Buy, Sell, Hold, Hold), Config{InitialCapital: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("ожидалась одна сделка, получено %d", len(result.Trades))
	}
	trade := result.Trades[0]
	if trade.EntryDate != candles[1].Date || trade.ExitDate != candles[2].Date {
		t.Fatalf("сделка исполнена не на следующих свечах: %+v", trade)
	}
	if trade.EntryPrice != 10 || trade.ExitPrice != 20 {
		t.Fatalf("неверные цены: вход %v, выход %v", trade.EntryPrice, trade.ExitPrice)
	}
	if !almostEqual(trade.Profit, 1000, 1e-9) || !almostEqual(trade.ReturnPct, 100, 1e-9) {
		t.Fatalf("неверный результат сделки: %+v", trade)
	}
	if !almostEqual(result.Metrics.FinalEquity, 2000, 1e-9) || !almostEqual(result.Metrics.TotalReturnPct, 100, 1e-9) {
		t.Fatalf("неверные итоговые метрики: %+v", result.Metrics)
	}
}

func TestRunCommissionAndSlippage(t *testing.T) {
	candles := testCandles(
		[]float64{10, 10, 20, 15},
		[]float64{11, 12, 21, 16},
	)
	cfg := Config{InitialCapital: 1000, Commission: 0.01, Slippage: 0.01}
	result, err := Run(candles, fixed(Buy, Sell, Hold, Hold), cfg)
	if err != nil {
		t.Fatal(err)
	}
	trade := result.Trades[0]
	if !almostEqual(trade.EntryPrice, 10.1, 1e-9) || !almostEqual(trade.ExitPrice, 19.8, 1e-9) {
		t.Fatalf("издержки не учтены в ценах: %+v", trade)
	}
	expectedShares := 1000 / 1.01 / 10.1
	if !almostEqual(trade.Shares, expectedShares, 1e-9) {
		t.Fatalf("акций %v, ожидалось %v", trade.Shares, expectedShares)
	}
	expectedProfit := expectedShares*19.8*(1-0.01) - 1000
	if !almostEqual(trade.Profit, expectedProfit, 1e-6) {
		t.Fatalf("прибыль %v, ожидалось %v", trade.Profit, expectedProfit)
	}
	if result.Metrics.FinalEquity >= 2000 {
		t.Fatal("издержки должны уменьшать капитал")
	}
}

func TestRunForcesCloseOnLastCandle(t *testing.T) {
	candles := testCandles(
		[]float64{10, 10, 20, 15},
		[]float64{11, 12, 21, 16},
	)
	result, err := Run(candles, fixed(Hold, Buy, Hold, Hold), Config{InitialCapital: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("ожидалась принудительно закрытая сделка, получено %d", len(result.Trades))
	}
	trade := result.Trades[0]
	if trade.ExitDate != candles[3].Date || trade.ExitPrice != 16 {
		t.Fatalf("позиция закрыта неверно: %+v", trade)
	}
	if !almostEqual(trade.Profit, -200, 1e-9) || !almostEqual(result.Equity[3].Equity, 800, 1e-9) {
		t.Fatalf("неверный результат принудительного закрытия: %+v", result.Trades[0])
	}
}

func TestRunIgnoresLastSignalAndInvalidSell(t *testing.T) {
	candles := testCandles(
		[]float64{10, 11, 12, 13, 14},
		[]float64{10, 11, 12, 13, 14},
	)
	result, err := Run(candles, fixed(Hold, Hold, Hold, Hold, Buy), Config{InitialCapital: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Trades) != 0 || result.Metrics.FinalEquity != 1000 {
		t.Fatal("сигнал последней свечи не должен исполняться")
	}

	result, err = Run(candles, fixed(Sell, Buy, Buy, Hold, Sell), Config{InitialCapital: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("должна быть только одна позиция, сделок: %d", len(result.Trades))
	}
}

func TestRunEquityIncludesUnrealizedProfit(t *testing.T) {
	candles := testCandles(
		[]float64{10, 10, 20},
		[]float64{11, 12, 21},
	)
	result, err := Run(candles, fixed(Buy, Hold, Hold), Config{InitialCapital: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if !almostEqual(result.Equity[1].Equity, 1200, 1e-9) {
		t.Fatalf("капитал с нереализованной прибылью %v, ожидалось 1200", result.Equity[1].Equity)
	}
}

func TestRunValidation(t *testing.T) {
	candles := testCandles([]float64{10, 11}, []float64{10, 11})
	if _, err := Run(candles, fixed(Buy, Buy), Config{InitialCapital: 1000}); err == nil {
		t.Fatal("ожидалась ошибка для короткого набора")
	}
	long := testCandles([]float64{10, 11, 12}, []float64{10, 11, 12})
	if _, err := Run(long, fixed(Buy), Config{InitialCapital: 1000}); err == nil {
		t.Fatal("ожидалась ошибка при несовпадении числа сигналов")
	}
	if _, err := Run(long, fixed(Hold, Hold, Hold), Config{}); err == nil {
		t.Fatal("ожидалась ошибка для нулевого капитала")
	}
}
