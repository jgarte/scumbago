package scumbag

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	irc "github.com/fluffle/goirc/client"
	"github.com/gocarina/gocsv"
)

const (
	// Date should be formatted as: MM-DD-YYYY
	coronaBaseURL = "https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_daily_reports/%s.csv"
)

var coronaHelp = []string{
	fmt.Sprintf("%s us <state> -- US details; state optional.", cmdCorona),
	fmt.Sprintf("%s <country>  -- Worldwide details; country optional.", cmdCorona),
}

type CoronaRecord struct {
	ProvinceState string `csv:"Province/State"`
	CountryRegion string `csv:"Country/Region"`
	LastUpdate    string `csv:"Last Update"`
	Confirmed     int    `csv:"Confirmed"`
	Deaths        int    `csv:"Deaths"`
	Recovered     int    `csv:"Recovered"`
	// Latitude
	// Longitude
}

// CoronaCommand pulls data from https://github.com/CSSEGISandData/COVID-19
type CoronaCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewCoronaCommand returns a new CoronaCommand instance.
func NewCoronaCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *CoronaCommand {
	return &CoronaCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *CoronaCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("CoronaCommand.Run()", err)
		return
	}

	fields := strings.Fields(args[0])

	if len(fields) <= 0 {
		go cmd.worldwide(channel)
		return
	}

	switch strings.ToLower(fields[0]) {
	case "us":
		go cmd.us(channel, fields[1:])
	default:
		go cmd.country(channel, fields)
	}
}

// Help displays the command help.
func (cmd *CoronaCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("CoronaCommand.Help()", err)
		return
	}

	for _, helpText := range coronaHelp {
		cmd.bot.Msg(cmd.conn, channel, helpText)
	}
}

func (cmd *CoronaCommand) us(channel string, args []string) {
	records, err := cmd.getRecords()
	if err != nil {
		cmd.bot.LogError("CoronaCommand.us()", err)
		return
	}

	var title string
	confirmed := 0
	deaths := 0
	recovered := 0

	state := strings.Join(args, " ")

	for _, r := range records {
		if r.CountryRegion == "US" {
			if len(args) > 0 {
				if strings.ToLower(r.ProvinceState) == strings.ToLower(state) {
					title = fmt.Sprintf("%s, US Corona Report", r.ProvinceState)
					confirmed = r.Confirmed
					deaths = r.Deaths
					recovered = r.Recovered
					break
				} else {
					title = fmt.Sprintf("State Not Found: %s", state)
				}
			} else {
				title = "US Corona Report"
				confirmed += r.Confirmed
				deaths += r.Deaths
				recovered += r.Recovered
			}
		}
	}

	details := cmd.buildDetails(confirmed, deaths, recovered)
	cmd.bot.Msg(cmd.conn, channel, "%s: %s", title, strings.Join(details, " / "))
}

func (cmd *CoronaCommand) country(channel string, args []string) {
	if len(args) == 0 {
		cmd.worldwide(channel)
		return
	}

	records, err := cmd.getRecords()
	if err != nil {
		cmd.bot.LogError("CoronaCommand.country()", err)
		return
	}

	var title string
	confirmed := 0
	deaths := 0
	recovered := 0

	country := strings.Join(args, " ")

	for _, r := range records {
		if strings.ToLower(r.CountryRegion) == strings.ToLower(country) {
			title = fmt.Sprintf("%s Corona Report", r.CountryRegion)
			confirmed = r.Confirmed
			deaths = r.Deaths
			recovered = r.Recovered
			break
		} else {
			title = fmt.Sprintf("Country Not Found: %s", country)
		}
	}

	details := cmd.buildDetails(confirmed, deaths, recovered)
	cmd.bot.Msg(cmd.conn, channel, "%s: %s", title, strings.Join(details, " / "))
}

func (cmd *CoronaCommand) worldwide(channel string) {
	records, err := cmd.getRecords()
	if err != nil {
		cmd.bot.LogError("CoronaCommand.worldwide()", err)
		return
	}

	confirmed := 0
	deaths := 0
	recovered := 0

	for _, r := range records {
		confirmed += r.Confirmed
		deaths += r.Deaths
		recovered += r.Recovered
	}

	details := cmd.buildDetails(confirmed, deaths, recovered)
	cmd.bot.Msg(cmd.conn, channel, "Worldwide Corona Report: %s", strings.Join(details, " / "))
}

func (cmd *CoronaCommand) getCSV(date time.Time) (*http.Response, error) {
	url := fmt.Sprintf(coronaBaseURL, date.Format("01-02-2006"))
	resp, err := getResponse(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (cmd *CoronaCommand) getRecords() ([]*CoronaRecord, error) {
	date := time.Now()

	csv, err := cmd.getCSV(date)
	if err != nil {
		return nil, err
	}
	if csv.StatusCode == http.StatusNotFound {
		cmd.bot.Log.Debug("CoronaCommand.getRecords(): 404: retrying with previous day")
		date = date.AddDate(0, 0, -1)
		csv, err = cmd.getCSV(date)
		if err != nil {
			return nil, err
		}
	}
	defer csv.Body.Close()

	records := []*CoronaRecord{}
	err = gocsv.Unmarshal(csv.Body, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (cmd *CoronaCommand) buildDetails(confirmed, deaths, recovered int) []string {
	return []string{
		fmt.Sprintf("Confirmed: %s", humanize.Comma(int64(confirmed))),
		fmt.Sprintf("Deaths: %s", humanize.Comma(int64(deaths))),
		fmt.Sprintf("Recovered: %s", humanize.Comma(int64(recovered))),
	}
}
