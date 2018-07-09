package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"bytes"
	"strings"
	"math/rand"
	"encoding/json"
	"net"
	"crypto/x509"
	"path/filepath"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go"
)

// opens a file in append,wronly,create,0600 mode
// or panics if that's not possible.
func openAppendOrDie(path string, log io.Writer) *os.File {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		writeOrDie(log, "ERR: Error %q", err.Error())
		panic(err)
	}

	return f
}

// writes a msg (fmt) with args or panics if that's not
// possible.
func writeOrDie(dst io.Writer, msg string, args... interface{}) {
	now := time.Now().Unix()
	_, err := io.WriteString(dst, fmt.Sprintf("%d\t", now) + fmt.Sprintf(msg, args...) + "\n")

	if err != nil {
		panic(err)
	}
}

type Stats struct {
	Success bool
	Size int64
	Speed float64
	Elapsed float64
	StatusCode int
	Now int64
	URL string
	Message string
}

// Create a HTTP Client using an H2Quic RoundTripper that determines
// the LocalAddr to listen on based on the iface name provided. 
// (it'll pick the first address of that interface). If the interface name
// provided is an empty string it'll listen on the zero address. 
func createHttpClient(ifaceName string, log io.Writer) (*http.Client, error) {
	if ifaceName == "" {
		hclient := &http.Client{
			Transport: &h2quic.RoundTripper{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		return hclient, nil
	}

	// Figure out the ip address of the specified interface
	// then pick the first one later on to use as LocalAddr
	iface, err := net.InterfaceByName(ifaceName)

	if err != nil {
		return nil, fmt.Errorf("ERR: Error using interface %q: %q", ifaceName, err.Error())
	}

	addrs, err := iface.Addrs()

	if err != nil {
		return nil, fmt.Errorf("ERR: Error using interface %q: %q", ifaceName, err.Error())
	}

	if len(addrs) < 1 {
		return nil, fmt.Errorf("ERR: Interface %q has no addresses?", ifaceName)
	}

	ipAddr := addrs[0].(*net.IPNet).IP
	udpAddr := &net.UDPAddr { IP: ipAddr }

	writeOrDie(log, "Using %q", udpAddr.String())

	hclient := &http.Client{
		Transport: &h2quic.RoundTripper{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			QuicConfig : &quic.Config { 
				ClientLocalAddr: udpAddr,
			},
		},
		Timeout : time.Second * 30, // maximum time a request may last is 30s
	}

	return hclient, nil
}

// Opens the logfile. If the path is an empty string
// it'll point to stdout. Panics if the logfile can not be opened.
func openLogfile(logfilePath string) *os.File {
	if logfilePath != "" {
		logfile := openAppendOrDie(logfilePath, nil)

		return logfile
	} else {
		return os.Stdout
	}
}

// Return the name of the next output file name
func getOFileName(nodeId string) string {
	return fmt.Sprintf("puic-poll-%d-%s.json", time.Now().UnixNano(), nodeId)
}

// Open the next output file in append,wronly,create,0600 mode.
func openNextOutputFile(odir, nodeId string) (*os.File, error) {
	fpath := filepath.Join(odir, getOFileName(nodeId))

	f, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return nil, err
	} else {
		return f, nil
	}
}

// Sleep for at least waitFrom but at most waitTo ms
func wait(waitFrom int, waitTo int) {
	wait := waitFrom + (rand.Int() % (waitTo - waitFrom))

	time.Sleep(time.Duration(wait) * time.Millisecond)
}

// load CA certs and set InsecureSkipVerify to false. 
func loadCerts(certs string, hclient *http.Client) error {
	rt := hclient.Transport.(*h2quic.RoundTripper)
	rt.TLSClientConfig.InsecureSkipVerify = false

	// Load CA certs
	caCert, err := ioutil.ReadFile(certs)

	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	rt.TLSClientConfig.RootCAs = caCertPool

	return nil
}

type monroeConfig struct {
	URLs string
	WaitFrom int
	WaitTo int
	Collect int
	IFaceName string
	Runs int
	NodeId string `json:nodeid`
}


func main() {
	// Default values
	urls := ""
	waitFrom := int(1000)
	waitTo := int(2000)
	collect := int(256)
	ifaceName := "op0"
	runs := int(4)

	var cfgFile = flag.String("config","/monroe/config","Path to config file!")
	var odirFlag = flag.String("odir","./tmp/","Path to output directory!")
	var certsFlag = flag.String("certs","/opt/monroe/rootCACert.pem","Path to certificates!")

	flag.Parse()

	odir := *odirFlag
	certs := *certsFlag

	file, err := os.OpenFile(*cfgFile, os.O_RDONLY, 0)

	if err != nil {
		panic(err.Error())
	}

	cfgData, err := ioutil.ReadAll(file)

	if err != nil {
		panic(err.Error())
	}

	monroeCfg := &monroeConfig{}

	err = json.Unmarshal(cfgData, &monroeCfg)

	if err != nil {
		panic(err.Error())
	}

	if monroeCfg.URLs != "" {
		urls = monroeCfg.URLs
	}

	if monroeCfg.WaitFrom != 0 {
		waitFrom = monroeCfg.WaitFrom
	}

	if monroeCfg.WaitTo != 0 {
		waitTo = monroeCfg.WaitTo
	}

	if monroeCfg.Collect != 0 {
		collect = monroeCfg.Collect
	}

	if monroeCfg.IFaceName != "" {
		ifaceName = monroeCfg.IFaceName
	}

	if monroeCfg.Runs != 0 {
		runs = monroeCfg.Runs
	}

	nodeId := monroeCfg.NodeId

	logfilePath := filepath.Join(odir, fmt.Sprintf("puic-poll-%d-%s.log", time.Now().Unix(), nodeId))

	
	// Logging setup
	collectedStats := make([]Stats, collect)
	j := 0

	logfile := openLogfile(logfilePath)
	defer logfile.Close()
	
	ofile, err := openNextOutputFile(odir, nodeId)

	if err != nil {
		writeOrDie(logfile, "ERR: Error opening output file: %q", err.Error())
		panic(err.Error())
	}

	hclient, err := createHttpClient(ifaceName, logfile)

	if err != nil {
		writeOrDie(logfile, "ERR: Error creating h2client: %q", err.Error())
		panic(err.Error())
	}

	if certs != "" {
		err := loadCerts(certs, hclient)

		if err != nil {
			writeOrDie(logfile, "ERR: Error loading certs: %q", err.Error())
			panic(err.Error())
		}
	}

	urlsToFetch := strings.Split(urls, ";")
	
	run := 0

	for {

		writeOrDie(logfile, "RUN [%d]", run)

		urlToFetch := urlsToFetch[rand.Int() % len(urlsToFetch)]

		size, speed, elapsed, statusCode, err := fetchOnce(urlToFetch, hclient, logfile)

		stats := Stats {
			Size : size,
			Speed : speed,
			Elapsed : elapsed,
			StatusCode : statusCode,
			Now : time.Now().UnixNano(),
			URL : urlToFetch,
		}

		if err != nil {
			stats.Success = false
			writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
			stats.Message = err.Error()
		} else {
			stats.Success = true
		}

		collectedStats[j] = stats

		statbytes, err := json.Marshal(stats)

		if err != nil {
			writeOrDie(logfile, fmt.Sprintf("ERR: Error: %q", err.Error()))
			panic(err.Error())
		}

		statstr := string(statbytes)

		_, err = io.WriteString(ofile, statstr+"\n")

		if err != nil {
			writeOrDie(logfile, "ERR: Error writing to output file: %q", err.Error())
		}

		io.WriteString(os.Stdout, statstr+"\n")

		j++

		if j >= collect {

			// Reset counter to zero, close the output file and open a new 
			// output file
			j = 0
			run++
			ofile.Close()

			if run >= runs {
				break
			}

			ofile, err = openNextOutputFile(odir, nodeId)

			if err != nil {
				writeOrDie(logfile, "ERR: Error opening output file: %q", err.Error())
				panic(err.Error())
			}
		}

		wait(waitFrom, waitTo)
	}
}

// Make a single GET request to url using hclient while logging to log
func fetchOnce(url string, hclient *http.Client, log io.Writer) (int64, float64, float64, int, error) {
	writeOrDie(log, "Start fetching %q", url)

	start := time.Now()

	rsp, err := hclient.Get(url)

	if err != nil {
		return 0, 0.0, 0.0, -1, err
	}

	writeOrDie(log, "Status code: %d", rsp.StatusCode)
	writeOrDie(log, "Content length: %d", rsp.ContentLength)

	n, err := io.Copy(&bytes.Buffer{}, rsp.Body)

	end := time.Now()
	elapsed := end.Sub(start).Seconds()

	speed := float64(n) / elapsed
	speed = speed / (1024.0 * 1024.0)

	writeOrDie(log, "Speed: %f MiB/s, elapsed: %f seconds", speed, elapsed)

	rsp.Body.Close()

	writeOrDie(log, "Done fetching %q", url)

	return n, speed, elapsed, rsp.StatusCode, err
}
