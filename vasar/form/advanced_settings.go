package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/cape"
	"github.com/vasar-network/vails/role"
	"golang.org/x/exp/slices"
	"regexp"
	"strings"
)

var (
	// capes contains a list of all the capes that are available to the user.
	capes = []string{"None"}
	// particleMultipliers contains all possible particle multipliers.
	particleMultipliers = []string{"x0", "x1", "x2"}
	// splashColours contains all possible customizations for the splash colour of potions.
	splashColours = []string{
		"Default",
		"Invisible",
		text.Colourf("<red>Red</red>"),
		text.Colourf("<gold>Orange</gold>"),
		text.Colourf("<yellow>Yellow</yellow>"),
		text.Colourf("<green>Green</green>"),
		text.Colourf("<aqua>Aqua</aqua>"),
		text.Colourf("<blue>Blue</blue>"),
		text.Colourf("<purple>Pink</purple>"),
		text.Colourf("<white>White</white>"),
		text.Colourf("<grey>Gray</grey>"),
		text.Colourf("<dark-grey>Black</dark-grey>"),
	}
	// plusColours contains all possible customizations for Vasar+ colours.
	plusColours = []string{
		text.Colourf("<black>Default</black>"),
		text.Colourf("<red>Red</red>"),
		text.Colourf("<green>Green</green>"),
		text.Colourf("<blue>Blue</blue>"),
	}

	// formatRegex is a regex used to clean color formatting on a string.
	formatRegex = regexp.MustCompile(`§[\da-gk-or]`)
)

// init initializes all capes.
func init() {
	for _, c := range cape.All() {
		capes = append(capes, c.Name())
	}
}

// advanced is a form that allows the user to manage their advanced settings.
type advanced struct {
	// Cape is a dropdown that allows the user to change their cape.
	Cape form.Dropdown
	// Particles is a dropdown that allows the user to change the multiplication of the combat particles.
	ParticleMultiplier form.Dropdown
	// PotionParticles is a dropdown that allows the user to change the color of their splash potion particles.
	PotionSplashColour form.Dropdown
	// VasarPlusColour is a dropdown that allows the user to change the color of the Vasar+ role.
	VasarPlusColour form.Dropdown
	// u is the user that is using the form.
	u *user.User
}

// NewAdvanced returns a new advanced form for the player to modify their advanced settings.
func NewAdvanced(u *user.User) form.Form {
	s := u.Settings()
	capeIndex := slices.Index(capes, s.Advanced.Cape)
	if capeIndex == -1 {
		capeIndex = 0
	}

	splashColourIndex := slices.IndexFunc(splashColours, func(colour string) bool {
		return strings.ToLower(stripFormatting(colour)) == s.Advanced.PotionSplashColor
	})
	if splashColourIndex == -1 {
		splashColourIndex = 0
	}

	plusColourIndex := slices.IndexFunc(plusColours, func(colour string) bool {
		return strings.TrimSuffix(stripResets(colour), stripFormatting(colour)) == s.Advanced.VasarPlusColour
	})
	if plusColourIndex == -1 {
		plusColourIndex = 0
	}

	return form.New(advanced{
		Cape:               form.NewDropdown("Cape:", capes, capeIndex),
		ParticleMultiplier: form.NewDropdown("Particle Multiplier:", particleMultipliers, s.Advanced.ParticleMultiplier),
		PotionSplashColour: form.NewDropdown("Potion Splash Color:", splashColours, splashColourIndex),
		VasarPlusColour:    form.NewDropdown("Vasar+ Variant:", plusColours, plusColourIndex),
		u:                  u,
	}, "Advanced Settings")
}

// Submit ...
func (a advanced) Submit(form.Submitter) {
	s := a.u.Settings()
	c, _ := cape.ByName(capes[a.Cape.Value()])
	if c.Premium() && !a.u.Roles().Contains(role.Plus{}) {
		a.u.Message("setting.plus")
		return
	}
	if name := c.Name(); s.Advanced.Cape != name {
		s.Advanced.Cape = name
		a.u.SetSettings(s)

		skin := a.u.Player().Skin()
		skin.Cape = c.Cape()
		a.u.Player().SetSkin(skin)
	}

	formattedPlusColour := plusColours[a.VasarPlusColour.Value()]
	if a.VasarPlusColour.Value() > 0 && !a.u.Variant(strings.ToLower(stripFormatting(formattedPlusColour))) {
		a.u.Message("variant.locked")
		return
	}

	splashColour := strings.ToLower(stripFormatting(splashColours[a.PotionSplashColour.Value()]))
	plusColour := strings.TrimSuffix(stripResets(formattedPlusColour), stripFormatting(formattedPlusColour))
	if s.Advanced.ParticleMultiplier != a.ParticleMultiplier.Value() || s.Advanced.PotionSplashColor != splashColour || s.Advanced.VasarPlusColour != plusColour {
		s.Advanced.ParticleMultiplier = a.ParticleMultiplier.Value()
		s.Advanced.PotionSplashColor = splashColour
		s.Advanced.VasarPlusColour = plusColour
		a.u.SetNameTagFromRole()
		if !a.u.Roles().Contains(role.Plus{}) {
			a.u.Message("setting.plus")
			return
		}
	}

	a.u.SetSettings(s)
	a.u.Player().SendForm(NewAdvanced(a.u))
}

// Close ...
func (a advanced) Close(form.Submitter) {
	a.u.Player().SendForm(NewSettings(a.u))
}

// stripResets removes all resets from a string.
func stripResets(s string) string {
	return strings.ReplaceAll(s, "§r", "")
}

// stripFormatting removes all formatting from a string.
func stripFormatting(s string) string {
	return formatRegex.ReplaceAllString(s, "")
}
