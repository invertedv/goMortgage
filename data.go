package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/invertedv/chutils"
	s "github.com/invertedv/chutils/sql"
	"github.com/invertedv/sampler"
	sea "github.com/invertedv/seafan"
)

// existing adds the output of an existing model to basePipe. This expects 4 files in modelRoot:
//   - The NNModel files modelP.nn and modelS.nn
//   - FTypes file that defines the features in the model.  The data in basePipe is re-normalized and re-mapped using
//     these values.
//   - target.specs.  This file specifies the name(s) of the fields to create in basePipe. It has the format:
//     <field name>:<target columns to coalesce separated by commas>.
func existing(modelRoot string, basePipe sea.Pipeline) error {
	modelRoot = slash(modelRoot)

	// see if there are any directories in here -- these would be input models to this model
	dirList, e := os.ReadDir(modelRoot)
	if e != nil {
		return e
	}

	hasFiles := false // this directory may be a directory of directories (submodels)
	for _, entry := range dirList {
		// load up the submodel
		if entry.IsDir() {
			if er := existing(modelRoot+entry.Name(), basePipe); er != nil {
				return er
			}
		} else {
			hasFiles = true
		}
	}
	if !hasFiles {
		return nil
	}

	fts, e := sea.LoadFTypes(modelRoot + "fieldDefs.jsn")
	if e != nil {
		return e
	}

	handle, e := os.Open(modelRoot + "targets.spec")
	if e != nil {
		return e
	}

	rdr := bufio.NewReader(handle)

	for line, err := rdr.ReadString('\n'); err == nil; line, err = rdr.ReadString('\n') {
		spl := toSlice(line, "{")
		if len(spl) != 2 {
			return fmt.Errorf("existing model %s error in target %s", modelRoot, line)
		}

		lvls := strings.Split(strings.ReplaceAll(spl[1], "}", ""), ",")
		fieldName := spl[0]
		targets := make([]int, 0)

		for _, lvl := range lvls {
			ilvl, e1 := strconv.ParseInt(lvl, base10, bits32)
			if e1 != nil {
				return fmt.Errorf("existing error parsing targets %s for model %s", line, modelRoot)
			}
			targets = append(targets, int(ilvl))
		}
		//TODO: decide logodds intelligently
		modSpec, e := sea.LoadModSpec(modelRoot + "modelS.nn")
		if e != nil {
			return e
		}
		var obsFt *sea.FType = nil

		if trg := modSpec.TargetName(); trg != "" {
			obsFt = fts.Get(trg)
		}

		if e := sea.AddFitted(basePipe, modelRoot+"model", targets, fieldName, fts, true, obsFt); e != nil {
			return e
		}
	}

	return nil
}

// allExisting runs through all the existing models in the inputModels directory.
func allExisting(rootDir string, basePipe sea.Pipeline) error {
	rootDir = slash(rootDir)
	entries, e := os.ReadDir(rootDir)
	if e != nil {
		return e
	}

	for _, dir := range entries {
		if !dir.IsDir() {
			continue
		}
		if e := existing(rootDir+dir.Name(), basePipe); e != nil {
			return e
		}
	}

	return nil
}

// newPipe creates a new data pipeline.
//
//   - qry: query to run against ClickHouse.
//   - name: name of pipeline.
//   - specs: user specs.
//   - bSize: batch size.  0 means batch size is equal to the # of rows in the data.
//   - conn: connection to ClickHouse.
func newPipe(qry, name string, specs specsMap, bSize int, fts sea.FTypes,
	conn *chutils.Connect) (sea.Pipeline, error) {
	rdr := s.NewReader(qry, conn)
	if e := rdr.Init("", chutils.MergeTree); e != nil {
		return nil, e
	}

	pipe := sea.NewChData(name)

	if fts != nil {
		sea.WithFtypes(fts)(pipe)
	}

	sea.WithReader(rdr)(pipe)
	sea.WithCats(specs.allCat()...)(pipe)
	sea.WithNormalized(specs.ctsFeatures()...)(pipe)
	if specs.targetType() == sea.FRCts {
		sea.WithNormalized(specs.target())(pipe)
	}

	for _, cat := range specs.ohFields() {
		sea.WithOneHot(cat+"Oh", cat)(pipe)
	}

	if e := pipe.Init(); e != nil {
		return nil, e
	}
	sea.WithBatchSize(bSize)(pipe)

	if specs.existing() == "" {
		return pipe, nil
	}

	if e := allExisting(specs.existing(), pipe); e != nil {
		return nil, e
	}

	// check we have all the fields we need
	for _, fld := range specs.allFields() {
		ft := pipe.GetFType(fld)
		if ft == nil {
			return nil, fmt.Errorf("field %s not in pipeline", fld)
		}
	}

	return pipe, nil
}

// pass1 samples the raw loan-level data to determine select the as-of dates.
// The pass1 query requires the following field replacements:
//   - mtgDb: loan-level goMortgage data table.
//   - fields: fields required for stratification and those that must be pulled for the as-of date. These latter are
//     generally (a) calculation at the as-of date or (b) monthly values pulled for the as-of date
//   - goodLoan: QA restrictions
//   - where: other restrictions to loans to be considered.
//
// specs fields used directly:
//   - where1: optional additional restrictions on the selection
//   - strats1:  fields to stratify on.
//   - pass1Sample: output table of loan-level sample.
//   - pass1Strat: output table of counts by strat
//   - stratsDir: location to place graphs of strats
//
// specs methods used:
//   - goodLoan
//   - pass1Fields
func pass1(specs specsMap, conn *chutils.Connect, log *os.File) error {
	specs["goodLoan"], specs["where"] = specs.goodLoan(), ""

	if where1, ok := specs["where1"]; ok {
		specs["where"] = where1
	}

	specs["fields"] = specs.pass1Fields()
	qry := buildQuery(withPass1, specs)

	sampleSize, e := strconv.ParseInt(specs["sampleSize1"], base10, bits32)
	if e != nil {
		return e
	}

	gen := sampler.NewGenerator(qry, specs["pass1Sample"], specs["pass1Strat"], int(sampleSize), true, conn)

	strats := toSlice(specs["strats1"], ",")

	if e := gen.CalcRates(strats...); e != nil {
		return e
	}

	if e := gen.MakeTable(); e != nil {
		return e
	}

	if e := gen.SampleStrats().Plot(specs["stratsDir"]+"pass1.html", specs.plotShow()); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("Pass 1 Strats:\n%v", gen), false)
	return nil
}

// pass2 starts with the sample from pass1 and selects the target date.
// The pass2 query requires the following field replacements:
//   - mtgDb: loan-level goMortgage data table.
//   - pass1Sample: table of sampled loans produced by pass1
//   - fields. The fields to keep.  These will be sourced from pass1 and the loan-level data.
//   - where.  This is optional.  Where clause to further restrict selection.
//
// specs fields used directly:
//   - strats2:  fields to stratify on.
//   - pass2Sample: output table of loan-level sample.
//   - pass2Strat: output table of counts by strat
//   - stratsDir: location to place graphs of strats
//
// specs methods used:
//   - pass2Fields
//   - mtgFields
//   - plotShow
func pass2(specs specsMap, conn *chutils.Connect, log *os.File) error {
	specs["where"] = ""
	if where2, ok := specs["where2"]; ok {
		specs["where"] = where2
	}
	specs["fields"] = fmt.Sprintf("%s, %s", specs.mtgFields(), specs.pass2Fields())
	qry := buildQuery(withPass2, specs)

	sampleSize, e := strconv.ParseInt(specs["sampleSize2"], base10, bits32)
	if e != nil {
		return e
	}

	gen := sampler.NewGenerator(qry, specs["pass2Sample"], specs["pass2Strat"], int(sampleSize), true, conn)

	strats := toSlice(specs["strats2"], ",")

	if e := gen.CalcRates(strats...); e != nil {
		return e
	}

	if e := gen.MakeTable(); e != nil {
		return e
	}

	if e := gen.SampleStrats().Plot(specs["stratsDir"]+"pass2.html", specs.plotShow()); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("Pass 2 Strats:\n%v", gen), false)

	return nil
}

// pass3 joins the output of pass2 with economic data.
// pass3 requires the following field replacements:
//   - with: With statement that defines to economic table
//   - fields: fields to keep
//   - pass2Sample: table of sampled loans output by pass2
//   - econFields: field name to join the econ table to the loan-level table.
//
// specs fields used directly:
//   - modelTable: name of output table
//
// specs methods used:
//   - econJoin
//   - pass3Fields
func pass3(specs specsMap, conn *chutils.Connect) error {
	econTable, econFields := specs.econJoin()
	specs["with"] = econTable
	specs["fields"] = econFields + "," + specs.pass3Fields()
	qry := buildQuery(withPass3, specs)
	rdr := s.NewReader(qry, conn)
	rdr.Name = specs.outTable()

	if e := rdr.Init(specs.tableKey(), chutils.MergeTree); e != nil {
		return e
	}

	if e := rdr.TableSpec().Create(conn, specs.outTable()); e != nil {
		return e
	}

	if e := rdr.Insert(); e != nil {
		return e
	}

	return nil
}

// data builds the modeling data
func data(specs specsMap, conn *chutils.Connect, log *os.File) error {
	start := time.Now()
	logger(log, fmt.Sprintf("starting data build @ %s", start.Format(time.UnixDate)), true)

	// pass 1
	if e := pass1(specs, conn, log); e != nil {
		return e
	}

	// pass 2
	if e := pass2(specs, conn, log); e != nil {
		return e
	}

	// pass 3
	if e := pass3(specs, conn); e != nil {
		return e
	}

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("data build run time: %0.1f minutes", elapsed), true)

	return nil
}

// addDefault sets the default value for FRCat fields to their mode. The default value can be needed when using the
// model to predict on new data.  The returned FTypes are restricted to those fields in keepFeatures
func addDefault(pipe sea.Pipeline, keepFeatures []string) (sea.FTypes, error) {
	fts := pipe.GetFTypes()

	drops := make([]string, 0)
	for _, ft := range fts {
		ok := false
		for _, mf := range keepFeatures {
			if mf == ft.Name {
				ok = true
				break
			}
		}
		if !ok {
			drops = append(drops, ft.Name)
		}
	}

	if len(drops) > 0 {
		fts = fts.DropFields(drops...)
	}

	for _, ft := range fts {
		if ft.Role != sea.FRCat {
			continue
		}

		x, e := pipe.GData().GetRaw(ft.Name)
		if e != nil {
			return nil, e
		}

		lvls := sea.ByCounts(x, nil)
		keys, _ := lvls.Sort(false, false)
		ft.FP.Default = keys[0]
	}

	return fts, nil
}

// export saves a pipeline back to ClickHouse.
//
// obsFT is the FType from the modeling pipeline
func export(pipe sea.Pipeline, specs specsMap, obsFt *sea.FType, conn *chutils.Connect) error {
	table, fields, targets, err := specs.saveTable()
	if err != nil {
		return err
	}

	if table == "" {
		return nil
	}
	// TODO: LOOK at nil below
	// are there model-output fields to add?
	if len(fields) > 0 {
		for ind, field := range fields {
			if e := sea.AddFitted(pipe, specs.modelRoot(), targets[ind], field, nil, false, obsFt); e != nil {
				return e
			}
		}
	}

	// make writer
	wtr := s.NewWriter(table, conn)
	defer func() { _ = wtr.Close() }()

	gd := pipe.GData()
	tb := gd.TableSpec()

	if e := tb.Create(conn, table); e != nil {
		return e
	}

	if e := chutils.Export(pipe.GData(), wtr, 0, false); e != nil {
		return e
	}

	return nil
}
