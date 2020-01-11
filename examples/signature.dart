import 'dart:convert';
import 'package:convert/convert.dart';
import 'package:crypto/crypto.dart';

void main() {
	var key = hex.decode("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881"); 
	var salt = hex.decode("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5"); 
	var url = "http://img.example.com/pretty/image.jpg"; 
	var resizing_type = 'fill';
	var width = 300;
	var height = 300;
	var gravity = 'no';
	var enlarge = 1;
	var extension = 'png';
  
	var url_encoded = urlSafeBase64(utf8.encode(url)); 

	var path = "/$resizing_type/$width/$height/$gravity/$enlarge/$url_encoded.$extension"; 

	var signature = sign(salt, utf8.encode(path), key);
	print("/$signature/$path"); 
}

String urlSafeBase64(buffer) {
	return base64.encode(buffer).replaceAll("=", "").replaceAll("+", "-").replaceAll("/", "_"); 
}

String sign(salt, path, key) {
	var hmac = Hmac(sha256, key); 
	var digest = hmac.convert(salt + path); 
	return urlSafeBase64(digest.bytes); 
}