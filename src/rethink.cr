require "kemal"
require "db"
require "sqlite3"
require "crystal-argon2"
require "ecr"

DATABASE = DB.open "sqlite3:./rethink.sqlite"

class Thought
  property :content, :date

  def initialize(content : String, date : Time)
    @content = content
    @date = date
  end
end

class Thoughts
  def initialize(@thoughts : Array(Thought))
  end

  ECR.def_to_s "src/views/thoughts.ecr"
end

class Feed
  def initialize(@user : String, @thoughts : Array(Thought))
  end

  ECR.def_to_s "src/views/feed.ecr"
end

def getThoughtsByUser(name : String) : Array(Thought)
  thoughts = [] of Thought

  id : Int32? = nil
  DATABASE.query("SELECT id FROM users WHERE name = ?", name) do |rows|
    rows.each do
      id = rows.read(Int32)
    end
  end

  if id.nil?
    raise "User not found"
  end

  DATABASE.query("SELECT content, date FROM thoughts WHERE author_id = ?", id) do |rows|
    rows.each do
      content = rows.read(String)
      date = rows.read(Time)
      thoughts << Thought.new(content, date)
    end
  end

  thoughts
end

get "/~:name" do |ctx|
  name = ctx.params.url["name"]
  begin
    thoughts = getThoughtsByUser(name)
  rescue ex
  end

  if thoughts.nil?
    ctx.response.status_code = 404
    next "User not found"
  end

  Thoughts.new(thoughts).to_s
end

get "/~:name/feed" do |ctx|
  ctx.response.headers["Content-Type"] = "application/atom+xml"
  name = ctx.params.url["name"]
  thoughts = getThoughtsByUser(name)
  Feed.new(name, thoughts).to_s
end

get "/" do
  render "public/index.html"
end

# Reuse this later for population scripts
# get "/api/hash" do |ctx|
#   Argon2::Password.create(ctx.request.body.as(IO).gets_to_end)
# end

# post to rethink
put "/api/think" do |ctx|
  unless ctx.request.headers.has_key?("authorization") && ctx.request.headers.has_key?("name")
    ctx.response.status_code = 401
    next "Unauthorized"
  end

  auth = ctx.request.headers["authorization"]
  username = ctx.request.headers["name"]

  id : Int32? = nil
  thought_key = ""

  DATABASE.query("SELECT id, thought_key FROM users WHERE name = ?", username) do |rows|
    rows.each do
      id = rows.read(Int32)
      thought_key = rows.read(String)
    end
  end

  authorized : Argon2::Response? = nil
  begin
    authorized = Argon2::Password.verify_password(auth, thought_key)
  rescue ex
  end

  unless authorized == Argon2::Response::ARGON2_OK
    ctx.response.status_code = 401
    next "Unauthorized"
  end

  thought = if ctx.request.body.nil?
              ""
            else
              ctx.request.body.as(IO).gets_to_end
            end
  DATABASE.exec("INSERT INTO thoughts (author_id, content) VALUES (?, ?)", id, thought)
  ctx.response.status_code = 201
end

begin
  Kemal.config.env = "production"
  Kemal.run
ensure
  DATABASE.close
end
