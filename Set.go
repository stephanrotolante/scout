package main

type Set map[string]struct{}

func NewSet() Set {
	return make(Set)
}

func (s Set) Add(item string) {
	s[item] = struct{}{}
}

func (s Set) Remove(item string) {
	delete(s, item)
}

func (s Set) Contains(item string) bool {
	_, found := s[item]
	return found
}

func (s Set) GetKeys() []string {

	keys := make([]string, 0, len(s))

	// Iterate over the map and append keys to the slice
	for key := range s {
		keys = append(keys, key)
	}

	return keys
}
