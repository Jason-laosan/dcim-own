package com.opc.alert.config.service;

import com.opc.alert.common.model.AlertReceiver;
import com.opc.alert.common.model.AlertRule;
import com.opc.alert.common.model.AlertTemplate;
import com.opc.alert.config.repository.AlertReceiverRepository;
import com.opc.alert.config.repository.AlertRuleRepository;
import com.opc.alert.config.repository.AlertTemplateRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Alert Configuration Service
 * Manages alert rules, templates, and receivers
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class AlertConfigService {

    private final AlertRuleRepository ruleRepository;
    private final AlertTemplateRepository templateRepository;
    private final AlertReceiverRepository receiverRepository;

    // ====== Alert Rules ======

    /**
     * Get all enabled alert rules (cached)
     */
    @Cacheable("alertRules")
    public List<AlertRule> getAllEnabledRules() {
        return ruleRepository.findAllEnabled();
    }

    /**
     * Get rules by metric name
     */
    public List<AlertRule> getRulesByMetricName(String metricName) {
        return ruleRepository.findByMetricName(metricName);
    }

    /**
     * Refresh alert rules cache periodically
     */
    @Scheduled(fixedDelayString = "${alert.config.refresh-interval:60000}")
    @CacheEvict(value = "alertRules", allEntries = true)
    public void refreshRulesCache() {
        log.debug("Refreshing alert rules cache");
    }

    // ====== Alert Templates ======

    /**
     * Get template by ID (cached)
     */
    @Cacheable("alertTemplates")
    public AlertTemplate getTemplateById(Long id) {
        return templateRepository.findById(id);
    }

    /**
     * Get all enabled templates
     */
    public List<AlertTemplate> getAllEnabledTemplates() {
        return templateRepository.findAllEnabled();
    }

    /**
     * Refresh templates cache
     */
    @CacheEvict(value = "alertTemplates", allEntries = true)
    public void refreshTemplatesCache() {
        log.debug("Refreshing alert templates cache");
    }

    // ====== Alert Receivers ======

    /**
     * Get all enabled receivers (cached)
     */
    @Cacheable("alertReceivers")
    public List<AlertReceiver> getAllEnabledReceivers() {
        return receiverRepository.findAllEnabled();
    }

    /**
     * Get receivers for specific alert level
     */
    public List<AlertReceiver> getReceiversForLevel(String level) {
        List<AlertReceiver> allReceivers = getAllEnabledReceivers();
        return allReceivers.stream()
                .filter(receiver -> receiver.shouldReceiveLevel(level))
                .toList();
    }

    /**
     * Refresh receivers cache
     */
    @CacheEvict(value = "alertReceivers", allEntries = true)
    public void refreshReceiversCache() {
        log.debug("Refreshing alert receivers cache");
    }

    // ====== Management Methods ======

    /**
     * Refresh all caches
     */
    @Scheduled(fixedDelayString = "${alert.config.refresh-interval:60000}")
    public void refreshAllCaches() {
        refreshRulesCache();
        refreshTemplatesCache();
        refreshReceiversCache();
        log.info("All alert configuration caches refreshed");
    }
}
