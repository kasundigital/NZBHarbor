package nntp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kasundigital/NZBHarbor/internal/config"
)

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
	code, msg, err := c.readResponse()
	if err != nil || (code != 200 && code != 201) {
		conn.Close()
		return nil, fmt.Errorf("banner %d %s: %w", code, msg, err)
	}
	if s.Username != "" {
		if err := c.cmdExpect("AUTHINFO USER "+s.Username, 281, 381); err != nil {
			c.Close()
			return nil, err
		}
		if s.Password != "" {
			if err := c.cmdExpect("AUTHINFO PASS "+s.Password, 281); err != nil {
				c.Close()
				return nil, err
			}
		}
	}
	return c, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	_ = c.command("QUIT")
	return c.conn.Close()
}

func (c *Client) Body(messageID string) ([]byte, error) {
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
		return nil, fmt.Errorf("article unavailable: %d %s", code, msg)
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
	return fmt.Errorf("%s: %d %s", strings.Fields(cmd)[0], code, msg)
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
