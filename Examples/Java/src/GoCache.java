import java.io.*;
import java.net.*;

public class GoCache {
    public static void main(String[] args) throws IOException {
        String hostname = "http://localhost";
        String port = "8080";

        URL url = new URL(hostname + ":" + port + "/cache/Hello/World");
        HttpURLConnection connection = (HttpURLConnection) url.openConnection();
        connection.setDoOutput(true);
        connection.setRequestMethod("POST");
        connection.getInputStream();

        url = new URL(hostname + ":" + port + "/cache/Hello");
        connection = (HttpURLConnection) url.openConnection();
        connection.setDoOutput(true);
        connection.setRequestMethod("GET");

        StringBuffer value = new StringBuffer();
        BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(connection.getInputStream()));
        String line;
        while ((line = bufferedReader.readLine()) != null) {
            value.append(line);
        }
        bufferedReader.close();

        System.out.println(String.format("Hello = %s ", value.toString()));
    }
}
