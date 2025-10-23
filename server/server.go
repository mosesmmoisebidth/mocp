package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/claudiodangelis/qrcp/qr"

	"github.com/claudiodangelis/qrcp/body"
	"github.com/claudiodangelis/qrcp/config"
	"github.com/claudiodangelis/qrcp/pages"
	"github.com/claudiodangelis/qrcp/style"
	"github.com/claudiodangelis/qrcp/util"
	"gopkg.in/cheggaaa/pb.v1"
	"sync/atomic"
	"time"
)

// Server is the server
type Server struct {
	BaseURL string
	// SendURL is the URL used to send the file
	SendURL string
	// ReceiveURL is the URL used to Receive the file
	ReceiveURL  string
	instance    *http.Server
	body        body.Body
	outputDir   string
	stopChannel chan bool
	// expectParallelRequests is set to true when qrcp sends files, in order
	// to support downloading of parallel chunks
	expectParallelRequests bool
}

// ReceiveTo sets the output directory
func (s *Server) ReceiveTo(dir string) error {
	output, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	// Check if the output dir exists
	fileinfo, err := os.Stat(output)
	if err != nil {
		return err
	}
	if !fileinfo.IsDir() {
		return fmt.Errorf("%s is not a valid directory", output)
	}
	s.outputDir = output
	return nil
}

// Send adds a handler for sending the file
func (s *Server) Send(p body.Body) {
	s.body = p
	s.expectParallelRequests = true
}

// DisplayQR creates a handler for serving the QR code in the browser
func (s *Server) DisplayQR(url string) {
	const PATH = "/qr"
	qrImg := qr.RenderImage(url)
	http.HandleFunc(PATH, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		if err := jpeg.Encode(w, qrImg, nil); err != nil {
			panic(err)
		}
	})
	openBrowser(s.BaseURL + PATH)
}

// Wait for transfer to be completed, it waits forever if kept awlive
func (s Server) Wait() error {
	<-s.stopChannel
	if err := s.instance.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}
	if s.body.DeleteAfterTransfer {
		if err := s.body.Delete(); err != nil {
			panic(err)
		}
	}
	return nil
}

// Shutdown the server
func (s Server) Shutdown() {
	s.stopChannel <- true
}

// New instance of the server
func New(cfg *config.Config) (*Server, error) {
	// Show the MOCP logo
	fmt.Println(style.Logo)

	app := &Server{}
	// Get the address of the configured interface to bind the server to.
	// If `bind` configuration parameter has been configured, it takes precedence
	bind, err := util.GetInterfaceAddress(cfg.Interface)
	if err != nil {
		return &Server{}, err
	}
	if cfg.Bind != "" {
		bind = cfg.Bind
	}
	// Create a listener. If `port: 0`, a random one is chosen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bind, cfg.Port))
	if err != nil {
		return nil, err
	}
	// Set the value of computed port
	port := listener.Addr().(*net.TCPAddr).Port
	// Set the host
	host := fmt.Sprintf("%s:%d", bind, port)
	// Get a random path to use
	path := cfg.Path
	if path == "" {
		path = util.GetRandomURLPath()
	}
	// Set the hostname
	hostname := fmt.Sprintf("%s:%d", bind, port)
	// Use external IP when using `interface: any`, unless a FQDN is set
	if bind == "0.0.0.0" && cfg.FQDN == "" {
		fmt.Println("Retrieving the external IP...")
		extIP, err := util.GetExternalIP()
		if err != nil {
			panic(err)
		}
		extIPString := extIP.String()
		fmtstring := "%s:%d"
		if strings.Count(extIPString, ":") >= 2 {
			// IPv6 address, wrap it in [] to add a port
			fmtstring = "[%s]:%d"
		}
		hostname = fmt.Sprintf(fmtstring, extIPString, port)
	}
	// Use a fully-qualified domain name if set
	if cfg.FQDN != "" {
		hostname = fmt.Sprintf("%s:%d", cfg.FQDN, port)
	}
	// Set URLs
	protocol := "http"
	if cfg.Secure {
		protocol = "https"
	}
	app.BaseURL = fmt.Sprintf("%s://%s", protocol, hostname)
	app.SendURL = fmt.Sprintf("%s/send/%s",
		app.BaseURL, path)
	app.ReceiveURL = fmt.Sprintf("%s/receive/%s",
		app.BaseURL, path)
	// Create a server
	httpserver := &http.Server{
		Addr: host,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	// Create channel to send message to stop server
	app.stopChannel = make(chan bool)
	// Create cookie used to verify request is coming from first client to connect
	cookie := http.Cookie{Name: "qrcp", Value: ""}
	// Gracefully shutdown when an OS signal is received or when "q" is pressed
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	start := time.Now()
	go func() {
		<-sig
		app.stopChannel <- true
	}()
	// The handler adds and removes from the sync.WaitGroup
	// When the group is zero all requests are completed
	// and the server is shutdown
	var waitgroup sync.WaitGroup
	waitgroup.Add(1)
	var initCookie sync.Once
	// Create handlers
	// Send handler (sends file to caller)
	http.HandleFunc("/send/"+path, func(w http.ResponseWriter, r *http.Request) {
		if !cfg.KeepAlive && strings.HasPrefix(r.Header.Get("User-Agent"), "Mozilla") {
			if cookie.Value == "" {
				initCookie.Do(func() {
					value, err := util.GetSessionID()
					if err != nil {
						log.Println("Unable to generate session ID", err)
						app.stopChannel <- true
						return
					}
					cookie.Value = value
					http.SetCookie(w, &cookie)
				})
			} else {
				// Check for the expected cookie and value
				// If it is missing or doesn't match
				// return a 400 status
				rcookie, err := r.Cookie(cookie.Name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if rcookie.Value != cookie.Value {
					http.Error(w, "mismatching cookie", http.StatusBadRequest)
					return
				}
				// If the cookie exits and matches
				// this is an aadditional request.
				// Increment the waitgroup
				waitgroup.Add(1)
			}
			// Remove connection from the waitgroup when done
			defer waitgroup.Done()
		}
		w.Header().Set("Content-Disposition", "attachment; filename=\""+
			app.body.Filename+
			"\"; filename*=UTF-8''"+
			url.QueryEscape(app.body.Filename))

		// Stream the file manually so we can show send-side progress
		fi, err := os.Stat(app.body.Path)
		if err != nil {
			http.Error(w, "Unable to stat file", http.StatusInternalServerError)
			log.Printf("Unable to stat file: %v", err)
			return
		}
		total := fi.Size()
		w.Header().Set("Content-Length", strconv.FormatInt(total, 10))
		// Content type fallback
		w.Header().Set("Content-Type", "application/octet-stream")

		f, err := os.Open(app.body.Path)
		if err != nil {
			http.Error(w, "Unable to open file", http.StatusInternalServerError)
			log.Printf("Unable to open file: %v", err)
			return
		}
		defer f.Close()

		// Start progress tracking
		progressBar := pb.New64(total)
		progressBar.ShowCounters = false
		progressBar.Prefix(app.body.Filename)
		// don't call Start() to avoid automatic terminal output; internal
		// counters are still updated via progressBar.Add.

		var bytesSent int64
		done := make(chan struct{})
		go func() {
			// use short filename as prefix for the progress bar
			prefix := filepath.Base(app.body.Filename)
			start := time.Now()
			ticker := time.NewTicker(300 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					cur := int(atomic.LoadInt64(&bytesSent))
					elapsed := time.Since(start)
					// rate in bytes/sec
					var rate float64
					if elapsed > 0 {
						rate = float64(cur) / elapsed.Seconds()
					}
					bar := style.AnimatedProgressBarWithStats(int64(cur), total, prefix, rate, elapsed)
					plain := style.AnimatedPlainProgressBarWithStats(int64(cur), total, prefix, rate, elapsed)
					func() {
						defer func() {
							if rec := recover(); rec != nil {
								log.Printf("Send progress: %s", bar)
								log.Printf("Send progress (plain): %s", plain)
							}
						}()
						// update reserved progress line under QR
						qr.UpdateProgressLine(bar)
						// also log progress so it's visible in environments that don't support ANSI cursor updates
						log.Printf("Send progress: %s", bar)
						// also log a plain version
						log.Printf("Send progress (plain): %s", plain)
					}()
				case <-done:
					return
				}
			}
		}()

		buf := make([]byte, 32*1024)
		// last immediate-log time to avoid spamming logs on every chunk
		lastLog := time.Now().Add(-time.Second)
		for {
			n, rerr := f.Read(buf)
			if n > 0 {
				wn, werr := w.Write(buf[:n])
				if werr != nil {
					log.Printf("Error writing to client: %v", werr)
					break
				}
				if wn > 0 {
					atomic.AddInt64(&bytesSent, int64(wn))
					progressBar.Add(wn)
					// throttle immediate plain logging to ~500ms
					if time.Since(lastLog) > 500*time.Millisecond {
						cur := int(atomic.LoadInt64(&bytesSent))
						elapsed := time.Since(start)
						var rate float64
						if elapsed > 0 {
							rate = float64(cur) / elapsed.Seconds()
						}
						colored := style.AnimatedProgressBarWithStats(int64(cur), total, filepath.Base(app.body.Filename), rate, elapsed)
						log.Printf("Send progress (immediate): %s", colored)
						lastLog = time.Now()
					}
				}
				if fl, ok := w.(http.Flusher); ok {
					fl.Flush()
				}
			}
			if rerr != nil {
				if rerr == io.EOF {
					break
				}
				log.Printf("Error reading file: %v", rerr)
				break
			}
		}

		// Do not call progressBar.FinishPrint — it prints the pb library's
		// own progress UI (the '=' bar). We use our custom progress output.
		close(done)
	})
	// Upload handler (serves the upload page)
	http.HandleFunc("/receive/"+path, func(w http.ResponseWriter, r *http.Request) {
		htmlVariables := struct {
			Route string
			File  string
		}{}
		htmlVariables.Route = "/receive/" + path
		switch r.Method {
		case "POST":
			filenames := util.ReadFilenames(app.outputDir)
			reader, err := r.MultipartReader()
			if err != nil {
				fmt.Fprintf(w, "Upload error: %v\n", err)
				log.Printf("Upload error: %v\n", err)
				app.stopChannel <- true
				return
			}
			// Log that we received a POST upload request and are ready to receive parts
			log.Printf("Upload request received — waiting to receive file parts from client %s", r.RemoteAddr)
			transferredFiles := []string{}
			progressBar := pb.New64(r.ContentLength)
			progressBar.ShowCounters = false

			// Non-invasive progress renderer: track bytes transferred using an atomic counter
			var bytesTransferred int64 = 0
			var progressPrefix string
			doneRendering := make(chan struct{})
			// last immediate-log time to avoid spamming logs on every chunk
			lastLog := time.Now().Add(-time.Second)
			start := time.Now()
			go func() {
				ticker := time.NewTicker(300 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						// Build a single-line progress string and attempt an in-place update
						current := int(atomic.LoadInt64(&bytesTransferred))
						total := int(r.ContentLength)
						elapsed := time.Since(start)
						var rate float64
						if elapsed > 0 {
							rate = float64(current) / elapsed.Seconds()
						}
						bar := style.AnimatedProgressBarWithStats(int64(current), int64(total), progressPrefix, rate, elapsed)
						// Update the reserved single progress line below the QR; also log as fallback
						func() {
							defer func() {
								if rec := recover(); rec != nil {
									// fallback to logging when ANSI escapes are not supported
									log.Printf("Progress: %s", bar)
								}
							}()
							qr.UpdateProgressLine(bar)
							// Also log so progress is visible even if cursor updates fail
							log.Printf("Progress: %s", bar)
							// Log a plain(no-ANSI) version for terminals that don't render colors
							plain := style.AnimatedPlainProgressBarWithStats(int64(current), int64(total), progressPrefix, rate, elapsed)
							log.Printf("Progress (plain): %s", plain)
						}()
					case <-doneRendering:
						return
					}
				}
			}()
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				// iIf part.FileName() is empty, skip this iteration.
				if part.FileName() == "" {
					continue
				}
				// Prepare the destination
				fileName := getFileName(filepath.Base(part.FileName()), filenames)
				out, err := os.Create(filepath.Join(app.outputDir, fileName))
				if err != nil {
					// Output to server
					fmt.Fprintf(w, "Unable to create the file for writing: %s\n", err)
					// Output to console
					log.Printf("Unable to create the file for writing: %s\n", err)
					// Send signal to server to shutdown
					app.stopChannel <- true
					return
				}
				defer out.Close()
				// Add name of new file
				filenames = append(filenames, fileName)
				// set prefix for progress rendering
				progressPrefix = fileName
				// Write the content from POSTed file to the out
				// Use log.Printf so it appears on stderr and is more reliably visible
				log.Printf("Transferring file: %s", out.Name())
				progressBar.Prefix(out.Name())
				// Do not call Start() to avoid automatic terminal rendering by the pb library.
				buf := make([]byte, 1024)
				for {
					// Read a chunk
					n, err := part.Read(buf)
					if err != nil && err != io.EOF {
						// Output to server
						fmt.Fprintf(w, "Unable to write file to disk: %v", err)
						// Output to console
						fmt.Printf("Unable to write file to disk: %v", err)
						// Send signal to server to shutdown
						app.stopChannel <- true
						return
					}
					if n == 0 {
						break
					}
					// Write a chunk
					if _, err := out.Write(buf[:n]); err != nil {
						// Output to server
						fmt.Fprintf(w, "Unable to write file to disk: %v", err)
						// Output to console
						log.Printf("Unable to write file to disk: %v", err)
						// Send signal to server to shutdown
						app.stopChannel <- true
						return
					}
					// Update progress counters
					progressBar.Add(n)
					atomic.AddInt64(&bytesTransferred, int64(n))
					// immediate plain logging (throttled)
					// Use a per-upload lastLog variable stored in closure
					if time.Since(lastLog) > 500*time.Millisecond {
						cur := int(atomic.LoadInt64(&bytesTransferred))
						elapsed := time.Since(start)
						var rate float64
						if elapsed > 0 {
							rate = float64(cur) / elapsed.Seconds()
						}
						colored := style.AnimatedProgressBarWithStats(int64(cur), int64(r.ContentLength), progressPrefix, rate, elapsed)
						log.Printf("Progress (immediate): %s", colored)
						lastLog = time.Now()
					}
				}
			}
			// Do not call progressBar.FinishPrint for the same reason as above.
			// Stop the QR+progress renderer
			close(doneRendering)
			// Set the value of the variable to the actually transferred files
			htmlVariables.File = strings.Join(transferredFiles, ", ")
			serveTemplate("done", pages.Done, w, htmlVariables)
			if !cfg.KeepAlive {
				app.stopChannel <- true
			}
		case "GET":
			serveTemplate("upload", pages.Upload, w, htmlVariables)
		}
	})
	// Wait for all wg to be done, then send shutdown signal
	go func() {
		waitgroup.Wait()
		if cfg.KeepAlive || !app.expectParallelRequests {
			return
		}
		app.stopChannel <- true
	}()

	go func() {
		netListener := tcpKeepAliveListener{listener.(*net.TCPListener)}
		if cfg.Secure {
			if err := httpserver.ServeTLS(netListener, cfg.TlsCert, cfg.TlsKey); err != http.ErrServerClosed {
				log.Fatalln("error starting the server:", err)
			}
		} else {
			if err := httpserver.Serve(netListener); err != http.ErrServerClosed {
				log.Fatalln("error starting the server", err)
			}
		}
	}()

	app.instance = httpserver
	return app, nil
}

// openBrowser navigates to a url using the default system browser
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("failed to open browser on platform: %s", runtime.GOOS)
	}
	if err != nil {
		log.Fatal(err)
	}
}
