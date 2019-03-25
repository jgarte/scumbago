package scumbag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	api "github.com/Henry-Sarabia/apicalypse"
	irc "github.com/fluffle/goirc/client"
)

const (
	igdbAPIURL   = "https://api-v3.igdb.com"
	igdbGamesURL = igdbAPIURL + "/games"
)

var gameHelp = []string{
	cmdGame + " <search>",
	cmdGame + " -recent",
	cmdGame + " -upcoming",
}

// From the /platforms endpoint.
var gamePlatforms = map[int]string{
	3:   "Linux",
	4:   "Nintendo 64",
	5:   "Wii",
	6:   "PC",
	7:   "PS1",
	8:   "PS2",
	9:   "PS3",
	11:  "Xbox",
	12:  "Xbox 360",
	13:  "DOS",
	14:  "Mac",
	15:  "Commodore C64/128",
	18:  "NES",
	19:  "SNES",
	20:  "Nintendo DS",
	21:  "GameCube",
	22:  "Game Boy Color",
	23:  "Dreamcast",
	24:  "Game Boy Advance",
	32:  "Sega Saturn",
	33:  "Game Boy",
	35:  "Sega Game Gear",
	36:  "Xbox Live Arcade",
	37:  "Nintendo 3DS",
	38:  "PSP",
	39:  "iOS",
	41:  "Wii U",
	45:  "PSN",
	46:  "PS Vita",
	47:  "Virtual Console (Nintendo)",
	48:  "PS4",
	49:  "Xbox One",
	51:  "Famicom Disk System",
	52:  "Arcade",
	55:  "Mobile",
	56:  "WiiWare",
	59:  "Atari 2600",
	60:  "Atari 7800",
	62:  "Atari Jaguar",
	65:  "Atari 8-bit",
	66:  "Atari 5200",
	67:  "Intellivision",
	68:  "ColecoVision",
	72:  "Ouya",
	74:  "Windows Phone",
	75:  "Apple II",
	78:  "Sega CD",
	80:  "Neo Geo AES",
	82:  "Web browser",
	87:  "Virtual Boy",
	92:  "SteamOS",
	115: "Apple IIGS",
	120: "Neo Geo Pocket Color",
	126: "TRS-80",
	129: "Texas Instruments TI-99",
	130: "Nintendo Switch",
	136: "Neo Geo CD",
	160: "Nintendo eShop",
	163: "SteamVR",
	164: "Daydream",
	165: "PlayStation VR",
}

// GameResult represents game data returned from the API.
type GameResult struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`

	FirstReleaseDate int64 `json:"first_release_date"`
	Platforms        []int `json:"platforms"`
}

func (g GameResult) ReleaseDate() string {
	t := time.Unix(g.FirstReleaseDate, 0)
	return fmt.Sprintf("%02d/%02d/%d", t.Month(), t.Day(), t.Year())
}

func (g GameResult) PlatformString() string {
	out := []string{}
	for _, p := range g.Platforms {
		out = append(out, gamePlatforms[p])
	}
	return strings.Join(out, "/")
}

// GameCommand interacts with the IGDB.com API.
type GameCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewGameCommand returns a new GameCommand instance.
func NewGameCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *GameCommand {
	return &GameCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *GameCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.Run()")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("GameCommand.Run(): No query")
		return
	}

	cmdArgs := strings.Fields(query)
	if len(cmdArgs) <= 0 {
		cmd.Help()
		return
	}

	switch cmdArgs[0] {
	case "-recent":
		cmd.recent(channel)
	case "-upcoming":
		cmd.upcoming(channel)
	default:
		cmd.search(channel, query)
	}
}

// Help displays the command help.
func (cmd *GameCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.Help()")
		return
	}

	for _, helpText := range gameHelp {
		cmd.bot.Msg(cmd.conn, channel, helpText)
	}
}

func (cmd *GameCommand) search(channel, query string) {
	// Released in the last 20 years.
	releaseDate := strconv.FormatInt(time.Now().AddDate(-20, 0, 0).Unix(), 10)

	cmd.bot.Log.WithField("query", query).Debug("GameCommand.search()")
	req, err := api.NewRequest(
		"POST",
		igdbGamesURL,
		api.Limit(1),
		api.Fields("*"),
		api.Where("rating >= 80 & release_dates.date > "+releaseDate),
		api.Search("", query),
	)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	req.Header.Add("user-key", cmd.bot.Config.IGDB.Key)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	content, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	cmd.bot.Log.WithField("content", string(content)).Debug("GameCommand.search()")

	results := make([]GameResult, 1)
	err = json.Unmarshal(content, &results)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	if len(results) <= 0 {
		cmd.bot.Msg(cmd.conn, channel, "Nothing found.")
		return
	}
	game := results[0]

	cmd.bot.Log.WithField("game", game).Debug("GameCommand.search()")

	info := fmt.Sprintf("[%s] %s - %s", game.ReleaseDate(), game.PlatformString(), game.Name)
	cmd.bot.Msg(cmd.conn, channel, info)
	cmd.bot.Msg(cmd.conn, channel, game.Summary)
}

func (cmd *GameCommand) recent(channel string) {
	cmd.bot.Log.Debug("GameCommand.recent()")
}

func (cmd *GameCommand) upcoming(channel string) {
	cmd.bot.Log.Debug("GameCommand.upcoming()")
}
