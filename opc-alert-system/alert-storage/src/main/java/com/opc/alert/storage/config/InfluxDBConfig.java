package com.opc.alert.storage.config;

import com.influxdb.client.InfluxDBClient;
import com.influxdb.client.InfluxDBClientFactory;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * InfluxDB Configuration
 */
@Slf4j
@Configuration
public class InfluxDBConfig {

    @Value("${influxdb.url:http://localhost:8086}")
    private String url;

    @Value("${influxdb.token}")
    private String token;

    @Value("${influxdb.org:opc_organization}")
    private String org;

    @Value("${influxdb.bucket:opc_data}")
    private String bucket;

    @Bean
    public InfluxDBClient influxDBClient() {
        log.info("Initializing InfluxDB client: url={}, org={}, bucket={}", url, org, bucket);
        return InfluxDBClientFactory.create(url, token.toCharArray(), org, bucket);
    }

    @Bean
    public String influxDBOrg() {
        return org;
    }

    @Bean
    public String influxDBBucket() {
        return bucket;
    }
}
