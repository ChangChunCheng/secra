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
}

func FetchNvdv2Feed(nvd_v2_url string, params NvdV2QueryParams) ([]byte, error) {

	query := url.Values{}
	query.Set("pubStartDate", params.PubStartDate.Format(time.RFC3339))
	query.Set("pubEndDate", params.PubEndDate.Format(time.RFC3339))
	query.Set("startIndex", fmt.Sprintf("%d", params.StartIndex))
	query.Set("resultsPerPage", fmt.Sprintf("%d", params.ResultsPerPage))

	fullURL := fmt.Sprintf("%s?%s", nvd_v2_url, query.Encode())

	log.Printf("Fetching NVD v2 feed from: %s", fullURL)

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	if params.ApiKey != "" {
		req.Header.Set("apiKey", params.ApiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
