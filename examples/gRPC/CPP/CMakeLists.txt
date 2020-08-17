cmake_minimum_required(VERSION 3.18)
project(GerduGRPCCPP)

include_directories("/usr/local/Cellar/grpc/1.30.2/include")
link_directories("/usr/local/Cellar/grpc/1.30.2/lib")

find_package(Protobuf REQUIRED )
message(STATUS "Using protobuf ${protobuf_VERSION}")
include_directories(${PROTOBUF_INCLUDE_DIR})

set(CMAKE_CXX_STANDARD 14)

add_executable(GerduGRPCCPP main.cpp proto/gerdu.grpc.pb.cc proto/gerdu.pb.cc)

target_link_libraries(GerduGRPCCPP grpc++ ${PROTOBUF_LIBRARY})
