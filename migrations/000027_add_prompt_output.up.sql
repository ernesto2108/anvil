-- Migration: 000027_add_prompt_output.up.sql
-- Adds output column to prompts table to store Claude's last_assistant_message
-- captured by the Stop hook.

ALTER TABLE prompts ADD COLUMN output TEXT NOT NULL DEFAULT '';
