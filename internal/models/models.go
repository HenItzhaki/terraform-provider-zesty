package models

import "time"

type (
	OnboardingStatus string
	CloudProvider    string
	Product          string
)

const (
	AWS   CloudProvider = "AWS"
	Azure CloudProvider = "Azure"
	GCP   CloudProvider = "GCP"

	Kompass   Product = "Kompass"
	CM        Product = "CM"
	ZestyDisk Product = "ZestyDisk"

	DefaultHostURL string = "https://api.zesty.co/kompass-platform"
)

type ProductDetails struct {
	Active bool `json:"active" dynamodbav:"active"`
}

type CurDetails struct {
	S3Bucket   string `json:"s3Bucket"`
	ExportName string `json:"exportName"`
	Type       string `json:"type"`
}

type AthenaDetails struct {
	AthenaDB        string `json:"athenaDB"`
	AthenaS3Bucket  string `json:"athenaS3Bucket"`
	AthenaProjectID string `json:"athenaProjectID"`
	AthenaRegion    string `json:"athenaRegion"`
	AthenaTable     string `json:"athenaTable"`
	AthenaWorkgroup string `json:"athenaWorkgroup"`
	AthenaCatalog   string `json:"athenaCatalog"`
}

type Payload struct {
	AccountID     string                     `json:"accountID"`
	CloudProvider CloudProvider              `json:"cloudProvider"`
	Region        *string                    `json:"region,omitempty"`
	RoleARN       string                     `json:"roleARN"`
	ExternalID    string                     `json:"externalID"`
	Products      map[Product]ProductDetails `json:"products"`
	Cur           *CurDetails                `json:"cur,omitempty"`
	Athena        *AthenaDetails             `json:"athena,omitempty"`
}

type Account struct {
	OrganizationID   int64
	OnboardingStatus OnboardingStatus
	AccountID        string
	Region           *string
	CloudProvider    CloudProvider
	Products         map[Product]ProductDetails
	Cur              *CurDetails
	Athena           *AthenaDetails

	CreatedAt      time.Time
	UpdatedAt      time.Time
	AdditionalData map[string]any
}
