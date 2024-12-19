package redisx

import (
	"github.com/redis/go-redis/v9"
	"github.com/whitekid/goxp/log"
)

var logger = log.Named("redisx")

type Client struct {
	*redis.Client
}

func New(c *redis.Client) *Client { return &Client{c} }
