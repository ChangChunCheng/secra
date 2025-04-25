// parser/nvd_v2_schema.go
package v2

import (
	"fmt"
	"strings"
	"time"
)

type NvdTime struct {
	time.Time
}

func (nt *NvdTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	layouts := []string{
		time.RFC3339,                    // 標準含時區
		"2006-01-02T15:04:05Z",          // 無 offset，Z 結尾
		"2006-01-02T15:04:05",           // 無時區
		"2006-01-02T15:04:05.000",       // 含毫秒無時區 ← ✅ 新增這個
		"2006-01-02T15:04Z",             // 精簡 Z
		"2006-01-02T15:04:05.000Z07:00", // 含毫秒含時區（如有）
	}

	var err error
	for _, layout := range layouts {
		nt.Time, err = time.Parse(layout, s)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("NvdTime: failed to parse %q: %w", s, err)
}

type Nvdv2CveFeed struct {
	ResultsPerPage  int            `json:"resultsPerPage"`
	StartIndex      int            `json:"startIndex"`
	TotalResults    int            `json:"totalResults"`
	Vulnerabilities []Nvdv2CveItem `json:"vulnerabilities"`
}

type Nvdv2CveItem struct {
	Cve Nvdv2CveCore `json:"cve"`
}

type Nvdv2CveCore struct {
	ID               string                `json:"id"`
	SourceIdentifier string                `json:"sourceIdentifier"`
	Published        NvdTime               `json:"published"`
	LastModified     NvdTime               `json:"lastModified"`
	VulnStatus       string                `json:"vulnStatus"`
	Descriptions     []LangValue           `json:"descriptions"`
	Metrics          *Nvdv2CvssMetrics     `json:"metrics,omitempty"`
	Configurations   []Nvdv2Configurations `json:"configurations,omitempty"`
	References       []Nvdv2Reference      `json:"references,omitempty"`
	CveMetadata      *Nvdv2Metadata        `json:"cveMetadata,omitempty"`
}

type Nvdv2CvssMetrics struct {
	CvssMetricV31 []CvssV3Entry `json:"cvssMetricV31,omitempty"`
	CvssMetricV2  []CvssV2Entry `json:"cvssMetricV2,omitempty"`
}

type CvssV3Entry struct {
	Source   string `json:"source"`
	Type     string `json:"type"`
	CvssData struct {
		BaseScore    float64 `json:"baseScore"`
		BaseSeverity string  `json:"baseSeverity"`
		VectorString string  `json:"vectorString"`
	} `json:"cvssData"`
}

type CvssV2Entry struct {
	Source   string `json:"source"`
	Type     string `json:"type"`
	CvssData struct {
		BaseScore    float64 `json:"baseScore"`
		Severity     string  `json:"severity"`
		VectorString string  `json:"vectorString"`
	} `json:"cvssData"`
}

type Nvdv2Reference struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags,omitempty"`
}

type Nvdv2Metadata struct {
	ID          string `json:"cveId"`
	State       string `json:"state"`
	AssignerOrg string `json:"assignerOrg"`
}

type Nvdv2Configurations struct {
	Nodes []Nvdv2ConfigNode `json:"nodes"`
}

type Nvdv2ConfigNode struct {
	Operator string            `json:"operator"`
	Negate   bool              `json:"negate"`
	CpeMatch []Nvdv2CpeMatch   `json:"cpeMatch"`
	Children []Nvdv2ConfigNode `json:"children,omitempty"` // 避免遞迴失敗
}

type Nvdv2CpeMatch struct {
	Vulnerable          bool   `json:"vulnerable"`
	Criteria            string `json:"criteria"`
	VersionEndIncluding string `json:"versionEndIncluding,omitempty"`
	VersionEndExcluding string `json:"versionEndExcluding,omitempty"`
	MatchCriteriaID     string `json:"matchCriteriaId"`
}

type LangValue struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}
