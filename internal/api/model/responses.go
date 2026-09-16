package model

type UsersResponse struct {
	Response UsersData `json:"response"`
}

type UsersData struct {
	Users      []User  `json:"users"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

type SingleUserResponse struct {
	Response User `json:"response"`
}

type DevicesResponse struct {
	Response DevicesData `json:"response"`
}

type DevicesData struct {
	Total   int      `json:"total"`
	Devices []Device `json:"devices"`
}
