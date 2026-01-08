package com.opc.alert.config.repository;

import com.opc.alert.common.model.AlertTemplate;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.time.Instant;
import java.util.List;

/**
 * Alert Template Repository
 */
@Slf4j
@Repository
@RequiredArgsConstructor
public class AlertTemplateRepository {

    private final JdbcTemplate jdbcTemplate;

    private static final RowMapper<AlertTemplate> ROW_MAPPER = (rs, rowNum) -> AlertTemplate.builder()
            .id(rs.getLong("id"))
            .templateName(rs.getString("template_name"))
            .titleTemplate(rs.getString("title_template"))
            .messageTemplate(rs.getString("message_template"))
            .channels(rs.getString("channels"))
            .enabled(rs.getBoolean("enabled"))
            .createdAt(toInstant(rs.getTimestamp("created_at")))
            .updatedAt(toInstant(rs.getTimestamp("updated_at")))
            .build();

    /**
     * Find template by ID
     */
    public AlertTemplate findById(Long id) {
        String sql = "SELECT * FROM alert_templates WHERE id = ?";
        List<AlertTemplate> templates = jdbcTemplate.query(sql, ROW_MAPPER, id);
        return templates.isEmpty() ? null : templates.get(0);
    }

    /**
     * Find all enabled templates
     */
    public List<AlertTemplate> findAllEnabled() {
        String sql = "SELECT * FROM alert_templates WHERE enabled = true";
        return jdbcTemplate.query(sql, ROW_MAPPER);
    }

    /**
     * Insert new template
     */
    public long insert(AlertTemplate template) {
        String sql = "INSERT INTO alert_templates " +
                "(template_name, title_template, message_template, channels, enabled) " +
                "VALUES (?, ?, ?, ?, ?) RETURNING id";

        return jdbcTemplate.queryForObject(sql, Long.class,
                template.getTemplateName(),
                template.getTitleTemplate(),
                template.getMessageTemplate(),
                template.getChannels(),
                template.getEnabled()
        );
    }

    /**
     * Update template
     */
    public void update(AlertTemplate template) {
        String sql = "UPDATE alert_templates SET " +
                "template_name = ?, title_template = ?, message_template = ?, " +
                "channels = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP " +
                "WHERE id = ?";

        jdbcTemplate.update(sql,
                template.getTemplateName(),
                template.getTitleTemplate(),
                template.getMessageTemplate(),
                template.getChannels(),
                template.getEnabled(),
                template.getId()
        );
    }

    private static Instant toInstant(Timestamp timestamp) {
        return timestamp != null ? timestamp.toInstant() : null;
    }
}
