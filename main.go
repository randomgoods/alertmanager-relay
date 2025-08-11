package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// VERSION contains the version info
var VERSION string = "v0.0.1 (2025-08-09)"

func main() {
	interval := getenvDuration("POLL_INTERVAL", 60*time.Second)
	src := getenv("SRC_ALERTMANAGER_URL", "http://src-alertmanager:9093")
	dst := getenv("DST_ALERTMANAGER_URL", "http://dst-alertmanager:9093")
	alertmanagerApiVersion := getenv("ALERTMANAGER_API_VERSION", "v1")
	httpPort := getenv("HTTP_PORT", "8080")

	log.Printf("alertmanager-relay %v", VERSION)
	log.Println("---")
	log.Printf("POLL_INTERVAL=%v", interval)
	log.Printf("SRC_ALERTMANAGER_URL=%v", src)
	log.Printf("DST_ALERTMANAGER_URL=%v", dst)
	log.Printf("ALERTMANAGER_API_VERSION=%v", alertmanagerApiVersion)
	log.Printf("HTTP_PORT=%v", httpPort)

	client := &http.Client{Timeout: 5 * time.Second}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	// Run health endpoint in background.
	go http.ListenAndServe(":"+httpPort, nil)

	// Run pull and push
	for {
		if err := pullAndPush(client, src, dst, alertmanagerApiVersion); err != nil {
			log.Printf("pull/push failed: %v", err)
		}
		time.Sleep(interval)
	}
}

// AlertV1 represents a single alert returned by /api/v1/alerts
type AlertV1 struct {
	Annotations  map[string]string `json:"annotations,omitempty"`
	Labels       map[string]string `json:"labels"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL,omitempty"`
}

// AlertV2 represents a single alert returned by /api/v2/alerts
type AlertV2 struct {
	Annotations  map[string]string `json:"annotations"`
	Labels       map[string]string `json:"labels"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	Fingerprint  string            `json:"fingerprint"`
	GeneratorURL string            `json:"generatorURL"`
	// Receivers    []string          `json:"receivers"`
	Status struct {
		InhibitedBy []string `json:"inhibitedBy"`
		SilencedBy  []string `json:"silencedBy"`
		State       string   `json:"state"`
	} `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// pullAndPush handles the API interactions with the source and destination alertmanager instances.
func pullAndPush(c *http.Client, src, dst, apiVersion string) error {
	// pull
	reqSrc, _ := http.NewRequestWithContext(context.Background(), "GET", src+"/api/"+apiVersion+"/alerts", nil)
	reqSrc.Header.Set("Content-Type", "application/json")

	if user, pass := os.Getenv("SRC_AUTH_USERNAME"), os.Getenv("SRC_AUTH_PASSWORD"); user != "" && pass != "" {
		reqSrc.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
	}

	respSrc, err := c.Do(reqSrc)
	if err != nil {
		return err
	}
	defer respSrc.Body.Close()

	switch apiVersion {
	case "v1":
		var payload struct {
			Data []AlertV1 `json:"data"`
		}
		if err := json.NewDecoder(respSrc.Body).Decode(&payload); err != nil {
			return err
		}
		if len(payload.Data) == 0 {
			log.Println("Nothing to forward. Awaiting next cycle.")
			return nil // nothing to forward
		}

		// push
		body, _ := json.Marshal(payload.Data)
		return post(c, dst, apiVersion, body)
	default:
		var alerts []AlertV2
		if err := json.NewDecoder(respSrc.Body).Decode(&alerts); err != nil {
			return err
		}
		if len(alerts) == 0 {
			log.Println("Nothing to forward. Awaiting next cycle.")
			return nil // nothing to forward
		}

		// push
		body, _ := json.Marshal(alerts)
		return post(c, dst, apiVersion, body)
	}
}

// post handles the REST request to send the data to the destination alertmanager.
func post(c *http.Client, dst, apiVersion string, body []byte) error {
	reqDst, _ := http.NewRequestWithContext(context.Background(), "POST", dst+"/api/"+apiVersion+"/alerts", bytes.NewReader(body))
	reqDst.Header.Set("Content-Type", "application/json")

	if user, pass := os.Getenv("DST_AUTH_USERNAME"), os.Getenv("DST_AUTH_PASSWORD"); user != "" && pass != "" {
		reqDst.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
	}

	respDst, err := c.Do(reqDst)
	if err != nil {
		return err
	}
	defer respDst.Body.Close()
	if respDst.StatusCode >= 300 {
		return fmt.Errorf("destination returned %s", respDst.Status)
	}
	log.Println("Forwarding successful")
	return nil
}

// helpers
// getenv fetches the value of an environment variable as string.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getenvDuration fetches the value of an environment variable as time.Duration.
func getenvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
