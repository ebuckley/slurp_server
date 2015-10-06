package main

import (
	"bytes"
	"encoding/binary"
	"github.com/ebuckley/slurp_server/LRU"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type cachePutRequest struct {
	name       string
	fileIsSent chan bool
	file       io.Reader
}
type cacheCheckRequest struct {
	name        string
	isCached    chan *cachePutRequest
	isNotCached chan bool
}

var (
	listenPort        string
	serveDirectory    string
	maxCacheSizeBytes int = 60 * 1024 * 1024
	serverListener    net.Listener
)

func getFile(filename string, cp chan *cachePutRequest, cacheCheck chan *cacheCheckRequest, fileSent chan bool) (io.Reader, error) {
	cacheReq := new(cacheCheckRequest)
	cacheReq.name = filename
	cacheReq.isNotCached = make(chan bool)
	cacheReq.isCached = make(chan *cachePutRequest)

	//send cache check request
	cacheCheck <- cacheReq
	//handle cache request
	select {
	case c := <-cacheReq.isCached:
		log.Println("Cache hit sending", c.name, "to the client")
		return c.file, nil
	case <-cacheReq.isNotCached:
		log.Println("Cache miss sending", filename, "to the client")

		fd, err := os.Open(filename)
		if err != nil {
			return fd, err
		}

		//when cacheFileReader.Read is called, responseFileReader will be written to
		responseFileReader := new(bytes.Buffer)
		cacheFileReader := io.TeeReader(fd, responseFileReader)

		cachePut := new(cachePutRequest)
		cachePut.name = filename
		cachePut.file = responseFileReader
		cachePut.fileIsSent = fileSent
		//the cache needs to know when the file has been read so it can write it to the

		cp <- cachePut
		return cacheFileReader, nil
	}
}
func handleCache(cacheRequests chan *cacheCheckRequest, cachePuts chan *cachePutRequest) {

	lru := LRU.NewCache(maxCacheSizeBytes)

	for {
		select {
		case cr := <-cacheRequests:

			cr.isNotCached <- true
			// data, ok := lru.Get(cr.name)
			// if !ok {
			// 	cr.isNotCached <- true
			// } else {
			// 	cachePut := new(cachePutRequest)
			// 	cachePut.name = cr.name
			// 	cachePut.file = bytes.NewBuffer(data)
			// 	cr.isCached <- cachePut
			// }
		case cp := <-cachePuts:

			//block until the file has complely sent before updating the cache
			fileSent := <-cp.fileIsSent

			if fileSent == true {
				buf, err := ioutil.ReadAll(cp.file)
				if err != nil {
					log.Println("err reading cachePutRequest on file", cp.name)
				}
				lru.Push(cp.name, buf)
			}
		}
	}
}

func handleConnection(conn net.Conn, cacheRequests chan *cacheCheckRequest, cachePuts chan *cachePutRequest) {
	defer conn.Close()

	//get the filename
	req := make([]byte, 255)
	n, err := conn.Read(req)
	if err != nil {
		log.Println("Error reading request:", req)
		return
	}
	if n != 255 {
		log.Println("WTF: didn't recieve the right number of bytes :(")
		return
	}

	requestFile := string(bytes.Trim(req, "\x00")) //remove nul bytes from the read
	requestPath := serveDirectory + "/" + requestFile
	log.Printf("client %s is requesting file %s", conn.RemoteAddr(), requestPath)

	//check existance of file
	fStat, err := os.Stat(requestPath)
	if err != nil {
		log.Printf("could not stat file: %s ", err)
		// TODO if IO.notFound then respond to the client that the file can't be found
		return
	}

	if fStat.IsDir() {
		log.Println(requestPath, "is actually a directory, sending directories is not supported")
		return
	}

	// send filesize
	err = binary.Write(conn, binary.BigEndian, fStat.Size())
	if err != nil {
		log.Println("could not write filesize message", err)
		return
	}

	fileIsSent := make(chan bool)

	fd, err := getFile(requestPath, cachePuts, cacheRequests, fileIsSent)
	if err != nil {
		log.Println("could not open file:", err)
		return
	}

	//send file
	totalWrite, err := io.Copy(conn, fd)
	if err != nil {
		fileIsSent <- false
		log.Println("could not send file over connection", err)
	}
	fileIsSent <- true

	if totalWrite != fStat.Size() {
		log.Println("file didn't get sent correctly... not enought bytes where sent -_- ")
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("ERROR: we expect two arguments, invocation of server should be \"slurp_server port_to_listen_on file_directory\"")
	}

	listenPort = os.Args[2]
	serveDirectory = os.Args[1]

	directory, err := os.Open(serveDirectory)
	if err != nil {
		log.Fatalf("%s was not succesfully opened because %s", serveDirectory, err)
		os.Exit(1)
		return
	}

	fStat, err := directory.Stat()
	if err != nil {
		log.Fatalf("%s could not get file stat coz %s", serveDirectory, err)
		os.Exit(1)
		return
	}

	if fStat.IsDir() == false {
		log.Fatalf("ERROR: first argument should be a valid directory")
		os.Exit(1)
		return
	}

	// setup chanels for handling caching
	cacheRequests := make(chan *cacheCheckRequest)
	cachePuts := make(chan *cachePutRequest)

	go handleCache(cacheRequests, cachePuts)

	//set up for interupts
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//application signal listener
	go func() {
		sig := <-sigs
		log.Println("recieved signal", sig)
		done <- true
	}()

	go func() {
		serverListener, err := net.Listen("tcp", listenPort)
		if err != nil {
			log.Fatalf("server couldn't listen on %s for some reason %s", listenPort, err)
			os.Exit(1)
			done <- true
		}

		for {
			conn, err := serverListener.Accept()
			if err != nil {
				log.Println("connection acception failed: %s", err)
			}
			go handleConnection(conn, cacheRequests, cachePuts)
		}
	}()

	//cleanup before exiting
	<-done
	log.Println("gracefull shutdown, cya")
	directory.Close()

	if serverListener != nil {
		serverListener.Close()
	}
}
