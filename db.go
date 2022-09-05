package helpers

import (
	"fmt"
)

func BuildDbConnString(
	Type string,
	Host string,
	Port int,
	Name string,
	User string,
	Pass string,
	SSlMode string,
) string {
	var connUrl string = ""
	connUrl = fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?sslmode=%s",
		Type,
		User,
		Pass,
		Host,
		Port,
		Name,
		SSlMode,
	)

	return connUrl
}
