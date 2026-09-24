package nntp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kasundigital/NZBHarbor/internal/config"
)

const articleIOTimeout = 90 * time.Second

type ResponseError struct {
	Code    int
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("NNTP %d %s", e.Code, e.Message)
}

func IsArticleMissing(err error) bool {
	var resp *ResponseError
	if !errors.As(err, &resp) {
		return false
	}
	return resp.Code == 423 || resp.Code == 430
}

type Client struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

func Dial(s config.NewsServer) (*Client, error) {
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	d := net.Dialer{Timeout: 15 * time.Second}
	var conn net.Conn
	var err error
	if s.TLS {
		conn, err = tls.DialWithDialer(&d, "tcp", addr, &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = d.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, r: bufio.NewReaderSize(conn, 128*1024), w: bufio.NewWriter(conn)}
	_ = conn.SetDeadline(time.Now().Add(articleIOTimeout))
	code, msg, err := c.readResponse()
	if err != nil || (code != 200 && code != 201) {
		_ = conn.Close()
		if err != nil {
			return nil, fmt.Errorf("banner: %w", err)
		}
		return nil, &ResponseError{Code: code, Message: msg}
	}
	if s.Username != "" {
		if err := c.cmdExpect("AUTHINFO USER "+s.Username, 281, 381); err != nil {
			_ = c.Close()
			return nil, err
		}
		if s.Password != "" {
			if err := c.cmdExpect("AUTHINFO PASS "+s.Password, 281); err != nil {
				_ = c.Close()
				return nil, err
			}
		}
	}
	_ = conn.SetDeadline(time.Time{})
	return c, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	_ = c.command("QUIT")
	return c.conn.Close()
}

func (c *Client) interrupt() {
	if c != nil && c.conn != nil {
		_ = c.conn.SetDeadline(time.Now())
	}
}

func (c *Client) Body(messageID string) ([]byte, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("NNTP connection is closed")
	}
	_ = c.conn.SetDeadline(time.Now().Add(articleIOTimeout))
	defer c.conn.SetDeadline(time.Time{})

	id := strings.TrimSpace(messageID)
	if !strings.HasPrefix(id, "<") {
		id = "<" + id + ">"
	}
	if err := c.command("BODY " + id); err != nil {
		return nil, err
	}
	code, msg, err := c.readResponse()
	if err != nil {
		return nil, err
	}
	if code != 222 {
		return nil, &ResponseError{Code: code, Message: msg}
	}
	var out []byte
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "." {
			break
		}
		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out, nil
}

func (c *Client) cmdExpect(cmd string, ok ...int) error {
	if err := c.command(cmd); err != nil {
		return err
	}
	code, msg, err := c.readResponse()
	if err != nil {
		return err
	}
	for _, x := range ok {
		if code == x {
			return nil
		}
	}
	return &ResponseError{Code: code, Message: msg}
}

func (c *Client) command(s string) error {
	if _, err := fmt.Fprintf(c.w, "%s\r\n", s); err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *Client) readResponse() (int, string, error) {
	line, err := c.r.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	line = strings.TrimSpace(line)
	if len(line) < 3 {
		return 0, line, io.ErrUnexpectedEOF
	}
	code, err := strconv.Atoi(line[:3])
	return code, line, err
}
