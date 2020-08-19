//
//  main.m
//  GerduGrpcObjC
//
//  Created by Amir Razmjou on 8/19/20.
//  Copyright © 2020 Amir Razmjou. All rights reserved.
//

#import <GRPCClient/GRPCCall+ChannelArg.h>
#import <GRPCClient/GRPCTransport.h>
#import <Gerdu.pbrpc.h>

static NSString *const kHostAddress = @"localhost:8081";

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        dispatch_group_t serviceGroup = dispatch_group_create();

        GRPCMutableCallOptions *options = [[GRPCMutableCallOptions alloc] init];
        options.transport = GRPCDefaultTransportImplList.core_insecure;
        options.userAgentPrefix = @"Gerdu/1.0";

        Gerdu *client = [[Gerdu alloc] initWithHost:kHostAddress callOptions:options];

        PutRequest *putRequest = [PutRequest message];
        putRequest.key = @"Hello";
        putRequest.value = [@"World" dataUsingEncoding:NSUTF8StringEncoding];

        dispatch_group_enter(serviceGroup);
        GRPCUnaryResponseHandler *putResponseHandler =
                [[GRPCUnaryResponseHandler alloc] initWithResponseHandler:^(PutResponse *response, NSError *error) {
                    dispatch_group_leave(serviceGroup);
                }                                   responseDispatchQueue:nil];
        [[client putWithMessage:putRequest
                responseHandler:putResponseHandler callOptions:nil] start];

        dispatch_group_wait(serviceGroup, 100);

        GetRequest *getRequest = [GetRequest message];
        getRequest.key = @"Hello";
        dispatch_group_enter(serviceGroup);
        GRPCUnaryResponseHandler *getResponseHandler =
                [[GRPCUnaryResponseHandler alloc] initWithResponseHandler:^(GetResponse *response, NSError *error) {
                    dispatch_group_leave(serviceGroup);

                    NSString *value = [[NSString alloc] initWithData:response.value encoding:NSUTF8StringEncoding];
                    NSLog(@"Hello = %@",value);

                }                                   responseDispatchQueue:nil];
        [[client getWithMessage:getRequest responseHandler:getResponseHandler callOptions:nil] start];

        dispatch_group_notify(serviceGroup, dispatch_get_main_queue(), ^{
            exit(EXIT_SUCCESS);
        });
        dispatch_main();
    }
    return 0;
}
