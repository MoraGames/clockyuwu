package structs

import (
	"fmt"
)

type User struct {
	TelegramID                      int64
	UserName                        string
	TotalPoints                     int
	TotalEventPartecipations        int
	TotalEventWins                  int
	TotalChampionshipPartecipations int
	TotalChampionshipWins           int
	Effects                         []*Effect
}

func NewUser(telegramID int64, username string) *User {
	return &User{telegramID, username, 0, 0, 0, 0, 0, make([]*Effect, 0)}
}

func (u *User) AddEffect(effectToAdd *Effect) {
	u.Effects = append(u.Effects, effectToAdd)
}

func (u *User) RemoveEffect(effectToRemove *Effect) {
	newUserEffects := make([]*Effect, 0)
	for _, userEffect := range u.Effects {
		if userEffect.Name != effectToRemove.Name {
			newUserEffects = append(newUserEffects, userEffect)
		}
	}
	u.Effects = newUserEffects
}

func (u *User) StringifyEffects() string {
	stringifiedEffects := ""
	for i, e := range u.Effects {
		if i != len(u.Effects)-1 {
			stringifiedEffects += fmt.Sprintf("%q, ", e.Name)
		} else {
			stringifiedEffects += fmt.Sprintf("%q", e.Name)
		}
	}
	return "[" + stringifiedEffects + "]"
}
