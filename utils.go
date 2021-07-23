package main

func mergeTags(one, two map[string]string) map[string]string {
	for k, v := range two {
		one[k] = v
	}
	return one
}
