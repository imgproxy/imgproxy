use std::process;

use base64; // https://crates.io/crates/base64
use hex::{self, FromHexError}; // https://crates.io/crates/hex
use hmac::{Hmac, Mac, NewMac}; // https://crates.io/crates/hmac
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
    let img = Image {
        src: "http://img.example.com/pretty/image.jpg",
        width: 100,
        height: 80,
        dpr: 2,
        ext: "webp",
    };
    match sign_url(img) {
        Ok(url) => {
            println!("{}", url);
            process::exit(0);
        }
        Err(err) => {
            eprintln!("{:#?}", err);
            process::exit(1);
        }
    }
}

pub fn sign_url(img: Image) -> Result<String, Error> {
    let url = format!(
        "/rs:{resize_type}:{width}:{height}:{enlarge}:{extend}/dpr:{dpr}/{src}.{ext}",
        resize_type = "auto",
        width = img.width,
        height = img.height,
        enlarge = 0,
        extend = 0,
        dpr = img.dpr,
        src = base64::encode_config(img.src.as_bytes(), base64::URL_SAFE_NO_PAD),
        ext = img.ext,
    );
    let decoded_key = match hex::decode(KEY) {
        Ok(key) => key,
        Err(err) => return Err(Error::InvalidKey(err)),
    };
    let decoded_salt = match hex::decode(SALT) {
        Ok(salt) => salt,
        Err(err) => return Err(Error::InvalidSalt(err)),
    };
    let mut hmac = HmacSha256::new_varkey(&decoded_key).unwrap();
    hmac.update(&decoded_salt);
    hmac.update(url.as_bytes());
    let signature = hmac.finalize().into_bytes();
    let signature = base64::encode_config(signature.as_slice(), base64::URL_SAFE_NO_PAD);
    let signed_url = format!("/{signature}{url}", signature = signature, url = url);

    Ok(signed_url)
}
