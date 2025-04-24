package fetcher

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

// DownloadNvdFeed 下載指定年份的 NVD JSON.gz 並解壓為原始 JSON bytes
func DownloadNvdv1FeedYear(year uint16, url string) ([]byte, error) {
	url = fmt.Sprintf("%s/nvdcve-1.1-%d.json.gz", url, year)

	return DownloadNvdv1GZIPandUnzip(url)
}

func DownloadNvdv1FeedRecent(url string) ([]byte, error) {
	url = fmt.Sprintf("%s/nvdcve-1.1-recent.json.gz", url)

	return DownloadNvdv1GZIPandUnzip(url)
}

func DownloadNvdv1FeedModified(url string) ([]byte, error) {
	url = fmt.Sprintf("%s/nvdcve-1.1-modified.json.gz", url)

	return DownloadNvdv1GZIPandUnzip(url)
}

func DownloadNvdv1GZIPandUnzip(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download NVD feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status downloading NVD feed: %s", resp.Status)
	}

	// 解壓 gzip
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gzipReader); err != nil {
		return nil, fmt.Errorf("failed to read uncompressed data: %w", err)
	}

	return buf.Bytes(), nil
}
