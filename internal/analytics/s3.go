package analytics

import (
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/inayathulla/cloudrift/internal/models"
)

// RunAnalytics executes all enabled S3 analytics features.
func RunAnalytics(cfg sdkaws.Config, analyticsConf models.AnalyticsConfig, buckets []models.S3Bucket) {
	if !(analyticsConf.Enabled && analyticsConf.S3.Enabled) {
		return // S3 disabled
	}

	if analyticsConf.S3.Features.ZombieBucket.Enabled {
		zombieDays := analyticsConf.S3.Features.ZombieBucket.Days
		if zombieDays == 0 {
			fmt.Println("Analytics: zombie_bucket.days not set or zero, skipping zombie bucket analytics")
			return
		}
		runZombieBucketAnalytics(cfg, buckets, zombieDays)
	}

	// Future S3 analytics features can be added here similarly.
}

func runZombieBucketAnalytics(cfg sdkaws.Config, buckets []models.S3Bucket, zombieDays int) {
	for _, bucket := range buckets {
		err := checkZombieBucket(cfg, bucket.Name, zombieDays)
		if err != nil {
			fmt.Printf("Analytics error on bucket %s: %v\n", bucket.Name, err)
		}
	}
}

// checkZombieBucket is your existing helper for checking last modified time and printing results.
func checkZombieBucket(cfg sdkaws.Config, bucketName string, zombieDays int) error {
	// Implement your existing logic here:
	// - list objects (limited)
	// - find last modified
	// - compare with zombieDays threshold
	// - print output
	return nil
}
