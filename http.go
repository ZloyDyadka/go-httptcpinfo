package tcpinfo

import (
	"context"
	"errors"
	"log"
	"net"
)

type ctxFDKey struct{}

// HTTPConnFDMiddleware extracts a file descriptor of client connection
// and returns a copy of parent context in which passed fd
func HTTPConnFDMiddleware(ctx context.Context, c net.Conn) context.Context {
	fd, err := ExtractFDFromConn(c)
	if err != nil {
		log.Println("can't extract FD from connection: ", err)
		return ctx
	}

	return setFD2Ctx(ctx, fd)
}

func setFD2Ctx(ctx context.Context, fd uintptr) context.Context {
	return context.WithValue(ctx, ctxFDKey{}, fd)
}

// ExtractFDFromCtx returns extracted file descriptor from context
// Bool flag indicates whether fd was found
func ExtractFDFromCtx(ctx context.Context) (uintptr, bool) {
	fd, ok := ctx.Value(ctxFDKey{}).(uintptr)
	return fd, ok
}

var (
	ErrNoTCPConn = errors.New("conn is not a net.TCPConn")
)

// ExtractFDFromConn returns a file descriptor of connection
func ExtractFDFromConn(conn net.Conn) (uintptr, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, ErrNoTCPConn
	}

	file, err := tcpConn.File()
	if err != nil {
		return 0, errors.New("get file of tcp connection")
	}

	return file.Fd(), nil
}
