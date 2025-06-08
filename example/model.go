package example

// go run main.go --input example/model.go --output gen/user_dto.go --type User

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
