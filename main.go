// package main
// trick or treat!
//
// this bot allows members to |trick or |treat every 10 minutes and receive a random candy!
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	//NotifChannel = "485713440258785302"
	NotifChannel = "543714149536890883" // #commands
	File         = "halloweeners.json"
	Period       = 15 * time.Minute

	Prefix  = "|"
	Pumpkin = " 🎃 "
	Loli    = " 🍬 "
)

var (
	States = []string{
		"prized", "golden", "limited-edition",
		"clean", "fresh", "mint-condition",
		"moldy", "expired", "suspicious-looking", "old", "suspect",
		"ok-looking", "meh", "strange-smelling", "musty",
	}
	Sizes = []string{
		"full", "fun", "family", "kid", "baby", "American", "industrial", "elephant", "decent",
	}
	Names = []string{
		"Snickers", "Milky ways", "Mars bars", "Pods", "Bountys", "Cadburys", "Freddo Frogs", "Caramello Koalas", "Aero bars",
		"Nerds", "Sour worms", "Sour Skittles", "Skittles", "Warheads", "Wizz Fizzes",
		"Minties", "5 gum",
		"Grapes", "Apples",
		"Toothbrushes", "Life lessons",
	}
)

// Treat encapsulates a treat with some quantity, state, and name
//
// e.g. "2 expired fun-size snicker's bar"
type Treat struct {
	Quantity int    `json:"quantity"`
	State    string `json:"state"`
	Size     string `json:"size"`
	Name     string `json:"name"`
}

// NewTreat creates a new treat
func NewTreat(quant int, size, state, name string) Treat {
	return Treat{quant, state, size, name}
}

// String returns a string form of the treat
func (t Treat) String() string {
	return strconv.Itoa(t.Quantity) + " " + t.State + ", " + t.Size + "-size " + t.Name
}

// Weener encapsulates a Halloweener and their bag of treats
type Weener struct {
	Uid      string  `json:"uid"`
	Treats   []Treat `json:"treats"`
	MaxCombo int     `json:"max combo"`
	tricked  bool
	combo    int
}

// NewWeener creates a new Weener struct for a halloweener
func NewWeener(uid string) Weener {
	return Weener{
		Uid:      uid,
		Treats:   []Treat{},
		MaxCombo: 0,
		tricked:  false,
		combo:    0,
	}
}

// Weeners is the collection of halloweeners
type Weeners map[string]Weener

// global halloweeners
var halloweeners = make(Weeners)

// GetComboLeader returns the uid of the current combo leader
func (w Weeners) GetComboLeader() string {
	max := 0
	leader := ""
	for _, v := range w {
		if v.MaxCombo > max {
			leader = v.Uid
			max = v.MaxCombo
		}
	}
	return leader
}

// GetTrickLeader returns the uid of the current amount leader
func (w Weeners) GetTrickLeader() string {
	max := 0
	leader := ""
	for _, v := range w {
		if len(v.Treats) > max {
			leader = v.Uid
			max = len(v.Treats)
		}
	}
	return leader
}

// GenLoli generates a random lolly in a random state to a user in the map
//
// Returns result of operation
func (w Weeners) GenLoli(uid string) string {
	// check weener
	if _, ok := halloweeners[uid]; !ok {
		halloweeners[uid] = NewWeener(uid)
	}

	// user has already been tricked
	if halloweeners[uid].tricked {
		return "You have already been tricked!" + Pumpkin
	}

	// randomly pick treat
	rand.Seed(time.Now().UnixNano())
	treat := NewTreat(rand.Intn(15)+1, Sizes[rand.Intn(len(Sizes))], States[rand.Intn(len(States))], Names[rand.Intn(len(Names))])

	// add treat to weener's treats
	weener := halloweeners[uid]
	weener.Treats = append(weener.Treats, treat)

	// roll to mark as tricked
	if rand.Intn(2+len(weener.Treats)/10) == 0 {
		if weener.combo > weener.MaxCombo {
			weener.MaxCombo = weener.combo
		}
		halloweeners[uid] = weener
		return "You got: **" + treat.String() + "**\n...keep going!" + Loli
	}

	// roll failed, trick the weener
	weener.tricked = true
	halloweeners[uid] = weener
	return "You got: **" + treat.String() + "**\n...and have been tricked!" + Pumpkin
}

func init() {
	// attempt to read from a saved file
	b, err := ioutil.ReadFile(File)
	if err == nil {
		err = json.Unmarshal(b, &halloweeners)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	var dgo *discordgo.Session
	var err error

	// open discord connection
	if key, ok := os.LookupEnv("SPOOK"); ok {
		dgo, err = discordgo.New("Bot " + key)
		if err != nil {
			log.Fatalln(err)
		}
		err = dgo.Open()
		if err != nil {
			log.Fatalln(err)
		}
		defer dgo.Close()
	} else {
		log.Fatalln(err)
	}

	log.Println("TRICK OR TREAT")
	dgo.UpdateStatus(0, Prefix+"trick or "+Prefix+"treat")

	// message handler
	dgo.AddHandler(func(ses *discordgo.Session, msg *discordgo.MessageCreate) {
		message := strings.TrimSpace(msg.Content)

		if msg.Author.Bot {
			return
		}

		if !strings.HasPrefix(message, Prefix) {
			return
		}

		// handle commands
		com := strings.ToLower(message[1:])
		if com == "trick" {
			ses.ChannelMessageSend(msg.ChannelID, halloweeners.GenLoli(msg.Author.ID))
			return
		}

		if com == "treat" {
			weener, ok := halloweeners[msg.Author.ID]
			if !ok || len(weener.Treats) == 0 {
				ses.ChannelMessageSend(msg.ChannelID, "You have no treats!")
				return
			}

			out := "You have " + strconv.Itoa(len(weener.Treats)) + " tricks:\n"
			for _, treat := range weener.Treats {
				out += treat.String() + "\n"
			}
			ses.ChannelMessageSend(msg.ChannelID, out)
			return
		}
	})

	// event loop
	var timer *time.Timer
	for {
		// reset all tricked fields
		for k, _ := range halloweeners {
			weener := halloweeners[k]
			weener.tricked = false
			halloweeners[k] = weener
		}

		// get leaders
		out := Pumpkin + " Everyone has been untricked! " + Pumpkin
		comboID := halloweeners.GetComboLeader()
		comboUser, err := dgo.User(comboID)
		if err == nil {
			out += "\n" + Loli + "Current Combo Leader: " + " [" + strconv.Itoa(halloweeners[comboID].MaxCombo) + "]" + comboUser.Username
		}

		trickID := halloweeners.GetTrickLeader()
		trickUser, err := dgo.User(trickID)
		if err == nil {
			out += "\n" + Loli + "Current Trick Leader: " + " [" + strconv.Itoa(len(halloweeners[trickID].Treats)) + "]" + trickUser.Username
		}
		dgo.ChannelMessageSend(NotifChannel, out)

		// and save map to a file
		b, err := json.MarshalIndent(halloweeners, "", "  ")
		if err != nil {
			log.Println(err)
		}

		err = ioutil.WriteFile(File, b, 0644)
		if err != nil {
			log.Println(err)
		}

		// wait another period
		timer = time.NewTimer(Period)
		<-timer.C
	}
}
