-- Migration: 000031_add_gate_decision.up.sql
-- Adds gate_decision column to agents table.
-- Stores the outcome of a human gate step (e.g. "approve", "reject", "skip").
-- NULL means no gate decision was recorded for this agent row.

ALTER TABLE agents ADD COLUMN gate_decision TEXT;
