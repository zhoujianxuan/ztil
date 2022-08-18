package main

import (
	"fmt"
	"hash/crc32"

	"github.com/urfave/cli/v2"
)

func NewCrc32() *cli.Command {
	return &cli.Command{
		Name:    "crc32",
		Usage:   "crc32 <key>",
		Aliases: []string{"crc"},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				fmt.Println("Periodic incomplete")
				return nil
			}
			key := c.Args().Get(0)
			var i = crc32.ChecksumIEEE([]byte(key)) % 10
			fmt.Println("hash 的结果: ", i)
			return nil
		}}
}
