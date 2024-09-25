package awsconditions

import (
	"net"
	"strings"
	"time"
)

func matchPattern(s, p string) bool {
	// Initialize the DP table
	m, n := len(s), len(p)
	dp := make([][]bool, m+1)
	for i := range dp {
		dp[i] = make([]bool, n+1)
	}

	// Base case: empty string and empty pattern match
	dp[0][0] = true

	// Handle patterns with '*' at the beginning
	for j := 1; j <= n; j++ {
		if p[j-1] == '*' {
			dp[0][j] = dp[0][j-1]
		}
	}

	// Fill the DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if p[j-1] == '*' {
				dp[i][j] = dp[i][j-1] || dp[i-1][j]
			} else if p[j-1] == '?' || s[i-1] == p[j-1] {
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	return dp[m][n]
}

func StringEquals(a string, b string) bool {
	return a == b
}

func StringNotEquals(a string, b string) bool {
	return !StringEquals(a, b)
}

func StringEqualsIgnoreCase(a string, b string) bool {
	// convert each string to lower case
	// and compare them
	return strings.ToLower(a) == strings.ToLower(b)
}

func StringNotEqualsIgnoreCase(a string, b string) bool {
	// convert each string to lower case
	// and compare them
	return !StringEqualsIgnoreCase(a, b)
}

func StringLike(a string, b string) bool {
	// Case-sensitive matching. The values can include multi-character
	// match wildcards (*) and single-character match wildcards (?)
	// anywhere in the string. You must specify wildcards to achieve
	// partial string matches.
	return matchPattern(a, b)

}

func StringNotLike(a string, b string) bool {
	return !StringLike(a, b)
}

func NumericEquals(a int, b int) bool {
	return a == b
}

func NumericNotEquals(a int, b int) bool {
	return !NumericEquals(a, b)
}

func NumericLessThan(a int, b int) bool {
	return a < b
}

func NumericLessThanEquals(a int, b int) bool {
	return a <= b
}

func NumericGreaterThan(a int, b int) bool {
	return a > b
}

func NumericGreaterThanEquals(a int, b int) bool {
	return a >= b
}

func DateEquals(a string, b string) bool {
	const iso8601 = "2000-01-01T00:00:00Z00:00"
	d1, err := time.Parse(iso8601, a)
	if err == nil {
		return false
	}

	d2, err := time.Parse(iso8601, b)
	if err == nil {
		return false
	}

	return d1 == d2
}

func DateNotEquals(a string, b string) bool {
	return !DateEquals(a, b)
}

func DateLessThan(a string, b string) bool {
	const iso8601 = "2000-01-01T00:00:00Z00:00"
	d1, err := time.Parse(iso8601, a)
	if err == nil {
		return false
	}

	d2, err := time.Parse(iso8601, b)
	if err == nil {
		return false
	}

	return d1.Before(d2)
}

func DateLessThanEquals(a string, b string) bool {
	const iso8601 = "2000-01-01T00:00:00Z00:00"
	d1, err := time.Parse(iso8601, a)
	if err == nil {
		return false
	}

	d2, err := time.Parse(iso8601, b)
	if err == nil {
		return false
	}

	return d1.Before(d2) || d1 == d2
}

func DateGreaterThan(a string, b string) bool {
	return !DateLessThanEquals(a, b)
}

func DateGreaterThanEquals(a string, b string) bool {
	return !DateLessThan(a, b)
}

func Bool(a string) bool {
	return a == "true"
}

func IpAddress(a string, b string) bool {
	// Determine if a is equal or in the CIDR range
	// of b

	// Convert a to an ip address
	parsedIP1 := net.ParseIP(a)
	if parsedIP1 == nil {
		return false
	}

	// Try to parse the second IP as an IP address
	parsedIP2 := net.ParseIP(b)
	if parsedIP2 != nil {
		// If both IPs are parsed successfully, compare them for equality
		return parsedIP1.Equal(parsedIP2)
	}

	// If the second IP is not a simple IP address, try to parse it as a CIDR range
	_, ipNet, err := net.ParseCIDR(b)
	if err != nil {
		return false
	}

	// Check if the first IP is within the CIDR range
	return ipNet.Contains(parsedIP1)
}

func NotIpAddress(a string, b string) bool {
	return !IpAddress(a, b)
}

func ArnEquals(a string, b string) bool {
	// Case-sensitive matching of the ARN. Each of the six
	// colon-delimited components of the ARN is checked separately
	// and each can include multi-character match wildcards (*)
	// or single-character match wildcards (?).

	// First split the ARNs by colon
	arn1 := strings.Split(a, ":")
	arn2 := strings.Split(b, ":")

	for i := 0; i < 6; i++ {
		if !StringLike(arn1[i], arn2[i]) {
			return false
		}
	}

	return true
}

func ArnNotEquals(a string, b string) bool {
	return !ArnEquals(a, b)
}

func ArnLike(a string, b string) bool {
	return ArnEquals(a, b)
}

func ArnNotLike(a string, b string) bool {
	return !ArnEquals(a, b)
}

func IfExists(conditionKey string) bool {
	return Null(conditionKey, false)
}

func Null(conditionKey string, isNull bool) bool {
	// TODO
	// Get the context of the path
	// pathContext := context.GetContext(path)
	// if pathContext.Get(conditionKey) == nil {
	// 	return true ^ isNull
	// }
	// return false ^ isNull
	return true
}
