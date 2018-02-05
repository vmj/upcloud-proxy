/*
 * upcloud-proxy
 * Copyright (C) 2018  <mikko@varri.fi>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"flag"
	"net/http"
	"log"
	"github.com/elazarl/goproxy"
	"strings"
	"os"
	"io/ioutil"
	"bytes"
)

func main() {
	username := flag.String("username", "", "UpCloud API username")
	password := flag.String("password", "", "UpCloud API password")
	addr 	 := flag.String("addr", ":8080", "Address to listen to")
	verbose  := flag.Bool("verbose", false, "Be verbose")

	flag.Parse()

	// if username and/or password weren't given on command line, try to get them from environment
	getFromEnvironment("username", *username, "UPCLOUD_API_USERNAME")
	getFromEnvironment("password", *password, "UPCLOUD_API_PASSWORD")

	// if username and/or password starts with @ character, try to read from from those files
	getFromFile("username", *username)
	getFromFile("password", *password)

	if *username == "" || *password == "" {
		log.Fatalln("both username and password must be given and non-empty")
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	// redirect all non-proxy requests to UpCloud API
	proxy.NonproxyHandler = http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = "api.upcloud.com"
		proxy.ServeHTTP(w, req)
	})

	// Fix all passing UpCloud API calls
	proxy.OnRequest().DoFunc(func (req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

		// UpCloud API does not like the port that some clients put in the Host header (api.upcloud.com:443)
		req.Host = "api.upcloud.com"

		// Add Authorization header unless it already is there
		if req.Header.Get("Authorization") == "" {
			req.SetBasicAuth(*username, *password)
		}

		// Force Accept header to application/json
		if !strings.HasPrefix(req.Header.Get("Accept"), "application/json") {
			req.Header.Set("Accept", "application/json; charset=UTF-8")
		}

		// Force Content-Type header to application/json if there seems to be a payload
		cl := req.Header.Get("Content-Length")
		if cl != "" && cl != "0" {
			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
				req.Header.Set("Content-Type", "application/json")
			}
		}

		// Add upcloud-proxy to the User-Agent string
		ua := req.Header.Get("User-Agent")
		if ua == "" {
			req.Header.Set("User-Agent", "upcloud-proxy/0.1")
		} else {
			req.Header.Set("User-Agent", "upcloud-proxy/0.1 " + ua)
		}

		return req, nil
	})

	log.Printf("Starting upcloud-proxy for username '%s'", *username)
	log.Fatal(http.ListenAndServe(*addr, proxy))
}

// When running in a container, it is easiest to pass the required parameters via environment values
func getFromEnvironment(variable, value, environmentVariable string) {
	if value == "" {
		flag.Set(variable, os.Getenv(environmentVariable))
	}
}

// When running on Kubernetes, it is customary to pass the credentials via secret mounts
func getFromFile(variable, value string) {
	if strings.HasPrefix(value, "@") {
		file := value[1:]
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("unable to read %s from %s: %s", variable, file, err.Error())
		}
		flag.Set(variable, bytes.NewBuffer(data).String())
	}
}
