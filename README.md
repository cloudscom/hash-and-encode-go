# hash-and-encode-go
A short demonstration program in Go that creates a webserver that:
1. Listens on localhost:8080/hash and accepts POST requests for "password=XXX" and then SHA512 hashes that password and saves the base64 encoded value of that hash and returns an *id* for later retrieval.
2. Listens on localhost:8080/hash/*id* and accepts GET requests to retrieve the computed and encoded hash value. Note that computed hashes are available 5 seconds after the request.
3. Listens on localhost:8080/stats and returns a JSON object containing the total number of hash requests completed and an average execution time in microseconds.
4. Listens on localhost:8080/shutdown and performs a graceful shutdown on the webserver after completing any outstanding requests.

# Testing
A short test script is included to start the webserver and compute a few hashes. Only the first hash is verified for correctness and, at the end, the webserver is shutdown.

# Running
Only standard Go libraries and a Go runtime environment are required.
Build the executable in the usual way - 'go build'
Execute the test - 'sh test' or make *test* executable and './test'
