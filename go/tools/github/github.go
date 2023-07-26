/*
 *
 * Copyright 2023 The Vitess Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package github

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type App struct {
	appID          int
	webHookSecret  string
	secretKey      string
	port           string
	installationID int

	client *github.Client
}

const (
	flagAppID          = "gh-app-id"
	flagWebHookSecret  = "gh-webhook-secret"
	flagSecretKey      = "gh-secret-key"
	flagPort           = "gh-port"
	flagInstallationID = "gh-installation-id"
)

// AddToCommand ...
func (a *App) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().IntVar(&a.appID, flagAppID, 0, "xxx")
	cmd.Flags().StringVar(&a.webHookSecret, flagWebHookSecret, "", "xxx")
	cmd.Flags().StringVar(&a.secretKey, flagSecretKey, "", "xxx")
	cmd.Flags().StringVar(&a.port, flagPort, "8181", "xxx")
	cmd.Flags().IntVar(&a.installationID, flagInstallationID, 0, "xxx")

	_ = viper.BindPFlag(flagAppID, cmd.Flags().Lookup(flagAppID))
	_ = viper.BindPFlag(flagWebHookSecret, cmd.Flags().Lookup(flagWebHookSecret))
	_ = viper.BindPFlag(flagSecretKey, cmd.Flags().Lookup(flagSecretKey))
	_ = viper.BindPFlag(flagPort, cmd.Flags().Lookup(flagPort))
	_ = viper.BindPFlag(flagInstallationID, cmd.Flags().Lookup(flagInstallationID))
}

func (a *App) Init() error {
	// Create an authenticated client using go-githubapp
	config := githubapp.Config{
		V3APIURL: "https://api.github.com/",
		V4APIURL: "https://api.github.com/graphql",
		App: struct {
			IntegrationID int64  `yaml:"integration_id" json:"integrationId"`
			WebhookSecret string `yaml:"webhook_secret" json:"webhookSecret"`
			PrivateKey    string `yaml:"private_key" json:"privateKey"`
		}{
			IntegrationID: int64(a.appID),
			WebhookSecret: a.webHookSecret,
			PrivateKey:    a.secretKey,
		},
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	metricsRegistry := metrics.DefaultRegistry

	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientUserAgent("arewefastyet-bot/1.0.0"),
		githubapp.WithClientTimeout(5*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(metricsRegistry),
		),
	)
	if err != nil {
		return err
	}

	client, err := clientCreator.NewInstallationClient(int64(a.installationID))
	if err != nil {
		return err
	}
	a.client = client

	go func() {
		prHandler := pullRequestHandler{
			ClientCreator: clientCreator,
		}

		webhookHandler := githubapp.NewEventDispatcher(
			[]githubapp.EventHandler{prHandler},
			config.App.WebhookSecret,
			githubapp.WithScheduler(
				githubapp.AsyncScheduler(),
			),
		)

		http.Handle(githubapp.DefaultWebhookRoute, webhookHandler)

		err = http.ListenAndServe(":"+a.port, webhookHandler)
		if err != nil {
			logger.Error().Err(err).Msg("failed to start server")
		}
	}()

	return nil
}

type PRInfo struct {
	ID        int
	Author    string
	Title     string
	CreatedAt *time.Time
	Base      string
	Head      string
}

func (a *App) GetPullRequestInfo(prNumber int) (PRInfo, error) {
	ctx := context.Background()
	pr, _, err := a.client.PullRequests.Get(ctx, "vitessio", "vitess", prNumber)
	if err != nil {
		return PRInfo{}, err
	}

	createAt := pr.GetCreatedAt().Time
	return PRInfo{
		ID:        prNumber,
		Author:    pr.User.GetLogin(),
		Title:     pr.GetTitle(),
		CreatedAt: &createAt,
	}, nil
}