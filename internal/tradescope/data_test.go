package tradescope

import (
	"strings"
	"testing"
)

func TestParseCSVValidAndSorted(t *testing.T) {
	input := `Date,Open,High,Low,Close,Volume
2024-01-03,10,12,9,11,1000
2024-01-02,9,11,8,10,900
2024-01-04,11,13,10,12,1100
`
	candles, err := ParseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(candles) != 3 {
		t.Fatalf("получено %d свечей, ожидалось 3", len(candles))
	}
	if got := candles[0].Date.Format("2006-01-02"); got != "2024-01-02" {
		t.Fatalf("данные не отсортированы: %s", got)
	}
	if candles[2].Close != 12 || candles[0].Volume != 900 {
		t.Fatalf("неверно разобраны значения: %+v", candles[0])
	}
}

func TestParseCSVErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"пустой файл", "", "пуст"},
		{"только заголовок", "Date,Open,High,Low,Close\n", "не содержит данных"},
		{"нет close", "Date,Open,High,Low\n2024-01-02,9,11,8\n2024-01-03,10,12,9\n", "close"},
		{"плохая дата", "Date,Open,High,Low,Close\n02/01/2024,9,11,8,10\n2024-01-03,10,12,9,11\n", "дата"},
		{"плохая цена", "Date,Open,High,Low,Close\n2024-01-02,9,11,8,abc\n2024-01-03,10,12,9,11\n", "число"},
		{"повтор даты", "Date,Open,High,Low,Close\n2024-01-02,9,11,8,10\n2024-01-02,10,12,9,11\n", "повторяющаяся"},
		{"OHLC", "Date,Open,High,Low,Close\n2024-01-02,9,8,8,10\n2024-01-03,10,12,9,11\n", "OHLC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseCSV(strings.NewReader(tc.input))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ожидалась ошибка %q, получено: %v", tc.want, err)
			}
		})
	}
}

func TestParseCSVSemicolonAndNoVolume(t *testing.T) {
	input := "Date;Open;High;Low;Close;Volume\n02.01.2024;9,5;11,5;8,5;10,5;900\n03.01.2024;10,5;12,5;9,5;11,5;1000\n"
	candles, err := ParseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if candles[0].Close != 10.5 {
		t.Fatalf("ожидалось 10.5, получено %v", candles[0].Close)
	}

	noVolume := "Date,Open,High,Low,Close\n2024-01-02,9,11,8,10\n2024-01-03,10,12,9,11\n"
	candles, err = ParseCSV(strings.NewReader(noVolume))
	if err != nil {
		t.Fatal(err)
	}
	if candles[0].Volume != 0 {
		t.Fatalf("объём должен быть нулевым, получено %v", candles[0].Volume)
	}
}
