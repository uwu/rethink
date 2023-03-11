require "db"
require "sqlite3"

db = DB.open "sqlite3:./rethink.sqlite"

db.exec("CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            thought_key TEXT
          )")

db.exec("CREATE TABLE IF NOT EXISTS thoughts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            author_id INTEGER,
            content TEXT,
            date DATE DEFAULT (datetime('now'))
          )")
