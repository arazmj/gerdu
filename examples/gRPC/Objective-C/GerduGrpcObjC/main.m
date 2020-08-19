//
//  main.m
//  GerduGrpcObjC
//
//  Created by Amir Razmjou on 8/19/20.
//  Copyright © 2020 Amir Razmjou. All rights reserved.
//

#import <GRPCClient/GRPCCall+ChannelArg.h>
#import <Gerdu.pbrpc.h>

static NSString *const kHostAddress = @"localhost:8081";

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        dispatch_group_t serviceGroup = dispatch_group_create();

        [GRPCCall useInsecureConnectionsForHost:kHostAddress];
        Gerdu *client = [[Gerdu alloc] initWithHost:kHostAddress];
        PutRequest *putRequest = [PutRequest message];
        putRequest.key = @"Hello";
        putRequest.value = [@"World" dataUsingEncoding:NSUTF8StringEncoding];

        dispatch_group_enter(serviceGroup);
        [client putWithRequest:putRequest handler:^(PutResponse *response, NSError *error) {
            GetRequest *getRequest = [GetRequest message];
            getRequest.key = @"Hello";
            [client getWithRequest:getRequest handler:^(GetResponse *response, NSError *error) {
                NSString *value = [[NSString alloc] initWithData:response.value encoding:NSUTF8StringEncoding];
                NSLog(@"Hello = %@",value);
                dispatch_group_leave(serviceGroup);
            }];
        }];

        dispatch_group_notify(serviceGroup, dispatch_get_main_queue(), ^{
            exit(EXIT_SUCCESS);
        });

    }
    dispatch_main();
}
