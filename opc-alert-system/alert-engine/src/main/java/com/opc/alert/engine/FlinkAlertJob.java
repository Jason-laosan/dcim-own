package com.opc.alert.engine;

import com.opc.alert.common.model.AlertEvent;
import com.opc.alert.common.model.ProcessedData;
import com.opc.alert.config.service.AlertConfigService;
import com.opc.alert.engine.function.AlertEvaluationFunction;
import com.opc.alert.engine.source.InfluxDBSource;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import org.apache.flink.connector.kafka.sink.KafkaRecordSerializationSchema;
import org.apache.flink.connector.kafka.sink.KafkaSink;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.time.Duration;

/**
 * Flink Alert Engine Job
 * Reads data from InfluxDB and evaluates alert rules in real-time
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class FlinkAlertJob implements CommandLineRunner {

    private final AlertConfigService configService;

    @Value("${kafka.bootstrap-servers:localhost:9092}")
    private String kafkaBootstrapServers;

    @Value("${kafka.alert-topic:alert-events}")
    private String alertTopic;

    @Value("${influxdb.url}")
    private String influxUrl;

    @Value("${influxdb.token}")
    private String influxToken;

    @Value("${influxdb.org}")
    private String influxOrg;

    @Value("${influxdb.bucket}")
    private String influxBucket;

    @Value("${flink.parallelism:4}")
    private int parallelism;

    @Value("${flink.checkpoint-interval:60000}")
    private long checkpointInterval;

    @Override
    public void run(String... args) throws Exception {
        log.info("Starting Flink Alert Engine");

        // Create Flink execution environment
        StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

        // Configure environment
        env.setParallelism(parallelism);
        env.enableCheckpointing(checkpointInterval);

        // Create InfluxDB source
        InfluxDBSource influxSource = new InfluxDBSource(
                influxUrl, influxToken, influxOrg, influxBucket
        );

        // Read from InfluxDB with 10-minute window
        DataStream<ProcessedData> dataStream = env
                .addSource(influxSource)
                .name("InfluxDB Source")
                .assignTimestampsAndWatermarks(
                        WatermarkStrategy
                                .<ProcessedData>forBoundedOutOfOrderness(Duration.ofSeconds(10))
                                .withTimestampAssigner((data, timestamp) ->
                                        data.getTimestamp().toEpochMilli())
                );

        // Evaluate alerts
        DataStream<AlertEvent> alertStream = dataStream
                .keyBy(ProcessedData::getDeviceId)
                .process(new AlertEvaluationFunction(configService))
                .name("Alert Evaluation");

        // Send alerts to Kafka
        KafkaSink<String> kafkaSink = KafkaSink.<String>builder()
                .setBootstrapServers(kafkaBootstrapServers)
                .setRecordSerializer(KafkaRecordSerializationSchema.builder()
                        .setTopic(alertTopic)
                        .setValueSerializationSchema(new SimpleStringSchema())
                        .build())
                .build();

        alertStream
                .map(alert -> toJson(alert))
                .sinkTo(kafkaSink)
                .name("Kafka Alert Sink");

        // Execute job
        log.info("Flink Alert Engine configured, starting execution...");
        env.execute("OPC Alert Engine");
    }

    /**
     * Convert AlertEvent to JSON
     */
    private String toJson(AlertEvent event) {
        try {
            // Simple JSON serialization (you can use Jackson for more complex scenarios)
            return String.format(
                    "{\"eventId\":\"%s\",\"ruleId\":%d,\"ruleName\":\"%s\"," +
                            "\"deviceId\":\"%s\",\"metricName\":\"%s\"," +
                            "\"currentValue\":%.2f,\"threshold\":%.2f," +
                            "\"level\":\"%s\",\"title\":\"%s\",\"message\":\"%s\"," +
                            "\"triggeredAt\":\"%s\",\"status\":\"%s\"}",
                    event.getEventId(),
                    event.getRuleId(),
                    event.getRuleName(),
                    event.getDeviceId(),
                    event.getMetricName(),
                    event.getCurrentValue(),
                    event.getThreshold(),
                    event.getLevel(),
                    event.getTitle(),
                    event.getMessage(),
                    event.getTriggeredAt(),
                    event.getStatus()
            );
        } catch (Exception e) {
            log.error("Failed to serialize alert event: {}", e.getMessage());
            return "{}";
        }
    }
}
