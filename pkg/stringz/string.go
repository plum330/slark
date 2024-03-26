package stringz

func SnakeCase(s string) string {
	l := len(s)
	b := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			b = append(b, '_')
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return string(b)
}
