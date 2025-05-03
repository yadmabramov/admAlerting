package handlers

import (
	"net/http"
	"strconv"
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
        {{METRICS_ROWS}}
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
		rows.WriteString("<tr><td>gauge</td><td>" + name + "</td><td>" + strconv.FormatFloat(value, 'f', 2, 64) + "</td></tr>")
	}

	// Добавляем counter метрики
	for name, value := range counters {
		rows.WriteString("<tr><td>counter</td><td>" + name + "</td><td>" + strconv.FormatInt(value, 10) + "</td></tr>")
	}

	w.Header().Set("Content-Type", "text/html")

	html := strings.Replace(htmlTemplate, "{{METRICS_ROWS}}", rows.String(), 1)
	w.Write([]byte(html))
}
