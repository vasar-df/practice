package vasar

import (
	"fmt"
	"github.com/brandenc40/romannumeral"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game"
	"math"
	"strings"
	"time"
)

// startLeaderboards spawns and starts updating lobby leaderboards. I sincerely apologize for the horrible code.
func (v *Vasar) startLeaderboards() {
	b := entity.NewText("", mgl64.Vec3{-6.5, 60, -15.5})
	v.srv.World().AddEntity(b)

	leaderboards := []string{"global"}
	for _, g := range game.Games() {
		leaderboards = append(leaderboards, g.String())
	}
	leaderboards = append(leaderboards, "wins", "playtime")

	c := make(chan struct{}, 1)
	c <- struct{}{}

	var cursor int
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case <-v.c:
			return
		case <-c:
			variant := leaderboards[cursor]

			sb := &strings.Builder{}
			sb.WriteString(text.Colourf("<bold><dark-aqua>TOP %v</dark-aqua></bold>\n", strings.ReplaceAll(strings.ToUpper(variant), "_", " ")))

			var query string
			switch variant {
			case "global":
				query = "-practice.elo"
			case "wins":
				query = "-practice.ranked_wins"
			case "playtime":
				query = "-playtime"
			default:
				query = fmt.Sprintf("-practice.game_elo.%v", variant)
			}

			leaders, err := data.OrderedOfflineUsers(query, 10)
			if err != nil {
				panic(err)
			}

			for i, leader := range leaders {
				var value any
				switch variant {
				case "global":
					value = leader.Stats.Elo
				case "wins":
					value = leader.Stats.RankedWins
				case "playtime":
					value = fmt.Sprintf("%v hours", int(math.Floor(leader.PlayTime().Hours())))
				default:
					value = leader.Stats.GameElo[variant]
				}

				position, _ := romannumeral.IntToString(i + 1)
				sb.WriteString(text.Colourf(
					"<grey>%v.</grey> <white>%v</white> <aqua>-</aqua> <grey>%v</grey>\n",
					position,
					leader.DisplayName(),
					value,
				))
			}

			cursor++
			if cursor == len(leaderboards) {
				cursor = 0
			}

			b.SetText(sb.String())
		case <-t.C:
			c <- struct{}{}
		}
	}
}
