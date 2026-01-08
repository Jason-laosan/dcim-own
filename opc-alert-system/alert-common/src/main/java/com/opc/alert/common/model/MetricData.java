package com.opc.alert.common.model;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.Map;

/**
 * OPC Metric Data from Kafka
 * Corresponds to opc-collector MetricData structure
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MetricData {

    @JsonProperty("device_id")
    private String deviceId;

    @JsonProperty("device_ip")
    private String deviceIp;

    @JsonProperty("timestamp")
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss.SSSXXX", timezone = "UTC")
    private Instant timestamp;

    @JsonProperty("metrics")
    private Map<String, MetricValue> metrics;

    @JsonProperty("tags")
    private Map<String, String> tags;

    @JsonProperty("quality")
    private String quality; // good, partial, bad

    /**
     * Metric Value
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class MetricValue {
        private String name;
        private Object value;
        private String unit;
        private String quality; // Good, Bad, Uncertain

        /**
         * Get value as Double
         */
        public Double getValueAsDouble() {
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

        /**
         * Get value as String
         */
        public String getValueAsString() {
            return value != null ? value.toString() : null;
        }
    }
}
