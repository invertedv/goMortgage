// A Self-Service Program to Build Mortgage Models
// What is a mortgage model?
//
// A mortgage model is a predictive model that forecasts some aspect of a mortgage's performance.
// Common performance metrics are delinquency, default, severity and prepayment.
//
// Such models are built on historical data and may be either at a loan level or pool level.
// goMortgage builds loan-level models.
//
// # About goMortgage
//
// goMortgage is an app that builds mortgage forecasting models.
//
// What aspects of mortgage performance can be modeled? Really, anything you can think of.  The software
// is agnostic about the model target and features.
//
// goMortgage takes a text file (*.gom) you create to direct all aspects of the modeling process
// from building datasets to model assessment.
//
// Since goMortgage is open source, goMortgage can be modified to suit your needs.
// Out of the box, it is configured to work with Freddie and Fannie data. There are
// instructions to set up your own data sources.
//
// # Why goMortgage?
//
// I wrote goMortgage for my own use. I made it open source because I thought others might find it useful, too.
// Who?
//   - People who are more interested in the output of the model than building the infrastructure --
//     say people on a trading desk or academics.  Especially those who want to focus on Fannie and Freddie,
//     since I have packages ([fannie],
//     [freddie])
//     to build this data.
//   - Those who don't have the bandwidth to start from scratch but enough to make the modifications for new data sources.
//
// # Requirements
//
// You need to be able to put up an instance of [ClickHouse].
//
// Currently, goMortgage doesn't support building DNNs on a GPU, but the run speed has been fine. You'll want
// 64GB of RAM, though you could get by with less probably.
//
// The raw files for Freddie and Fannie are quite large, so you'll want a few TBs of disk.
//
// And, of course, you need to be able to compile Go.
//
// # Documentation
//
// For details, see the [docs].
//
// [docs]: https://invertedv.github.io/goMortgage
// [ClickHouse]: https://clickhouse.com/clickhouse
// [freddie]: https://pkg.go.dev/github.com/invertedv/freddie)
// [fannie]: https://pkg.go.dev/github.com/invertedv/fannie
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
