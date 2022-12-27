package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	currencypb "something/pb"
	"sort"
	"sync"
	"syscall"
	"time"
)

func init() {
	rand.Seed(1) //time.Now().UnixNano())
}

var addr = flag.String("addr", "localhost:8080", "the address to connect to")

func factorial(n chan int, ch chan int) {

	result := 1
	nn := <-n
	for i := 1; i <= nn; i++ {
		result *= i
	}
	fmt.Println(n, "-", result)

	ch <- result
}

type WorkNumber struct {
	i int
	v int
}

func main() {
	flag.Parse()
	var pool, _ = x509.SystemCertPool()

	var insecureTransportCreds = credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            pool,
	})

	ctx, ccancel := signal.NotifyContext(context.Background(), os.Kill, syscall.SIGINT, syscall.SIGTERM)
	defer ccancel()
	//sctx, scancel := context.WithCancel(ctx)
	//defer scancel()
	//dset := []int{1, 3, 4, 7, 2, 6, 8, 4435, 45}
	//dres := make([]int, len(dset))
	//
	//intCh := make(chan int)
	//go factorial(intCh2, intCh)
	//fmt.Println(<-intCh)
	//fmt.Println("The End")
	//
	//return
	////wg := new(sync.WaitGroup)
	//

	c := make(chan WorkNumber)
	sc := make(chan WorkNumber)
	start := time.Now()
	wg := new(sync.WaitGroup)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range c {
				select {
				case <-time.After(time.Millisecond * 1000):
				case <-ctx.Done():
					return

				}
				v.v = v.v * 2
				select {
				case sc <- v:
				case <-ctx.Done():
					return
				}

			}
		}()
	}
	go func() {
		wg.Wait()
		close(sc)
	}()

	in := make([]int, 10)
	go func() {
		defer close(c)
		for i := 0; i < len(in); i++ {
			v := rand.Intn(10)
			in[i] = v
			select {
			case c <- WorkNumber{i: i, v: v}:
			case <-ctx.Done():
				return
			}
		}
	}()
	out := make([]WorkNumber, 0)
	for v := range sc {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].i < out[j].i
	})
	out2 := make([]int, len(out))
	for i, v := range out {
		out2[i] = v.v
	}
	fmt.Printf("in  %v\n", in)
	fmt.Printf("out %v\n", out2)
	log.Printf("time %v\n", time.Now().Sub(start))

	////sem := semaphore.NewWeighted(4)
	//for i_, v_ := range dset {
	//	c <- v_
	//	i := i_
	//	//v := v_
	//	//if err := sem.Acquire(ctx, 1); err != nil {
	//	//	break
	//	//}
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		//defer sem.Release(1)
	//		dres[i] = <-c * 2
	//		<-time.After(time.Millisecond * 300)
	//	}()
	//}
	//wg.Wait()
	//log.Printf("dres %v\n", dres)
	//log.Printf("time %v\n", time.Now().Sub(start))
	return

	//insecureTransportCreds = insecure.NewCredentials()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecureTransportCreds),
	}
	conn, err := grpc.Dial(*addr, opts...)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	var country = currencypb.Countries(rand.Intn(5))

	defer conn.Close()

	client := currencypb.NewCurrencyClient(conn)
	request := &currencypb.CurRequest{
		Message: country,
	}
	response, err := client.Do(ctx, request)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	fmt.Println(response.Currency, response.Value)

	stream, err := client.DoStrm(ctx, request)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	for ctx.Err() == nil {
		if feature, err := stream.Recv(); err != nil {
			if err == io.EOF {
				break
			}
			grpclog.Warningf("fail to dial: %v", err)
		} else {
			log.Println(feature.Currency, feature.Value)
		}
	}
	log.Println("Done")
}
