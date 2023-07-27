package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	backend := flag.String("backend", "https://api.openai.com/v1", "address of the backend service")
	listenOn := flag.String("listen_on", ":8090", "address to listen on")
	dbFile := flag.String("db_file", "./requests.db", "path to the requests.db file")
	apiKey := flag.String("api_key", "", "API key")
	openaiOrg := flag.String("openai_org", "", "OpenAI Organization")
	flag.Parse()

	_, err := url.Parse(*backend)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", *dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create requests table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		request TEXT,
		response TEXT,
		url_path TEXT,
		duration_ms INTEGER
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_show/") && r.Method == http.MethodGet {
			ShowHandler(w, r, db)
			return
		}

		if r.URL.Path == "/_list" && r.Method == http.MethodGet {
			ListHandler(w, r, db)
			return
		}

		if r.URL.Path == "/favicon.ico" {
			// Return an empty/blank logo
			w.Header().Set("Content-Type", "image/x-icon")
			_, err := w.Write([]byte{})
			if err != nil {
				log.Println("Failed to write response body:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

		fmt.Println(r.URL.Path)

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Failed to read request body:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create a new request to the backend
		backendReq, err := http.NewRequest(r.Method, *backend+r.URL.Path, bytes.NewReader(body))
		if err != nil {
			log.Println("Failed to create backend request:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Copy all the headers from the client request to the backend request
		for key, values := range r.Header {
			for _, value := range values {
				backendReq.Header.Add(key, value)
			}
		}

		// Add Authorization header with API key
		if *apiKey != "" {
			backendReq.Header.Set("Authorization", "Bearer "+*apiKey)
		} else {
			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey != "" {
				backendReq.Header.Set("Authorization", "Bearer "+apiKey)
			}
		}

		// Add OpenAI-Organization header
		if *openaiOrg != "" {
			backendReq.Header.Set("OpenAI-Organization", *openaiOrg)
		}

		// Measure the time spent on calling the backend
		startTime := time.Now()
		resp, err := http.DefaultClient.Do(backendReq)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			log.Println("Failed to make backend request:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy all the headers from the backend response to the client response
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Persist the request, response, and duration to the database
		stmt, err := db.Prepare("INSERT INTO requests (request, response, url_path, duration_ms) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Println("Failed to prepare database statement:", err)
			return
		}
		defer stmt.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body:", err)
			return
		}

		os.WriteFile("/tmp/gateway-response-body.json", responseBody, os.ModePerm)

		insertBody := responseBody
		// Check if the response is gzipped
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(bytes.NewReader(responseBody))
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

			insertBody = uncompressedBody
		}
		_, err = stmt.Exec(string(body), string(insertBody), r.URL.Path, duration)
		if err != nil {
			log.Println("Failed to execute database statement:", err)
			return
		}

		// Write the response body to the client
		_, err = w.Write(responseBody)
		if err != nil {
			log.Println("Failed to write response body:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(*listenOn, nil))
}

func convertTimestamp(timestamp time.Time) time.Time {
	// Convert timestamp to current timezone
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Println("Failed to load timezone location:", err)
		return timestamp
	}
	return timestamp.In(location)
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}
