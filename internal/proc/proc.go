package proc

import (
	"math"
	"proc/internal/storage"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

const TIME_FORMAT = "15:04:05"
const MIN_SIGNAL_THRESHOLD = 20
const SIGNAL_DISCRIMINATION_PERCENTAGE = 10

type lnDurPair struct {
	lnVal  float64
	durVal float64
}

type lnDurPairAprox struct {
	aproxLnVals  []float64
	aproxDurVals []float64
}

type currentGroup struct {
	wholeSetOfLnVals  []float64
	wholeSetOfDurVals []float64
}

func procData(strg *storage.Storage) (map[int][]lnDurPair, map[int]int, []int, error) {
	chanArr := make([]int, 0)
	dataMap := make(map[int][]lnDurPair)
	chanNMap := make(map[int]int)

	cols, err := strg.GetData()
	if err != nil {
		return nil, nil, nil, err
	}

	t0, _ := time.Parse(TIME_FORMAT, cols[0][1])

	for i := 1; i <= len(cols[0])-1; i++ {
		channel, err := strconv.Atoi(cols[1][i])
		if err != nil {
			return nil, nil, nil, err
		}

		_, ok := dataMap[channel]

		if !ok {
			lnDurPairArr := make([]lnDurPair, 0, len(cols[0])-1)
			t, _ := time.Parse(TIME_FORMAT, cols[0][i])
			dur := t.Sub(t0)
			lnDurPairArr = append(lnDurPairArr, lnDurPair{math.Log(1), dur.Seconds()})
			dataMap[channel] = lnDurPairArr
			chanNMap[channel] = 1
			chanArr = append(chanArr, channel)
		} else {
			t, _ := time.Parse(TIME_FORMAT, cols[0][i])
			dur := t.Sub(t0)
			dataMap[channel] = append(dataMap[channel], lnDurPair{math.Log(float64(len(dataMap[channel]) + 1)), dur.Seconds()})
			chanNMap[channel]++
		}
	}
	sort.Ints(chanArr)

	return dataMap, chanNMap, chanArr, nil
}

func procXae(dataMap map[int][]lnDurPair, start, end float64) (map[int]float64, map[int]lnDurPairAprox) {
	xaeMap := make(map[int]float64)
	aproxMap := make(map[int]lnDurPairAprox)

	for currentChan, lnDurPairArr := range dataMap {
		var tempV []float64
		var tempT []float64
		for _, pair := range lnDurPairArr {
			if pair.durVal >= start && pair.durVal <= end {
				tempV = append(tempV, pair.lnVal)
				tempT = append(tempT, pair.durVal)
			}
		}
		a, b := linearTrend(tempT, tempV)
		xaeMap[currentChan] = a

		aproxLnVals := generateAproxLnVals(a, b, tempT)
		aproxMap[currentChan] = lnDurPairAprox{aproxLnVals, tempT}
	}

	return xaeMap, aproxMap
}

func generateAproxLnVals(a, b float64, time []float64) []float64 {
	items := make([]float64, 0, len(time))

	for _, curTime := range time {
		items = append(items, a*curTime+b)
	}

	return items
}

func graphImgCreateStandard(dataMap map[int][]lnDurPair, channelsTotal int, aproxMap map[int]lnDurPairAprox, strg *storage.Storage) error {
	for currentChan := 1; currentChan <= channelsTotal; currentChan++ {
		var nTime, nVals, aproxDurVals, aproxLnVals []float64

		for _, pair := range dataMap[currentChan] {
			nTime = append(nTime, pair.durVal)
			nVals = append(nVals, pair.lnVal)
		}

		aproxDurVals = append(aproxDurVals, aproxMap[currentChan].aproxDurVals...)
		aproxLnVals = append(aproxLnVals, aproxMap[currentChan].aproxLnVals...)

		err := strg.CreateChartStandard(nTime, nVals, aproxDurVals, aproxLnVals, currentChan)
		if err != nil {
			return err
		}
	}
	return nil
}

func xaeResultStandard(xaeChanMap map[int]float64, dataMap map[int][]lnDurPair, chanNMap map[int]int, channelsArr []int, strg *storage.Storage) error {
	var tempN, tempChannel, discriminationThreshold int

	nArr := make([]opts.BarData, 0)
	xaeArr := make([]float64, 0)
	tempNArr := make([]int, 0)
	tempXaeArr := make([]float64, len(channelsArr)+1)

	for _, n := range chanNMap {
		tempNArr = append(tempNArr, n)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(tempNArr)))

	for currentChan, xae := range xaeChanMap {
		xaeArr = append(xaeArr, xae)
		tempXaeArr[currentChan] = xae
	}

	sort.Float64s(xaeArr)

	xaeBarData := make([]opts.BarData, 0)
	confirmedChannelsArr := make([]int, 0)

	discriminationThreshold = tempNArr[0] / SIGNAL_DISCRIMINATION_PERCENTAGE

	for _, xaeV := range xaeArr {
		for channel, xae := range tempXaeArr {
			if xaeV == xae {
				tempN = len(dataMap[channel])
				tempChannel = channel
				break
			}
		}
		if tempN >= discriminationThreshold && tempN >= MIN_SIGNAL_THRESHOLD {
			xaeBarData = append(xaeBarData, opts.BarData{Value: barValue(xaeV)})
			confirmedChannelsArr = append(confirmedChannelsArr, tempChannel)
		}
	}

	for _, v := range confirmedChannelsArr {
		nArr = append(nArr, opts.BarData{Value: chanNMap[v]})
	}

	err := strg.MakeStandardPage(confirmedChannelsArr, xaeBarData, nArr)
	if err != nil {
		return err
	}
	return nil
}

func StandardProc(start, end float64, strg *storage.Storage) error {
	dataMap, chanNMap, chanArr, err := procData(strg)
	if err != nil {
		return err
	}

	xaeMap, aproxMap := procXae(dataMap, start, end)

	err = graphImgCreateStandard(dataMap, len(xaeMap), aproxMap, strg)
	if err != nil {
		return err
	}

	err = xaeResultStandard(xaeMap, dataMap, chanNMap, chanArr, strg)
	if err != nil {
		return err
	}

	return nil
}

func GroupModeProc(intVals [][]int, start, end float64, strg *storage.Storage) error {
	dataMap, _, _, err := procData(strg)
	if err != nil {
		return err
	}

	aproxDataMap, groupArr := procAproxData(dataMap, intVals)

	xae, aproxMap := procXae(aproxDataMap, start, end)

	err = graphImgRenderGroup(groupArr, aproxMap, strg)
	if err != nil {
		return err
	}

	err = xaeResultGroup(xae, strg)
	if err != nil {
		return err
	}

	return nil
}

func xaeResultGroup(xae map[int]float64, strg *storage.Storage) error {
	xaeBarData := make([]opts.BarData, 0)
	confirmedChannelsArr := make([]int, 0)

	for i := len(xae); i >= 1; i-- {
		xaeBarData = append(xaeBarData, opts.BarData{Value: barValue(xae[i-1])})
		confirmedChannelsArr = append(confirmedChannelsArr, i)
	}
	err := strg.MakeGroupPage(confirmedChannelsArr, xaeBarData)
	if err != nil {
		return err
	}
	return nil
}

func graphImgRenderGroup(groupArr []currentGroup, aproxMap map[int]lnDurPairAprox, strg *storage.Storage) error {
	for groupNumber, group := range groupArr {
		aproxDurVals := make([]float64, 0)
		aproxLnVals := make([]float64, 0)

		aproxDurVals = append(aproxDurVals, aproxMap[groupNumber].aproxDurVals...)
		aproxLnVals = append(aproxLnVals, aproxMap[groupNumber].aproxLnVals...)

		err := strg.CreateChartGroup(group.wholeSetOfDurVals, group.wholeSetOfLnVals, aproxDurVals, aproxLnVals, groupNumber)
		if err != nil {
			return err
		}
	}
	return nil
}

func procAproxData(dataMap map[int][]lnDurPair, intVals [][]int) (map[int][]lnDurPair, []currentGroup) {
	var groupArr []currentGroup

	for _, chans := range intVals {
		var cg currentGroup
		for _, curChan := range chans {
			for _, v := range dataMap[curChan] {
				cg.wholeSetOfDurVals = append(cg.wholeSetOfDurVals, v.durVal)
				cg.wholeSetOfLnVals = append(cg.wholeSetOfLnVals, v.lnVal)
			}
		}
		groupArr = append(groupArr, cg)
	}

	for _, v := range groupArr {
		sort.Float64s(v.wholeSetOfDurVals)
		sort.Float64s(v.wholeSetOfLnVals)
	}

	aproxDataMap := changeSignature(groupArr)

	return aproxDataMap, groupArr
}

func changeSignature(groupArr []currentGroup) map[int][]lnDurPair {
	dataMap := make(map[int][]lnDurPair)

	for id, v := range groupArr {
		for i, ln := range v.wholeSetOfLnVals {
			dataMap[id] = append(dataMap[id], lnDurPair{ln, v.wholeSetOfDurVals[i]})
		}
	}

	return dataMap
}

func barValue(n float64) float64 {
	return math.Round((n*1000)*100) / 100
}

func linearTrend(dataT, dataV []float64) (float64, float64) {
	var sumX, sumY, sumXY, sumXX float64

	n := len(dataV)
	for i, y := range dataV {
		x := dataT[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / float64(n)

	return slope, intercept
}
