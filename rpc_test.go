package rpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

type testRPC int

func (testRPC) Test(ctx context.Context, in string, out *string) error {
	*out = in
	return nil
}
func (testRPC) TestWithCancel(ctx context.Context, in string, out *string) error {
	timer := time.NewTimer(time.Second * 5)
	fmt.Printf("server: started (%s)\n", ContextID(ctx))
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("server: canceled (%s)\n", ContextID(ctx))
			return ctx.Err()
		case <-timer.C:
			fmt.Printf("server: complete (%s)\n", ContextID(ctx))
			*out = in
			return nil
		}
	}
}

// TestRPC is a super simple functional test of calling over TCP
func TestRPC(t *testing.T) {
	srv := NewServer()
	if err := srv.RegisterName("Test", new(testRPC)); err != nil {
		t.Fatalf("failed to register service: %v", err)
		return
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		srv.Accept(context.Background(), ln)
		wg.Done()
	}()

	client, err := Dial(context.Background(), ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		ln.Close()
		t.Fatalf("failed to dial: %v", err)
		return
	}

	in := "test"
	var out string
	if err := client.Call(context.Background(), "Test.Test", in, &out); err != nil {
		ln.Close()
		t.Fatalf("failed to call Test.Test: %v", err)
		return
	}

	if in != out {
		t.Errorf("Unexpected output: expected %q got %q", in, out)
	}

	ln.Close()
	wg.Wait()
}
func TestCancelableRPC(t *testing.T) {
	srv := NewServer()
	if err := srv.RegisterName("Test", new(testRPC)); err != nil {
		t.Fatalf("failed to register service: %v", err)
		return
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		srv.Accept(context.Background(), ln)
		wg.Done()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	client, err := Dial(ctx, ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		ln.Close()
		t.Fatalf("client: failed to dial: %v", err)
		return
	}

	doCancel := false // change this to modify server behavior

	go func() {
		in := "test"
		var out string
		if err := client.Call(ctx, "Test.TestWithCancel", in, &out); err != nil {
			t.Fatalf("client: failed to call Test.TestWithCancel: %v", err)
			time.Sleep(time.Second)
			ln.Close()
			return
		}
		if !doCancel && in != out {
			t.Errorf("client: Unexpected output: expected %q got %q", in, out)
		} else if doCancel && out != "" {
			t.Errorf("client: Unexpected output: expected %q got %q", "", out)
		}
	}()

	time.Sleep(time.Second * 2)
	_ = cancel
	if doCancel {
		cancel()
	}
	time.Sleep(time.Second * 10)

	ln.Close()
	wg.Wait()
}
