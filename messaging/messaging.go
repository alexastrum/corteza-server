package messaging

import (
	"context"

	"github.com/go-chi/chi"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/titpetric/factory"

	migrate "github.com/cortezaproject/corteza-server/messaging/db"
	"github.com/cortezaproject/corteza-server/messaging/internal/service"
	"github.com/cortezaproject/corteza-server/messaging/rest"
	"github.com/cortezaproject/corteza-server/messaging/websocket"
	"github.com/cortezaproject/corteza-server/pkg/cli"
	"github.com/cortezaproject/corteza-server/pkg/cli/flags"
)

const (
	messaging = "messaging"
)

func Configure() *cli.Config {
	var (
		// Messaging API Server specific
		websocketOpt *flags.WebsocketOpt

		// Websocket handler
		ws *websocket.Websocket

		accessControlSetup = func(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
			// Calling grant directly on internal permissions service to avoid AC check for "grant"
			var p = service.DefaultPermissions
			var ac = service.DefaultAccessControl
			return p.Grant(ctx, ac.Whitelist(), ac.DefaultRules()...)
		}
	)

	return &cli.Config{
		ServiceName: messaging,

		RootCommandPreRun: cli.Runners{
			func(ctx context.Context, cmd *cobra.Command, c *cli.Config) (err error) {
				if c.ProvisionOpt.MigrateDatabase {
					cli.HandleError(c.ProvisionMigrateDatabase.Run(ctx, cmd, c))
				}

				cli.HandleError(service.Init(ctx, c.Log))

				ws = websocket.Init(ctx, &websocket.Config{
					Timeout:     websocketOpt.Timeout,
					PingTimeout: websocketOpt.PingTimeout,
					PingPeriod:  websocketOpt.PingPeriod,
				})

				if c.ProvisionOpt.AutoSetup {
					cli.HandleError(accessControlSetup(ctx, cmd, c))
				}

				return
			},
		},

		ApiServerAdtFlags: cli.FlagBinders{
			func(cmd *cobra.Command, c *cli.Config) {
				websocketOpt = flags.Websocket(cmd, messaging)
			},
		},

		ApiServerPreRun: cli.Runners{
			func(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
				go service.Watchers(ctx)
				return nil
			},
		},

		ApiServerRoutes: cli.Mounters{
			rest.MountRoutes,
			// Wrap in func() to assure ws is set when mounted
			func(r chi.Router) { ws.ApiServerRoutes(r) },
		},

		ProvisionMigrateDatabase: cli.Runners{
			func(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
				var db, err = factory.Database.Get(messaging)
				if err != nil {
					return err
				}

				db = db.With(ctx)
				// Disable profiler for migrations
				db.Profiler = nil

				return migrate.Migrate(db, c.Log)
			},
		},

		ProvisionAccessControl: cli.Runners{
			accessControlSetup,
		},
	}
}