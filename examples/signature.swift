//
// You also need a Bridging-Header.h to import CommonCrypto:
//   #import <CommonCrypto/CommonCrypto.h>
//

import Foundation

// https://stackoverflow.com/a/41965688/326257
extension Data {
    func hmac256(key: Data) -> String {
        let cKey = (key as NSData).bytes
        let cData = (self as NSData).bytes
        var result = [CUnsignedChar](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        CCHmac(CCHmacAlgorithm(kCCHmacAlgSHA256), cKey, key.count, cData, self.count, &result)
        let hmacData:NSData = NSData(bytes: result, length: (Int(CC_SHA256_DIGEST_LENGTH)))
        return customBase64(input: hmacData as Data)
    }
}

// https://stackoverflow.com/a/26502285/326257
extension String {
    func hexadecimal() -> Data? {
        var data = Data(capacity: characters.count / 2)

        let regex = try! NSRegularExpression(pattern: "[0-9a-f]{1,2}", options: .caseInsensitive)
        regex.enumerateMatches(in: self, range: NSMakeRange(0, utf16.count)) { match, flags, stop in
            let byteString = (self as NSString).substring(with: match!.range)
            var num = UInt8(byteString, radix: 16)!
            data.append(&num, count: 1)
        }

        guard data.count > 0 else { return nil }

        return data
    }

}

func customBase64(input: Data) -> String {
    return input.base64EncodedString()
        .replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_")
        .replacingOccurrences(of: "=+$", with: "", options: .regularExpression)
}

let key = "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881".hexadecimal()!;
let salt = "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5".hexadecimal()!;

let partialPath = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg"
let toSign = salt + partialPath.utf8

let signature = toSign.hmac256(key: key)

let path = "/\(signature)\(partialPath)"
print(path)
