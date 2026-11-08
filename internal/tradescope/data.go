// Package tradescope содержит данные, стратегии, бэктестинг и метрики TradeScope.
package tradescope

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Candle — одна дневная свеча.
type Candle struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

var aliases = map[string]string{
	"date": "date", "дата": "date",
	"open": "open", "open price": "open",
	"high": "high", "high price": "high",
	"low": "low", "low price": "low",
	"close": "close", "close price": "close", "adj close": "close",
	"volume": "volume", "vol": "volume",
}

// LoadCSV читает и проверяет файл с котировками.
func LoadCSV(path string) ([]Candle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть %s: %w", path, err)
	}
	defer file.Close()
	return ParseCSV(file)
}

// ParseCSV разбирает CSV с обязательными столбцами Date, Open, High, Low, Close.
// Поддерживаются запятая и точка с запятой как разделители.
func ParseCSV(r io.Reader) ([]Candle, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать CSV: %w", err)
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, fmt.Errorf("CSV-файл пуст")
	}

	delimiter := ','
	if firstLine := bytes.SplitN(raw, []byte("\n"), 2)[0]; bytes.Count(firstLine, []byte(";")) > bytes.Count(firstLine, []byte(",")) {
		delimiter = ';'
	}

	reader := csv.NewReader(bytes.NewReader(raw))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV-файл не содержит данных")
	}

	columns, err := columnIndex(records[0])
	if err != nil {
		return nil, err
	}

	candles := make([]Candle, 0, len(records)-1)
	for i, record := range records[1:] {
		if blank(record) {
			continue
		}
		candle, err := parseRow(record, columns, delimiter)
		if err != nil {
			return nil, fmt.Errorf("строка %d: %w", i+2, err)
		}
		candles = append(candles, candle)
	}
	if len(candles) < 2 {
		return nil, fmt.Errorf("нужно не менее двух свечей, получено %d", len(candles))
	}

	sort.SliceStable(candles, func(i, j int) bool {
		return candles[i].Date.Before(candles[j].Date)
	})
	for i := 1; i < len(candles); i++ {
		if candles[i].Date.Equal(candles[i-1].Date) {
			return nil, fmt.Errorf("повторяющаяся дата %s", candles[i].Date.Format("2006-01-02"))
		}
	}
	return candles, nil
}

func columnIndex(header []string) (map[string]int, error) {
	index := make(map[string]int, len(header))
	for i, name := range header {
		name = strings.ToLower(strings.TrimSpace(name))
		if field, ok := aliases[name]; ok {
			index[field] = i
		}
	}
	for _, field := range []string{"date", "open", "high", "low", "close"} {
		if _, ok := index[field]; !ok {
			return nil, fmt.Errorf("в CSV отсутствует обязательный столбец %s", field)
		}
	}
	return index, nil
}

func blank(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseRow(record []string, columns map[string]int, delimiter rune) (Candle, error) {
	value := func(field string) (string, error) {
		position, ok := columns[field]
		if !ok {
			return "", nil
		}
		if position >= len(record) {
			return "", fmt.Errorf("нет значения в столбце %s", field)
		}
		return strings.TrimSpace(record[position]), nil
	}

	dateText, err := value("date")
	if err != nil {
		return Candle{}, err
	}
	date, err := parseDate(dateText)
	if err != nil {
		return Candle{}, err
	}

	numbers := make(map[string]float64, 5)
	for _, field := range []string{"open", "high", "low", "close", "volume"} {
		text, err := value(field)
		if err != nil {
			return Candle{}, err
		}
		if text == "" {
			if field == "volume" {
				continue
			}
			return Candle{}, fmt.Errorf("пустое значение в столбце %s", field)
		}
		numbers[field], err = parseNumber(text, delimiter)
		if err != nil {
			return Candle{}, fmt.Errorf("столбец %s: %w", field, err)
		}
	}

	candle := Candle{
		Date:   date,
		Open:   numbers["open"],
		High:   numbers["high"],
		Low:    numbers["low"],
		Close:  numbers["close"],
		Volume: numbers["volume"],
	}
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
	return candle, nil
}

func parseDate(text string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "02.01.2006", "2006/01/02"} {
		if date, err := time.Parse(layout, text); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("некорректная дата %q", text)
}

func parseNumber(text string, delimiter rune) (float64, error) {
	text = strings.ReplaceAll(text, "\u00a0", "")
	text = strings.ReplaceAll(text, " ", "")
	if delimiter == ';' {
		if strings.Contains(text, ",") && !strings.Contains(text, ".") {
			text = strings.ReplaceAll(text, ",", ".")
		} else {
			text = strings.ReplaceAll(text, ",", "")
		}
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("некорректное число %q", text)
	}
	return value, nil
}
