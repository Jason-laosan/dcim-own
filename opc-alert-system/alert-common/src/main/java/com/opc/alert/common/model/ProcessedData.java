package com.opc.alert.common.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

/**
 * Processed Data ready for storage
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProcessedData {

    private String deviceId;
    private String deviceIp;
    private Instant timestamp;

    /**
     * Processed metric fields
     * Key: metric name
     * Value: processed value
     */
    @Builder.Default
    private Map<String, Object> fields = new HashMap<>();

    /**
     * Tags for InfluxDB
     */
    @Builder.Default
    private Map<String, String> tags = new HashMap<>();

    /**
     * Measurement name for InfluxDB
     */
    private String measurement;

    /**
     * Data quality indicator
     */
    private String quality;

    /**
     * Add a field
     */
    public void addField(String name, Object value) {
        this.fields.put(name, value);
    }

    /**
     * Add a tag
     */
    public void addTag(String key, String value) {
        this.tags.put(key, value);
    }

    /**
     * Get field as Double
     */
    public Double getFieldAsDouble(String fieldName) {
        Object value = fields.get(fieldName);
        if (value == null) {
            return null;
        }
        if (value instanceof Number) {
            return ((Number) value).doubleValue();
        }
        try {
            return Double.parseDouble(value.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }
}
