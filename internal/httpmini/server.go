package httpmini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	Addr      string
	PublicDir string
}

func (s *Server) ListenAndServe(handle func(conn net.Conn)) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		handle(conn)
	}
}

func BuildResponse(statusLine, contentType string, body []byte) []byte {
	headers := fmt.Sprintf(
		"%s\r\nDate: %s\r\nContent-Length: %d\r\nContent-Type: %s\r\nConnection: close\r\n\r\n",
		statusLine,
		time.Now().UTC().Format(time.RFC1123),
		len(body),
		contentType,
	)
	var buf bytes.Buffer
	buf.WriteString(headers)
	buf.Write(body)
	return buf.Bytes()
}

func (s *Server) HandleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimRight(line, "\r\n")
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		conn.Write(BuildResponse("HTTP/1.1 400 Bad Request", "text/plain; charset=utf-8", []byte("bad request")))
		return
	}
	method, target, version := parts[0], parts[1], parts[2]

	if method != "GET" {
		conn.Write(BuildResponse("HTTP/1.1 405 Method Not Allowed", "text/plain; charset=utf-8", []byte("method not allowed")))
		return
	}
	if version != "HTTP/1.1" {
		conn.Write(BuildResponse("HTTP/1.1 505 HTTP Version Not Supported", "text/plain; charset=utf-8", []byte("version not supported")))
		return
	}

	for {
		h, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		h = strings.TrimRight(h, "\r\n")
		if h == "" {
			break
		}
	}

	if target == "/" {
		target = "/index.html"
	}
	if strings.Contains(target, "..") {
		target = "/404.html"
	}

	full := filepath.Join(s.PublicDir, filepath.Clean(target))
	body, err := os.ReadFile(full)
	status := "HTTP/1.1 200 OK"
	if err != nil {
		body, _ = os.ReadFile(filepath.Join(s.PublicDir, "404.html"))
		status = "HTTP/1.1 404 Not Found"
	}

	ct := GuessContentType(full)
	resp := BuildResponse(status, ct, body)
	_, _ = io.Copy(conn, bytes.NewReader(resp))
}
