# GoTowardsEntropyExample: Static File Server

This is an example static file server and client using GoTowardsEntropy.

The server hosts files from `./files` on localhost:8080. The client iterates over the files and requests them from the server.

## Running

Start the server: `cd server && go run .`

Run the client: `cd client && go run .`

## Verifying compression ratios

If you want to verify that GoTowardsEntropy is improving compression ratios (eg that custom dictionaries make a difference for your data), you should look to the `standalone` example.

GoTowardsEntropy (like any reasonable compression system) uses streaming compression. This means that the server can't accurately set the Content-Length header - data is streaming back to the client before the compression is completed. However, if you want to be very sure, you can use wireshark or similar tools to inspect localhost traffic to see the transported data size.

## Notes

This example demonstrates how to use the GoTowardsEntropy library. If you are interested in using this library yourself, here are a few key things to look for:

1. The config initialization. Both client and server must initialize the config. Without this initialization, GoTowardsEntropy will look for dictionaries at `./dictionaries`. Of course, requests will still work and be compressed with Zstandard, but they won't be compressed with a dictionary.
2. Clients making a request should use the transport generated from `NewDecompressingTransport`. In most cases, you should follow the example and create an `http.Client` object using the transport. This client can then be passed around or be stored in an made accessible location. Any request using this client will go through the GoTowardsEntropy compression system.
3. Webservers handling requests should call `NewCompressionHandler` to generate a handler. This handler will wrap your logic and perform compression for you transparently. No need to make any changes to your logic to support this.

