#import "../dozen_template.typ": dozen_code

#heading(numbering: none, level: 2)[А.1. Движок бэктестинга]

Фрагмент показывает основной цикл исполнения сигналов. Сигнал, рассчитанный на свече $i$, сохраняется в переменной `pending` и исполняется на следующей итерации по цене открытия. Открытая в конце позиция закрывается принудительно.

#dozen_code(caption: [Основной цикл движка бэктестинга])[
  ```go
  for i, candle := range candles {
      switch {
      case pending == Buy && open == nil:
          entryPrice := candle.Open * (1 + cfg.Slippage)
          if entryPrice > 0 {
              cost := cash / (1 + cfg.Commission)
              fee := cost * cfg.Commission
              shares := cost / entryPrice
              open = &position{
                  shares:    shares,
                  entryDate: candle.Date,
                  entryBar:  i,
                  totalCost: cost + fee,
              }
              cash -= cost + fee
          }
      case pending == Sell && open != nil:
          cash = sell(cash, open, candle, candle.Open, i, cfg, &trades)
          open = nil
      }
      pending = signals[i]

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
  ```
]

#heading(numbering: none, level: 2)[А.2. Проверка CSV]

Проверка OHLC и положительности цен выполняется до добавления свечи в набор. При ошибке возвращается сообщение с номером строки, а приложение не завершается аварийно.

#dozen_code(caption: [Проверка цен и OHLC])[
  ```go
  if candle.Open <= 0 || candle.High <= 0 || candle.Low <= 0 || candle.Close <= 0 {
      return Candle{}, fmt.Errorf("цены должны быть положительными")
  }
  if candle.High < candle.Low || candle.High < candle.Open || candle.High < candle.Close ||
      candle.Low > candle.Open || candle.Low > candle.Close {
      return Candle{}, fmt.Errorf("нарушено соотношение OHLC")
  }
  if candle.Volume < 0 {
      return Candle{}, fmt.Errorf("объём не может быть отрицательным")
  }
  ```
]

#heading(numbering: none, level: 2)[А.3. Запуск и воспроизводимость]

Приложение запускается командой `go run ./cmd/server` и открывается по адресу `http://localhost:8080`. Эксперимент воспроизводится командой `go run ./cmd/experiment`; она сохраняет JSON с результатами и формирует файл с числами для отчёта. Проверка проекта выполняется командами `go test ./...`, `go vet ./...` и `go build ./...`. Компиляция пояснительной записки выполняется командой `typst compile report/main.typ report/coursework.pdf`.
