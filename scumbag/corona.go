package scumbag

import (
	"fmt"
	"strings"

	"github.com/Oshuma/corona"
	"github.com/dustin/go-humanize"
	irc "github.com/fluffle/goirc/client"
)

const (
	coronaBaseURL = "https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_daily_reports/%s.csv"
)

var coronaHelp = []string{
	fmt.Sprintf("%s us <state> -- US details; state optional.", cmdCorona),
	fmt.Sprintf("%s <country>  -- Worldwide details; country optional.", cmdCorona),
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
	var title string
	var confirmed, deaths, recovered int

	if len(args) > 0 {
		state := strings.Join(args, " ")

		cases, err := corona.DailyByState(state)
		if err == corona.ErrorNoCasesFound {
			cmd.bot.Msg(cmd.conn, channel, "State Not Found: %s", state)
			return
		} else if err != nil {
			cmd.bot.LogError("CoronaCommand.us()", err)
			return
		}

		title = fmt.Sprintf("%s, US Corona Report", cases.ProvinceState)
		confirmed = cases.Confirmed
		deaths = cases.Deaths
		recovered = cases.Recovered
	} else {
		title = "US Corona Report"

		cases, err := corona.DailyByCountry("US")
		if err == corona.ErrorNoCasesFound {
			cmd.bot.Msg(cmd.conn, channel, "No data for US (something probably fucked up)")
			return
		} else if err != nil {
			cmd.bot.LogError("CoronaCommand.us()", err)
			return
		}

		for _, c := range cases {
			confirmed += c.Confirmed
			deaths += c.Deaths
			recovered += c.Recovered
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

	country := strings.Join(args, " ")
	cases, err := corona.DailyByCountry(country)
	if err == corona.ErrorNoCasesFound {
		cmd.bot.Msg(cmd.conn, channel, "Country Not Found: %s", country)
		return
	} else if err != nil {
		cmd.bot.LogError("CoronaCommand.country()", err)
		return
	}

	title := cases[0].CountryRegion
	var confirmed, deaths, recovered int

	for _, c := range cases {
		confirmed += c.Confirmed
		deaths += c.Deaths
		recovered += c.Recovered
	}

	details := cmd.buildDetails(confirmed, deaths, recovered)
	cmd.bot.Msg(cmd.conn, channel, "%s: %s", title, strings.Join(details, " / "))
}

func (cmd *CoronaCommand) worldwide(channel string) {
	cases, err := corona.DailyWorldwide()
	if err != nil {
		cmd.bot.LogError("CoronaCommand.worldwide()", err)
		return
	}

	var confirmed, deaths, recovered int
	for _, r := range cases {
		confirmed += r.Confirmed
		deaths += r.Deaths
		recovered += r.Recovered
	}

	details := cmd.buildDetails(confirmed, deaths, recovered)
	cmd.bot.Msg(cmd.conn, channel, "Worldwide Corona Report: %s", strings.Join(details, " / "))
}

func (cmd *CoronaCommand) buildDetails(confirmed, deaths, recovered int) []string {
	return []string{
		fmt.Sprintf("Confirmed: %s", humanize.Comma(int64(confirmed))),
		fmt.Sprintf("Deaths: %s", humanize.Comma(int64(deaths))),
		fmt.Sprintf("Recovered: %s", humanize.Comma(int64(recovered))),
	}
}
