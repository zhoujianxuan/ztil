package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

type TextBody struct {
	Content             string   `json:"content"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type DDMsg struct {
	MsgType string   `json:"msgtype"`
	Text    TextBody `json:"text"`
}

func NewWXRobot() *cli.Command {
	return &cli.Command{
		Name:      "wx robot",
		Aliases:   []string{"wr"},
		Usage:     "wr",
		UsageText: "wr [msg] [robot url]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				fmt.Println("Check the parameters???")
				return nil
			}

			url := c.Args().Get(0)
			msg := c.Args().Get(1)

			ddMsg, _ := json.Marshal(&DDMsg{
				Text:    TextBody{Content: msg},
				MsgType: "text",
			})
			_, err := http.Post(url, "application/json", bytes.NewReader(ddMsg))
			if err != nil {
				return err
			}
			return nil
		},
	}
}
