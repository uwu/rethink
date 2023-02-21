require "option_parser"
require "db"
require "sqlite3"
require "uuid"
require "crystal-argon2"

file = "./rethink.sqlite"

enum Command
  User
end

enum Action
  Create
  Delete
  Query
end

enum DeleteType
  Id
  LikeName
end

command : Command? = nil
action : Action? = nil

force = false
create_key : String? = nil
query_filter : String? = nil

parser = OptionParser.parse do |parser|
  parser.banner = "Usage: admin <command> [flags]"

  parser.separator "\nGlobal flags:"
  parser.on("-d", "--database", "Select databse to work on (defaults to \"./rethink.sqlite\")") do |arg|
    file = arg
  end
  parser.on("-h", "--help", "Show this help") do
    puts parser
    exit
  end

  parser.separator "\nCommands:"

  parser.on("user", "Manipulate users") do
    parser.banner = "Usage: admin user <action> [flags]\n\nGlobal flags:"

    command = Command::User

    parser.separator "\nActions:"
    parser.on("create", "Create a user") do
      parser.banner = "Usage: admin user create <name>\n\nGlobal flags:"

      action = Action::Create

      parser.separator "\nFlags:"
      parser.on("-k", "--key [KEY]", "Custom thought key") do |arg|
        create_key = arg
      end
    end

    parser.on("delete", "Delete a user") do
      parser.banner = "Usage: admin user delete <filter>\n\nGlobal flags:"

      action = Action::Delete

      parser.separator "\nFlags:"
      parser.on("-f", "--force", "Pass to allow deletion of more than one user") do
        force = true
      end
    end

    parser.on("query", "Query users") do
      parser.banner = "Usage: admin user query [flags]\n\nGlobal flags:"

      action = Action::Query

      parser.separator "\nFlags:"
      parser.on("-n", "--name NAME", "Filter LIKE name") do |arg|
        query_filter = arg
      end
    end
  end

  # parser.on("thought", "Mind control (Manipulate thoughts)") do
  #   parser.banner = "Usage: admin thought <action> [flags]\n\nGlobal flags:"
  # end

  parser.separator "\nRun a command followed by --help to see command specific help."
end

if command.nil? || action.nil?
  puts parser
  exit
end

db = DB.open "sqlite3:#{file}"

case command
when Command::User
  case action
  when Action::Create
    if ARGV.size < 1
      puts "no name was specified"
      exit 1
    end

    name = ARGV[0]
    exists = false

    db.query("SELECT 1 FROM users WHERE name = ?", name) do |rs|
      rs.each do exists = true end
    end

    if exists
      puts "user already exists"
      exit 1
    end

    pass = if create_key.nil? UUID.random else create_key end
    hash = Argon2::Password.create(pass.to_s)
    db.exec("INSERT INTO users (name, thought_key) VALUES (?, ?)", name, hash)
    puts "#{name}:#{pass}"
  when Action::Delete
    if ARGV.size < 1
      puts "no filter was specified"
      exit 1
    end
    split = ARGV[0].split(":", limit: 2)

    if split.size < 2
      puts "invalid filter, no the format is not documented anywhere fuck you"
      exit 1
    end

    rawType, rawFilter = split

    # Multiple assignment is not allowed for constants
    type =
      case rawType
      when "id"
        DeleteType::Id
      when "name"
        DeleteType::LikeName
      else
        nil
      end

    filter =
      case type
      when DeleteType::Id
        rawFilter.to_i { nil }
      when DeleteType::LikeName
        rawFilter
      else
        nil
      end

    if filter.nil?
      puts "invalid filter, no the format is not documented anywhere fuck you"
      exit 1
    end

    query = String.build do |str|
      str << "SELECT id, name FROM users WHERE "
      case type
      when DeleteType::Id
        str << "id = ?"
      when DeleteType::LikeName
        str << "name LIKE ?"
      end
    end

    users = [] of {Int32, String}
    db.query(query, filter) do |rs|
      rs.each do users << {rs.read(Int32), rs.read(String)} end
    end

    if users.empty?
      puts "no users like that found"
      exit 1
    end

    if users.size > 1 && !force
      puts "cannot delete #{users.size} users without force"
      exit 1
    end

    query = String.build do |str|
      str << "DELETE FROM users WHERE "
      case type
      when DeleteType::Id
        str << "id = ?"
      when DeleteType::LikeName
        str << "name LIKE ?"
      end
    end

    db.exec(query, filter)
    puts "deleted #{users.size} users: #{users.map { |u| u[1] }.join(", ")}"
    res = db.exec("DELETE FROM thoughts WHERE author_id in (#{users.map { |u| u[0] }.join(", ")})")
    puts "deleted #{res.rows_affected} accompanying thoughts"
  when Action::Query
    if query_filter.nil?
      db.query("SELECT id, name FROM users") do |rs|
        rs.each do
          puts "#{rs.read(Int32)}\t#{rs.read(String)}"
        end
      end
    else
      db.query("SELECT id, name FROM users WHERE name LIKE ?", query_filter) do |rs|
        rs.each do
          puts "#{rs.read(Int32)}\t#{rs.read(String)}"
        end
      end
    end
  end
end
