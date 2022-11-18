package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/invertedv/chutils"
	sea "github.com/invertedv/seafan"
)

var (
	//go:embed sql/passes/pass1.sql
	withPass1 string

	//go:embed sql/passes/pass2.sql
	withPass2 string

	//go:embed sql/passes/pass3.sql
	withPass3 string

	//go:embed sql/fannie/mtgFields.sql
	fannieMtgFields string

	//go:embed sql/fannie/goodloan.sql
	fannieGoodLoan string

	//go:embed sql/fannie/pass1Fields.sql
	fanniePass1 string

	//go:embed sql/fannie/pass2Fields.sql
	fanniePass2Fields string

	//go:embed sql/fannie/pass3Fields.sql
	fanniePass3Calcs string

	//go:embed sql/econ/zip3With.sql
	econTable3 string

	//go:embed sql/econ/zip3Fields.sql
	econFields3 string
)

// inits initializes exported vars Specs, Conn, LogFile
func inits(host, user, pw, specsFile string, maxMemory, maxGroupBy int64) (specsMap, *chutils.Connect, *os.File, error) {
	var e error

	sea.Verbose = false
	conn, e := chutils.NewConnect(host, user, pw, clickhouse.Settings{
		"max_memory_usage":                   maxMemory,
		"max_bytes_before_external_group_by": maxGroupBy,
	})
	if e != nil {
		return nil, nil, nil, e
	}

	specs, e := readSpecsMap(specsFile)
	if e != nil {
		return nil, nil, nil, e
	}

	if er := specs.check(); er != nil {
		return nil, nil, nil, er
	}

	if specs.buildData() || specs.buildModel() {
		if er := os.RemoveAll(specs["outDir"]); er != nil {
			return nil, nil, nil, er
		}

		if er := os.MkdirAll(specs["outDir"], os.ModePerm); er != nil {
			return nil, nil, nil, er
		}
	}
	switch specs.buildData() || specs.buildModel() {
	case true:
		if er := os.RemoveAll(specs["outDir"]); er != nil {
			return nil, nil, nil, er
		}

		if er := os.MkdirAll(specs["outDir"], os.ModePerm); er != nil {
			return nil, nil, nil, er
		}

		if specs["modelDir"], e = makeSubDir(specs["outDir"], specs.modelKey()); e != nil {
			return nil, nil, nil, e
		}

	case false:
		specs["modelDir"] = fmt.Sprintf("%s%s/", slash(specs["outDir"]), slash(specs.modelKey()))

		if er := os.RemoveAll(fmt.Sprintf("%s%s", slash(specs["outDir"]), specs.graphsKey())); er != nil {
			return nil, nil, nil, er
		}
	}

	if specs["graphDir"], e = makeSubDir(specs["outDir"], specs.graphsKey()); e != nil {
		return nil, nil, nil, e
	}

	if specs["valDir"], e = makeSubDir(specs["graphDir"], "validation"); e != nil {
		return nil, nil, nil, e
	}

	if specs["margDir"], e = makeSubDir(specs["graphDir"], "marginal"); e != nil {
		return nil, nil, nil, e
	}

	if specs["costDir"], e = makeSubDir(specs["graphDir"], "cost"); e != nil {
		return nil, nil, nil, e
	}

	if specs["stratsDir"], e = makeSubDir(specs["graphDir"], "strats"); e != nil {
		return nil, nil, nil, e
	}

	if specs["curvesDir"], e = makeSubDir(specs["graphDir"], "curves"); e != nil {
		return nil, nil, nil, e
	}

	if specs["inputDir"], e = makeSubDir(specs.modelDir(), "inputModels"); e != nil {
		return nil, nil, nil, e
	}

	// just doing assessment ... append to existing log file, don't copy .gom file or input models
	if !specs.buildData() && !specs.buildModel() {
		// load up the needed cts and cat feature list in this case
		if er := specs.features(specs.modelDir()); er != nil {
			return nil, nil, nil, er
		}

		logFile, er := os.OpenFile(specs["outDir"]+"model.log", os.O_APPEND|os.O_WRONLY, os.ModePerm)

		if er != nil {
			return nil, nil, nil, er
		}

		return specs, conn, logFile, nil
	}

	// copy over the spec file
	if er := copyFile(specsFile, specs["outDir"]+"model.gom"); er != nil {
		return nil, nil, nil, er
	}

	if er := specs.inputModels(); er != nil {
		return nil, nil, nil, er
	}

	logFile, e := os.Create(specs["outDir"] + "model.log")

	return specs, conn, logFile, e
}

// buildQuery builds a query from a skeleton.  Any time the skeleton contains <key>-where key is a key in replacers-
// it is replaced with the value map[key].
func buildQuery(baseWith string, replacers map[string]string) string {
	qry := baseWith
	for k, v := range replacers {
		krep := fmt.Sprintf("<%s>", k)
		// add whitespace around v
		qry = strings.ReplaceAll(qry, krep, fmt.Sprintf(" %s ", v))
	}

	return fmt.Sprintf("%s SELECT * FROM d", qry)
}

// slash appends a trailing backslash if there is not one
func slash(path string) string {
	if path[len(path)-1:] == "/" {
		return path
	}

	return path + "/"
}

// copyFiles recursively copies files from fromDir to toDir
func copyFiles(fromDir, toDir string) error {
	fromDir = slash(fromDir)
	toDir = slash(toDir)

	dirList, e := os.ReadDir(fromDir)
	if e != nil {
		return e
	}

	// skip if directory is empty
	if len(dirList) == 0 {
		return nil
	}

	if e := os.MkdirAll(toDir, os.ModePerm); e != nil {
		return e
	}

	for _, file := range dirList {
		if file.IsDir() {
			if e := copyFiles(fromDir+file.Name(), toDir+file.Name()); e != nil {
				return e
			}
			continue
		}

		if e := copyFile(fmt.Sprintf("%s%s", fromDir, file.Name()), fmt.Sprintf("%s%s", toDir, file.Name())); e != nil {
			return e
		}
	}

	return nil
}

// makeSubDir creates subDir under dir and returns the full path to it.
func makeSubDir(dir, subDir string) (path string, err error) {
	path = slash(dir) + slash(subDir)

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(slash(dir)+subDir, os.ModePerm)
		return
	}

	return path, nil
}

// plotCosts plots the cost function value vs epoch with title costName
func plotCosts(fit *sea.Fit, costName string, specs specsMap) error {
	var best string

	switch fit.OutCosts() == nil {
	case true:
		best = fmt.Sprintf("Best Epoch: %d (model Sample)", fit.BestEpoch())
	case false:
		best = fmt.Sprintf("Best Epoch: %d (Validation Sample)", fit.BestEpoch())
	}

	if e := fit.InCosts().Plot(&sea.PlotDef{
		Title:    fmt.Sprintf("model Sample Cost-%s", costName),
		XTitle:   "Epoch",
		YTitle:   "Cost",
		STitle:   best,
		Legend:   false,
		Height:   specs.plotHeight(),
		Width:    specs.plotWidth(),
		Show:     specs.plotShow(),
		FileName: specs.costDir() + "modelSample.html",
	}, true); e != nil {
		return e
	}

	if fit.OutCosts() == nil {
		return nil
	}

	if e := fit.OutCosts().Plot(&sea.PlotDef{
		Title:    fmt.Sprintf("Validation Sample Cost-%s", costName),
		XTitle:   "Epoch",
		YTitle:   "Cost",
		STitle:   best,
		Legend:   false,
		Height:   specs.plotHeight(),
		Width:    specs.plotWidth(),
		Show:     specs.plotShow(),
		FileName: specs.costDir() + "validationSample.html",
	}, true); e != nil {
		return e
	}

	return nil
}

// toSlice returns a slice by splitting str on sep
func toSlice(str, sep string) []string {
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "\n", "")

	// check for no entries
	if str == "" {
		return nil
	}
	return strings.Split(str, sep)
}

// copyFile copies sourceFile to destFile
func copyFile(sourceFile, destFile string) error {
	inFile, e := os.Open(sourceFile)
	if e != nil {
		return e
	}
	defer func() { _ = inFile.Close() }()

	outFile, e := os.Create(destFile)
	if e != nil {
		return e
	}
	defer func() { _ = outFile.Close() }()

	_, e = io.Copy(outFile, inFile)

	return e
}

// logger makes a log entry.
func logger(log *os.File, text string, toConsole bool) {
	_, _ = fmt.Fprintln(log, text)
	if toConsole {
		fmt.Println(text)
	}
}
