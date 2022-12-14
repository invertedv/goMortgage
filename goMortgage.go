package main

import (
	"flag"
	"fmt"
	"time"
)

const (
	maxMemoryDef  = 40000000000
	maxGroupByDef = 20000000000

	yes = "yes"
)

// for strconv.ParseInt
const (
	base10 = 10
	bits32 = 32
	bits64 = 64
)

func main() {
	// modeling options
	specsFile := flag.String("specs", "", "string")

	// ClickHouse credentials
	host := flag.String("host", "127.0.0.1", "string") // ClickHouse db
	user := flag.String("user", "", "string")          // ClickHouse username
	pw := flag.String("pw", "", "string")              // password for user

	// ClickHouse options
	maxMemory := flag.Int64("memory", maxMemoryDef, "int64")
	maxGroupby := flag.Int64("groupby", maxGroupByDef, "int64")

	flag.Parse()

	specs, conn, log, e := inits(*host, *user, *pw, *specsFile, *maxMemory, *maxGroupby)
	if e != nil {
		panic(e)
	}
	defer func() {
		if ex := log.Close(); e != nil {
			panic(ex)
		}
	}()
	defer func() {
		if ex := conn.Close(); e != nil {
			panic(ex)
		}
	}()

	start := time.Now()

	if specs.buildData() {
		if e = data(specs, conn, log); e != nil {
			panic(e)
		}
	}

	if specs.buildModel() {
		if e = model(specs, conn, log); e != nil {
			panic(e)
		}
	}

	if specs.biasCorrect() {
		if e = biasCorrect(specs, conn, log); e != nil {
			panic(e)
		}
	}

	if specs.assessModel() {
		if e := assessModel(specs, conn, log); e != nil {
			panic(e)
		}
	}

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("total run time: %0.1f minutes", elapsed), true)
}
