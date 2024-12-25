package storage

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
)

type Storage struct {
	StoragePath   string
	WorksheetName string
}

func NewStorage(StoragePath string, WorksheetName string) *Storage {
	return &Storage{StoragePath: StoragePath, WorksheetName: WorksheetName}
}

func (s *Storage) GetData() ([][]string, error) {
	f, err := excelize.OpenFile(fmt.Sprintf("%s/files/file.xlsx", s.StoragePath))
	defer f.Close()

	if err != nil {
		return nil, err
	}

	cols, err := f.GetCols(s.WorksheetName)
	if err != nil {
		return nil, err
	}

	return cols, nil
}

func (s *Storage) CreateChartStandard(nTime, nVals, aproxDurVals, aproxLnVals []float64, currentChan int) error {
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

	f, err := os.Create(fmt.Sprintf("%s/charts/%d.jpg", s.StoragePath, currentChan))
	if err != nil {
		return err
	}
	err = graph.Render(chart.PNG, f)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) MakeStandardPage(confirmedChannelsArr []int, xaeBarData []opts.BarData, nArr []opts.BarData) error {
	page := components.NewPage()
	page.AddCharts(
		makeStandardBarChart(confirmedChannelsArr, xaeBarData),
		makeStandardBarChartN(confirmedChannelsArr, nArr),
	)

	f, err := os.Create(fmt.Sprintf("%s/bar/bar.html", s.StoragePath))
	if err != nil {
		return err
	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return err
	}
	return nil
}

func makeStandardBarChart(confirmedChannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
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

	bar.SetXAxis(confirmedChannelsArr).
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

func makeStandardBarChartN(confirmedChannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
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

	bar.SetXAxis(confirmedChannelsArr).
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

func (s *Storage) CreateChartGroup(wholeSetOfDurVals, wholeSetOfLnVals, aproxDurVals, aproxLnVals []float64, groupNumber int) error {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Время, с",
		},
		YAxis: chart.YAxis{
			Name: "ln(Σn)",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: wholeSetOfDurVals,
				YValues: wholeSetOfLnVals,
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

	f, err := os.Create(fmt.Sprintf("%s/group_charts/group%d.jpg", s.StoragePath, groupNumber+1))
	if err != nil {
		return err
	}
	err = graph.Render(chart.PNG, f)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) MakeGroupPage(confirmedChannelsArr []int, xaeBarData []opts.BarData) error {
	page := components.NewPage()
	page.AddCharts(
		makeGroupBarChart(confirmedChannelsArr, xaeBarData),
	)

	f, err := os.Create(fmt.Sprintf("%s/bar/bar2.html", s.StoragePath))
	if err != nil {
		return err
	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return err
	}

	return nil
}

func makeGroupBarChart(confirmedChannelsArr []int, xaeBarData []opts.BarData) *charts.Bar {
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

	bar.SetXAxis(confirmedChannelsArr).
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
