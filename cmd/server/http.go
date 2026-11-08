package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/chesnokov-an/TradeScope/internal/tradescope"
)

type dataset struct {
	Name    string
	Source  string
	Candles []tradescope.Candle
}

type backtestRequest struct {
	CSV        string            `json:"csv"`
	Capital    float64           `json:"capital"`
	Commission float64           `json:"commission"`
	Slippage   float64           `json:"slippage"`
	Params     tradescope.Params `json:"params"`
}

func defaultRequest() backtestRequest {
	return backtestRequest{
		Capital:    10000,
		Commission: 0.1,
		Slippage:   0.05,
		Params:     tradescope.DefaultParams(),
	}
}

func backtestHandler(sample dataset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "используйте POST"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

		request := defaultRequest()
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON: " + err.Error()})
			return
		}
		if err := validate(request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		current := sample
		if strings.TrimSpace(request.CSV) != "" {
			candles, err := tradescope.ParseCSV(strings.NewReader(request.CSV))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			current = dataset{Name: "Загруженный CSV", Source: "файл пользователя", Candles: candles}
		}

		strategies, err := tradescope.BuildStrategies(current.Candles, request.Params)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		for _, strategy := range strategies {
			if len(current.Candles) <= strategy.Warmup+2 {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("для стратегии %s нужно больше %d свечей", strategy.Name, strategy.Warmup+2),
				})
				return
			}
		}

		cfg := tradescope.Config{
			InitialCapital: request.Capital,
			Commission:     request.Commission / 100,
			Slippage:       request.Slippage / 100,
		}
		results := make([]tradescope.Result, 0, len(strategies))
		responses := make([]strategyResponse, 0, len(strategies))
		for _, strategy := range strategies {
			result, err := tradescope.Run(current.Candles, strategy, cfg)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			results = append(results, result)
			responses = append(responses, buildResponse(result))
		}

		writeJSON(w, http.StatusOK, backtestResponse{
			Dataset: datasetInfo{
				Name: current.Name, Source: current.Source,
				Start: current.Candles[0].Date.Format("2006-01-02"),
				End:   current.Candles[len(current.Candles)-1].Date.Format("2006-01-02"),
				Rows:  len(current.Candles),
			},
			Config: configInfo{
				InitialCapital: request.Capital,
				Commission:     request.Commission,
				Slippage:       request.Slippage,
			},
			Price:       priceSeries(current.Candles),
			Results:     responses,
			Conclusions: tradescope.Compare(results).Conclusions,
		})
	}
}

func validate(r backtestRequest) error {
	values := []struct {
		name  string
		value float64
	}{
		{"капитал", r.Capital}, {"комиссия", r.Commission}, {"проскальзывание", r.Slippage},
		{"нижний уровень RSI", r.Params.RSIOversold}, {"верхний уровень RSI", r.Params.RSIOverbought},
	}
	for _, item := range values {
		if math.IsNaN(item.value) || math.IsInf(item.value, 0) {
			return fmt.Errorf("параметр «%s» содержит недопустимое число", item.name)
		}
	}
	if r.Capital < 100 || r.Capital > 1e9 {
		return fmt.Errorf("капитал должен быть от 100 до 1 000 000 000")
	}
	if r.Commission < 0 || r.Commission > 5 {
		return fmt.Errorf("комиссия должна быть от 0 до 5 %%")
	}
	if r.Slippage < 0 || r.Slippage > 5 {
		return fmt.Errorf("проскальзывание должно быть от 0 до 5 %%")
	}
	return nil
}

type point struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type datasetInfo struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Rows   int    `json:"rows"`
}

type configInfo struct {
	InitialCapital float64 `json:"initialCapital"`
	Commission     float64 `json:"commission"`
	Slippage       float64 `json:"slippage"`
}

type strategyResponse struct {
	Key      string             `json:"key"`
	Name     string             `json:"name"`
	Params   []tradescope.Param `json:"params"`
	Metrics  tradescope.Metrics `json:"metrics"`
	Equity   []point            `json:"equity"`
	Drawdown []point            `json:"drawdown"`
	Trades   []tradescope.Trade `json:"trades"`
}

type backtestResponse struct {
	Dataset     datasetInfo        `json:"dataset"`
	Config      configInfo         `json:"config"`
	Price       []point            `json:"price"`
	Results     []strategyResponse `json:"results"`
	Conclusions []string           `json:"conclusions"`
}

func priceSeries(candles []tradescope.Candle) []point {
	points := make([]point, len(candles))
	for i, candle := range candles {
		points[i] = point{Date: candle.Date.Format("2006-01-02"), Value: candle.Close}
	}
	return points
}

func buildResponse(result tradescope.Result) strategyResponse {
	equityValues := make([]float64, len(result.Equity))
	equity := make([]point, len(result.Equity))
	for i, item := range result.Equity {
		equityValues[i] = item.Equity
		equity[i] = point{Date: item.Date.Format("2006-01-02"), Value: item.Equity}
	}
	drawdownValues := tradescope.DrawdownSeries(result.Config.InitialCapital, equityValues)
	drawdown := make([]point, len(result.Equity))
	for i, item := range result.Equity {
		drawdown[i] = point{Date: item.Date.Format("2006-01-02"), Value: drawdownValues[i]}
	}
	trades := result.Trades
	if trades == nil {
		trades = []tradescope.Trade{}
	}
	return strategyResponse{
		Key: result.StrategyKey, Name: result.StrategyName, Params: result.Params,
		Metrics: result.Metrics, Equity: equity, Drawdown: drawdown, Trades: trades,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("не удалось записать ответ: %v", err)
	}
}
