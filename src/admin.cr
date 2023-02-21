require "option_parser"
require "db"
require "sqlite3"
require "uuid"
require "crystal-argon2"

enum Action
Create
Reset
end

file = "./rethink.sqlite"
action : Action? = nil
parser = OptionParser.parse do |parser|
  parser.banner = "Usage: admin <action> [flags]"

  parser.separator "\nActions:"
  parser.on("create", "Create a user") do
    parser.banner = "Usage: admin create <name>\n\nGlobal flags:"

    action = Action::Create
  end
  parser.on("reset", "Reset thought key of a user") do
    parser.banner = "Usage: admin reset <name>\n\nGlobal flags:"

    action = Action::Reset
  end

  parser.separator "\nFlags:"
  parser.on("-d", "--database", "Select databse to work on (defaults to \"./rethink.sqlite\")") do |arg|
    file = arg
  end
  parser.on("-h", "--help", "Show this help") do
    puts parser
    exit
  end
end

if action.nil?
  puts parser
  exit
end

db = DB.open "sqlite3:#{file}"

if ARGV.size < 1
  puts "no username was specified"
  exit 1
end
name = ARGV[0]

exists = false
db.query("SELECT 1 FROM users WHERE name = ?", name) do |rs|
  rs.each do exists = true end
end

case action
when Action::Create
  if exists
    puts "user already exists"
    exit 1
  end

  pass = UUID.random
  hash = Argon2::Password.create(pass.to_s)
  db.exec("INSERT INTO users (name, thought_key) VALUES (?, ?)", name, hash)
  puts "#{name}:#{pass}"
when Action::Reset
  if !exists
    puts "user does not exist"
    exit 1
  end

  pass = UUID.random
  hash = Argon2::Password.create(pass.to_s)
  db.exec("UPDATE users SET thought_key = ? WHERE name = ?", hash, name)
  puts "#{name}:#{pass}"
end
