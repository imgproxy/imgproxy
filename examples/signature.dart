import 'dart:convert';
import 'package:convert/convert.dart';
import 'package:crypto/crypto.dart';

void main() {
	var key = hex.decode("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881");
	var salt = hex.decode("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5");

	var path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg";

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
