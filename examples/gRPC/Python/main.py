# Gerdu gRPC Python Example
# Amir Razmjou 2020

from proto.gerdu_pb2 import *
from proto.gerdu_pb2_grpc import *

if __name__ == '__main__':
    channel = grpc.insecure_channel('localhost:8081')
    stub = GerduStub(channel)
    stub.Put(PutRequest(key="Hello", value=b"World"))
    response = stub.Get(GetRequest(key="Hello"))
    values = response.value.decode("utf-8")
    print("Hello =", values)
    stub.Delete(DeleteRequest(key="Hello"))
    try:
        response = stub.Get(GetRequest(key="Hello"))
    except:
        print('The key is deleted successfully')


