package vasar

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/oomph-ac/oomph/check"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	_ "github.com/vasar-network/vails/command"
	_ "github.com/vasar-network/vails/console"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/tebex"
	"github.com/vasar-network/vails/worlds"
	"golang.org/x/text/language"
	"math"
	"net/netip"
	"sync"
	"time"
	_ "unsafe"
)

var (
	connectionsMu sync.Mutex
	connections   = make(map[netip.Addr]int)
)

// Vasar creates a new instance of Vasar Practice.
type Vasar struct {
	name string

	log    *logrus.Logger
	config Config

	worlds *worlds.Manager
	srv    *server.Server
	store  *tebex.Client

	mute      atomic.Bool
	pvp       atomic.Bool
	startTime time.Time

	c chan struct{}
}

// New creates a new instance of Vasar.
func New(log *logrus.Logger, config Config) (*Vasar, error) {
	config.WorldConfig = func(def world.Config) world.Config {
		def.Generator = nil
		def.PortalDestination = nil
		def.ReadOnly = true
		return def
	}

	v := &Vasar{
		name: text.Colourf("<bold><dark-aqua>VASAR</dark-aqua></bold>") + "§8",

		srv: server.New(&config.Config, log),
		c:   make(chan struct{}),

		pvp: *atomic.NewBool(true),

		log:    log,
		config: config,
	}
	v.srv.Allow(&allower{v: v})
	v.srv.CloseOnProgramEnd()
	v.srv.PlayerProvider(&Provider{srv: v.srv})
	v.srv.SetName(v.name)

	p, err := resource.Compile(config.Pack.Path)
	if err != nil {
		return nil, err
	}
	v.srv.AddResourcePack(p.WithContentKey(config.Pack.Key))

	v.loadLocales()
	v.loadTexts()
	v.loadArenas()
	v.loadStore()
	go v.startBroadcasts()
	go v.startPlayerBroadcasts()
	return v, nil
}

// Start starts the server.
func (v *Vasar) Start() error {
	if err := v.srv.Start(); err != nil {
		return err
	}
	v.startTime = time.Now()

	v.srv.World()
	w := v.srv.World()
	w.Handle(WorldHandler{})
	w.StopWeatherCycle()
	w.SetDefaultGameMode(world.GameModeAdventure)
	w.SetTime(6000)
	w.StopTime()
	w.SetTickRange(0)

	for v.srv.Accept(v.accept) {
		// Do nothing.
	}

	close(v.c)
	v.closeArenas()
	return nil
}

// CheckAlias returns the translated reason of an anti-cheat check, and if the check should ban.
func (v *Vasar) CheckAlias(l language.Tag, ch check.Check) (reason string, ban bool) {
	ban = true
	switch ch.(type) {
	case *check.AimAssistA:
		reason = lang.Translate(l, "cheat.aim.assist")
	case *check.AutoClickerA, *check.AutoClickerB, *check.AutoClickerC, *check.AutoClickerD:
		reason, ban = lang.Translate(l, "cheat.auto.clicker"), false
	case *check.InvalidMovementC:
		reason = lang.Translate(l, "cheat.invalid.movement")
	case *check.KillAuraA, *check.KillAuraB:
		reason = lang.Translate(l, "cheat.kill.aura")
	case *check.OSSpoofer:
		reason, ban = lang.Translate(l, "cheat.os.spoof"), false
	case *check.ReachA:
		reason, ban = lang.Translate(l, "cheat.reach"), false
	case *check.TimerA:
		reason, ban = lang.Translate(l, "cheat.timer"), false
	default:
		name, _ := ch.Name()
		reason = name
	}
	return
}

// ToggleGlobalMute toggles the global mute state of the chat.
func (v *Vasar) ToggleGlobalMute() (old bool) {
	return v.mute.Toggle()
}

// GlobalMuted checks if the chat is globally muted.
func (v *Vasar) GlobalMuted() bool {
	return v.mute.Load()
}

// TogglePvP toggles server-wide pvp.
func (v *Vasar) TogglePvP() (old bool) {
	return v.pvp.Toggle()
}

// PvP checks if pvp is enabled on the server.
func (v *Vasar) PvP() bool {
	return v.pvp.Load()
}

// StartTime returns the start time of the server.
func (v *Vasar) StartTime() time.Time {
	return v.startTime
}

// accept welcomes and accepts the player provided.
func (v *Vasar) accept(p *player.Player) {
	addr, _ := netip.ParseAddrPort(p.Addr().String())
	ip := addr.Addr()

	connectionsMu.Lock()
	if connections[ip] >= 5 {
		p.Disconnect(lang.Translatef(p.Locale(), "user.connections.limit"))
		connectionsMu.Unlock()
		return
	}
	connections[ip]++
	connectionsMu.Unlock()

	u, err := data.LoadUser(p)
	if err != nil {
		p.Disconnect(lang.Translatef(p.Locale(), "user.account.error"))
		return
	}
	_ = data.SaveUser(u) // Ensure the user is saved on join, in case this is their first join.

	p.Handle(newHandler(u, v))
	v.store.ExecuteCommands(p)
}

// loadLocales loads all supported locales to Vails.
func (v *Vasar) loadLocales() {
	lang.Register(language.English)
	// TODO: More languages in the future?
}

// loadStore initializes the Tebex store connection.
func (v *Vasar) loadStore() {
	v.store = tebex.NewClient(v.log, time.Second*5, v.config.Vasar.Tebex)
	name, domain, err := v.store.Information()
	if err != nil {
		v.log.Fatalf("tebex: %v", err)
	}
	v.log.Infof("Connected to Tebex under %v (%v).", name, domain)
}

// loadTexts loads all relevant floating texts to the lobby world.
func (v *Vasar) loadTexts() {
	w := v.srv.World()
	c := v.config.Vasar
	b := mgl64.Vec3{-6.5, 58.35, 47.5}
	for _, e := range []*entity.Text{
		entity.NewText(text.Colourf("<b><dark-aqua>VASAR</dark-aqua></b>"), mgl64.Vec3{b.X(), b.Y() + 2.4, b.Z()}),
		entity.NewText(text.Colourf("<purple>Season %v began on %v.</purple>", c.Season, c.Start), mgl64.Vec3{b.X(), b.Y() + 1.9, b.Z()}),
		entity.NewText(text.Colourf("<purple>It will conclude on %v.</purple>", c.End), mgl64.Vec3{b.X(), b.Y() + 1.6, b.Z()}),
		entity.NewText(text.Colourf("Store: <aqua>https://vasar.tebex.io</aqua>"), mgl64.Vec3{b.X(), b.Y() + 1.1, b.Z()}),
		entity.NewText(text.Colourf("Discord: <aqua>discord.gg/vasar</aqua>"), mgl64.Vec3{b.X(), b.Y() + 0.5, b.Z()}),
		entity.NewText(text.Colourf("<grey>vasar.land</grey>"), b),
	} {
		w.AddEntity(e)
	}
	l := world.NewLoader(6, w, world.NopViewer{})
	l.Move(w.Spawn().Vec3Middle())
	l.Load(int(math.Round(math.Pi * 36)))
	go v.startLeaderboards()
}

// loadArenas loads all default arenas.
func (v *Vasar) loadArenas() {
	lobby.NewProvider(v.srv.World())

	v.worlds = worlds.New(v.srv, "assets/arenas/ffa", v.log)
	for _, g := range game.FFA() {
		err := v.worlds.LoadWorld(g.String(), "Vasar "+g.Name())
		if err != nil {
			v.log.Fatalf("failed to load %v world: %v", g.String(), err)
		}
		w := v.worlds.AssertWorld("Vasar " + g.Name())
		w.Handle(WorldHandler{})

		ffa.NewProvider(g, w)
	}

	match.NewUnrankedProvider(v.log, true)
	match.NewRankedProvider(v.log)
	match.Unranked().World().Handle(WorldHandler{})
	match.Ranked().World().Handle(WorldHandler{})
}

// closeArenas closes all arenas.
func (v *Vasar) closeArenas() {
	err := v.worlds.Close()
	if err != nil {
		v.log.Fatalf("failed to close worlds: %v", err)
	}
}
