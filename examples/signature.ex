defmodule App.Imgproxy do
  @key Base.decode16!("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881", case: :lower)
  @salt Base.decode16!("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5", case: :lower)

  def sign_path(path) do
    signature = gen_signature(path)
    Path.join(["/", signature, path])
  end

  defp gen_signature(path) do
    :sha256
    |> :crypto.hmac(@key, @salt <> path)
    |> Base.url_encode64(padding: false)
  end
end

# Usage
#
# App.Imgproxy.sign_path(
#   "/rs:fit:300:300/plain/https://myawesomedomain.com/raw-image.png"
# )
