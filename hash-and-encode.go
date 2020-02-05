package main

import (
    "os"
    "fmt"
    "sync"
    "encoding/json"
    "encoding/base64"
    "crypto/sha512"
    "time"
    "net/http"
    "strconv"
    "log"
    "strings"
)

type hashEncode struct {
    hash string
    elapsed time.Duration
}

type pwHashes struct {
    sync.Mutex
    run bool
    hashSlice []hashEncode
}

type StatsReply struct {
    Total  int      `json:"total"`
    Average int64   `json:"average"`
}

// delay and then compute the hash
func (myHash *pwHashes) doHash(index int, password string) {
    // sleep 5 seconds
    time.Sleep(5 * time.Second)

    if password != "" {
        start := time.Now()

        // do the hash of the password
        crypt := sha512.New()
        crypt.Write([]byte(password))
        value := base64.StdEncoding.EncodeToString(crypt.Sum(nil))
        myHash.hashSlice[index].hash = value

        t := time.Now()

        // save elapsed
        myHash.hashSlice[index].elapsed = t.Sub(start)
    }
}


// Handler for "getHash" request
func (myHash *pwHashes) getHash(w http.ResponseWriter, req *http.Request) {
    if myHash.run != true {
        return
    }

    err := req.ParseForm()
    if err != nil {
        // log.Printf("Error from ParseForm %s", err)   // TEST
        // handle error
        return
    }

    switch req.Method {
    case "GET":
        // parse the url, ignore the error
        is := strings.TrimPrefix(req.URL.Path, "/hash/")
        if is == "" {
            // handle error
            // log.Printf("Null path\n")    // TEST
            return
        }

        index, err := strconv.Atoi(is)
        if err != nil {
            // handle error
            // log.Printf("Error not an int %s", is)   // TEST
            log.Println(err)
            return
        }

        // if hash is done, return it
        result := myHash.hashSlice[index].hash
        if result == "" {
            // handle error
            // log.Printf("Null result\n")  // TEST
            return
        }

        // return hash
        fmt.Fprintf(w, "%s\n", result)

    default:
        // handle error
        // log.Printf("Not a POST/GET request")     // TEST
        return
    }
}

// Handler for "hash" request
func (myHash *pwHashes) hash(w http.ResponseWriter, req *http.Request) {
    if myHash.run != true {
        return
    }

    err := req.ParseForm()
    if err != nil {
        // handle error
        // log.Printf("Error from ParseForm %s", err)   // TEST
        return
    }

    switch req.Method {
    case "POST":
        // get the password
        value := req.FormValue("password")
        if value == "" {
            // handle error
            // log.Printf("Form %v", req)   // TEST
            return
        }

        // crit start
        myHash.Lock()
        he := hashEncode{"", 0}
        myHash.hashSlice = append(myHash.hashSlice, he)
        index := len(myHash.hashSlice) - 1
        myHash.Unlock()
        // crit end

        // Use goroutine to delay and compute hash
        go myHash.doHash(index, value)

        // return the index immediately
        fmt.Fprintf(w, "%d\n", index)

    default:
        // handle error
        // log.Printf("Not a POST/GET request")     // TEST
        return
    }
}

// Handler for "stats"
func (myHash *pwHashes) stats(w http.ResponseWriter, req *http.Request) {
    if myHash.run != true {
        return
    }

    var totalElapsed, avgElapsed int64

    // no locking - read only here
    sliceLen := len(myHash.hashSlice)
    for i := 0; i < sliceLen; i++ {
        // calculate average elapsed
        if myHash.hashSlice[i].elapsed == 0 {
            // bail at first uninitialized elapsed and reset len
            sliceLen = i + 1
            break
        }
        totalElapsed += int64(myHash.hashSlice[i].elapsed)
    }

    // time math in go is awkward
    avgElapsed = totalElapsed / int64(sliceLen)

    // create JSON object with all the data
    m := StatsReply{sliceLen, avgElapsed}

    // Only exported fields will be marshalled
    // Remember the `json:"name"` hint to get lowercase
    reply, err := json.Marshal(m)
    if err != nil {
        // handle error
        return
    }

    // Return stats
    fmt.Fprintf(w, "%s\n", reply)
}

// exit
func closeMe() {
    os.Exit(0)
}

// Handler for "shutdown"
func (myHash *pwHashes) shutdown(w http.ResponseWriter, req *http.Request) {
    if myHash.run != true {
        return
    }

    myHash.run = false

    for len(myHash.hashSlice) > 0 {
        // Wait for last hash
        if myHash.hashSlice[len(myHash.hashSlice) - 1].elapsed != 0 {
            break;
        }
    }

    fmt.Fprintf(w, "Shutdown\n")

    // goroutine to allow flush
    go closeMe()
}

// main
func main() {
    myHash := new(pwHashes)
    myHash.run = true

    http.HandleFunc("/stats", myHash.stats)
    http.HandleFunc("/shutdown", myHash.shutdown)
    http.HandleFunc("/hash", myHash.hash)
    http.HandleFunc("/hash/", myHash.getHash)

    http.ListenAndServe(":8080", nil)
}

