CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY,
  conversation_id UUID,
  role VARCHAR(50),
  content TEXT,
  created_at TIMESTAMP,
  is_file BOOLEAN,
  FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);