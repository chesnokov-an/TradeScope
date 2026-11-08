package tradescope

import (
	"fmt"
	"time"
)

// Config задаёт общие для всех стратегий правила исполнения сделок.
type Config struct {
	InitialCapital float64 `json:"initialCapital"`
	// Commission — доля от суммы сделки, 0.001 соответствует 0,1 %.
	Commission float64 `json:"commission"`
	// Slippage — доля ухудшения цены, 0.0005 соответствует 0,05 %.
	Slippage float64 `json:"slippage"`
}

// DefaultConfig возвращает параметры эксперимента по умолчанию.
func DefaultConfig() Config {
	return Config{InitialCapital: 10000, Commission: 0.001, Slippage: 0.0005}
}

// Validate проверяет параметры бэктестинга.
func (c Config) Validate() error {
	if c.InitialCapital <= 0 {
		return fmt.Errorf("начальный капитал должен быть положительным")
	}
	if c.Commission < 0 || c.Commission >= 1 {
		return fmt.Errorf("комиссия должна быть в диапазоне [0, 1)")
	}
	if c.Slippage < 0 || c.Slippage >= 1 {
		return fmt.Errorf("проскальзывание должно быть в диапазоне [0, 1)")
	}
	return nil
}

// Trade — завершённая торговая операция.
type Trade struct {
	EntryDate  time.Time `json:"entryDate"`
	ExitDate   time.Time `json:"exitDate"`
	EntryPrice float64   `json:"entryPrice"`
	ExitPrice  float64   `json:"exitPrice"`
	Shares     float64   `json:"shares"`
	Profit     float64   `json:"profit"`
	ReturnPct  float64   `json:"returnPct"`
	BarsHeld   int       `json:"barsHeld"`
}

// EquityPoint — капитал на закрытии свечи.
type EquityPoint struct {
	Date   time.Time
	Equity float64
}

// Result — результат бэктестинга одной стратегии.
type Result struct {
	StrategyKey  string
	StrategyName string
	Params       []Param
	Config       Config
	Trades       []Trade
	Equity       []EquityPoint
	Metrics      Metrics
}

type position struct {
	shares    float64
	entryDate time.Time
	entryBar  int
	totalCost float64
}

// Run моделирует торговлю по сигналам стратегии.
// Сигнал по закрытию свечи i исполняется по открытию свечи i+1.
// Позиция, открытая в конце периода, закрывается по цене закрытия последней свечи.
func Run(candles []Candle, strategy Strategy, cfg Config) (Result, error) {
	if err := cfg.Validate(); err != nil {
		return Result{}, err
	}
	if len(candles) < 3 {
		return Result{}, fmt.Errorf("для бэктестинга нужно не менее трёх свечей")
	}
	if len(strategy.Signals) != len(candles) {
		return Result{}, fmt.Errorf("стратегия вернула %d сигналов для %d свечей", len(strategy.Signals), len(candles))
	}
	return run(candles, strategy, cfg), nil
}

func run(candles []Candle, s Strategy, cfg Config) Result {
	cash := cfg.InitialCapital
	var open *position
	trades := make([]Trade, 0)
	equity := make([]EquityPoint, 0, len(candles))
	pending := Hold

	for i, candle := range candles {
		switch {
		case pending == Buy && open == nil:
			entryPrice := candle.Open * (1 + cfg.Slippage)
			if entryPrice > 0 {
				cost := cash / (1 + cfg.Commission)
				fee := cost * cfg.Commission
				shares := cost / entryPrice
				open = &position{shares: shares, entryDate: candle.Date, entryBar: i, totalCost: cost + fee}
				cash -= cost + fee
				if cash < 0 && cash > -1e-9 {
					cash = 0
				}
			}
		case pending == Sell && open != nil:
			cash = sell(cash, open, candle, candle.Open, i, cfg, &trades)
			open = nil
		}
		pending = s.Signals[i]

		value := cash
		if open != nil {
			value += open.shares * candle.Close
		}
		equity = append(equity, EquityPoint{Date: candle.Date, Equity: value})
	}

	if open != nil {
		last := candles[len(candles)-1]
		cash = sell(cash, open, last, last.Close, len(candles)-1, cfg, &trades)
		equity[len(equity)-1].Equity = cash
	}

	values := make([]float64, len(equity))
	for i, point := range equity {
		values[i] = point.Equity
	}
	return Result{
		StrategyKey:  s.Key,
		StrategyName: s.Name,
		Params:       s.Params,
		Config:       cfg,
		Trades:       trades,
		Equity:       equity,
		Metrics:      Compute(cfg.InitialCapital, values, trades, 252),
	}
}

func sell(cash float64, open *position, candle Candle, execPrice float64, bar int, cfg Config, trades *[]Trade) float64 {
	exitPrice := execPrice * (1 - cfg.Slippage)
	proceeds := open.shares * exitPrice
	fee := proceeds * cfg.Commission
	profit := cash + proceeds - fee - open.totalCost
	returnPct := 0.0
	if open.totalCost > 0 {
		returnPct = profit / open.totalCost * 100
	}
	*trades = append(*trades, Trade{
		EntryDate:  open.entryDate,
		ExitDate:   candle.Date,
		EntryPrice: open.totalCost / open.shares / (1 + cfg.Commission),
		ExitPrice:  exitPrice,
		Shares:     open.shares,
		Profit:     profit,
		ReturnPct:  returnPct,
		BarsHeld:   bar - open.entryBar,
	})
	return cash + proceeds - fee
}
