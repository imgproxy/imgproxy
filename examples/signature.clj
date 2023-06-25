(ns imgproxy-test
  (:require
   [clojure.test :refer [deftest is]])
  (:import
   (java.util Base64)
   (javax.crypto Mac)
   (javax.crypto.spec SecretKeySpec)))

(defn hex-string-to-byte-array
  [hex]
  (let [res (byte-array (/ (count hex) 2))]
    (dotimes [i (count res)]
      (aset res
            i
            (unchecked-byte (bit-or (bit-shift-left (Character/digit (nth hex (* i 2)) 16) 4)
                                    (Character/digit (nth hex (+ (* i 2) 1)) 16)))))
    res))

(defn sign-path
  [key salt path]
  (let [mac (doto (Mac/getInstance "HmacSHA256") (.init (SecretKeySpec. key "HmacSHA256")))
        _ (.update mac salt)
        hash (.doFinal mac (.getBytes path "UTF-8"))
        encoded-hash (.. (Base64/getUrlEncoder) withoutPadding (encodeToString hash))]
    (str "/" encoded-hash path)))

(deftest test-with-hmac-base64-img-proxy-test
  (is
   (=
    "/m3k5QADfcKPDj-SDI2AIogZbC3FlAXszuwhtWXYqavc/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg"
    (sign-path (hex-string-to-byte-array "943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881")
               (hex-string-to-byte-array "520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5")
               "/rs:fit:300:300/plain/http://img.example.com/pretty/image.jpg"))))
