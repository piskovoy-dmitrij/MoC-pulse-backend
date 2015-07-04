package redis

import (
	//	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

type Vote struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	owner string
	date  int64
}

type VoteResult struct {
	id    string
	value int
	vote  string
	date  int64
}

func NewVote(name string, owner string) *Vote {
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	return &Vote{
		Name:  name,
		date:  time.Now().UnixNano(),
		owner: owner,
		Id:    id,
	}
}
