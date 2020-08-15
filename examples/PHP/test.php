<html>
<head>
    <title>GoCache PHP Test</title>
</head>
<body>

<?php
$url = "http://localhost:8080/cache/Hello";
$ch = curl_init($url);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "PUT");
curl_setopt($ch, CURLOPT_POSTFIELDS, "World");

curl_exec($ch);

curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "GET");
curl_setopt($ch,CURLOPT_RETURNTRANSFER,true);

$response = curl_exec($ch);

echo "Hello = " . $response
?>
</body>
</html>
