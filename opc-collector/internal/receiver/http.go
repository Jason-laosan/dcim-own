package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"

	"go.uber.org/zap"
)

// HTTPReceiver HTTP接收器
// 提供REST API接口接收下游主动推送的数据
type HTTPReceiver struct {
	config       *config.HTTPReceiverConfig
	dataHandler  DataHandler
	server       *http.Server
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	logger       *zap.Logger

	// 统计指标
	receivedCount uint64
	errorCount    uint64
}

// NewHTTPReceiver 创建HTTP接收器实例
func NewHTTPReceiver(cfg *config.HTTPReceiverConfig, handler DataHandler) (*HTTPReceiver, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &HTTPReceiver{
		config:      cfg,
		dataHandler: handler,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger.Log,
	}

	return r, nil
}

// Start 启动HTTP接收器
func (r *HTTPReceiver) Start() error {
	mux := http.NewServeMux()

	// 注册数据接收端点
	mux.HandleFunc(r.config.Endpoint, r.handleMetricData)

	// 批量数据接收端点
	mux.HandleFunc(r.config.Endpoint+"/batch", r.handleBatchMetricData)

	// 健康检查端点
	mux.HandleFunc("/health", r.handleHealth)

	// 统计端点
	mux.HandleFunc("/stats", r.handleStats)

	r.server = &http.Server{
		Addr:         r.config.ListenAddr,
		Handler:      r.wrapWithAuth(mux),
		ReadTimeout:  time.Duration(r.config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(r.config.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动HTTP服务器
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		r.logger.Info("HTTP receiver listening",
			zap.String("address", r.config.ListenAddr),
			zap.String("endpoint", r.config.Endpoint))

		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.logger.Error("HTTP receiver server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止HTTP接收器
func (r *HTTPReceiver) Stop() {
	r.logger.Info("stopping HTTP receiver")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.server.Shutdown(ctx); err != nil {
		r.logger.Error("failed to shutdown HTTP server gracefully", zap.Error(err))
	}

	r.cancel()
	r.wg.Wait()

	r.logger.Info("HTTP receiver stopped")
}

// Name 返回接收器名称
func (r *HTTPReceiver) Name() string {
	return "http"
}

// handleMetricData 处理单条数据推送
func (r *HTTPReceiver) handleMetricData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("failed to read request body", zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// 解析JSON数据
	var data models.MetricData
	if err := json.Unmarshal(body, &data); err != nil {
		r.logger.Error("failed to unmarshal metric data",
			zap.Error(err),
			zap.String("body", string(body)))
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证必要字段
	if data.DeviceID == "" {
		r.logger.Warn("received data without device_id")
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Missing device_id", http.StatusBadRequest)
		return
	}

	// 如果时间戳为空，使用当前时间
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	r.logger.Debug("received metric data",
		zap.String("device_id", data.DeviceID),
		zap.Int("metric_count", len(data.Metrics)))

	// 调用数据处理回调
	if err := r.dataHandler(&data); err != nil {
		r.logger.Error("failed to handle metric data",
			zap.String("device_id", data.DeviceID),
			zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	atomic.AddUint64(&r.receivedCount, 1)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Data received successfully",
	})
}

// handleBatchMetricData 处理批量数据推送
func (r *HTTPReceiver) handleBatchMetricData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("failed to read request body", zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// 解析JSON数组
	var dataList []models.MetricData
	if err := json.Unmarshal(body, &dataList); err != nil {
		r.logger.Error("failed to unmarshal batch data", zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	r.logger.Debug("received batch metric data", zap.Int("count", len(dataList)))

	// 处理每条数据
	successCount := 0
	failedCount := 0

	for i := range dataList {
		data := &dataList[i]

		// 如果时间戳为空，使用当前时间
		if data.Timestamp.IsZero() {
			data.Timestamp = time.Now()
		}

		// 调用数据处理回调
		if err := r.dataHandler(data); err != nil {
			r.logger.Error("failed to handle metric data in batch",
				zap.String("device_id", data.DeviceID),
				zap.Error(err))
			failedCount++
		} else {
			successCount++
		}
	}

	atomic.AddUint64(&r.receivedCount, uint64(successCount))
	atomic.AddUint64(&r.errorCount, uint64(failedCount))

	// 返回处理结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"total":         len(dataList),
		"success_count": successCount,
		"failed_count":  failedCount,
	})
}

// handleHealth 健康检查
func (r *HTTPReceiver) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"receiver": "http",
	})
}

// handleStats 统计信息
func (r *HTTPReceiver) handleStats(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"received_count": atomic.LoadUint64(&r.receivedCount),
		"error_count":    atomic.LoadUint64(&r.errorCount),
	})
}

// wrapWithAuth 添加认证中间件
func (r *HTTPReceiver) wrapWithAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 如果启用了认证
		if r.config.Auth.Enabled {
			// 检查Bearer Token
			if r.config.Auth.Type == "bearer" {
				authHeader := req.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
					return
				}

				expectedAuth := "Bearer " + r.config.Auth.Token
				if authHeader != expectedAuth {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}
			}

			// 检查Basic Auth
			if r.config.Auth.Type == "basic" {
				username, password, ok := req.BasicAuth()
				if !ok || username != r.config.Auth.Username || password != r.config.Auth.Password {
					w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
		}

		// 继续处理请求
		handler.ServeHTTP(w, req)
	})
}
