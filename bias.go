package main

import (
	"fmt"
	"io"
	"math"
	"os"

	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
	G "gorgonia.org/gorgonia"
	"gorgonia.org/tensor"

	"github.com/invertedv/chutils"
	s "github.com/invertedv/chutils/sql"
	sea "github.com/invertedv/seafan"
)

type objFn func(x []float64) float64

func BiasCorrect(pipe sea.Pipeline, specs specsMap, conn *chutils.Connect, log *os.File) error {
	var sseFn objFn
	var bAdj []float64
	var e error
	var optimal *optimize.Result

	nnModel, err := sea.PredictNN(specs.modelDir()+"model", pipe, false)
	if err != nil {
		return err
	}

	modSpec := nnModel.ModSpec()

	var outLayer *sea.FCLayer
	var outLayLoc int

	for outLayLoc = len(modSpec); outLayLoc >= 0; outLayLoc-- {
		outLayer = modSpec.FC(outLayLoc)
		if outLayer != nil {
			break
		}
	}

	if outLayer.Act != sea.SoftMax {
		return fmt.Errorf("bias correction: model output is not softmax")
	}

	biasQ := specs.biasQuery()
	if biasQ == "" {
		return nil
	}

	if sseFn, bAdj, e = buildObj(pipe, nnModel, biasQ, specs, conn); e != nil {
		return e
	}

	logger(log, "conducting bias correction", true)

	grad := func(grad, x []float64) {
		fd.Gradient(grad, sseFn, x, nil)
	}
	hess := func(h *mat.SymDense, x []float64) {
		fd.Hessian(h, sseFn, x, nil)
	}
	problem := optimize.Problem{Func: sseFn, Grad: grad, Hess: hess}

	if optimal, e = optimize.Minimize(problem, bAdj, nil, &optimize.Newton{}); e != nil {
		return e
	}

	logger(log, fmt.Sprintln("bias corrections factors", optimal.X), true)
	logger(log, fmt.Sprintf("fit SSE: %0.5f", sseFn(optimal.X)), true)

	nodeName := fmt.Sprintf("lBias%d", outLayLoc)
	node := nnModel.G().ByName(nodeName)
	// output bias values
	vals := node.Nodes()[0].Value().Data().([]float64)

	if len(vals) != len(optimal.X) {
		return fmt.Errorf("Bias and Adjustment have differing lengths: %d and %d", len(vals), len(optimal.X))
	}

	for ind := 0; ind < len(vals); ind++ {
		vals[ind] += optimal.X[ind]
	}

	t := tensor.New(tensor.WithBacking(vals), tensor.WithShape(1, len(vals)))
	if e = G.Let(node.Nodes()[0], t); e != nil {
		return e
	}

	var loc string

	if loc, e = makeSubDir(specs["outDir"], specs.biasDir()); e != nil {
		return e
	}

	if e = nnModel.Save(loc + "model"); e != nil {
		return e
	}

	if e = copyFiles(specs.modelDir(), loc); e != nil {
		return e
	}

	specs["modelDir"] = loc

	return nil
}

func buildObj(pipe sea.Pipeline, nnModel *sea.NNModel, biasQuery string, specs specsMap, conn *chutils.Connect) (objFn, []float64, error) {

	// get fit probabilities
	probs := nnModel.FitSlice()

	nCol := nnModel.Cols()
	if nCol == 1 {
		return nil, nil, fmt.Errorf("bias only for categorical models")
	}

	nRow := pipe.Rows()
	logOdds := make([]float64, nRow*(nCol-1))

	// logodds is log(p[c]/p[nCol]) where c runs through first nCol-1 columns
	avgLogs := make([]float64, nCol-1)
	for row := 0; row < nRow; row++ {
		for col := 0; col < nCol-1; col++ {
			pDen := probs[row*nCol+nCol-1]
			pNum := probs[row*nCol+col]

			if pDen <= 0.0 || pDen >= 1.0 || pNum <= 0.0 || pNum >= 1.0 {
				return nil, nil, fmt.Errorf("encountered 0 or 1 probability")
			}

			lo := math.Log(pNum / pDen)
			logOdds[row*(nCol-1)+col] = lo
			avgLogs[col] += lo
		}
	}

	// get target average for each level of target
	biasQ := s.NewReader(biasQuery, conn)
	if e := biasQ.Init("", chutils.MergeTree); e != nil {
		return nil, nil, e
	}

	targets := make([]float64, 0)
	vals := make([]any, 0)

	for {
		rowVal, _, e := biasQ.Read(1, false)
		if e != nil && e != io.EOF {
			return nil, nil, e
		}

		if e != nil {
			break
		}

		if len(rowVal[0]) != 2 {
			return nil, nil, fmt.Errorf("bias query can return only two fields, got %d", len(rowVal))
		}

		var val1 any = rowVal[0][1]
		flt, ok := val1.(*float64)
		if !ok {
			return nil, nil, fmt.Errorf("bias query value is not float64 %v", rowVal[0][1])
		}

		vals = append(vals, rowVal[0][0])
		targets = append(targets, *flt)
	}

	if len(targets) != nCol {
		return nil, nil, fmt.Errorf("bias query returned %d rows, expected %d rows", len(targets), nCol)
	}

	targetsOrdered := make([]float64, nCol)
	ft := pipe.GetFType(specs.target())

	if ft == nil {
		return nil, nil, fmt.Errorf("target is missing from pipeline, bias corrections")
	}

	if ft.FP.Lvl == nil {
		return nil, nil, fmt.Errorf("target in bias pipeline isn't categorical")
	}

	for ind := 0; ind < nCol; ind++ {
		indLoc, ok := ft.FP.Lvl[vals[ind]]
		if !ok {
			return nil, nil, fmt.Errorf("value not in target levels %v", vals[ind])
		}
		targetsOrdered[indLoc] = targets[ind]
	}

	// build objective function for optimizer...sse of average phat to bias query target
	biasSse := func(biasAdj []float64) float64 {
		p := make([]float64, nCol)
		avgP := make([]float64, nCol)

		for row := 0; row < nRow; row++ {
			tot := 1.0

			for col := 0; col < nCol-1; col++ {
				p[col] = math.Exp(logOdds[row*(nCol-1)+col] + biasAdj[col])
				tot += p[col]
			}

			p[nCol-1] = 1.0

			for col := 0; col < nCol; col++ {
				p[col] /= tot
				avgP[col] += p[col]
			}
		}

		sse := 0.0
		nFlt := float64(nRow)
		wts := []float64{1.0, 1.0, 1.0}
		for col := 0; col < nCol; col++ {
			avg := avgP[col] / nFlt
			errv := (avg - targetsOrdered[col])
			sse += errv * errv * wts[col]
			// 9.118235469254579 5.518683539248588]
			// sse -= targetsOrdered[col] * math.Log(10000000.0*avg)
		}
		return sse
	}

	// starting values
	var bAdj = make([]float64, nnModel.Cols()-1)
	for ind := 0; ind < len(bAdj); ind++ {
		targ := math.Log(targetsOrdered[ind] / targetsOrdered[nCol-1])
		bAdj[ind] = targ - avgLogs[ind]/float64(nRow)
	}

	return biasSse, bAdj, nil
}
