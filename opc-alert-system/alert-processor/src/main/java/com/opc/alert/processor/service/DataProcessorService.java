package com.opc.alert.processor.service;

import com.opc.alert.common.model.MetricData;
import com.opc.alert.common.model.ProcessedData;
import com.opc.alert.storage.service.InfluxDBWriter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;

/**
 * Data Processor Service
 * Processes raw metric data and prepares it for storage
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class DataProcessorService {

    private final InfluxDBWriter influxDBWriter;

    /**
     * Process metric data
     * This method can be extended to add custom data transformations:
     * - Data validation
     * - Unit conversion
     * - Data aggregation
     * - Calculated fields
     */
    public void process(MetricData metricData) {
        try {
            log.debug("Processing metric data for device: {}", metricData.getDeviceId());

            // Convert to ProcessedData
            ProcessedData processedData = transformData(metricData);

            // Apply custom transformations
            applyTransformations(processedData, metricData);

            // Write to InfluxDB
            influxDBWriter.write(processedData);

            log.debug("Successfully processed data for device: {}", metricData.getDeviceId());
        } catch (Exception e) {
            log.error("Failed to process metric data for device: {}", metricData.getDeviceId(), e);
            throw new RuntimeException("Data processing failed", e);
        }
    }

    /**
     * Transform MetricData to ProcessedData
     */
    private ProcessedData transformData(MetricData metricData) {
        ProcessedData.ProcessedDataBuilder builder = ProcessedData.builder()
                .deviceId(metricData.getDeviceId())
                .deviceIp(metricData.getDeviceIp())
                .timestamp(metricData.getTimestamp())
                .measurement("opc_metrics")
                .quality(metricData.getQuality());

        ProcessedData processedData = builder.build();

        // Add metrics as fields
        if (metricData.getMetrics() != null) {
            for (Map.Entry<String, MetricData.MetricValue> entry : metricData.getMetrics().entrySet()) {
                String fieldName = entry.getKey();
                MetricData.MetricValue metricValue = entry.getValue();

                // Store the value
                processedData.addField(fieldName, metricValue.getValue());

                // Store quality as a separate field if needed
                if (metricValue.getQuality() != null) {
                    processedData.addField(fieldName + "_quality", metricValue.getQuality());
                }

                // Store unit as a tag if needed
                if (metricValue.getUnit() != null && !metricValue.getUnit().isEmpty()) {
                    processedData.addTag(fieldName + "_unit", metricValue.getUnit());
                }
            }
        }

        // Add custom tags
        if (metricData.getTags() != null) {
            for (Map.Entry<String, String> entry : metricData.getTags().entrySet()) {
                processedData.addTag(entry.getKey(), entry.getValue());
            }
        }

        return processedData;
    }

    /**
     * Apply custom transformations
     * Override this method or add configuration to customize data processing
     */
    private void applyTransformations(ProcessedData processedData, MetricData originalData) {
        // Example: Add calculated fields
        // For temperature metrics, you might want to add Fahrenheit if Celsius is provided
        if (processedData.getFields().containsKey("temperature")) {
            Double celsius = processedData.getFieldAsDouble("temperature");
            if (celsius != null) {
                Double fahrenheit = celsius * 9.0 / 5.0 + 32.0;
                processedData.addField("temperature_f", fahrenheit);
            }
        }

        // Example: Add timestamp-based tags
        processedData.addTag("hour", String.valueOf(processedData.getTimestamp().atZone(java.time.ZoneId.of("UTC")).getHour()));

        // Example: Data validation - filter out invalid data
        validateData(processedData);

        // Add more custom transformations here as needed
    }

    /**
     * Validate processed data
     */
    private void validateData(ProcessedData processedData) {
        // Example: Remove null or invalid values
        processedData.getFields().entrySet().removeIf(entry -> {
            Object value = entry.getValue();
            if (value == null) {
                log.warn("Removing null field: {} for device: {}",
                        entry.getKey(), processedData.getDeviceId());
                return true;
            }
            // Add more validation logic as needed
            return false;
        });
    }
}
