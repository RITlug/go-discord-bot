package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	// Variables used for command line parameters
	CommandLineOpts *flag.FlagSet
	Token           string
	// Used in matching the user's message text against the different branches of the knock-knock joke
	WhosThere *regexp.Regexp = regexp.MustCompile(`who['’]?s there\??`)
)

// Returns whether or not the given message mentions the given user
func messageMentions(m *discordgo.Message, user *discordgo.User) bool {
	return strings.Contains(m.Content, user.Mention())
}

func init() {
	CommandLineOpts = flag.NewFlagSet("Discord Bot", flag.ExitOnError)
	CommandLineOpts.StringVar(&Token, "t", "", "Bot Token")
	CommandLineOpts.Parse(os.Args[1:])

	if len(Token) == 0 {
		CommandLineOpts.Usage()
		os.Exit(2)
	}

}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		os.Exit(1)
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		os.Exit(3)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// What the user typed in the message
	text := strings.ToLower(m.Content)

	fmt.Fprintln(os.Stderr, "Got message:", m.Content)

	if messageMentions(m.Message, s.State.User) && strings.Contains(text, "tell me a joke") {
		// If the user @-s the bot with a message containing the string "tell me a joke" (case-insensitive)
		response := fmt.Sprintf("%s Knock, knock.", m.Author.Mention())
		s.ChannelMessageSend(m.ChannelID, response)
	} else if WhosThere.Match([]byte(text)) {
		// Response after saying "Knock, knock."
		s.ChannelMessageSend(m.ChannelID, "To.")
	} else if strings.Contains(text, "to who?") {
		// Response after saying "To."
		s.ChannelMessageSend(m.ChannelID, "Actually, it's to _whom_. :nerd:")
	}
}
