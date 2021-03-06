package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"
)

type Puller struct {
	SubscriptionsService *pubsub.ProjectsSubscriptionsService
	Ack                  bool
	Follow               bool
	Fqn                  string
	Interval             int
	MaxMessages          int64
	ReturnImmediately    bool
	Verbose              bool
	pullRequest          *pubsub.PullRequest
}

func (p *Puller) Setup() {
	p.pullRequest = &pubsub.PullRequest{
		ReturnImmediately: p.ReturnImmediately,
		MaxMessages:       p.MaxMessages,
	}

	if p.Verbose {
		p.ShowFields()
	}
}

func (p *Puller) ShowFields() {
	fmt.Fprintf(os.Stderr, "Puller\n"+
		"  Ack              : %v\n"+
		"  Follow           : %v\n"+
		"  Fqn              : %v\n"+
		"  Interval         : %v\n"+
		"  MaxMessages      : %v\n"+
		"  ReturnImmediately: %v\n"+
		"  Verbose          : %v\n"+
		"  pullRequest      : \n"+
		"    MaxMessages: %v\n"+
		"    ReturnImmediately: %v\n"+
		"",
		p.Ack, p.Follow, p.Fqn, p.Interval, p.MaxMessages, p.ReturnImmediately, p.Verbose,
		p.pullRequest.MaxMessages, p.pullRequest.ReturnImmediately)
}

func (p *Puller) Run() error {
	if p.Follow {
		return p.Subscribe()
	} else {
		return p.Execute()
	}
}

func (p *Puller) Subscribe() error {
	for {
		err := p.Execute()
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(p.Interval) * time.Second)
	}
}

func (p *Puller) Execute() error {
	res, err := p.SubscriptionsService.Pull(p.Fqn, p.pullRequest).Do()
	if err != nil {
		fmt.Printf("Failed to pull from %v cause of %v\n", p.Fqn, err)
		return err
	}

	for _, recvMsg := range res.ReceivedMessages {
		m := recvMsg.Message
		var decodedData string
		decoded, err := base64.StdEncoding.DecodeString(m.Data)
		if err != nil {
			decodedData = fmt.Sprintf("Failed to decode data by base64 because of %v", err)
		} else {
			decodedData = string(decoded)
		}
		fmt.Printf("%v %s: %v %s\n", m.PublishTime, m.MessageId, m.Attributes, decodedData)

		if p.Ack {
			err = p.Acknowledge(recvMsg.AckId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Puller) Acknowledge(ackId string) error {
	ackRequest := &pubsub.AcknowledgeRequest{
		AckIds: []string{ackId},
	}
	_, err := p.SubscriptionsService.Acknowledge(p.Fqn, ackRequest).Do()
	if err != nil {
		fmt.Printf("Failed to Acknowledge to %v cause of %v\n", p.Fqn, err)
		return err
	}
	return nil
}
