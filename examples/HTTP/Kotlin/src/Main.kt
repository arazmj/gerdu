import java.net.HttpURLConnection
import java.net.URL

fun main() {
    val url = "http://localhost:8080/cache/Hello"
    with(URL(url).openConnection() as HttpURLConnection) {
        requestMethod = "PUT"
        doOutput = true
        outputStream.write("World".toByteArray())
        outputStream.flush()
        inputStream.read()
        val value = URL(url).readText()
        print("Hello = $value")
    }
}