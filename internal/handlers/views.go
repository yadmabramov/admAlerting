package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <table>
        <tr><th>Type</th><th>Name</th><th>Value</th></tr>
        %s
    </table>
    <style>
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</body>
</html>`

func (h *MetricsHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	gauges, counters := h.storage.GetAllMetrics()

	var rows strings.Builder

	// Добавляем gauge метрики
	for name, value := range gauges {
		rows.WriteString(fmt.Sprintf(
			"<tr><td>gauge</td><td>%s</td><td>%.2f</td></tr>", // 2 знака после запятой
			name, value,
		))
	}

	// Добавляем counter метрики
	for name, value := range counters {
		rows.WriteString(fmt.Sprintf(
			"<tr><td>counter</td><td>%s</td><td>%d</td></tr>",
			name, value,
		))
	}

	w.Header().Set("Content-Type", "text/html")
	// Исправленная строка - убираем лишнее форматирование
	fmt.Fprintf(w, htmlTemplate, rows.String())
}
