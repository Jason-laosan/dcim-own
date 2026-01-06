针对你**500万物联网节点 + 10分钟滑动窗口**的场景，我为你整理了一份**生产级 Flink 作业提交 Shell 脚本模板**。这份脚本包含了**检查点恢复、并行度优化、RocksDB 配置、资源限制**等所有最佳实践参数，可直接用于线上作业的提交、重启和版本升级。

### 一、 脚本核心功能
1.  支持**从保存点（Savepoint）恢复**（版本升级、集群扩容必备）
2.  支持**指定并行度、作业名称、配置文件**（适配大规模场景）
3.  支持**后台运行**（生产环境必选）
4.  包含**参数校验、日志输出、错误处理**（避免手动提交的误操作）
5.  集成**RocksDB 性能参数**（与之前的配置完全匹配）

### 二、 Flink 作业提交脚本（`submit_iot_window_job.sh`）
```bash
#!/bin/bash
set -e  # 遇到错误立即退出

################################################################################
# 1. 基础配置（根据你的集群环境修改，核心！）
################################################################################
# Flink 安装目录
FLINK_HOME="/opt/flink-1.17.1"
# Flink 客户端脚本路径
FLINK_BIN="${FLINK_HOME}/bin/flink"
# 作业 JAR 包路径（包含所有依赖，建议使用 uber-jar）
JAR_PATH="/opt/flink-jobs/iot-sliding-window-job-1.0.0.jar"
# 作业主类（包含包名）
MAIN_CLASS="com.iot.flink.job.IotSlidingWindowJob"
# 作业名称（需与 Grafana 监控模板中的 job_name 匹配）
JOB_NAME="iot-500w-node-10min-sliding-window"
# 并行度（与 flink-conf.yaml 中的 parallelism.default 一致，建议 64-128）
PARALLELISM=64
# 保存点目录（与 flink-conf.yaml 中的 state.savepoints.dir 一致）
SAVEPOINT_DIR="hdfs:///flink/savepoints"
# 作业日志目录（建议独立存储，方便排查问题）
LOG_DIR="/var/log/flink/jobs"
# 配置文件路径（作业的自定义配置，如 Kafka 地址、窗口参数等）
JOB_CONFIG="/opt/flink-jobs/config/iot-job.properties"

################################################################################
# 2. 可选参数（根据作业需求修改）
################################################################################
# RocksDB 额外配置（覆盖 flink-conf.yaml 中的配置，针对当前作业优化）
ROCKSDB_CONFIG="
-Dstate.backend.rocksdb.block.cache-size=16g \
-Dstate.backend.rocksdb.write-buffer-size=128m \
-Dstate.backend.rocksdb.compaction.style=UNIVERSAL \
-Dstate.backend.rocksdb.compression-type.l0-l1=LZ4 \
-Dstate.backend.rocksdb.compression-type.l2-plus=ZSTD
"

# JVM 优化参数（针对大规模作业，避免 OOM）
JVM_OPTS="
-Xms4g -Xmx4g \
-XX:+UseG1GC \
-XX:MaxGCPauseMillis=200 \
-XX:+HeapDumpOnOutOfMemoryError \
-XX:HeapDumpPath=${LOG_DIR}/heapdump.hprof \
-XX:+PrintGCDetails \
-XX:+PrintGCTimeStamps \
-XX:GCLogFileSize=100M \
-XX:NumberOfGCLogFiles=5 \
-Xloggc:${LOG_DIR}/gc.log
"

################################################################################
# 3. 命令行参数解析（支持提交、重启、停止作业）
################################################################################
usage() {
    echo "Usage: $0 [submit|restart|stop] [savepoint-path (optional)]"
    echo "  submit: 提交新作业"
    echo "  restart: 从保存点重启作业（需指定保存点路径）"
    echo "  stop: 停止作业并生成保存点"
    exit 1
}

# 检查参数数量
if [ $# -lt 1 ]; then
    usage
fi

COMMAND=$1
SAVEPOINT_PATH=$2

# 创建日志目录
mkdir -p ${LOG_DIR}

################################################################################
# 4. 作业操作函数（提交、重启、停止）
################################################################################
# 提交新作业
submit_job() {
    echo "====================================="
    echo "提交 Flink 作业: ${JOB_NAME}"
    echo "并行度: ${PARALLELISM}"
    echo "JAR 包: ${JAR_PATH}"
    echo "主类: ${MAIN_CLASS}"
    echo "日志目录: ${LOG_DIR}"
    echo "====================================="

    ${FLINK_BIN} run \
        -d \  # 后台运行（生产必选）
        -p ${PARALLELISM} \  # 指定并行度
        -jobname ${JOB_NAME} \  # 指定作业名称
        -class ${MAIN_CLASS} \  # 指定主类
        -Djobmanager.memory.process.size=16g \  # JobManager 内存
        -Dtaskmanager.memory.process.size=64g \  # TaskManager 内存
        -Dtaskmanager.numberOfTaskSlots=8 \  # TaskManager Slot 数量
        ${ROCKSDB_CONFIG} \  # RocksDB 额外配置
        ${JVM_OPTS} \  # JVM 优化参数
        ${JAR_PATH} \  # 作业 JAR 包
        --config ${JOB_CONFIG}  # 作业自定义配置

    echo "作业提交成功！可通过 Flink Web UI 查看状态: http://jobmanager:8081"
}

# 从保存点重启作业
restart_job() {
    if [ -z "${SAVEPOINT_PATH}" ]; then
        echo "错误：重启作业必须指定保存点路径！"
        usage
    fi

    echo "====================================="
    echo "从保存点重启 Flink 作业: ${JOB_NAME}"
    echo "保存点路径: ${SAVEPOINT_PATH}"
    echo "并行度: ${PARALLELISM}"
    echo "====================================="

    ${FLINK_BIN} run \
        -d \
        -p ${PARALLELISM} \
        -jobname ${JOB_NAME} \
        -class ${MAIN_CLASS} \
        -s ${SAVEPOINT_PATH} \  # 从保存点恢复（核心！）
        -Djobmanager.memory.process.size=16g \
        -Dtaskmanager.memory.process.size=64g \
        -Dtaskmanager.numberOfTaskSlots=8 \
        ${ROCKSDB_CONFIG} \
        ${JVM_OPTS} \
        ${JAR_PATH} \
        --config ${JOB_CONFIG}

    echo "作业重启成功！"
}

# 停止作业并生成保存点
stop_job() {
    echo "====================================="
    echo "停止 Flink 作业: ${JOB_NAME} 并生成保存点"
    echo "保存点目录: ${SAVEPOINT_DIR}"
    echo "====================================="

    # 获取作业 ID（通过作业名称过滤）
    JOB_ID=$(${FLINK_BIN} list -running | grep ${JOB_NAME} | awk '{print $4}')

    if [ -z "${JOB_ID}" ]; then
        echo "错误：未找到运行中的作业 ${JOB_NAME}！"
        exit 1
    fi

    # 停止作业并生成保存点
    ${FLINK_BIN} stop \
        -p ${SAVEPOINT_DIR} \  # 保存点输出目录
        ${JOB_ID}

    echo "作业停止成功！保存点已生成到 ${SAVEPOINT_DIR}"
}

################################################################################
# 5. 执行命令
################################################################################
case ${COMMAND} in
    submit)
        submit_job
        ;;
    restart)
        restart_job
        ;;
    stop)
        stop_job
        ;;
    *)
        usage
        ;;
esac
```

### 三、 作业自定义配置文件（`iot-job.properties`）
脚本中引用了 `--config ${JOB_CONFIG}` 参数，用于传递作业的自定义配置（如 Kafka 地址、窗口参数、设备类型等），避免硬编码。示例如下：
```properties
# Kafka 数据源配置
kafka.bootstrap.servers=kafka-node1:9092,kafka-node2:9092,kafka-node3:9092
kafka.consumer.topic=iot_topic
kafka.consumer.group.id=iot-flink-window-group
kafka.producer.topic=iot_window_result_topic

# 窗口参数配置（与代码中的 10 分钟滑动窗口一致）
window.size.minutes=10
window.slide.seconds=30
watermark.outoforderness.seconds=30

# 物联网设备配置
device.id.prefix=dev_
device.count=5000000
value.range.min=0.0
value.range.max=100.0

# 输出配置
output.format=json
output.batch.size=1000
```

### 四、 脚本使用方法
1.  **赋予执行权限**
    ```bash
    chmod +x submit_iot_window_job.sh
    ```
2.  **提交新作业**
    ```bash
    ./submit_iot_window_job.sh submit
    ```
3.  **停止作业并生成保存点**（用于版本升级、集群扩容）
    ```bash
    ./submit_iot_window_job.sh stop
    ```
4.  **从保存点重启作业**（需指定保存点路径，路径来自 `stop` 命令的输出）
    ```bash
    ./submit_iot_window_job.sh restart hdfs:///flink/savepoints/savepoint-abc123
    ```

### 五、 生产环境关键最佳实践
1.  **使用 Uber-Jar**：作业 JAR 包必须包含所有依赖（如 Flink 连接器、JSON 解析库等），避免集群依赖缺失。可通过 Maven/Gradle 打包为 uber-jar。
2.  **后台运行（-d 参数）**：生产环境必须使用 `-d` 参数，避免脚本退出导致作业终止。
3.  **保存点恢复（-s 参数）**：版本升级、集群扩容时，必须从保存点恢复作业，避免状态丢失和数据重复。
4.  **并行度匹配**：脚本中的 `PARALLELISM` 必须与 `flink-conf.yaml` 中的 `parallelism.default` 一致，避免资源浪费。
5.  **日志独立存储**：作业日志目录 `LOG_DIR` 建议独立挂载磁盘，避免日志占满系统盘。
6.  **JVM 优化**：脚本中的 `JVM_OPTS` 已针对大规模作业优化，使用 G1GC 垃圾收集器，避免 Full GC 导致的作业延迟。

### 六、 常见问题与解决方案
| 问题现象 | 原因 | 解决方案 |
|----------|------|----------|
| 作业提交失败，提示内存不足 | JVM 堆内存配置过小 | 增大 `JVM_OPTS` 中的 `-Xms` 和 `-Xmx`（建议 4G-8G） |
| 从保存点恢复失败 | 保存点路径错误或权限不足 | 检查保存点路径是否存在，Flink 进程是否有 HDFS 读写权限 |
| 作业并行度不生效 | 脚本中的 `PARALLELISM` 与代码中的并行度冲突 | 代码中不指定并行度，完全由脚本控制 |
| RocksDB 性能不佳 | 脚本中的 `ROCKSDB_CONFIG` 与 `flink-conf.yaml` 冲突 | 保持两者配置一致，或在脚本中覆盖需要调整的参数 |

这份脚本已经完全适配你的 500 万物联网节点滑动窗口场景，可直接用于生产环境。需要我帮你整理一份**作业上线前的检查清单**吗？里面包含集群配置、作业参数、监控告警等所有需要验证的项，确保上线一次成功。