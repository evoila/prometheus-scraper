package model

// ServiceInstance defined for each Service Instance
type ServiceInstance struct {
	ID                  string                 `json:"id" bson:"id"`
	ServiceDefinitionID string                 `json:"serviceDefinitionId" bson:"serviceDefinitionId"`
	PlanID              string                 `json:"planId" bson:"planId"`
	OrganizationGUID    string                 `json:"organizationGuid" bson:"organizationGuid"`
	SpaceGUID           string                 `json:"spaceGuid" bson:"spaceGuid"`
	DashboardURL        string                 `json:"dashboardUrl" bson:"dashboardUrl"`
	Parameters          map[string]interface{} `json:"parameters" bson:"parameters"`
	InternalID          string                 `json:"internalId" bson:"internalId"`
	Hosts               []ServerAddress        `json:"hosts" bson:"hosts"`
	Context             map[string]string      `json:"context" bson:"context"`
}
