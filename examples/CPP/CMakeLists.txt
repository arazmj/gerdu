cmake_minimum_required(VERSION 3.17)
project(GOCACHECPP)

include_directories(/usr/local/include)
link_directories(/usr/local/Cellar/curlpp/0.8.1/lib/)

set(CMAKE_CXX_STANDARD 11)

add_executable(GOCACHECPP main.cpp)

target_link_libraries(GOCACHECPP -lcurlpp -lcurl)
