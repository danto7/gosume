package main

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {

	envConf, err := config.NewEnvConfig()
	if err != nil {
		panic(err)
	}

	slog.Info("envConf", slog.String("conf", envConf.SharedConfigFile), slog.String("creds", envConf.SharedCredentialsFile))

	ctx := context.TODO()
	conf, err := config.LoadSharedConfigProfile(ctx, "test-source",
		func(lsco *config.LoadSharedConfigOptions) {
			lsco.ConfigFiles = []string{envConf.SharedConfigFile}
			lsco.CredentialsFiles = []string{envConf.SharedCredentialsFile}
		})

	// conf, err := config.LoadDefaultConfig(ctx,
	// 	config.WithLogger(logging.NewStandardLogger(os.Stdout)),
	// 	config.WithSharedConfigProfile("gosume-test-source"))
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info(fmt.Sprintf("STS %#v", conf))
}
