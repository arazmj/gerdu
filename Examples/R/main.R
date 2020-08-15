url <- "http://localhost:8080/cache/Hello"
httpPUT(url, "World")
response <- httpGET(url)
cat("Hello =", response[1])