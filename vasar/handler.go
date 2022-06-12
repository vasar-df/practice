package vasar

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/module"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/role"
	"golang.org/x/text/cases"
	"net/netip"
	"regexp"
	"strings"
	"time"
)

// handler is a base handler that forwards all events to their respective modules.
// TODO: Merge this with user.User or do something similar.
type handler struct {
	player.NopHandler

	u   *user.User
	srv *Vasar

	cl *module.Click
	co *module.Combat
	i  *module.Inventory
	g  *module.Game
	r  *module.Rods
	p  *module.Protection
	s  *module.Settings
	m  *module.Moderation
}

var (
	// tlds is a list of top level domains used for checking for advertisements.
	tlds = [...]string{".me", ".club", "www.", ".com", ".net", ".gg", ".cc", ".net", ".co", ".co.uk", ".ddns", ".ddns.net", ".cf", ".live", ".ml", ".gov", "http://", "https://", ",club", "www,", ",com", ",cc", ",net", ",gg", ",co", ",couk", ",ddns", ",ddns.net", ",cf", ",live", ",ml", ",gov", ",http://", "https://", "gg/"}
	// emojis is a map between emojis and their unicode representation.
	emojis = map[string]string{
		":l:":     "\uE107",
		":skull:": "\uE105",
		":fire:":  "\uE108",
		":eyes:":  "\uE109",
		":clown:": "\uE10A",
		":100:":   "\uE10B",
		":heart:": "\uE10C",
	}
)

// newHandler ...
func newHandler(u *user.User, v *Vasar) *handler {
	p := u.Player()
	ha := &handler{
		srv: v,
		u:   u,

		cl: module.NewClick(u),
		co: module.NewCombat(u),
		g:  module.NewGame(u),
		i:  module.NewInventory(u),
		m:  module.NewModeration(u),
		p:  module.NewProtection(p),
		r:  module.NewRods(u),
		s:  module.NewSettings(u),
	}
	lobby.Lobby().AddPlayer(p)
	ha.m.HandleJoin()
	ha.s.HandleJoin()

	s := p.Skin()
	if s.Persona {
		p.SetSkin(steve)
	} else if percent, err := searchTransparency(s); err != nil || percent >= 0.05 {
		p.SetSkin(steve)
	} else if !bytes.Equal(s.Model, steve.Model) {
		s.Model = steve.Model
		p.SetSkin(s)
	}
	return ha
}

// formatRegex is a regex used to clean color formatting on a string.
var formatRegex = regexp.MustCompile(`§[\da-gk-or]`)

// HandleChat ...
func (h *handler) HandleChat(ctx *event.Context, message *string) {
	ctx.Cancel()

	p := h.u.Player()
	operator := h.u.Roles().Contains(role.Operator{})
	if h.srv.GlobalMuted() && !operator {
		h.u.Message("user.globalmuted")
		return
	}

	if msg := strings.TrimSpace(*message); len(msg) > 0 {
		msg = formatRegex.ReplaceAllString(msg, "") // Remove color formatting.
		for _, word := range strings.Split(msg, " ") {
			if emoji, ok := emojis[strings.ToLower(word)]; ok {
				msg = strings.ReplaceAll(msg, word, emoji)
			}
		}
		if h.u.ChatType() == user.ChatTypeStaff() || msg[0] == '!' && h.u.Roles().Staff() {
			userName, roleName := h.u.Player().Name(), h.u.Roles().Highest().Name()
			msg = strings.TrimPrefix(msg, "!")
			for _, s := range user.Staff() {
				s.Message("staff.chat",
					cases.Title(s.Player().Locale()).String(roleName),
					userName,
					msg,
				)
			}
			return
		}

		formatted := role.Default{}.Chat(h.u.DisplayName(), msg)
		if h.u.DisplayName() == h.u.Player().Name() {
			formatted = h.u.Roles().Highest().Chat(h.u.DisplayName(), msg)
			if _, ok := h.u.Roles().Highest().(role.Plus); ok {
				formatted = strings.ReplaceAll(formatted, "§0", h.u.Settings().Advanced.VasarPlusColour)
			}
		}
		if !h.u.CanSendMessage() && !operator {
			p.Message(formatted)
			return
		}
		if !operator {
			for _, tld := range tlds {
				if strings.Contains(strings.ToLower(msg), tld) {
					p.Message(formatted)
					return
				}
			}
		}
		_, _ = chat.Global.WriteString(formatted)
		h.u.RenewLastMessage()
	}
}

// HandleSkinChange ...
func (h *handler) HandleSkinChange(ctx *event.Context, s *skin.Skin) {
	if s.Persona {
		*s = steve
	} else if percent, err := searchTransparency(*s); err != nil || percent >= 0.05 {
		*s = steve
	} else if !bytes.Equal(s.Model, steve.Model) {
		s.Model = steve.Model
	}
	h.s.HandleSkinChange(ctx, s)
}

// HandleCommandExecution ...
func (h *handler) HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string) {
	h.m.HandleCommandExecution(ctx, command, args)
	if _, ok := ffa.LookupProvider(h.u.Player()); ok {
		h.co.HandleCommandExecution(ctx, command, args)
	}
}

// HandleFoodLoss ...
func (h *handler) HandleFoodLoss(ctx *event.Context, from, to int) {
	h.p.HandleFoodLoss(ctx, from, to)
}

// HandleHurt ...
func (h *handler) HandleHurt(ctx *event.Context, dmg *float64, immunity *time.Duration, src damage.Source) {
	h.m.HandleHurt(ctx, dmg, immunity, src)
	h.p.HandleHurt(ctx, dmg, immunity, src)
	h.g.HandleHurt(ctx, dmg, immunity, src)
	h.co.HandleHurt(ctx, dmg, immunity, src)

	if (h.u.Player().Health()-h.u.Player().FinalDamageFrom(*dmg, src) <= 0 || (src == damage.SourceVoid{})) && !ctx.Cancelled() {
		ctx.Cancel()
		h.HandleDeath(src)
	}
}

// HandleDeath ...
func (h *handler) HandleDeath(source damage.Source) {
	h.g.HandleDeath(source)
	h.s.HandleDeath(source)
	h.co.HandleDeath(source)
}

// HandleBlockPlace ...
func (h *handler) HandleBlockPlace(ctx *event.Context, pos cube.Pos, b world.Block) {
	h.p.HandleBlockPlace(ctx, pos, b)
}

// HandleBlockBreak ...
func (h *handler) HandleBlockBreak(ctx *event.Context, pos cube.Pos, drops *[]item.Stack) {
	h.p.HandleBlockBreak(ctx, pos, drops)
}

// HandleItemUseOnBlock ...
func (h *handler) HandleItemUseOnBlock(ctx *event.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	h.i.HandleItemUseOnBlock(ctx, pos, face, clickPos)

	p := h.u.Player()
	held, _ := p.HeldItems()
	if h.u.DeviceGroup() == user.DeviceGroupMobile() {
		if _, ok := held.Item().(item.Usable); ok {
			p.UseItem()
		}
	}
	if _, ok := lobby.LookupProvider(p); ok {
		ctx.Cancel()
	} else if _, ok := ffa.LookupProvider(p); ok { // todo: remove this once df adds a way to access ui inventories
		if _, ok := held.Item().(item.Bucket); ok {
			ctx.Cancel()
		}
	}
}

// HandleItemUse ...
func (h *handler) HandleItemUse(ctx *event.Context) {
	if _, ok := ffa.LookupProvider(h.u.Player()); ok { // todo: remove this once df adds a way to access ui inventories
		held, _ := h.u.Player().HeldItems()
		if _, ok := held.Item().(item.SplashPotion); ok {
			ctx.Cancel()
		}
	}
	h.m.HandleItemUse(ctx)
	h.i.HandleItemUse(ctx)
	h.g.HandleItemUse(ctx)
	h.r.HandleItemUse(ctx)
}

// HandleItemConsume ...
func (h *handler) HandleItemConsume(ctx *event.Context, i item.Stack) {
	if _, ok := ffa.LookupProvider(h.u.Player()); ok { // todo: remove this once df adds a way to access ui inventories
		i := i.Item()
		if _, ok := i.(item.EnchantedApple); ok {
			ctx.Cancel()
		}
		if _, ok := i.(item.GoldenApple); ok {
			ctx.Cancel()
		}
	}
}

// HandleAttackEntity ...
func (h *handler) HandleAttackEntity(ctx *event.Context, e world.Entity, force, height *float64, critical *bool) {
	h.m.HandleAttackEntity(ctx, e, force, height, critical)
	h.cl.HandleAttackEntity(ctx, e, force, height, critical)
	h.p.HandleAttackEntity(ctx, e, force, height, critical)
	if !h.srv.PvP() {
		h.u.Message("pvp.disabled")
		ctx.Cancel()
	}
	h.g.HandleAttackEntity(ctx, e, force, height, critical)
	h.s.HandleAttackEntity(ctx, e, force, height, critical)
}

// HandlePunchAir ...
func (h *handler) HandlePunchAir(ctx *event.Context) {
	h.cl.HandlePunchAir(ctx)
	p := h.u.Player()
	if p.World() == h.srv.srv.World() && !p.OnGround() {
		h.u.Launch()
	}
	if h.u.DeviceGroup() == user.DeviceGroupMobile() {
		held, _ := p.HeldItems()
		if _, ok := held.Item().(item.Usable); ok {
			p.UseItem()
		}
	}
}

// HandleMove ...
func (h *handler) HandleMove(ctx *event.Context, newPos mgl64.Vec3, newYaw, newPitch float64) {
	h.g.HandleMove(ctx, newPos, newYaw, newPitch)
	h.s.HandleMove(ctx, newPos, newYaw, newPitch)
}

// HandleItemDrop ...
func (h *handler) HandleItemDrop(ctx *event.Context, e *entity.Item) {
	h.g.HandleItemDrop(ctx, e)
}

// HandleItemDamage ...
func (h *handler) HandleItemDamage(ctx *event.Context, i item.Stack, d int) {
	h.g.HandleItemDamage(ctx, i, d)
}

// HandleQuit ...
func (h *handler) HandleQuit() {
	h.u.StopWatchingClicks()
	for _, w := range h.u.ClickWatchers() {
		w.StopWatchingClicks()
	}

	addr, _ := netip.ParseAddrPort(h.u.Address().String())
	ip := addr.Addr()

	connectionsMu.Lock()
	if connections[ip] <= 1 {
		delete(connections, ip)
	} else {
		connections[ip]--
	}
	connectionsMu.Unlock()

	if h.u.Tagged() {
		attacker := h.u.Attacker()
		u, ok := user.Lookup(attacker)
		if ok && u.Settings().Gameplay.AutoReapplyKit {
			if prov, ok := ffa.LookupProvider(attacker); ok {
				kit.Apply(prov.Game().Kit(true), attacker)
			}
		}
		h.g.Death(attacker.World())
	}
	h.g.HandleQuit()
	h.u.Close()

	_ = data.SaveUser(h.u)

	for _, otherP := range lobby.Lobby().Players() {
		if otherU, ok := user.Lookup(otherP); ok {
			otherU.Board().SendScoreboard(otherP)
		}
	}
}
