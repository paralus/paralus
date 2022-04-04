package util

import (
	"fmt"
	"strconv"
	"strings"
)

func getAlterString(v int) string {
	s := strconv.Itoa(v)
	return "-" + s
}

func stripCtlFromBytes(inputStr string) string {
	str := strings.ToLower(inputStr)
	runes := []rune(str)
	var ret string
	for i := 0; i < len(runes); i++ {
		ascii := int(runes[i])
		if ascii < 48 || (ascii > 57) && (ascii < 97) || (ascii > 122) {
			ret = ret + getAlterString(ascii)
		} else {
			s := rune(ascii)
			ret = ret + string(s)
		}
	}
	return ret
}

// SanitizeUsername sanitizes username as a k8s name
func SanitizeUsername(username string) (name string) {
	name = strings.ToLower(username)
	name = stripCtlFromBytes(name)
	if len(name) > 253 {
		name = string([]rune(name)[:253])
	}
	return
}

const (
	templatePrefix        = "template/"
	templatePrefixLen     = len(templatePrefix)
	clusterPrefix         = "cluster/"
	clusterPrefixLen      = len(clusterPrefix)
	userPrefix            = "user/"
	userPrefixLen         = len(userPrefix)
	ssoUserPrefix         = "ssouser/"
	ssoUserPrefixLen      = len(ssoUserPrefix)
	organizationPrefix    = "organization/"
	organizationPrefixLen = len(organizationPrefix)
)

// GetTemplateScope returns template scope from url
func GetTemplateScope(templateScope string) (scope string, err error) {
	if strings.HasPrefix(templateScope, templatePrefix) {
		scope = templateScope[templatePrefixLen:]
		return
	}
	err = fmt.Errorf("invalid template scope %s", templateScope)

	return
}

// ToTemplateScope converts scope into template scope
func ToTemplateScope(scope string) string {
	return fmt.Sprintf("%s%s", templatePrefix, scope)
}

// GetClusterScope returns cluster scope from url
func GetClusterScope(clusterScope string) (scope string, err error) {
	if strings.HasPrefix(clusterScope, clusterPrefix) {
		scope = clusterScope[clusterPrefixLen:]
		return
	}
	err = fmt.Errorf("invalid template scope %s", clusterScope)

	return
}

// ToClusterScope converts scope into cluster scope
func ToClusterScope(scope string) string {
	return fmt.Sprintf("%s%s", clusterPrefix, scope)
}

// GetUserScope returns user scope from url
func GetUserScope(userScope string) (scope string, err error) {
	if strings.HasPrefix(userScope, ssoUserPrefix) {
		scope = userScope[ssoUserPrefixLen:]
		return
	}
	if strings.HasPrefix(userScope, userPrefix) {
		scope = userScope[userPrefixLen:]
		return
	}
	err = fmt.Errorf("invalid template scope %s", userScope)

	return
}

// ToUserScope converts scope into user scope
func ToUserScope(scope string) string {
	return fmt.Sprintf("%s%s", userPrefix, scope)
}

// GetOrganizationScope returns organization scope from url
func GetOrganizationScope(organizationScope string) (scope string, err error) {
	if strings.HasPrefix(organizationScope, organizationPrefix) {
		scope = organizationScope[organizationPrefixLen:]
		return
	}
	err = fmt.Errorf("invalid template scope %s", organizationScope)

	return
}

// ToOrganizationScope converts scope into organization scope
func ToOrganizationScope(scope string) string {
	return fmt.Sprintf("%s%s", organizationPrefix, scope)
}
