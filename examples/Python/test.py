import requests

url = "http://localhost"    # The URL on which GoCache is running
port = "8080"               # The port on which GoCache is listening

# add key value pair ('Hello', World) to the cache
requests.put("%s:%s/cache/Hello" % (url, port), "World")

# get the value for 'Hello' key
response = requests.get("%s:%s/cache/Hello" % (url, port))
print("Hello =", response.text)

