package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"golang.org/x/exp/slices"
	"strings"
)

// VariantUnlock unlocks a Vasar+ variant for a given user.
type VariantUnlock struct {
	Sub     unlockVariant
	Targets []cmd.Target `cmd:"target"`
	Variant variantType  `cmd:"variant"`
}

// Run ...
func (v VariantUnlock) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(v.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := v.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	variant := strings.ToLower(string(v.Variant))
	if u.Variant(variant) {
		o.Error(lang.Translatef(l, "command.variant.unlocked.already", u.DisplayName(), variant))
		return
	}
	u.UnlockVariant(variant)
	o.Print(lang.Translatef(l, "command.variant.unlocked", variant, u.DisplayName()))
}

// VariantUnlockOffline unlocks a Vasar+ variant for an offline user.
type VariantUnlockOffline struct {
	Sub     unlockVariant
	Target  string      `cmd:"target"`
	Variant variantType `cmd:"variant"`
}

// Run ...
func (v VariantUnlockOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(v.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	variant := strings.ToLower(string(v.Variant))
	if slices.Contains(u.Variants, variant) {
		o.Error(lang.Translatef(l, "command.variant.unlocked.already", u.DisplayName(), variant))
		return
	}
	u.Variants = append(u.Variants, variant)
	_ = data.SaveOfflineUser(u)
	o.Print(lang.Translatef(l, "command.variant.unlocked", variant, u.DisplayName()))
}

// VariantLock locks a Vasar+ variant for a given user.
type VariantLock struct {
	Sub     lockVariant
	Targets []cmd.Target `cmd:"target"`
	Variant variantType  `cmd:"variant"`
}

// Run ...
func (v VariantLock) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(v.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := v.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	variant := strings.ToLower(string(v.Variant))
	if !u.Variant(variant) {
		o.Error(lang.Translatef(l, "command.variant.locked.already", u.DisplayName(), variant))
		return
	}
	u.LockVariant(variant)
	o.Print(lang.Translatef(l, "command.variant.locked", variant, u.DisplayName()))
}

// VariantLockOffline locks a Vasar+ variant for an offline user.
type VariantLockOffline struct {
	Sub     lockVariant
	Target  string      `cmd:"target"`
	Variant variantType `cmd:"variant"`
}

// Run ...
func (v VariantLockOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(v.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	variant := strings.ToLower(string(v.Variant))
	if !slices.Contains(u.Variants, variant) {
		o.Error(lang.Translatef(l, "command.variant.locked.already", u.DisplayName(), variant))
		return
	}
	ind := slices.Index(u.Variants, variant)
	u.Variants = slices.Delete(u.Variants, ind, ind+1)
	_ = data.SaveOfflineUser(u)
	o.Print(lang.Translatef(l, "command.variant.locked", variant, u.DisplayName()))
}

// Allow ...
func (VariantUnlock) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (VariantUnlockOffline) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (VariantLock) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (VariantLockOffline) Allow(s cmd.Source) bool {
	return allow(s, true)
}

type (
	unlockVariant string
	lockVariant   string
	variantType   string
)

// SubName ...
func (unlockVariant) SubName() string {
	return "unlock"
}

// SubName ...
func (lockVariant) SubName() string {
	return "lock"
}

// Type ...
func (variantType) Type() string {
	return "variant"
}

// Options ...
func (variantType) Options(cmd.Source) []string {
	return []string{
		"red",
		"green",
		"blue",
	}
}
