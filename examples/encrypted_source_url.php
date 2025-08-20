<?php

declare(strict_types=1);

$key = hex2bin('1eb5b0e971ad7f45324c1bb15c947cb207c43152fa5c6c7f35c4f36e0c18e0f1');

function encrypt(string $target, string $key): string
{
    $method = 'aes-256-cbc';
    $block_size = 16;
    $iv = openssl_random_pseudo_bytes($block_size);
    $pad_length = $block_size - (strlen($target) % $block_size);
    $padded_url = $target.str_repeat(chr($pad_length), $pad_length);
    $encrypted = openssl_encrypt(
        data: $padded_url,
        cipher_algo: $method,
        passphrase: $key,
        options: OPENSSL_RAW_DATA | OPENSSL_ZERO_PADDING,
        iv: $iv,
    );

    return \rtrim(\strtr(\base64_encode($iv.$encrypted), '+/', '-_'), '=');
}
function decrypt(string $encrypted_url, string $key): string
{
    $blockSize = 16;
    $raw = base64_decode(
        strtr(
            $encrypted_url,
            '-_',
            '+/',
        ),
        true,
    );

    $iv = substr($raw, 0, $blockSize);
    $encrypted_data = substr($raw, $blockSize);

    $decrypted = openssl_decrypt(data: $encrypted_data, cipher_algo: 'aes-256-cbc', passphrase: $key, options: OPENSSL_RAW_DATA | OPENSSL_ZERO_PADDING, iv: $iv);

    $len = strlen($decrypted);
    if ($len === 0) {
        throw new InvalidArgumentException('Input data is empty.');
    }

    $pad = ord($decrypted[$len - 1]);

    if ($pad < 1 || $pad > $blockSize) {
        throw new UnexpectedValueException('Invalid padding value.');
    }

    // verify that all padding bytes are the same
    if (substr($decrypted, -$pad) !== str_repeat(chr($pad), $pad)) {
        throw new UnexpectedValueException('Invalid PKCS#7 padding.');
    }

    return substr($decrypted, 0, $len - $pad);
}

$url = 'http://img.example.com/pretty/image.jpg';
$encrypted_url = encrypt($url, $key);

// We don't sign the URL in this example but it is highly recommended to sign
// imgproxy URLs when imgproxy is being used in production.
// Signing URLs is especially important when using encrypted source URLs to
// prevent a padding oracle attack
$path = "/unsafe/rs:fit:300:300/enc/{$encrypted_url}.jpg";

echo $path;

$dUrl = decrypt($encrypted_url, $key);

assert($dUrl === $url);
