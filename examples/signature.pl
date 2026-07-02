use v5.12;
use warnings;
use Digest::SHA qw(hmac_sha256);
use MIME::Base64 qw(encode_base64url);

my $key = pack 'H*', '943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881';
my $salt = pack 'H*', '520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5';

my $path = '/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg';

my $hmac = hmac_sha256($salt . $path, $key);
my $signature = encode_base64url($hmac);

say "/$signature$path";
