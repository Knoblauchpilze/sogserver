package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// upgradablesModule :
// Fill a similar role to the `ResourcesModule` (see doc in
// this file for more info). The goal of this module is to
// provide a base for other game elements such as the ships,
// buildings, technologies and defenses. All these elements
// have in common that they cannot be produced unless some
// specific conditions are met (meaning dependencies exists
// between elements). This provide a common structure for
// the way to store some data and we want to mutualize it
// as much as possible.
//
// The `uType` defines the type of the upgradable associated
// this module. It will help to determine the tables that
// need to be fetched to populate the internal data.
//
// The `buildingsDeps` is used to register the buildings
// dependencies for each element associated to this module.
//
// The `techDeps` fills a similar role to `buildingDeps`
// but is used to register dependencies on technologies.
type upgradablesModule struct {
	associationTable
	baseModule

	uType         upgradable
	buildingsDeps map[string][]Dependency
	techDeps      map[string][]Dependency
}

// upgradable :
// Describes the underlying support for an upgradable.
type upgradable int

// Define the possible upgradable types in the game.
const (
	Building upgradable = iota
	Technology
	Ship
	Defense
)

// String :
// Implementation of the `Stringer` interface to allow
// easy manipulation of an upgradable type.
//
// Returns the string corresponding to this upgradable.
func (u upgradable) String() string {
	switch u {
	case Building:
		return "buildings"
	case Technology:
		return "technologies"
	case Ship:
		return "ships"
	case Defense:
		return "defenses"
	default:
		return "unknown"
	}
}

// UpgradableDesc :
// Defines the abstract representation of a game element.
// This description defines the identifier and name for
// the module along with some dependencies, both in terms
// of buildings and technologies.
//
// The `ID` defines the unique identifier for this element.
//
// The `Name` defines a human readable name associated to
// the element. It is mostly used for display purposes.
//
// The `BuildingDeps` defines a list of dependencies that
// should be met for this upgradable to be available on a
// planet (or to a player).
//
// The `TechnologiesDeps` fills a similar purpose but is
// used to register dependencies on technologies.
type UpgradableDesc struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	BuildingsDeps    []Dependency `json:"buildings_dependencies"`
	TechnologiesDeps []Dependency `json:"technologies_dependencies"`
}

// Dependency :
// Defines a way to represent a dependency between an element
// of the game and another. It defines only the prerequisite
// to the element itself along with the level that the item
// should reach so that the dependency is met.
//
// The `ID` defines an identifier of the requirement needed
// for the element. It should correspond to a valid element.
//
// The `Level` defines a minimum threshold below which the
// dependency is not met.
type Dependency struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
}

// newUpgradablesModule :
// Used to create a new module allowing to handle the list
// of upgradables of some sort. No dependencies are created
// right now, one should call the `init` method to do so.
//
// The `log` defines the logging layer to forward to the
// base `baseModule` element.
//
// The `kind` defines the type of upgradable associated to
// this module. It will help to determine which tables are
// relevant when fetching data from the DB.
//
// The `module` defines the string identifying the module
// to forward to the logging layer.
func newUpgradablesModule(log logger.Logger, kind upgradable, module string) *upgradablesModule {
	return &upgradablesModule{
		associationTable: associationTable{},
		baseModule:       newBaseModule(log, module),
		uType:            kind,
		buildingsDeps:    nil,
		techDeps:         nil,
	}
}

// valid :
// Refinement of the base `associationTable` valid method
// in order to perform some checks on the dependencies that
// are loaded in this module.
//
// Returns `true` if the association table is valid and
// the internal resources as well.
func (um *upgradablesModule) valid() bool {
	return um.associationTable.valid() && len(um.buildingsDeps) > 0 && len(um.techDeps) > 0
}

// Init :
// Implementation of the `DBModule` interface to allow
// fetching information from the input DB and load to
// local memory.
//
// The `dbase` represents the main data source to use
// to initialize the resources data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (um *upgradablesModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		um.trace(logger.Error, fmt.Sprintf("Unable to initialize upgradables module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if um.valid() && !force {
		return nil
	}

	// Initialize internal values.
	um.buildingsDeps = make(map[string][]Dependency)
	um.techDeps = make(map[string][]Dependency)

	proxy := db.NewProxy(dbase)

	// First update the buildings dependencies.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"requirement",
			"level",
		},
		Table:   fmt.Sprintf("tech_tree_%s_vs_buildings", um.uType),
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		um.trace(logger.Error, fmt.Sprintf("Unable to initialize upgradables %s module (err: %v)", um.uType, err))
		return ErrNotInitialized
	}

	err = um.initDeps(rows, &um.buildingsDeps)
	if err != nil {
		um.trace(logger.Error, fmt.Sprintf("Unable to initialize upgradables %s module (err: %v)", um.uType, err))
		return err
	}

	// Also update the technologies dependencies.
	rows.Close()
	query.Table = fmt.Sprintf("tech_tree_%s_vs_technologies", um.uType)

	rows, err = proxy.FetchFromDB(query)
	defer rows.Close()

	err = um.initDeps(rows, &um.techDeps)
	if err != nil {
		um.trace(logger.Error, fmt.Sprintf("Unable to initialize upgradables %s module (err: %v)", um.uType, err))
		return err
	}

	return nil
}

// initDeps :
// Used to initialize the provided dependencies map from
// the input query result. If the query is invalid or if
// some inconsistencies are detected in the data an error
// will be raised.
//
// The `rows` defines the query result to analyze.
//
// The `deps` defines the dependencies map to populate.
//
// Returns any error.
func (um *upgradablesModule) initDeps(rows db.QueryResult, deps *map[string][]Dependency) error {
	if rows.Err != nil {
		um.trace(logger.Error, fmt.Sprintf("Invalid query to initialize upgradables %s module (err: %v)", um.uType, rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var elem, req string
	var dep int

	override := false
	sanity := make(map[string]map[string]int)

	for rows.Next() {
		err := rows.Scan(
			&elem,
			&req,
			&dep,
		)

		if err != nil {
			um.trace(logger.Error, fmt.Sprintf("Failed to initialize dependency from row (err: %v)", err))
			continue
		}

		// Check for overrides.
		eDeps, ok := sanity[elem]
		if !ok {
			eDeps = make(map[string]int)
			eDeps[req] = dep
		} else {
			e, ok := eDeps[req]
			if ok {
				um.trace(logger.Error, fmt.Sprintf("Prevented override of upgradable \"%s\" dependency on \"%s\" (%d to %d)", elem, req, e, dep))
				override = true

				continue
			}

			eDeps[req] = dep
		}

		sanity[elem] = eDeps

		// Register this value.
		uDeps, ok := (*deps)[elem]
		if !ok {
			uDeps = make([]Dependency, 0)
		}

		uDeps = append(
			uDeps,
			Dependency{},
		)

		(*deps)[elem] = uDeps
	}

	if override {
		return ErrInconsistentDB
	}

	return nil
}

// getDependencyFromID :
// Used to retrieve information on the dependencies
// for the input identifier assuming it is an element
// of the same type described by this module. If no
// element with a corresponding ID exists an error is
// returned.
//
// The `id` defines the identifier of the upgradable
// element to fetch.
//
// Returns the description for the dependencies of
// the element along with any errors.
func (um *upgradablesModule) getDependencyFromID(id string) (UpgradableDesc, error) {
	// Find this element in the association table.
	if um.existsID(id) {
		um.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for upgradable \"%s\"", id))
		return UpgradableDesc{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := um.getNameFromID(id)

	// If the key does not exist the zero value will be
	// assigned to the left operands which is okay (and
	// even desired).
	res := UpgradableDesc{
		ID:               id,
		Name:             name,
		BuildingsDeps:    um.buildingsDeps[id],
		TechnologiesDeps: um.techDeps[id],
	}

	return res, nil
}

// getDependencyFromName :
// Calls internally the `getDependencyFromID` in order
// to forward the call to the above method. Failures
// happen in similar cases.
//
// The `name` defines the name of the upgradable elem
// for which a description should be provided.
//
// Returns the description for this resource along with
// any errors.
func (um *upgradablesModule) getDependencyFromName(name string) (UpgradableDesc, error) {
	// Find this element in the association table.
	id, err := um.getIDFromName(name)
	if err != nil {
		um.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for upgradable \"%s\" (err: %v)", name, err))
		return UpgradableDesc{}, ErrNotFound
	}

	return um.getDependencyFromID(id)
}
