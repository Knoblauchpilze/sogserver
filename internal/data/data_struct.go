package data

// ProductionEffect :
// Defines a production effect that a building upgrade
// action can have on the production of a planet. It is
// used to regroup the resource and the value of the
// change brought by the building upgrade action.
//
// The `Action` defines the identifier of the action to
// which this effect is linked.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Effect` defines the actual effect of the upgrade
// action. This value should be substituted to the planet
// production if the upgrade action completes.
type ProductionEffect struct {
	Action   string  `json:"action"`
	Resource string  `json:"res"`
	Effect   float32 `json:"new_production"`
}

// StorageEffect :
// Defines a storage effect that a building upgrade
// action can have on the capacity of a resource that
// can be stored on a planet. It is used to regroup
// the resource and the value of the change brought
// by the building upgrade action.
//
// The `Action` defines the identifier of the action
// to which this effect is linked.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Effect` defines the actual effect of the upgrade
// action. This value should be substituted to the planet
// storage capacity for the resource if the upgrade
// action completes.
type StorageEffect struct {
	Action   string  `json:"action"`
	Resource string  `json:"res"`
	Effect   float32 `json:"new_storage_capacity"`
}
