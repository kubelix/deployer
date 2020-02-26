package service

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
