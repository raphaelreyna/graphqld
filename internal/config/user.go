package config

import (
	"os/user"
	"strconv"

	"github.com/rs/zerolog/log"
)

type User struct {
	Name    string
	HomeDir string
	Uid     uint32
	Gid     uint32
}

func userFromName(name string) *User {
	var (
		u User
	)

	uu, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to get current user from the os")
	}

	// non root users cant set uid
	if uu.Uid != "0" {
		return nil
	}

	if name != "" {
		uu, err = user.Lookup(name)
		if err != nil {
			log.Fatal().Err(err).
				Str("user", name).
				Msg("unable to get user from the os")
		}
	} else {
		return nil
	}

	var uid, gid uint64
	if uid, err = strconv.ParseUint(uu.Uid, 10, 32); err != nil {
		log.Fatal().Err(err).
			Msg("this os does not support uint32 user ids")
	}
	if gid, err = strconv.ParseUint(uu.Gid, 10, 32); err != nil {
		log.Fatal().Err(err).
			Msg("this os does not support uint32 group ids")
	}

	u.Uid = uint32(uid)
	u.Gid = uint32(gid)
	u.Name = name
	u.HomeDir = uu.HomeDir

	return &u
}
