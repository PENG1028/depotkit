package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/config"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [service]",
	Short: "Print connection strings for all or a specific service",
	Long: `Print connection information for enabled services.

Services: postgres, redis, object, mongo

If no service is specified, connection strings for all enabled services are printed.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if len(args) == 0 {
			// Print all services
			fmt.Printf("Project: %s\n", cfg.Project)
			fmt.Println()

			if cfg.Services.Postgres.Enabled {
				printPostgresConnect(cfg)
				fmt.Println()
			}
			if cfg.Services.Redis.Enabled {
				printRedisConnect(cfg)
				fmt.Println()
			}
			if cfg.Services.Object.Enabled {
				printObjectConnect(cfg)
				fmt.Println()
			}
			if cfg.Services.Mongo.Enabled {
				printMongoConnect(cfg)
				fmt.Println()
			}
		} else {
			service := strings.ToLower(args[0])
			switch service {
			case "postgres", "pg":
				if !cfg.Services.Postgres.Enabled {
					ExitError("PostgreSQL is not enabled in config")
				}
				printPostgresConnect(cfg)
			case "redis":
				if !cfg.Services.Redis.Enabled {
					ExitError("Redis is not enabled in config")
				}
				printRedisConnect(cfg)
			case "object", "s3", "minio":
				if !cfg.Services.Object.Enabled {
					ExitError("Object storage is not enabled in config")
				}
				printObjectConnect(cfg)
			case "mongo", "mongodb":
				if !cfg.Services.Mongo.Enabled {
					ExitError("MongoDB is not enabled in config")
				}
				printMongoConnect(cfg)
			default:
				ExitError("Unknown service: %s (supported: postgres, redis, object, mongo)", service)
			}
		}
	},
}

func printPostgresConnect(cfg *config.Config) {
	host := "localhost"
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Services.Postgres.User,
		cfg.Services.Postgres.Password,
		host,
		cfg.Services.Postgres.Port,
		cfg.Services.Postgres.Database,
	)
	fmt.Println("PostgreSQL:")
	fmt.Printf("  DATABASE_URL=%s\n", url)
}

func printRedisConnect(cfg *config.Config) {
	url := fmt.Sprintf("redis://localhost:%d", cfg.Services.Redis.Port)
	fmt.Println("Redis:")
	fmt.Printf("  REDIS_URL=%s\n", url)
}

func printObjectConnect(cfg *config.Config) {
	fmt.Println("Object Storage (MinIO/S3):")
	fmt.Printf("  S3_ENDPOINT=http://localhost:%d\n", cfg.Services.Object.Port)
	fmt.Printf("  S3_ACCESS_KEY=%s\n", cfg.Services.Object.AccessKey)
	fmt.Printf("  S3_SECRET_KEY=%s\n", cfg.Services.Object.SecretKey)
	fmt.Printf("  S3_BUCKET=%s\n", cfg.Services.Object.Bucket)
	fmt.Printf("  CONSOLE_URL=http://localhost:%d\n", cfg.Services.Object.ConsolePort)
}

func printMongoConnect(cfg *config.Config) {
	url := fmt.Sprintf("mongodb://localhost:%d/%s", cfg.Services.Mongo.Port, cfg.Services.Mongo.Database)
	fmt.Println("MongoDB:")
	fmt.Printf("  MONGO_URL=%s\n", url)
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
