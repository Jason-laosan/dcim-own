package com.opc.alert.common.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * Alert Rule Configuration
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AlertRule {

    /**
     * Rule ID
     */
    private Long id;

    /**
     * Rule name
     */
    private String ruleName;

    /**
     * Rule description
     */
    private String description;

    /**
     * Metric name to monitor
     */
    private String metricName;

    /**
     * Condition type: >, <, >=, <=, ==, !=
     */
    private String conditionType;

    /**
     * Threshold value
     */
    private Double threshold;

    /**
     * Alert level: INFO, WARNING, ERROR, CRITICAL
     */
    private String level;

    /**
     * Time window in seconds for evaluation
     */
    private Integer timeWindowSeconds;

    /**
     * Consecutive violations required to trigger alert
     */
    private Integer consecutiveCount;

    /**
     * Device filter (null means all devices)
     */
    private String deviceFilter;

    /**
     * Tag filters (JSON format)
     */
    private String tagFilters;

    /**
     * Alert template ID
     */
    private Long templateId;

    /**
     * Is rule enabled
     */
    private Boolean enabled;

    /**
     * Created at
     */
    private Instant createdAt;

    /**
     * Updated at
     */
    private Instant updatedAt;

    /**
     * Check if rule matches device
     */
    public boolean matchesDevice(String deviceId) {
        if (deviceFilter == null || deviceFilter.trim().isEmpty()) {
            return true;
        }
        return deviceId.matches(deviceFilter);
    }

    /**
     * Evaluate condition
     */
    public boolean evaluateCondition(Double value) {
        if (value == null || threshold == null) {
            return false;
        }

        return switch (conditionType) {
            case ">" -> value > threshold;
            case "<" -> value < threshold;
            case ">=" -> value >= threshold;
            case "<=" -> value <= threshold;
            case "==" -> Math.abs(value - threshold) < 0.0001;
            case "!=" -> Math.abs(value - threshold) >= 0.0001;
            default -> false;
        };
    }
}
