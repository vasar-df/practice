package user

// Settings contains all user configurable settings for Vasar.
type Settings struct {
	// Display is a section of settings related to what the user interface should look like.
	Display struct {
		// CPS is true if the CPS counter should be enabled.
		CPS bool `bson:"cps"`
		// Scoreboard is true if the scoreboard should be enabled.
		Scoreboard bool `bson:"scoreboard"`
	} `bson:"display"`
	// Visual is a section of settings related to visual features, such as lightning or potion splashes.
	Visual struct {
		// Lightning is true if lightning deaths should be enabled.
		Lightning bool `bson:"lightning"`
		// Splashes is true if potion splashes should be enabled.
		Splashes bool `bson:"splashes"`
		// PearlAnimation is true if players should appear to zoom instead of instantly teleport.
		PearlAnimation bool `bson:"pearl_animation"`
	} `bson:"visual"`
	// Gameplay is a section of settings related to gameplay features, such as the pearl animation or instant respawn.
	Gameplay struct {
		// ToggleSprint is true if the user should automatically toggle sprinting.
		ToggleSprint bool `bson:"toggle_sprint"`
		// AutoReapplyKit is true if the user should automatically reapply the kit.
		AutoReapplyKit bool `bson:"auto_reapply_kit"`
		// PreventInterference is true if the user should prevent interference with other players.
		PreventInterference bool `bson:"prevent_interference"`
		// PreventClutter is true if clutter should be prevented.
		PreventClutter bool `bson:"prevent_clutter"`
		// InstantRespawn is true if the user should respawn instantly.
		InstantRespawn bool `bson:"instant_respawn"`
	} `bson:"gameplay"`
	// Privacy is a section of settings related to privacy features, such as duel requests or PMs.
	Privacy struct {
		// PrivateMessages is true if private messages should be allowed.
		PrivateMessages bool `bson:"private_messages"`
		// PublicStatistics is true if the user's statistics should be public.
		PublicStatistics bool `bson:"public_statistics"`
		// DuelRequests is true if duel requests should be allowed.
		DuelRequests bool `bson:"duel_requests"`
	} `bson:"privacy"`
	// Matchmaking is a section of settings related to matchmaking features, such as the accepted ping range.
	Matchmaking struct {
		// MatchWithMobile is true if the user should be matched with mobile players.
		MatchWithMobile bool `bson:"match_with_mobile"`
		// MatchWithController is true if the user should be matched with controller players.
		MatchWithController bool `bson:"match_with_controller"`
		// MatchWithKeyboard is true if the user should be matched with keyboard players.
		MatchWithKeyboard bool `bson:"match_with_keyboard"`
		// PingRange is the maximum ping range opponents of the user can have.
		PingRange uint8 `bson:"ping_range"`
	} `bson:"matchmaking"`
	// Advanced is a section of settings related to advanced features, such as capes or splash colours.
	Advanced struct {
		// Cape is the name of the user's cape.
		Cape string `bson:"cape"`
		// ParticleMultiplier is the multiplier of combat particles.
		ParticleMultiplier int `bson:"particle_multiplier"`
		// PotionSplashColor is the colour of the potion splash particles.
		PotionSplashColor string `bson:"potion_splash_colour"`
		// VasarPlusColour is the colour of the user's Vasar Plus role.
		VasarPlusColour string `bson:"vasar_plus_colour"`
	} `bson:"advanced"`
}

// DefaultSettings returns the default settings for a new user.
func DefaultSettings() Settings {
	s := Settings{}
	s.Display.CPS = true
	s.Display.Scoreboard = true

	s.Visual.Lightning = true
	s.Visual.Splashes = true

	s.Gameplay.AutoReapplyKit = true

	s.Privacy.PrivateMessages = true
	s.Privacy.PublicStatistics = true
	s.Privacy.DuelRequests = true

	s.Matchmaking.MatchWithMobile = true
	s.Matchmaking.MatchWithController = true
	s.Matchmaking.MatchWithKeyboard = true

	s.Advanced.VasarPlusColour = "§0"
	return s
}
