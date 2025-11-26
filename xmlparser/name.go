package xmlparser

import "strings"

type Name string

func (n Name) Split() (string, string) {
	ind := strings.IndexByte(string(n), ':')
	if ind == -1 {
		return "", string(n)
	}
	return string(n[:ind]), string(n[ind+1:])
}

func (n Name) Space() string {
	space, _ := n.Split()
	return space
}

func (n Name) Local() string {
	_, local := n.Split()
	return local
}

func (n Name) String() string {
	return string(n)
}
