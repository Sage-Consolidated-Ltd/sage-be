package entra

type SignInEvent struct {
	ID              string `json:"id"`
	UserDisplayName string `json:"userDisplayName"`
	UserID          string `json:"userId"`
	AppDisplayName  string `json:"appDisplayName"`
	IPAddress       string `json:"ipAddress"`
	CreatedDateTime string `json:"createdDateTime"`
	Status          Status `json:"status"`
}

type Status struct {
	ErrorCode int `json:"errorCode"`
}

type GraphResponse struct {
	Value    []SignInEvent `json:"value"`
	NextLink string        `json:"@odata.nextLink"`
}

type Checkpoint struct {
	LastCreatedTime string `json:"last_created_time"`
	UpdatedAt       string `json:"updated_at"`
}
