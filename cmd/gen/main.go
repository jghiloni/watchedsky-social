package main

import (
	"github.com/jghiloni/watchedsky-social/backend/atproto"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	genCfg := cbg.Gen{
		MaxStringLength: 1_000_000,
	}

	if err := genCfg.WriteMapEncodersToFile("backend/atproto/cbor_gen.go", "atproto", atproto.Alert{}); err != nil {
		panic(err)
	}
}
