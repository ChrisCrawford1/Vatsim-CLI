package main

import (
	"MyVatsimCLI/data"
	"MyVatsimCLI/services"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"log"
	"sort"
	"strconv"
	"time"
)

var utcZone, _ = time.LoadLocation("UTC")

func main() {
	var pulledData = getCurrentData()

	// This is already in UTC, no need to convert
	var nextUpdate = pulledData.General.UpdateTimestamp.Add(time.Minute * 5)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	defer ui.Close()

	/**
	Initial setup for the grid.
	*/
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	renderUi := func(paragraph *widgets.Paragraph, table *widgets.Table) (*widgets.Paragraph, *widgets.Table) {
		ui.Clear()

		grid.Set(
			ui.NewRow(1.0/2,
				ui.NewCol(1.0/2, paragraph),
				ui.NewCol(1.0/2, table),
			),
		)

		ui.Render(grid)

		return paragraph, table
	}

	p, _ := renderUi(
		infoPanel(nextUpdateTime(pulledData.General.UpdateTimestamp), pulledData.GetConnectionsPerATCRating()),
		table(&pulledData),
	)

	updateTitle := func(p *widgets.Paragraph) {
		p.Title = "Vatsim CLI " + time.Now().In(utcZone).Format(time.RFC1123)
	}

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "r", "<C-r>":
				manualRefreshData := getCurrentData()
				p, _ = renderUi(
					infoPanel(
						nextUpdateTime(manualRefreshData.General.UpdateTimestamp),
						manualRefreshData.GetConnectionsPerATCRating(),
					),
					table(&manualRefreshData),
				)
				nextUpdate = manualRefreshData.General.UpdateTimestamp.Add(time.Minute * 5)
				updateTitle(p)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		case <-ticker:
			if time.Now().In(utcZone).After(nextUpdate) {
				updatedData := getCurrentData()
				p, _ = renderUi(
					infoPanel(
						nextUpdateTime(updatedData.General.UpdateTimestamp),
						updatedData.GetConnectionsPerATCRating(),
					),
					table(&updatedData),
				)
				nextUpdate = updatedData.General.UpdateTimestamp.Add(time.Minute * 5)
				updateTitle(p)
				continue
			}
			updateTitle(p)
			ui.Clear()
			ui.Render(grid)
		}
	}
}

func table(data *data.Datafile) *widgets.Table {
	dataTable := widgets.NewTable()
	var rows [][]string

	rows = append(rows, []string{"Departures", "Arrivals"})
	airfields := data.GetPopularAirfields()
	departures := airfields.Departures[0:5]
	arrivals := airfields.Arrivals[0:5]

	for i, departure := range departures {
		rows = append(rows, []string{
			departure.Icao + ":" + strconv.Itoa(departure.Count),
			arrivals[i].Icao + ":" + strconv.Itoa(arrivals[i].Count),
		})
	}

	dataTable.Rows = rows
	dataTable.TextStyle = ui.NewStyle(ui.ColorCyan)
	dataTable.Title = "Most Popular Airfields"
	dataTable.BorderStyle.Fg = ui.ColorGreen
	dataTable.SetRect(20, 2, 100, 10)

	return dataTable
}

func infoPanel(nextUpdateTime string, connections map[string]int) *widgets.Paragraph {
	displayLine := "Next update: " + nextUpdateTime + "\nCurrent Controller Connections By Rating:\n"

	/**
	This probably shouldn't live here, its just to get alpha sorting initially.
	TODO - Refactor me!
	*/
	keys := make([]string, 0)
	for k, _ := range connections {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, rating := range keys {
		displayLine += rating + "[" + strconv.Itoa(connections[rating]) + "]\n"
	}

	p := widgets.NewParagraph()
	p.Title = "Vatsim CLI"
	p.TextStyle.Fg = ui.ColorGreen
	p.Text = displayLine
	p.TextStyle.Fg = ui.ColorCyan
	p.BorderStyle.Fg = ui.ColorMagenta

	return p
}

func getCurrentData() data.Datafile {
	return services.FetchCurrentData()
}

func nextUpdateTime(providedTime time.Time) string {
	return providedTime.Add(time.Minute * 5).Format(time.RFC822)
}
