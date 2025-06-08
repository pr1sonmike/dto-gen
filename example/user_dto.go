package example

type UserDTO struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

func ToUserDTO(in User) UserDTO {
	return UserDTO{
		ID: in.ID,
		Name: in.Name,
		Email: in.Email,
	}
}
