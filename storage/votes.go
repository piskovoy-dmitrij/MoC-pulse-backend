package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
)

func (this *StorageConnection) NewVote(name string, owner string) (*Vote, error) {
	funcPrefix := "Inserting new vote to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	vote := &Vote{
		Name:  name,
		Date:  time.Now().Unix(),
		Owner: owner,
		Id:    id,
		Voted: false,
	}

	log.Debug.Printf("%s: marshaling vote...\n", funcPrefix)
	serialized, err := json.Marshal(vote)
	if err != nil {
		log.Error.Printf("%s: marshaling vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	err = this.client.Set("vote:"+id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return vote, nil
}

func (this *StorageConnection) GetVote(id string) (*Vote, error) {
	return this.LoadVote("vote:" + id)
}

func (this *StorageConnection) LoadVote(id string) (*Vote, error) {
	funcPrefix := "Getting vote from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting vote...\n", funcPrefix)
	val, err := this.client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting vote failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not found")
	}

	var vote Vote
	log.Debug.Printf("%s: unmarshaling vote...\n", funcPrefix)
	jsonString, _ := base64.StdEncoding.DecodeString(val)
	err = json.Unmarshal(jsonString, &vote)
	if err != nil {
		log.Error.Printf("%s: unmarshaling vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return &vote, nil
}

func (this *StorageConnection) GetAllVotesWithResult(user auth.User) ([]VoteWithResult, error) {
	funcPrefix := "Getting all votes from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting vote keys...\n", funcPrefix)
	voteKeys, count, err := this.getKeys("vote:*")
	if err != nil {
		log.Error.Printf("%s: getting vote keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var votes []VoteWithResult

	if count > 0 {
		log.Debug.Printf("%s: getting votes by keys...\n", funcPrefix)
		data, err := this.client.MGet(voteKeys...).Result()
		if err != nil {
			log.Error.Printf("%s: getting votes failed: %s\n", funcPrefix, err.Error())
			return nil, err
		}

		usersCount := this.UsersCount()
		
		log.Debug.Printf("%s: getting results of each vote by key...\n", funcPrefix)
		for _, value := range data {
			vote := &Vote{}
			jsonString, _ := base64.StdEncoding.DecodeString(value.(string))
			err = json.Unmarshal(jsonString, vote)
			if err != nil {
				log.Error.Printf("%s: unmarshaling vote failed: %s\n", funcPrefix, err.Error())
				return nil, err
			}
			item, error2 := this.GetVoteResultStatus(*vote, user, usersCount)
			if error2 == nil {
				votes = append(votes, item.Vote)
			}
		}
	}	

	return votes, nil
}

func (this *StorageConnection) VoteProcessing(vote Vote, user auth.User, value int) error {
	funcPrefix := "Saving result to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	voteResult := &VoteResult{
		Id:    vote.Id + ":" + user.Id,
		Value: value,
		Vote:  vote.Id,
		Date:  time.Now().UnixNano(),
	}

	log.Debug.Printf("%s: marshaling result...\n", funcPrefix)
	serialized, err := json.Marshal(voteResult)

	if err != nil {
		log.Error.Printf("%s: marshaling result failed: %s\n", funcPrefix, err.Error())
		return err
	}
	err = this.client.Set("result:"+voteResult.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting result failed: %s\n", funcPrefix, err.Error())
		return err
	}

	return nil
}

func (this *StorageConnection) LoadVoteResult(id string) (*VoteResult, error) {
	funcPrefix := "Getting result from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting result...\n", funcPrefix)
	data, err := this.client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting result failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not exist")
	}

	log.Debug.Printf("%s: unmarshaling vote result...\n", funcPrefix)
	voteResult := &VoteResult{}
	jsonString, _ := base64.StdEncoding.DecodeString(data)
	err = json.Unmarshal(jsonString, &voteResult)
	if err != nil {
		log.Error.Printf("%s: unmarshaling vote result failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return voteResult, nil
}

func (this *StorageConnection) IsVotedByUser(vote Vote, user auth.User) bool {
	funcPrefix := "Checking if user has voted"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	_, err := this.client.Get("result:" + vote.Id + ":" + user.Id).Result()
	if err != nil {
		log.Debug.Printf("%s: user '%s' has not voted in vote with id '%s' yet\n", funcPrefix, user.Id, vote.Id)
		return false
	}

	log.Debug.Printf("%s: user '%s' has already voted in vote with id '%s'\n", funcPrefix, user.Id, vote.Id)
	return true
}

func (this *StorageConnection) GetVoteResultStatus(vote Vote, user auth.User, usersCount int) (*VoteResultStatus, error) {
	funcPrefix := "Getting vote result status from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting vote result keys...\n", funcPrefix)
	resultKeys, count, err := this.getKeys("result:" + vote.Id + ":*")
	if err != nil {
		log.Error.Printf("%s: getting vote result keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var yellow int
	var red int
	var green int

	if count > 0 {
		log.Debug.Printf("%s: getting results by keys...\n", funcPrefix)
		data, err := this.client.MGet(resultKeys...).Result()
		if err != nil {
			log.Error.Printf("%s: getting results failed: %s\n", funcPrefix, err.Error())
			return nil, err
		}
		
		log.Debug.Printf("%s: getting vote result by result key...\n", funcPrefix)
		for _, item := range data {
			voteResult := &VoteResult{}
			jsonString, _ := base64.StdEncoding.DecodeString(item.(string))
			err = json.Unmarshal(jsonString, &voteResult)
			if err != nil {
				log.Error.Printf("%s: unmarshaling vote result failed: %s\n", funcPrefix, err.Error())
				return nil, err
			}
			switch voteResult.Value {
			case 0:
				red++
			case 1:
				yellow++
			default:
				green++
			}
		}
	}	

	log.Debug.Printf("%s: getting vote owner...\n", funcPrefix)
	ownerUser, error := this.LoadUser("user:" + vote.Owner)
	if error != nil {
		log.Error.Printf("%s: getting vote owner failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return &VoteResultStatus{
		Vote: VoteWithResult{
			Name:  vote.Name,
			Id:    vote.Id,
			Owner: *ownerUser,
			Date:  vote.Date,
			Voted: this.IsVotedByUser(vote, user),
			Result: Result{
				Yellow:    yellow,
				Green:     green,
				Red:       red,
				AllUsers:  usersCount,
				VoteUsers: count,
			},
		},
	}, nil
}
