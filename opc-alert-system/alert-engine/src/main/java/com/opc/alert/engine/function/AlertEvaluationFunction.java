package com.opc.alert.engine.function;

import com.opc.alert.common.model.AlertEvent;
import com.opc.alert.common.model.AlertReceiver;
import com.opc.alert.common.model.AlertRule;
import com.opc.alert.common.model.AlertTemplate;
import com.opc.alert.common.model.ProcessedData;
import com.opc.alert.config.service.AlertConfigService;
import lombok.extern.slf4j.Slf4j;
import org.apache.flink.api.common.state.MapState;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.state.ValueState;
import org.apache.flink.api.common.state.ValueStateDescriptor;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.KeyedProcessFunction;
import org.apache.flink.util.Collector;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Alert Evaluation Function
 * Evaluates metric data against alert rules using Flink state
 */
@Slf4j
public class AlertEvaluationFunction extends KeyedProcessFunction<String, ProcessedData, AlertEvent> {

    private final AlertConfigService configService;

    // State to track consecutive violations per rule
    private transient MapState<Long, Integer> violationCountState;

    // State to track last alert time per rule (for deduplication)
    private transient MapState<Long, Long> lastAlertTimeState;

    // Minimum time between alerts for same rule (milliseconds)
    private static final long MIN_ALERT_INTERVAL = 300000; // 5 minutes

    public AlertEvaluationFunction(AlertConfigService configService) {
        this.configService = configService;
    }

    @Override
    public void open(Configuration parameters) {
        // Initialize state for tracking violations
        MapStateDescriptor<Long, Integer> violationDescriptor = new MapStateDescriptor<>(
                "violation-count",
                Long.class,
                Integer.class
        );
        violationCountState = getRuntimeContext().getMapState(violationDescriptor);

        // Initialize state for tracking last alert time
        MapStateDescriptor<Long, Long> lastAlertDescriptor = new MapStateDescriptor<>(
                "last-alert-time",
                Long.class,
                Long.class
        );
        lastAlertTimeState = getRuntimeContext().getMapState(lastAlertDescriptor);
    }

    @Override
    public void processElement(
            ProcessedData data,
            Context ctx,
            Collector<AlertEvent> out) throws Exception {

        String deviceId = data.getDeviceId();
        log.debug("Evaluating alerts for device: {}", deviceId);

        // Get all enabled alert rules
        List<AlertRule> rules = configService.getAllEnabledRules();

        for (AlertRule rule : rules) {
            try {
                evaluateRule(rule, data, ctx, out);
            } catch (Exception e) {
                log.error("Failed to evaluate rule {} for device {}: {}",
                        rule.getId(), deviceId, e.getMessage(), e);
            }
        }
    }

    /**
     * Evaluate a single rule against data
     */
    private void evaluateRule(
            AlertRule rule,
            ProcessedData data,
            Context ctx,
            Collector<AlertEvent> out) throws Exception {

        // Check if rule applies to this device
        if (!rule.matchesDevice(data.getDeviceId())) {
            return;
        }

        // Get the metric value
        Double value = data.getFieldAsDouble(rule.getMetricName());
        if (value == null) {
            // Metric not present in this data
            return;
        }

        // Evaluate condition
        boolean violated = rule.evaluateCondition(value);

        if (violated) {
            handleViolation(rule, data, value, ctx, out);
        } else {
            handleNonViolation(rule);
        }
    }

    /**
     * Handle rule violation
     */
    private void handleViolation(
            AlertRule rule,
            ProcessedData data,
            Double value,
            Context ctx,
            Collector<AlertEvent> out) throws Exception {

        Long ruleId = rule.getId();

        // Get current violation count
        Integer count = violationCountState.get(ruleId);
        count = (count == null) ? 1 : count + 1;
        violationCountState.put(ruleId, count);

        log.debug("Rule {} violated for device {}, count: {}/{}",
                ruleId, data.getDeviceId(), count, rule.getConsecutiveCount());

        // Check if consecutive violations threshold is met
        if (count >= rule.getConsecutiveCount()) {
            // Check if we should send alert (deduplication)
            Long lastAlertTime = lastAlertTimeState.get(ruleId);
            long currentTime = System.currentTimeMillis();

            if (lastAlertTime == null || (currentTime - lastAlertTime) >= MIN_ALERT_INTERVAL) {
                // Trigger alert
                AlertEvent event = createAlertEvent(rule, data, value);
                out.collect(event);

                // Update last alert time
                lastAlertTimeState.put(ruleId, currentTime);

                log.info("Alert triggered: rule={}, device={}, value={}, threshold={}",
                        rule.getRuleName(), data.getDeviceId(), value, rule.getThreshold());

                // Reset violation count after triggering
                violationCountState.put(ruleId, 0);
            }
        }
    }

    /**
     * Handle non-violation (reset counter)
     */
    private void handleNonViolation(AlertRule rule) throws Exception {
        violationCountState.put(rule.getId(), 0);
    }

    /**
     * Create alert event
     */
    private AlertEvent createAlertEvent(AlertRule rule, ProcessedData data, Double value) {
        // Get template
        AlertTemplate template = configService.getTemplateById(rule.getTemplateId());

        // Get receivers
        List<AlertReceiver> receivers = configService.getReceiversForLevel(rule.getLevel());

        // Prepare template variables
        Map<String, Object> variables = new HashMap<>();
        variables.put("deviceId", data.getDeviceId());
        variables.put("deviceIp", data.getDeviceIp());
        variables.put("metricName", rule.getMetricName());
        variables.put("value", String.format("%.2f", value));
        variables.put("threshold", String.format("%.2f", rule.getThreshold()));
        variables.put("level", rule.getLevel());
        variables.put("timestamp", data.getTimestamp().toString());

        // Render title and message
        String title = template != null ? template.renderTitle(variables) :
                String.format("Alert: %s", rule.getRuleName());
        String message = template != null ? template.renderMessage(variables) :
                String.format("Device %s metric %s = %.2f %s threshold %.2f",
                        data.getDeviceId(), rule.getMetricName(), value,
                        rule.getConditionType(), rule.getThreshold());

        return AlertEvent.builder()
                .eventId(UUID.randomUUID().toString())
                .ruleId(rule.getId())
                .ruleName(rule.getRuleName())
                .deviceId(data.getDeviceId())
                .deviceIp(data.getDeviceIp())
                .metricName(rule.getMetricName())
                .currentValue(value)
                .threshold(rule.getThreshold())
                .level(rule.getLevel())
                .title(title)
                .message(message)
                .triggeredAt(Instant.now())
                .receivers(receivers)
                .status("TRIGGERED")
                .build();
    }
}
