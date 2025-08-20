<?php

$key = hex2bin('1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1');

function encrypt(string $target, string $key): string
{
    $method = 'aes-256-cbc';
    $block_size = 16;
    $iv = openssl_random_pseudo_bytes($block_size);
    $pad_length = $block_size - (strlen($target) % $block_size);
    $padded_url = $target . str_repeat(chr($pad_length), $pad_length);
    $encrypted = openssl_encrypt(
        data: $padded_url,
        cipher_algo: $method,
        passphrase: $key,
        options: OPENSSL_RAW_DATA | OPENSSL_ZERO_PADDING,
        iv: $iv
    );

    return \rtrim(\strtr(\base64_encode($iv . $encrypted), '+/', '-_'), '=');
}

$url = 'http://img.example.com/pretty/image.jpg';
$encrypted_url = encrypt($url, $key);

// We don't sign the URL in this example but it is highly recommended to sign
// imgproxy URLs when imgproxy is being used in production.
// Signing URLs is especially important when using encrypted source URLs to
// prevent a padding oracle attack
$path = "/unsafe/rs:fit:300:300/enc/{$encrypted_url}.jpg";

echo $path;