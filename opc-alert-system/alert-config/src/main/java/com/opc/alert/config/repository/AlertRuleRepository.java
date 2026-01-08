package com.opc.alert.config.repository;

import com.opc.alert.common.model.AlertRule;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.time.Instant;
import java.util.List;

/**
 * Alert Rule Repository
 */
@Slf4j
@Repository
@RequiredArgsConstructor
public class AlertRuleRepository {

    private final JdbcTemplate jdbcTemplate;

    private static final RowMapper<AlertRule> ROW_MAPPER = (rs, rowNum) -> AlertRule.builder()
            .id(rs.getLong("id"))
            .ruleName(rs.getString("rule_name"))
            .description(rs.getString("description"))
            .metricName(rs.getString("metric_name"))
            .conditionType(rs.getString("condition_type"))
            .threshold(rs.getDouble("threshold"))
            .level(rs.getString("level"))
            .timeWindowSeconds(rs.getInt("time_window_seconds"))
            .consecutiveCount(rs.getInt("consecutive_count"))
            .deviceFilter(rs.getString("device_filter"))
            .tagFilters(rs.getString("tag_filters"))
            .templateId(rs.getLong("template_id"))
            .enabled(rs.getBoolean("enabled"))
            .createdAt(toInstant(rs.getTimestamp("created_at")))
            .updatedAt(toInstant(rs.getTimestamp("updated_at")))
            .build();

    /**
     * Find all enabled alert rules
     */
    public List<AlertRule> findAllEnabled() {
        String sql = "SELECT * FROM alert_rules WHERE enabled = true ORDER BY id";
        return jdbcTemplate.query(sql, ROW_MAPPER);
    }

    /**
     * Find rule by ID
     */
    public AlertRule findById(Long id) {
        String sql = "SELECT * FROM alert_rules WHERE id = ?";
        List<AlertRule> rules = jdbcTemplate.query(sql, ROW_MAPPER, id);
        return rules.isEmpty() ? null : rules.get(0);
    }

    /**
     * Find rules by metric name
     */
    public List<AlertRule> findByMetricName(String metricName) {
        String sql = "SELECT * FROM alert_rules WHERE metric_name = ? AND enabled = true";
        return jdbcTemplate.query(sql, ROW_MAPPER, metricName);
    }

    /**
     * Insert new rule
     */
    public long insert(AlertRule rule) {
        String sql = "INSERT INTO alert_rules " +
                "(rule_name, description, metric_name, condition_type, threshold, level, " +
                "time_window_seconds, consecutive_count, device_filter, tag_filters, template_id, enabled) " +
                "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id";

        return jdbcTemplate.queryForObject(sql, Long.class,
                rule.getRuleName(),
                rule.getDescription(),
                rule.getMetricName(),
                rule.getConditionType(),
                rule.getThreshold(),
                rule.getLevel(),
                rule.getTimeWindowSeconds(),
                rule.getConsecutiveCount(),
                rule.getDeviceFilter(),
                rule.getTagFilters(),
                rule.getTemplateId(),
                rule.getEnabled()
        );
    }

    /**
     * Update rule
     */
    public void update(AlertRule rule) {
        String sql = "UPDATE alert_rules SET " +
                "rule_name = ?, description = ?, metric_name = ?, condition_type = ?, " +
                "threshold = ?, level = ?, time_window_seconds = ?, consecutive_count = ?, " +
                "device_filter = ?, tag_filters = ?, template_id = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP " +
                "WHERE id = ?";

        jdbcTemplate.update(sql,
                rule.getRuleName(),
                rule.getDescription(),
                rule.getMetricName(),
                rule.getConditionType(),
                rule.getThreshold(),
                rule.getLevel(),
                rule.getTimeWindowSeconds(),
                rule.getConsecutiveCount(),
                rule.getDeviceFilter(),
                rule.getTagFilters(),
                rule.getTemplateId(),
                rule.getEnabled(),
                rule.getId()
        );
    }

    /**
     * Delete rule
     */
    public void delete(Long id) {
        String sql = "DELETE FROM alert_rules WHERE id = ?";
        jdbcTemplate.update(sql, id);
    }

    private static Instant toInstant(Timestamp timestamp) {
        return timestamp != null ? timestamp.toInstant() : null;
    }
}
