CREATE TABLE IF NOT EXISTS logs_index (
    project_id   UInt32,
    trace_id     String,
    span_id      String,
    service_name LowCardinality(String),
    severity     LowCardinality(String),
    body         String,
    time         DateTime CODEC(Delta, ZSTD(1)),
    attr_keys    Array(LowCardinality(String)),
    attr_values  Array(String),
    INDEX idx_trace_id  trace_id    TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_body      body        TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1,
    INDEX idx_attr_keys attr_keys   TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_attr_vals attr_values TYPE bloom_filter(0.01) GRANULARITY 1
) ENGINE = MergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, service_name, time)
TTL time + INTERVAL 30 DAY
SETTINGS ttl_only_drop_parts = 1
