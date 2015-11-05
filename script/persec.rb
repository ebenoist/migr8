require 'redis'

source_redis = Redis.new(:host => ARGV[0], :port => 6379, :db => 0)

while true do
  old = source_redis.info("keyspace")["db0"].split(",")[0].split('=')[1].to_i
  sleep 1
  new = source_redis.info("keyspace")["db0"].split(",")[0].split('=')[1].to_i
  puts "keys per second (#{new - old})"
end
