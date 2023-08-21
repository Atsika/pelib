package pelib

// sdbm hashing algorithm
func Hash(str string) uint32 {
	var hash uint32 = 0
	for i := 0; i < len(str); i++ {
		hash = uint32(str[i]) + (hash << 6) + (hash << 16) - hash
	}
	return hash
}
