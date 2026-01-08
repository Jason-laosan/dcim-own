package com.opc.alert.engine;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.scheduling.annotation.EnableScheduling;

/**
 * OPC Alert Engine Application
 */
@SpringBootApplication
@EnableCaching
@EnableScheduling
@ComponentScan(basePackages = {
        "com.opc.alert.engine",
        "com.opc.alert.config"
})
public class AlertEngineApplication {

    public static void main(String[] args) {
        SpringApplication.run(AlertEngineApplication.class, args);
    }
}
