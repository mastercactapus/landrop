package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"flag"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/mastercactapus/landrop/app"
	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

func main() {
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", ":8000", "Listen address.")
	name := flag.String("svc", "landrop.local", "Service name.")
	dir := flag.String("dir", ".", "Directory to serve.")
	uploads := flag.Bool("w", false, "Allow drag-and-drop uploads.")
	scan := flag.Bool("s", false, "Scan file contents for description (requires `file` tool).")
	https := flag.Bool("https", false, "Serve over HTTPS (if -cert and -key are omitted, will generate self-signed cert).")
	cert := flag.String("cert", "", "Path to TLS certificate.")
	key := flag.String("key", "", "Path to TLS key.")
	flag.Parse()

	mdnsAddr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	if err != nil {
		log.Fatalln("resolve mdns:", err)
	}

	l, err := net.ListenUDP("udp4", mdnsAddr)
	if err != nil {
		log.Fatalln("listen mdns:", err)
	}
	defer l.Close()

	_, err = mdns.Server(ipv4.NewPacketConn(l), &mdns.Config{
		LocalNames: []string{*name},
	})
	if err != nil {
		log.Fatalln("mdns server:", err)
	}

	hl, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("http listen:", err)
	}
	defer hl.Close()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalln("abs dir:", err)
	}

	if *uploads {
		log.Println("Uploads enabled.")
	}

	srv := &http.Server{
		Handler: app.NewServer(app.Config{
			Name:     *name,
			Dir:      absDir,
			Writable: *uploads,
			Scan:     *scan,
		}),
	}
	if *https && *cert == "" && *key == "" {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		l := lru.New(100)
		var mx sync.Mutex
		srv.TLSConfig = &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				c, ok := l.Get(hello.ServerName)
				if ok {
					return c.(*tls.Certificate), nil
				}
				mx.Lock()
				defer mx.Unlock()

				// Check again in case another goroutine set it.
				c, ok = l.Get(hello.ServerName)
				if ok {
					return c.(*tls.Certificate), nil
				}

				now := time.Now()
				template := &x509.Certificate{
					SerialNumber: big.NewInt(now.Unix()),
					Subject: pkix.Name{
						CommonName:         hello.ServerName,
						OrganizationalUnit: []string{"landrop"},
					},
					NotBefore: now,
					NotAfter:  now.AddDate(0, 0, 1), // Valid for one day

					BasicConstraintsValid: true,
					IsCA:                  true,
					ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
					KeyUsage: x509.KeyUsageKeyEncipherment |
						x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
				}
				cert, err := x509.CreateCertificate(rand.Reader, template, template,
					priv.Public(), priv)
				if err != nil {
					return nil, err
				}

				var outCert tls.Certificate
				outCert.Certificate = append(outCert.Certificate, cert)
				outCert.PrivateKey = priv
				l.Add(hello.ServerName, &outCert)
				return &outCert, nil
			},
		}
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(*name, strconv.Itoa(hl.Addr().(*net.TCPAddr).Port)),
	}

	if *https {
		u.Scheme = "https"
	}

	nameURL := u.String()
	u.Host = hl.Addr().String()
	addrURL := u.String()

	log.Printf("Serving %s on %s (%s)", absDir, nameURL, addrURL)
	if *https {
		err = srv.ServeTLS(hl, *cert, *key)
	} else {
		err = srv.Serve(hl)
	}
	if err != nil {
		log.Fatal(err)
	}
}
