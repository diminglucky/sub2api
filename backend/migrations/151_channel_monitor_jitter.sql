-- Migration: 151_channel_monitor_jitter
-- Add optional channel monitor schedule jitter. 0 keeps previous fixed-interval behavior.

ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS jitter_seconds INTEGER NOT NULL DEFAULT 0;
