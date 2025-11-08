CREATE TABLE IF NOT EXISTS responses ( 
  req_hash TEXT PRIMARY KEY,
  body TEXT,
  headers TEXT,
  status_code INT,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
