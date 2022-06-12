package user

import (
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/vasar-network/practice/vasar/board"
	ent "github.com/vasar-network/practice/vasar/entity"
	"github.com/vasar-network/practice/vasar/game"
	it "github.com/vasar-network/practice/vasar/item"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/sets"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"
	_ "unsafe"
)

// User is an extension of the Dragonfly player that adds a few extra features required by Vasar.
type User struct {
	p *player.Player

	hashedAddress string
	whitelisted   atomic.Bool
	address       net.Addr

	firstLogin time.Time
	joinTime   time.Time
	playTime   time.Duration

	displayName atomic.Value[string]

	lastMessageFrom atomic.Value[string]
	lastMessage     atomic.Value[time.Time]

	launchDelay         atomic.Value[time.Time]
	pearlCoolDown       atomic.Bool
	projectilesDisabled atomic.Bool

	frozen atomic.Bool

	tagMu         sync.Mutex
	tagExpiration time.Time
	attacker      *player.Player
	tagC          chan struct{}

	clickWatchersMu sync.Mutex
	clickWatchers   sets.Set[*User]
	watchingClick   *User

	rodMu sync.Mutex
	hook  *ent.FishingHook

	vanished atomic.Bool

	antiCheatAlerts    atomic.Bool
	antiCheatDelay     atomic.Value[time.Duration]
	lastAntiCheatAlert atomic.Value[time.Time]

	recentOpponent atomic.Value[string]

	settings atomic.Value[Settings]
	stats    atomic.Value[Stats]

	board atomic.Value[board.Provider]

	variantsMu sync.Mutex
	variants   []string

	postMatchStatsMu sync.Mutex
	postMatchStats   map[string]any
	queuedSince      atomic.Value[time.Time]
	reportSince      atomic.Value[time.Time]

	inAir       atomic.Bool
	airDuration atomic.Value[time.Time]

	roles     *Roles
	mute, ban atomic.Value[Punishment]

	deviceGroup atomic.Value[DeviceGroup]
	pingRange   atomic.Value[PingRange]
	eloRange    atomic.Value[EloRange]

	chatType atomic.Value[ChatType]

	s *session.Session
}

var (
	userMu    sync.Mutex
	users     = map[*player.Player]*User{}
	staff     = map[*player.Player]*User{}
	admins    = map[*player.Player]*User{}
	usersXUID = map[string]*User{}

	frozen = sets.New[string]()
)

// All returns a slice of all the users.
func All() []*User {
	userMu.Lock()
	defer userMu.Unlock()
	return maps.Values(users)
}

// Staff returns a slice of all staff online.
func Staff() []*User {
	userMu.Lock()
	defer userMu.Unlock()
	return maps.Values(staff)
}

// Admins returns a slice of all admins online.
func Admins() []*User {
	userMu.Lock()
	defer userMu.Unlock()
	return maps.Values(admins)
}

// Count returns the total user count.
func Count() int {
	userMu.Lock()
	defer userMu.Unlock()
	return len(users)
}

// Lookup looks up the user.User of a player.Player passed.
func Lookup(p *player.Player) (*User, bool) {
	userMu.Lock()
	defer userMu.Unlock()
	u, ok := users[p]
	return u, ok
}

// LookupXUID looks up the user.User of a XUID passed.
func LookupXUID(xuid string) (*User, bool) {
	userMu.Lock()
	defer userMu.Unlock()
	u, ok := usersXUID[xuid]
	return u, ok
}

// Alert alerts all staff users with an action performed by a cmd.Source.
func Alert(s cmd.Source, key string, args ...any) {
	for _, u := range Admins() {
		u.Message("staff.alert",
			s.Name(),
			fmt.Sprintf(lang.Translate(u.Player().Locale(), key), args...),
		)
	}
}

// Broadcast broadcasts a message to every user using that user's locale.
func Broadcast(key string, args ...any) {
	for _, u := range All() {
		u.Message(key, args...)
	}
}

// NewUser creates a new user from a Dragonfly player along with list of roles and settings.
func NewUser(p *player.Player, r *Roles, settings Settings, stats Stats, firstLogin time.Time, playTime time.Duration, hashedAddress string, whitelisted bool, variants []string, mute, ban Punishment) *User {
	u := &User{
		p:             p,
		address:       p.Addr(),
		whitelisted:   *atomic.NewBool(whitelisted),
		hashedAddress: hashedAddress,
		s:             player_session(p),

		joinTime:   time.Now(),
		firstLogin: firstLogin,
		playTime:   playTime,

		tagC: make(chan struct{}, 1),

		roles: r,
		mute:  *atomic.NewValue(mute),
		ban:   *atomic.NewValue(ban),

		variants:      variants,
		clickWatchers: sets.New[*User](),

		antiCheatAlerts: *atomic.NewBool(true),

		settings:    *atomic.NewValue(settings),
		stats:       *atomic.NewValue(stats),
		deviceGroup: *atomic.NewValue(DeviceGroupKeyboardMouse()),
	}
	u.displayName.Store(p.Name())
	if u.s != session.Nop {
		f := reflect.ValueOf(u.s).Elem().FieldByName("handlers")

		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
		f.SetMapIndex(reflect.ValueOf(uint32(packet.IDPlayerAuthInput)), reflect.ValueOf(PlayerAuthInputHandler{u: u}))
	}
	u.roles.sortRoles()
	u.SetNameTagFromRole()

	userMu.Lock()
	users[p] = u
	if u.roles.Staff() {
		staff[p] = u
		for _, s := range staff {
			l := s.Player().Locale()
			s.Player().Message(lang.Translatef(l,
				"staff.joined",
				cases.Title(l).String(u.Roles().Highest().Name()),
				u.Player().Name(),
			))
		}
	}
	if u.roles.Contains(role.Admin{}, role.Operator{}) {
		admins[p] = u
	}
	usersXUID[p.XUID()] = u
	if frozen.Contains(p.XUID()) {
		u.p.SetImmobile()
		u.frozen.Toggle()
	}
	userMu.Unlock()
	return u
}

// DisplayName returns the display name of the user.
func (u *User) DisplayName() string {
	return u.displayName.Load()
}

// SetDisplayName sets the display name of the user.
func (u *User) SetDisplayName(name string) {
	u.displayName.Store(name)
}

// Vanished returns whether the user is vanished or not.
func (u *User) Vanished() bool {
	return u.vanished.Load()
}

// ToggleVanish toggles the user's vanish state.
func (u *User) ToggleVanish() {
	u.vanished.Toggle()
}

// AntiCheatAlerts returns whether the user has anti-cheat alerts enabled or not.
func (u *User) AntiCheatAlerts() bool {
	return u.antiCheatAlerts.Load()
}

// ToggleAntiCheatAlerts toggles the user's anti-cheat alerts.
func (u *User) ToggleAntiCheatAlerts() (old bool) {
	return u.antiCheatAlerts.Toggle()
}

// AntiCheatAlertDelay returns the user's anti-cheat alert delay.
func (u *User) AntiCheatAlertDelay() time.Duration {
	return u.antiCheatDelay.Load()
}

// LastAntiCheatAlert returns the timestamp of the last received anti-cheat alert.
func (u *User) LastAntiCheatAlert() time.Time {
	return u.lastAntiCheatAlert.Load()
}

// RenewLastAntiCheatAlert renews the last time the user received an anti-cheat alert.
func (u *User) RenewLastAntiCheatAlert() {
	u.lastAntiCheatAlert.Store(time.Now())
}

// SetAntiCheatAlertDelay sets the user's anti-cheat alert delay.
func (u *User) SetAntiCheatAlertDelay(delay time.Duration) {
	u.antiCheatDelay.Store(delay)
}

// Potions returns the amount of potions the user has.
func (u *User) Potions() (n int) {
	for _, i := range u.p.Inventory().Items() {
		if p, ok := i.Item().(it.VasarPotion); ok && p.Type == potion.StrongHealing() {
			n++
		}
	}
	return n
}

// Message sends a localized message to the user, implementing chat.Subscriber.
func (u *User) Message(key string, args ...any) {
	u.Player().Message(lang.Translatef(u.Player().Locale(), key, args...))
}

// DesiredPingRange returns the ping range that opponents of the user must be within.
func (u *User) DesiredPingRange() PingRange {
	return PingRanges()[u.Settings().Matchmaking.PingRange]
}

// RenewEloRange renews the elo range of the user, resetting it to the default.
func (u *User) RenewEloRange(g game.Game) {
	u.eloRange.Store(NewEloRange(int(u.Stats().GameElo[g.String()])))
}

// RenewPingRange renews the ping range of the user, resetting it to the default.
func (u *User) RenewPingRange() {
	if u.DesiredPingRange() == PingRangeUnrestricted() {
		u.pingRange.Store(PingRangeUnrestricted())
		return
	}
	u.pingRange.Store(NewPingRange(int(u.Latency())))
}

// RenewLastMessage renews the last time a message was sent from a player.
func (u *User) RenewLastMessage() {
	u.lastMessage.Store(time.Now())
}

// RenewAirDuration renews the last time the user was in the air.
func (u *User) RenewAirDuration() {
	u.airDuration.Store(time.Now())
	u.inAir.Toggle()
}

// ResetAirDuration resets the last time the user was in the air.
func (u *User) ResetAirDuration() {
	u.inAir.Toggle()
}

// UpdateEloRange updates the elo range of the user.
func (u *User) UpdateEloRange(ra EloRange) {
	u.eloRange.Store(ra)
}

// EloRange returns the elo range the user is currently within.
func (u *User) EloRange() EloRange {
	return u.eloRange.Load()
}

// UpdatePingRange updates the ping range of the user.
func (u *User) UpdatePingRange(ra PingRange) {
	u.pingRange.Store(ra)
}

// PingRange returns the ping range the user is currently within.
func (u *User) PingRange() PingRange {
	return u.pingRange.Load()
}

// UpdateChatType updates the chat type for the user.
func (u *User) UpdateChatType(t ChatType) {
	u.chatType.Store(t)
}

// ChatType returns the chat type the user is currently using.
func (u *User) ChatType() ChatType {
	return u.chatType.Load()
}

// DesiredDeviceGroups returns the device groups that the user will match against.
func (u *User) DesiredDeviceGroups() []DeviceGroup {
	settings := u.Settings()
	groups := make([]DeviceGroup, 0, 3)
	if settings.Matchmaking.MatchWithMobile {
		groups = append(groups, DeviceGroupMobile())
	}
	if settings.Matchmaking.MatchWithController {
		groups = append(groups, DeviceGroupController())
	}
	if settings.Matchmaking.MatchWithKeyboard {
		groups = append(groups, DeviceGroupKeyboardMouse())
	}
	return groups
}

// CanSendMessage returns true if the user can send a message.
func (u *User) CanSendMessage() bool {
	if u.Roles().Contains(role.Operator{}) {
		return true
	}
	if _, ok := u.Mute(); ok {
		return false
	}
	return time.Since(u.lastMessage.Load()) > time.Second*2
}

// AirDuration returns the time the user has been in the air.
func (u *User) AirDuration() (time.Duration, bool) {
	return time.Since(u.airDuration.Load()), u.inAir.Load()
}

// StartWatchingClicks starts watchingClick the user's clicks.
func (u *User) StartWatchingClicks(user *User) {
	u.clickWatchersMu.Lock()
	if u.watchingClick != nil {
		u.watchingClick.RemoveClickWatcher(u)
	}
	u.watchingClick = user
	u.clickWatchersMu.Unlock()
	user.AddClickWatcher(u)
}

// StopWatchingClicks stops watchingClick the user's clicks.
func (u *User) StopWatchingClicks() {
	u.clickWatchersMu.Lock()
	if u.watchingClick == nil {
		u.clickWatchersMu.Unlock()
		return
	}
	user := u.watchingClick
	u.watchingClick = nil
	u.clickWatchersMu.Unlock()
	user.RemoveClickWatcher(u)
}

// Variant returns true if the user has the given Vasar+ variant.
func (u *User) Variant(variant string) bool {
	u.variantsMu.Lock()
	defer u.variantsMu.Unlock()
	return slices.Contains(u.variants, variant)
}

// UnlockVariant unlocks the given variant for the user.
func (u *User) UnlockVariant(variant string) {
	u.variantsMu.Lock()
	defer u.variantsMu.Unlock()
	u.variants = append(u.variants, variant)
}

// LockVariant locks the given variant for the user.
func (u *User) LockVariant(variant string) {
	u.variantsMu.Lock()
	defer u.variantsMu.Unlock()
	ind := slices.Index(u.variants, variant)
	u.variants = slices.Delete(u.variants, ind, ind+1)
}

// Variants returns the variants the user has unlocked.
func (u *User) Variants() []string {
	u.variantsMu.Lock()
	defer u.variantsMu.Unlock()
	return u.variants
}

// AddClickWatcher adds a user to the clicksWatchers set.
func (u *User) AddClickWatcher(user *User) {
	u.clickWatchersMu.Lock()
	defer u.clickWatchersMu.Unlock()
	u.clickWatchers.Add(user)
}

// RemoveClickWatcher removes a user from the clicksWatchers set.
func (u *User) RemoveClickWatcher(user *User) {
	u.clickWatchersMu.Lock()
	defer u.clickWatchersMu.Unlock()
	u.clickWatchers.Delete(user)
}

// ClickWatchers returns the users watchingClick the user.
func (u *User) ClickWatchers() (users []*User) {
	u.clickWatchersMu.Lock()
	for usr := range u.clickWatchers {
		users = append(users, usr)
	}
	u.clickWatchersMu.Unlock()
	return users
}

// WatchingClicks returns the user it is currently watching clicks.
func (u *User) WatchingClicks() (*User, bool) {
	u.clickWatchersMu.Lock()
	defer u.clickWatchersMu.Unlock()
	return u.watchingClick, u.watchingClick != nil
}

// SetRecentOpponent sets the opponent that the user last matched against.
func (u *User) SetRecentOpponent(opponent *User) {
	u.recentOpponent.Store(opponent.Player().XUID())
}

// ResetRecentOpponent resets the opponent that the user last matched against.
func (u *User) ResetRecentOpponent() {
	u.recentOpponent.Store("")
}

// RecentOpponent returns the opponent that the user last matched against.
func (u *User) RecentOpponent() (*User, bool) {
	return LookupXUID(u.recentOpponent.Load())
}

// RenewQueuedSince renews the last time the user has been queued.
func (u *User) RenewQueuedSince() {
	u.queuedSince.Store(time.Now())
}

// RenewReportSince renews the last time the user has made a report.
func (u *User) RenewReportSince() {
	u.reportSince.Store(time.Now())
}

// ReportSince returns the last time the user has made a report.
func (u *User) ReportSince() time.Time {
	return u.reportSince.Load()
}

// QueuedSince returns the last time the user has been queued.
func (u *User) QueuedSince() time.Time {
	return u.queuedSince.Load()
}

// Latency returns the full round trip latency of the user in milliseconds.
func (u *User) Latency() int64 {
	return u.Player().Latency().Milliseconds() * 2
}

// DeviceGroup returns the device group of the user.
func (u *User) DeviceGroup() DeviceGroup {
	return u.deviceGroup.Load()
}

// Whitelisted returns true if the user is whitelisted.
func (u *User) Whitelisted() bool {
	return u.whitelisted.Load()
}

// Whitelist adds the user to the whitelist.
func (u *User) Whitelist() {
	u.whitelisted.Store(true)
}

// Unwhitelist removes the user from the whitelist.
func (u *User) Unwhitelist() {
	u.whitelisted.Store(false)
}

// FirstLogin returns the time the user first logged in.
func (u *User) FirstLogin() time.Time {
	return u.firstLogin
}

// JoinTime returns the time the user joined.
func (u *User) JoinTime() time.Time {
	return u.joinTime
}

// PlayTime returns the time the user has played.
func (u *User) PlayTime() time.Duration {
	return u.playTime + time.Since(u.joinTime)
}

// Roles returns the role manager of the user.
func (u *User) Roles() *Roles {
	return u.roles
}

// SetBan sets the ban data of the user.
func (u *User) SetBan(p Punishment) {
	u.ban.Store(p)
}

// Ban returns the ban data of the user, this should only be valid once, when the user gets banned.
func (u *User) Ban() Punishment {
	return u.ban.Load()
}

// SetMute sets the mute data of the user.
func (u *User) SetMute(p Punishment) {
	u.mute.Store(p)
}

// Mute returns the mute data of the user and true if the data is valid. Otherwise, it will return false.
func (u *User) Mute() (Punishment, bool) {
	p := u.mute.Load()
	if p.Expired() {
		u.SetMute(Punishment{})
		return Punishment{}, false
	}
	return p, true
}

// SetNameTagFromRole sets the name tag from the user's highest role.
func (u *User) SetNameTagFromRole() {
	if u.DisplayName() != u.Player().Name() {
		u.Player().SetNameTag(role.Default{}.Tag(u.DisplayName()))
		return
	}
	highest := u.Roles().Highest()
	tag := highest.Tag(u.DisplayName())
	if _, ok := highest.(role.Plus); ok {
		tag = strings.ReplaceAll(tag, "§0", u.Settings().Advanced.VasarPlusColour)
	}
	u.Player().SetNameTag(tag)
}

// MultiplyParticles multiplies the hit particles for the user.
func (u *User) MultiplyParticles(e world.Entity, multiplier int) {
	for i := 0; i < multiplier; i++ {
		u.s.ViewEntityAction(e, entity.CriticalHitAction{})
	}
}

// Rotate rotates the user with the specified yaw and pitch deltas.
// TODO: Remove this once Dragonfly supports a way to do this properly.
func (u *User) Rotate(deltaYaw, deltaPitch float64) {
	currentYaw, currentPitch := u.p.Rotation()
	session_writePacket(u.s, &packet.MovePlayer{
		EntityRuntimeID: 1, // Always 1 on Dragonfly.
		Position:        vec64To32(u.p.Position().Add(mgl64.Vec3{0, 1.62})),
		Pitch:           float32(currentPitch + deltaPitch),
		Yaw:             float32(currentYaw + deltaYaw),
		HeadYaw:         float32(currentYaw + deltaYaw),
		Mode:            packet.MoveModeTeleport,
		OnGround:        u.p.OnGround(),
	})
	u.p.Move(mgl64.Vec3{}, deltaYaw, deltaPitch)
}

// Launch launches the user in their direction vector.
func (u *User) Launch() {
	now := time.Now()
	if now.Before(u.launchDelay.Load()) {
		return
	}

	u.SendCustomParticle(8, 0, u.p.Position(), true) // Add a flame particle.
	u.SendCustomSound("mob.vex.hurt", 1, 0.5, true)

	motion := entity.DirectionVector(u.p).Mul(1.5)
	motion[1] = 0.85

	u.p.StopSprinting()
	u.p.SetVelocity(motion)

	u.launchDelay.Store(now.Add(time.Second * 2))
}

// SetLastMessageFrom sets the player passed as the last player who messaged the user.
func (u *User) SetLastMessageFrom(p *player.Player) {
	u.lastMessageFrom.Store(p.XUID())
}

// LastMessageFrom returns the last user that messaged the user.
func (u *User) LastMessageFrom() (*User, bool) {
	u, ok := LookupXUID(u.lastMessageFrom.Load())
	return u, ok
}

// ToggleFreeze toggles the frozen state of the user.
func (u *User) ToggleFreeze() {
	userMu.Lock()
	defer userMu.Unlock()
	if u.frozen.Toggle() {
		u.p.SetMobile()
		frozen.Delete(u.p.XUID())
		return
	}
	u.p.SetImmobile()
	frozen.Add(u.p.XUID())
}

// Frozen returns the frozen state of the user.
func (u *User) Frozen() bool { return u.frozen.Load() }

// ToggleRod toggles a fishing hook. If the user is already using a hook, it will be removed, otherwise a new hook will
// be created.
func (u *User) ToggleRod() {
	u.rodMu.Lock()
	defer u.rodMu.Unlock()

	if u.hook == nil || u.hook.World() == nil {
		if u.ProjectilesDisabled() {
			u.Message("projectiles.disabled")
			return
		}
		u.hook = ent.NewFishingHook(entity.EyePosition(u.p), entity.DirectionVector(u.p).Mul(1.3), u.p)

		w := u.p.World()
		w.AddEntity(u.hook)
	} else {
		_ = u.hook.Close()
	}
}

// PearlCoolDown returns true if ender pearls currently are on cool down.
func (u *User) PearlCoolDown() bool {
	return u.pearlCoolDown.Load()
}

// TogglePearlCoolDown toggles the ender pearl cool down.
func (u *User) TogglePearlCoolDown() {
	if u.pearlCoolDown.Toggle() {
		u.ResetExperienceProgress()
	}
}

// DisableProjectiles disables the user's projectiles.
func (u *User) DisableProjectiles() {
	u.projectilesDisabled.Store(true)
}

// EnableProjectiles enables the user's projectiles.
func (u *User) EnableProjectiles() {
	u.projectilesDisabled.Store(false)
}

// ProjectilesDisabled returns true if the user's projectiles are disabled.
func (u *User) ProjectilesDisabled() bool {
	return u.projectilesDisabled.Load()
}

// Settings returns the settings of the user.
func (u *User) Settings() Settings {
	return u.settings.Load()
}

// SetSettings sets the settings of the user.
func (u *User) SetSettings(settings Settings) {
	u.settings.Store(settings)
}

// Stats returns the stats of the user.
func (u *User) Stats() Stats {
	return u.stats.Load()
}

// SetStats sets the stats of the user.
func (u *User) SetStats(stats Stats) {
	u.stats.Store(stats)
}

// SetPostMatchStats sets the post match stats of the user.
func (u *User) SetPostMatchStats(stats map[string]any) {
	u.postMatchStatsMu.Lock()
	defer u.postMatchStatsMu.Unlock()
	u.postMatchStats = stats
}

// PostMatchStats returns the post match stats of the user.
func (u *User) PostMatchStats() (map[string]any, bool) {
	u.postMatchStatsMu.Lock()
	defer u.postMatchStatsMu.Unlock()
	return u.postMatchStats, u.postMatchStats != nil
}

// SetBoard stores the current board provider of the user.
func (u *User) SetBoard(board board.Provider) {
	u.board.Store(board)
	if !u.Settings().Display.Scoreboard {
		// If the scoreboard is disabled, we shouldn't do anything.
		return
	}
	p := u.Player()
	p.RemoveScoreboard()
	board.SendScoreboard(p)
}

// Board returns the current scoreboard provider of the user.
func (u *User) Board() board.Provider {
	if b := u.board.Load(); b != nil && u.Settings().Display.Scoreboard {
		return b
	}
	return board.NopProvider{}
}

// Player ...
func (u *User) Player() *player.Player {
	return u.p
}

// Device returns the device of the user.
func (u *User) Device() string {
	return u.s.ClientData().DeviceModel
}

// DeviceID returns the device ID of the user.
func (u *User) DeviceID() string {
	return u.s.ClientData().DeviceID
}

// SelfSignedID returns the self-signed ID of the user.
func (u *User) SelfSignedID() string {
	return u.s.ClientData().SelfSignedID
}

// Address returns the address of the user.
func (u *User) Address() net.Addr {
	return u.address
}

// HashedAddress returns the hashed IP address of the user.
func (u *User) HashedAddress() string {
	return u.hashedAddress
}

// Close ...
func (u *User) Close() {
	if u.Roles().Staff() {
		for _, s := range staff {
			l := s.Player().Locale()
			s.Player().Message(lang.Translatef(l,
				"staff.left",
				cases.Title(l).String(u.Roles().Highest().Name()),
				u.Player().Name(),
			))
		}
	}

	u.tagMu.Lock()
	close(u.tagC)
	u.tagMu.Unlock()

	userMu.Lock()
	delete(users, u.p)
	delete(staff, u.p)
	delete(admins, u.p)
	delete(usersXUID, u.p.XUID())
	userMu.Unlock()
}

// viewers returns a list of all viewers of the Player.
func (u *User) viewers() []world.Viewer {
	viewers := u.p.World().Viewers(u.p.Position())
	for _, v := range viewers {
		if v == u.s {
			return viewers
		}
	}
	return append(viewers, u.s)
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

//go:linkname player_session github.com/df-mc/dragonfly/server/player.(*Player).session
//noinspection ALL
func player_session(*player.Player) *session.Session

//go:linkname session_writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
//noinspection ALL
func session_writePacket(*session.Session, packet.Packet)
