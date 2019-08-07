package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	HEAD string = `<title>Simple Server</title>
				<meta name="viewport" content="width=device-width, initial-scale=1">
				<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
				<style>

					ol {
						list-style-type: none;
					}
				</style>`
	FOLDER_HTML string = `<i class="material-icons">folder</i>`

	FILE_HTML string = `<i class="material-icons">file_download</i>`
)

type Server struct {
	config     config
	router     *mux.Router
	httpServer *http.Server
}

func (server *Server) Init(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &server.config)
	if err != nil {
		return err
	}

	server.router = mux.NewRouter()
	server.router.PathPrefix("/").HandlerFunc(server.handleRequest)

	server.httpServer = new(http.Server)
	server.httpServer.Addr = server.config.Host + ":" + server.config.Port
	server.httpServer.WriteTimeout = time.Second * 10
	server.httpServer.ReadTimeout = time.Second * 10
	server.httpServer.IdleTimeout = time.Second * 10
	server.httpServer.Handler = server.router

	return nil
}

func (server *Server) Run() {
	go func() {
		if err := server.httpServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("Listen and server:", server.httpServer.Addr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	server.httpServer.Shutdown(context.Background())

	log.Println("Shut down")
}

func (server *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Receive request from:", r.RemoteAddr, "Url:", r.URL.String())
	newQueryURL, err := url.PathUnescape(r.URL.String())
	fullPath := strings.TrimRight(server.config.Root, "/") + "/" + strings.TrimLeft(newQueryURL, "/")

	if err != nil {
		log.Println(err)
		w.Write([]byte(""))
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		log.Println(err)
		w.Write([]byte(""))
		return
	}

	if info.IsDir() {
		server.handleDirRequest(w, r)
	} else {
		server.handleFileRequest(w, r)
	}
}

func (server *Server) handleFileRequest(w http.ResponseWriter, r *http.Request) {
	newURL, _ := url.PathUnescape(r.URL.String())
	fullPath := strings.TrimRight(server.config.Root, "/") + "/" + strings.TrimLeft(newURL, "/")

	file, err := os.Open(fullPath)
	if err != nil {
		log.Println(err)
		w.Write([]byte(""))
		return
	}

	attachmentType := `"attachment; filename="%s"`
	attachmentType = fmt.Sprintf(attachmentType, path.Base(file.Name()))

	fileStat, _ := file.Stat()
	fileSize := strconv.FormatInt(fileStat.Size(), 10)

	data := make([]byte, 250)
	file.Read(data)
	contentType := http.DetectContentType(data)

	w.Header().Set("Content-Disposition", attachmentType)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fileSize)

	file.Seek(0, 0)
	io.Copy(w, file)
}

func (server *Server) handleDirRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	result := `<!DOCTYPE html><html><head>%s</head><body>%s</body></html>`
	body := ""

	fullPath := strings.TrimRight(server.config.Root, "/") + "/" + strings.TrimLeft(r.URL.String(), "/")
	dirs, files, err := server.getDirsAndFilesList(fullPath)
	if err != nil {
		log.Println(err)
		result = fmt.Sprintf(result, HEAD, body)
		w.Write([]byte(result))
		return
	}

	table := `<table>`

	for _, dir := range dirs {

		item := `<tr><td>%s</td><td>%s</td></tr>`

		newDirname := url.PathEscape(dir.Name())

		actualLink := r.URL.String() + "/" + newDirname
		actualLink = strings.TrimLeft(actualLink, "/")
		link := `<a href="/%s">%s</a>`
		link = fmt.Sprintf(link, actualLink, dir.Name())
		item = fmt.Sprintf(item, FOLDER_HTML, link)
		table += item
	}

	for _, file := range files {
		item := `<tr><td>%s</td><td>%s</td></tr>`
		newFilename := url.PathEscape(file.Name())

		actualLink := r.URL.String() + "/" + newFilename
		actualLink = strings.TrimLeft(actualLink, "/")

		link := `<a href="/%s">%s</a>`
		link = fmt.Sprintf(link, actualLink, file.Name())

		item = fmt.Sprintf(item, FILE_HTML, link)
		table += item
	}
	table += `</table>`

	body = table

	result = fmt.Sprintf(result, HEAD, body)
	w.Write([]byte(result))
}

func (server *Server) getDirsAndFilesList(fullPath string) (dirs []os.FileInfo, files []os.FileInfo, err error) {
	infos, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return dirs, files, err
	}

	for _, info := range infos {
		if len(info.Name()) > 0 && info.Name() != "/" && info.Name()[0] != '.' {
			if info.IsDir() {
				dirs = append(dirs, info)
			} else {
				files = append(files, info)
			}
		}
	}

	return dirs, files, nil
}
