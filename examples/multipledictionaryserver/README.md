# GoTowardsEntropyExample: Multiple Dictionary Static File Server

This is an example static file server and client using GoTowardsEntropy. Everything is identical to the `staticfileserver` example, except this one shows utilization of multiple dictionaries depending on the path of the data.

The server hosts files from `./files` on localhost:8080. The client iterates over the files and requests them from the server.

## Running

Start the server: `cd server && go run .`

Run the client: `cd client && go run .`