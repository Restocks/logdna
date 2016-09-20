package logdna

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IngestBaseURL is the base URL for the LogDNA ingest API.
const IngestBaseURL = "https://logs.logdna.com/logs/ingest"

// DefaultFlushLimit is the number of log lines before we flush to LogDNA
const DefaultFlushLimit = 5000

// Config is used by NewClient to configure new clients.
type Config struct {
	APIKey     string
	LogFile    string
	Hostname   string
	FlushLimit int
}

// Client is a client to the LogDNA logging service.
type Client struct {
	config  Config
	payload payloadJSON
	apiURL  url.URL
}

// logLineJSON represents a log line in the LogDNA ingest API JSON payload.
type logLineJSON struct {
	Timestamp int64  `json:"timestamp"`
	Line      string `json:"line"`
	File      string `json:"file"`
}

// payloadJSON is the complete JSON payload that will be sent to the LogDNA
// ingest API.
type payloadJSON struct {
	Lines []logLineJSON `json:"lines"`
}

// makeIngestURL creats a new URL to the a full LogDNA ingest API endpoint with
// API key and requierd parameters.
func makeIngestURL(cfg Config) (url.URL, error) {
	u, err := url.Parse(IngestBaseURL)

	if err != nil {
		return url.URL{}, err
	}

	u.User = url.User(cfg.APIKey)
	values := url.Values{}
	values.Set("hostname", cfg.Hostname)
	values.Set("now", strconv.FormatInt(time.Time{}.UnixNano(), 10))
	u.RawQuery = values.Encode()

	return *u, err
}

// NewClient returns a Client configured to send logs to the LogDNA ingest API.
func NewClient(cfg Config) *Client {
	if cfg.FlushLimit == 0 {
		cfg.FlushLimit = DefaultFlushLimit
	}

	iu, err := makeIngestURL(cfg)
	if err != nil {
		log.Fatal(err)
	}

	client := Client{
		apiURL: iu,
		config: cfg,
	}

	return &client
}

// Log adds a new log line to Client's payload.
//
// To actually send the logs, Flush() needs to be called.
//
// Flush is called automatically if we reach the client's flush limit.
func (c *Client) Log(t time.Time, msg string) {
	if c.Size() == c.config.FlushLimit {
		c.Flush()
	}

	// Ingest API wants timestamp in milliseconds so we need to round timestamp
	// down from nanoseconds.
	logLine := logLineJSON{
		Timestamp: t.UnixNano() / 1000000,
		Line:      msg,
		File:      c.config.LogFile,
	}
	c.payload.Lines = append(c.payload.Lines, logLine)
}

// Size returns the number of lines waiting to be sent.
func (c *Client) Size() int {
	return len(c.payload.Lines)
}

// Flush sends any buffered logs to LogDNA and clears the buffered logs.
func (c *Client) Flush() error {
	// Return immediately if no logs to send
	if c.Size() == 0 {
		return nil
	}

	jsonPayload, err := json.Marshal(c.payload)
	if err != nil {
		return err
	}

	jsonReader := bytes.NewReader(jsonPayload)

	_, err = http.Post(c.apiURL.String(), "application/json", jsonReader)

	if err != nil {
		return err
	}

	c.payload = payloadJSON{}

	return err
}

// Close closes the client. It also sends any buffered logs.
func (c *Client) Close() error {
	return c.Flush()
}
