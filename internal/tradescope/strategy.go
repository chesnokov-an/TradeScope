package tradescope

import (
	"fmt"
	"math"
	"strconv"
)

// Signal — торговый сигнал, рассчитанный по закрытию свечи.
type Signal int

const (
	Hold Signal = iota
	Buy
	Sell
)

// Param — отображаемый параметр стратегии.
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Params задаёт настройки всех трёх стратегий.
type Params struct {
	SMAShort      int     `json:"smaShort"`
	SMALong       int     `json:"smaLong"`
	RSIPeriod     int     `json:"rsiPeriod"`
	RSIOversold   float64 `json:"rsiOversold"`
	RSIOverbought float64 `json:"rsiOverbought"`
	MACDFast      int     `json:"macdFast"`
	MACDSlow      int     `json:"macdSlow"`
	MACDSignal    int     `json:"macdSignal"`
}

// DefaultParams возвращает параметры эксперимента по умолчанию.
func DefaultParams() Params {
	return Params{
		SMAShort: 20, SMALong: 50,
		RSIPeriod: 14, RSIOversold: 30, RSIOverbought: 70,
		MACDFast: 12, MACDSlow: 26, MACDSignal: 9,
	}
}

func (p Params) validate() error {
	if p.SMAShort < 2 || p.SMALong <= p.SMAShort {
		return fmt.Errorf("SMA: короткий период должен быть не меньше 2 и меньше длинного")
	}
	if p.RSIPeriod < 2 {
		return fmt.Errorf("период RSI должен быть не меньше 2")
	}
	if p.RSIOversold <= 0 || p.RSIOverbought >= 100 || p.RSIOversold >= p.RSIOverbought {
		return fmt.Errorf("уровни RSI должны удовлетворять 0 < нижний < верхний < 100")
	}
	if p.MACDFast < 2 || p.MACDSignal < 2 || p.MACDSlow <= p.MACDFast {
		return fmt.Errorf("MACD: периоды должны быть не меньше 2, медленный больше быстрого")
	}
	return nil
}

// Strategy — готовая стратегия с рассчитанными сигналами.
type Strategy struct {
	Key     string
	Name    string
	Warmup  int
	Params  []Param
	Signals []Signal
}

// BuildStrategies проверяет параметры и рассчитывает сигналы трёх стратегий.
func BuildStrategies(candles []Candle, p Params) ([]Strategy, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	values := closes(candles)
	shortSMA := SMA(values, p.SMAShort)
	longSMA := SMA(values, p.SMALong)
	rsi := RSI(values, p.RSIPeriod)
	macdLine, signalLine, _ := MACD(values, p.MACDFast, p.MACDSlow, p.MACDSignal)

	return []Strategy{
		{
			Key: "sma", Name: "SMA Crossover", Warmup: p.SMALong,
			Params:  []Param{{"Короткий период", strconv.Itoa(p.SMAShort)}, {"Длинный период", strconv.Itoa(p.SMALong)}},
			Signals: crossSignals(shortSMA, longSMA),
		},
		{
			Key: "rsi", Name: "RSI Mean Reversion", Warmup: p.RSIPeriod + 1,
			Params:  []Param{{"Период RSI", strconv.Itoa(p.RSIPeriod)}, {"Перепроданность", paramNumber(p.RSIOversold)}, {"Перекупленность", paramNumber(p.RSIOverbought)}},
			Signals: rsiSignals(rsi, p.RSIOversold, p.RSIOverbought),
		},
		{
			Key: "macd", Name: "MACD", Warmup: p.MACDSlow + p.MACDSignal - 1,
			Params:  []Param{{"Быстрая EMA", strconv.Itoa(p.MACDFast)}, {"Медленная EMA", strconv.Itoa(p.MACDSlow)}, {"Сигнальная EMA", strconv.Itoa(p.MACDSignal)}},
			Signals: crossSignals(macdLine, signalLine),
		},
	}, nil
}

func closes(candles []Candle) []float64 {
	values := make([]float64, len(candles))
	for i, candle := range candles {
		values[i] = candle.Close
	}
	return values
}

func crossSignals(fast, slow []float64) []Signal {
	signals := make([]Signal, len(fast))
	for i := range fast {
		switch {
		case crossedAbove(fast, slow, i):
			signals[i] = Buy
		case crossedBelow(fast, slow, i):
			signals[i] = Sell
		}
	}
	return signals
}

func rsiSignals(values []float64, low, high float64) []Signal {
	signals := make([]Signal, len(values))
	for i := range values {
		switch {
		case crossedAboveLevel(values, i, low):
			signals[i] = Buy
		case crossedAboveLevel(values, i, high):
			signals[i] = Sell
		}
	}
	return signals
}

// SMA возвращает простую скользящую среднюю; первые period-1 значений — NaN.
func SMA(values []float64, period int) []float64 {
	result := nanSlice(len(values))
	if period <= 0 || len(values) < period {
		return result
	}
	sum := 0.0
	for i, value := range values {
		sum += value
		if i >= period {
			sum -= values[i-period]
		}
		if i >= period-1 {
			result[i] = sum / float64(period)
		}
	}
	return result
}

// EMA возвращает экспоненциальную скользящую среднюю с коэффициентом 2/(period+1).
func EMA(values []float64, period int) []float64 {
	result := nanSlice(len(values))
	if period <= 0 {
		return result
	}
	start := firstValid(values)
	if start < 0 || start+period > len(values) {
		return result
	}
	sum := 0.0
	for i := start; i < start+period; i++ {
		if math.IsNaN(values[i]) {
			return result
		}
		sum += values[i]
	}
	result[start+period-1] = sum / float64(period)
	alpha := 2.0 / float64(period+1)
	for i := start + period; i < len(values); i++ {
		if math.IsNaN(values[i]) {
			return result
		}
		result[i] = alpha*values[i] + (1-alpha)*result[i-1]
	}
	return result
}

// RSI рассчитывает индекс относительной силы по методу Уайлдера.
func RSI(values []float64, period int) []float64 {
	result := nanSlice(len(values))
	if period <= 0 || len(values) <= period {
		return result
	}
	gainSum, lossSum := 0.0, 0.0
	for i := 1; i <= period; i++ {
		change := values[i] - values[i-1]
		if change > 0 {
			gainSum += change
		} else {
			lossSum -= change
		}
	}
	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)
	result[period] = rsiValue(avgGain, avgLoss)
	for i := period + 1; i < len(values); i++ {
		change := values[i] - values[i-1]
		avgGain = (avgGain*float64(period-1) + math.Max(change, 0)) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + math.Max(-change, 0)) / float64(period)
		result[i] = rsiValue(avgGain, avgLoss)
	}
	return result
}

// MACD возвращает линию MACD, сигнальную линию и гистограмму.
func MACD(values []float64, fast, slow, signal int) ([]float64, []float64, []float64) {
	fastEMA := EMA(values, fast)
	slowEMA := EMA(values, slow)
	line := nanSlice(len(values))
	for i := range values {
		if !math.IsNaN(fastEMA[i]) && !math.IsNaN(slowEMA[i]) {
			line[i] = fastEMA[i] - slowEMA[i]
		}
	}
	signalLine := EMA(line, signal)
	histogram := nanSlice(len(values))
	for i := range values {
		if !math.IsNaN(line[i]) && !math.IsNaN(signalLine[i]) {
			histogram[i] = line[i] - signalLine[i]
		}
	}
	return line, signalLine, histogram
}

func rsiValue(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		if avgGain == 0 {
			return 50
		}
		return 100
	}
	return 100 - 100/(1+avgGain/avgLoss)
}

func nanSlice(length int) []float64 {
	result := make([]float64, length)
	for i := range result {
		result[i] = math.NaN()
	}
	return result
}

func firstValid(values []float64) int {
	for i, value := range values {
		if !math.IsNaN(value) {
			return i
		}
	}
	return -1
}

func validAt(values []float64, i int) bool {
	return i >= 0 && i < len(values) && !math.IsNaN(values[i])
}

func crossedAbove(first, second []float64, i int) bool {
	return i > 0 && validAt(first, i) && validAt(second, i) &&
		validAt(first, i-1) && validAt(second, i-1) &&
		first[i-1] <= second[i-1] && first[i] > second[i]
}

func crossedBelow(first, second []float64, i int) bool {
	return i > 0 && validAt(first, i) && validAt(second, i) &&
		validAt(first, i-1) && validAt(second, i-1) &&
		first[i-1] >= second[i-1] && first[i] < second[i]
}

func crossedAboveLevel(values []float64, i int, level float64) bool {
	return i > 0 && validAt(values, i) && validAt(values, i-1) &&
		values[i-1] <= level && values[i] > level
}

func paramNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
