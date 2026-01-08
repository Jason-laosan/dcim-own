package com.opc.alert.common.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * Alert Receiver Configuration
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AlertReceiver {

    /**
     * Receiver ID
     */
    private Long id;

    /**
     * Receiver name
     */
    private String receiverName;

    /**
     * Receiver type: EMAIL, SMS, WEBHOOK
     */
    private String receiverType;

    /**
     * Contact information (email, phone, webhook URL)
     */
    private String contact;

    /**
     * Alert level filter (comma-separated): INFO,WARNING,ERROR,CRITICAL
     * If null or empty, receives all levels
     */
    private String levelFilter;

    /**
     * Is receiver enabled
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
     * Check if receiver should receive alert of given level
     */
    public boolean shouldReceiveLevel(String alertLevel) {
        if (levelFilter == null || levelFilter.trim().isEmpty()) {
            return true;
        }
        String[] levels = levelFilter.split(",");
        for (String level : levels) {
            if (level.trim().equalsIgnoreCase(alertLevel)) {
                return true;
            }
        }
        return false;
    }
}
