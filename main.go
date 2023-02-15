package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func NewV50() *cli.Command {
	return &cli.Command{Name: "v50", Action: func(c *cli.Context) error {
		fmt.Println("KFC Crazy Thursday V me 50")
		return nil
	}}
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			NewV50(),
			NewKafkaCommand(),
			NewKafkaConsumerCommand(),
			NewMySQLCommand(),
			NewRedisCommand(),
			NewCrc32(),
			NewWXRobot(),
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
