package model

type UsersData struct {
	Users      []User  `json:"users"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

type UsersResponse struct {
	Response UsersData `json:"response"`
}

type SingleUserResponse struct {
	Response User `json:"response"`
}

type DevicesData struct {
	Total   int      `json:"total"`
	Devices []Device `json:"devices"`
}

type DevicesResponse struct {
	Response DevicesData `json:"response"`
}
