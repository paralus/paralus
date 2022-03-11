package service

func unique(items []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range items {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
