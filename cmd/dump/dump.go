package main

import (
	"context"
	"flag"
	"fmt"

	whitelabel "github.com/fastmail/pobox-bulk-go"
	"github.com/fastmail/pobox-bulk-go/auth"
	"github.com/golang/glog"
	"time"
)

var (
	authFile = flag.String("authfile", ".pobox-api-auth", "file containing auth secrets")
	timeout  = flag.Duration("timeout", 5*time.Minute, "overall timeout")
	delim    = flag.String("d", "\t", "output delimiter")
)

// WhitelabelClient defines the subset of whitelabel.Client we use in this package.
type WhitelabelClient interface {
	GetRoutes(context.Context) (whitelabel.Routes, error)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	u, p := auth.MustLoad(*authFile)

	ctx := context.Background()

	// TODO: context expiry is only checked when http requests are made.
	ctx, cxl := context.WithTimeout(ctx, *timeout)
	defer cxl()

	var c WhitelabelClient = &whitelabel.Client{User: u, Pass: p}

	routes, err := c.GetRoutes(ctx)
	if err != nil {
		glog.Exitf("error getting routes: %v", err)
	}

	for a, r := range routes {
		fmt.Printf("%s%s%s\n", a, *delim, r.Fwd)
	}
}
