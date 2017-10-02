package main

import (
	"encoding/base64"
	"fmt"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"
)

type Puller struct {
	SubscriptionsService *pubsub.ProjectsSubscriptionsService
	Fqn                  string
	Interval             int
}

func (p *Puller) Follow() error {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: false,
		MaxMessages:       1,
	}
	for {
		err := p.Execute(pullRequest)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(p.Interval) * time.Second)
	}
}

func (p *Puller) Execute(pullRequest *pubsub.PullRequest) error {
	res, err := p.SubscriptionsService.Pull(p.Fqn, pullRequest).Do()
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
		ackRequest := &pubsub.AcknowledgeRequest{
			AckIds: []string{recvMsg.AckId},
		}
		_, err = p.SubscriptionsService.Acknowledge(p.Fqn, ackRequest).Do()
		if err != nil {
			fmt.Printf("Failed to Acknowledge to %v cause of %v\n", p.Fqn, err)
			return err
		}
	}
	return nil
}
