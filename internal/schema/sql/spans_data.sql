CREATE TABLE IF NOT EXISTS spans_data (
    project_id UInt32,
    trace_id   String,
    span_id    String,
    time       DateTime64(6) CODEC(Delta, ZSTD(1)),
    data       String CODEC(ZSTD(1))
) ENGINE = MergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, trace_id, span_id)
TTL toDateTime(time) + INTERVAL 7 DAY
SETTINGS ttl_only_drop_parts = 1
