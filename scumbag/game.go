package scumbag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	api "github.com/Henry-Sarabia/apicalypse"
	irc "github.com/fluffle/goirc/client"
)

const (
	igdbAPIURL          = "https://api-v3.igdb.com"
	igdbGamesURL        = igdbAPIURL + "/games"
	igdbReleaseDatesURL = igdbAPIURL + "/release_dates"

	igdbLimit     = 5
	igdbTimeframe = 7
)

var gameHelp = []string{
	cmdGame + " <search>",
	cmdGame + " -recent <pc/ps/mobile/nin/xbox>",
	cmdGame + " -upcoming <pc/ps/mobile/nin/xbox>",
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
	34:  "Android",
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

// Pulled from relevant gamePlatforms; used in recent and upcoming subcommands.
var platformCategories = map[string][]int{
	"pc":     []int{3, 6, 13, 14, 82, 92},            // PC (Steam, etc.)
	"ps":     []int{9, 38, 45, 46, 48, 165},          // PlayStation
	"mobile": []int{34, 39, 55, 74},                  // Mobile games
	"nin":    []int{5, 20, 37, 41, 47, 56, 130, 160}, // Nintendo
	"xbox":   []int{11, 12, 36, 49},                  // Xbox
}

// Game represents game data returned from the API.
type Game struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`

	FirstReleaseDate int64 `json:"first_release_date"`
	Platforms        []int `json:"platforms"`
}

// GameRelease represents game release data.
type GameRelease struct {
	ID   int64 `json:"id"`
	Date int64 `json:"date"`
	Game Game  `json:"game"`
}

func (g Game) Info() string {
	return fmt.Sprintf("[%s] %s - %s", g.ReleaseDate(), g.PlatformString(), g.Name)
}

func (g Game) ReleaseDate() string {
	return formatDate(g.FirstReleaseDate)
}

func (r GameRelease) ReleaseDate() string {
	return formatDate(r.Date)
}

func (g Game) PlatformString() string {
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
		cmd.recent(channel, strings.Join(cmdArgs[1:], ""))
	case "-upcoming":
		cmd.upcoming(channel, strings.Join(cmdArgs[1:], ""))
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
	req, err := api.NewRequest(
		"POST",
		igdbGamesURL,
		api.Limit(1),
		api.Fields("name, summary, first_release_date, platforms"),
		api.Where("rating >= 80"),
		api.Search("", query),
	)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	cmd.addHeaders(req)

	content, err := getContentBytes(req)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.search()")
		return
	}

	results := make([]Game, 1)
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

	cmd.bot.Msg(cmd.conn, channel, game.Info())
	cmd.bot.Msg(cmd.conn, channel, game.Summary)
}

func (cmd *GameCommand) recent(channel, platform string) {
	if len(platformCategories[platform]) == 0 {
		cmd.bot.Msg(cmd.conn, channel, "Unknown platform: "+platform)
		cmd.Help()
		return
	}

	// Games released in the past N days.
	date := time.Now().AddDate(0, 0, -igdbTimeframe)
	releases, err := cmd.releases(date, platformCategories[platform])
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.recent()")
		return
	}

	if len(releases) <= 0 {
		cmd.bot.Msg(cmd.conn, channel, "Nothing.")
		return
	}

	for _, release := range releases {
		cmd.bot.Msg(cmd.conn, channel, "[%s] - %s", release.ReleaseDate(), release.Game.Name)
	}
}

func (cmd *GameCommand) upcoming(channel, platform string) {
	if len(platformCategories[platform]) == 0 {
		cmd.bot.Msg(cmd.conn, channel, "Unknown platform: "+platform)
		cmd.Help()
		return
	}

	// Games being released in the next N days.
	date := time.Now().AddDate(0, 0, igdbTimeframe)
	releases, err := cmd.releases(date, platformCategories[platform])
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("GameCommand.upcoming()")
		return
	}

	if len(releases) <= 0 {
		cmd.bot.Msg(cmd.conn, channel, "Nothing.")
		return
	}

	for _, release := range releases {
		cmd.bot.Msg(cmd.conn, channel, "[%s] - %s", release.ReleaseDate(), release.Game.Name)
	}
}

func (cmd *GameCommand) addHeaders(req *http.Request) {
	req.Header.Add("user-key", cmd.bot.Config.IGDB.Key)
	req.Header.Add("Accept", "application/json")
}

func (cmd *GameCommand) releases(date time.Time, platforms []int) ([]GameRelease, error) {
	req, err := api.NewRequest(
		"POST",
		igdbReleaseDatesURL,
		api.Limit(igdbLimit),
		api.Fields("*, game.name, game.summary, game.first_release_date, game.platforms"),
		api.Where(
			fmt.Sprintf("platform = %s", joinPlatforms(platforms)),
			fmt.Sprintf("date > %s", strconv.FormatInt(date.Unix(), 10)),
		),
		api.Sort("date", "asc"),
	)
	if err != nil {
		return nil, err
	}

	cmd.addHeaders(req)

	content, err := getContentBytes(req)
	if err != nil {
		return nil, err
	}

	results := make([]GameRelease, igdbLimit)
	err = json.Unmarshal(content, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func formatDate(date int64) string {
	t := time.Unix(date, 0)
	return fmt.Sprintf("%02d/%02d/%d", t.Month(), t.Day(), t.Year())
}

func joinPlatforms(platformIDs []int) string {
	s := fmt.Sprint(platformIDs)
	s = strings.Replace(s, "[", "(", -1)
	s = strings.Replace(s, "]", ")", -1)
	s = strings.Join(strings.Fields(s), ",")
	return s
}
