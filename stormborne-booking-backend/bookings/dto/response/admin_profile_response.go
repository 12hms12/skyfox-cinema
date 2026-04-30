package response

type AdminProfileResponse struct{
	Id int
	Name string
	Username string
	CounterNo int
}

func NewAdminProfileResponse(Id int,Name string,Username string,CounterNo int) *AdminProfileResponse{
	return &AdminProfileResponse{
		Id: Id,
		Name: Name,
		Username: Username,
		CounterNo: CounterNo,
	}
}
