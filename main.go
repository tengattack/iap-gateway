package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"./common"
	yaml "gopkg.in/yaml.v2"

	"github.com/nytimes/gziphandler"
	"golang.org/x/net/proxy"
)

const (
	AppStoreURL        = "https://buy.itunes.apple.com/verifyReceipt"
	AppStoreSandboxURL = "https://sandbox.itunes.apple.com/verifyReceipt"
)

// Config is the main config
type Config struct {
	Core struct {
		Port       int  `yaml:"port"`
		Production bool `yaml:"production"`
	} `yaml:"core"`
	Proxies []string `yaml:"proxies"`
}

var (
	// Version of iap-gateway
	Version = "dev"
)
var conf Config
var trs []*http.Transport

func initTransports(proxies []string) {
	trs = make([]*http.Transport, len(proxies))

	for i, addr := range proxies {
		if addr == "" {
			// direct
			trs[i] = nil
			continue
		}
		proxyURL, _ := url.Parse(addr)
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host,
			nil,
			&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
			},
		)
		if err != nil {
			panic(err)
		}
		tr := &http.Transport{
			Proxy:               nil,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		// direct is the first one
		trs[i] = tr
	}
}

func verify(data []byte) (*common.MyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan *common.MyResponse, len(trs))
	errs := make(chan error, len(trs))

	for i, tr := range trs {
		go func(i int, tr *http.Transport) {
			var url string
			if conf.Core.Production {
				url = AppStoreURL
			} else {
				url = AppStoreSandboxURL
			}
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			if err != nil {
				errs <- err
				return
			}
			req = req.WithContext(ctx)

			t0 := time.Now().UnixNano()
			var client *http.Client
			if tr != nil {
				client = &http.Client{
					Transport: tr,
					Timeout:   5 * time.Second,
				}
			} else {
				client = &http.Client{
					Timeout: 5 * time.Second,
				}
			}
			resp, err := common.DoHTTPRequest(req, true, client)
			if err != nil {
				errs <- err
				return
			}
			t1 := time.Now().UnixNano()
			done <- resp
			if tr == nil {
				fmt.Printf("direct%d time cost: %fs\n", i, float64(t1-t0)/float64(time.Second))
			} else {
				fmt.Printf("proxy%d time cost: %fs\n", i, float64(t1-t0)/float64(time.Second))
			}
		}(i, tr)
	}

	select {
	case resp := <-done:
		return resp, nil
	case <-ctx.Done():
		select {
		case err := <-errs:
			return nil, err
		default:
			return nil, errors.New("Timeout")
		}
	}
}

func verifyReceipt(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()

	resp, err := verify(body)
	if err != nil {
		fmt.Printf("verify error: %v\n", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	header := w.Header()
	for key, values := range resp.Header {
		if strings.ToLower(key) == "content-length" {
			continue
		}
		for _, value := range values {
			header.Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	_, err = w.Write(resp.Body)
	if err != nil {
		fmt.Printf("write error: %v\n", err)
	}
}

func main() {
	flag := flag.NewFlagSet(os.Args[0]+" "+Version, flag.ExitOnError)
	configFilePath := flag.String("config", "", "config file path")
	showHelp := flag.Bool("help", false, "show help message")
	flag.Parse(os.Args[1:])

	if *showHelp {
		flag.Usage()
		return
	}

	configFile, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(configFile, &conf)
	if err != nil {
		panic(err)
	}

	if len(conf.Proxies) <= 0 {
		log.Fatal("no proxies found")
	}

	initTransports(conf.Proxies)

	http.Handle("/verifyReceipt", gziphandler.GzipHandler(http.HandlerFunc(verifyReceipt)))
	http.ListenAndServe(":"+strconv.Itoa(conf.Core.Port), nil)
}
