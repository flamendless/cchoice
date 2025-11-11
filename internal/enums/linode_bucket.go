package enums

//go:generate go tool stringer -type=LinodeBucketEnum -trimprefix=LINODE_BUCKET_

type LinodeBucketEnum int

const (
	LINODE_BUCKET_UNDEFINED LinodeBucketEnum = iota
	LINODE_BUCKET_PUBLIC
	LINODE_BUCKET_PRIVATE
)

func ParseLinodeBucketEnum(e string) LinodeBucketEnum {
	switch e {
	case "PUBLIC":
		return LINODE_BUCKET_PUBLIC
	case "PRIVATE":
		return LINODE_BUCKET_PRIVATE
	case "UNDEFINED":
		return LINODE_BUCKET_UNDEFINED
	default:
		return LINODE_BUCKET_UNDEFINED
	}
}
