CREATE TABLE IF NOT EXISTS responses ( 
  req_hash TEXT PRIMARY KEY,
  body TEXT,
  headers TEXT,
  status_code INT,
  updated_at TEXT DEFAULT (datetime('now', 'localtime'))
);
