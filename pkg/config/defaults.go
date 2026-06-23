package config

// DefaultConfig returns the default configuration for a new project.
func DefaultConfig(projectName string) *Config {
	return &Config{
		Project: projectName,
		Runtime: RuntimeConfig{
			Mode:        "docker",
			ComposeFile: ".datadock/runtime/docker-compose.yml",
		},
		Services: ServiceConfig{
			Postgres: PostgresService{
				Enabled:       true,
				Image:         "postgres:16",
				ContainerName: projectName + "-postgres",
				Port:          5432,
				Database:      "app",
				User:          "app",
				Password:      "app_password",
				Volume:        projectName + "_postgres_data",
			},
			Redis: RedisService{
				Enabled:       true,
				Image:         "redis:7",
				ContainerName: projectName + "-redis",
				Port:          6379,
				Volume:        projectName + "_redis_data",
			},
			Object: ObjectService{
				Enabled:       true,
				Provider:      "minio",
				Image:         "minio/minio",
				ContainerName: projectName + "-minio",
				Port:          9000,
				ConsolePort:   9001,
				AccessKey:     "minio",
				SecretKey:     "minio_password",
				Bucket:        "app",
				Volume:        projectName + "_minio_data",
			},
			Mongo: MongoService{
				Enabled:       true,
				Image:         "mongo:7",
				ContainerName: projectName + "-mongo",
				Port:          27017,
				Database:      "app",
				Volume:        projectName + "_mongo_data",
			},
		},
		Postgres: PostgresConfig{
			Migrations: "db/migrations/postgres",
			Schema:     "db/schema/postgres/latest.sql",
			Backups:    ".datadock/backups/postgres",
		},
		Mongo: MongoConfig{
			Backups: ".datadock/backups/mongo",
		},
		Object: ObjectConfig{
			BackupPrefix: "backups/",
		},
	}
}
