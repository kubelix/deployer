package service

var ptrOne, ptrThree *int32

func init() {
	ptrOne = ptrInt32(1)
	ptrThree = ptrInt32(3)
}

func ptrInt32(num int32) *int32 {
	return &num
}

func ptrInt64(num int64) *int64 {
	return &num
}

const (
	dockerConfigContent = `{"auths": {"%s": {"auth": "%s"}}}`
	fieldIsImmutable = "field is immutable"
)
