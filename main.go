package main

import (
    "fmt"
    "encoding/json"
    "io/ioutil"
	"log"
	"os"
    "sort"
    "strings"

	"github.com/nlopes/slack"
)

const (
    freeOwner = "Free"
)

type Resources struct {
    Resources []Resource `json:"resources"`
}

type Resource struct {
        Name string         `json:"name"`
        Description string  `json:"description"`
        Owner string        `json:"owner"`
    }

var resource Resources

func updateFile() {
	file, _ := json.MarshalIndent(resource, "", " ")
	_ = ioutil.WriteFile("gotcha.json", file, 0644)
}

func processMessage(ev *slack.MessageEvent, api *slack.Client) {
    botInfo, err := api.GetBotInfo("BMLAVS5N3")
    if err != nil {
        log.Printf("Get User Identity error: %s\n", err)
    }
    log.Printf("TEST USER IDENTITY: %s-%s-%s", ev.SubType, botInfo.ID, ev.BotID)
    if ev.SubType == "bot_message" && botInfo.ID == ev.BotID {
        return
    }
    idx := 0
    // case *slackevents.AppMentionEvent:
    //  idx += 1
	user, err := api.GetUserInfo(ev.User)
	if err != nil {
		log.Printf("Get User Info error (message event): %s\n", err)
        return
	}
    s := strings.Fields(ev.Text)
    if len(s) > (idx+1) || s[idx] == "list" || s[idx] == "help" {
        var userParam slack.GetConversationsForUserParameters
        channels, _, err := api.GetConversationsForUser(&userParam)
        if err != nil {
            log.Printf("Get Conversations error: %s\n", err)
        }
        channelId := channels[0].ID
        switch s[idx] {
        case "lock":
            name := s[idx+1]
            for i, s := range resource.Resources {
                if s.Name == name {
                    if s.Owner == freeOwner {
                        resource.Resources[i].Owner = ev.User
                        api.PostMessage(channelId, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                    } else if s.Owner == ev.User {
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " already locked by " + user.RealName + " (You)", false))
                    } else {
                        user, err := api.GetUserInfo(s.Owner)
                        if err != nil {
                            log.Printf("Get User Info error (lock): %s\n", err)
                        }
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                    }
                    return
                }
            }
			updateFile()
            api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
        case "unlock":
            name := s[idx+1]
            for i, s := range resource.Resources {
                if s.Name == name {
                    if s.Owner == ev.User {
                        resource.Resources[i].Owner = freeOwner
                        api.PostMessage(channelId, slack.MsgOptionText(name + " unlocked by " + user.RealName, false))
                    } else if s.Owner == freeOwner {
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " already unlocked", false))
                    } else {
                        user, err := api.GetUserInfo(s.Owner)
                        if err != nil {
                            log.Printf("Get User Info error (unlock): %s\n", err)
                        }
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                    }
                    return
                }
            }
			updateFile()
            api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
        case "add":
            name := s[idx+1]
            description := "-"
            if len(s) > 2 {
                description = strings.Join(s[idx+2:], " ")
            }
            for _, s := range resource.Resources {
                if s.Name == name {
                    api.PostMessage(ev.Channel, slack.MsgOptionText(name + " already added", false))
                    return
                }
            }
            newResource := Resource {
                Name: name,
                Description: description,
                Owner: freeOwner,
            }
            resource.Resources = append(resource.Resources, newResource)
            sort.Slice(resource.Resources, func(i, j int) bool {
                return resource.Resources[i].Name < resource.Resources[j].Name
            })
			updateFile()
            api.PostMessage(channelId, slack.MsgOptionText(name + " (" + description + ") added by " + user.RealName, false))
        case "delete":
            name := s[idx+1]
            for i, s := range resource.Resources {
                if s.Name == name {
                    if s.Owner == freeOwner {
                        resource.Resources = append(resource.Resources[:i], resource.Resources[i+1:]...)
                        api.PostMessage(channelId, slack.MsgOptionText(name + " deleted by " + user.RealName, false))
                    } else if s.Owner == ev.User {
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName + " (You)", false))
                    } else {
                        user, err := api.GetUserInfo(s.Owner)
                        if err != nil {
                            log.Printf("Get User Info error (delete): %s\n", err)
                        }
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                    }
                    return
                }
            }
			updateFile()
            api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
        case "update":
            name := s[idx+1]
            description := "-"
            if len(s) > 2 {
                description = strings.Join(s[idx+2:], " ")
            }
            for i, s := range resource.Resources {
                if s.Name == name {
                    if s.Owner == ev.User || s.Owner == freeOwner {
                        resource.Resources[i].Description = description
                        api.PostMessage(channelId, slack.MsgOptionText(name + " (" + description + ") updated by " + user.RealName, false))
                    } else {
                        user, err := api.GetUserInfo(s.Owner)
                        if err != nil {
                            log.Printf("Get User Info error (update): %s\n", err)
                        }
                        api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                    }
                    return
                }
            }
            api.PostMessage(ev.Channel, slack.MsgOptionText(name + "not added ", false))
        case "list":
            var text string
            if len(resource.Resources) != 0 {
                text += fmt.Sprintf("```%15.15s %-20.20s %s\n", "Name", "Owner", "Description")
                for _, s := range resource.Resources {
                    owner := freeOwner
                    if s.Owner != freeOwner {
                        user, _ = api.GetUserInfo(s.Owner)
                        owner = user.RealName
                    }
                    text += fmt.Sprintf("%15.15s %-20.20s %s\n", s.Name, owner, s.Description)
                }
                text += "```"
            } else {
                text = "No resource added yet. Please refer to the add command to create one."
            }
            api.PostMessage(ev.Channel, slack.MsgOptionText(text, false))
        case "help":
            api.PostMessage(ev.Channel, slack.MsgOptionText("Available commands are:\nlock <resource>\nunlock <resource>\nadd <resource> <description>\ndelete <resource>\nupdate <resource> <description>\nlist", false))
        }
    }
}

func getenv(name string) string {
    v := os.Getenv(name)
    if v == "" {
        panic("missing required environment variable " + name)
    }
    return v
}

func main() {
    token := getenv("SLACKTOKEN")
	api := slack.New(
		token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	file, _ := ioutil.ReadFile("gotcha.json")
	_ = json.Unmarshal([]byte(file), &resource)

	for msg := range rtm.IncomingEvents {
		log.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			log.Println("Infos:", ev.Info)
			log.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.MessageEvent:
            channel, err := api.GetConversationInfo(ev.Channel, false)
            if err != nil {
                log.Printf("Get Conversation info error: %s\n", err)
                return
            }

            if channel.IsIM {
                processMessage(ev, api)
            } else {
                log.Printf("The message has not come from a direct message: %s\n", ev.Text)
            }

		case *slack.PresenceChangeEvent:
			log.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			log.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return

		default:
			log.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
