package com.opc.alert.common.constants;

/**
 * Application Constants
 */
public final class Constants {

    private Constants() {
    }

    /**
     * Kafka Topics
     */
    public static final class Topics {
        public static final String OPC_METRICS = "opc-metrics";
        public static final String PROCESSED_DATA = "processed-data";
        public static final String ALERT_EVENTS = "alert-events";
    }

    /**
     * Alert Levels
     */
    public static final class AlertLevel {
        public static final String INFO = "INFO";
        public static final String WARNING = "WARNING";
        public static final String ERROR = "ERROR";
        public static final String CRITICAL = "CRITICAL";
    }

    /**
     * Alert Status
     */
    public static final class AlertStatus {
        public static final String TRIGGERED = "TRIGGERED";
        public static final String SENT = "SENT";
        public static final String FAILED = "FAILED";
    }

    /**
     * Receiver Types
     */
    public static final class ReceiverType {
        public static final String EMAIL = "EMAIL";
        public static final String SMS = "SMS";
        public static final String WEBHOOK = "WEBHOOK";
    }

    /**
     * Data Quality
     */
    public static final class DataQuality {
        public static final String GOOD = "good";
        public static final String PARTIAL = "partial";
        public static final String BAD = "bad";
    }

    /**
     * Condition Types
     */
    public static final class ConditionType {
        public static final String GREATER_THAN = ">";
        public static final String LESS_THAN = "<";
        public static final String GREATER_EQUAL = ">=";
        public static final String LESS_EQUAL = "<=";
        public static final String EQUAL = "==";
        public static final String NOT_EQUAL = "!=";
    }

    /**
     * InfluxDB Configuration
     */
    public static final class InfluxDB {
        public static final String DEFAULT_MEASUREMENT = "opc_metrics";
        public static final String DEFAULT_BUCKET = "opc_data";
        public static final String DEFAULT_ORG = "opc_organization";
    }
}
