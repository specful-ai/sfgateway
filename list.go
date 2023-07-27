package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func ListHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT id, timestamp, url_path, duration_ms, request, response FROM requests ORDER BY timestamp DESC LIMIT 1000")
	if err != nil {
		log.Println("Failed to query database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var (
		id        int
		timestamp time.Time
		urlPath   string
		duration  int
		request   string
		response  string
	)

	var tableRows []string
	for rows.Next() {
		err := rows.Scan(&id, &timestamp, &urlPath, &duration, &request, &response)
		if err != nil {
			log.Println("Failed to scan database row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if bytes.HasPrefix([]byte(response), []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00}) {
			gzipReader, err := gzip.NewReader(bytes.NewReader([]byte(response)))
			if err != nil {
				log.Println("Failed to create gzip reader:", err)
				return
			}
			defer gzipReader.Close()

			uncompressedBody, err := io.ReadAll(gzipReader)
			if err != nil {
				log.Println("Failed to uncompress response body:", err)
				return
			}

			response = string(uncompressedBody)
		} else if !utf8.Valid([]byte(response)) {
			response = "(invalid utf8)"
		}
		tableRow := fmt.Sprintf(
			"<tr><td><pre><a href=\"/_show/%d\">#%d</a>\n\n%s\n\n%s\n\n%d ms</pre></td><td><pre>request: %d bytes\n\n%s</pre></td><td><pre>response: %d bytes\n\n%s</pre></td></tr>",
			id, id,
			convertTimestamp(timestamp).Format("2006-01-02 15:04:05 -07:00"),
			urlPath,
			duration,
			len(request),
			html.EscapeString(truncateString(request, 600)),
			len(response),
			html.EscapeString(truncateString(response, 400)))
		tableRows = append(tableRows, tableRow)
	}

	if len(tableRows) == 0 {
		tableRows = append(tableRows, "<tr><td colspan=\"3\">No results</td></tr>")
	}

	html := fmt.Sprintf("<html><head><style>%s</style></head><body><table>%s</table></body></html>", `
table {
  border-collapse: collapse;
}
td {
  border: 1px solid black;
  padding: 8px;
  vertical-align: top;
}
pre {
  overflow: hidden;
  white-space: pre-wrap;
  max-width: 50ch;
}`, strings.Join(tableRows, ""))
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Println("Failed to write response body:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
