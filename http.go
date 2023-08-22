package tcpinfo

import (
	"context"
	"errors"
	"net"
)

type ctxConnKey struct{}

// HTTPConnFDMiddleware extracts a file descriptor of client connection
// and returns a copy of parent context in which passed fd
func HTTPConnFDMiddleware(ctx context.Context, conn net.Conn) context.Context {
	return setConn2Ctx(ctx, conn)
}

func setConn2Ctx(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, ctxConnKey{}, conn)
}

// ExtractFDFromCtx returns extracted file descriptor from context
// Bool flag indicates whether fd was found
func ExtractFDFromCtx(ctx context.Context) (uintptr, bool) {
	fd, ok := ctx.Value(ctxConnKey{}).(uintptr)
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
