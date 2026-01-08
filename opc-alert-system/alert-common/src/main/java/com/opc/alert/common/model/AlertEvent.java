package com.opc.alert.common.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;

/**
 * Alert Event
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AlertEvent {

    /**
     * Event ID
     */
    private String eventId;

    /**
     * Rule ID that triggered this alert
     */
    private Long ruleId;

    /**
     * Rule name
     */
    private String ruleName;

    /**
     * Device ID
     */
    private String deviceId;

    /**
     * Device IP
     */
    private String deviceIp;

    /**
     * Metric name
     */
    private String metricName;

    /**
     * Current value
     */
    private Double currentValue;

    /**
     * Threshold value
     */
    private Double threshold;

    /**
     * Alert level: INFO, WARNING, ERROR, CRITICAL
     */
    private String level;

    /**
     * Alert title
     */
    private String title;

    /**
     * Alert message
     */
    private String message;

    /**
     * Triggered timestamp
     */
    private Instant triggeredAt;

    /**
     * Receivers for this alert
     */
    private List<AlertReceiver> receivers;

    /**
     * Alert status: TRIGGERED, SENT, FAILED
     */
    private String status;
}
