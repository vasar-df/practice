package game

import (
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/vails"
	"strings"
)

// Game represents a Variant of a game.
type Game struct {
	id      uint8
	name    string
	texture string
}

// NoDebuff returns the NoDebuff game.
func NoDebuff() Game {
	return Game{0, "NoDebuff", "textures/items/potion_bottle_splash_heal"}
}

// Debuff returns the Debuff game.
func Debuff() Game {
	return Game{1, "Debuff", "textures/items/potion_bottle_splash_poison"}
}

// Gapple returns the Gapple game.
func Gapple() Game {
	return Game{2, "Gapple", "textures/items/apple_golden"}
}

// Soup returns the Soup game.
func Soup() Game {
	return Game{3, "Soup", "textures/items/mushroom_stew"}
}

// Boxing returns the Boxing game.
func Boxing() Game {
	return Game{4, "Boxing", "textures/items/slimeball"}
}

// StickFight returns the StickFight game.
func StickFight() Game {
	return Game{5, "Stick Fight", "textures/items/stick"}
}

// Sumo returns the Sumo game.
func Sumo() Game {
	return Game{id: 6, name: "Sumo", texture: "textures/items/feather"}
}

// Combo returns the Combo game.
func Combo() Game {
	return Game{7, "Combo", "textures/items/fish_pufferfish_raw"}
}

// BuildUHC returns the BuildUHC game.
func BuildUHC() Game {
	return Game{8, "BuildUHC", "textures/items/bucket_lava"}
}

// Games returns all the games supported.
func Games() []Game {
	return []Game{NoDebuff(), Boxing(), BuildUHC(), Sumo(), StickFight(), Soup(), Combo(), Gapple(), Debuff()}
}

// FFA returns all FFA-supported games.
func FFA() []Game {
	return []Game{NoDebuff(), Sumo()}
}

// ByName returns the game with the given name.
func ByName(name string) Game {
	for _, g := range Games() {
		if g.name == name {
			return g
		}
	}
	panic("should never happen")
}

// ByString returns the game with the given string.
func ByString(s string) Game {
	for _, g := range Games() {
		if g.String() == s {
			return g
		}
	}
	panic("should never happen")
}

// ByID returns the game with the given ID.
func ByID(id byte) Game {
	for _, g := range Games() {
		if g.id == id {
			return g
		}
	}
	panic("should never happen")
}

// Name ...
func (g Game) Name() string {
	return g.name
}

// String ...
func (g Game) String() string {
	return strings.ReplaceAll(strings.ToLower(g.name), " ", "_")
}

// Texture ...
func (g Game) Texture() string {
	return g.texture
}

// Kit ...
func (g Game) Kit(ffa bool) vails.Kit {
	switch g {
	case NoDebuff():
		return kit.NoDebuff{FFA: ffa}
	case Debuff():
		return kit.Debuff{}
	case Gapple():
		return kit.Gapple{}
	case Soup():
		return kit.Soup{FFA: ffa}
	case Boxing():
		return kit.Boxing{}
	case StickFight():
		return kit.StickFight{}
	case Sumo():
		return kit.Sumo{FFA: ffa}
	case Combo():
		return kit.Combo{}
	case BuildUHC():
		return kit.BuildUHC{}
	}
	panic("should never happen")
}

// Cap ...
func (g Game) Cap() int {
	switch g {
	case NoDebuff():
		return 60
	case Sumo():
		return 30
	}
	panic("should never happen")
}
