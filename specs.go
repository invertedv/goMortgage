package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	sea "github.com/invertedv/seafan"
)

const (
	fannie = "fannie"

	plotWidth  = 1600.0
	plotHeight = 1200.0
	plotShow   = false
)

// specsMap holds the specs provided by the user
type specsMap map[string]string

// learnRate returns the starting and ending learning rate
func (sf specsMap) learnRate() (start, end float64, err error) {
	start, err = strconv.ParseFloat(strings.ReplaceAll(sf["learningRateStart"], " ", ""), bits64)
	if err != nil {
		return
	}

	end, err = strconv.ParseFloat(strings.ReplaceAll(sf["learningRateEnd"], " ", ""), bits64)

	return
}

// getQuery returns to query to pull data from ClickHouse
func (sf specsMap) getQuery(table string) string {
	flds := strings.Join(sf.queryFields(), ",")
	key := fmt.Sprintf("%sQuery", table)
	return fmt.Sprintf(sf[key], flds, sf["modelTable"]) + " " // add trailing blank
}

// requiredKeys returns the keys required in the specs file for each type of model we can build.
func (sf specsMap) requiredKeys() string {
	switch sf["model"] {
	case "mod", "dq", "death", "netpro":
		return `mtgDb, econDb, pass1Strat, pass1Sample, pass2Strat,
                    pass2Sample, log, mtgFields, econFields, modelTable,
                    sampleSize1, strats1, sampleSize2, strats2, where1, where2, layer1,
                    batchSize, epochs, earlyStopping, targetType,
                    learningRateStart, learningRateEnd, modelQuery, validateQuery, assessQuery`
	}

	return ""
}

// get returns the value in sp
func (sf specsMap) get(key string) string {
	if val, ok := sf[key]; ok {
		return val
	}

	return ""
}

// slices struct holds the details a feature to group by and the model output to use.
// The structure in the specs file is:
// <base>Name<shortName> : <name>
// <base>Target<shortName>: <targetStr>
// <base>Slice<shortName>: <feature>
//
// <base> can be: curves or assess,
//
// Example:
// assessNameState: Property State
// assessTargetState: state
// assessSliceState: 1, 2
//
// Will assess the model sliceing the assess data by state and using the sum of model output columns 1 and 2 as the metric.
type slices struct {
	name      string // display name (e.g. for plots)
	shortName string // name as used in the key in specs file
	feature   string // name of feature we're operating on
	targetStr string // target values, as a string
	target    []int  // target values as []int
}

// l2 returns L2 regularization parameter
func (sf specsMap) l2() (float64, error) {
	l2Str, ok := sf["l2Reg"]
	if !ok {
		return 0.0, nil
	}
	l2, err := strconv.ParseFloat(strings.ReplaceAll(l2Str, " ", ""), bits64)
	if err != nil {
		return 0.0, err
	}
	return l2, nil
}

// gDir returns and creates the directory to place the graphs for sl
func (sf specsMap) gDir(dirType string, sl *slices) (path string, err error) {
	baseDir := ""
	switch dirType {
	case "assess":
		baseDir = sf["valDir"]
	case "marginal":
		baseDir = sf["margDir"]
	default:
		return "", fmt.Errorf("(specsMap) gDir: invalid option %s", dirType)
	}
	path, err = makeSubDir(baseDir, sl.shortName)

	return path, err
}

// slicer returns an array of slicers specified in specs for the base category (assess or curves)
func (sf specsMap) slicer(base string) []slices {
	vals := make([]slices, 0)
	keyFind := fmt.Sprintf("%sName", base)

	for k, v := range sf {
		var item slices
		if !strings.Contains(k, keyFind) {
			continue
		}
		if len(base) == len(k) {
			item = slices{name: k, feature: "", targetStr: "", target: nil}
			return append(vals, item)
		}
		shortName := k[len(keyFind):]
		targetStr := sf.get(fmt.Sprintf("%sTarget%s", base, shortName))
		if targetStr == "" {
			item = slices{name: k, feature: "", targetStr: "", target: nil}
			return append(vals, item)
		}
		spl := toSlice(targetStr, ",")
		trgs := make([]int, 0)
		for _, trg := range spl {
			i, e := strconv.ParseInt(strings.ReplaceAll(trg, " ", ""), base10, bits32)
			if e != nil {
				item = slices{name: k, feature: "", targetStr: "", target: nil}
				return append(vals, item)
			}
			trgs = append(trgs, int(i))
		}
		item = slices{
			name:      v,
			feature:   sf.get(fmt.Sprintf("%sSlicer%s", base, shortName)),
			shortName: shortName,
			targetStr: targetStr,
			target:    trgs,
		}
		vals = append(vals, item)
	}

	return vals
}

// layers returns the model layers specified by the user.  The layers are specified in the specs file as
// layer<num>: <seafan layer>.
// The layers are sequential, ordered by <num> starting with 1.
func (sf specsMap) layers() (model []string) {
	model = make([]string, 0)

	lyr := 1
	for lyrStr, ok := sf[fmt.Sprintf("layer%d", lyr)]; ok; lyrStr, ok = sf[fmt.Sprintf("layer%d", lyr)] {
		model = append(model, lyrStr)
		lyr++
	}
	return
}

// mtgFields returns the Pass3 goMortgage fields from the data source specified in the specs field <mtgFields>
func (sf specsMap) mtgFields() string {
	switch sf["mtgFields"] {
	case fannie:
		return fannieMtgFields
	default:
		return ""
	}
}

// goodLoan returns the Pass1 WHERE clause for the data source the restricts selections to loans that pass QA.
func (sf specsMap) goodLoan() string {
	switch sf["mtgFields"] {
	case fannie:
		return fannieGoodLoan
	default:
		return ""
	}
}

// pass1Fields returns the Pass1 fields for the data source.
func (sf specsMap) pass1Fields() string {
	switch sf["mtgFields"] {
	case fannie:
		return fanniePass1
	default:
		return ""
	}
}

// pass2Fields returns the Pass2 fields for the data source.
func (sf specsMap) pass2Fields() string {
	switch sf["mtgFields"] {
	case fannie:
		return fanniePass2Fields
	default:
		return ""
	}
}

// pass3Fields returns the Pass3 fields for the data source.
func (sf specsMap) pass3Fields() string {
	switch sf["mtgFields"] {
	case fannie:
		return fanniePass3Calcs
	default:
		return ""
	}
}

// econJoin is called for pass3 which joins the sampled goMortgage data to economic data.
// The returns are
//
//   - table : WITH statement that generates the econ data.
//   - fields: field list of econ fields to return from query
//
// Economic data is pulled at 3 time periods per loan:
//
//   - first pay date
//   - as-of date
//   - target date
//
// To accomodate this, the field list has the form:
//
//	<corr><base field> AS <pre><base field>
//
// where <base field> is the root field (e.g. HPI), <corr> is the corrlation for the table (since it's joined 3 times)
// and <pre> is a prefix (org, ao, trg)
func (sf specsMap) econJoin() (table, fields string) {
	var fieldList string
	switch sf["econFields"] {
	case "zip3":
		table, fieldList = econTable3, econFields3 // these are embedded files
	default:
		return "", ""
	}

	fields3 := make([]string, 0)
	fields3 = append(fields3, strings.ReplaceAll(strings.ReplaceAll(fieldList, "<corr>", "b."), "<pre>", "trg"),
		strings.ReplaceAll(strings.ReplaceAll(fieldList, "<corr>", "c."), "<pre>", "ao"),
		strings.ReplaceAll(strings.ReplaceAll(fieldList, "<corr>", "d."), "<pre>", "org"),
		strings.ReplaceAll(strings.ReplaceAll(fieldList, "<corr>", "x2020."), "<pre>", "y20"))
	return table, strings.Join(fields3, ",")
}

// batchSize returns batch size
func (sf specsMap) batchSize() (int, error) {
	bSize, e := strconv.ParseInt(strings.ReplaceAll(sf["batchSize"], " ", ""), base10, bits32)
	return int(bSize), e
}

// epochs returns # of epochs
func (sf specsMap) epochs() (int, error) {
	epochs, e := strconv.ParseInt(strings.ReplaceAll(sf["epochs"], " ", ""), base10, bits32)
	return int(epochs), e
}

// earlyStopping returns # of epochs with no improvement to trigger early stopping
func (sf specsMap) earlyStopping() (int, error) {
	eStop, e := strconv.ParseInt(strings.ReplaceAll(sf["earlyStopping"], " ", ""), base10, bits32)
	return int(eStop), e
}

// plotShow returns true if the user wants to show all the plots in a browser, too.
func (sf specsMap) plotShow() bool {
	show, ok := sf["plotShow"]
	if !ok {
		return plotShow
	}
	return show == yes
}

// plotWidth returns plot width (in pixels)
func (sf specsMap) plotWidth() float64 {
	pw, ok := sf["PlotWidth"]
	if !ok {
		return plotWidth
	}
	pwFl, e := strconv.ParseFloat(strings.ReplaceAll(pw, " ", ""), bits64)
	if e != nil {
		return plotWidth
	}
	return pwFl
}

// plotHeight returns plot height (in pixels)
func (sf specsMap) plotHeight() float64 {
	pw, ok := sf["PlotHeight"]
	if !ok {
		return plotHeight
	}
	pwFl, e := strconv.ParseFloat(strings.ReplaceAll(pw, " ", ""), bits64)
	if e != nil {
		return plotHeight
	}
	return pwFl
}

// saveTable returns details to save the assess data back to ClickHouse.
// tableName: fully qualified table name
// fields: extra fields from the model to include
// targets: target columns corresponding to the extra fields.
//
// In the specs file, this looks like:
//
//	saveTable: mtg.outDqT12
//	saveTableTargets: d120:4,5,6,7,8,9,10,11,12; d30:1
//
// The assess data is saved to mtg.outDqT12.  It will have two extra fields: d120 and d30.  D120 is the sum
// of columns 4-12 of the model output, and d30 is column 1.
func (sf specsMap) saveTable() (tableName string, fields []string, targets [][]int, err error) {
	table, ok := sf["saveTable"]
	if !ok {
		return "", nil, nil, nil
	}

	// save table w/o adding fitted
	fTargs, ok := sf["saveTableTargets"]
	if !ok {
		return table, nil, nil, nil
	}

	fields = make([]string, 0)
	targets = make([][]int, 0)

	// fTargs = strings.ReplaceAll(fTargs, " ", "")
	fts := strings.Split(fTargs, ";")
	for _, ft := range fts {
		fldTarg := strings.Split(ft, ":")
		if len(fldTarg) != 2 {
			return "", nil, nil, fmt.Errorf("cannot parse saveTableTargets: %s", fTargs)
		}

		field := fldTarg[0]
		targsStr := strings.Split(fldTarg[1], ",")
		targs := make([]int, 0)

		for _, targStr := range targsStr {
			targ, e := strconv.ParseInt(strings.ReplaceAll(targStr, " ", ""), base10, bits32)
			if e != nil {
				return "", nil, nil, fmt.Errorf("cannot ParseInt targets %s", targsStr)
			}
			targs = append(targs, int(targ))
		}

		targets = append(targets, targs)
		fields = append(fields, field)
	}

	return table, fields, targets, nil
}

// inputModels copies input models to the subdirectory model/inputModels in the directory for this model.
//
// The format in the specs file is:
//
//	inputModel: mod
//	modLocation: /home/user/goMortgage/mod/
//	modTargets: pMod:1
//
// The output of the model, column 1, will be called pMod and be available as a feature.
func (sf specsMap) inputModels() error {
	for k, v := range sf {
		if !strings.Contains(k, "inputModel") {
			continue
		}

		modelName := v

		loc, ok := sf["location"+modelName]
		if !ok {
			return fmt.Errorf("(specsMap) inputModels: no location for model %s", modelName)
		}

		targets, ok := sf["targets"+modelName]
		if !ok {
			return fmt.Errorf("(specsMap) inputModels: no target for model %s", modelName)
		}

		path := slash(sf["inputDir"] + modelName)
		if e := copyFiles(slash(loc)+"model", path); e != nil {
			return e
		}

		// create specs file. If there are nested models, the .spec file will already exist
		handle, e := os.Create(path + "targets.spec")
		if e != nil {
			return e
		}
		targetSl := strings.Split(targets, ";")
		for _, trg := range targetSl {
			if _, e = handle.WriteString(fmt.Sprintf("%s\n", strings.ReplaceAll(trg, " ", ""))); e != nil {
				return e
			}
		}
		if e := handle.Close(); e != nil {
			return e
		}
	}
	return nil
}

// check checks that required keys are available in sf
func (sf specsMap) check(required string) error {
	for _, item := range sf.slicer("curves") {
		if item.feature == "" || len(item.target) == 0 {
			return fmt.Errorf("curves for %s missing target or slicer", item.name)
		}
	}

	for _, item := range sf.slicer("assess") {
		if item.feature == "" || len(item.target) == 0 {
			return fmt.Errorf("curves for %s missing target or slicer", item.name)
		}
	}

	reqd := toSlice(required, ",")
	for _, req := range reqd {
		_, ok := sf[req]
		if !ok {
			return fmt.Errorf("required key %s not in specs file", req)
		}
	}
	switch sf["model"] {
	case "mod", "dq", "death", "netpro":
		return nil
	default:
		return fmt.Errorf("invalid model choice: %s", sf["model"])
	}
}

// ctsFeatures returns a slice of continuous features in the model
func (sf specsMap) ctsFeatures() []string {
	if _, ok := sf["cts"]; !ok {
		return nil
	}
	return toSlice(sf["cts"], ",")
}

// ohFeatures returns a slice of the one-hot features in the model
func (sf specsMap) ohFeatures() []string {
	if _, ok := sf["cat"]; !ok {
		return nil
	}
	return toSlice(sf["cat"], ",")
}

// ohFields slice is all fields that need one-hot encoding (cat features, emb features and target, if categorical)
func (sf specsMap) ohFields() []string {
	flds := sf.ohFeatures()
	eFld, _ := sf.embFeatures(false)
	flds = append(flds, eFld...)
	if sf.targetType() == sea.FRCat {
		flds = append(flds, sf.target())
	}
	return flds
}

// addlCats slice is user-specified additional fields that should be sea.FRCat
func (sf specsMap) addlCats() []string {
	if _, ok := sf["addlCats"]; !ok {
		return nil
	}
	return toSlice(sf["addlCats"], ",")
}

// allCat returns the one-hot features plus additional one-hot features specified by the addlCats key in specs.
func (sf specsMap) allCat() []string {
	all := append(sf.ohFeatures(), sf.addlCats()...)
	emb, _ := sf.embFeatures(false)
	all = append(all, emb...)
	if sf.targetType() == sea.FRCat {
		all = append(all, sf.target())
	}

	return all
}

// allFeatures returns all the features in the model (continuous, one-hot, embedded)
func (sf specsMap) allFeatures() []string {
	embF, _ := sf.embFeatures(false)
	return append(append(sf.ctsFeatures(), sf.ohFeatures()...), embF...)
}

// targetType returns the type of the target feature.
func (sf specsMap) targetType() sea.FRole {
	if sf["targetType"] == "cat" {
		return sea.FRCat
	}
	return sea.FRCts
}

// assessFields returns a slice of all the fields to use in the model assessment.  This consists of all the features
// in the model plus any features specified in "assessAddl" key.
func (sf specsMap) assessFields() []string {
	flds := sf.allFeatures()
	addl, ok := sf["assessAddl"]
	if ok {
		flds = append(flds, strings.Split(strings.ReplaceAll(addl, " ", ""), ",")...)
	}

	return flds
}

// costFunc returns the cost function for the model.
func (sf specsMap) costFunc() sea.CostFunc {
	switch sf.targetType() {
	case sea.FRCat:
		return sea.CrossEntropy
	case sea.FRCts:
		return sea.RMS
	}
	return nil
}

// embFeatures returns a slice of the embedded features.
// If complete is true, then the list is suitable for seafan (E(<feature>Oh:<embeddingColumns>).
// If complete is false, then the list is just the "from" features.
func (sf specsMap) embFeatures(complete bool) (parsed []string, err error) {
	if _, ok := sf["emb"]; !ok {
		return nil, nil
	}

	for _, emb := range toSlice(sf["emb"], ",") {
		fact := strings.Split(emb, ":")
		if len(fact) != 2 {
			return nil, fmt.Errorf("invalid embedding format: %s", emb)
		}
		switch complete {
		case true:
			_, e := strconv.ParseInt(strings.ReplaceAll(fact[1], " ", ""), base10, bits32)
			if e != nil {
				return nil, fmt.Errorf("invalid dimension in embedding format: %s", emb)
			}
			parsed = append(parsed, fmt.Sprintf("E(%s,%s)", fact[0]+"Oh", fact[1]))
		case false:
			parsed = append(parsed, fact[0])
		}
	}

	return
}

// target returns the name of the target feature.
func (sf specsMap) target() string {
	return sf["target"]
}

// existing returns the directory that holds the existing models
func (sf specsMap) existing() string {
	return sf["modelDir"] + "inputModels"
}

// modelRoot returns the location+root name of the model we're fitting
func (sf specsMap) modelRoot() string {
	return sf["modelDir"] + "model"
}

// costDir returns the directory for the cost graphs
func (sf specsMap) costDir() string {
	return sf["costDir"]
}

// allFields returns a slice of all the fields required by the run
func (sf specsMap) allFields() []string {
	aFld := sf.allFeatures()
	aFld = append(aFld, sf.assessFields()...)
	aFld = append(aFld, sf.addlCats()...)

	//what was this?
	//	if sf.targetType() != sea.FRCat {
	//		aFld = append(aFld, sf.target())
	//	}

	aFld = append(aFld, sf.target())
	for _, sl := range sf.slicer("curves") {
		aFld = append(aFld, sl.feature)
	}

	for _, sl := range sf.slicer("assess") {
		aFld = append(aFld, sl.feature)
	}

	if _, ok := sf["addlKeep"]; ok {
		aFld = append(aFld, toSlice(sf["addlKeep"], ",")...)
	}
	// de-dupe
	sort.Strings(aFld)
	outFld := []string{aFld[0]}
	for ind := 1; ind < len(aFld); ind++ {
		if aFld[ind] != aFld[ind-1] {
			outFld = append(outFld, aFld[ind])
		}
	}

	return outFld
}

// calcFields returns the fields derived from input models (key inputModel)
func (sf specsMap) calcFields() []string {
	cFlds := make([]string, 0)
	for k, v := range sf {
		if !strings.Contains(k, "inputModel") {
			continue
		}
		targs := sf[fmt.Sprintf("targets%s", v)]
		for _, trg := range strings.Split(targs, ";") {
			each := strings.Split(trg, ":")
			cFlds = append(cFlds, strings.ReplaceAll(each[0], " ", ""))
		}
	}
	return cFlds
}

// queryFields returns the fields to be pulled from the ClickHouse table.
func (sf specsMap) queryFields() []string {
	allFlds := sf.allFields()
	// fields from input models
	calcFlds := sf.calcFields()
	if len(calcFlds) == 0 {
		return allFlds
	}
	qFlds := make([]string, 0)
	for _, fld := range allFlds {
		isCalc := false
		for _, cfld := range calcFlds {
			if fld == cfld {
				isCalc = true
				break
			}
		}
		if !isCalc {
			qFlds = append(qFlds, fld)
		}
	}

	return qFlds
}

func (sf specsMap) title() string {
	if title, ok := sf["title"]; ok {
		return title
	}
	return ""
}

// readSpecsMap reads the specfile and creates the specMap.
func readSpecsMap(specFile, required string) (specsMap, error) {
	handle, e := os.Open(specFile)
	if e != nil {
		return nil, e
	}
	defer func() { _ = handle.Close() }()

	rdr := bufio.NewReader(handle)

	sMap := make(specsMap)
	for {
		var line string
		line, e = rdr.ReadString('\n')
		line = strings.TrimLeft(strings.TrimRight(line, "\n"), " ")
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		// skip blank lines
		if strings.TrimRight(strings.ReplaceAll(line, " ", ""), "\n") == "" {
			continue
		}
		// skip comments
		if line[0:2] == "//" {
			continue
		}

		if ind := strings.Index(line, "//"); ind >= 0 {
			line = line[0:ind]
			line = strings.TrimRight(line, " ")
		}

		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("bad key val: %s in specs file %s", line, specFile)
		}

		key := strings.ReplaceAll(kv[0], " ", "")
		ind := 0
		for _, ok := sMap[key]; ok; _, ok = sMap[key] {
			ind++
			key = fmt.Sprintf("%s%d", key, ind)
		}
		sMap[key] = strings.TrimLeft(kv[1], " ")
	}
	// all models need these
	if e := sMap.check(required); e != nil {
		return nil, e
	}

	sMap["outDir"] = slash(sMap["outDir"])

	return sMap, nil
}
