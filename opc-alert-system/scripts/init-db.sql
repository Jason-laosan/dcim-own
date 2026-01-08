-- PostgreSQL Database Initialization Script
-- OPC Alert System

-- ============================================
-- Create Database (if needed)
-- ============================================
-- CREATE DATABASE opc_alert_db;
-- \c opc_alert_db;

-- ============================================
-- Create Tables
-- ============================================

-- Alert Rules Table
CREATE TABLE IF NOT EXISTS alert_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    metric_name VARCHAR(255) NOT NULL,
    condition_type VARCHAR(10) NOT NULL CHECK (condition_type IN ('>', '<', '>=', '<=', '==', '!=')),
    threshold DOUBLE PRECISION NOT NULL,
    level VARCHAR(20) NOT NULL CHECK (level IN ('INFO', 'WARNING', 'ERROR', 'CRITICAL')),
    time_window_seconds INTEGER NOT NULL DEFAULT 600,
    consecutive_count INTEGER NOT NULL DEFAULT 1,
    device_filter VARCHAR(255),
    tag_filters TEXT,
    template_id BIGINT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alert Templates Table
CREATE TABLE IF NOT EXISTS alert_templates (
    id BIGSERIAL PRIMARY KEY,
    template_name VARCHAR(255) NOT NULL UNIQUE,
    title_template TEXT NOT NULL,
    message_template TEXT NOT NULL,
    channels VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alert Receivers Table
CREATE TABLE IF NOT EXISTS alert_receivers (
    id BIGSERIAL PRIMARY KEY,
    receiver_name VARCHAR(255) NOT NULL UNIQUE,
    receiver_type VARCHAR(50) NOT NULL CHECK (receiver_type IN ('EMAIL', 'SMS', 'WEBHOOK')),
    contact VARCHAR(500) NOT NULL,
    level_filter VARCHAR(100),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alert History Table (optional, for tracking sent alerts)
CREATE TABLE IF NOT EXISTS alert_history (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(100) NOT NULL UNIQUE,
    rule_id BIGINT NOT NULL,
    rule_name VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_ip VARCHAR(50),
    metric_name VARCHAR(255) NOT NULL,
    current_value DOUBLE PRECISION NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    level VARCHAR(20) NOT NULL,
    title TEXT,
    message TEXT,
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'TRIGGERED',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- Create Indexes
-- ============================================

CREATE INDEX idx_alert_rules_metric_name ON alert_rules(metric_name);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX idx_alert_history_device_id ON alert_history(device_id);
CREATE INDEX idx_alert_history_triggered_at ON alert_history(triggered_at);
CREATE INDEX idx_alert_history_rule_id ON alert_history(rule_id);

-- ============================================
-- Insert Sample Data
-- ============================================

-- Sample Alert Template
INSERT INTO alert_templates (template_name, title_template, message_template, channels, enabled)
VALUES (
    'Default Alert Template',
    '[${level}] Alert: ${metricName} on ${deviceId}',
    'Device: ${deviceId} (${deviceIp})
Metric: ${metricName}
Current Value: ${value}
Threshold: ${threshold}
Alert Level: ${level}
Timestamp: ${timestamp}',
    'EMAIL,WEBHOOK',
    true
) ON CONFLICT (template_name) DO NOTHING;

-- Sample Alert Rule: High Temperature
INSERT INTO alert_rules (
    rule_name, description, metric_name, condition_type, threshold,
    level, time_window_seconds, consecutive_count, template_id, enabled
)
VALUES (
    'High Temperature Alert',
    'Trigger alert when temperature exceeds 80 degrees',
    'temperature',
    '>',
    80.0,
    'WARNING',
    600,
    3,
    1,
    true
) ON CONFLICT (rule_name) DO NOTHING;

-- Sample Alert Rule: Critical Pressure
INSERT INTO alert_rules (
    rule_name, description, metric_name, condition_type, threshold,
    level, time_window_seconds, consecutive_count, template_id, enabled
)
VALUES (
    'Critical Pressure Alert',
    'Trigger alert when pressure exceeds critical threshold',
    'pressure',
    '>',
    100.0,
    'CRITICAL',
    300,
    2,
    1,
    true
) ON CONFLICT (rule_name) DO NOTHING;

-- Sample Alert Receiver: Email
INSERT INTO alert_receivers (receiver_name, receiver_type, contact, level_filter, enabled)
VALUES (
    'Operations Team Email',
    'EMAIL',
    'ops-team@example.com',
    'WARNING,ERROR,CRITICAL',
    true
) ON CONFLICT (receiver_name) DO NOTHING;

-- Sample Alert Receiver: Webhook
INSERT INTO alert_receivers (receiver_name, receiver_type, contact, level_filter, enabled)
VALUES (
    'Alert Webhook',
    'WEBHOOK',
    'https://your-webhook-endpoint.com/alerts',
    'ERROR,CRITICAL',
    true
) ON CONFLICT (receiver_name) DO NOTHING;

-- ============================================
-- Create Functions and Triggers
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
DROP TRIGGER IF EXISTS update_alert_rules_updated_at ON alert_rules;
CREATE TRIGGER update_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_templates_updated_at ON alert_templates;
CREATE TRIGGER update_alert_templates_updated_at
    BEFORE UPDATE ON alert_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_receivers_updated_at ON alert_receivers;
CREATE TRIGGER update_alert_receivers_updated_at
    BEFORE UPDATE ON alert_receivers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Grant Permissions (adjust user as needed)
-- ============================================

-- Example: GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO opc_alert_user;
-- Example: GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO opc_alert_user;

COMMENT ON TABLE alert_rules IS 'Alert rule configurations';
COMMENT ON TABLE alert_templates IS 'Alert message templates';
COMMENT ON TABLE alert_receivers IS 'Alert notification receivers';
COMMENT ON TABLE alert_history IS 'Historical alert events';
