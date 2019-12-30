// sync synchronizes a CSV file to the Pobox bulk forwarding API.
package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"strings"
	"time"

	whitelabel "github.com/fastmail/pobox-bulk-go"
	"github.com/fastmail/pobox-bulk-go/auth"
	"github.com/golang/glog"
)

var (
	mapFile  = flag.String("map", "", "map file, defaults to stdin")
	domain   = flag.String("domain", "", "domain to configure")
	dryRun   = flag.Bool("n", false, "dry run, don't do anything permanent")
	maxCount = flag.Int("max", 100, "maximum changes to make in a single request")
	authFile = flag.String("authfile", ".pobox-api-auth", "file containing auth secrets")
	timeout  = flag.Duration("timeout", 5*time.Minute, "overall timeout")
	delay    = flag.Duration("delay", 2*time.Second, "delay between modifications")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	if *domain == "" {
		glog.Exitf("required flag --domain not specified")
	}

	u, p := auth.MustLoad(*authFile)

	ctx := context.Background()

	// TODO: context expiry is only checked when http requests are made.
	ctx, cxl := context.WithTimeout(ctx, *timeout)
	defer cxl()

	c := whitelabel.Client{User: u, Pass: p}

	have, err := c.GetRoutes(ctx)
	if err != nil {
		glog.Exitf("error getting routes: %v", err)
	}
	glog.Infof("%d total existing routes", len(have))

	// API returns aliases for all domains covered by the same account.
	have = filterToDomain(have, *domain)
	glog.Infof("%d existing routes for @%s", len(have), *domain)

	want, err := loadCSV(*mapFile)
	if err != nil {
		glog.Exitf("error reading map file: %v", err)
	}
	glog.Infof("map has %d routes", len(want))
	glog.V(1).Infof("details: %v", want)

	mod := computeChanges(have, want)
	glog.Infof("found %d changes to make", len(mod))
	mods := splitChanges(mod, *maxCount)

	for _, m := range mods {
		glog.Infof("modifying %d routes", len(m))
		glog.V(1).Infof("details: %v", m)
		if *dryRun {
			// TODO: consider implementing dryrun with a no-op interface implementation
			glog.Infof("not making modifications due to dryRun flag")
			continue
		}
		_, err = c.SetRoutes(ctx, m)
		if err != nil {
			glog.Fatalf("error setting routes: %v", err)
		}
		time.Sleep(*delay)
	}
	glog.Info("🙂")
}

func computeChanges(have, want whitelabel.Routes) whitelabel.Routes {
	// additions are in want but not have
	// deletions are in have but not want
	// changes are differences in the Fwd.

	mod := make(whitelabel.Routes)

	for ha, hr := range have {
		wr, ok := want[ha]
		if !ok {
			// delete
			mod[ha] = nil
			continue
		}
		if hr.Fwd != wr.Fwd {
			// change
			mod[ha] = wr
		}
	}

	for wa, wr := range want {
		if _, ok := have[wa]; !ok {
			// add
			mod[wa] = wr
		}
		// don't have to check for changes on this pass
	}

	return mod
}

// loadCSV reads a CSV where the first column is the alias name, and the second
// is the destination.  It defaults to reading from stdin if no filename is
// provided.
func loadCSV(fn string) (whitelabel.Routes, error) {
	routes := make(whitelabel.Routes)
	fh := os.Stdin
	if fn != "" {
		var err error
		fh, err = os.Open(fn)
		if err != nil {
			return nil, err
		}
	}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		l := scanner.Text()
		// ignore comments
		if strings.HasPrefix(l, "#") {
			continue
		}
		ms := strings.Split(l, ",")
		if len(ms) != 2 {
			glog.Infof("can't parse %q: ", l)
			continue
		}
		alias := ms[0] + "@" + *domain
		// Pobox normalizes all aliases to lowercase.  If there's multiple (case
		// different) entries in a slowforward mapping file, the last one will
		// win.
		alias = strings.ToLower(strings.TrimSpace(alias))
		routes[alias] = &whitelabel.Route{Fwd: strings.TrimSpace(ms[1])}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

func filterToDomain(in whitelabel.Routes, domain string) whitelabel.Routes {
	var out = make(whitelabel.Routes)
	for k, v := range in {
		ps := strings.Split(k, "@")
		if domain == ps[1] {
			out[k] = v
		}
	}
	return out
}

func splitChanges(mod whitelabel.Routes, limit int) []whitelabel.Routes {
	if len(mod) == 0 {
		return []whitelabel.Routes{}
	}
	if len(mod) < limit {
		return []whitelabel.Routes{mod}
	}

	i := -1
	var ret []whitelabel.Routes
	var new whitelabel.Routes

	// TODO: this logic works, but is hairier than it should be.
	for k, v := range mod {
		i++
		if new == nil {
			new = make(whitelabel.Routes)
		}
		new[k] = v
		if i%limit == limit-1 {
			if len(new) > 0 {
				ret = append(ret, new)
			}
			new = nil
		}
	}
	ret = append(ret, new)

	return ret
}
