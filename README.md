![Build status](https://github.com/Towards-Entropy/GoTowardsEntropy/actions/workflows/go.yml/badge.svg)

# Go Towards Entropy

## Subtitle

GoTowardsEntropy is middleware for custom dictionary compression using Zstandard in Go. Configure the library with dictionaries you have trained as well as path matching for those dictionaries. If you then use this library on both sides of a network request, you will get transparent compression with your custom dictionaries.

## Custom Dictionaries

Zstandard (and other compression systems) allow you to train and utilize custom dictionaries. This dictionaries can dramatically reduce the size of compressed data. The only requirement is that the same dictionary is used for compression and decompression.

Train Zstandard dictionaires like so:

`zstd -q 19 --train training_set/* -o dictionary_name.dict`

or, if your training set is a single large file, you can do something like:

`zstd -q 19 --train training_set/big_file -B10kb -o dictionary_name.dict`

## Install

## Usage

### Configuration

Before using GoTowardsEntropy, make sure you configure the library. Here is an example of how to configure and initialize the library.

```
cfg := towardsentropy.Config{
  CompressionLevel:    5,
  BufferSize:          1024,
  DictionaryDirectory: "../../../testdata/dictionaries",
  DictionaryMatchMap:  map[string]string{"*": "supply_chain"},
}
towardsentropy.InitWithStruct(cfg)
```

You _must_ `InitWithStruct` before using the library or you will get default configuration. You can see the full set of configuration options in towardsentropy/config.go.

### HTTP Middleware

GoTowardsEntropy supports HTTP middleware that allows you to wrap handlers and requests to get transparent compression. As long as this middleware is used on both sides of a request, you will be using dictionary compression!

#### HTTP Handler

Using the handler is a simple matter of wrapping an existing handler with `towardsentropy.NewTowardsEntropyHandler`

```
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    io.WriteString(w, "Hello, World")
})
compressedHandler := towardsentropy.NewTowardsEntropyHandler(handler)
http.Handle("/", compressedHandler)
```

#### HTTP Transport

Simply wrap whatever transport you are currently using with `towardsentropy.NewTowardsEntropyTransport`, then use that as a normal transport.

```
transport := towardsentropy.NewTowardsEntropyTransport(http.DefaultTransport)
client := &http.Client{Transport: transport}
```

### Direct Compression

GoTowardsEntropy also supports usage directly via the `towardsentropy.Compress` and `towardsentropy.Decompress` calls.

Compression:

```
b := dataToCompress()
var compressed bytes.Buffer
reader := bytes.NewReader(b)
err := towardsentropy.Compress(reader, &compressed, "dictionary_id")
```

Decompression:

```
b := compressedData()
var decompressed bytes.Buffer
reader := bytes.NewReader(b)
err := towardsentropy.Compress(reader, &decompressed, "dictionary_id")
```


## Examples

There are a bunch of example usages of GoTowardsEntropy in the `/examples` dir. You should be able to copy-paste code from them to get started!

## Testing and Development

Patches are welcome! Issues are welcome!

To run unit tests in this repo: `cd towardsentropy && go test . -v`.

## Licence

[Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/)
