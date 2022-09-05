package imgproxytest;

import org.junit.jupiter.api.Test;

import java.util.Base64;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

import static org.junit.jupiter.api.Assertions.*;

public class ImgProxy{

    @Test
    public void testWithJavaHmacApacheBase64ImgProxyTest() throws Exception {
        byte[] key = hexStringToByteArray("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881");
        byte[] salt = hexStringToByteArray("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5");

        String path = "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg";

        String pathWithHash = signPath(key, salt, path);

        assertEquals("/m3k5QADfcKPDj-SDI2AIogZbC3FlAXszuwhtWXYqavc/rs:fit:300:300/plain/http://img.example.com/pretty/image.jp", pathWithHash);
    }

    public static String signPath(byte[] key, byte[] salt, String path) throws Exception {
        final String HMACSHA256 = "HmacSHA256";

        Mac sha256HMAC = Mac.getInstance(HMACSHA256);
        SecretKeySpec secretKey = new SecretKeySpec(key, HMACSHA256);
        sha256HMAC.init(secretKey);
        sha256HMAC.update(salt);

        String hash = Base64.getUrlEncoder().withoutPadding().encodeToString(sha256HMAC.doFinal(path.getBytes()));

        return "/" + hash + path;
    }

    private static byte[] hexStringToByteArray(String hex){
        if (hex.length() % 2 != 0)
            throw new IllegalArgumentException("Even-length string required");
        byte[] res = new byte[hex.length() / 2];
        for (int i = 0; i < res.length; i++) {
            res[i]=(byte)((Character.digit(hex.charAt(i * 2), 16) << 4) | (Character.digit(hex.charAt(i * 2 + 1), 16)));
        }
        return res;
    }
}
