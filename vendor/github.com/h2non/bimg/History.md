
## v1.0.14 / 2017-09-12

  * Merge pull request #192 from greut/trim
  * Adding trim operation.
  * Merge pull request #191 from greut/alpha4
  * Update 8.6 to alpha4.

## v1.0.13 / 2017-09-11

  * Merge pull request #190 from greut/typos
  * Fix typo and small cleanup.

## v1.0.12 / 2017-09-10

  * Merge branch '99designs-vips-reduce'
  * fix(reduce): resolve conflicts with master
  * Use vips reduce when downscaling

## v1.0.11 / 2017-09-10

  * feat(#189): allow strip image metadata via bimg.Options.StripMetadata = bool
  * fix(resize): code format issue
  * refactor(resize): add Go version comment
  * refactor(tests): fix minor code formatting issues
  * fix(#162): garbage collection fix. split Resize() implementation for Go runtime specific
  * feat(travis): add go 1.9
  * Merge pull request #183 from greut/autorotate
  * Proper handling of the EXIF cases.
  * Merge pull request #184 from greut/libvips858
  * Merge branch 'master' into libvips858
  * Merge pull request #185 from greut/libvips860
  * Add libvips 8.6 pre-release
  * Update to libvips 8.5.8
  * fix(resize): runtime.KeepAlive is only Go
  * fix(#159): prevent buf to be freed by the GC before resize function exits
  * Merge pull request #171 from greut/fix-170
  * Check the length before jumping into buffer.
  * Merge pull request #168 from Traum-Ferienwohnungen/icc_transform
  * Add option to convert embedded ICC profiles
  * Merge pull request #166 from danjou-a/patch-1
  * Fix Resize verification value
  * Merge pull request #165 from greut/libvips846
  * Testing using libvips8.4.6 from Github.

## v1.0.10 / 2017-06-25

  * Merge pull request #164 from greut/length
  * Add Image.Length()
  * Merge pull request #163 from greut/libvips856
  * Run libvips 8.5.6 on Travis.
  * Merge pull request #161 from henry-blip/master
  * Expose vips cache memory management functions.
  * feat(docs): add watermark image note in features

## v1.0.9 / 2017-05-25

  * Merge pull request #156 from Dynom/SmartCropToGravity
  * Adding a test, verifying both ways of enabling SmartCrop work
  * Merge pull request #149 from waldophotos/master
  * Replacing SmartCrop with a Gravity option
  * refactor(docs): v8.4
  * Change for older LIBVIPS versions. `vips_bandjoin_const1` is added in libvips 8.2.
  * Second try, watermarking memory issue fix

## v1.0.8 / 2017-05-18

  * Merge pull request #145 from greut/smartcrop
  * Merge pull request #155 from greut/libvips8.5.5
  * Update libvips to 8.5.5.
  * Adding basic smartcrop support.
  * Merge pull request #153 from abracadaber/master
  * Added Linux Mint 17.3+ distro names
  * feat(docs): add new maintainer notice (thanks to @kirillDanshin)
  * Merge pull request #152 from greut/libvips85
  * Download latest version of libvips from github.
  * Merge pull request #147 from h2non/revert-143-master
  * Revert "Fix for memory issue when watermarking images"
  * Merge pull request #146 from greut/minor-major
  * Merge pull request #143 from waldophotos/master
  * Merge pull request #144 from greut/go18
  * Fix tests where minor/major were mixed up
  * Enabled go 1.8 builds.
  * Fix the unref of images, when image isn't transparent
  * Fix for memory issue when watermarking images
  * feat(docs): add maintainers sections
  * Merge pull request #132 from jaume-pinyol/WATERMARK_SUPPORT
  * Add support for image watermarks
  * Merge pull request #131 from greut/versions
  * Running tests on more specific versions.
  * refactor(preinstall.sh): remove deprecation notice
  * Update preinstall.sh
  * fix(requirements): required libvips 7.42
  * fix(History): typo
  * chore(History): add breaking change note

## v1.0.7 / 13-01-2017

- fix(#128): crop image calculation for missing width or height axis.
- feat: add TIFF save output format (**note**: this introduces a minor interface breaking change in `bimg.IsImageTypeSupportedByVips` auxiliary function).

## v1.0.6 / 12-11-2016

- feat(#118): handle 16-bit PNGs.
- feat(#119): adds JPEG2000 file for the type tests.
- feat(#121): test bimg against multiple libvips versions.

## v1.0.5 / 01-10-2016

- feat(#92): support Extend param with optional background.
- fix(#106): allow image area extraction without explicit x/y axis.
- feat(api): add Extend type with `libvips` enum alias.

## v1.0.4 / 29-09-2016

- fix(#111): safe check of magick image type support.

## v1.0.3 / 28-09-2016

- fix(#95): better image type inference and support check.
- fix(background): pass proper background RGB color for PNG image conversion.
- feat(types): validate supported image types by current `libvips` compilation.
- feat(types): consistent SVG image checking.
- feat(api): add public functions `VipsIsTypeSupported()`, `IsImageTypeSupportedByVips()` and `IsSVGImage()`.

## v1.0.2 / 27-09-2016

- feat(#95): support GIF, SVG and PDF formats.
- fix(#108): auto-width and height calculations now round instead of floor.

## v1.0.1 / 22-06-2016

- fix(#90): Do not not dereference the original image a second time.

## v1.0.0 / 21-04-2016

- refactor(api): breaking changes: normalize public members to follow Go naming idioms.
- feat(version): bump to major version. API contract won't be compromised in `v1`.
- feat(docs): add missing inline godoc documentation.
