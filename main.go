// package main
// trick or treat!
//
// this bot allows members to |trick or |treat every 10 minutes and receive a random candy!
//
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
	NotifChannel = "485713440258785302"
	Prefix       = "|"
	Pumpkin      = " 🎃 "
	Loli         = " 🍬 "
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
	return strconv.Itoa(t.Quantity) + " " + t.Size + "-size, " + t.State + " " + t.Name
}

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
		"Toothbrushes", "Life-lessons",
	}
)

var (
	halloweeners = make(Weeners)
)

// Weener encapsulates a Halloweener and their bag of treats
type Weener struct {
	Uid     string  `json:"uid"`
	Treats  []Treat `json:"treats"`
	Tricked bool    `json:"tricked"`
	// TODO: combo   int      `json:"combo"`
}

// NewWeener creates a new Weener struct for a halloweener
func NewWeener(uid string) Weener {
	return Weener{uid, []Treat{}, false}
}

type Weeners map[string]Weener

// GenLoli generates a random lolly in a random state to a user in the map
//
// Returns result of operation
func (w Weeners) GenLoli(uid string) string {
	// check weener
	if _, ok := halloweeners[uid]; !ok {
		halloweeners[uid] = NewWeener(uid)
	}

	// user has already been tricked
	if halloweeners[uid].Tricked {
		return "You have already been tricked!" + Pumpkin
	}

	// randomly pick treat
	rand.Seed(time.Now().UnixNano())
	treat := NewTreat(rand.Intn(20)+1, Sizes[rand.Intn(len(Sizes))], States[rand.Intn(len(States))], Names[rand.Intn(len(Names))])

	// add treat to weener's treats
	weener := halloweeners[uid]
	weener.Treats = append(weener.Treats, treat)
	halloweeners[uid] = weener

	// roll to mark as tricked
	if rand.Intn(5) == 0 {
		return "You got: **" + treat.String() + "**\n...and have been spared, keep going!" + Loli
	}

	// roll failed, trick the weener
	weener.Tricked = true
	halloweeners[uid] = weener
	return "You got: **" + treat.String() + "**\n...and have been tricked!" + Pumpkin
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
	log.Println("Started spookyness!")

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

			out := "Your treats:\n"
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
		// reset all tricked fields every so often
		for k, _ := range halloweeners {
			weener := halloweeners[k]
			weener.Tricked = false
			halloweeners[k] = weener
		}
		dgo.ChannelMessageSend(NotifChannel, Pumpkin+" Everyone has been untricked! "+Pumpkin)

		// and save map to a file
		b, err := json.MarshalIndent(halloweeners, "  ", "  ")
		if err != nil {
			log.Println(err)
		}

		err = ioutil.WriteFile("halloween.json", b, 0644)
		if err != nil {
			log.Println(err)
		}

		timer = time.NewTimer(10 * time.Second)
		<-timer.C
	}
}
