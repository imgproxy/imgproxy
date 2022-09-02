use std::process;

use base64; // https://crates.io/crates/base64
use hex::{self, FromHexError}; // https://crates.io/crates/hex
use hmac::{Hmac, Mac}; // https://crates.io/crates/hmac
use sha2::Sha256; // https://crates.io/crates/sha2

type HmacSha256 = Hmac<Sha256>;

const KEY: &'static str = "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881";
const SALT: &'static str = "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5";

pub struct Image {
    pub src: &'static str,
    pub width: usize,
    pub height: usize,
    pub dpr: u8,
    pub ext: &'static str,
}

#[derive(Debug)]
pub enum Error {
    InvalidKey(FromHexError),
    InvalidSalt(FromHexError),
}

fn main() {
    let path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg";

    match sign_path(path) {
        Ok(signed_path) => {
            println!("{}", signed_path);
            process::exit(0);
        }
        Err(err) => {
            eprintln!("{:#?}", err);
            process::exit(1);
        }
    }
}

pub fn sign_path(path: &str) -> Result<String, Error> {
    let decoded_key = match hex::decode(KEY) {
        Ok(key) => key,
        Err(err) => return Err(Error::InvalidKey(err)),
    };
    let decoded_salt = match hex::decode(SALT) {
        Ok(salt) => salt,
        Err(err) => return Err(Error::InvalidSalt(err)),
    };

    let mut hmac = HmacSha256::new_from_slice(&decoded_key).unwrap();
    hmac.update(&decoded_salt);
    hmac.update(path.as_bytes());

    let signature = hmac.finalize().into_bytes();
    let signature = base64::encode_config(signature.as_slice(), base64::URL_SAFE_NO_PAD);

    let signed_path = format!("/{}{}", signature, path);

    Ok(signed_path)
}
