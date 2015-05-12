// Functions dealing with Org related API requests. Some of these are direct,
// pass-through implementations of Github API requests, others collect,
// manipulate and/or summarize the data generated by them
package git

import (
	"fmt"
)

// Pass-through for GET /orgs/:org
func Org(org string, token OAuthToken) map[string]interface{} {
	req := NewRequest(fmt.Sprintf("orgs/%s", org))
	js := APIRequest(req, token)
	return js[0]
}

// Pass-through for GET /orgs/:org/members
func OrgMembers(org string, token OAuthToken) []map[string]interface{} {
	req := NewRequest(fmt.Sprintf("orgs/%s/members", org))
	return APIRequest(req, token)
}

// Returns a string slice of all member github handles, useful for iterating
// over all members of an org and getting more detailed stats
func OrgMemberHandles(org string, token OAuthToken) []string {
	members := OrgMembers(org, token)
	vals := ValuesForKey("login", members)
	return StringifyInterfaceSlice(vals)
}
