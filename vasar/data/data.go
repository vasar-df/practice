package data

import (
	"encoding/json"
	"fmt"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mongo"
	"github.com/vasar-network/practice/vasar/user"
	"os"
	"time"
)

// userData is the data structure that is used to store the user data in the database.
type userData struct {
	// XUID is the XUID of the user.
	XUID string `bson:"xuid"`
	// Name is the last username of the user.
	Name string `bson:"name"`
	// DisplayName is the name displayed to other users.
	DisplayName string `bson:"display_name"`
	// DeviceID is the device ID of the last device the user logged in from.
	DeviceID string `bson:"did"`
	// SelfSignedID is the self-signed ID of the last client session of the user.
	SelfSignedID string `bson:"ssid"`
	// Address is the hashed IP address of the user.
	Address string `bson:"address"`
	// Whitelisted is true if the user is whitelisted.
	Whitelisted bool `bson:"whitelisted"`
	// FirstLogin is the time the user first logged in.
	FirstLogin time.Time `bson:"first_login"`
	// PlayTime is the duration the user has played Vasar Practice for.
	PlayTime time.Duration `bson:"playtime"`

	// Variants is a list of Vasar+ variants the user has unlocked.
	Variants []string `bson:"variants"`
	// Roles is a list of roles that the user has.
	Roles []roleData `bson:"roles"`
	// Punishments is a list of active punishments that the user has.
	Punishments punishmentData `bson:"punishments"`

	// Settings is a list of settings that the user has.
	Settings user.Settings `bson:"settings"`
	// Practice is a list of user statistics specific to Vasar Practice.
	Practice user.Stats `bson:"practice"`
}

// roleData is the data structure that is used to store roles in the database.
type roleData struct {
	// Name is the name of the role.
	Name string `bson:"name"`
	// Expires is true if the role expires.
	Expires bool `bson:"expires"`
	// Expiration is the expiration time of the role.
	Expiration time.Time `bson:"expiration"`
}

// punishmentData is the data structure that is used to store punishments in the database.
type punishmentData struct {
	// Mute contains punishment data on the user's mute.
	Mute user.Punishment `bson:"mute"`
	// Ban contains punishment data on the user's ban.
	Ban user.Punishment `bson:"ban"`
}

// salt contains the salt that starts with "THIS NIGGA GOT A FAT DICK" used for hashing. Please don't judge.
const salt = "THIS NIGGA GOT A FAT DICK I LOVE BIG DICKS YEAH UH, PUSSY ROLLLLLLLLIN, I LIKE BIG DICKS!!! PUSSY ROLLLLLLLIN OH YEAH WHATS GOOD NEEGUS!!!! BIG FAT COCK AND SHE SUCK WHEN I SAY SO TWO BIG TITTIES AND A BIG ASS SEX-SO THREE MORE CUM DROPS ALL ON HER ASS-HOLE, PULL UP THIS DICK THATS A BIG FUCKING PUSSY HOLE YEAHHHHHHHHHHHH!"

// sess is the Upper database session.
var sess db.Session

// init creates the Upper database connection.
func init() {
	path := os.Getenv("VASAR_DB")
	if len(path) == 0 {
		panic("vasar: mongo environment variable is not set")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("vasar: %s", err))
	}

	var settings mongo.ConnectionURL
	err = json.Unmarshal(b, &settings)
	if err != nil {
		panic(fmt.Sprintf("vasr: %s", err))
	}

	sess, err = mongo.Open(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to start mongo connection: %v", err))
	}
}
