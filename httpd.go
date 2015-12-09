package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
)

const SERVING = 5

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

/* sort files by size. */
type BySize []os.FileInfo

func (files BySize) Len() int {
	return len(files)
}

func (files BySize) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}

func (files BySize) Less(i, j int) bool {
	return files[i].Size() > files[j].Size()
}

func get_local_ipaddr() string {
	// send udp data to a server, so we can find the ipaddr in real use.
	conn, _ := net.Dial("udp", "163.com:9")
	host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	return host
}

func generate_random_port() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int(math.Floor(random.Float64() * (65535 - 1024)))
}

func find_files_to_serve() []string {
	// read all files under current directory. calculating default files
	// to be served over http.
	var count = 0
	serving := make([]string, 0)
	currentFiles, _ := ioutil.ReadDir("./")
	sort.Sort(BySize(currentFiles))
	for _, file := range currentFiles {
		if !file.IsDir() {
			serving = append(serving, file.Name())
		}
		count += 1
		if count >= SERVING {
			break
		}
	}
	return serving
}

/* 处理命令行参数, 起http服务的逻辑. */
func main() {
	serving := make([]string, 0)
	ipaddr := get_local_ipaddr()
	var address string

	var opts struct {
		Verbose   bool     `short:"v" long:"verbose" description:"Show verbose debug information"`
		Port      int      `short:"p" default:"0" long:"port" default-mask:"a random port" description:"Specify listen port manually."`
		Bind      string   `short:"b" default:"0.0.0.0" long:"bind" description:"host to bind to."`
		Filepaths []string `short:"f" long:"filename" description:"Files to be served"`
	}
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(0) // print help message and exit.
	}

	if opts.Port == 0 {
		// generate a random number to serve as http port if port not specified
		_port := generate_random_port()
		opts.Port = _port
	}
	address = ipaddr + ":" + strconv.Itoa(opts.Port)

	if len(opts.Filepaths) == 0 {
		opts.Filepaths = find_files_to_serve()
	}

	for _, file := range opts.Filepaths {
		serving = append(serving, string(file))
	}
	fmt.Println("Listening in " + address + "\n")

	for _, file := range serving {
		http.HandleFunc("/"+file, DownloadFile)
		fmt.Println("link: http://" + address + "/" + url.QueryEscape(file))
	}
	http.ListenAndServe(address, nil)
}
