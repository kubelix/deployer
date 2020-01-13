package service

var ptrOne, ptrThree *int32

func init() {
	ptrOne = ptrNum(1)
	ptrThree = ptrNum(3)
}

func ptrNum(num int32) *int32 {
	return &num
}
