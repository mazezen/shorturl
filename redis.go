package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mazezen/base62"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	URLIDKEY           = "next.url.id"
	SHORTLINKKEY       = "shortlink:%s:url"
	URLHASHKEY         = "urlhash:%s:url"
	SHORTLINKDETAILKEY = "shortlink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func toSha1(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func NewRedisClient(addr string, password string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisCli{Cli: c}
}

func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	h := toSha1(url)

	d, err := r.Cli.Get(fmt.Sprintf(URLHASHKEY, h)).Result()
	fmt.Println(err)
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	if d != "" {
		return d, nil
	}

	r.Cli.Incr(URLIDKEY)

	id, err := r.Cli.Get(URLIDKEY).Int64()
	if err != nil {
		return "", err
	}

	sl := base62.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(id))))
	//eid := Encode(uint64(id))

	err = r.Cli.Set(fmt.Sprintf(SHORTLINKKEY, sl), url, 30*time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(URLHASHKEY, h), sl, 30*time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// 存储详细信息
	detail := URLDetail{
		URL:                 url,
		CreatedAt:           time.Now().String(),
		ExpirationInMinutes: time.Duration(exp),
	}
	detailJSON, _ := json.Marshal(detail)
	err = r.Cli.Set(fmt.Sprintf(SHORTLINKDETAILKEY, sl), detailJSON, time.Duration(exp)*time.Minute*30).Err()
	if err != nil {
		return "", err
	}

	return sl, nil
}

func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	fmt.Println(fmt.Sprintf(SHORTLINKDETAILKEY, eid))
	detailJSON, err := r.Cli.Get(fmt.Sprintf(SHORTLINKDETAILKEY, eid)).Result()
	if err != nil {
		return nil, err
	}

	var detail URLDetail
	err = json.Unmarshal([]byte(detailJSON), &detail)
	if err != nil {
		return nil, err
	}

	return detail, nil
}

func (r *RedisCli) Unshorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(SHORTLINKKEY, eid)).Result()
	if err != nil {
		return "", err
	}
	return url, nil
}
