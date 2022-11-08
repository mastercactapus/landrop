package main

import (
	_ "embed"
	"flag"
	"log"
	"net"
	"net/http"
	"path/filepath"

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
	log.Printf("Serving %s on http://%s:%d (http://%s)", absDir, *name, hl.Addr().(*net.TCPAddr).Port, hl.Addr())

	err = http.Serve(hl, app.NewServer(app.Config{
		Name:     *name,
		Dir:      absDir,
		Writable: *uploads,
		Scan:     *scan,
	}))
	if err != nil {
		log.Fatal(err)
	}
}
