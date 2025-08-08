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
var VERSION string = "v0.0.0 (2025-08-09)"

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

	// Run health endpoint in background.
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})
		_ = http.ListenAndServe(":"+httpPort, nil)
	}()

	// Run pull and push in background.
	go func() {
		for {
			if err := pullAndPush(client, src, dst, alertmanagerApiVersion); err != nil {
				log.Printf("pull/push failed: %v", err)
			}
			time.Sleep(interval)
		}
	}()

	// Keep the process running indefinetly.
	// s: https://stackoverflow.com/a/36419222
	select {}
}

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

	var payload struct {
		Data []map[string]interface{} `json:"data"`
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
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getenvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
