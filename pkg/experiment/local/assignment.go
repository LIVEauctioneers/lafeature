package local

import (
	"sort"
	"strings"
	"time"

	"github.com/LIVEauctioneers/lafeature/pkg/experiment"
)

type assignment struct {
	user      *experiment.User
	results   map[string]experiment.Variant
	timestamp int64
}

func newAssignment(user *experiment.User, results map[string]experiment.Variant) *assignment {
	assignment := &assignment{
		user:      user,
		results:   results,
		timestamp: time.Now().UnixMilli(),
	}

	return assignment
}

func (a *assignment) Canonicalize() string {
	var sb strings.Builder

	if a.user != nil {
		sb.WriteString(a.user.UserId)
		sb.WriteString(" ")
		sb.WriteString(a.user.DeviceId)
		sb.WriteString(" ")
	}

	keys := make([]string, 0, len(a.results))
	for key := range a.results {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := a.results[key].Key
		sb.WriteString(key)
		sb.WriteString(" ")
		sb.WriteString(value)
		sb.WriteString(" ")
	}

	return sb.String()
}
