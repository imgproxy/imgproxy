using System;
using System.Collections.Generic;
using System.Linq;
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
            const string Url = "http://img.example.com/pretty/image.jpg";

            const string Resize = "fill";
            const int Width = 300;
            const int Height = 300;
            const string Gravity = "no";
            const int Enlarge = 1;
            const string Extension = "png";

            var url = GenerateUrl(Key, Salt, Url, Resize, Width, Height, Gravity, Enlarge, Extension);

            Console.WriteLine(url);
        }

        static string GenerateUrl(string key, string salt, string url, string resize, int width, int height, string gravity, int enlarge, string extension)
        {
            var keybin = StringToByteArray(key);
            var saltBin = StringToByteArray(salt);

            var encodedUrl = string.Join("/", WholeChunks(Convert.ToBase64String(Encoding.UTF8.GetBytes(url)).TrimEnd('='), 16));

            var path = $"/{resize}/{width}/{height}/{gravity}/{enlarge}/{encodedUrl}.{extension}";

            using (var hmac = new HMACSHA256(keybin))
            {
                var hash = hmac.ComputeHash(saltBin.Concat(Encoding.UTF8.GetBytes(path)).ToArray());
                var urlSafeBase64 = Convert.ToBase64String(hash).TrimEnd('=').Replace('+', '-').Replace('/', '_');
                return $"/{urlSafeBase64}{path}";
            }
        }

        static byte[] StringToByteArray(string hex)
        {
            return Enumerable.Range(0, hex.Length)
                             .Where(x => x % 2 == 0)
                             .Select(x => Convert.ToByte(hex.Substring(x, 2), 16))
                             .ToArray();
        }

        static IEnumerable<string> WholeChunks(string str, int maxChunkSize)
        {
            for (int i = 0; i < str.Length; i += maxChunkSize)
                yield return str.Substring(i, Math.Min(maxChunkSize, str.Length - i));
        }
    }
}
