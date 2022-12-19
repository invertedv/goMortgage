package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/invertedv/chutils"
	sea "github.com/invertedv/seafan"
)

// embed sql into strings
// BYOD
var (
	//go:embed sql/passes/pass1.sql
	withPass1 string

	//go:embed sql/passes/pass2.sql
	withPass2 string

	//go:embed sql/passes/pass3.sql
	withPass3 string

	//go:embed sql/fannie/mtgFieldsStatic.sql
	fannieMtgFieldsStat string

	//go:embed sql/freddie/mtgFieldsStatic.sql
	freddieMtgFieldsStat string

	//go:embed sql/fannie/mtgFieldsMonthly.sql
	fannieMtgFieldsMon string

	//go:embed sql/freddie/mtgFieldsMonthly.sql
	freddieMtgFieldsMon string

	//go:embed sql/fannie/goodloan.sql
	fannieGoodLoan string

	//go:embed sql/freddie/goodloan.sql
	freddieGoodLoan string

	//go:embed sql/fannie/pass1Fields.sql
	fanniePass1 string

	//go:embed sql/freddie/pass1Fields.sql
	freddiePass1 string

	//go:embed sql/fannie/pass2Fields.sql
	fanniePass2Fields string

	//go:embed sql/freddie/pass2Fields.sql
	freddiePass2Fields string

	//go:embed sql/fannie/pass2FieldsWindow.sql
	fanniePass2FieldsWindow string

	//go:embed sql/freddie/pass2FieldsWindow.sql
	freddiePass2FieldsWindow string

	//go:embed sql/fannie/pass3Fields.sql
	fanniePass3Calcs string

	//go:embed sql/freddie/pass3Fields.sql
	freddiePass3Calcs string

	//go:embed sql/econ/zip3With.sql
	econTable3 string

	//go:embed sql/econ/zip3Fields.sql
	econFields3 string

	// list of all known .gom keys
	//go:embed strings/keys.txt
	allKeys string
)

// inits initializes exported vars Specs, Conn, LogFile
func inits(host, user, pw, specsFile string, maxMemory, maxGroupBy int64) (specsMap, *chutils.Connect, *os.File, error) {
	var e error
	var modelDir string

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

	// check specs make sense
	if er := specs.check(); er != nil {
		return nil, nil, nil, er
	}

	outDir := slash(specs.getkeyVal("outDir", true))
	switch specs.buildData() || specs.buildModel() {
	case true:
		// if we're building the data or the model, clean out the outDir
		if er := os.RemoveAll(outDir); er != nil {
			return nil, nil, nil, er
		}

		if er := os.MkdirAll(outDir, os.ModePerm); er != nil {
			return nil, nil, nil, er
		}

		if modelDir, e = makeSubDir(outDir, specs.modelKey()); e != nil {
			return nil, nil, nil, e
		}
		specs.assign("modelDir", modelDir)

	case false:
		// otherwise, clean out the graphs directory only
		specs.assign("modelDir", fmt.Sprintf("%s%s/", outDir, specs.modelKey()))

		if er := os.RemoveAll(fmt.Sprintf("%s%s", outDir, specs.graphsKey())); er != nil {
			return nil, nil, nil, er
		}
	}

	// create graph directory structure.
	var graphDir string
	if graphDir, e = makeSubDir(outDir, specs.graphsKey()); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("graphDir", graphDir)

	var dir string
	if dir, e = makeSubDir(graphDir, "validation"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("valDir", dir)

	if dir, e = makeSubDir(graphDir, "marginal"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("margDir", dir)

	if dir, e = makeSubDir(graphDir, "cost"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("costDir", dir)

	if dir, e = makeSubDir(graphDir, "strats"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("stratsDir", dir)

	if dir, e = makeSubDir(graphDir, "curves"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("curvesDir", dir)

	// create inputModel subdirectory
	if dir, e = makeSubDir(specs.getkeyVal("modelDir", true), "inputModels"); e != nil {
		return nil, nil, nil, e
	}
	specs.assign("inputDir", dir)

	// just doing assessment/bias adjustment ... append to existing log file, don't copy .gom file or input models
	if !specs.buildData() && !specs.buildModel() {
		// there is already a .gom file, so copy this to the date/time
		dttm := time.Now().Format("060102150405")
		toFile := fmt.Sprintf("%s%sdmodel.gom", outDir, dttm)
		if er := copyFile(specsFile, toFile); er != nil {
			return nil, nil, nil, er
		}

		// load up the needed cts and cat feature list in this case
		if er := specs.findFeatures(specs.getkeyVal("modelDir", true), true); er != nil {
			return nil, nil, nil, er
		}

		logFile, er := os.OpenFile(outDir+"model.log", os.O_APPEND|os.O_WRONLY, os.ModePerm)

		if er != nil {
			return nil, nil, nil, er
		}

		return specs, conn, logFile, nil
	}

	// copy over the spec file
	if er := copyFile(specsFile, outDir+"model.gom"); er != nil {
		return nil, nil, nil, er
	}

	// process any input models
	if er := specs.inputModels(); er != nil {
		return nil, nil, nil, er
	}

	// load up required features from inputModels (there will be no model yet in modelDir() but inputModels may be populated)
	if er := specs.findFeatures(specs.getkeyVal("modelDir", true), true); er != nil {
		return nil, nil, nil, er
	}

	// crerate log file
	logFile, e := os.Create(outDir + "model.log")

	return specs, conn, logFile, e
}

// buildQuery builds a query from a skeleton.  Any time the skeleton contains <key>-where key is a key in replacers-
// it is replaced with the value map[key].
func buildQuery(baseWith string, replacers map[string]string) string {
	qry := baseWith

	// go through this twice since some of the strings substituted in may also have "<xyz>" replacers
	for ind := 0; ind < 2; ind++ {
		for k, v := range replacers {
			krep := fmt.Sprintf("<%s>", k)
			// add whitespace around v
			qry = strings.ReplaceAll(qry, krep, fmt.Sprintf(" %s ", v))
		}
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
		FileName: specs.getkeyVal("costDir", true) + "modelSample.html",
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
		FileName: specs.getkeyVal("costDir", true) + "validationSample.html",
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

// logger makes a log entry.
func logger(log *os.File, text string, toConsole bool) {
	_, _ = fmt.Fprintln(log, text)
	if toConsole {
		fmt.Println(text)
	}
}

// inModel determines whether the feature is in the input statement from a sea.ModSpec
func inModel(input, feature string) bool {
	const minLen = 5 // "input" has 5 letters
	if len(input) < minLen {
		return false
	}

	if input[0:5] != "input" {
		return false
	}
	inputs := strings.Split(input, "+")
	for _, inp := range inputs {
		if strings.Contains(inp, feature) {
			return true
		}
	}

	return false
}
