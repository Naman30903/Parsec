-- Create database
CREATE DATABASE IF NOT EXISTS logs;

-- Use the database
USE logs;

-- Create log events table with optimized schema
CREATE TABLE IF NOT EXISTS events
(
    id String,
    tenant_id String,
    timestamp DateTime64(3, 'UTC'),
    severity LowCardinality(String),
    source LowCardinality(String),
    message String,
    metadata Map(String, String),
    trace_id String,
    span_id String,
    received_at DateTime64(3, 'UTC'),
    ingest_node LowCardinality(String),
    partition_key String,
    
    -- Indexes
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_source source TYPE bloom_filter GRANULARITY 1,
    INDEX idx_severity severity TYPE set(0) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, timestamp, id)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Create materialized view for hourly aggregations by severity
CREATE MATERIALIZED VIEW IF NOT EXISTS events_by_severity_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (tenant_id, severity, hour)
AS SELECT
    tenant_id,
    severity,
    toStartOfHour(timestamp) as hour,
    count() as count
FROM events
GROUP BY tenant_id, severity, hour;

-- Create materialized view for hourly aggregations by source
CREATE MATERIALIZED VIEW IF NOT EXISTS events_by_source_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (tenant_id, source, hour)
AS SELECT
    tenant_id,
    source,
    toStartOfHour(timestamp) as hour,
    count() as count
FROM events
GROUP BY tenant_id, source, hour;

-- Create table for error tracking
CREATE TABLE IF NOT EXISTS error_summary
(
    tenant_id String,
    source String,
    error_pattern String,
    first_seen DateTime64(3, 'UTC'),
    last_seen DateTime64(3, 'UTC'),
    count UInt64
)
ENGINE = SummingMergeTree()
ORDER BY (tenant_id, source, error_pattern)
SETTINGS index_granularity = 8192;

-- Grant permissions to parsec user
GRANT ALL ON logs.* TO parsec;

-- Show created tables
SHOW TABLES FROM logs;