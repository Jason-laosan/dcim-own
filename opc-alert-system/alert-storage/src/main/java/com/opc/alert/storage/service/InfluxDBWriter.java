package com.opc.alert.storage.service;

import com.influxdb.client.InfluxDBClient;
import com.influxdb.client.WriteApiBlocking;
import com.influxdb.client.domain.WritePrecision;
import com.influxdb.client.write.Point;
import com.opc.alert.common.model.ProcessedData;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

/**
 * InfluxDB Writer Service
 */
@Slf4j
@Service
public class InfluxDBWriter {

    private final InfluxDBClient influxDBClient;
    private final String bucket;
    private final String org;

    public InfluxDBWriter(InfluxDBClient influxDBClient,
                          @Value("${influxdb.bucket:opc_data}") String bucket,
                          @Value("${influxdb.org:opc_organization}") String org) {
        this.influxDBClient = influxDBClient;
        this.bucket = bucket;
        this.org = org;
    }

    /**
     * Write single data point
     */
    public void write(ProcessedData data) {
        try {
            Point point = convertToPoint(data);
            WriteApiBlocking writeApi = influxDBClient.getWriteApiBlocking();
            writeApi.writePoint(bucket, org, point);

            log.debug("Successfully wrote data for device: {}", data.getDeviceId());
        } catch (Exception e) {
            log.error("Failed to write data to InfluxDB for device: {}", data.getDeviceId(), e);
            throw new RuntimeException("Failed to write to InfluxDB", e);
        }
    }

    /**
     * Write batch data points
     */
    public void writeBatch(List<ProcessedData> dataList) {
        if (dataList == null || dataList.isEmpty()) {
            return;
        }

        try {
            WriteApiBlocking writeApi = influxDBClient.getWriteApiBlocking();
            List<Point> points = dataList.stream()
                    .map(this::convertToPoint)
                    .toList();

            writeApi.writePoints(bucket, org, points);

            log.info("Successfully wrote {} data points to InfluxDB", dataList.size());
        } catch (Exception e) {
            log.error("Failed to write batch data to InfluxDB, size: {}", dataList.size(), e);
            throw new RuntimeException("Failed to write batch to InfluxDB", e);
        }
    }

    /**
     * Convert ProcessedData to InfluxDB Point
     */
    private Point convertToPoint(ProcessedData data) {
        String measurement = data.getMeasurement() != null ?
                data.getMeasurement() : "opc_metrics";

        Point point = Point.measurement(measurement)
                .time(data.getTimestamp(), WritePrecision.MS);

        // Add tags
        point.addTag("device_id", data.getDeviceId());
        if (data.getDeviceIp() != null) {
            point.addTag("device_ip", data.getDeviceIp());
        }
        if (data.getQuality() != null) {
            point.addTag("quality", data.getQuality());
        }

        // Add custom tags
        if (data.getTags() != null) {
            for (Map.Entry<String, String> entry : data.getTags().entrySet()) {
                point.addTag(entry.getKey(), entry.getValue());
            }
        }

        // Add fields
        if (data.getFields() != null) {
            for (Map.Entry<String, Object> entry : data.getFields().entrySet()) {
                addField(point, entry.getKey(), entry.getValue());
            }
        }

        return point;
    }

    /**
     * Add field to point based on value type
     */
    private void addField(Point point, String name, Object value) {
        if (value == null) {
            return;
        }

        if (value instanceof Number) {
            if (value instanceof Double || value instanceof Float) {
                point.addField(name, ((Number) value).doubleValue());
            } else if (value instanceof Long) {
                point.addField(name, ((Number) value).longValue());
            } else {
                point.addField(name, ((Number) value).longValue());
            }
        } else if (value instanceof Boolean) {
            point.addField(name, (Boolean) value);
        } else {
            point.addField(name, value.toString());
        }
    }

    /**
     * Health check
     */
    public boolean isHealthy() {
        try {
            return influxDBClient.ping();
        } catch (Exception e) {
            log.error("InfluxDB health check failed", e);
            return false;
        }
    }
}
