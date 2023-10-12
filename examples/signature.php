<?php

$key = '943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881';
$salt = '520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5';

$keyBin = pack("H*" , $key);
if(empty($keyBin)) {
	die('Key expected to be hex-encoded string');
}

$saltBin = pack("H*" , $salt);
if(empty($saltBin)) {
	die('Salt expected to be hex-encoded string');
}

$path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg";

$signature = rtrim(strtr(base64_encode(hash_hmac('sha256', $saltBin.$path, $keyBin, true)), '+/', '-_'), '=');

print(sprintf("/%s%s", $signature, $path));
