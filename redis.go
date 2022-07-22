package main

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/urfave/cli/v2"
)

func getRedisClient(addr, pass string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})
}

func NewRedisCommand() *cli.Command {
	return &cli.Command{
		Name:      "redis",
		Aliases:   []string{"r"},
		Usage:     "redis",
		UsageText: "redis <addr> <db> <operate> <key> [pass]. \noperate:Get",
		Action: func(c *cli.Context) error {
			if c.NArg() < 4 {
				fmt.Println("Periodic incomplete")
				return nil
			}

			addr := c.Args().Get(0)
			dbStr := c.Args().Get(1)
			db, err := strconv.Atoi(dbStr)
			if err != nil {
				return err
			}

			operate := c.Args().Get(2)
			key := c.Args().Get(3)
			var pass string
			if c.NArg() == 5 {
				pass = c.Args().Get(4)
			}

			var r string
			client := getRedisClient(addr, pass, db)
			switch operate {
			case "Get":
				r, err = client.Get(key).Result()
				if err != nil {
					return err
				}
			default:
				fmt.Println("当前不支持该操作类型")
			}
			fmt.Println(r)
			return nil
		},
	}
}
