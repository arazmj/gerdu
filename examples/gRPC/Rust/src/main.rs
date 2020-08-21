mod gerdu;
mod gerdu_grpc;

use std::sync::Arc;
use grpcio::{ChannelBuilder, EnvBuilder};
use gerdu::*;
use gerdu_grpc::GerduClient;

fn main() {
    let env = Arc::new(EnvBuilder::new().build());
    let ch = ChannelBuilder::new(env).connect("localhost:8081");
    let client = GerduClient::new(ch);

    let mut put_request = PutRequest::default();
    put_request.set_key("Hello".to_owned());
    put_request.set_value("World".as_bytes().to_vec());
    client.put(&put_request).expect("rpc");

    let mut get_request = GetRequest::default();
    get_request.set_key("Hello".to_owned());
    let get_response = client.get(&get_request).expect("rpc");
    let value = String::from_utf8(get_response.value).expect("Found invalid UTF-8");
    println!("Hello = {}", value);
}
