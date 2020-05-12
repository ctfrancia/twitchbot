package bot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

// Regex for parsing PRIVMSG strings.
//
// First matched group is the user's name and the second matched group is the content of the
// user's message.
var msgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

// Regex for parsing user commands, from already parsed PRIVMSG strings.
//
// First matched group is the command name and the second matched group is the argument for the
// command.
var cmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

// PSTFormat is the format of dates
const PSTFormat = "2 Jan 15:04:05"

// BasicBot struct
type BasicBot struct {
	Channel     string
	conn        net.Conn
	Credentials *OAuthCred
	MsgRate     time.Duration
	Name        string
	Port        string
	PrivatePath string
	Server      string
	startTime   time.Time
}

// OAuthCred struct
type OAuthCred struct {
	Password string `json:"password,omitempty"`
}

// TwitchBot interface
type TwitchBot interface {
	Connect()
	Disconnect()
	HandleChat() error
	JoinChannel()
	ReadCredentials() error
	Start()
}

// Connect method for connecting to the twitch channel
func (bb *BasicBot) Connect() {
	var err error
	fmt.Printf("[%s] Connecting to %s...\n", timeStamp(), bb.Server)

	// makes connection to Twitch IRC server
	bb.conn, err = net.Dial("tcp", bb.Server+":"+bb.Port)
	if err != nil {
		fmt.Printf("[%s] cannot connect to %s, retrying.\n", timeStamp(), bb.Server)
		return
	}

	fmt.Printf("[%s] Connected to %s!\n", timeStamp(), bb.Server)
	bb.startTime = time.Now()
}

// HandleChat reads the messages of the channel
func (bb *BasicBot) HandleChat() error {
	fmt.Printf("[%s] Watching #%s...\n", timeStamp(), bb.Channel)

	// reads from connection
	tp := textproto.NewReader(bufio.NewReader(bb.conn))

	// reads messages
	for {
		line, err := tp.ReadLine()
		if err != nil {
			bb.Disconnect()
			return errors.New("bb.Bot.HandleChat: Failed to read from channel. Disconnected")
		}
		fmt.Printf("[%s] %s\n", timeStamp(), line)
		if "PING :tmi.twitch.tv" == line {
			// respond to PING message with a PONG message, to maintain the connection
			bb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
			continue
		} else {
			matches := msgRegex.FindStringSubmatch(line)
			if nil != matches {
				userName := matches[1]
				msgType := matches[2]

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					fmt.Printf("[%s] %s: %s\n", timeStamp(), userName, msg)

					// parse commands from user message
					cmdMatches := cmdRegex.FindStringSubmatch(msg)
					if cmdMatches != nil {
						cmd := cmdMatches[1]
						// arg := cmdMatches[2]

						// channel-owener specific commands
						if userName == bb.Channel {
							switch cmd {
							case "tbdown":
								fmt.Printf(
									"[%s] Shutdown command received. Shutting down now...\n",
									timeStamp(),
								)
								bb.Disconnect()
								return nil

							default:
								fmt.Printf("%s command received", cmd)
							}
						}
					}
				default:
					// do nothing
				}
			}

		}
		time.Sleep(bb.MsgRate)

	}

}

// Start starts a loop where the bot will attempt to connect to the Twitch channel
// it will continue to do so until told to shut down
func (bb *BasicBot) Start() {
	err := bb.ReadCredentials()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting...")
		return
	}

	for {
		bb.Connect()
		bb.JoinChannel()
		err = bb.HandleChat()
		if err != nil {
			// attempts to reconnect upon chat error
			time.Sleep(1 * time.Second)
			fmt.Println(err)
			fmt.Println("Starting bot again...")
		} else {
			return
		}
	}
}

// Say speaks to the channel
func (bb *BasicBot) Say(msg string) error {
	if msg == "" {
		return errors.New("BasicBot.Say: msg was empty")
	}
	_, err := bb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s %s\r\n", bb.Channel, msg)))
	if err != nil {
		return err
	}
	return nil
}

// JoinChannel joines the requested channel
func (bb *BasicBot) JoinChannel() {
	fmt.Printf("[%s] Joining #%s...\n", timeStamp(), bb.Channel)
	bb.conn.Write([]byte("PASS " + bb.Credentials.Password + "\r\n"))
	bb.conn.Write([]byte("NICK " + bb.Name + "\r\n"))
	bb.conn.Write([]byte("JOIN #" + bb.Channel + "\r\n"))

	fmt.Printf("[%s] Joined #%s as @%s!\n", timeStamp(), bb.Channel, bb.Name)
}

// ReadCredentials reads the credentials from a path in order to make a connection
func (bb *BasicBot) ReadCredentials() error {
	// reads from the file
	credFile, err := ioutil.ReadFile(bb.PrivatePath)
	if err != nil {
		return err
	}

	bb.Credentials = &OAuthCred{}
	// parses the file contents
	dec := json.NewDecoder(strings.NewReader(string(credFile)))
	if err = dec.Decode(bb.Credentials); err != nil && io.EOF != err {
		return err
	}

	return nil
}

// Disconnect will disconnect from the twitch channel connected
func (bb *BasicBot) Disconnect() {
	bb.conn.Close()
	// upTime := time.Now().Sub(bb.startTime).Seconds()
	fmt.Printf("[%s] Closed connection from %s | Live for:", timeStamp(), bb.Server)
}
func timeStamp() string {
	return TimeStamp(PSTFormat)
}

// TimeStamp formats the time
func TimeStamp(format string) string {
	return time.Now().Format(format)
}
