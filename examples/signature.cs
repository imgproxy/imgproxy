using System;
using System.Collections.Generic;
using System.IO;
using System.Security.Cryptography;
using System.Text;

namespace ImgProxy.Examples
{
    class Program
    {
        static void Main(string[] args)
        {
            const string Key = "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881";
            const string Salt = "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5";

            var path = SignerHelper.SignPath(Key, Salt, "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg");

            Console.WriteLine(url);
        }
    }

    public static class SignerHelper
    {
        public static string SignPath(string key, string salt, string path)
        {
            var keybin = HexadecimalStringToByteArray(key);
            var saltBin = HexadecimalStringToByteArray(salt);

            var passwordWithSaltBytes = new List<byte>();
            passwordWithSaltBytes.AddRange(saltBin);
            passwordWithSaltBytes.AddRange(Encoding.UTF8.GetBytes(path));

            using var hmac = new HMACSHA256(keybin);
            byte[] digestBytes = hmac.ComputeHash(passwordWithSaltBytes.ToArray());
            var urlSafeBase64 = EncodeBase64URLSafeString(digestBytes);
            return $"/{urlSafeBase64}{path}";
        }

        static byte[] HexadecimalStringToByteArray(string input)
        {
            var outputLength = input.Length / 2;
            var output = new byte[outputLength];
            using (var sr = new StringReader(input))
            {
                for (var i = 0; i < outputLength; i++)
                    output[i] = Convert.ToByte(new string(new char[2] { (char)sr.Read(), (char)sr.Read() }), 16);
            }
            return output;
        }

        static string EncodeBase64URLSafeString(this byte[] stream)
            => Convert.ToBase64String(stream).TrimEnd('=').Replace('+', '-').Replace('/', '_');

        static string EncodeBase64URLSafeString(this string str)
            => EncodeBase64URLSafeString(Encoding.UTF8.GetBytes(str));
    }
}
