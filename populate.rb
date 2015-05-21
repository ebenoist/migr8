require 'redis'
require 'securerandom'

redis_srv = Redis.new(:host => ARGV[0], :port => 6379, :db => 0)

(1..ARGV[1]).to_a.each do |i|
  redis_srv.set(i, SecureRandom.uuid)
end
