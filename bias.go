package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
	G "gorgonia.org/gorgonia"
	"gorgonia.org/tensor"

	"github.com/invertedv/chutils"
	sea "github.com/invertedv/seafan"
)

// objective function for bias correction
type objFn func(x []float64) float64

// getOutLayer retrives the output layer from modSpec and returns the layer and its position in modSpec.
// The nnModel parameters are indexed by the layer position.
func getOutLayer(modSpec sea.ModSpec) (outLayer *sea.FCLayer, outLayLoc int) {
	for outLayLoc = len(modSpec); outLayLoc >= 0; outLayLoc-- {
		outLayer = modSpec.FC(outLayLoc)
		if outLayer != nil {
			break
		}
	}

	if outLayer == nil {
		return nil, 0
	}

	if outLayer.Act != sea.SoftMax {
		return nil, 0
	}

	return outLayer, outLayLoc
}

// biasCorrect corrects the bias in a NNModel that is caused by stratifying on the response.
// The user supplies a query that produces the desired rates
func biasCorrect(specs specsMap, conn *chutils.Connect, log *os.File) error {
	var (
		sseFn     objFn
		bAdj      []float64
		e         error
		optimal   *optimize.Result
		fts       sea.FTypes
		modelPipe sea.Pipeline
	)

	start := time.Now()
	logger(log, fmt.Sprintf("starting bias correction @ %s", start.Format(time.UnixDate)), true)

	if fts, e = sea.LoadFTypes(specs.modelDir() + "fieldDefs.jsn"); e != nil {
		return e
	}

	if modelPipe, e = newPipe(specs.getQuery("model"), "model data", specs, 0, fts, conn); e != nil {
		return e
	}

	nnModel, err := sea.PredictNN(specs.modelDir()+"model", modelPipe, false)
	if err != nil {
		return err
	}

	var (
		outLayer  *sea.FCLayer
		outLayLoc int
	)

	if outLayer, outLayLoc = getOutLayer(nnModel.ModSpec()); outLayer == nil {
		return fmt.Errorf("bias correction: error in ModSpec")
	}

	if sseFn, bAdj, e = buildObj(modelPipe, nnModel, log); e != nil {
		return e
	}

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
		return fmt.Errorf("bias and adjustment have differing lengths: %d and %d", len(vals), len(optimal.X))
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

	if e = copyFiles(specs.modelDir(), loc); e != nil {
		return e
	}

	if e = nnModel.Save(loc + "model"); e != nil {
		return e
	}

	specs["modelDir"] = loc

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("assessment run time: %0.1f minutes", elapsed), true)

	return nil
}

func buildObj(pipe sea.Pipeline, nnModel *sea.NNModel, log *os.File) (objFn, []float64, error) {
	// get fit probabilities
	probs := nnModel.FitSlice()

	nCol := nnModel.Cols()
	if nCol == 1 {
		return nil, nil, fmt.Errorf("bias only for categorical models")
	}

	nRow := pipe.Rows()
	logOdds := make([]float64, nRow*(nCol-1))

	trgFt := pipe.GetFType(nnModel.ModSpec().TargetName())
	if trgFt == nil {
		return nil, nil, fmt.Errorf("target is missing from pipeline, bias corrections")
	}

	trgGData := pipe.Get(trgFt.From)
	trgData := trgGData.Data.([]int32)
	trgRates := make([]float64, nCol)

	// logodds is log(p[c]/p[nCol]) where c runs through first nCol-1 columns
	avgLogs := make([]float64, nCol-1)
	for row := 0; row < nRow; row++ {
		trgRates[trgData[row]]++
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

	for ind := 0; ind < nCol; ind++ {
		trgRates[ind] /= float64(nRow)
	}

	logger(log, fmt.Sprintf("bias correction target rates: %v", trgRates), true)

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
			errv := avg - trgRates[col]
			sse += errv * errv * wts[col]
		}

		return sse
	}

	// starting values
	var bAdj = make([]float64, nnModel.Cols()-1)
	for ind := 0; ind < len(bAdj); ind++ {
		targ := math.Log(trgRates[ind] / trgRates[nCol-1])
		bAdj[ind] = targ - avgLogs[ind]/float64(nRow)
	}

	return biasSse, bAdj, nil
}
