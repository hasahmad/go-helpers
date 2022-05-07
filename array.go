package helpers

func InArray[T string | int](subset []T, allSet []T, checkAll bool) bool {
	count := 0
	for _, v := range allSet {
		for _, s := range subset {
			if s == v {
				if !checkAll {
					return true
				}
				count += 1
			}
		}
	}

	return count == len(subset)
}
