package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func ShowHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/_show/"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	row := db.QueryRow("SELECT timestamp, url_path, duration_ms, request, response FROM requests WHERE id = ?", id)

	var (
		timestamp time.Time
		urlPath   string
		duration  int
		request   string
		response  string
	)
	err = row.Scan(&timestamp, &urlPath, &duration, &request, &response)
	if err != nil {
		log.Println("Failed to query database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req Request
	err = json.Unmarshal([]byte(request), &req)
	if err != nil {
		log.Println("Failed to unmarshal request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var resp Response
	err = json.Unmarshal([]byte(response), &resp)
	if err != nil {
		log.Println("Failed to unmarshal response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	requestHTML := fmt.Sprintf(
		"<dl><dt>Model:</dt><dd>%s</dd><dt>Messages:</dt><dd>%s</dd><dt>Temperature:</dt><dd>%f</dd></dl>",
		html.EscapeString(req.Model),
		renderMessages(req.Messages),
		req.Temperature,
	)

	responseHTML := fmt.Sprintf(
		"<dl><dt>ID:</dt><dd>%s</dd><dt>Object:</dt><dd>%s</dd><dt>Created:</dt><dd>%d</dd><dt>Model:</dt><dd>%s</dd><dt>Choices:</dt><dd>%s</dd><dt>Usage:</dt><dd>%s</dd></dl>",
		html.EscapeString(resp.ID),
		html.EscapeString(resp.Object),
		resp.Created,
		html.EscapeString(resp.Model),
		renderChoices(resp.Choices),
		renderUsage(resp.Usage),
	)

	html := fmt.Sprintf(
		"<html><head><style>%s</style></head><body><dl><dt>ID:</dt><dd>%d</dd><dt>Timestamp:</dt><dd>%s</dd><dt>URL Path:</dt><dd>%s</dd><dt>Duration (ms):</dt><dd>%d</dd><dt>Request:</dt><dd><p>%d bytes</p>%s</dd><dt>Response:</dt><dd><p>%d bytes</p>%s</dd></dl></body></html>", `
table {
  border-collapse: collapse;
}
td {
  border: 1px solid black;
  padding: 8px;
  vertical-align: top;
}
pre {
  white-space: pre-wrap;
}`,
		id,
		convertTimestamp(timestamp).Format("2006-01-02 15:04:05 -07:00"),
		urlPath,
		duration,
		len(request),
		requestHTML,
		len(response),
		responseHTML,
	)
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Println("Failed to write response body:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func renderMessages(messages []Message) string {
	var sb strings.Builder
	sb.WriteString("<table>")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf(
			"<tr><td><pre>%s</pre></td><td><pre>%s</pre></td></tr>",
			html.EscapeString(msg.Role),
			html.EscapeString(msg.Content)))
	}
	sb.WriteString("</table>")
	return sb.String()
}

func renderChoices(choices []Choice) string {
	var sb strings.Builder
	sb.WriteString("<table>")
	for _, choice := range choices {
		sb.WriteString(fmt.Sprintf(
			"<tr><td><pre>%d</pre></td><td><pre>%s</pre></td><td><pre>%s</pre></td></tr>",
			choice.Index,
			html.EscapeString(choice.Message.Role),
			html.EscapeString(choice.Message.Content)))
	}
	sb.WriteString("</table>")
	return sb.String()
}

func renderUsage(usage Usage) string {
	return fmt.Sprintf(
		"<dl><dt>Prompt Tokens:</dt><dd>%d</dd><dt>Completion Tokens:</dt><dd>%d</dd><dt>Total Tokens:</dt><dd>%d</dd></dl>",
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
	)
}
