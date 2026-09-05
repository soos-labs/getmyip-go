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
	"os"
	"strings"
	"time"

	getmyip "github.com/soos-labs/getmyip-go"
)

func main() {
	jsonOut := flag.Bool("json", false, "print raw JSON")
	field := flag.String("field", "", "print a single field (ip, country, country_code, region, city, org, asn, timezone)")
	key := flag.String("key", os.Getenv("GETMYIP_API_KEY"), "API key (or set GETMYIP_API_KEY) — optional")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "getmyip — IP geolocation (https://getmyip.pro)\n\nUsage:\n  getmyip [flags] [ip]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	c := getmyip.New(getmyip.WithAPIKey(*key))
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	var (
		res *getmyip.Result
		err error
	)
	if flag.NArg() > 0 {
		res, err = c.Lookup(ctx, flag.Arg(0))
	} else {
		res, err = c.Me(ctx)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	switch {
	case *field != "":
		v, ok := fieldOf(res, *field)
		if !ok {
			fmt.Fprintln(os.Stderr, "unknown field:", *field)
			os.Exit(2)
		}
		fmt.Println(v)
	case *jsonOut:
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	default:
		printHuman(res)
	}
}

func printHuman(r *getmyip.Result) {
	loc := r.Country
	if r.City != "" {
		loc = r.City + ", " + loc
	}
	if r.CountryCode != "" {
		loc += " (" + r.CountryCode + ")"
	}
	fmt.Printf("IP:       %s\n", r.IP)
	if r.IsPrivate {
		fmt.Println("Private/reserved address — no public geolocation.")
		return
	}
	if loc != "" {
		fmt.Printf("Location: %s\n", loc)
	}
	if r.Org != "" {
		fmt.Printf("Network:  %s (AS%d)\n", r.Org, r.ASN)
	}
	if r.Timezone != "" {
		fmt.Printf("Timezone: %s\n", r.Timezone)
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
