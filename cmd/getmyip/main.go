// Command getmyip prints the geolocation of an IP address — or your own IP when
// called with no argument. Powered by the getmyip.pro API (privacy-first).
//
//	getmyip                 # your IP + geo
//	getmyip 8.8.8.8         # geo for an address
//	getmyip -json 8.8.8.8   # raw JSON
//	getmyip -field country 8.8.8.8
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	getmyip "github.com/soos-labs/getmyip-go"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("getmyip", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "print raw JSON")
	field := fs.String("field", "", "print a single field (ip, country, country_code, region, city, org, asn, timezone)")
	key := fs.String("key", os.Getenv("GETMYIP_API_KEY"), "API key (or set GETMYIP_API_KEY) — optional")
	fs.Usage = func() {
		fmt.Fprint(stderr, "getmyip — IP geolocation (https://getmyip.pro)\n\nUsage:\n  getmyip [flags] [ip]\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}

	opts := []getmyip.Option{getmyip.WithAPIKey(*key)}
	// свой инстанс (self-hosted) или мок в тестах
	if base := os.Getenv("GETMYIP_BASE_URL"); base != "" {
		opts = append(opts, getmyip.WithBaseURL(base))
	}
	c := getmyip.New(opts...)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	var (
		res *getmyip.Result
		err error
	)
	if fs.NArg() > 0 {
		res, err = c.Lookup(ctx, fs.Arg(0))
	} else {
		res, err = c.Me(ctx)
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	switch {
	case *field != "":
		v, ok := fieldOf(res, *field)
		if !ok {
			fmt.Fprintln(stderr, "unknown field:", *field)
			return 2
		}
		fmt.Fprintln(stdout, v)
	case *jsonOut:
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Fprintln(stdout, string(b))
	default:
		printHuman(stdout, res)
	}
	return 0
}

func printHuman(w io.Writer, r *getmyip.Result) {
	loc := r.Country
	if r.City != "" {
		loc = r.City + ", " + loc
	}
	if r.CountryCode != "" {
		loc += " (" + r.CountryCode + ")"
	}
	fmt.Fprintf(w, "IP:       %s\n", r.IP)
	if r.IsPrivate {
		fmt.Fprintln(w, "Private/reserved address — no public geolocation.")
		return
	}
	if loc != "" {
		fmt.Fprintf(w, "Location: %s\n", loc)
	}
	if r.Org != "" {
		fmt.Fprintf(w, "Network:  %s (AS%d)\n", r.Org, r.ASN)
	}
	if r.Timezone != "" {
		fmt.Fprintf(w, "Timezone: %s\n", r.Timezone)
	}
}

func fieldOf(r *getmyip.Result, name string) (string, bool) {
	switch strings.ToLower(name) {
	case "ip":
		return r.IP, true
	case "country":
		return r.Country, true
	case "country_code", "cc":
		return r.CountryCode, true
	case "region":
		return r.Region, true
	case "city":
		return r.City, true
	case "postal":
		return r.Postal, true
	case "org", "isp":
		return r.Org, true
	case "asn":
		return fmt.Sprintf("%d", r.ASN), true
	case "timezone", "tz":
		return r.Timezone, true
	default:
		return "", false
	}
}
