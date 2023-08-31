# GoTowardsEntropyExample: Upload Server

This is an example of a server hosting uploads. The key demonstration here is using the transport for dictionary compression of write operations (POST/PATCH/PUT) and the handler for decompression.

## Running

Start the server: `cd server && go run .`

Run the client: `cd client && go run .`
