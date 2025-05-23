package model

type User struct {
	Name     string
	Played   int
	Won      int
	Attempts int
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (u *User) RecordGame(won bool, attempts int) {
	u.Played++
	if won {
		u.Won++
	}
	u.Attempts += attempts
}

func (u *User) Stats() (games int, wins int, avg float64) {
	if u.Played == 0 {
		return 0, 0, 0
	}
	return u.Played, u.Won, float64(u.Attempts) / float64(u.Played)
}
