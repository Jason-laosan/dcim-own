package models

import (
	"bytes"
	"fmt"
	"time"
)

// DataQuality represents the overall quality of collected data
type DataQuality string

const (
	QualityGood    DataQuality = "good"
	QualityPartial DataQuality = "partial"
	QualityBad     DataQuality = "bad"
)

// MetricData represents collected metrics from a device
type MetricData struct {
	DeviceID  string                `json:"device_id"`
	DeviceIP  string                `json:"device_ip"`
	Timestamp time.Time             `json:"timestamp"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Tags      map[string]string     `json:"tags"`
	Quality   DataQuality           `json:"quality"`
}

// MetricValue represents a single metric value
type MetricValue struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Unit    string      `json:"unit"`
	Quality string      `json:"quality"` // Good, Bad, Uncertain
}

// ToLineProtocol converts MetricData to InfluxDB line protocol format
// Format: measurement,tag1=val1,tag2=val2 field1=val1,field2=val2 timestamp
func (m *MetricData) ToLineProtocol(measurement string) []byte {
	var buf bytes.Buffer

	// Measurement name
	buf.WriteString(measurement)

	// Tags
	buf.WriteString(",device_id=")
	buf.WriteString(escapeTagValue(m.DeviceID))
	buf.WriteString(",device_ip=")
	buf.WriteString(escapeTagValue(m.DeviceIP))

	// Additional tags
	for key, value := range m.Tags {
		buf.WriteString(",")
		buf.WriteString(escapeTagKey(key))
		buf.WriteString("=")
		buf.WriteString(escapeTagValue(value))
	}

	buf.WriteString(" ")

	// Fields
	first := true
	for name, metric := range m.Metrics {
		if !first {
			buf.WriteString(",")
		}
		first = false

		buf.WriteString(escapeFieldKey(name))
		buf.WriteString("=")
		buf.WriteString(formatFieldValue(metric.Value))
	}

	// Timestamp (nanoseconds)
	buf.WriteString(" ")
	buf.WriteString(fmt.Sprintf("%d", m.Timestamp.UnixNano()))

	return buf.Bytes()
}

// Helper functions for escaping InfluxDB line protocol
func escapeTagKey(s string) string {
	// Tags keys: escape commas, equal signs, and spaces
	return escape(s, []byte{',', '=', ' '})
}

func escapeTagValue(s string) string {
	// Tag values: escape commas, equal signs, and spaces
	return escape(s, []byte{',', '=', ' '})
}

func escapeFieldKey(s string) string {
	// Field keys: escape commas, equal signs, and spaces
	return escape(s, []byte{',', '=', ' '})
}

func escape(s string, chars []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		needsEscape := false
		for _, ec := range chars {
			if c == ec {
				needsEscape = true
				break
			}
		}
		if needsEscape {
			buf.WriteByte('\\')
		}
		buf.WriteByte(c)
	}
	return buf.String()
}

func formatFieldValue(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return fmt.Sprintf("%f", val)
	case float32:
		return fmt.Sprintf("%f", val)
	case int:
		return fmt.Sprintf("%di", val)
	case int64:
		return fmt.Sprintf("%di", val)
	case int32:
		return fmt.Sprintf("%di", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		// String values must be quoted
		return fmt.Sprintf("\"%s\"", escapeStringValue(val))
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

func escapeStringValue(s string) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			buf.WriteByte('\\')
		}
		buf.WriteByte(c)
	}
	return buf.String()
}
