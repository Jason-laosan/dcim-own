package com.opc.alert.config.repository;

import com.opc.alert.common.model.AlertReceiver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.time.Instant;
import java.util.List;

/**
 * Alert Receiver Repository
 */
@Slf4j
@Repository
@RequiredArgsConstructor
public class AlertReceiverRepository {

    private final JdbcTemplate jdbcTemplate;

    private static final RowMapper<AlertReceiver> ROW_MAPPER = (rs, rowNum) -> AlertReceiver.builder()
            .id(rs.getLong("id"))
            .receiverName(rs.getString("receiver_name"))
            .receiverType(rs.getString("receiver_type"))
            .contact(rs.getString("contact"))
            .levelFilter(rs.getString("level_filter"))
            .enabled(rs.getBoolean("enabled"))
            .createdAt(toInstant(rs.getTimestamp("created_at")))
            .updatedAt(toInstant(rs.getTimestamp("updated_at")))
            .build();

    /**
     * Find all enabled receivers
     */
    public List<AlertReceiver> findAllEnabled() {
        String sql = "SELECT * FROM alert_receivers WHERE enabled = true";
        return jdbcTemplate.query(sql, ROW_MAPPER);
    }

    /**
     * Find receivers by level
     */
    public List<AlertReceiver> findByLevel(String level) {
        String sql = "SELECT * FROM alert_receivers WHERE enabled = true " +
                "AND (level_filter IS NULL OR level_filter = '' OR level_filter LIKE ?)";
        return jdbcTemplate.query(sql, ROW_MAPPER, "%" + level + "%");
    }

    /**
     * Find receiver by ID
     */
    public AlertReceiver findById(Long id) {
        String sql = "SELECT * FROM alert_receivers WHERE id = ?";
        List<AlertReceiver> receivers = jdbcTemplate.query(sql, ROW_MAPPER, id);
        return receivers.isEmpty() ? null : receivers.get(0);
    }

    /**
     * Insert new receiver
     */
    public long insert(AlertReceiver receiver) {
        String sql = "INSERT INTO alert_receivers " +
                "(receiver_name, receiver_type, contact, level_filter, enabled) " +
                "VALUES (?, ?, ?, ?, ?) RETURNING id";

        return jdbcTemplate.queryForObject(sql, Long.class,
                receiver.getReceiverName(),
                receiver.getReceiverType(),
                receiver.getContact(),
                receiver.getLevelFilter(),
                receiver.getEnabled()
        );
    }

    /**
     * Update receiver
     */
    public void update(AlertReceiver receiver) {
        String sql = "UPDATE alert_receivers SET " +
                "receiver_name = ?, receiver_type = ?, contact = ?, " +
                "level_filter = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP " +
                "WHERE id = ?";

        jdbcTemplate.update(sql,
                receiver.getReceiverName(),
                receiver.getReceiverType(),
                receiver.getContact(),
                receiver.getLevelFilter(),
                receiver.getEnabled(),
                receiver.getId()
        );
    }

    private static Instant toInstant(Timestamp timestamp) {
        return timestamp != null ? timestamp.toInstant() : null;
    }
}
