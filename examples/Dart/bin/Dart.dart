import 'package:http/http.dart' as http;

void main(List<String> arguments) {
  var uri = Uri.encodeFull('http://localhost:8080/cache/Hello');
  http.put(uri, body: 'World')
      .then((_) {
        http.get(uri).then((value) => print('Hello = ${value.body}'));
  });
}
