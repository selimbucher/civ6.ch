package storage

// S3Backend will target Hetzner Object Storage (S3-compatible API).
//
// Required env vars when switching to s3:
//
//	S3_ENDPOINT   — e.g. https://nbg1.your-objectstorage.com
//	S3_BUCKET     — bucket name
//	S3_REGION     — e.g. eu-central-1
//	S3_ACCESS_KEY
//	S3_SECRET_KEY
//
// Implementation: use github.com/aws/aws-sdk-go-v2/service/s3 — it speaks
// the S3 protocol that Hetzner exposes.  For public reads, return a plain
// https URL; for private objects, generate a presigned GET URL (15 min TTL).
//
// TODO: implement once bucket credentials are available.
