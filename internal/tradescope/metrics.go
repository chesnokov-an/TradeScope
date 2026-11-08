package tradescope

import (
	"fmt"
	"math"
	"strings"
)

// Metrics — показатели эффективности одной стратегии.
type Metrics struct {
	InitialCapital    float64 `json:"initialCapital"`
	FinalEquity       float64 `json:"finalEquity"`
	TotalReturnPct    float64 `json:"totalReturnPct"`
	MaxDrawdownPct    float64 `json:"maxDrawdownPct"`
	NumberOfTrades    int     `json:"numberOfTrades"`
	WinningTrades     int     `json:"winningTrades"`
	LosingTrades      int     `json:"losingTrades"`
	WinRatePct        float64 `json:"winRatePct"`
	AvgTradeReturnPct float64 `json:"avgTradeReturnPct"`
	SharpeRatio       float64 `json:"sharpeRatio"`
}

// Compute рассчитывает метрики по кривой капитала и завершённым сделкам.
func Compute(initialCapital float64, equity []float64, trades []Trade, periodsPerYear float64) Metrics {
	result := Metrics{InitialCapital: initialCapital, NumberOfTrades: len(trades)}
	if len(equity) > 0 {
		result.FinalEquity = equity[len(equity)-1]
	} else {
		result.FinalEquity = initialCapital
	}
	if initialCapital > 0 {
		result.TotalReturnPct = (result.FinalEquity/initialCapital - 1) * 100
	}
	result.MaxDrawdownPct = MaxDrawdownPct(initialCapital, equity)

	sum := 0.0
	for _, trade := range trades {
		sum += trade.ReturnPct
		if trade.Profit > 0 {
			result.WinningTrades++
		} else if trade.Profit < 0 {
			result.LosingTrades++
		}
	}
	if len(trades) > 0 {
		result.WinRatePct = float64(result.WinningTrades) / float64(len(trades)) * 100
		result.AvgTradeReturnPct = sum / float64(len(trades))
	}
	result.SharpeRatio = SharpeRatio(equity, periodsPerYear)
	return result
}

// MaxDrawdownPct возвращает максимальную просадку капитала в процентах.
func MaxDrawdownPct(initialCapital float64, equity []float64) float64 {
	peak, maximum := initialCapital, 0.0
	for _, value := range equity {
		if value > peak {
			peak = value
		}
		if peak > 0 {
			maximum = math.Max(maximum, (peak-value)/peak*100)
		}
	}
	return maximum
}

// DrawdownSeries возвращает просадку в процентах для каждой точки капитала.
func DrawdownSeries(initialCapital float64, equity []float64) []float64 {
	result := make([]float64, len(equity))
	peak := initialCapital
	for i, value := range equity {
		if value > peak {
			peak = value
		}
		if peak > 0 {
			result[i] = (peak - value) / peak * 100
		}
	}
	return result
}

// SharpeRatio считает годовой коэффициент Шарпа по дневным доходностям.
// Безрисковая ставка принята равной нулю.
func SharpeRatio(equity []float64, periodsPerYear float64) float64 {
	if len(equity) < 3 || periodsPerYear <= 0 {
		return 0
	}
	returns := make([]float64, 0, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		if equity[i-1] != 0 {
			returns = append(returns, equity[i]/equity[i-1]-1)
		}
	}
	if len(returns) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range returns {
		mean += value
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, value := range returns {
		variance += (value - mean) * (value - mean)
	}
	std := math.Sqrt(variance / float64(len(returns)-1))
	if std == 0 || math.IsNaN(std) {
		return 0
	}
	return mean / std * math.Sqrt(periodsPerYear)
}

// Summary — итоги сравнения стратегий.
type Summary struct {
	BestReturn       string   `json:"bestReturn"`
	LowestDrawdown   string   `json:"lowestDrawdown"`
	MostTrades       string   `json:"mostTrades"`
	BestRiskAdjusted string   `json:"bestRiskAdjusted"`
	ReturnSpreadPct  float64  `json:"returnSpreadPct"`
	Conclusions      []string `json:"conclusions"`
}

// Compare выбирает лидеров и формирует автоматические выводы.
func Compare(results []Result) Summary {
	summary := Summary{}
	if len(results) == 0 {
		summary.Conclusions = []string{"Нет результатов для сравнения."}
		return summary
	}
	bestReturn, lowestDrawdown, mostTrades := results[0], results[0], results[0]
	bestRatio, bestRatioName := math.Inf(-1), ""
	minReturn, maxReturn := results[0].Metrics.TotalReturnPct, results[0].Metrics.TotalReturnPct

	for _, result := range results {
		m := result.Metrics
		if m.TotalReturnPct > bestReturn.Metrics.TotalReturnPct {
			bestReturn = result
		}
		if m.MaxDrawdownPct < lowestDrawdown.Metrics.MaxDrawdownPct {
			lowestDrawdown = result
		}
		if m.NumberOfTrades > mostTrades.Metrics.NumberOfTrades {
			mostTrades = result
		}
		minReturn = math.Min(minReturn, m.TotalReturnPct)
		maxReturn = math.Max(maxReturn, m.TotalReturnPct)
		ratio := m.TotalReturnPct
		if m.MaxDrawdownPct > 0 {
			ratio /= m.MaxDrawdownPct
		}
		if ratio > bestRatio {
			bestRatio, bestRatioName = ratio, result.StrategyName
		}
	}

	summary.BestReturn = bestReturn.StrategyName
	summary.LowestDrawdown = lowestDrawdown.StrategyName
	summary.MostTrades = mostTrades.StrategyName
	summary.BestRiskAdjusted = bestRatioName
	summary.ReturnSpreadPct = maxReturn - minReturn
	summary.Conclusions = []string{
		fmt.Sprintf("Наибольшую доходность показала стратегия %s: %s %%.", bestReturn.StrategyName, commaNumber(bestReturn.Metrics.TotalReturnPct)),
		fmt.Sprintf("Наименьшая максимальная просадка у стратегии %s: %s %%.", lowestDrawdown.StrategyName, commaNumber(lowestDrawdown.Metrics.MaxDrawdownPct)),
		fmt.Sprintf("Больше всего завершённых сделок у стратегии %s: %d.", mostTrades.StrategyName, mostTrades.Metrics.NumberOfTrades),
		fmt.Sprintf("Разброс доходности между стратегиями — %s п.п.", commaNumber(summary.ReturnSpreadPct)),
	}
	if bestReturn.StrategyName == lowestDrawdown.StrategyName {
		summary.Conclusions = append(summary.Conclusions, "Лидер по доходности одновременно имеет наименьшую просадку.")
	} else {
		summary.Conclusions = append(summary.Conclusions, "Лидер по доходности не совпадает с лидером по минимальной просадке.")
	}
	return summary
}

func commaNumber(value float64) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f", value), ".", ",")
}
