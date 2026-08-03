package entra

type SignInEvent struct {
	ID                      string `json:"id"`
	CreatedDateTime         string `json:"createdDateTime"`
	UserDisplayName         string `json:"userDisplayName"`
	UserPrincipalName       string `json:"userPrincipalName"`
	UserID                  string `json:"userId"`
	AppID                   string `json:"appId"`
	AppDisplayName          string `json:"appDisplayName"`
	IPAddress               string `json:"ipAddress"`
	ClientAppUsed           string `json:"clientAppUsed"`
	CorrelationID           string `json:"correlationId"`
	ConditionalAccessStatus string `json:"conditionalAccessStatus"`
	IsInteractive           bool   `json:"isInteractive"`

	RiskDetail            string `json:"riskDetail"`
	RiskLevelAggregated   string `json:"riskLevelAggregated"`
	RiskLevelDuringSignIn string `json:"riskLevelDuringSignIn"`
	RiskState             string `json:"riskState"`

	ResourceDisplayName string `json:"resourceDisplayName"`
	ResourceID          string `json:"resourceId"`

	Status                           Status                    `json:"status"`
	DeviceDetail                     DeviceDetail              `json:"deviceDetail"`
	Location                         Location                  `json:"location"`
	AppliedConditionalAccessPolicies []ConditionalAccessPolicy `json:"appliedConditionalAccessPolicies"`
}

type Status struct {
	ErrorCode         int     `json:"errorCode"`
	FailureReason     *string `json:"failureReason"`
	AdditionalDetails *string `json:"additionalDetails"`
}

type DeviceDetail struct {
	DeviceID        string  `json:"deviceId"`
	DisplayName     *string `json:"displayName"`
	OperatingSystem string  `json:"operatingSystem"`
	Browser         string  `json:"browser"`
	IsCompliant     *bool   `json:"isCompliant"`
	IsManaged       *bool   `json:"isManaged"`
	TrustType       *string `json:"trustType"`
}

type Location struct {
	City            string         `json:"city"`
	State           string         `json:"state"`
	CountryOrRegion string         `json:"countryOrRegion"`
	GeoCoordinates  GeoCoordinates `json:"geoCoordinates"`
}

type GeoCoordinates struct {
	Altitude  *float64 `json:"altitude"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
}

type ConditionalAccessPolicy struct {
	ID                      string   `json:"id"`
	DisplayName             string   `json:"displayName"`
	EnforcedGrantControls   []string `json:"enforcedGrantControls"`
	EnforcedSessionControls []string `json:"enforcedSessionControls"`
	Result                  string   `json:"result"`
}

type GraphResponse struct {
	Value    []SignInEvent `json:"value"`
	NextLink string        `json:"@odata.nextLink"`
}
