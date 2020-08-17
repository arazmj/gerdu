package net.amirrazmjou;

import com.amirrazmjou.gerdu.java.GerduGrpc;
import com.amirrazmjou.gerdu.java.GetRequest;
import com.amirrazmjou.gerdu.java.GetResponse;
import com.amirrazmjou.gerdu.java.PutRequest;
import com.google.protobuf.ByteString;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

import java.nio.charset.Charset;
import java.util.concurrent.TimeUnit;

public class Main {
    public static void main(String[] args) throws InterruptedException {
        String target = "localhost:8081";
        ManagedChannel channel = ManagedChannelBuilder.forTarget(target)
                // Channels are secure by default (via SSL/TLS). For the example we disable TLS to avoid
                // needing certificates.
                .usePlaintext()
                .build();
        try {
            GerduGrpc.GerduBlockingStub blockingStub =
                    GerduGrpc.newBlockingStub(channel);

            ByteString byteStringValue = ByteString.copyFrom("World",
                    Charset.defaultCharset());
            PutRequest putRequest = PutRequest.newBuilder()
                    .setKey("Hello")
                    .setValue(byteStringValue)
                    .build();
            blockingStub.put(putRequest);

            GetRequest getRequest = GetRequest.newBuilder()
                    .setKey("Hello")
                    .build();

            GetResponse response = blockingStub.get(getRequest);
            ByteString byteString = response.getValue();
            String value = byteString.toString(Charset.defaultCharset());
            System.out.printf("Hello = %s%n", value);
        } finally {
            channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
        }
    }
}
