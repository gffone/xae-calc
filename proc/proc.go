package proc

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/xuri/excelize/v2"
)

const TIME_FORMAT = "15:04:05"

const ABSOLUTE_MINIMUM_NUMBER_OF_SIGNALS = 20
const PERCENTAGE_OF_DISCRIMINATION_FROM_THE_MAXIMUM_NUMBER_OF_SIGNALS = 10

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

func procData() (map[int][]LnDurPair, map[int]int, []int) {
	chanArr := make([]int, 0)
	dataMap := make(map[int][]LnDurPair)
	chanNMap := make(map[int]int)

	f, err := excelize.OpenFile("web/temp/files/file.xlsx")
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	cols, err := f.GetCols("Лист1")
	if err != nil {
		fmt.Println(err)
	}

	t0, _ := time.Parse(TIME_FORMAT, cols[0][1])

	for i := 1; i <= len(cols[0])-1; i++ {
		channel, _ := strconv.Atoi(cols[1][i])

		_, ok := dataMap[channel]

		if !ok {
			lnDurPairArr := make([]lnDurPair, 0, len(cols[0])-1)
			t, _ := time.Parse(TIME_FORMAT, cols[0][i])
			dur := t.Sub(t0)
			lnDurPairArr = append(lnDurPairArr, lnDurPair{math.Log(1), dur.Seconds()})
			math.Log(float64(len(dataMap[channel])))
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

	return dataMap, chanNMap, chanArr
}

func procXae(dataMap map[int][]lnDurPair, start, end float64) (map[int]float64, map[int]lnDurPairAprox) {

	xaeMap := make(map[int]float64)
	aproxMap := make(map[int]lnDurPairAprox)

	for currentChan, lnDurPairArr := range dataMap {
		tempV := []float64{}
		tempT := []float64{}
		for _, lnDurPair := range lnDurPairArr {
			if lnDurPair.durVal >= start && lnDurPair.durVal <= end {
				tempV = append(tempV, LnDurPair.lnVal)
				tempT = append(tempT, LnDurPair.durVal)
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

func graphImgRender(dataMap map[int][]LnDurPair, xaeMap map[int]float64, aproxMap map[int]lnDurPairAprox) {

	channelsTotal := len(xaeMap)

	for currentChan := 1; currentChan <= channelsTotal; currentChan++ {

		nTime := make([]float64, 0)
		nVals := make([]float64, 0)
		aproxDurVals := make([]float64, 0)
		aproxLnVals := make([]float64, 0)

		for _, LnDurPair := range dataMap[currentChan] {
			nTime = append(nTime, LnDurPair.durVal)
			nVals = append(nVals, LnDurPair.lnVal)
		}

		aproxDurVals = append(aproxDurVals, aproxMap[currentChan].aproxDurVals...)
		aproxLnVals = append(aproxLnVals, aproxMap[currentChan].aproxLnVals...)

		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name: "Время, с",
			},
			YAxis: chart.YAxis{
				Name: "ln(Σn)",
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					XValues: nTime,
					YValues: nVals,
				},
				chart.ContinuousSeries{
					Style: chart.Style{
						StrokeColor: drawing.ColorBlack,
						FillColor:   drawing.ColorBlack.WithAlpha(64),
					},
					XValues: aproxDurVals,
					YValues: aproxLnVals,
				},
			},
		}

		f, err := os.Create(fmt.Sprintf("web/temp/charts/%d.jpg", currentChan))
		if err != nil {
			fmt.Println(err)
		}
		graph.Render(chart.PNG, f)
	}
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

func xaeResult(xaeChanMap map[int]float64, dataMap map[int][]LnDurPair, chanNMap map[int]int, channelsArr []int) {
	var tempN int
	var tempChannel int
	var discriminationThreshold int

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
	confirmedСhannelsArr := make([]int, 0)

	discriminationThreshold = tempNArr[0] / PERCENTAGE_OF_DISCRIMINATION_FROM_THE_MAXIMUM_NUMBER_OF_SIGNALS

	for _, xaeV := range xaeArr {
		for channel, xae := range tempXaeArr {
			if xaeV == xae {
				tempN = len(dataMap[channel])
				tempChannel = channel
				break
			}
		}
		if tempN >= discriminationThreshold && tempN >= ABSOLUTE_MINIMUM_NUMBER_OF_SIGNALS {
			xaeBarData = append(xaeBarData, opts.BarData{Value: math.Round((xaeV*1000)*100) / 100})
			confirmedСhannelsArr = append(confirmedСhannelsArr, tempChannel)
		}
	}

	for _, v := range confirmedСhannelsArr {
		nArr = append(nArr, opts.BarData{Value: chanNMap[v]})
	}

	makeStandartPage(confirmedСhannelsArr, xaeBarData, nArr)
}

func makeStandartPage(confirmedСhannelsArr []int, xaeBarData []opts.BarData, nArr []opts.BarData) {
	page := components.NewPage()
	page.AddCharts(
		MakeStandartBarChart(confirmedСhannelsArr, xaeBarData),
		MakeStandartBarChartN(confirmedСhannelsArr, nArr),
	)
	f, err := os.Create("web/temp/bar/bar.html")
	if err != nil {
		fmt.Println(err)
	}
	page.Render(io.MultiWriter(f))
}

func makeStandartBarChart(confirmedСhannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1300px",
			Height: "900px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Xae, ∙10⁻³ с⁻¹",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Канал",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
		charts.WithColorsOpts(opts.Colors{"#0D6EFD"}),
		charts.WithLegendOpts(opts.Legend{Right: "10%"}),
	)

	bar.SetXAxis(confirmedСhannelsArr).
		AddSeries("Xae", xaeBarData, charts.WithSeriesAnimation(true)).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Position: "inside",
				Show:     true,
			}),
		)

	bar.XYReversal()
	return bar
}

func makeStandartBarChartN(confirmedСhannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1300px",
			Height: "900px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Суммарный счёт",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Канал",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
		charts.WithColorsOpts(opts.Colors{"#0D6EFD"}),
		charts.WithLegendOpts(opts.Legend{Right: "10%"}),
	)

	bar.SetXAxis(confirmedСhannelsArr).
		AddSeries("Xae", xaeBarData, charts.WithSeriesAnimation(true)).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Position: "inside",
				Show:     true,
			}),
		)

	bar.XYReversal()
	return bar
}

func StandartProc(start, end float64) {
	dataMap, chanNMap, chanArr := procData()
	
	xaeMap, aproxMap := procXae(dataMap, start, end)
	
	graphImgRender(dataMap, xaeMap, aproxMap)
	
	xaeResult(xaeMap, dataMap, chanNMap, chanArr)
}

func GroupModeProc(int_vals [][]int, start, end float64) {
	errR := os.RemoveAll("web/temp/group_charts/")
	if errR != nil {
		fmt.Println(errR)
	}
	errM := os.MkdirAll("web/temp/group_charts/", os.ModePerm)
	if errM != nil {
		fmt.Println(errM)
	}

	groupArr := make([]currentGroup, 0)

	dataMap, _, _ := procData()

	for _, chans := range int_vals {
		var currentGroup currentGroup
		for _, cur_chan := range chans {
			for _, v := range dataMap[cur_chan] {
				currentGroup.wholeSetOfDurVals = append(currentGroup.wholeSetOfDurVals, v.durVal)
				currentGroup.wholeSetOfLnVals = append(currentGroup.wholeSetOfLnVals, v.lnVal)
			}
		}
		groupArr = append(groupArr, CurrentGroup)
	}

	for _, v := range groupArr {
		sort.Float64s(v.wholeSetOfDurVals)
		sort.Float64s(v.wholeSetOfLnVals)
	}

	aproxDataMap := changeSignature(groupArr)

	xae, aproxMap := procXae(aproxDataMap, start, end)

	for i, v := range groupArr {

		aproxDurVals := make([]float64, 0)
		aproxLnVals := make([]float64, 0)

		aproxDurVals = append(aproxDurVals, aproxMap[i].aproxDurVals...)
		aproxLnVals = append(aproxLnVals, aproxMap[i].aproxLnVals...)

		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name: "Время, с",
			},
			YAxis: chart.YAxis{
				Name: "ln(Σn)",
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					XValues: v.wholeSetOfDurVals,
					YValues: v.wholeSetOfLnVals,
				},
				chart.ContinuousSeries{
					Style: chart.Style{
						StrokeColor: drawing.ColorBlack,
						FillColor:   drawing.ColorBlack.WithAlpha(64),
					},
					XValues: aproxDurVals,
					YValues: aproxLnVals,
				},
			},
		}
		f, errC := os.Create(fmt.Sprintf("web/temp/group_charts/group%d.jpg", i+1))
		if errC != nil {
			fmt.Println(errC)
		}
		graph.Render(chart.PNG, f)
	}

	xaeBarData := make([]opts.BarData, 0)
	confirmedСhannelsArr := make([]int, 0)

	for i := len(xae); i >= 1; i-- {
		xaeBarData = append(xaeBarData, opts.BarData{Value: math.Round((xae[i-1]*1000)*100) / 100})
		confirmedСhannelsArr = append(confirmedСhannelsArr, i)
	}
	makeGroupPage(confirmedСhannelsArr, xaeBarData)
}

func makeGroupPage(confirmedСhannelsArr []int, xaeBarData []opts.BarData) {
	page := components.NewPage()
	page.AddCharts(
		makeGroupBarChart(confirmedСhannelsArr, xaeBarData),
	)
	f, errC := os.Create("web/temp/bar/bar2.html")
	if errC != nil {
		fmt.Println(errC)
	}

	errR := page.Render(io.MultiWriter(f))
	if errR != nil {
		fmt.Println(errR)
	}
}

func makeGroupBarChart(confirmedСhannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "900px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Xae, ∙10⁻³ с⁻¹",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Группа",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
		charts.WithColorsOpts(opts.Colors{"#0D6EFD"}),
		charts.WithLegendOpts(opts.Legend{Right: "10%"}),
	)

	bar.SetXAxis(confirmedСhannelsArr).
		AddSeries("Xae", xaeBarData, charts.WithSeriesAnimation(true)).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Position: "inside",
				Show:     true,
			}),
		)
	bar.XYReversal()
	return bar
}
func changeSignature(groupArr []CurrentGroup) map[int][]LnDurPair {

	dataMap := make(map[int][]lnDurPair)

	for id, v := range groupArr {
		for i, ln := range v.wholeSetOfLnVals {
			dataMap[id] = append(dataMap[id], lnDurPair{ln, v.wholeSetOfDurVals[i]})
		}
	}
	return dataMap
}
