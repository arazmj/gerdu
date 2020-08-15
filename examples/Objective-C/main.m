//
//  main.m
//  GoCacheObjC
//
//  Created by Amir Razmjou on 8/13/20.
//  Copyright © 2020 Amir Razmjou. All rights reserved.
//

#import <Foundation/Foundation.h>

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        dispatch_group_t serviceGroup = dispatch_group_create();
        NSURLSession *defaultSession = [NSURLSession sharedSession];
        NSURL *url = [NSURL URLWithString:@"http://localhost:8080/cache/Hello"];

        NSMutableURLRequest *putRequest = [[NSMutableURLRequest alloc] initWithURL:url];
        putRequest.HTTPMethod = @"PUT";
        NSData *postBody = [@"World" dataUsingEncoding:NSUTF8StringEncoding];
        [putRequest setHTTPBody:postBody];

        dispatch_group_enter(serviceGroup);

        [[defaultSession dataTaskWithRequest:putRequest
                           completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
                               NSMutableURLRequest *getRequest = [[NSMutableURLRequest alloc] initWithURL:url];
                               getRequest.HTTPMethod = @"GET";
                               [[[NSURLSession sharedSession] dataTaskWithRequest:getRequest completionHandler:
                                       ^(NSData *_Nullable data,
                                               NSURLResponse *_Nullable response,
                                               NSError *_Nullable error) {

                                           NSString *value = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
                                           NSLog(@"Hello = %@", value);
                                           dispatch_group_leave(serviceGroup);
                                       }] resume];
                           }] resume];

        dispatch_group_wait(serviceGroup, DISPATCH_TIME_FOREVER);
    }
    return 0;
}