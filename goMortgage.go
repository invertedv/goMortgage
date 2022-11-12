// goMortgage fits models that forecast the performance of goMortgage loans at the loan level.
//
// The model-build specs files is in key:val format.  Legal values are:
//
//   - model: <model name>
//   - buildData: <yes>, <no>
//   - buildModel: <yes>, <no>
//   - mtgDb: <ClickHouse table>
//   - econDb: <ClickHouse table>
//   - econFields: <zip3>, <zip>   key of econDb to use for joins to mtgDb
//   - outDir: <full path to output directory for this run>
//   - pass1Strat: <ClickHouse table>  name of table to store strats from pass 1 sampling
//   - pass1Table: <ClickHouse table>  name of table to create with pass 1 sample of loans
//   - pass2Strat: <ClickHouse table>  name of table to store strats from pass 2 sampling
//   - pass2Table: <ClickHouse table>  name of table to create with pass 2 sample of loans
//   - modelTable: <Clickhouse table>  name of table to create with model-build sample
//   - log: <log file>  name of log file to create in outDir
//   - show: <yes>, <no>  if yes, show all graphs in browser, too.
//   - inputModel: <name>.  Name of input model to add to model-build pipeline.
//   - <name>Location: <path to directory containing model <name>>
//   - <name>Targets: <field name:targets> (repeating with semi-colons).  For model <name>, the field name to create in
//     pipeline along with the columns to coalesce into this output.
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
		if e := log.Close(); e != nil {
			panic(e)
		}
	}()
	defer func() {
		if e := conn.Close(); e != nil {
			panic(e)
		}
	}()

	start := time.Now()

	if specs.buildData() {
		if e := data(specs, conn, log); e != nil {
			panic(e)
		}
	}

	if specs.buildModel() {
		if e := model(specs, conn, log); e != nil {
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
