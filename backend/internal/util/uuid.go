package util

import (
	"github.com/google/uuid"
)

// SecraNamespace is a fixed UUID for generating deterministic v5 UUIDs
var SecraNamespace = uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

// NewUUIDv5 generates a deterministic UUID based on a name string
func NewUUIDv5(name string) string {
	return uuid.NewSHA1(SecraNamespace, []byte(name)).String()
}

// Generate IDs for specific models
func VendorID(name string) string { return NewUUIDv5("vendor:" + name) }
func ProductID(vendorName, productName string) string { 
	return NewUUIDv5("product:" + vendorName + ":" + productName) 
}
func CVEID(sourceUID string) string { return NewUUIDv5("cve:" + sourceUID) }
func SourceID(name string) string { return NewUUIDv5("source:" + name) }
