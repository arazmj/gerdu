#include <iostream>
#include <grpcpp/grpcpp.h>
#include "proto/gerdu.grpc.pb.h"
using namespace gerdu;
using namespace std;
using namespace grpc;

int main() {
    const char *target = "localhost:8081";
    auto channel = CreateChannel(
            target, InsecureChannelCredentials());
    unique_ptr<Gerdu::Stub> stub(Gerdu::NewStub(channel));
    PutRequest putRequest;
    putRequest.set_key("Hello");
    putRequest.set_value("World");
    PutResponse putResponse;
    ClientContext context1;
    stub->Put(&context1, putRequest, &putResponse);
    // putResponse.created == 1
    GetRequest getRequest;
    getRequest.set_key("Hello");
    GetResponse getResponse;
    ClientContext context2;
    stub->Get(&context2, getRequest, &getResponse);
    cout << "Hello = " << getResponse.value() << endl;
    return 0;
}
