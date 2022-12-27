package main

import (
	"crypto/tls"
	"embed"
	_ "embed"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"something/currency"
	currencypb "something/pb"
	"strings"
	"time"
)

//go:embed playground/*
var swaggerui embed.FS

//go:embed server.crt
var tlsCert []byte

//go:embed server.key
var tlsKey []byte

func main() {
	go func() {
		<-time.After(10 * time.Second)
		//panic("crash")
	}()
	cert, _ := tls.X509KeyPair(tlsCert, tlsKey)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
	}
	grpcServer := grpc.NewServer()
	currencypb.RegisterCurrencyServer(grpcServer, &currency.GRPCServer{})
	wrappedServer := grpcweb.WrapServer(grpcServer,
		grpcweb.WithWebsockets(true),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)

	if l, err := net.Listen("tcp", ":8080"); err != nil {
		log.Fatal(err)
	} else {
		if wrappedServer != nil {
		}
		tlsl := tls.NewListener(l, tlsConfig)
		m := cmux.New(tlsl)
		//m := cmux.New(l)
		httpL := m.Match(cmux.HTTP1())
		grpcL := m.Match(cmux.HTTP2())
		swaggerServer := http.FileServer(http.FS(swaggerui))
		handler := func(resp http.ResponseWriter, req *http.Request) {
			execFile, _ := os.Executable()
			resp.Header().Set("Cache-Control", "private, max-age=3600")
			resp.Header().Set("Micro-Service", execFile)
			if strings.HasPrefix(req.RequestURI, "/playground") {
				swaggerServer.ServeHTTP(resp, req)
				return
			}
			if req.ProtoMajor == 2 || wrappedServer.IsGrpcWebRequest(req) || req.Method == "OPTIONS" {
				wrappedServer.ServeHTTP(resp, req)
			}
		}
		httpS := &http.Server{
			Handler: http.HandlerFunc(handler),
		}
		//grpcServer.Serve(l)
		go grpcServer.Serve(grpcL)
		go httpS.Serve(httpL)
		if err := m.Serve(); err != nil {
			log.Fatal(err)
		}

	}
	/*mux := runtime.NewServeMux()

	err := currencypb.RegisterCurrencyHandlerFromEndpoint(context.Background(), mux, "localhost:8080", nil)
	//	[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Handler: mux,
	}



	err = server.Serve(l)

	if err != nil {
		log.Fatal(l)
	}
	*/
	//if err := http.ListenAndServe("8080", mux); err != nil {
	//	panic(err)
	//}
	//s := grpc.NewServer()
	//srv := &currency.GRPCServer{}
	//currencypb.RegisterCurrencyServer(s, srv)

}
