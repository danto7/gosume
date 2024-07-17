package gosume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
)

func NewClientFromConfig(sourceConfig *aws.Config) Client {
	return Client{
		config:      sourceConfig,
		SessionName: "gosume",
	}
}

type Client struct {
	SessionName string
	config      *aws.Config
	credentials *types.Credentials
}

func (c *Client) AssumeRole(ctx context.Context, targetRoleArn string) (*types.Credentials, error) {
	log.Debug("assuming arn", "targetArn", targetRoleArn)
	stsClient := sts.NewFromConfig(*c.config)
	session, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &targetRoleArn,
		RoleSessionName: &c.SessionName,
	})
	if err != nil {
		return nil, fmt.Errorf("could not assume roleArn %s: \n\t%w", targetRoleArn, err)
	}
	cfg, err := createConfigFromStaticCredentials(ctx, session.Credentials)
	c.config = cfg
	return session.Credentials, err
}

func (c *Client) Config() aws.Config {
	return c.config.Copy()
}
