// fetcher/nvd_v2.go
package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type NvdV2QueryParams struct {
	PubStartDate   time.Time
	PubEndDate     time.Time
	StartIndex     int
	ResultsPerPage int
	ApiKey         string

	MaxRetries int
	RetryDelay time.Duration
}

func FetchNvdv2Feed(nvd_v2_url string, params NvdV2QueryParams) ([]byte, error) {
	query := url.Values{}
	query.Set("pubStartDate", params.PubStartDate.Format(time.RFC3339))
	query.Set("pubEndDate", params.PubEndDate.Format(time.RFC3339))
	query.Set("startIndex", fmt.Sprintf("%d", params.StartIndex))
	query.Set("resultsPerPage", fmt.Sprintf("%d", params.ResultsPerPage))

	fullURL := fmt.Sprintf("%s?%s", nvd_v2_url, query.Encode())

	maxRetries := params.MaxRetries
	if maxRetries <= 0 { maxRetries = 3 }
	
	retryDelay := params.RetryDelay
	if retryDelay <= 0 { retryDelay = 5 * time.Second }

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			log.Printf("⏳ [Retry %d/%d] Waiting %v before retrying...", i, maxRetries, retryDelay)
			time.Sleep(retryDelay)
		}

		log.Printf("📥 Fetching NVD v2 feed: %s", fullURL)

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "SECRA-Vulnerability-Bot/1.0")
		if params.ApiKey != "" {
			req.Header.Set("apiKey", params.ApiKey)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("⚠️ Network error: %v", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			continue 
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected response: %s, body: %s", resp.Status, string(body))
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		return data, err
	}

	return nil, fmt.Errorf("failed to fetch NVD feed after %d retries", maxRetries)
}
