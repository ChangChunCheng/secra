package parser

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
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04Z",    // fallback for short Z
		"2006-01-02T15:04:05Z", // fallback full second + Z
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

// 根目錄結構
type NvdCveFeed struct {
	CVEDataType    string    `json:"CVE_data_type"`
	CVEDataFormat  string    `json:"CVE_data_format"`
	CVEDataVersion string    `json:"CVE_data_version"`
	NumberOfCVEs   string    `json:"CVE_data_numberOfCVEs"`
	Timestamp      string    `json:"CVE_data_timestamp"`
	Items          []CveItem `json:"CVE_Items"`
}

// 每筆 CVE 條目
type CveItem struct {
	CVE              CveCore        `json:"cve"`
	PublishedDate    NvdTime        `json:"publishedDate"`
	LastModifiedDate NvdTime        `json:"lastModifiedDate"`
	Impact           *CveImpact     `json:"impact,omitempty"`
	Configurations   Configurations `json:"configurations"`
}

// 核心內容
type CveCore struct {
	DataMeta    CveMeta        `json:"CVE_data_meta"`
	Description CveDescription `json:"description"`
	// 可擴充 vendor、references、problemtype
}

type CveMeta struct {
	ID string `json:"ID"`
}

type CveDescription struct {
	DescriptionData []LangValue `json:"description_data"`
}

type LangValue struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// Impact 區塊，可能同時含有多版本 CVSS
type CveImpact struct {
	CVSSv1 *CvssMetrics `json:"cvssv1,omitempty"`
	CVSSv2 *CvssV2      `json:"baseMetricV2,omitempty"`
	CVSSv3 *CvssV3      `json:"baseMetricV3,omitempty"`
	CVSSv4 *CvssV4      `json:"cvssMetricV40,omitempty"`
}

// 以下定義各版本 CVSS 結構（僅必要欄位）

type CvssMetrics struct {
	BaseScore float64 `json:"baseScore"`
	Severity  string  `json:"severity"`
	Vector    string  `json:"vectorString"`
}

// CVSS v2 結構（舊版）
type CvssV2 struct {
	CvssData struct {
		BaseScore        float64 `json:"baseScore"`
		Severity         string  `json:"severity"`
		VectorString     string  `json:"vectorString"`
		AccessVector     string  `json:"accessVector"`
		AccessComplexity string  `json:"accessComplexity"`
	} `json:"cvssV2"`
	ExploitabilityScore float64 `json:"exploitabilityScore"`
	ImpactScore         float64 `json:"impactScore"`
}

// CVSS v3 結構（常見於 2016 後）
type CvssV3 struct {
	CvssData struct {
		BaseScore        float64 `json:"baseScore"`
		Severity         string  `json:"baseSeverity"`
		VectorString     string  `json:"vectorString"`
		AttackVector     string  `json:"attackVector"`
		AttackComplexity string  `json:"attackComplexity"`
	} `json:"cvssV3"`
	ExploitabilityScore float64 `json:"exploitabilityScore"`
	ImpactScore         float64 `json:"impactScore"`
}

// CVSS v4 結構（2023 起逐步開始）
type CvssV4 struct {
	CvssData struct {
		Version          string  `json:"version"`
		BaseScore        float64 `json:"baseScore"`
		BaseSeverity     string  `json:"baseSeverity"`
		VectorString     string  `json:"vectorString"`
		AttackVector     string  `json:"attackVector"`
		AttackComplexity string  `json:"attackComplexity"`
	} `json:"cvssData"`
	// 除了 cvssData 外還可包含 threatMetrics 等（如需可擴充）
}

type Configurations struct {
	Nodes []ConfigNode `json:"nodes"`
}

type ConfigNode struct {
	CpeMatch []CpeMatch   `json:"cpe_match"`
	Children []ConfigNode `json:"children,omitempty"`
}

type CpeMatch struct {
	Vulnerable bool   `json:"vulnerable"`
	Cpe23Uri   string `json:"cpe23Uri"`
}
