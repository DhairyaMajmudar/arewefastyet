/*
 *
 * Copyright 2021 The Vitess Authors.
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

package server

import (
	"errors"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/vitessio/arewefastyet/go/slack"
	"github.com/vitessio/arewefastyet/go/storage/psdb"
	"github.com/vitessio/arewefastyet/go/tools/github"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ErrorIncorrectConfiguration = "incorrect configuration"

	flagPort                                 = "web-port"
	flagVitessPath                           = "web-vitess-path"
	flagMode                                 = "web-mode"
	flagCronSchedule                         = "web-cron-schedule"
	flagCronSchedulePullRequests             = "web-cron-schedule-pull-requests"
	flagCronScheduleTags                     = "web-cron-schedule-tags"
	flagPullRequestLabelTrigger              = "web-pr-label-trigger"
	flagPullRequestLabelTriggerWithPlannerV3 = "web-pr-label-trigger-planner-v3"
	flagCronNbRetry                          = "web-cron-nb-retry"
	flagBenchmarkConfigPath                  = "web-benchmark-config-path"
	flagFilterBySource                       = "web-source-filter"
	flagExcludeFilterBySource                = "web-source-exclude-filter"

	// keyMinimumVitessVersion is used to define on which minimum Vitess version a given
	// benchmark should be run. Only the major version is counted. This key/value is located
	// in the benchmarks' configuration files.
	keyMinimumVitessVersion = "minimum-version"
)

type benchmarkConfig struct {
	file string
	v    *viper.Viper
	skip bool
}

type Server struct {
	port   string
	router *gin.Engine

	vitessPathMu    sync.Mutex
	localVitessPath string

	dbCfg    *psdb.Config
	dbClient *psdb.Client

	// Configuration used to send message to Slack.
	slackConfig slack.Config

	cronSchedule             string
	cronSchedulePullRequests string
	cronScheduleTags         string
	cronNbRetry              int

	benchmarkConfigPath string

	prLabelTrigger   string
	prLabelTriggerV3 string

	benchmarkConfig map[string]benchmarkConfig
	benchmarkTypes  []string

	sourceFilter        []string
	excludeSourceFilter []string

	ghApp *github.App

	// Mode used to run the server.
	Mode
}

func (s *Server) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.port, flagPort, "8080", "Port used for the HTTP server")
	cmd.Flags().StringVar(&s.localVitessPath, flagVitessPath, "/", "Absolute path where the vitess directory is located or where it should be cloned")
	cmd.Flags().Var(&s.Mode, flagMode, "Specify the mode on which the server will run")

	// execution configuration flags
	cmd.Flags().StringVar(&s.benchmarkConfigPath, flagBenchmarkConfigPath, "", "Path to the configuration file folder for the benchmarks.")
	cmd.Flags().StringVar(&s.cronSchedule, flagCronSchedule, "@midnight", "Execution CRON schedule defaults to every day at midnight. An empty string will result in no CRON.")
	cmd.Flags().StringVar(&s.cronSchedulePullRequests, flagCronSchedulePullRequests, "*/5 * * * *", "Execution CRON schedule for pull requests benchmarks. An empty string will result in no CRON. Defaults to an execution every 5 minutes.")
	cmd.Flags().StringVar(&s.cronScheduleTags, flagCronScheduleTags, "*/1 * * * *", "Execution CRON schedule for tags/releases benchmarks. An empty string will result in no CRON. Defaults to an execution every minute.")
	cmd.Flags().IntVar(&s.cronNbRetry, flagCronNbRetry, 1, "Number of retries allowed for each cron job.")
	cmd.Flags().StringVar(&s.prLabelTrigger, flagPullRequestLabelTrigger, "Benchmark me", "GitHub Pull Request label that will trigger the execution of new execution.")
	cmd.Flags().StringVar(&s.prLabelTriggerV3, flagPullRequestLabelTriggerWithPlannerV3, "Benchmark me (V3)", "GitHub Pull Request label that will trigger the execution of new execution using the V3 planner.")
	cmd.Flags().StringSliceVar(&s.sourceFilter, flagFilterBySource, nil, "List of execution source that should be run. By default, all sources are ran.")
	cmd.Flags().StringSliceVar(&s.excludeSourceFilter, flagExcludeFilterBySource, nil, "List of execution source to not execute. By default, all sources are ran.")

	_ = viper.BindPFlag(flagPort, cmd.Flags().Lookup(flagPort))
	_ = viper.BindPFlag(flagVitessPath, cmd.Flags().Lookup(flagVitessPath))
	_ = viper.BindPFlag(flagMode, cmd.Flags().Lookup(flagMode))
	_ = viper.BindPFlag(flagCronSchedule, cmd.Flags().Lookup(flagCronSchedule))
	_ = viper.BindPFlag(flagCronSchedulePullRequests, cmd.Flags().Lookup(flagCronSchedulePullRequests))
	_ = viper.BindPFlag(flagCronScheduleTags, cmd.Flags().Lookup(flagCronScheduleTags))
	_ = viper.BindPFlag(flagCronNbRetry, cmd.Flags().Lookup(flagCronNbRetry))
	_ = viper.BindPFlag(flagPullRequestLabelTrigger, cmd.Flags().Lookup(flagPullRequestLabelTrigger))
	_ = viper.BindPFlag(flagPullRequestLabelTriggerWithPlannerV3, cmd.Flags().Lookup(flagPullRequestLabelTriggerWithPlannerV3))
	_ = viper.BindPFlag(flagFilterBySource, cmd.Flags().Lookup(flagFilterBySource))
	_ = viper.BindPFlag(flagExcludeFilterBySource, cmd.Flags().Lookup(flagExcludeFilterBySource))

	s.slackConfig.AddToCommand(cmd)
	if s.dbCfg == nil {
		s.dbCfg = &psdb.Config{}
	}
	s.dbCfg.AddToCommand(cmd)
	if s.ghApp == nil {
		s.ghApp = &github.App{}
	}
	s.ghApp.AddToCommand(cmd)
}

func (s *Server) isReady() bool {
	return s.port != "" && s.localVitessPath != ""
}

func (s *Server) Init() error {
	if !s.isReady() {
		return errors.New(ErrorIncorrectConfiguration)
	}

	if s.Mode != "" && !s.Mode.correct() {
		return errors.New(ErrorIncorrectMode)
	} else if s.Mode == "" {
		s.Mode.useDefault()
	}

	if slog == nil {
		err := s.initLogger()
		if err != nil {
			return err
		}
		defer cleanLogger()
	}

	if err := s.setupLocalVitess(); err != nil {
		return err
	}

	if err := s.createStorages(); err != nil {
		return err
	}

	s.benchmarkConfig = map[string]benchmarkConfig{
		// "micro":         {file: path.Join(s.benchmarkConfigPath, "micro.yaml"), v: viper.New(), skip: true},
		"oltp":          {file: path.Join(s.benchmarkConfigPath, "oltp.yaml"), v: viper.New()},
		// "oltp-set":      {file: path.Join(s.benchmarkConfigPath, "oltp-set.yaml"), v: viper.New()},
		// "oltp-readonly": {file: path.Join(s.benchmarkConfigPath, "oltp-readonly.yaml"), v: viper.New()},
		//
		// // TODO: oltp-readonly-olap benchmarks are skipped for now as they fail very often due to
		// // MySQL connections being dropped. This issue will be investigated soon.
		// "oltp-readonly-olap": {file: path.Join(s.benchmarkConfigPath, "olap-readonly.yaml"), v: viper.New(), skip: true},
		//
		// "tpcc": {file: path.Join(s.benchmarkConfigPath, "tpcc.yaml"), v: viper.New()},
	}
	for configName, config := range s.benchmarkConfig {
		config.v.SetConfigFile(config.file)
		if err := config.v.ReadInConfig(); err != nil {
			slog.Error(err)
		}
		if configName == "micro" {
			continue
		}
		s.benchmarkTypes = append(s.benchmarkTypes, strings.ToUpper(configName))
	}
	return nil
}

func (s *Server) Run() error {
	if !s.isReady() {
		return errors.New(ErrorIncorrectConfiguration)
	}

	err := s.createCrons()
	if err != nil {
		return err
	}

	err = s.ghApp.Init()
	if err != nil {
		return err
	}

	s.prepareGin()
	s.router = gin.Default()

	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API
	s.router.GET("/api/recent", s.getRecentExecutions)
	s.router.GET("/api/queue", s.getExecutionsQueue)
	s.router.GET("/api/vitess/refs", s.getLatestVitessGitRef)
	s.router.GET("/api/macrobench/compare", s.compareMacrobenchmarks)
	s.router.GET("/api/microbench/compare", s.compareMicrobenchmarks)
	s.router.GET("/api/search", s.searchBenchmarck)
	s.router.GET("/api/macrobench/compare/queries", s.queriesCompareMacrobenchmarks)
	s.router.GET("/api/pr/list", s.getPullRequest)
	s.router.GET("/api/pr/info/:nb", s.getPullRequestInfo)
	s.router.GET("/api/daily/summary", s.getDailySummary)
	s.router.GET("/api/daily", s.getDaily)
	s.router.GET("/api/status/stats", s.getStatusStats)

	return s.router.Run(":" + s.port)
}

func (s *Server) prepareGin() {
	switch s.Mode {
	case ProductionMode:
		gin.SetMode(gin.ReleaseMode)
	case DevelopmentMode:
		gin.SetMode(gin.DebugMode)
	}
}

func Run(port, localVitessPath string) error {
	s := Server{
		port:            port,
		localVitessPath: localVitessPath,
	}
	err := s.Init()
	if err != nil {
		return err
	}
	return s.Run()
}
