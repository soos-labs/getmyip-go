package getmyip_test

import (
	"context"
	"fmt"

	getmyip "github.com/soos-labs/getmyip-go"
)

// Look up any address; the anonymous tier needs no API key.
func ExampleClient_Lookup() {
	c := getmyip.New() // getmyip.New(getmyip.WithAPIKey("...")) to raise limits
	r, err := c.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(r.Country, r.Org)
}

// Geolocation for your own public IP.
func ExampleClient_Me() {
	c := getmyip.New()
	me, err := c.Me(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(me.IP, me.City)
}
