// This file is generated. Do not edit
// @generated

// https://github.com/Manishearth/rust-clippy/issues/702
#![allow(unknown_lints)]
#![allow(clippy::all)]

#![cfg_attr(rustfmt, rustfmt_skip)]

#![allow(box_pointers)]
#![allow(dead_code)]
#![allow(missing_docs)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]
#![allow(non_upper_case_globals)]
#![allow(trivial_casts)]
#![allow(unsafe_code)]
#![allow(unused_imports)]
#![allow(unused_results)]

const METHOD_GERDU_PUT: ::grpcio::Method<super::gerdu::PutRequest, super::gerdu::PutResponse> = ::grpcio::Method {
    ty: ::grpcio::MethodType::Unary,
    name: "/gerdu.Gerdu/Put",
    req_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
    resp_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
};

const METHOD_GERDU_GET: ::grpcio::Method<super::gerdu::GetRequest, super::gerdu::GetResponse> = ::grpcio::Method {
    ty: ::grpcio::MethodType::Unary,
    name: "/gerdu.Gerdu/Get",
    req_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
    resp_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
};

const METHOD_GERDU_DELETE: ::grpcio::Method<super::gerdu::DeleteRequest, super::gerdu::DeleteResponse> = ::grpcio::Method {
    ty: ::grpcio::MethodType::Unary,
    name: "/gerdu.Gerdu/Delete",
    req_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
    resp_mar: ::grpcio::Marshaller { ser: ::grpcio::pb_ser, de: ::grpcio::pb_de },
};

#[derive(Clone)]
pub struct GerduClient {
    client: ::grpcio::Client,
}

impl GerduClient {
    pub fn new(channel: ::grpcio::Channel) -> Self {
        GerduClient {
            client: ::grpcio::Client::new(channel),
        }
    }

    pub fn put_opt(&self, req: &super::gerdu::PutRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<super::gerdu::PutResponse> {
        self.client.unary_call(&METHOD_GERDU_PUT, req, opt)
    }

    pub fn put(&self, req: &super::gerdu::PutRequest) -> ::grpcio::Result<super::gerdu::PutResponse> {
        self.put_opt(req, ::grpcio::CallOption::default())
    }

    pub fn put_async_opt(&self, req: &super::gerdu::PutRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::PutResponse>> {
        self.client.unary_call_async(&METHOD_GERDU_PUT, req, opt)
    }

    pub fn put_async(&self, req: &super::gerdu::PutRequest) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::PutResponse>> {
        self.put_async_opt(req, ::grpcio::CallOption::default())
    }

    pub fn get_opt(&self, req: &super::gerdu::GetRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<super::gerdu::GetResponse> {
        self.client.unary_call(&METHOD_GERDU_GET, req, opt)
    }

    pub fn get(&self, req: &super::gerdu::GetRequest) -> ::grpcio::Result<super::gerdu::GetResponse> {
        self.get_opt(req, ::grpcio::CallOption::default())
    }

    pub fn get_async_opt(&self, req: &super::gerdu::GetRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::GetResponse>> {
        self.client.unary_call_async(&METHOD_GERDU_GET, req, opt)
    }

    pub fn get_async(&self, req: &super::gerdu::GetRequest) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::GetResponse>> {
        self.get_async_opt(req, ::grpcio::CallOption::default())
    }

    pub fn delete_opt(&self, req: &super::gerdu::DeleteRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<super::gerdu::DeleteResponse> {
        self.client.unary_call(&METHOD_GERDU_DELETE, req, opt)
    }

    pub fn delete(&self, req: &super::gerdu::DeleteRequest) -> ::grpcio::Result<super::gerdu::DeleteResponse> {
        self.delete_opt(req, ::grpcio::CallOption::default())
    }

    pub fn delete_async_opt(&self, req: &super::gerdu::DeleteRequest, opt: ::grpcio::CallOption) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::DeleteResponse>> {
        self.client.unary_call_async(&METHOD_GERDU_DELETE, req, opt)
    }

    pub fn delete_async(&self, req: &super::gerdu::DeleteRequest) -> ::grpcio::Result<::grpcio::ClientUnaryReceiver<super::gerdu::DeleteResponse>> {
        self.delete_async_opt(req, ::grpcio::CallOption::default())
    }
    pub fn spawn<F>(&self, f: F) where F: ::futures::Future<Output = ()> + Send + 'static {
        self.client.spawn(f)
    }
}

pub trait Gerdu {
    fn put(&mut self, ctx: ::grpcio::RpcContext, req: super::gerdu::PutRequest, sink: ::grpcio::UnarySink<super::gerdu::PutResponse>);
    fn get(&mut self, ctx: ::grpcio::RpcContext, req: super::gerdu::GetRequest, sink: ::grpcio::UnarySink<super::gerdu::GetResponse>);
    fn delete(&mut self, ctx: ::grpcio::RpcContext, req: super::gerdu::DeleteRequest, sink: ::grpcio::UnarySink<super::gerdu::DeleteResponse>);
}

pub fn create_gerdu<S: Gerdu + Send + Clone + 'static>(s: S) -> ::grpcio::Service {
    let mut builder = ::grpcio::ServiceBuilder::new();
    let mut instance = s.clone();
    builder = builder.add_unary_handler(&METHOD_GERDU_PUT, move |ctx, req, resp| {
        instance.put(ctx, req, resp)
    });
    let mut instance = s.clone();
    builder = builder.add_unary_handler(&METHOD_GERDU_GET, move |ctx, req, resp| {
        instance.get(ctx, req, resp)
    });
    let mut instance = s;
    builder = builder.add_unary_handler(&METHOD_GERDU_DELETE, move |ctx, req, resp| {
        instance.delete(ctx, req, resp)
    });
    builder.build()
}
