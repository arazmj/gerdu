require 'grpc'
require_relative  'lib/gerdu_pb'
require_relative 'lib/gerdu_services_pb'

include Gerdu

def main
  stub = Gerdu::Gerdu::Stub.new('localhost:8081', :this_channel_is_insecure)
  stub.put(PutRequest.new(key: 'Hello', value: 'World'))
  response = stub.get(GetRequest.new(key: 'Hello'))
  print('Hello = ', response.value)
end

main

