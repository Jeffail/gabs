// +build gofuzz

package gabs

func Fuzz(input []byte) int {
	result, err := ParseJSON(input)
	if err != nil {
		return 0
	}
	result.String()
	return 1
}
