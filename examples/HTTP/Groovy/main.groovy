def url = new URL('http://localhost:8080/cache/Hello')
def http = url.openConnection()
http.setDoOutput(true)
http.setRequestMethod('PUT')

def out = new OutputStreamWriter(http.outputStream)
out.write('World')
out.flush()
http.inputStream.read()

http = url.openConnection()
value = http.inputStream.readLines()
print ("Hello = " + value.get(0))