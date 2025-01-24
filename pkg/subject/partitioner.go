package subject

func GetLastPartition(subject string) string {
	var lastDot int
	for i := 0; i < len(subject); i++ {
		if subject[i] == '.' {
			lastDot = i
		}
	}
	return subject[lastDot+1:]
}
