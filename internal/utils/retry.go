package utils

import (
	"errors"
	"net"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var RetriableErrors = []error{
	net.ErrClosed,
	&net.OpError{},
	&net.DNSError{},
}

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	for _, re := range RetriableErrors {
		if errors.Is(err, re) {
			return true
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.SerializationFailure:
			return true
		}
	}

	return false
}

func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil || !IsRetriable(err) {
			break
		}

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return err
}
