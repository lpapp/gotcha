package main

import (
	"encoding/json"
	"errors"
	"fmt"
    "log"
    "net/http"
    "sort"
    "strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

    "github.com/nlopes/slack"
    "github.com/nlopes/slack/slackevents"
)

var api = slack.New("xoxb-720718298770-732510316389-uuwrs1SltTvfFOjxLQGe7RiX")

var freeOwner = "Free"

type Switches struct {
    Switches []Switch `json:"switches"`
}

type Switch struct {
        Name string         `json:"name"`
        Description string  `json:"description"`
        Owner string        `json:"owner"`
    }

var switches Switches

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if request.HTTPMethod == "GET" {

		return events.APIGatewayProxyResponse{
            Body:       fmt.Sprintf("Welcome to go-tcha, the resource allocator!"),
			StatusCode: http.StatusOK,
		}, nil
	}

	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, errors.New("Empty HTTP body")
	}

    log.Print("TEST BODY: " + request.Body)

	eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(request.Body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: "DY9JYMzemYF46BoQA6THCHfL"}))
	if e != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Internal Server Error"),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(request.Body), &r)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("Internal Server Error"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			Body: r.Challenge,
			StatusCode: http.StatusOK,
		}, nil
	}
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
        case *slackevents.MessageEvent:
            botInfo, err := api.GetBotInfo("BMLAVS5N3")
			if err != nil {
                log.Printf("Get User Identity error: %s\n", err)
			}
            log.Printf("TEST USER IDENTITY: %s-%s", botInfo, ev.User)
            if botInfo.ID == ev.BotID {
                return events.APIGatewayProxyResponse{
                    StatusCode: http.StatusOK,
                }, nil
            }
            idx := 0
        // case *slackevents.AppMentionEvent:
        //  idx += 1
			user, err := api.GetUserInfo(ev.User)
			if err != nil {
                log.Printf("Get User Info error (message event): %s\n", err)
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
					for i, s := range switches.Switches {
						if s.Name == name {
							if s.Owner == freeOwner {
								switches.Switches[i].Owner = ev.User
                                api.PostMessage(channelId, slack.MsgOptionText(name + " locked by " + user.RealName, false))
							} else {
                                user, err := api.GetUserInfo(s.Owner)
                                if err != nil {
                                    log.Printf("Get User Info error (lock): %s\n", err)
                                }
                                api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
							}
                            return events.APIGatewayProxyResponse{
                                StatusCode: http.StatusOK,
                            }, nil
						}
					}
                    api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
                case "unlock":
					name := s[idx+1]
                    for i, s := range switches.Switches {
                        if s.Name == name {
                            if s.Owner == ev.User {
                                switches.Switches[i].Owner = freeOwner
                                api.PostMessage(channelId, slack.MsgOptionText(name + " unlocked by " + user.RealName, false))
                            } else {
                                user, err := api.GetUserInfo(s.Owner)
                                if err != nil {
                                    log.Printf("Get User Info error (unlock): %s\n", err)
                                }
                                api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                            }
                            return events.APIGatewayProxyResponse{
                                StatusCode: http.StatusOK,
                            }, nil
                        }
                    }
                    api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
                case "add":
					name := s[idx+1]
                    description := "-"
                    if len(s) > 2 {
                        description = strings.Join(s[idx+2:], " ")
                    }
					for _, s := range switches.Switches {
						if s.Name == name {
                            api.PostMessage(ev.Channel, slack.MsgOptionText(name + " already added", false))
							return events.APIGatewayProxyResponse{
								StatusCode: http.StatusOK,
							}, nil
						}
					}
					newSwitch := Switch {
						Name: name,
						Description: description,
						Owner: freeOwner,
					}
					switches.Switches = append(switches.Switches, newSwitch)
					sort.Slice(switches.Switches, func(i, j int) bool {
						return switches.Switches[i].Name < switches.Switches[j].Name
					})
                    api.PostMessage(channelId, slack.MsgOptionText(name + " added by " + user.RealName, false))
                case "delete":
					name := s[idx+1]
					for i, s := range switches.Switches {
						if s.Name == name {
                            if s.Owner == freeOwner {
                                switches.Switches = append(switches.Switches[:i], switches.Switches[i+1:]...)
                                api.PostMessage(channelId, slack.MsgOptionText(name + " deleted by " + user.RealName, false))
                            } else {
                                user, err := api.GetUserInfo(s.Owner)
                                if err != nil {
                                    log.Printf("Get User Info error (delete): %s\n", err)
                                }
                                api.PostMessage(ev.Channel, slack.MsgOptionText(name + " locked by " + user.RealName, false))
                            }
                            return events.APIGatewayProxyResponse{
                                StatusCode: http.StatusOK,
                            }, nil
						}
					}
                    api.PostMessage(ev.Channel, slack.MsgOptionText(name + " not added", false))
                case "list":
                    var text string
                    if len(switches.Switches) != 0 {
					    text += fmt.Sprintf("```%15.15s %-20.20s %s\n", "Name", "Owner", "Description")
                        for _, s := range switches.Switches {
                            owner := freeOwner
                            if s.Owner != freeOwner {
                                user, _ = api.GetUserInfo(s.Owner)
                                owner = user.RealName
                            }
                            text += fmt.Sprintf("%20.20s %-20.20s %s\n", s.Name, owner, s.Description)
                        }
                        text += "```"
                    } else {
                        text = "No resource added yet. Please refer to the add command to create one."
                    }
                    api.PostMessage(ev.Channel, slack.MsgOptionText(text, false))
                default:
                    api.PostMessage(ev.Channel, slack.MsgOptionText("Available commands are:\nlock <resource>\nunlock <resource>\nadd <resource> <description>\ndelete <resource>\nlist", false))
                    return events.APIGatewayProxyResponse{
                        StatusCode: http.StatusOK,
                    }, nil
                }
                return events.APIGatewayProxyResponse{
                    StatusCode: http.StatusOK,
                }, nil
            }
		}
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Invalid request"),
		StatusCode: http.StatusBadRequest,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
