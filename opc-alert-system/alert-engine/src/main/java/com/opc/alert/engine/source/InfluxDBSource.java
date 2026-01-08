package com.opc.alert.engine.source;

import com.influxdb.client.InfluxDBClient;
import com.influxdb.client.InfluxDBClientFactory;
import com.influxdb.query.FluxRecord;
import com.influxdb.query.FluxTable;
import com.opc.alert.common.model.ProcessedData;
import lombok.extern.slf4j.Slf4j;
import org.apache.flink.streaming.api.functions.source.RichSourceFunction;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * InfluxDB Source for Flink
 * Reads recent data from InfluxDB for alert evaluation
 */
@Slf4j
public class InfluxDBSource extends RichSourceFunction<ProcessedData> {

    private final String url;
    private final String token;
    private final String org;
    private final String bucket;

    private transient InfluxDBClient client;
    private volatile boolean running = true;

    // Query interval in milliseconds (default: 30 seconds)
    private static final long QUERY_INTERVAL = 30000;

    // Time window in minutes (default: 10 minutes)
    private static final int TIME_WINDOW_MINUTES = 10;

    public InfluxDBSource(String url, String token, String org, String bucket) {
        this.url = url;
        this.token = token;
        this.org = org;
        this.bucket = bucket;
    }

    @Override
    public void open(org.apache.flink.configuration.Configuration parameters) {
        log.info("Opening InfluxDB connection: url={}, org={}, bucket={}", url, org, bucket);
        client = InfluxDBClientFactory.create(url, token.toCharArray(), org, bucket);
    }

    @Override
    public void run(SourceContext<ProcessedData> ctx) throws Exception {
        log.info("Starting InfluxDB source, polling interval: {}ms, time window: {}min",
                QUERY_INTERVAL, TIME_WINDOW_MINUTES);

        while (running) {
            try {
                // Query recent data
                List<ProcessedData> data = queryRecentData();

                // Emit data
                for (ProcessedData item : data) {
                    ctx.collect(item);
                }

                log.debug("Fetched {} records from InfluxDB", data.size());

                // Wait before next query
                Thread.sleep(QUERY_INTERVAL);

            } catch (InterruptedException e) {
                log.info("InfluxDB source interrupted");
                break;
            } catch (Exception e) {
                log.error("Error querying InfluxDB: {}", e.getMessage(), e);
                // Continue running even if one query fails
                Thread.sleep(QUERY_INTERVAL);
            }
        }
    }

    /**
     * Query recent data from InfluxDB
     */
    private List<ProcessedData> queryRecentData() {
        String flux = String.format(
                "from(bucket: \"%s\") " +
                        "|> range(start: -%dm) " +
                        "|> filter(fn: (r) => r._measurement == \"opc_metrics\")",
                bucket, TIME_WINDOW_MINUTES
        );

        List<FluxTable> tables = client.getQueryApi().query(flux, org);
        return parseResults(tables);
    }

    /**
     * Parse Flux query results to ProcessedData
     */
    private List<ProcessedData> parseResults(List<FluxTable> tables) {
        Map<String, ProcessedData.ProcessedDataBuilder> dataMap = new HashMap<>();

        for (FluxTable table : tables) {
            for (FluxRecord record : table.getRecords()) {
                String deviceId = (String) record.getValueByKey("device_id");
                if (deviceId == null) continue;

                Instant timestamp = record.getTime();
                if (timestamp == null) continue;

                String key = deviceId + "_" + timestamp.toEpochMilli();

                ProcessedData.ProcessedDataBuilder builder = dataMap.computeIfAbsent(key, k ->
                        ProcessedData.builder()
                                .deviceId(deviceId)
                                .deviceIp((String) record.getValueByKey("device_ip"))
                                .timestamp(timestamp)
                                .measurement("opc_metrics")
                                .tags(new HashMap<>())
                                .fields(new HashMap<>())
                );

                // Add field
                String field = record.getField();
                Object value = record.getValue();
                if (field != null && value != null) {
                    builder.build().addField(field, value);
                }

                // Add tags
                record.getValues().forEach((tagKey, tagValue) -> {
                    if (tagKey.startsWith("_") || tagKey.equals("result") || tagKey.equals("table")) {
                        return;
                    }
                    if (tagValue != null && !tagKey.equals("device_id") && !tagKey.equals("device_ip")) {
                        builder.build().addTag(tagKey, tagValue.toString());
                    }
                });
            }
        }

        return dataMap.values().stream()
                .map(ProcessedData.ProcessedDataBuilder::build)
                .toList();
    }

    @Override
    public void cancel() {
        log.info("Cancelling InfluxDB source");
        running = false;
    }

    @Override
    public void close() {
        log.info("Closing InfluxDB connection");
        if (client != null) {
            client.close();
        }
    }
}
