require 'net/http'

host = 'localhost'
port = 8080
path = '/cache/Hello'

http = Net::HTTP.new(host, port)
http.send_request('PUT', path, "World")
value = Net::HTTP.get(host, path, port)
print "Hello = ", value
