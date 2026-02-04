package models

type AnalyticsConfig struct {
	Enabled bool
	S3      AnalyticsServiceConfig
}

type AnalyticsServiceConfig struct {
	Enabled  bool
	Features AnalyticsS3FeaturesConfig
}

type AnalyticsS3FeaturesConfig struct {
	ZombieBucket ZombieBucketFeatureConfig
}

type ZombieBucketFeatureConfig struct {
	Enabled bool
	Days    int
}
