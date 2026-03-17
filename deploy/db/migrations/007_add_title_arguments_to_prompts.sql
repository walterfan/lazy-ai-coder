-- Migration: Add title and arguments columns to prompts table
-- Date: 2025-11-22
-- Description:
--   Adds 'title' field for human-readable display names
--   Adds 'arguments' field for explicit argument definitions as JSON array
--   Format: [{"name": "arg1", "description": "desc", "required": true}]

-- Add title column for human-readable display name
ALTER TABLE prompt ADD COLUMN title TEXT;

-- Add arguments column to store JSON array of argument definitions
ALTER TABLE prompt ADD COLUMN arguments TEXT;

-- Update existing prompts to set title from name (you can update these manually later)
UPDATE prompt SET title = name WHERE title IS NULL OR title = '';
