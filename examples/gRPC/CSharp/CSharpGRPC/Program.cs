using System;
using System.Net;
using Gerdu;
using Google.Protobuf;
using Grpc.Core;

namespace CSharpGRPC
{
    class Program
    {
        static void Main(string[] args)
        {
            Channel channel = new Channel("localhost:8081", ChannelCredentials.Insecure);
            var client = new Gerdu.Gerdu.GerduClient(channel);

            client.Put(new PutRequest()
            {
                Key = "Hello",
                Value = ByteString.CopyFromUtf8("World")
            });

            var response = client.Get(new GetRequest() { Key = "Hello" });
            var value = response.Value.ToStringUtf8();
            Console.WriteLine("Hello = " + value);

            channel.ShutdownAsync().Wait();
        }
    }
}