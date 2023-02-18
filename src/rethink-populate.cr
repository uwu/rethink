require "crystal-argon2"
require "db"
require "option_parser"
require "uuid"
require "sqlite3"

db = DB.open "sqlite3:./rethink.sqlite"
uname = ""

OptionParser.parse do |parser|
  parser.banner = "Usage: rethink-populate [arguments]"
  parser.on("-n NAME", "--name NAME", "Username to add") { |name| uname = name }
  parser.on("-h", "--help", "Show this help") do
    puts parser
    exit
  end
  parser.invalid_option do |flag|
    STDERR.puts "Error: #{flag} is not a valid option."
    STDERR.puts parser
    exit(1)
  end
end

if uname.empty?
  STDERR.puts "You didn't provide a name! See --help for info."
  exit(1)
end

uuid = UUID.random
hash = Argon2::Password.create(uuid.to_s)
db.exec("INSERT INTO users (name, thought_key) VALUES (?, ?)", uname, hash)

puts uuid
