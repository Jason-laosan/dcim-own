package com.opc.alert.consumer;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;

/**
 * OPC Alert Consumer Application
 */
@SpringBootApplication
@ComponentScan(basePackages = {
        "com.opc.alert.consumer",
        "com.opc.alert.processor",
        "com.opc.alert.storage"
})
public class AlertConsumerApplication {

    public static void main(String[] args) {
        SpringApplication.run(AlertConsumerApplication.class, args);
    }
}
