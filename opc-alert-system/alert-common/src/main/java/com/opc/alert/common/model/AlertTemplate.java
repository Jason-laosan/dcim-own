package com.opc.alert.common.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Alert Template Configuration
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AlertTemplate {

    /**
     * Template ID
     */
    private Long id;

    /**
     * Template name
     */
    private String templateName;

    /**
     * Alert title template
     * Supports variables: ${deviceId}, ${metricName}, ${value}, ${threshold}, ${level}
     */
    private String titleTemplate;

    /**
     * Alert message template
     * Supports variables: ${deviceId}, ${deviceIp}, ${metricName}, ${value}, ${threshold}, ${level}, ${timestamp}
     */
    private String messageTemplate;

    /**
     * Notification channels: EMAIL, SMS, WEBHOOK
     */
    private String channels;

    /**
     * Is template enabled
     */
    private Boolean enabled;

    /**
     * Created at
     */
    private Instant createdAt;

    /**
     * Updated at
     */
    private Instant updatedAt;

    private static final Pattern VARIABLE_PATTERN = Pattern.compile("\\$\\{([^}]+)}");

    /**
     * Render title with variables
     */
    public String renderTitle(Map<String, Object> variables) {
        return renderTemplate(titleTemplate, variables);
    }

    /**
     * Render message with variables
     */
    public String renderMessage(Map<String, Object> variables) {
        return renderTemplate(messageTemplate, variables);
    }

    /**
     * Render template with variables
     */
    private String renderTemplate(String template, Map<String, Object> variables) {
        if (template == null || template.trim().isEmpty()) {
            return "";
        }

        StringBuffer result = new StringBuffer();
        Matcher matcher = VARIABLE_PATTERN.matcher(template);

        while (matcher.find()) {
            String varName = matcher.group(1);
            Object value = variables.get(varName);
            String replacement = value != null ? value.toString() : "";
            matcher.appendReplacement(result, Matcher.quoteReplacement(replacement));
        }
        matcher.appendTail(result);

        return result.toString();
    }
}
