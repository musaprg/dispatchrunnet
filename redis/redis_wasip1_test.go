//go:build wasip1

package redis_test

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/musaprg/dispatchrunnet/wasip1"
)

func TestRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:   "localhost:6379",
		Dialer: wasip1.DialContext,
	})
	defer client.Close()

	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	echo, err := client.Echo(ctx, "Hello, World!\n").Result()
	if err != nil {
		t.Fatal(err)
	}
	if echo != "Hello, World!\n" {
		t.Errorf("wrong echo message received: %q", echo)
	}
}
