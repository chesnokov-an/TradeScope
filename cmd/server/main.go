// Команда server запускает веб-приложение TradeScope.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/chesnokov-an/TradeScope/internal/tradescope"
)

func main() {
	addr := flag.String("addr", ":8080", "адрес HTTP-сервера")
	dataPath := flag.String("data", "data/samples/aapl_daily.csv", "демонстрационный CSV")
	webDir := flag.String("web", "web", "каталог со статическими файлами")
	flag.Parse()

	candles, err := tradescope.LoadCSV(*dataPath)
	if err != nil {
		log.Fatalf("демонстрационные данные: %v", err)
	}
	sample := dataset{
		Name:    "AAPL — дневные свечи",
		Source:  "Yahoo Finance, AAPL, 02.01.2019 — 31.12.2024",
		Candles: candles,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/backtest", backtestHandler(sample))
	mux.Handle("/", http.FileServer(http.Dir(*webDir)))

	log.Printf("TradeScope: http://localhost%s, свечей в демо-наборе: %d", *addr, len(candles))
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
