package main

import (
	"context"
	"fmt"
	"github.com/bswaterb/goX/brpc"
	"github.com/bswaterb/goX/brpc/xclient"
	"github.com/bswaterb/goX/bttp/middlewares"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bswaterb/goX/bttp"
)

func main() {
	RunBrpc2()
}

func RunBttp() {
	e := bttp.NewEngine()

	e.GET("/", func(c *bttp.Context) {
		c.String(http.StatusOK, "Hello Test")
	})

	e.POST("/path1", func(c *bttp.Context) {
		fmt.Println("req info: ", c.Method, c.Path, c.Req.RemoteAddr)
		c.JSON(http.StatusOK, map[string]string{
			"field1": "value1",
			"field2": "value2",
		})
	})

	e.GET("/path2/:name/search", func(c *bttp.Context) {
		name := c.Params["name"]
		c.JSON(http.StatusOK, map[string]string{
			"name": name,
		})
	})

	e.GET("/path2/a/b/*filepath", func(c *bttp.Context) {
		path := c.Params["filepath"]
		c.JSON(http.StatusOK, map[string]string{
			"filepath": path,
		})
	})

	e.GET("/hello", func(c *bttp.Context) {
		c.String(http.StatusOK, "query params 'name' value is : %s\n", c.Query("name"), c.Path)
	})

	e.Run(":8080")
}

func RunBttpGroup() {
	e := bttp.NewEngine()
	g1 := e.Group("/v1")
	g1.Use(middlewares.NewLoggerMiddleware(), middlewares.NewRecoveryMiddleware())
	{
		g1.GET("/", func(c *bttp.Context) {
			c.String(http.StatusOK, "Hello Test")
		})
		g1.POST("/path1", func(c *bttp.Context) {
			fmt.Println("req info: ", c.Method, c.Path, c.Req.RemoteAddr)
			c.JSON(http.StatusOK, map[string]string{
				"field1": "value1",
				"field2": "value2",
			})
		})
	}

	g2 := e.Group("/v2")
	{
		g2.GET("/", func(c *bttp.Context) {
			c.String(http.StatusOK, "Hello Test")
		})
		g2.GET("/path2/:name/search", func(c *bttp.Context) {
			name := c.Params["name"]
			c.JSON(http.StatusOK, map[string]string{
				"name": name,
			})
		})
		g2.GET("/path2/a/b/*filepath", func(c *bttp.Context) {
			path := c.Params["filepath"]
			c.JSON(http.StatusOK, map[string]string{
				"filepath": path,
			})
		})
	}

	g3 := e.Group("/auth")
	{
		g3.POST("/login", func(c *bttp.Context) {
			c.JSON(http.StatusOK, bttp.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
		g3.POST("/logout", func(c *bttp.Context) {
			fmt.Printf("user %s logout", c.PostForm("username"))
			c.JSON(http.StatusOK, bttp.H{
				"code": 200,
				"msg":  "success",
			})
		})
	}

	e.Run(":8080")
}

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startBrpcServer1(addr chan string) {
	var foo Foo
	if err := brpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	brpc.Accept(l)
}

func RunBrpc1() {
	addr := make(chan string)
	go startBrpcServer1(addr)
	client, _ := brpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if err := client.Call(ctx, "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func startBrpcServer2(addrCh chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := brpc.NewServer()
	_ = server.Register(&foo)
	addrCh <- l.Addr().String()
	server.Accept(l)
}

func traceCallingLog(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			traceCallingLog(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			traceCallingLog(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			traceCallingLog(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func RunBrpc2() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)
	// start two servers
	go startBrpcServer2(ch1)
	go startBrpcServer2(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}
