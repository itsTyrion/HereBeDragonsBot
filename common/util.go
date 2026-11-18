package common

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func NewCooldown(defaultDuration time.Duration) *Cooldown {
	return &Cooldown{
		defaultDuration: defaultDuration,
		users:           make(map[snowflake.ID]bool),
	}
}

type Cooldown struct {
	defaultDuration time.Duration
	users           map[snowflake.ID]bool
}

func (cd *Cooldown) Add(id snowflake.ID) bool {
	return cd.AddD(id, time.Duration(cd.defaultDuration)*time.Second)
}

func (cd *Cooldown) AddD(id snowflake.ID, d time.Duration) bool {
	if cd.users[id] {
		return false
	}
	cd.users[id] = true
	go func() {
		time.Sleep(d)
		delete(cd.users, id)
	}()
	return true
}
