//go:build wasip1

package mysql

import (
	"context"
	"net"

	"github.com/go-sql-driver/mysql"
	"github.com/musaprg/dispatchrunnet/wasip1"
)

func init() {
	networks := []string{"tcp", "tcp4", "tcp6"}
	for i := range networks {
		network := networks[i]
		mysql.RegisterDialContext(network, func(ctx context.Context, address string) (net.Conn, error) {
			return wasip1.DialContext(ctx, network, address)
		})
	}
}
