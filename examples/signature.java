package imgproxytest;

import com.amazonaws.util.Base16;
import org.apache.commons.codec.binary.Base64;
import org.junit.Assert;
import org.junit.Test;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

public class imgproxy {

    @Test
    public void testWithJavaHmacApacheBase64ImgProxyTest() throws Exception {

        String Key = "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881";
        String Salt = "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5";
        String Url = "http://img.example.com/pretty/image.jpg";

        String Resize = "fill";
        int Width = 300;
        int Height = 300;
        String Gravity = "no";
        int Enlarge = 1;
        String Extension = "png";
        String urlWithHash = GenerateSignedUrlImgProxy(Key, Salt, Url, Resize, Width, Height, Gravity, Enlarge, Extension);

        Assert.assertEquals("/_PQ4ytCQMMp-1w1m_vP6g8Qb-Q7yF9mwghf6PddqxLw/fill/300/300/no/1/aHR0cDovL2ltZy5leGFtcGxlLmNvbS9wcmV0dHkvaW1hZ2UuanBn.png", urlWithHash);
    }


    public static String GenerateSignedUrlImgProxy(String key, String salt, String url, String resize, int width, int height, String gravity, int enlarge, String extension) throws Exception {
        final String HMACSHA256 = "HmacSHA256";

        byte[] keybin = Base16.decode(key);
        byte[] saltBin = Base16.decode(salt);

        String encodedUrl = Base64.encodeBase64URLSafeString(url.getBytes());

        String path = "/" + resize + "/" + width + "/" + height + "/" + gravity + "/" + enlarge + "/" + encodedUrl + "." + extension;

        Mac sha256_HMAC = Mac.getInstance(HMACSHA256);
        SecretKeySpec secret_key = new SecretKeySpec(keybin, HMACSHA256);
        sha256_HMAC.init(secret_key);
        sha256_HMAC.update(saltBin);

        String hash = Base64.encodeBase64URLSafeString(sha256_HMAC.doFinal(path.getBytes()));

        return "/" + hash + path;
    }
}
