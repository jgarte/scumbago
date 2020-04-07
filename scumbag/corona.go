package scumbag

import (
	"fmt"
	"strings"

	"github.com/Oshuma/corona"
	"github.com/dustin/go-humanize"
	irc "github.com/fluffle/goirc/client"
)

var coronaHelp = []string{
	fmt.Sprintf("%s us <state> -- US details; state optional.", cmdCorona),
	fmt.Sprintf("%s <country>  -- Worldwide details; country optional.", cmdCorona),
}

type CoronaReport struct {
	Title     string
	Channel   string
	Confirmed float64
	Deaths    float64
}

func (r *CoronaReport) Details() string {
	return strings.Join([]string{
		fmt.Sprintf("Confirmed: %s", humanize.Comma(int64(r.Confirmed))),
		fmt.Sprintf("Deaths: %s", humanize.Comma(int64(r.Deaths))),
	}, " / ")
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
	var confirmed, deaths float64

	if len(args) > 0 {
		state := strings.Join(args, " ")

		usCases, err := corona.DailyByCountryCode("US")
		if err != nil {
			cmd.bot.LogError("CoronaCommand.us()", err)
			return
		}

		cases, err := usCases.FilterRegionName(state)
		if err == corona.ErrorNoReportsFound {
			cmd.bot.Msg(cmd.conn, channel, "State not found: %s", state)
			return
		} else if err != nil {
			cmd.bot.LogError("CoronaCommand.us()", err)
			return
		}

		title = cases[0].Region.Name

		for _, r := range cases {
			confirmed += r.Confirmed
			deaths += r.Deaths
		}
	} else {
		total, err := corona.TotalByCountryCode("US")
		if err != nil {
			cmd.bot.LogError("CoronaCommand.us()", err)
			return
		}
		title = "US Corona Report"
		confirmed = total.Confirmed
		deaths = total.Deaths
	}

	r := &CoronaReport{
		Title:     title,
		Channel:   channel,
		Confirmed: confirmed,
		Deaths:    deaths,
	}
	cmd.sendReport(r)
}

func (cmd *CoronaCommand) country(channel string, args []string) {
	if len(args) == 0 {
		cmd.worldwide(channel)
		return
	}

	country := strings.Join(args, " ")

	total, err := corona.TotalByCountryName(country)
	if err == corona.ErrorNoReportsFound {
		cmd.bot.Msg(cmd.conn, channel, "Country Not Found: %s", country)
		return
	} else if err != nil {
		cmd.bot.LogError("CoronaCommand.country()", err)
		return
	}

	r := &CoronaReport{
		Title:     fmt.Sprintf("%s Corona Report", total.Country.Name),
		Channel:   channel,
		Confirmed: total.Confirmed,
		Deaths:    total.Deaths,
	}
	cmd.sendReport(r)
}

func (cmd *CoronaCommand) worldwide(channel string) {
	cases, err := corona.DailyWorldwide()
	if err != nil {
		cmd.bot.LogError("CoronaCommand.worldwide()", err)
		return
	}

	cases, err = cases.FilterRegionName("")
	if err != nil {
		cmd.bot.LogError("CoronaCommand.worldwide()", err)
		return
	}

	var confirmed, deaths float64
	for _, c := range cases {
		confirmed += c.Confirmed
		deaths += c.Deaths
	}

	r := &CoronaReport{
		Title:     "Worldwide Corona Report",
		Channel:   channel,
		Confirmed: confirmed,
		Deaths:    deaths,
	}
	cmd.sendReport(r)
}

func (cmd *CoronaCommand) sendReport(r *CoronaReport) {
	cmd.bot.Msg(cmd.conn, r.Channel, "%s: %s", r.Title, r.Details())
}
