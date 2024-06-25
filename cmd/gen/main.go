package main

import (
	"github.com/jghiloni/watchedsky-social/backend/bsky"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	genCfg := cbg.Gen{
		MaxStringLength: 1_000_000,
	}

	if err := genCfg.WriteMapEncodersToFile("backend/bsky/cbor_gen.go", "bsky", bsky.Alert{}); err != nil {
		panic(err)
	}
}
