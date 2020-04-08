package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ZloyDyadka/go-httptcpinfo"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fd, ok := tcpinfo.ExtractFDFromCtx(r.Context())
		if !ok {
			log.Println("fd is not found in context")
			return
		}

		info, err := tcpinfo.GetTCPInfoByFD(fd)
		if err != nil {
			log.Println("can't get tcpinfo by fd:", err)
			return
		}

		if err := json.NewEncoder(w).Encode(info); err != nil {
			log.Println("can't encode tcpinfo to JSON:", err)
			return
		}
	})

	httpServer := http.Server{
		Addr:        ":8080",
		Handler:     mux,
		ConnContext: tcpinfo.HTTPConnFDMiddleware,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		return errors.Wrap(err, "listen and serve HTTP")
	}

	return nil
}
