package main

import (
	"github.com/RestartFU/gophig"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"github.com/vasar-network/practice/vasar"
	"github.com/vasar-network/practice/vasar/command"
	_ "github.com/vasar-network/vails/command"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.InfoLevel

	conf := vasar.DefaultConfig()
	g := gophig.NewGophig("./config", "toml", 0777)
	if err := g.GetConf(&conf); os.IsNotExist(err) {
		err = g.SetConf(conf)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Release: conf.Sentry.Release,
		Dsn:     conf.Sentry.Dsn,
	}); err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
			panic(err)
		}
	}()

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:1111", nil))
	}()

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	newDiamond := item.ArmourTierDiamond
	newDiamond.Toughness = math.MaxFloat64
	item.ArmourTierDiamond = newDiamond

	newIron := item.ArmourTierIron
	newIron.Toughness = math.MaxFloat64
	item.ArmourTierIron = newIron

	v, err := vasar.New(log, conf)
	if err != nil {
		log.Fatalln(err)
	}

	cmd.Register(cmd.New("alias", "", nil, command.AliasOnline{}, command.AliasOffline{}))
	cmd.Register(cmd.New("announce", "Announce a message to the entire server.", nil, command.Announce{}))
	cmd.Register(cmd.New("ban", "", nil, command.BanLiftOffline{}, command.BanList{}, command.BanInfoOffline{}, command.BanForm{}, command.Ban{}, command.BanOffline{}))
	cmd.Register(cmd.New("blacklist", "", nil, command.BlacklistLiftOffline{}, command.BlacklistList{}, command.BlacklistInfoOffline{}, command.BlacklistForm{}, command.Blacklist{}, command.BlacklistOffline{}))
	cmd.Register(cmd.New("chat", "Switch between your available chats.", nil, command.Chat{}, command.ChatSwitch{}))
	cmd.Register(cmd.New("cps", "View a players' clicks-per-second.", nil, command.CPS{}))
	cmd.Register(cmd.New("disguise", "Manage disguises.", nil, command.Disguise{}))
	cmd.Register(cmd.New("duel", "Duel a player.", nil, command.DuelRequest{}, command.DuelRespond{}))
	cmd.Register(cmd.New("entity", "", nil, command.Entity{}))
	cmd.Register(cmd.New("ffa", "Manage FFA arenas.", nil, command.FFA{}))
	cmd.Register(cmd.New("fly", "Manage flight.", nil, command.Fly{}))
	cmd.Register(cmd.New("freeze", "Freeze a player.", []string{"ss"}, command.Freeze{}))
	cmd.Register(cmd.New("gamemode", "Change the gamemode of yourself or others.", []string{"gm"}, command.GameMode{}))
	cmd.Register(cmd.New("globalmute", "", nil, command.NewGlobalMute(v)))
	cmd.Register(cmd.New("kick", "Remove a player from the server.", nil, command.Kick{}))
	cmd.Register(cmd.New("kill", "", nil, command.Kill{}))
	cmd.Register(cmd.New("mute", "", nil, command.MuteList{}, command.MuteInfo{}, command.MuteInfoOffline{}, command.MuteLift{}, command.MuteLiftOffline{}, command.MuteForm{}, command.Mute{}, command.MuteOffline{}))
	cmd.Register(cmd.New("nick", "", nil, command.NickReset{}, command.Nick{}))
	cmd.Register(cmd.New("online", "", nil, command.Online{}))
	cmd.Register(cmd.New("pvp", "Modify the damage by players per arena.", nil, command.NewPvP(v)))
	cmd.Register(cmd.New("rekit", "", []string{"kit"}, command.Rekit{}))
	cmd.Register(cmd.New("reply", "", []string{"r"}, command.Reply{}))
	cmd.Register(cmd.New("report", "Report a player for breaking the rules.", nil, command.Report{}))
	cmd.Register(cmd.New("reset", "", nil, command.ResetElo{}, command.ResetEloOffline{}, command.ResetStats{}, command.ResetStatsOffline{}, command.ResetSeason{}))
	cmd.Register(cmd.New("restart", "Restart the server. (DO NOT USE)", nil, command.Restart{}))
	cmd.Register(cmd.New("role", "", nil, command.RoleAdd{}, command.RoleRemove{}, command.RoleAddOffline{}, command.RoleRemoveOffline{}, command.RoleList{}))
	cmd.Register(cmd.New("settings", "Manage Settings.", nil, command.Settings{}))
	cmd.Register(cmd.New("spawn", "Teleport to the server lobby.", nil, command.Spawn{}))
	cmd.Register(cmd.New("spectate", "", nil, command.Spectate{}))
	cmd.Register(cmd.New("stats", "", nil, command.Stats{}, command.StatsOffline{}))
	cmd.Register(cmd.New("teleport", "Teleport yourself or another player to a position.", []string{"tp"}, command.TeleportToPos{}, command.TeleportToTarget{}, command.TeleportTargetsToPos{}, command.TeleportTargetsToTarget{}))
	cmd.Register(cmd.New("vanish", "", nil, command.Vanish{}))
	cmd.Register(cmd.New("variant", "Manage players' accessible Vasar+ variants.", nil, command.VariantUnlock{}, command.VariantUnlockOffline{}, command.VariantLock{}, command.VariantLockOffline{}))
	cmd.Register(cmd.New("w", "Send a private message to a player.", []string{"msg", "tell"}, command.Whisper{}))
	cmd.Register(cmd.New("whitelist", "Manage the server whitelist.", nil, command.Whitelist{}, command.WhitelistAdd{}, command.WhitelistAddOffline{}, command.WhitelistRemove{}, command.WhitelistRemoveOffline{}, command.WhitelistClear{}))
	cmd.Register(cmd.New("who", "", nil, command.Who{}))

	if err = v.Start(); err != nil {
		log.Fatalln(err)
	}
}
