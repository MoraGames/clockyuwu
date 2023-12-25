package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MoraGames/clockyuwu/events"
	"github.com/MoraGames/clockyuwu/pkg/types"
	"github.com/MoraGames/clockyuwu/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// Rank is the type used for manage /ranking sorting
type Rank struct {
	Username       string
	Points         int
	Partecipations int
}

// switch for all the commands that the bot can receive
func manageCommands(update tgbotapi.Update, utils types.Utils, data types.Data, curTime time.Time, eventKey events.EventKey) {
	switch update.Message.Command() {
	case "ping":
		// Respond with a "pong" message. Useful for checking if the bot is online
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "pong")
		msg.ReplyToMessageID = update.Message.MessageID
		data.Bot.Send(msg)
		utils.Logger.WithFields(logrus.Fields{
			"usr": update.Message.From.UserName,
			"msg": update.Message.Text,
		}).Debug("Ping-Pong sent")
	case "ranking":
		// Respond with the ranking based on users' points
		// Generate the ranking
		ranking := make([]Rank, 0)
		for _, u := range Users {
			if u != nil {
				ranking = append(ranking, Rank{u.UserName, u.TotalPoints, u.TotalEventPartecipations})
			}
		}

		// Sort the ranking by points (and partecipations if points are equal)
		sort.Slice(
			ranking,
			func(i, j int) bool {
				if ranking[i].Points == ranking[j].Points {
					return ranking[i].Partecipations < ranking[j].Partecipations
				}
				return ranking[i].Points > ranking[j].Points
			},
		)

		// Generate the string to send
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ancora nessun utente ha partecipato agli eventi della season.")
		if len(ranking) != 0 {
			leadersPoints := ranking[0].Points
			rankingString := ""
			for i, r := range ranking {
				rankingString += fmt.Sprintf("%v] %v: %v (-%v)\n", i+1, r.Username, r.Points, leadersPoints-r.Points)
			}

			// Send the message
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("La classifica è la seguente:\n\n%v", rankingString))
		}
		msg.ReplyToMessageID = update.Message.MessageID
		data.Bot.Send(msg)

		// Log the /ranking command sent
		utils.Logger.Debug("Ranking sent")
	case "stats":
		// Respond with the user's stats (or the specified user's stats)

		// Check if the command has arguments
		if update.Message.CommandArguments() == "" {
			// Get the user from the Users data structure
			u := Users[update.Message.From.ID]

			// Check (and eventually update) the user effects
			UpdateUserEffects(update.Message.From.ID)

			// Send the message with user's stats
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Non hai ancora partecipato a nessun evento.")
			if u != nil {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Le tue statistiche sono:\n\nPunti totali: %v\nPartecipazioni totali: %v\nVittorie totali: %v\nEffetti attivi: %v", u.TotalPoints, u.TotalEventPartecipations, u.TotalEventWins, u.Effects))
			}
			msg.ReplyToMessageID = update.Message.MessageID
			data.Bot.Send(msg)

			// Log the /stats command sent
			utils.Logger.Debug("Stats sent")
		} else {
			// Split the command arguments
			cmdArgs := strings.Split(update.Message.CommandArguments(), " ")

			if len(cmdArgs) != 1 {
				// Respond with a message indicating that the command arguments are wrong
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Il comando è /stats [user]")
				msg.ReplyToMessageID = update.Message.MessageID
				data.Bot.Send(msg)
				utils.Logger.WithFields(logrus.Fields{
					"usr": update.Message.From.UserName,
					"msg": update.Message.Text,
				}).Debug("Wrong command")
			} else {
				// Get the id and check if the user exists
				username := cmdArgs[0]
				var userKey int64
				var founded bool
				for userID, user := range Users {
					if user.UserName == username {
						founded = true
						userKey = userID
					}
				}

				if !founded {
					// Respond with a message indicating that the user does not exist
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Utente non trovato")
					msg.ReplyToMessageID = update.Message.MessageID
					data.Bot.Send(msg)
					utils.Logger.WithFields(logrus.Fields{
						"usr": update.Message.From.UserName,
						"msg": update.Message.Text,
					}).Debug("User not found")
				} else {
					// Get the user from the Users data structure
					u := Users[userKey]

					// Check (and eventually update) the user effects
					UpdateUserEffects(userKey)

					// Send the message with user's stats
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v non ha ancora partecipato a nessun evento.", username))
					if u != nil {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Le tue statistiche sono:\n\nPunti totali: %v\nPartecipazioni totali: %v\nVittorie totali: %v\nEffetti attivi: %v", u.TotalPoints, u.TotalEventPartecipations, u.TotalEventWins, u.Effects))
					}
					msg.ReplyToMessageID = update.Message.MessageID
					data.Bot.Send(msg)

					// Log the /stats command sent
					utils.Logger.Debug("Stats sent")
				}
			}
		}

		// Log the /stats command sent
		utils.Logger.Debug("Stats sent")
	case "reset":
		// Reset the events or users data structure
		// Check if the user is an bot-admin
		if !isAdmin(update.Message.From, utils) {
			// Respond and log with a message indicating that the user is not authorized to use this command
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Non sei autorizzato ad usare questo comando")
			msg.ReplyToMessageID = update.Message.MessageID
			data.Bot.Send(msg)
			utils.Logger.WithFields(logrus.Fields{
				"usr": update.Message.From.UserName,
				"cmd": update.Message.Command(),
			}).Debug("Unauthorized user")
		} else {
			// Split the command arguments
			cmdArgs := strings.Split(update.Message.CommandArguments(), " ")

			// Check if the command arguments are in the form /reset <events|users>
			if len(cmdArgs) != 1 {
				// Respond with a message indicating that the command arguments are wrong
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Il comando è /reset <events|users>")
				msg.ReplyToMessageID = update.Message.MessageID
				data.Bot.Send(msg)
				utils.Logger.WithFields(logrus.Fields{
					"usr": update.Message.From.UserName,
					"msg": update.Message.Text,
				}).Debug("Wrong command")
			} else {
				// Check if the command argument is events or users
				switch cmdArgs[0] {
				case "events":
					// Reset the events data structure
					events.Events.Reset(true, types.WriteMessageData{Bot: data.Bot, ChatID: update.Message.Chat.ID, ReplyMessageID: update.Message.MessageID})

					// Respond with command executed successfully
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Eventi resettati")
					msg.ReplyToMessageID = update.Message.MessageID
					data.Bot.Send(msg)

					// Log the /reset command sent
					utils.Logger.Debug("Events resetted")
				case "users":
					// Reset the users data structure
					Users = make(map[int64]*structs.User)

					// Overwrite the users.json file with the new (and empty) data structure
					file, err := json.MarshalIndent(Users, "", " ")
					if err != nil {
						utils.Logger.WithFields(logrus.Fields{
							"err":  err,
							"note": "preoccupati",
						}).Error("Error while marshalling data")
						utils.Logger.Error(Users)
					}
					err = os.WriteFile("./users.json", file, 0644)
					if err != nil {
						utils.Logger.WithFields(logrus.Fields{
							"err":  err,
							"note": "preoccupati tanto",
						}).Error("Error while writing data")
						utils.Logger.Error(Users)
					}

					// Respond with command executed successfully
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Utenti resettati")
					msg.ReplyToMessageID = update.Message.MessageID
					data.Bot.Send(msg)

					// Log the /reset command sent
					utils.Logger.Debug("Users resetted")
				default:
					// Respond with a message indicating that the command arguments are wrong
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Il comando è /reset <events|users>")
					msg.ReplyToMessageID = update.Message.MessageID
					data.Bot.Send(msg)

					// Log the /reset command executed in a wrong form
					utils.Logger.WithFields(logrus.Fields{
						"usr": update.Message.From.UserName,
						"cmd": update.Message.Command(),
					}).Debug("Wrong command")
				}
			}
		}
	case "update":
		// Update points value property of an event
		// Check if the user is an bot-admin
		if !isAdmin(update.Message.From, utils) {
			// Respond and log with a message indicating that the user is not authorized to use this command
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Non sei autorizzato ad usare questo comando")
			msg.ReplyToMessageID = update.Message.MessageID
			data.Bot.Send(msg)
			utils.Logger.WithFields(logrus.Fields{
				"usr": update.Message.From.UserName,
				"cmd": update.Message.Command(),
			}).Debug("Unauthorized user")
		} else {
			// Split the command arguments
			cmdArgs := strings.Split(update.Message.CommandArguments(), " ")

			// Check if the command arguments are in the form /update <event|user> <points>
			if len(cmdArgs) != 3 {
				// Respond with a message indicating that the command arguments are wrong
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Il comando è /update (<\"event\" event>|\"user\" <user>) <points>")
				msg.ReplyToMessageID = update.Message.MessageID
				data.Bot.Send(msg)
				utils.Logger.WithFields(logrus.Fields{
					"usr": update.Message.From.UserName,
					"msg": update.Message.Text,
				}).Debug("Wrong command")
			} else {
				// Get and check if the event exists
				targetType := cmdArgs[0]
				switch targetType {
				case "event":
					eventKeyString := cmdArgs[1]
					if event, ok := events.Events[events.EventKey(eventKeyString)]; !ok {
						// Respond with a message indicating that the event does not exist
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Evento non trovato")
						msg.ReplyToMessageID = update.Message.MessageID
						data.Bot.Send(msg)
						utils.Logger.WithFields(logrus.Fields{
							"usr": update.Message.From.UserName,
							"msg": update.Message.Text,
						}).Debug("Event not found")
					} else {
						// Get and check if the points value is a number
						points, err := strconv.Atoi(cmdArgs[2])
						if err != nil {
							// Respond with a message indicating that the points value is not a number
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "parametro points deve essere un numero.")
							msg.ReplyToMessageID = update.Message.MessageID
							data.Bot.Send(msg)
							utils.Logger.WithFields(logrus.Fields{
								"usr": update.Message.From.UserName,
								"msg": update.Message.Text,
							}).Debug("Wrong command")
						} else {
							// Update the event points value
							events.Events[events.EventKey(eventKeyString)] = &events.EventValue{Points: points, Activated: event.Activated, ActivatedBy: event.ActivatedBy, ActivatedAt: event.ActivatedAt, ArrivedAt: event.ArrivedAt, Partecipations: event.Partecipations}

							// Respond with command executed successfully
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Evento aggiornato")
							msg.ReplyToMessageID = update.Message.MessageID
							data.Bot.Send(msg)

							// Log the /update command executed successfully
							utils.Logger.Debug("Event updated")
						}
					}
				case "user":
					username := cmdArgs[1]
					var userKey int64
					for userID, user := range Users {
						if user != nil && user.UserName == username {
							userKey = userID
						}
					}

					if user, ok := Users[userKey]; !ok {
						// Respond with a message indicating that the event does not exist
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Utente non trovato")
						msg.ReplyToMessageID = update.Message.MessageID
						data.Bot.Send(msg)
						utils.Logger.WithFields(logrus.Fields{
							"usr": update.Message.From.UserName,
							"msg": update.Message.Text,
						}).Debug("Event not found")
					} else {
						// Get and check if the points value is a number
						points, err := strconv.Atoi(cmdArgs[2])
						if err != nil {
							// Respond with a message indicating that the points value is not a number
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "parametro points deve essere un numero.")
							msg.ReplyToMessageID = update.Message.MessageID
							data.Bot.Send(msg)
							utils.Logger.WithFields(logrus.Fields{
								"usr": update.Message.From.UserName,
								"msg": update.Message.Text,
							}).Debug("Wrong command")
						} else {
							// Update the user points value
							Users[userKey] = &structs.User{UserName: user.UserName, TotalPoints: points, TotalEventPartecipations: user.TotalEventPartecipations, TotalEventWins: user.TotalEventWins, TotalChampionshipPartecipations: user.TotalChampionshipPartecipations, TotalChampionshipWins: user.TotalChampionshipWins}

							// Respond with command executed successfully
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Utente aggiornato")
							msg.ReplyToMessageID = update.Message.MessageID
							data.Bot.Send(msg)

							// Log the /update command executed successfully
							utils.Logger.Debug("Event updated")
						}
					}
				}
			}
		}
	}
}

// Check if a user is considered bot-admin (saved in .env file)
func isAdmin(user *tgbotapi.User, utils types.Utils) bool {
	adminUserIDStr := os.Getenv("TELEGRAM_ADMIN_ID")
	if adminUserIDStr == "" {
		utils.Logger.WithFields(logrus.Fields{
			"env": "TELEGRAM_ADMIN_ID",
		}).Panic("Env not set")
	}
	adminUserID, err := strconv.ParseInt(adminUserIDStr, 10, 64)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Error while parsing AdminUserIDStr")
	}

	return user.ID == adminUserID
}
