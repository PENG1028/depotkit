package mongo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// VersionCount represents the count of documents for each schemaVersion.
type VersionCount struct {
	Version string
	Count   int64
}

// SchemaVersionReport returns the distribution of schemaVersion values in a collection.
func SchemaVersionReport(db *mongo.Database, collection string) ([]VersionCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	coll := db.Collection(collection)

	// Aggregate to get schemaVersion distribution
	pipeline := bson.A{
		bson.M{
			"$group": bson.M{
				"_id":   "$schemaVersion",
				"count": bson.M{"$sum": 1},
			},
		},
		bson.M{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregating schemaVersion for %s: %w", collection, err)
	}
	defer cursor.Close(ctx)

	var results []VersionCount
	totalWithVersion := int64(0)

	for cursor.Next(ctx) {
		var result struct {
			ID    interface{} `bson:"_id"`
			Count int64       `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decoding aggregation result: %w", err)
		}

		versionStr := "null"
		if result.ID != nil {
			switch v := result.ID.(type) {
			case float64:
				versionStr = fmt.Sprintf("%.0f", v)
			case int32:
				versionStr = fmt.Sprintf("%d", v)
			case int64:
				versionStr = fmt.Sprintf("%d", v)
			default:
				versionStr = fmt.Sprintf("%v", v)
			}
		}

		results = append(results, VersionCount{Version: versionStr, Count: result.Count})
		totalWithVersion += result.Count
	}

	return results, nil
}

// CountDocuments counts total documents in a collection.
func CountDocuments(db *mongo.Database, collection string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := db.Collection(collection)
	return coll.EstimatedDocumentCount(ctx)
}

// Backup runs mongodump for the database.
func Backup(host string, port int, database, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	outputDir := backupDir + "/" + timestamp

	cmd := exec.Command("mongodump",
		"--host", host,
		"--port", fmt.Sprintf("%d", port),
		"--db", database,
		"--out", outputDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mongodump failed: %w\n%s", err, string(output))
	}

	return outputDir, nil
}

// Restore runs mongorestore for the database.
func Restore(host string, port int, database, backupDir string, drop bool) error {
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory not found: %s", backupDir)
	}

	args := []string{
		"--host", host,
		"--port", fmt.Sprintf("%d", port),
		"--db", database,
	}

	if drop {
		args = append(args, "--drop")
	}

	args = append(args, backupDir)

	cmd := exec.Command("mongorestore", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mongorestore failed: %w\n%s", err, string(output))
	}

	return nil
}
