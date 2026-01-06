针对你**500万物联网节点 + 10分钟滑动窗口**的场景，我为你整理了一份**专属的 RocksDB + Flink Grafana 仪表盘模板**。这份模板聚焦于**状态读写性能、检查点效率、合并（Compaction）压力、磁盘 IO**四大核心维度，完美适配大规模滑动窗口的监控需求。

### 一、 仪表盘核心功能
1.  **RocksDB 状态总览**：状态大小、Block Cache 命中率、读写 QPS
2.  **检查点（Checkpoint）监控**：检查点持续时间、成功率、本地/远程传输速度
3.  **合并（Compaction）监控**：待合并任务数、合并持续时间、CPU 占用
4.  **磁盘 IO 监控**：本地磁盘读写速度、RocksDB 刷盘（Flush）压力
5.  **异常告警**：Block Cache 命中率过低、检查点超时、合并任务积压

### 二、 Grafana 仪表盘 JSON 模板
> 适用版本：Grafana 8.0+、Prometheus 2.0+、Flink 1.17+
> 导入方式：Grafana → Create → Import → 粘贴以下 JSON → 选择 Prometheus 数据源

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": {
          "type": "grafana",
          "uid": "-- Grafana --"
        },
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": 1001,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "collapsed": false,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 26,
      "panels": [],
      "title": "1. RocksDB 状态总览（500万节点滑动窗口核心）",
      "type": "row"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "red",
                "value": 80
              }
            ]
          },
          "unit": "bytes"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 1
      },
      "id": 1,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "sum(flink_taskmanager_rocksdb_state_size{job_name=~\"$job_name\"}) by (taskmanager_id)",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "RocksDB 状态总大小（按 TaskManager）",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "命中率",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "red",
                "value": null
              },
              {
                "color": "orange",
                "value": 0.8
              },
              {
                "color": "green",
                "value": 0.9
              }
            ]
          },
          "unit": "percentunit"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 1
      },
      "id": 2,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "avg(flink_taskmanager_rocksdb_block_cache_hit_ratio{job_name=~\"$job_name\"}) by (taskmanager_id)",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "Block Cache 命中率（≥90% 为优）",
      "type": "timeseries"
    },
    {
      "collapsed": false,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 9
      },
      "id": 27,
      "panels": [],
      "title": "2. 检查点（Checkpoint）监控（滑动窗口容错核心）",
      "type": "row"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "持续时间（s）",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "orange",
                "value": 30
              },
              {
                "color": "red",
                "value": 60
              }
            ]
          },
          "unit": "s"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 10
      },
      "id": 3,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "flink_jobmanager_checkpoint_duration{job_name=~\"$job_name\", checkpoint_type=\"COMPLETED\"}",
          "interval": "",
          "legendFormat": "检查点持续时间",
          "refId": "A"
        }
      ],
      "title": "检查点持续时间（≤30s 为优）",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "成功率",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "red",
                "value": null
              },
              {
                "color": "green",
                "value": 1
              }
            ]
          },
          "unit": "percentunit"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 10
      },
      "id": 4,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "sum(flink_jobmanager_checkpoint_completed{job_name=~\"$job_name\"}) / sum(flink_jobmanager_checkpoint_started{job_name=~\"$job_name\"})",
          "interval": "",
          "legendFormat": "检查点成功率",
          "refId": "A"
        }
      ],
      "title": "检查点成功率（100% 为优）",
      "type": "timeseries"
    },
    {
      "collapsed": false,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 18
      },
      "id": 28,
      "panels": [],
      "title": "3. 合并（Compaction）监控（避免读写阻塞核心）",
      "type": "row"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "任务数",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "orange",
                "value": 5
              },
              {
                "color": "red",
                "value": 10
              }
            ]
          }
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 19
      },
      "id": 5,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "avg(flink_taskmanager_rocksdb_compaction_pending{job_name=~\"$job_name\"}) by (taskmanager_id)",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "待合并任务数（≤5 为优）",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "持续时间（ms）",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "orange",
                "value": 1000
              },
              {
                "color": "red",
                "value": 5000
              }
            ]
          },
          "unit": "ms"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 19
      },
      "id": 6,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "avg(flink_taskmanager_rocksdb_compaction_duration{job_name=~\"$job_name\"}) by (taskmanager_id)",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "合并持续时间（≤1s 为优）",
      "type": "timeseries"
    },
    {
      "collapsed": false,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 27
      },
      "id": 29,
      "panels": [],
      "title": "4. 磁盘 IO 监控（SSD 性能验证）",
      "type": "row"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "速度（MB/s）",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "orange",
                "value": 200
              },
              {
                "color": "red",
                "value": 300
              }
            ]
          },
          "unit": "MB/s"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 28
      },
      "id": 7,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "sum(flink_taskmanager_rocksdb_disk_write_rate{job_name=~\"$job_name\"}) by (taskmanager_id) / 1024 / 1024",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "RocksDB 磁盘写入速度",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "速度（MB/s）",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "orange",
                "value": 200
              },
              {
                "color": "red",
                "value": 300
              }
            ]
          },
          "unit": "MB/s"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 28
      },
      "id": 8,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "expr": "sum(flink_taskmanager_rocksdb_disk_read_rate{job_name=~\"$job_name\"}) by (taskmanager_id) / 1024 / 1024",
          "interval": "",
          "legendFormat": "{{taskmanager_id}}",
          "refId": "A"
        }
      ],
      "title": "RocksDB 磁盘读取速度",
      "type": "timeseries"
    }
  ],
  "refresh": "10s",
  "schemaVersion": 38,
  "style": "dark",
  "tags": [
    "flink",
    "rocksdb",
    "iot",
    "sliding-window"
  ],
  "templating": {
    "list": [
      {
        "current": {
          "selected": false,
          "text": "All",
          "value": "$__all"
        },
        "datasource": {
          "type": "prometheus",
          "uid": "prometheus"
        },
        "definition": "label_values(flink_job_name)",
        "description": "选择 Flink 作业名称",
        "error": null,
        "hide": 0,
        "includeAll": true,
        "label": "Job Name",
        "multi": false,
        "name": "job_name",
        "options": [],
        "query": {
          "query": "label_values(flink_job_name)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 1,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }
    ]
  },
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ]
  },
  "timezone": "",
  "title": "Flink RocksDB 物联网滑动窗口监控仪表盘",
  "uid": "flink-rocksdb-iot",
  "version": 1,
  "weekStart": 0
}
```

### 三、 模板使用前置条件
1.  **Flink 已启用 Prometheus 监控**（已在 `flink-conf.yaml` 中配置）
    ```yaml
    metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
    metrics.reporter.prom.port: 9250-9260
    ```
2.  **Prometheus 已配置 Flink 抓取任务**
    ```yaml
    scrape_configs:
      - job_name: 'flink'
        static_configs:
          - targets: ['jobmanager:8081', 'taskmanager1:9250', 'taskmanager2:9250']
        scrape_interval: 10s
    ```
3.  **Grafana 已添加 Prometheus 数据源**（UID 为 `prometheus`，与模板中一致）

### 四、 关键告警规则配置（需手动添加到 Prometheus）
在 `prometheus.yml` 中添加以下告警规则，针对大规模滑动窗口的核心异常场景：
```yaml
rule_files:
  - "alert_rules.yml"
```
创建 `alert_rules.yml`：
```yaml
groups:
  - name: flink_rocksdb_iot_alerts
    rules:
      # 1. Block Cache 命中率过低（<90% 持续 5 分钟）
      - alert: RocksDBBlockCacheHitRatioLow
        expr: avg(flink_taskmanager_rocksdb_block_cache_hit_ratio{job_name=~"iot.*"}) by (job_name) < 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "RocksDB Block Cache 命中率过低"
          description: "作业 {{ $labels.job_name }} 的 Block Cache 命中率为 {{ $value | humanizePercentage }}，低于 90%，需增大 Block Cache 大小。"
      # 2. 检查点超时（持续时间 > 60s 持续 3 次）
      - alert: FlinkCheckpointTimeout
        expr: flink_jobmanager_checkpoint_duration{job_name=~"iot.*", checkpoint_type="COMPLETED"} > 60
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "Flink 检查点超时"
          description: "作业 {{ $labels.job_name }} 的检查点持续时间超过 60s，需启用本地检查点或优化 RocksDB 配置。"
      # 3. 待合并任务数过多（>10 持续 5 分钟）
      - alert: RocksDBCompactionPendingHigh
        expr: avg(flink_taskmanager_rocksdb_compaction_pending{job_name=~"iot.*"}) by (taskmanager_id) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "RocksDB 待合并任务数过多"
          description: "TaskManager {{ $labels.taskmanager_id }} 的待合并任务数为 {{ $value }}，超过 10，需增加合并线程数。"
```

### 五、 仪表盘使用建议
1.  **时间范围选择**：默认 `now-1h`，故障排查时可切换到 `now-10m` 或 `now-30m`。
2.  **作业筛选**：通过顶部的 `Job Name` 下拉框，筛选你关注的物联网滑动窗口作业。
3.  **告警阈值调整**：根据你的集群硬件配置（如 SSD 性能、CPU 核数），适当调整面板中的阈值（如磁盘写入速度、合并持续时间）。
4.  **日常监控重点**：
    - 每小时查看 **Block Cache 命中率**（确保 ≥90%）。
    - 每天查看 **状态总大小**（确保不超过磁盘容量）。
    - 实时监控 **检查点成功率**（确保 100%）。
