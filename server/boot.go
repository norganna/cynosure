package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/gateway"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
	"github.com/norganna/cynosure/proto/cynosure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Run starts up the server.
func Run(config *common.Config, args []string) {
	ctx := context.Background()
	log := config.Log()

	// #### Add brokers ####

	for identity, b := range config.Brokers {
		err := deps.NewInstance(identity, "", b.Kind, b.Config.Default)
		if err != nil {
			log.Fatalf("Failed to initialise broker identity %s: %s", identity, err.Error())
		}

		if b.Config.Namespaced != nil {
			for ns, config := range b.Config.Namespaced {
				err = deps.NewInstance(identity, ns, b.Kind, config)
				if err != nil {
					log.Fatalf("Failed to initialise broker identity %s in %s namespace: %s", identity, ns, err.Error())
				}
			}
		}
	}

	// #### Setup TLS ####

	sHost, sPort, err := net.SplitHostPort(config.Server)
	if err != nil {
		log.Fatal("Failed to parse server address: ", err)
	}

	gPort, err := strconv.ParseInt(sPort, 10, 32)
	if err != nil {
		log.Fatal("Failed to parse server port: ", err)
	}
	gPort++

	serverCrt, err := config.ServerCert(365)
	if err != nil {
		log.Fatal("Failed to get server certificate: ", err)
		return
	}

	clientCrt, err := config.ClientCert(365)
	if err != nil {
		log.Fatal("Failed to get client certificate: ", err)
		return
	}

	pool, err := config.CertPool()
	if err != nil {
		log.Fatal("Failed to obtain a certificate pool: ", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverCrt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}
	serverCreds := credentials.NewTLS(tlsConfig)

	clientCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{*clientCrt},
		RootCAs:      pool,
		ServerName:   "api.cynosure",
	})

	// #### gRPC SERVER ####

	rpcListen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", sHost, gPort))
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}

	rpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(grpc_validator.UnaryServerInterceptor()),
		grpc.StreamInterceptor(grpc_validator.StreamServerInterceptor()),
	)
	cynosure.RegisterAPIServer(rpcServer, newHandler(config))

	// Serve gRPC Server
	log.Info("Serving gRPC on ", rpcListen.Addr())
	go func() {
		log.Fatal(rpcServer.Serve(rpcListen))
	}()

	// #### WEB SERVER ####

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(w, strings.NewReader(cynosure.Swagger))
	})

	jsonPB := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonPB),
		runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)

	if sHost == "" {
		sHost = "localhost"
	}
	dialAddr := fmt.Sprintf("passthrough:///%s:%d", sHost, gPort)

	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal("Failed to dial server: ", err)
	}

	err = cynosure.RegisterAPIHandler(
		ctx, gwMux, conn,
	)
	if err != nil {
		log.Fatal("Failed to register API handler: ", err)
		return
	}
	mux.Handle("/", gwMux)

	webListen, err := net.Listen("tcp", config.Server)
	if err != nil {
		log.Fatalln("Failed to listen: ", err)
	}

	webServer := &http.Server{
		Addr:      config.Server,
		Handler:   grpcHandler(rpcServer, mux),
		TLSConfig: tlsConfig,
	}

	log.Info("Listening on ", webListen.Addr())
	if err = webServer.Serve(tls.NewListener(webListen, webServer.TLSConfig)); err != nil {
		log.Fatal("Stopped: ", err)
	}
}

func grpcHandler(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && r.Method == "PRI" && r.RequestURI == "*" {
			fmt.Println("Serving GRPC")
			grpcServer.ServeHTTP(w, r)
		} else {
			fmt.Println("Serving other", r.RequestURI, r.Host, r.Method, r.Trailer, r.TransferEncoding, r.Proto)
			otherHandler.ServeHTTP(w, r)
		}
	})
}
