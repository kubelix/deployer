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

func mergeLabels(labels1, labels2 map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range labels1 {
		result[k] = v
	}
	for k, v := range labels2 {
		result[k] = v
	}

	return result
}

const (
	dockerConfigContent = `{"auths": {"%s": {"auth": "%s"}}}`
	fieldIsImmutable = "field is immutable"
)
