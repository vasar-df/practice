package user

import (
	"github.com/vasar-network/vails"
	"github.com/vasar-network/vails/role"
	"golang.org/x/exp/slices"
	"sort"
	"sync"
	"time"
)

// Roles is a user-based role manager for both offline users and online users.
type Roles struct {
	roleMu          sync.Mutex
	roles           []vails.Role
	roleExpirations map[vails.Role]time.Time
}

// NewRoles creates a new Roles instance.
func NewRoles(roles []vails.Role, expirations map[vails.Role]time.Time) *Roles {
	return &Roles{
		roles:           roles,
		roleExpirations: expirations,
	}
}

// Add adds a role to the manager's role list.
func (r *Roles) Add(ro vails.Role) {
	r.roleMu.Lock()
	r.roles = append(r.roles, ro)
	r.roleMu.Unlock()
	r.sortRoles()
}

// Remove removes a role from the manager's role list. Users are responsible for updating the highest role usages if
// changed.
func (r *Roles) Remove(ro vails.Role) bool {
	if _, ok := ro.(role.Default); ok {
		// You can't remove the default role.
		return false
	}

	r.roleMu.Lock()
	i := slices.IndexFunc(r.roles, func(other vails.Role) bool {
		return ro == other
	})
	r.roles = slices.Delete(r.roles, i, i+1)
	delete(r.roleExpirations, ro)
	r.roleMu.Unlock()
	r.sortRoles()
	return true
}

// Staff returns true if the roles contains a staff role.
func (r *Roles) Staff() bool {
	return r.Contains(role.Trial{}, role.Operator{})
}

// Contains returns true if the manager has any of the given roles. Users are responsible for updating the highest role
// usages if changed.
func (r *Roles) Contains(roles ...vails.Role) bool {
	r.roleMu.Lock()
	defer r.roleMu.Unlock()

	var actualRoles []vails.Role
	for _, ro := range r.roles {
		r.propagateRoles(&actualRoles, ro)
	}

	for _, r := range roles {
		if i := slices.IndexFunc(actualRoles, func(other vails.Role) bool {
			return r == other
		}); i >= 0 {
			return true
		}
	}
	return false
}

// Expiration returns the expiration time for a role. If the role does not expire, the second return value will be false.
func (r *Roles) Expiration(ro vails.Role) (time.Time, bool) {
	r.roleMu.Lock()
	defer r.roleMu.Unlock()
	e, ok := r.roleExpirations[ro]
	return e, ok
}

// Expire sets the expiration time for a role. If the role does not expire, the second return value will be false.
func (r *Roles) Expire(ro vails.Role, t time.Time) {
	r.roleMu.Lock()
	defer r.roleMu.Unlock()
	r.roleExpirations[ro] = t
}

// Highest returns the highest role the manager has, in terms of hierarchy.
func (r *Roles) Highest() vails.Role {
	r.roleMu.Lock()
	defer r.roleMu.Unlock()
	return r.roles[len(r.roles)-1]
}

// All returns the user's roles.
func (r *Roles) All() []vails.Role {
	r.roleMu.Lock()
	defer r.roleMu.Unlock()
	return append(make([]vails.Role, 0, len(r.roles)), r.roles...)
}

// propagateRoles propagates roles to the user's role list.
func (r *Roles) propagateRoles(actualRoles *[]vails.Role, role vails.Role) {
	*actualRoles = append(*actualRoles, role)
	if h, ok := role.(vails.HeirRole); ok {
		r.propagateRoles(actualRoles, h.Inherits())
	}
}

// sortRoles sorts the roles in the user's role list.
func (r *Roles) sortRoles() {
	sort.SliceStable(r.roles, func(i, j int) bool {
		return role.Tier(r.roles[i]) < role.Tier(r.roles[j])
	})
}
