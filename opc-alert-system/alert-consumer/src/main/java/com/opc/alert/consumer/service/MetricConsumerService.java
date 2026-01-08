package com.opc.alert.consumer.service;

import com.opc.alert.common.constants.Constants;
import com.opc.alert.common.model.MetricData;
import com.opc.alert.processor.service.DataProcessorService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.Acknowledgment;
import org.springframework.kafka.support.KafkaHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Kafka Metric Consumer Service
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class MetricConsumerService {

    private final DataProcessorService dataProcessorService;

    /**
     * Consume OPC metrics from Kafka
     */
    @KafkaListener(
            topics = Constants.Topics.OPC_METRICS,
            containerFactory = "kafkaListenerContainerFactory"
    )
    public void consumeMetrics(
            @Payload List<MetricData> metrics,
            @Header(KafkaHeaders.RECEIVED_PARTITION) int partition,
            @Header(KafkaHeaders.OFFSET) long offset,
            Acknowledgment acknowledgment) {

        try {
            log.debug("Received {} metrics from partition {} at offset {}",
                    metrics.size(), partition, offset);

            // Process each metric
            for (MetricData metric : metrics) {
                try {
                    dataProcessorService.process(metric);
                } catch (Exception e) {
                    log.error("Failed to process metric for device: {}, error: {}",
                            metric.getDeviceId(), e.getMessage(), e);
                    // Continue processing other metrics even if one fails
                }
            }

            // Acknowledge after successful processing
            if (acknowledgment != null) {
                acknowledgment.acknowledge();
            }

            log.info("Successfully processed {} metrics from partition {}", metrics.size(), partition);

        } catch (Exception e) {
            log.error("Failed to consume metrics from partition {} at offset {}: {}",
                    partition, offset, e.getMessage(), e);
            // Don't acknowledge - message will be reprocessed
            throw e;
        }
    }

    /**
     * Single message consumer (fallback)
     */
    @KafkaListener(
            topics = Constants.Topics.OPC_METRICS,
            groupId = "opc-alert-consumer-single",
            containerFactory = "kafkaListenerContainerFactory"
    )
    public void consumeSingleMetric(
            @Payload MetricData metric,
            Acknowledgment acknowledgment) {

        try {
            log.debug("Received single metric for device: {}", metric.getDeviceId());

            dataProcessorService.process(metric);

            if (acknowledgment != null) {
                acknowledgment.acknowledge();
            }

            log.debug("Successfully processed metric for device: {}", metric.getDeviceId());

        } catch (Exception e) {
            log.error("Failed to process metric for device: {}: {}",
                    metric.getDeviceId(), e.getMessage(), e);
            throw e;
        }
    }
}
