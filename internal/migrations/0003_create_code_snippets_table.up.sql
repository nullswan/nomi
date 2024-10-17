CREATE TABLE IF NOT EXISTS code_snippets (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP,
  description TEXT,
  code TEXT,
  language TEXT
);