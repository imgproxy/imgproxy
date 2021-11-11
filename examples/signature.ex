defmodule App.Imgproxy do
  @prefix "https://imgproxy.mybiz.xyz"
  @key Base.decode16!("943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881", case: :lower)
  @salt Base.decode16!("520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5", case: :lower)

  def build_url(img_url, opts) do
    path = build_path(img_url, opts)
    signature = gen_signature(path)

    Path.join([@prefix, signature, path])
  end

  defp build_path(img_url, opts) do
    Path.join([
      "/",
      "rs:#{opts.resize}:#{opts.width}:#{opts.height}:#{opts.enlarge}",
      "g:#{opts.gravity}",
      Base.url_encode64(img_url, padding: false) <> "." <> opts.extension
    ])
  end

  defp gen_signature(path) do
    :sha256
    |> :crypto.hmac(@key, @salt <> path)
    |> Base.url_encode64(padding: false)
  end
end

# Usage
#
# App.Imgproxy.build_url(
#   "https://myawesomedomain.com/raw-image.png",
#   %{resize: "fit", width: 1000, height: 400, gravity: "ce", enlarge: 0, extension: "jpg"}
# )
