package tradescope

import (
	"math"
	"testing"
	"time"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func candlesFromCloses(values []float64) []Candle {
	candles := make([]Candle, len(values))
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, value := range values {
		candles[i] = Candle{
			Date: start.AddDate(0, 0, i), Open: value,
			High: value + 1, Low: value - 1, Close: value, Volume: 1000,
		}
	}
	return candles
}

func TestSMAAndEMA(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	sma := SMA(values, 3)
	ema := EMA(values, 3)
	for i := 0; i < 2; i++ {
		if !math.IsNaN(sma[i]) || !math.IsNaN(ema[i]) {
			t.Fatalf("до прогрева должны быть NaN, индекс %d", i)
		}
	}
	for i, want := range []float64{2, 3, 4} {
		if !almostEqual(sma[i+2], want, 1e-9) || !almostEqual(ema[i+2], want, 1e-9) {
			t.Fatalf("индекс %d: SMA=%v EMA=%v, ожидалось %v", i+2, sma[i+2], ema[i+2], want)
		}
	}
}

func TestRSIWilderExample(t *testing.T) {
	values := []float64{
		44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42,
		45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28,
	}
	result := RSI(values, 14)
	if !math.IsNaN(result[13]) {
		t.Fatal("RSI до прогрева должен быть NaN")
	}
	if !almostEqual(result[14], 70.46, 0.1) {
		t.Fatalf("RSI[14] = %v, ожидалось около 70.46", result[14])
	}
	if got := RSI(make([]float64, 30), 14); !almostEqual(got[14], 50, 1e-9) {
		t.Fatalf("для неизменного ряда RSI должен быть 50, получено %v", got[14])
	}
}

func TestMACDZeroOnFlatSeries(t *testing.T) {
	values := make([]float64, 80)
	for i := range values {
		values[i] = 42
	}
	line, signal, histogram := MACD(values, 12, 26, 9)
	if !math.IsNaN(line[24]) {
		t.Fatal("MACD до прогрева должен быть NaN")
	}
	for i := 25; i < len(values); i++ {
		if !almostEqual(line[i], 0, 1e-9) {
			t.Fatalf("MACD[%d] = %v, ожидался 0", i, line[i])
		}
	}
	if !almostEqual(signal[33], 0, 1e-9) || !almostEqual(histogram[40], 0, 1e-9) {
		t.Fatal("на постоянном ряде сигнальная линия и гистограмма должны быть нулевыми")
	}
}

func TestSMASignals(t *testing.T) {
	params := DefaultParams()
	params.SMAShort, params.SMALong = 2, 4
	strategies, err := BuildStrategies(candlesFromCloses([]float64{5, 4, 3, 2, 3, 4, 5, 4, 3, 2}), params)
	if err != nil {
		t.Fatal(err)
	}
	signals := strategies[0].Signals
	if signals[5] != Buy || signals[8] != Sell {
		t.Fatalf("ожидались Buy на 5 и Sell на 8, получено %v и %v", signals[5], signals[8])
	}
	for i := 0; i < 5; i++ {
		if signals[i] != Hold {
			t.Fatalf("до прогрева на индексе %d не должно быть сигнала", i)
		}
	}
}

func TestRSIAndMACDSignals(t *testing.T) {
	closes := make([]float64, 0, 80)
	price := 100.0
	for i := 0; i < 30; i++ {
		price -= 1
		closes = append(closes, price)
	}
	for i := 0; i < 50; i++ {
		price += 2
		closes = append(closes, price)
	}
	params := DefaultParams()
	strategies, err := BuildStrategies(candlesFromCloses(closes), params)
	if err != nil {
		t.Fatal(err)
	}
	rsi := strategies[1].Signals
	buy, sell := -1, -1
	for i, signal := range rsi {
		if signal == Buy && buy < 0 {
			buy = i
		}
		if signal == Sell && buy >= 0 && sell < 0 {
			sell = i
		}
	}
	if buy <= strategies[1].Warmup || sell <= buy {
		t.Fatalf("неверная последовательность RSI-сигналов: buy=%d sell=%d", buy, sell)
	}

	macdCloses := make([]float64, 0, 120)
	price = 100.0
	for i := 0; i < 40; i++ {
		price -= 0.5
		macdCloses = append(macdCloses, price)
	}
	for i := 0; i < 50; i++ {
		price += 1.2
		macdCloses = append(macdCloses, price)
	}
	for i := 0; i < 30; i++ {
		price -= 1.0
		macdCloses = append(macdCloses, price)
	}
	macdStrategies, err := BuildStrategies(candlesFromCloses(macdCloses), params)
	if err != nil {
		t.Fatal(err)
	}
	macd := macdStrategies[2].Signals
	var buys, sells int
	for _, signal := range macd {
		if signal == Buy {
			buys++
		}
		if signal == Sell {
			sells++
		}
	}
	if buys == 0 || sells == 0 {
		t.Fatalf("MACD должен давать покупки и продажи: buy=%d sell=%d", buys, sells)
	}
}

func TestInvalidStrategyParams(t *testing.T) {
	params := DefaultParams()
	params.SMALong = params.SMAShort
	if _, err := BuildStrategies(candlesFromCloses([]float64{1, 2, 3}), params); err == nil {
		t.Fatal("ожидалась ошибка для длинного периода SMA не больше короткого")
	}
}
