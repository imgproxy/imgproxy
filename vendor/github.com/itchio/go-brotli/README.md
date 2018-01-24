# Go bindings for the Brotli compression library

[![GoDoc](https://godoc.org/github.com/itchio/go-brotli?status.svg)](https://godoc.org/github.com/itchio/go-brotli)
[![Build Status](https://travis-ci.org/itchio/go-brotli.svg)](https://travis-ci.org/itchio/go-brotli)

This is a fork of Mike Houston's <https://github.com/kothar/brotli-go> with
the following changes:

  * Bumped to a more recent upstream (post-1.0.0)
  * Removed custom dictionary support (which was removed from upstream)

See <https://github.com/google/brotli> for the upstream C/C++ source, and
the `VERSION.md` file to find out the currently vendored version.

### Usage

Instead of including potentially-outdated examples in the README, 
please refer to the `Examples` tests on the following godoc pages:

  * Decompression: <https://godoc.org/github.com/itchio/go-brotli/dec>
  * Compression: <https://godoc.org/github.com/itchio/go-brotli/enc>

### Bindings

This is a very basic Cgo wrapper for the enc and dec directories from the Brotli sources.

A few minor changes have been made to get things working with Go:

1. The default dictionary has been extracted to a separate 'common' package to allow linking the enc and dec cgo modules if you use both. Otherwise there are duplicate symbols, as described in the dictionary.h header files.

2. The dictionary variable name for the dec package has been modified for the same reason, to avoid linker collisions.

### Links

  * original bindings: <https://github.com/kothar/brotli-go>
  * upstream cgo bindings (requires separate library compilation): <https://github.com/google/brotli/tree/master/go/cbrotli>
  * brotli streaming decompression written in pure go: <https://github.com/dsnet/compress>

### License

Brotli and these bindings are open-sourced under the MIT License - see the LICENSE file.
